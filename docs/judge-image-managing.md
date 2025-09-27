# ジャッジイメージ管理

## 概要
- `library-checker-judge` リポジトリ内で GitHub Actions と Packer を用いてジャッジ用 VM イメージを管理しています。
- `judge-image-build.yml` はベースイメージの更新（任意）と、環境ごとのジャッジイメージ生成を常に担当します。
- ワークフローは Terraform の出力に依存しており、Google Cloud の認証、イメージファミリー、ストレージ/データベース設定を決定します。
- イメージは `${ENV}-library-checker-project` GCP プロジェクトに保存され、ベース層は `v3-${ENV}-base-image`、ジャッジランタイムは Terraform 出力で指定されるファミリーを利用します。
- 古いジャッジイメージは `judge-image-prune.yml` が `scripts/prune_gce_images.py` を呼び出して最近のものを保持しつつ古いものを削除します。

## judge-image-build ワークフロー
- トリガーモード: `workflow_call` による再利用（`env` 入力が必須）と、`env`（`dev` または `prod`）および `build-base` フラグを受け取る手動の `workflow_dispatch`。
- 並行実行制御は環境ごと（`${env}-judge-image-build`）に設定され、ビルドの重複を避けます。
- グローバルな環境変数で Terraform Cloud の組織（`yosupo06-org`）とワークスペース（`${env}-library-checker`）を選択します。

### Base ジョブ
- `inputs.build-base` が true のとき（デフォルトで有効）に実行され、後続のジャッジビルドが利用するベースイメージを準備します。
- 手順:
  1. リポジトリをチェックアウトし、Terraform を初期化して `TF_API_TOKEN` で outputs を取得します。
  2. Terraform の出力（`gh_provider_id`, `judge_deployer_sa_email`）を用いて Workload Identity で Google Cloud に認証します。
  3. `packer/base/` で `packer init` を実行し、タイムスタンプ付きの一時イメージ名（`v3-tmp-base-image-<timestamp>`）を作成します。
  4. `packer build` で `ubuntu-2204-lts` を元に VM を構築し、Docker、Python ツール群、Cloud Ops Agent、各言語ランタイムを `/tmp/langs` にインストールします。
     - `python3 /tmp/langs/build.py` で全ランタイムのコンテナイメージをビルドし、不要な Docker アーティファクトを削除して後段で再利用する Docker ベースディレクトリを準備します。
     - Systemd ユニット `prepare-docker.service` と補助スクリプトをコピーし、ランタイムノードがベイク済みイメージから Docker レイヤーを展開できるようにします。
  5. Packer の処理完了後、`gcloud compute images create` で `v3-${env}-base-image-<timestamp>` という永続的なベースイメージを共有ファミリー `v3-${env}-base-image` に作成します。
- このジョブには現在 `# TODO: test` コメントが付いており、公開前に自動検証は実施されていません。

### Judge ジョブ
- Base ジョブに依存しており、Base ステージが成功するか、キャンセルされずにスキップされた場合のみ実行されます。
- 手順:
  1. Terraform の初期化と Google Cloud 認証を再実行し、環境固有のシークレットを取得します。
  2. Secret Manager から MinIO の HMAC シークレットを取得します（`get-secretmanager-secrets`）。
  3. リポジトリルートで `./run_protoc.sh` を実行して gRPC スタブを再生成します（Go ビルドに必要）。
  4. `./judge` ディレクトリ内で `go build .` を実行し、後でプロビジョニングするための `packer/judge/../../judge/judge` バイナリを生成します。
  5. `packer/judge/` で `packer init` と `packer build` を実行し、以下のパラメータを渡します:
     - `env`: GCP プロジェクト（`${env}-library-checker-project`）およびベースイメージファミリー（`v3-${env}-base-image`）を選択します。
     - `image_family`: Terraform で提供されるジャッジイメージファミリー（例: `v3-judge-image`）。
     - ストレージ資格情報（`minio_host`, `minio_id`, `minio_secret`, `minio_bucket`, `minio_public_bucket`）。
     - データベース接続情報（`db_connection_name`, `pg_user`）。
  6. Packer は最新のベースイメージファミリーを基に VM をプロビジョニングし、cloud-init の完了を待ち、Cloud SQL Proxy を配置し、systemd ユニットをインストールします。
     - `cloudsql.service.pkrtpl` が対象の接続名に対してプライベート IP と IAM 認証で `cloud-sql-proxy` を起動します。
     - コンパイル済みジャッジバイナリを `/root/judge` に配置し、`judge.service.pkrtpl` が MinIO と PostgreSQL の環境変数を設定してサービスを有効化します。
- イメージは設定されたファミリー（`image_family`）にタイムスタンプ付きの名前で直接出力されます。

## Packer 設定概要

### `packer/base/build.pkr.hcl`
- 標準の Ubuntu 22.04 LTS イメージから開始し、システムの初期設定を行います。
- Docker と補助コンポーネント（`crun`、カスタム daemon 設定、プリビルド済み言語イメージ）をインストールします。
- ランタイムノードが Docker 状態を復元できるように systemd ヘルパー（`prepare-docker.service`, `/root/prepare-docker.sh`）を準備します。
- Docker データを `/var/lib/docker-base` に配置し、後続のジャッジ層が引き継げるようにします。

### `packer/judge/build.pkr.hcl`
- 最新のベースイメージファミリー（`v3-${env}-base-image`）を利用します。
- systemd ユニットのテンプレート化を通じてランタイムシークレットを埋め込み、平文の機密情報が systemd 設定の外に残らないようにします。
- `judge.service` が Docker と Cloud SQL Proxy の両方に依存するよう構成し、ジャッジプロセスに必要な環境変数を登録します。

## イメージのクリーンアップ
- `judge-image-prune.yml` は UTC 18:00 に毎日実行され（手動トリガーも可能）、ジャッジイメージを整理します。
- ワークフローはビルドジョブと同じ認証手順を踏み、その後 `scripts/prune_gce_images.py` を次のパラメータで呼び出します:
  - `--family v3-judge-image`
  - `--keep 3`
  - `--min-age-days 14`
  - ワークフロー入力の `--dry-run` フラグ（任意）
- スクリプトは `creationTimestamp` 順にイメージを並べ、新しい `keep` 件を残し、最小経過日数の条件を満たした古いイメージを削除します。

## 運用メモ
- `judge-image-build` を手動で起動する際はベースイメージの変更有無を確認し、最新の `v3-${env}-base-image` を再利用したい場合は `build-base` をスキップして実行時間を短縮してください。
- ワークフローを動かす前に Terraform のステート/ワークスペースの出力が最新であることを確認してください。これらはサービスアカウント、イメージファミリー名、ストレージ/データベース設定を提供します。
- 新しいベースイメージを公開すると同じワークフロー内のジャッジジョブが自動的にそれを取り込みます。Packer のソースがファミリーの最新イメージを参照するためです。
- クリーンアップはジャッジイメージファミリーのみに適用されます。ベースイメージは GCE ファミリールールに従って蓄積されるため、不要な増加を避けるために Base ジョブの実行は必要な場合に限定してください。
