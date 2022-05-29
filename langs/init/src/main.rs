use std::env;
use std::ffi::{CString, CStr};
use std::process::exit;
use nix::sys::stat::Mode;
use nix::unistd::{dup2, execvp};
use nix::fcntl::{open, OFlag};

fn main() {
    let args: Vec<String> = env::args().collect();

    if args.len() <= 3 {
        println!("usage: init infile outfile command args");
        exit(1)
    }

    let stdin = open(args[1].as_str(), OFlag::O_RDONLY, Mode::empty()).expect("failed to open input file");
    let stdout = open(args[2].as_str(), OFlag::O_WRONLY | OFlag::O_CREAT, Mode::empty()).expect("failed to open output file");

    dup2(stdin, 0).expect("failed to dup2 stdin");
    dup2(stdout, 1).expect("failed to dup2 stdout");

    let new_args = args[3..].iter().map(|x| CString::new(x.clone()).unwrap()).collect::<Vec<CString>>();
    let new_args_slice = new_args.iter().map(|x| &x[..]).collect::<Vec<&CStr>>();
    execvp(&CString::new(args[3].clone()).unwrap(), &new_args_slice).expect("failed to execvp");
}
