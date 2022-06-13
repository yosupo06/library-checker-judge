use std::io;

fn main() {
    let mut input = String::new();
    io::stdin().read_line(&mut input).unwrap();
    let vec : Vec<i32> = input.trim().split_whitespace().map(|x| x.parse().ok().unwrap()).collect();

    println!("{}", vec[0] + vec[1]);
}
