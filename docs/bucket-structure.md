# バケット構成とファイル

- 目的: Public / Private バケットの構成と、各オブジェクトキーの仕様を整理する。

**バケット**
- Private バケット: `STORAGE_PRIVATE_BUCKET`（デフォルト `testcase`）— テストケースの tarball（非公開オブジェクト）を格納。
- Public バケット: `STORAGE_PUBLIC_BUCKET`（デフォルト `testcase-public`）— 公開ファイルと例題 I/O を格納。

**オブジェクトキー構造（Object Keys）**
- 現行（v3 運用 — 利用側の参照は当面こちら）
  - Private（テストケース tarball）: `v3/{problem}/testcase/{testcase_hash}.tar.gz`
  - Public（例題 I/O）:
    - 入力: `v3/{problem}/testcase/{testcase_hash}/in/example_*.in`
    - 出力: `v3/{problem}/testcase/{testcase_hash}/out/example_*.out`
  - Public（公開ファイル — 問題バージョン依存）: `v3/{problem}/files/{version}/{path}`

- 新スキーマ（v4 仕様 — Phase 1 で並列に全て用意）
  - Private（テストケース tarball）: `v4/testcase/{problem}/{testcase_hash}.tar.gz`
  - Public（例題 I/O）:
    - 入力: `v4/examples/{problem}/{testcase_hash}/in/example_*.in`
    - 出力: `v4/examples/{problem}/{testcase_hash}/out/example_*.out`
  - Public（公開ファイル — 問題バージョン依存; v4 は `Version -> OverallVersion` を用いる）:
    - `v4/files/{problem}/{overall_version}/common/...` — リポジトリルート `common/` の中身
    - `v4/files/{problem}/{overall_version}/{problem_name}/...` — 問題ディレクトリ全体

**例（problem: `aplusb`, version: `v`, testcase hash: `h`）**
- v3（現行参照）
  - Private tarball: `v3/aplusb/testcase/h.tar.gz`
  - 例題入力: `v3/aplusb/testcase/h/in/example_00.in`
  - 例題出力: `v3/aplusb/testcase/h/out/example_00.out`
  - 公開ファイル: `v3/aplusb/files/v/task.md` 等
- v4（Phase 1 で並列に作成）
  - Private tarball: `v4/testcase/aplusb/h.tar.gz`
  - 例題入力: `v4/examples/aplusb/h/in/example_00.in`
  - 例題出力: `v4/examples/aplusb/h/out/example_00.out`
  - 公開ファイル（OverallVersion = ov）:
    - `v4/files/aplusb/ov/common/fastio.h`
    - `v4/files/aplusb/ov/aplusb/task.md`

**Private バケット**
- 非公開。ジャッジサーバーのみがアクセス。
- ジャッジは提出の評価時に tarball をダウンロード・展開し、ローカルにキャッシュして再ダウンロードを回避する。

**Public バケット**
- 公開。誰でも参照可能。
- Phase 1 中の運用: 参照は引き続き v3 を利用しつつ、v4 の全構造（examples/files）を並行で作成・維持する。

**バージョンとハッシュの定義**
- テストケースバージョン（`TestCaseVersion`）:
  - 問題ディレクトリ内の `hash.json` を集約（各テストケースファイルのハッシュを結合）して算出。
  - 提出の再ジャッジ判定に利用（保守的）。
- 問題バージョン（v3: `Version` / v4: `OverallVersion`）:
  - 以下の内容をハッシュ化して算出。
    - `TestCaseVersion`
    - v3: 公開ファイル集合の固定リスト（互換）
    - v4: Git で追跡されている `common/` と `{problem_name}/` 配下の全ファイルの内容（untracked は除外）
  - 軽量な公開ファイル群のみの再アップロードで済むため、更新運用は容易。

**アップロードのタイミング**
- `TestCaseVersion` または `Version / OverallVersion` が変化した場合のみアップロード（高速化のため）。
- 変化検知はデータベースに保存されたバージョンとの比較で行う。
- Phase 1 実施内容（v3 運用は維持しつつ v4 にも並列アップロード）:
  - Private: `v3/{problem}/testcase/{hash}.tar.gz` と `v4/testcase/{problem}/{hash}.tar.gz` の両方にアップロード。
  - Public（例題 I/O）: `v3/{problem}/testcase/{hash}/{in,out}/...` と `v4/examples/{problem}/{hash}/{in,out}/...` の両方にアップロード。
  - Public（公開ファイル）: `v3/{problem}/files/{version}/...`（従来の Version）と `v4/files/{problem}/{overall_version}/...`（新 OverallVersion）の両方にアップロード。

**将来（最終 v4 仕様の方向性）**
- 例題 I/O: `v4/examples/{problem}/{testcase_hash}/{in,out}/example_*`
- 公開ファイル: `v4/files/{problem}/{overall_version}/common/...` および `v4/files/{problem}/{overall_version}/{problem_name}/...`

**関連コード**
- アップローダー本体: `uploader/main.go`
- ストレージ（キー生成・アップロード処理）: `storage/*`
  - `storage/problem.go` — キー生成（v4 仕様へ更新予定）
  - `storage/upload.go` — 公開ファイルの収集とアップロード（`common/` + 問題ディレクトリ全体）
  - `storage/download.go` — ジャッジ側のダウンロード処理（v4 仕様へ更新予定）
- 問題リポジトリ: `library-checker-problems`（例: `sample/aplusb`）

## 移行プラン（v3 → v4）

目的
- Phase 1 で v4 の全構造（private tarball / examples / files）を一気に整備し、以後も v3 と並列でアップロード維持。
- 利用側（Judge/Frontend）は当面 v3 を参照し、十分な検証後に v4 へ切替。
- DB では v3 の `Version` は従来どおり保持し、v4 用に新たに `OverallVersion` を別フィールドで保持。

フェーズ概要
1) Phase 1（並列化・完全 v4 構築）
   - 目的: v3 のアップロードを維持しつつ、同時に v4（testcase/examples/files）にも全て書き込む「デュアルライティング」を実現。
   - アップローダ変更:
     - バージョン計算を二系統に分離: `Version`（現行の対象）と `OverallVersion`（`common/` + `{problem_name}/` 全ファイル）。
     - Private: 生成した tarball を v3 と v4 の両方へアップロード。
     - Public 例題: v3 と v4（`v4/examples/...`）の両方へアップロード（再生成コストが高い場合は v3→v4 のサーバーサイドコピーでも可）。
     - Public 公開ファイル: v3（`files/{Version}`）と v4（`files/{OverallVersion}`）の両方へアップロード。
   - DB 変更:
     - `problems` テーブルに `overall_version`（文字列）を追加（NULL/空文字許容）。
     - 当面 `testcases_version` は共通（v3/v4 で同一）。
   - 運用:
     - 夜間に `--force` で全問題を再生成・再アップロード（Private v4 tarball を作るため）。
     - 以後は通常の差分アップロード時に v3/v4 並行更新。

2) Phase 2（参照側の切替準備とフォールバック）
   - Judge/Frontend に v4 参照機能を追加（初期は OFF）。
   - 切替時は v4 を優先参照、v4 不在時のみ v3 へフォールバック（短期間）。

3) Phase 3（クリーンアップ）
   - 安定後に v3 Public を段階整理。Private v3 tarball は運用方針に応じて整理。

実装メモ/注意
- バージョン定義:
  - `Version`: 現行の選定ルール（従来の固定ファイル群）。
  - `OverallVersion`: `common/` 配下と `{problem_name}/` 配下の全ファイルをハッシュ化して算出。
- サーバーサイドコピー（例題 I/O の使い回し向け）:
  - GCS: `gsutil -m cp -r gs://<bucket>/v3/{problem}/testcase/{hash}/ gs://<bucket>/v4/examples/{problem}/{hash}/`
- 冪等性: 上書き許容。存在チェックでスキップしても良い。
- キャッシュ: 切替直前に v4 へ全ファイル反映→必要なら `v4/meta/{problem}.json` を最後に更新して原子的に切替（任意）。
- アクセス制御: Public バケットのポリシーは v4 パスにも適用されることを事前確認。
