use clap::{App, Arg};
use anyhow::{format_err, Error};
use libc::pid_t;
use libc::{rlimit, setrlimit, RLIMIT_STACK, RLIM_INFINITY};
use log::{info, warn};
use nix::mount::{mount, MsFlags};
use nix::sched::{unshare, CloneFlags};
use nix::sys::signal::{kill, SIGKILL};
use nix::sys::wait::{waitpid, WaitStatus};
use nix::unistd::{
    chdir, chroot, close, execvp, fork, getpid, pipe, read, setgid, setuid, write, ForkResult, Gid,
    Pid, Uid,
};
use rand::distributions::Alphanumeric;
use rand::Rng;
use std::env;
use std::ffi::CString;
use std::fs::File;
use std::fs::{create_dir, read_to_string, remove_dir_all, set_permissions, OpenOptions};
use std::io;
use std::io::Write;
use std::iter;
use std::mem::size_of;
use std::os::unix::fs::PermissionsExt;
use std::os::unix::io::RawFd;
use std::path::{Path, PathBuf};
use std::process::exit;
use std::process::Command;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::thread;
use std::time::{Duration, Instant};

fn random_string(len: usize) -> String {
    let mut rng = rand::thread_rng();
    iter::repeat(())
        .map(|()| rng.sample(Alphanumeric))
        .take(len)
        .collect()
}

fn tempdir(temp_dir: &Path) -> io::Result<PathBuf> {
    for _ in 0..10 {
        let chars: String = random_string(10);
        let temp_dir = temp_dir.join(chars);
        if temp_dir.exists() {
            continue;
        }
        create_dir(&temp_dir)?;
        return Ok(temp_dir);
    }
    Err(io::Error::new(
        io::ErrorKind::Other,
        "failed to create a temp dir",
    ))
}

fn prepare_mount(base_dir: &Path, temp_dir: &Path, overlay: bool) -> Result<(), Error> {
    let root_dir = temp_dir.join("root");
    let sand_dir = root_dir.join("sand");
    create_dir(&root_dir)?;
    create_dir(&sand_dir)?;

    if overlay {
        let work_dir = tempdir(&temp_dir)?;
        let upper_dir = tempdir(&temp_dir)?;
        set_permissions(&work_dir, PermissionsExt::from_mode(0o777))?;
        set_permissions(&upper_dir, PermissionsExt::from_mode(0o777))?;
        let option = format!(
            "lowerdir={},upperdir={},workdir={}",
            "./",
            upper_dir.to_str().unwrap_or(""),
            work_dir.to_str().unwrap_or("")
        );
        mount(
            None::<&str>,
            &sand_dir,
            Some("overlay"),
            MsFlags::empty(),
            Some(&option[..]),
        )?;
    } else {
        // mount --bind ./ tempdir
        mount(
            Some(base_dir),
            &sand_dir,
            None::<&str>,
            MsFlags::MS_BIND,
            None::<&str>,
        )?;
    }

    // make /tmp
    create_dir(root_dir.join("tmp"))?;
    set_permissions(&root_dir.join("tmp"), PermissionsExt::from_mode(0o777))?;
    // mount -t proc proc /proc
    let proc_dir = root_dir.join("proc");
    create_dir(&proc_dir)?;
    mount(
        Some("proc"),
        &proc_dir,
        Some("proc"),
        MsFlags::empty(),
        None::<&str>,
    )?;
    // mount files
    for dir_name in vec![
        "dev", "sys", "bin", "sbin", "lib", "lib64", "usr", "etc", "opt", "var", "home",
    ] {
        let dir = root_dir.join(dir_name);
        create_dir(&dir)?;
        mount(
            Some(&Path::new("/").join(dir_name)),
            &dir,
            None::<&str>,
            MsFlags::MS_BIND | MsFlags::MS_RDONLY,
            None::<&str>,
        )?;
    }
    Ok(())
}

fn prepare_cgroup(pid: &Pid) -> Result<(), Error> {
    Command::new("cgdelete")
        .arg("pids,cpuset,memory:/lib-judge")
        .status()?;
    if !Command::new("cgcreate")
        .args(&["-g", "pids,cpuset,memory:/lib-judge"])
        .status()?
        .success()
    {
        return Err(format_err!("failed cgcreate"));
    }
    fn cgset(arg: &str) -> Result<(), Error> {
        if !Command::new("cgset")
            .args(&["-r", arg, "/lib-judge"])
            .status()?
            .success()
        {
            return Err(format_err!("failed cgset pids.max"));
        }
        Ok(())
    }
    cgset("pids.max=1000")?;
    cgset("cpuset.cpus=0")?;
    cgset("cpuset.mems=0")?;
    cgset("memory.limit_in_bytes=1G")?;
    cgset("memory.memsw.limit_in_bytes=1G")?;
    fn cgexec(group: &str, pid: &Pid) -> Result<(), Error> {
        let mut file = OpenOptions::new()
            .write(true)
            .truncate(true)
            .open(format!("/sys/fs/cgroup/{}/lib-judge/cgroup.procs", group))?;
        file.write((pid.as_raw().to_string() + "\n").as_bytes())?;
        Ok(())
    }
    cgexec("pids", &pid)?;
    cgexec("cpuset", &pid)?;
    cgexec("memory", &pid)?;
    Ok(())
}

fn change_uid() -> Result<(), Error> {
    let output = Command::new("id")
        .args(&["-g", "library-checker-user"])
        .output()?;
    if !output.status.success() {
        return Err(format_err!("failed: id -g library-checker-user"));
    }
    let gid = String::from_utf8(output.stdout)?.trim().parse()?;
    setgid(Gid::from_raw(gid))?;
    let output = Command::new("id")
        .args(&["-u", "library-checker-user"])
        .output()?;
    if !output.status.success() {
        return Err(format_err!("failed: id -u library-checker-user"));
    }
    let uid = String::from_utf8(output.stdout)?.trim().parse()?;
    setuid(Uid::from_raw(uid))?;

    Ok(())
}

fn execute_unshared(
    app: &clap::ArgMatches,
    temp_dir: &Path,
    user_args: &[String],
    start_pipe_write: RawFd,
) -> Result<(), Error> {
    let base_dir = Path::new(app.value_of("cwd").unwrap_or("."));
    let overlay: bool = app.is_present("overlay");
    prepare_mount(&base_dir, &temp_dir, overlay)?;
    prepare_cgroup(&getpid())?;
    unsafe {
        let res = setrlimit(
            RLIMIT_STACK,
            &rlimit {
                rlim_cur: RLIM_INFINITY,
                rlim_max: RLIM_INFINITY,
            },
        );
        if res != 0 {
            return Err(format_err!("setrlimit(stack) failed"));
        }
    }
    chdir(&temp_dir.join("root").join("sand"))?;
    chroot("..")?;
    change_uid()?;
    env::remove_var("TMPDIR");
    env::set_var("HOME", "/home/library-checker-user");
    let program = CString::new(user_args[0].clone())?;
    let args: Vec<std::ffi::CString> = user_args[..]
        .iter()
        .map(|s| CString::new(s.clone()).unwrap())
        .collect();
    let args: Vec<&std::ffi::CStr> = args.iter().map(|s| &s[..]).collect();
    write(start_pipe_write, &[0])?;
    close(start_pipe_write)?;
    execvp(&program, &args[..])?;
    Ok(())
}
#[derive(Debug)]
pub struct ExecResult {
    pub status: i32,
    pub time: f64,
    pub memory: i64,
    pub tle: bool,
}

fn execute(app: &clap::ArgMatches, user_args: &[String]) -> Result<ExecResult, Error> {
    let temp_dir = tempdir(&std::env::temp_dir())?;
    let (pipe_read, pipe_write) = pipe()?;
    let (start_pipe_read, start_pipe_write) = pipe()?;
    info!("working dir: {:?}", temp_dir);
    set_permissions(&temp_dir, PermissionsExt::from_mode(0o777))?;
    match fork()? {
        ForkResult::Child => {
            close(pipe_read)?;
            close(start_pipe_read)?;
            unshare(CloneFlags::CLONE_NEWPID | CloneFlags::CLONE_NEWNS | CloneFlags::CLONE_NEWNET)
                .expect("unshare failed");
            match fork()? {
                ForkResult::Child => {
                    close(pipe_write)?;
                    mount(
                        None::<&str>,
                        "/",
                        None::<&str>,
                        MsFlags::MS_REC | MsFlags::MS_PRIVATE,
                        None::<&str>,
                    )?;
                    mount(
                        Some("none"),
                        "/proc",
                        None::<&str>,
                        MsFlags::MS_PRIVATE | MsFlags::MS_REC,
                        None::<&str>,
                    )?;
                    mount(
                        Some("proc"),
                        "/proc",
                        Some("proc"),
                        MsFlags::MS_NOSUID | MsFlags::MS_NOEXEC | MsFlags::MS_NODEV,
                        None::<&str>,
                    )?;
                    exit(
                        match execute_unshared(app, &temp_dir, user_args, start_pipe_write) {
                            Ok(()) => 0,
                            Err(msg) => {
                                warn!("{}", msg);
                                1
                            }
                        },
                    )
                }
                ForkResult::Parent { child, .. } => {
                    close(start_pipe_write)?;
                    write(pipe_write, &child.as_raw().to_le_bytes())?;
                    match waitpid(child, None)? {
                        WaitStatus::Exited(_, status) => {
                            write(pipe_write, &status.to_le_bytes())?;
                        }
                        WaitStatus::Signaled(_, status, _) => {
                            write(pipe_write, &(status as i32).to_le_bytes())?;
                        }
                        _ => {
                            warn!("failed waitpid");
                            return Err(format_err!("waitpid: unusual return value"));
                        }
                    }
                    close(pipe_write)?;
                }
            }
            exit(0);
        }
        ForkResult::Parent { child, .. } => {
            close(pipe_write)?;
            close(start_pipe_write)?;
            let tl: f64 = app
                .value_of("timelimit")
                .unwrap_or("3600")
                .parse()
                .expect("--tl must be f64");
            let tl_msec = (tl * 1000.0) as u64;
            let mut buf = [0; size_of::<pid_t>()];
            let size = read(pipe_read, &mut buf[..])?;
            if size == 0 {
                return Err(format_err!("pipe broken: unshared may be failed"));
            }
            let inside: i32 = pid_t::from_le_bytes(buf);
            let tle = Arc::new(AtomicBool::new(false));
            let tle_clone = tle.clone();
            read(start_pipe_read, &mut [0])?;
            thread::spawn(move || {
                thread::sleep(Duration::from_millis(tl_msec + 200));
                match kill(Pid::from_raw(inside), SIGKILL) {
                    Ok(()) => {
                        tle_clone.store(true, Ordering::Relaxed);
                    }
                    Err(..) => {}
                }
            });
            let start = Instant::now();
            match waitpid(child, None)? {
                WaitStatus::Exited(_, status) => {
                    if status != 0 {
                        return Err(format_err!("execute failed {:?}", status));
                    }
                }
                _ => {
                    return Err(format_err!("waitpid: unusual return value!"));
                }
            }
            let time = start.elapsed().as_secs_f64();
            remove_dir_all(&temp_dir)?;
            let mut buf = [0; size_of::<i32>()];
            read(pipe_read, &mut buf[..])?;
            let mut result = ExecResult {
                status: 0,
                time: -1.0,
                memory: -1,
                tle: false,
            };
            result.status = i32::from_le_bytes(buf);
            result.time = time;
            if tl < result.time {
                tle.store(true, Ordering::Relaxed);
                result.time = tl;
            }
            let mem = read_to_string("/sys/fs/cgroup/memory/lib-judge/memory.max_usage_in_bytes")?;
            result.memory = mem.trim().parse()?;
            result.tle = tle.load(Ordering::Relaxed);
            Ok(result)
        }
    }
}

pub fn execute_main(my_args: &[String], user_args: &[String]) -> Result<ExecResult, Error> {
    let matches = App::new("executor")
        .arg(
            Arg::with_name("cwd")
                .long("cwd")
                .takes_value(true)
                .required(false),
        )
        .arg(Arg::with_name("overlay").long("overlay").required(false))
        .arg(
            Arg::with_name("timelimit")
                .long("tl")
                .takes_value(true)
                .required(false),
        )
        .arg(
            Arg::with_name("result")
                .long("result")
                .takes_value(true)
                .required(false),
        )
        .get_matches_from(my_args);

    let tl: f64 = matches
        .value_of("timelimit")
        .unwrap_or("3600")
        .parse()
        .expect("--tl must be f64");
    if !(0.0 <= tl && tl <= 3600.0) {
        warn!("invalid timelimit: {}", tl);
        panic!();
    }
    let result = execute(&matches, user_args)?;
    info!("result: {:?}", result);
    if let Some(result_file) = matches.value_of("result") {
        let mut file = File::create(result_file)?;
        writeln!(
            file,
            "{{\"returncode\": {}, \"time\": {}, \"memory\": {}, \"tle\": {}}}",
            result.status, result.time, result.memory, result.tle
        )?;
    }

    Ok(result)
}
