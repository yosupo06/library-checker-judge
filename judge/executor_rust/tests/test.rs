use executor_rust::{execute_main, ExecResult};
use anyhow::Error;
use rand::distributions::Alphanumeric;
use rand::Rng;
use std::fs::copy;
use std::fs::set_permissions;
use std::iter;
use std::env;
use std::os::unix::fs::PermissionsExt;
use std::path::Path;
use tempfile::{tempdir, TempDir};

fn random_string(len: usize) -> String {
    let mut rng = rand::thread_rng();
    iter::repeat(())
        .map(|()| rng.sample(Alphanumeric))
        .take(len)
        .collect()
}

fn get_dir(src: &Path) -> Result<TempDir, Error> {
    let dir = tempdir()?;
    copy(src, dir.path().join(src.file_name().unwrap()))?;
    set_permissions(&dir.path(), PermissionsExt::from_mode(0o777))?;
    println!("tempdir: {:?}", dir);
    Ok(dir)
}

fn to_string_vec(s: Vec<&str>) -> Vec<String> {
    s.iter().map(|x| x.to_string()).collect()
}

fn assert_result(result: &ExecResult, status_expect: Option<i32>, time_upper: Option<f64>) {
    println!("result: {:?}", result);
    if let Some(status) = status_expect {
        assert!(result.status == status);
    }
    if let Some(time) = time_upper {
        assert!(result.time <= time);
    }
}

#[test]
fn hello_world_cpp() {
    let temp = get_dir(Path::new("../test_src/Hello.cpp")).expect("failed");
    let temp = temp.path();
    let result = execute_main(
        &to_string_vec(vec!["executor", "--cwd", temp.to_str().unwrap()]),
        &to_string_vec(vec!["g++", "Hello.cpp"]),
    )
    .expect("failed");
    assert_result(&result, Some(0), None);

    assert!(temp.join("a.out").exists());

    let result = execute_main(
        &to_string_vec(vec!["executor", "--cwd", temp.to_str().unwrap()]),
        &to_string_vec(vec!["./a.out"]),
    )
    .expect("failed");
    assert_result(&result, Some(0), Some(0.05));
}

#[test]
fn hello_world_cpp_flag() {
    let temp = get_dir(Path::new("../test_src/Hello.cpp")).expect("failed");
    let temp = temp.path();
    let result = execute_main(
        &to_string_vec(vec!["executor", "--cwd", temp.to_str().unwrap()]),
        &to_string_vec(vec!["g++", "Hello.cpp", "-o", "Hello"]),
    )
    .expect("failed");
    assert_result(&result, Some(0), None);

    assert!(temp.join("Hello").exists());

    let result = execute_main(
        &to_string_vec(vec!["executor", "--cwd", temp.to_str().unwrap()]),
        &to_string_vec(vec!["./Hello"]),
    )
    .expect("failed");
    assert_result(&result, Some(0), Some(0.05));
}

#[test]
fn test_unused_tle() {
    let temp = get_dir(Path::new("../test_src/Hello.cpp")).expect("failed");
    let temp = temp.path();
    let result = execute_main(
        &to_string_vec(vec!["executor", "--cwd", temp.to_str().unwrap()]),
        &to_string_vec(vec!["g++", "Hello.cpp"]),
    )
    .expect("failed");
    assert_result(&result, Some(0), None);

    assert!(temp.join("a.out").exists());

    let result = execute_main(
        &to_string_vec(vec![
            "executor",
            "--cwd",
            temp.to_str().unwrap(),
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
    let temp = get_dir(Path::new("../test_src/TLE.cpp")).expect("failed");
    let temp = temp.path();
    let result = execute_main(
        &to_string_vec(vec!["executor", "--cwd", temp.to_str().unwrap()]),
        &to_string_vec(vec!["g++", "TLE.cpp"]),
    )
    .expect("failed");
    assert_result(&result, Some(0), None);
    assert!(temp.join("a.out").exists());
    let result = execute_main(
        &to_string_vec(vec![
            "executor",
            "--cwd",
            temp.to_str().unwrap(),
            "--tl",
            "2.0",
        ]),
        &to_string_vec(vec!["./a.out"]),
    )
    .expect("failed");
    assert!(result.status != 0);
    assert!(result.tle);
}

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

#[test]
fn test_other_tmp() {
    let temp = tempdir().expect("failed");
    let temp = temp.path();
    let prev_tmp = env::var("TMPDIR");
    env::set_var("TMPDIR", temp);
    let result = execute_main(
        &to_string_vec(vec![]),
        &to_string_vec(vec!["mktemp"]),
    ).expect("failed");
    match prev_tmp {
        Ok(s) => env::set_var("TMPDIR", s),
        Err(..) => env::remove_var("TMPDIR"),
    }
    assert_result(&result, Some(0), None);
}

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
    let temp = tempdir().expect("failed");
    let temp = temp.path();
    let result = execute_main(
        &to_string_vec(vec!["executor", "--cwd", temp.to_str().unwrap()]),
        &to_string_vec(vec!["cat", "dummy.dummy"]),
    )
    .expect("failed");
    assert_result(&result, Some(1), Some(0.05));
}

#[test]
fn test_no_binary() {
    let temp = tempdir().expect("failed");
    let temp = temp.path();
    let result = execute_main(
        &to_string_vec(vec!["executor", "--cwd", temp.to_str().unwrap()]),
        &to_string_vec(vec!["./dummy_binary"]),
    )
    .expect("failed");
    assert_result(&result, Some(1), Some(0.05));
}

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
    println!("{:?}", result);
    assert!(result.tle);
}
