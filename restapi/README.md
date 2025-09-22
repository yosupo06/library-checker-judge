# REST API (OpenAPI) — ranking only

このディレクトリは、Library Checker の最小 REST API サーバーです。現状は Ranking API のみを実装しています（/ranking）。gRPC 本体とは別プロセスで動きます。

- デフォルトポート: `12381`（環境変数 `PORT` で変更可）
- エンドポイント:
  - `GET /health` — ヘルスチェック（"SERVING" を返す）
  - `GET /openapi.yaml` — OpenAPI 定義
  - `GET /ranking?skip&limit` — ランキング取得（JSON）
  - `GET /problems` — 問題一覧（name, title）
  - `GET /problems/{name}` — 問題詳細（title, source_url, time_limit, version, testcases_version, overall_version）

## 1) Docker Compose で動かす（おすすめ）

依存サービス（PostgreSQL, Cloud Storage エミュレータ, Firebase emulator など）と一緒に立ち上げます。OpenAPI からのコード生成も Docker イメージ内で自動実行されます。

```bash
# ルート: library-checker-judge/
docker compose up -d --build

# 状態確認
docker compose ps

# REST のヘルスチェック
curl http://localhost:12381/health
# => SERVING

# ランキング
curl "http://localhost:12381/ranking?skip=0&limit=100"

# 問題一覧 / 詳細
curl http://localhost:12381/problems
curl http://localhost:12381/problems/aplusb
```

個別に REST だけ起動したい場合（依存は自動解決）:

```bash
docker compose up -d --build db db-init gcs bootstrap-gcs api-rest
```

## 2) Go 単体でローカル実行（Docker なし）

OpenAPI のコード生成が必要です。生成後は普通に `go run` で起動できます。

### 前提
- Go 1.24+
- PostgreSQL が動いていること（DB 初期化が未実施なら後述のマイグレーションを実行）

### OpenAPI コード生成（補完を効かせたい人向け）
エディタで補完を効かせるには、生成コード（`internal/api/api.gen.go`）が手元に存在する必要があります。以下のいずれかを実行してください。

- シンプル: `make gen`

```bash
cd library-checker-judge/restapi
make gen   # = go generate ./... && go mod tidy
```

- 直接 `go generate` を使う:

```bash
cd library-checker-judge/restapi
go generate ./...
go mod tidy
```

### DB マイグレーション（初回のみ）
PostgreSQL にテーブルを作成します（環境変数で接続先を指定）。

- Docker で実行する場合（DB コンテナを使う）
```bash
cd library-checker-judge
# DB を先に起動
docker compose up -d db
# マイグレーションを実行
docker compose run --rm db-init
```

- Go で直接実行する場合（ホストの PostgreSQL を使う）
```bash
# 例: ローカルの Postgres に対して実行
cd library-checker-judge
PGHOST=localhost PGPORT=5432 PGDATABASE=librarychecker PGUSER=postgres PGPASSWORD=lcdummypassword \
  go run ./migrator
```

### REST サーバーの起動
```bash
cd library-checker-judge/restapi

# 例: ローカルの Postgres に接続して起動
PORT=12381 \
PGHOST=localhost PGPORT=5432 PGDATABASE=librarychecker PGUSER=postgres PGPASSWORD=lcdummypassword \
  go run .

# 動作確認
curl http://localhost:12381/health
curl "http://localhost:12381/ranking?skip=0&limit=100"
```

## フロントエンドから叩く
- フロントの環境変数 `VITE_REST_API_URL` に REST の URL（例: `http://localhost:12381`）を設定します。
- 例: `library-checker-frontend/.env.development` には既に `VITE_REST_API_URL=http://localhost:12381` が入っています。

## よくあるハマりどころ / トラブルシュート
- ビルド時に `missing go.sum entry for ... oapi-codegen ...` と出る
  - 上記「OpenAPI コード生成」後に `go mod tidy` を実行してください。
- DB 接続に失敗する
  - `PGHOST, PGPORT, PGDATABASE, PGUSER, PGPASSWORD` が正しいか確認してください。
  - Docker Compose を使っている場合は `PGHOST=db`（デフォルト）になります。
- 404 / `openapi.yaml` が返らない
  - `restapi/openapi/openapi.yaml` が存在するか確認してください。

## 実装のメモ
- 生成コードは `internal/api/api.gen.go` に出力します。
- ルーティングは `restapi.HandlerFromMux(&server{db}, r)` で登録しています（oapi-codegen v2）。
- DB は `gorm`（`*gorm.DB`）で接続しています。接続情報は環境変数から取得します。
