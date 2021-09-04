# Library Checker Judge

ジャッジ / API のソースコードです

## Requirements

- Ubuntu 18.04(Judge Server)
- docker, docker-compose(API, SQL)

## Launch API & SQL

```sh
./launch_local.sh
```

dockerグループに自分を登録することでsudoなしでdockerが使えるようになる [Reference](https://qiita.com/DQNEO/items/da5df074c48b012152ee)
sudoをつけて実行してもいいが、色んなところにrootでフォルダが作られて面倒な事になるので、非推奨

APIサーバー(localhost:50051)とSQL(Postgre SQL)がdocker-composeで立ち上がり、`aplusb, unionfind`がデプロイされる。

APIサーバーへは gRPC でアクセスします。例えばクライアントとして [evans](https://github.com/ktr0731/evans) を使うなら、以下のようにアクセス

```sh
evans --host localhost --port 50051 api/proto/library_checker.proto
evans --host apiv1.yosupo.com --port 443 library-checker-judge/api/proto/library_checker.proto -t
```

## Launch Judge Server

```sh
sudo apt install postgresql-client libpq-dev python3 python3-dev python3-pip g++ cgroup-tools libcap2-bin

pip3 install termcolor toml psycopg2 psutil

cargo test -- --test-threads=1 --nocapture
```

など

### cgroupでmemory swapを管理する

/etc/default/grubに以下を書き、reboot

```sh
GRUB_CMDLINE_LINUX="swapaccount=1"
```

- References: https://unix.stackexchange.com/questions/147158/how-to-enable-swap-accounting-for-memory-cgroup-in-archlinux


### ジャッジ用のシステムユーザーを作成する

```sh
sudo useradd library-checker-user -u 990 -r -s /sbin/nologin -M
```

ジャッジはpkill -u library-checker-user(このユーザーのプロセスを全部消す)を使用するため、UIDが他のユーザーと被ってはいけない。
特にpostgreコンテナはデフォルトで999をUIDとして使うため注意。

どちらかを変更すること

## Local Test

- library-checker-problems / library-chcker-judge は同じディレクトリにcloneしておくこと

APIのテスト(今のgo sourceではなく、今立ち上がってるAPIサーバーに対してテストすることに注意)

```sh
cd library-checker-judge/api
go test . -v
```

### Launch Judge

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
