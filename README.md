# Library Checker Judge

Judge server / API server のソースコードです

## Requirements

- Docker
- Ubuntu 22.04 (Judge Server)
- Go 1.24+

## API Server

### 起動

```sh
./launch_local.sh
```

APIサーバー(localhost:50051)とSQL(PostgreSQL)がDocker Composeで立ち上がり、`aplusb, unionfind`がデプロイされる。

### 動作確認

gRPC-web のAPIサーバーが起動します。

- gRPC API: localhost:50051
- gRPC-web API: localhost:12380
 - REST API (separate service): localhost:12381
 - REST (partial):
   - GET http://localhost:12380/api/langs
   - GET http://localhost:12380/api/problems

OpenAPI (partial, for REST) is defined at:

- restapi/openapi/openapi.yaml

```sh
evans --host localhost --port 50051 api/proto/library_checker.proto
```

## Judge Server

Judge serverはGoで書かれたAPIサーバーと通信するクライアント(`/judge`)と、実行環境(`/executor`)からなる。

### 準備

```sh
sudo apt install postgresql-client libpq-dev python3 python3-dev python3-pip g++ cgroup-tools libcap2-bin
pip3 install termcolor toml psycopg2 psutil
pip3 install -r deploy/requirements.txt
pip3 install -r ../library-checker-problems/requirements.txt
```

など

#### 実行環境の準備

Judge serverは各種プログラミング言語の実行環境が必要です。詳細は[langs/langs.toml](./langs/langs.toml)を参照してください。


### 起動

```
cd library-checker-judge/judge
go run .
```

## Local Test

- library-checker-problems / library-chcker-judge は同じディレクトリにcloneしておくこと

### 全体テスト（推奨）

全モジュールのテストを一括実行。PostgreSQL等の環境も自動で起動/停止します。

```sh
./test-all.sh
```

このスクリプトは以下を実行します：
- `./launch_local.sh` による環境起動（必要に応じて）
- Database, API, Storage モジュールのテスト
- 静的解析（go vet, gofmt）
- ビルド確認
- 環境のクリーンアップ（必要に応じて）

### 個別モジュールテスト

個別モジュールテストを実行する場合は、事前に `./launch_local.sh` でPostgreSQL等を起動しておく必要があります。

#### API Server のテスト

実行中のAPIサーバーに対してテストを実行します。

```sh
cd api
go test . -v
```

#### Database のテスト

```sh
cd database
go test . -v
```

#### Storage のテスト

```sh
cd storage
go test . -v
```

#### Judge Server のテスト

```sh
cd judge
go test . -v
```


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
