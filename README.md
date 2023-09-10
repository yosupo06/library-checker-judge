# Library Checker Judge

Judge server / API server のソースコードです

## Requirements

- docker
- Ubuntu 22.04(Judge Server)


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

grpc-web のAPIサーバーが建つ

```sh
evans --host localhost --port 18080 api/proto/library_checker.proto --web
```

## Judge Server

Judge serverはgoで書かれたAPIサーバーと通信するクライアント(`/judge`)と、このクライアントが呼び出す軽量コンテナ(`/executor`)からなる。

### 準備

```sh
sudo apt install postgresql-client libpq-dev python3 python3-dev python3-pip g++ cgroup-tools libcap2-bin
pip3 install termcolor toml psycopg2 psutil
pip3 install -r deploy/requirements.txt
pip3 install -r ../library-checker-problems/requirements.txt
```

など

#### executorをinstallする

executorの[README](./executor/README.md)を参照。Ubuntu以外で動作確認をしていない、かつ色々準備が必要なので注意。
`executor` をビルドして PATH の通ったところに置く。

```sh
cd library-checker-judge/executor
cargo install --path . --features sandbox
# or: cargo build --release --features sandbox && cp target/release/executor_rust path/to/...
```

#### 実行環境を作る

設定情報が書かれたファイル `judge/secret.toml` を作る。

```
cd library-checker-judge/judge
./make_secret.sh
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
