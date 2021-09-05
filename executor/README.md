# library-checker-executor

ジャッジサーバーが提出コードを実行するためのコンテナです。

## How to Install

```sh
cargo build
```

もしくは

```sh
cargo build --features sandbox
```

`--features sandbox`をつけない場合、最低限の機能しかつかず、例えばfork爆弾などを防げなくなるので注意。
sandboxをつけて動かすためには次の準備を行い、かつrootで実行する必要がある。正確には[build.pkr.hcl]('../packer/build.pkr.hck')を参照。

### ジャッジ用のシステムユーザーを作成する

```sh
sudo useradd library-checker-user -u 2000 -m
```

### cgroupでswap(memory)を管理できるようにする

`/etc/default/grub` に以下を書き、rebootする

```sh
GRUB_CMDLINE_LINUX="swapaccount=1"
```

- References: https://unix.stackexchange.com/questions/147158/how-to-enable-swap-accounting-for-memory-cgroup-in-archlinux

## How to Use

```sh
$ cargo run -- -- echo a
a
$ cargo run -- --tl 2.0 --result a.txt -- sleep 1 && cat a.txt
{"returncode": 0, "time": 1.0006987, "memory": 0, "tle": false}
$ cargo run -- --tl 2.0 --result a.txt -- sleep 10 && cat a.txt
{"returncode": 9, "time": 2, "memory": 0, "tle": true}
```

## Test

```sh
cargo test
cargo test --features sandbox
```
