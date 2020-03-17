use clap::{App, Arg};
use failure::{format_err, Error};
use libc::{rlimit, setrlimit, RLIMIT_STACK, RLIM_INFINITY};
use log::warn;
use nix::NixPath;
use nix::mount::{mount, MsFlags};
use nix::sched::{unshare, CloneFlags};
use nix::sys::wait::{waitpid, WaitStatus};
use nix::unistd::{chdir, chroot, execvp, fork, ForkResult};
use std::env;
use std::ffi::CString;
use std::fs::{create_dir, set_permissions};
use std::os::unix::fs::PermissionsExt;
use std::path::Path;
use tempfile::{tempdir, TempDir};

fn prepare_mount(temp_dir: &Path, overlay: bool) -> Result<(), Error> {
    let sand_dir = temp_dir.join("sand");
    create_dir(&sand_dir)?;

    mount(Some("./"), &sand_dir, Some(""), MsFlags::MS_BIND, Some(""));
    // if overlay {
    //     let work_dir = tempdir()?;
    //     let work_dir = work_dir.path().to_owned();
    //     let upper_dir = tempdir()?;
    //     let upper_dir = upper_dir.path().to_owned();
    //     set_permissions(work_dir, PermissionsExt::from_mode(0o777))?;
    //     set_permissions(upper_dir, PermissionsExt::from_mode(0o777))?;
    //     mount()
    // } else {

    // }

    let proc_dir = temp_dir.join("proc");
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
    //chroot(".")?;
    warn!("UNKO");
    let program = CString::new(user_args[0].clone())?;
    let args: Vec<std::ffi::CString> = user_args[..]
        .iter()
        .map(|s| CString::new(s.clone()).unwrap())
        .collect();
    let args: Vec<&std::ffi::CStr> = args.iter().map(|s| &s[..]).collect();
    execvp(&program, &args[..])?;

    Ok(())
}

fn execute_unshared(app: &clap::ArgMatches, user_args: &[String]) -> Result<(), Error> {
    let temp_dir = tempdir()?;
    println!("{:?}", temp_dir);
    let temp_dir = temp_dir.path().to_owned();
    set_permissions(&temp_dir, PermissionsExt::from_mode(0o777))?;
    prepare_mount(&temp_dir, false)?;

    match fork().expect("fork failed") {
        ForkResult::Child => {
            execute_child(&temp_dir.join("sand"), user_args)?
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

fn execute(app: &clap::ArgMatches, user_args: &[String]) -> Result<(), Error> {
    unshare(CloneFlags::CLONE_NEWPID | CloneFlags::CLONE_NEWNS | CloneFlags::CLONE_NEWNET)?;
    execute_unshared(app, user_args)
}

fn main() {
    let args: Vec<String> = env::args().collect();
    env::set_var("RUST_LOG", "warn");
    println!("{:?}", args);
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
