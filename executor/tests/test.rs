use anyhow::Error;
use executor::{execute_main, tempdir, ExecResult};
#[cfg(feature = "sandbox")]
use rand::distributions::Alphanumeric;
#[cfg(feature = "sandbox")]
use rand::Rng;
use std::fs::File;

#[cfg(feature = "sandbox")]
use std::env;
#[cfg(feature = "sandbox")]
use std::iter;

use std::io::Write;
use std::path::PathBuf;

#[cfg(test)]
#[ctor::ctor]
fn init() {
    env_logger::Builder::from_default_env()
        .format_timestamp(None)
        .format_module_path(false)
        .init();
}

fn str2testdir(src: &str, name: &str) -> Result<PathBuf, Error> {
    let temp_dir = tempdir(&std::env::temp_dir())?;
    let file = File::create(temp_dir.join(name))?;
    writeln!(&file, "{}", src)?;
    Ok(temp_dir)
}

fn to_string_vec(s: Vec<&str>) -> Vec<String> {
    s.iter().map(|x| x.to_string()).collect()
}

fn assert_result(result: &ExecResult, status_expect: Option<i32>, time_upper: Option<f64>) {
    println!("result: {:?}", result);
    if let Some(status) = status_expect {
        assert_eq!(result.status, status);
    }
    if let Some(time) = time_upper {
        assert!(result.time <= time);
    }
}

fn compile_cpp_file(src: &str) -> Result<PathBuf, Error> {
    let dir = str2testdir(src, "source.cpp")?;

    let result = execute_main(
        &to_string_vec(vec!["executor", "--cwd", dir.to_str().unwrap()]),
        &to_string_vec(vec!["g++", "source.cpp"]),
    )?;
    assert_result(&result, Some(0), None);

    Ok(dir)
}

#[test]
fn hello_world_cpp() {
    let dir = compile_cpp_file(include_str!("../res/Hello.cpp")).expect("compile failed");

    let result = execute_main(
        &to_string_vec(vec!["executor", "--cwd", dir.to_str().unwrap()]),
        &to_string_vec(vec!["./a.out"]),
    )
    .expect("failed");

    assert_result(&result, Some(0), Some(0.05));
}

#[test]
fn test_unused_tle() {
    let dir = compile_cpp_file(include_str!("../res/Hello.cpp")).expect("compile failed");

    let result = execute_main(
        &to_string_vec(vec![
            "executor",
            "--cwd",
            dir.to_str().unwrap(),
            "--tl",
            "2.0",
        ]),
        &to_string_vec(vec!["./a.out"]),
    )
    .expect("failed");

    assert_result(&result, Some(0), Some(0.05));
}

#[test]
fn test_tle() {
    let dir = compile_cpp_file(include_str!("../res/TLE.cpp")).expect("compile failed");

    let result = execute_main(
        &to_string_vec(vec![
            "executor",
            "--cwd",
            dir.to_str().unwrap(),
            "--tl",
            "2.0",
        ]),
        &to_string_vec(vec!["./a.out"]),
    )
    .expect("failed");

    assert!(result.status != 0);
    assert!(result.tle);
}

#[cfg(feature = "sandbox")]
#[test]
fn test_overlay() {
    let temp = tempdir().expect("failed");
    let temp = temp.path();
    let result = execute_main(
        &to_string_vec(vec![
            "executor",
            "--cwd",
            temp.to_str().unwrap(),
            "--overlay",
        ]),
        &to_string_vec(vec!["touch", "test.txt"]),
    )
    .expect("failed");
    assert_result(&result, Some(0), None);
    assert!(!temp.join("test.txt").exists());
}

#[cfg(feature = "sandbox")]
#[test]
fn test_other_tmp() {
    let temp = tempdir().expect("failed");
    let temp = temp.path();
    let prev_tmp = env::var("TMPDIR");
    env::set_var("TMPDIR", temp);
    let result =
        execute_main(&to_string_vec(vec![]), &to_string_vec(vec!["mktemp"])).expect("failed");
    match prev_tmp {
        Ok(s) => env::set_var("TMPDIR", s),
        Err(..) => env::remove_var("TMPDIR"),
    }
    assert_result(&result, Some(0), None);
}

#[cfg(feature = "sandbox")]
#[test]
fn test_tempdir() {
    let temp = tempdir().expect("failed");
    let temp = temp.path();
    let name = format!("/tmp/{}", random_string(20));
    let result = execute_main(
        &to_string_vec(vec!["executor", "--cwd", temp.to_str().unwrap()]),
        &to_string_vec(vec!["touch", &name]),
    )
    .expect("failed");
    assert_result(&result, Some(0), None);
    assert!(!Path::new(&name).exists());
}

#[test]
fn test_re() {
    let temp = tempdir(&std::env::temp_dir()).expect("failed");
    let result = execute_main(
        &to_string_vec(vec!["executor", "--cwd", temp.to_str().unwrap()]),
        &to_string_vec(vec!["cat", "dummy.dummy"]),
    )
    .expect("failed");
    assert_result(&result, Some(1), Some(0.05));
}

#[test]
fn test_no_binary() {
    let temp = tempdir(&std::env::temp_dir()).expect("failed");
    let result = execute_main(
        &to_string_vec(vec!["executor", "--cwd", temp.to_str().unwrap()]),
        &to_string_vec(vec!["./dummy_binary"]),
    )
    .expect("failed");
    assert_result(&result, Some(1), Some(0.05));
}

#[cfg(feature = "sandbox")]
#[test]
fn test_stack_over_flow() {
    let temp = get_dir(Path::new("../test_src/stack.cpp")).expect("failed");
    let temp = temp.path();
    let result = execute_main(
        &to_string_vec(vec!["executor", "--cwd", temp.to_str().unwrap()]),
        &to_string_vec(vec!["g++", "stack.cpp"]),
    )
    .expect("failed");
    assert_result(&result, Some(0), None);
    assert!(temp.join("a.out").exists());
    let result = execute_main(
        &to_string_vec(vec!["executor", "--cwd", temp.to_str().unwrap()]),
        &to_string_vec(vec!["./a.out"]),
    )
    .expect("failed");
    assert_result(&result, Some(0), None);
}

#[cfg(feature = "sandbox")]
#[test]
fn test_fork_bomb() {
    let temp = get_dir(Path::new("../test_src/fork_bomb.sh")).expect("failed");
    let temp = temp.path();
    let result = execute_main(
        &to_string_vec(vec![
            "executor",
            "--cwd",
            temp.to_str().unwrap(),
            "--tl",
            "1.0",
        ]),
        &to_string_vec(vec!["./fork_bomb.sh"]),
    )
    .expect("failed");
    assert!(result.tle);
}
