use executor::execute_main;
use std::env;

fn main() {
    env_logger::Builder::from_default_env()
        .format_timestamp(None)
        .format_module_path(false)
        .init();
    let args: Vec<String> = env::args().collect();
    let split_pos = args
        .iter()
        .position(|s| s == "--")
        .expect(&format!("args must have --, args={:?}", args));
    let my_args = &args[..split_pos];
    let user_args = &args[split_pos + 1..];
    execute_main(my_args, user_args).expect("execute failed");
}
