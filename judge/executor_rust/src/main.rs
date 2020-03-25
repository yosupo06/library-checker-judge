use clap::{App, Arg};
use failure::{format_err, Error};
use libc::{rlimit, setrlimit, RLIMIT_STACK, RLIM_INFINITY};
use log::warn;
use nix::mount::{mount, MsFlags};
use nix::sched::{unshare, CloneFlags};
use nix::sys::wait::{waitpid, WaitStatus};
use nix::unistd::{chdir, chroot, execvp, fork, ForkResult};
use std::env;
use std::ffi::CString;
use std::fs::{create_dir, set_permissions};
use std::os::unix::fs::PermissionsExt;
use std::path::{Path, PathBuf};
use std::process::exit;
use tempfile::tempdir;
use std::io;
use std::iter;
use rand::Rng;
use rand::distributions::Alphanumeric;
use std::process::Command;

fn inside_tempdir(temp_dir: &Path) -> io::Result<PathBuf> {
    let mut rng = rand::thread_rng();    
    for _ in 0..10 {
        let chars: String = iter::repeat(())
                .map(|()| rng.sample(Alphanumeric))
                .take(10)
                .collect();
        let temp_dir = temp_dir.join(chars);
        if temp_dir.exists() {
            continue
        }
        create_dir(&temp_dir)?;
        return Ok(temp_dir);
    }
    Err(io::Error::new(io::ErrorKind::Other, "failed to create a temp dir"))
}

fn prepare_mount(temp_dir: &Path, overlay: bool) -> Result<(), Error> {
    let sand_dir = temp_dir.join("sand");
    create_dir(&sand_dir)?;

    if overlay {
        let work_dir = inside_tempdir(&temp_dir)?;
        let upper_dir = inside_tempdir(&temp_dir)?;
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
            Some("./"),
            &sand_dir,
            None::<&str>,
            MsFlags::MS_BIND,
            None::<&str>,
        )?;
    }

    let proc_dir = temp_dir.join("proc");
    // make /tmp
    create_dir(temp_dir.join("tmp"))?;
    set_permissions(&temp_dir.join("tmp"), PermissionsExt::from_mode(0o777))?;
    // mount -t proc proc /proc
    create_dir(temp_dir.join("proc"))?;
    mount(
        Some("proc"),
        &proc_dir,
        Some("proc"),
        MsFlags::empty(),
        None::<&str>,
    )?;
    // mount files
    for dir_name in vec![
        "dev", "sys", "bin", "lib", "lib64", "usr", "etc", "opt", "var", "home",
    ] {
        let dir = temp_dir.join(dir_name);
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

fn prepare_cgroup() -> Result<(), Error> {
    Command::new("cgdelete")
    .arg("pids,cpuset,memory:/lib-judge")
    .status()?;
    Ok(())
}

fn execute_child(sand_dir: &Path, user_args: &[String]) -> Result<(), Error> {
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

    chdir(sand_dir)?;
    chroot("..")?;
    let program = CString::new(user_args[0].clone())?;
    let args: Vec<std::ffi::CString> = user_args[..]
        .iter()
        .map(|s| CString::new(s.clone()).unwrap())
        .collect();
    let args: Vec<&std::ffi::CStr> = args.iter().map(|s| &s[..]).collect();
    execvp(&program, &args[..])?;

    Ok(())
}

fn execute_unshared(
    app: &clap::ArgMatches,
    temp_dir: &Path,
    user_args: &[String],
) -> Result<(), Error> {
    let overlay: bool = app.is_present("overlay");
    prepare_mount(&temp_dir, overlay)?;
    prepare_cgroup()?;
    match fork().expect("fork failed") {
        ForkResult::Child => execute_child(&temp_dir.join("sand"), user_args)?,
        ForkResult::Parent { child, .. } => match waitpid(child, None)? {
            WaitStatus::Exited(pid, status) => {
                println!("exit!: pid={:?}, status={:?}", pid, status);
            }
            WaitStatus::Signaled(pid, status, _) => {
                println!("signal!: pid={:?}, status={:?}", pid, status)
            }
            _ => {
                warn!("failed waitpid");
                return Err(format_err!("abnormal exit!"));
            }
        },
    }

    Ok(())
}

fn execute(app: &clap::ArgMatches, user_args: &[String]) -> Result<(), Error> {
    let temp_dir = tempdir()?;
    println!("{:?}", temp_dir);
    let temp_dir = temp_dir.path().to_owned();
    set_permissions(&temp_dir, PermissionsExt::from_mode(0o777))?;
    match fork().expect("fork failed") {
        ForkResult::Child => {
            unshare(CloneFlags::CLONE_NEWPID | CloneFlags::CLONE_NEWNS | CloneFlags::CLONE_NEWNET)?;
            mount(
                None::<&str>,
                &Some("/"),
                None::<&str>,
                MsFlags::MS_REC | MsFlags::MS_PRIVATE,
                None::<&str>,
            )?;
            execute_unshared(app, &temp_dir, user_args)?;
            exit(0);
        }
        ForkResult::Parent { child, .. } => match waitpid(child, None)? {
            WaitStatus::Exited(pid, status) => {
                println!("exit!: pid={:?}, status={:?}", pid, status)
            }
            WaitStatus::Signaled(pid, status, _) => {
                println!("signal!: pid={:?}, status={:?}", pid, status)
            }
            _ => {
                warn!("failed waitpid");
                return Err(format_err!("abnormal exit!"));
            }
        },
    }
    Ok(())
}

fn main() {
    let args: Vec<String> = env::args().collect();
    env::set_var("RUST_LOG", "warn");
    let split_pos = args
        .iter()
        .position(|s| s == "--")
        .expect("args must have --");
    let my_args = &args[..split_pos];
    let user_args = &args[split_pos + 1..];
    env_logger::Builder::from_default_env()
        .format_timestamp(None)
        .format_module_path(false)
        .init();
    let matches = App::new("executor")
        .arg(
            Arg::with_name("stdin")
                .long("stdin")
                .takes_value(true)
                .required(false),
        )
        .arg(
            Arg::with_name("stdout")
                .long("stdout")
                .takes_value(true)
                .required(false),
        )
        .arg(
            Arg::with_name("stderr")
                .long("stderr")
                .takes_value(true)
                .required(false),
        )
        .arg(Arg::with_name("overlay").long("overlay").required(false))
        .arg(
            Arg::with_name("result")
                .long("result")
                .takes_value(true)
                .required(false),
        )
        .arg(
            Arg::with_name("timelimit")
                .long("tl")
                .takes_value(true)
                .required(false),
        )
        .get_matches_from(my_args);
    println!("{:?} {:?} {:?}", my_args, user_args, matches);

    let tl: f64 = matches
        .value_of("timelimit")
        .unwrap_or("3600")
        .parse()
        .expect("--tl must be f64");
    if !(0.0 <= tl && tl <= 3600.0) {
        warn!("invalid timelimit: {}", tl);
        panic!();
    }

    execute(&matches, user_args).expect("execute failed");
}
