# Library Checker Judge

ジャッジ / API のソースコードです

## Requirements

- Ubuntu 18.04(Judge Server)
- docker, docker-compose(API, SQL)

## API Server

### 準備

dockerグループに自分を登録することでsudoなしでdockerが使えるようになる [Reference](https://qiita.com/DQNEO/items/da5df074c48b012152ee)
sudoをつけて実行してもいいが、色んなところにrootでフォルダが作られて面倒な事になるので、非推奨

```sh
sudo gpasswd -a $USER docker
sudo systemctl restart docker
```

APIサーバー(localhost:50051)とSQL(Postgre SQL)がdocker-composeで立ち上がり、`aplusb, unionfind`がデプロイされる。

APIサーバーへは gRPC でアクセスします。例えばクライアントとして [evans](https://github.com/ktr0731/evans) を使うなら、以下のようにアクセス

### 起動

```sh
./launch_local.sh
```

`launch_local.sh` は default だと `aplusb` しかデプロイしないので、必要ならば `deploy/problems_deploy.py` も叩くとよい。

### 動作確認

通常の gRPC 版と gRPC-Web 版のふたつが建つ

```sh
evans --host localhost --port 50051 api/proto/library_checker.proto
evans --host localhost --port 58080 api/proto/library_checker.proto --web
```

## Judge Server

### 準備

```sh
sudo apt install postgresql-client libpq-dev python3 python3-dev python3-pip g++ cgroup-tools libcap2-bin
pip3 install termcolor toml psycopg2 psutil
pip3 -r install deploy/requirements.txt
pip3 -r ../library-checker-problems/requirements.txt
```

など

#### cgroupでmemory swapを管理する

/etc/default/grubに以下を書き、reboot

```sh
GRUB_CMDLINE_LINUX="swapaccount=1"
```

- References: https://unix.stackexchange.com/questions/147158/how-to-enable-swap-accounting-for-memory-cgroup-in-archlinux


#### ジャッジ用のシステムユーザーを作成する

```sh
sudo useradd library-checker-user -u 990 -r -s /sbin/nologin -M
```

ジャッジはpkill -u library-checker-user(このユーザーのプロセスを全部消す)を使用するため、UIDが他のユーザーと被ってはいけない。
特にpostgreコンテナはデフォルトで999をUIDとして使うため注意。

どちらかを変更すること

#### 実行環境を作る

設定情報が書かれたファイル `judge/secret.toml` を作る。

```
cd library-checker-judge/judge
./make_secret.sh
```

`executor_rust` をビルドして PATH の通ったところに置く。

```
cd library-checker-judge/judge/executor_rust
cargo build --release
cp target/release/executor_rust path/to/...
```

ユーザの提出を実行するための処理系をインストールする。
[api/langs.toml](https://github.com/yosupo06/library-checker-judge/blob/master/api/langs.toml) を見ながら適当にする。

```
sudo apt install g++ clang++ python3.8 pypy3 openjdk-11-jdk haskell-stack sbcl ...
```

### 起動

```
cd library-checker-judge/judge
go run .
```

## Local Test

- library-checker-problems / library-chcker-judge は同じディレクトリにcloneしておくこと

### API Server のテスト

今のgo sourceではなく、今立ち上がってるAPIサーバーに対してテストすることに注意

```sh
cd library-checker-judge/api
go test . -v
```

### Judge Server のテスト

```sh
cd library-checker-judge/judge
sudo go run *.go
```

各種機能をガンガン使うのでrootじゃないと動かない　多分

### Build Judge Image for GCP

```
gcloud auth application-default login
cd packer
packer build .
```

Library Checkerで稼働するジャッジサーバーのイメージは[packer](https://www.packer.io/)でビルドされている。


## Contribution

なんでも歓迎

## library-checker-project

- problems: [library-checker-problems](https://github.com/yosupo06/library-checker-problems)
- judge: [library-checker-judge](https://github.com/yosupo06/library-checker-judge)
- frontend: [library-checker-frontend](https://github.com/yosupo06/library-checker-frontend)
