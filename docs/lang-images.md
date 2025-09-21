# Docker 言語イメージとテストの設計方針

本ドキュメントは library-checker-judge における「言語用 Docker イメージ」と、それを用いたテスト戦略（Judge / Langs / Executor / Integration）の役割分担と CI 方針をまとめます。CI 時間の短縮と責務の明確化が目的です。

## 役割分担（結論）
- Judge モジュール（最小構成のロジック検証）
  - C++（`gcc` イメージ）と Python（`python3` イメージ）に限定した A+B と主要判定（AC / WA / PE / TLE / RE / CE / Fail）で、判定結果・進捗更新・チェック連携など Judge のロジックを検証します。
  - 多言語の正当性は Judge では担保しません（= 速度最優先）。
  - CI では最小セットの言語イメージを `python3 ./build.py gcc python3`（`langs/` ディレクトリ）でビルドし、`go test -v .`（`judge/` ディレクトリ）を実行します。
- Langs モジュール（定義とビルドの中枢）
  - `langs/langs.toml` に言語 ID、コンパイル・実行コマンド、`image_name` を定義し、`langs/langs.go` で読み込みます。特殊言語（checker / verifier / generator）もここで定義します。
  - `langs/build.py` は各言語の Dockerfile をビルドするスクリプトで、サブセット指定（例: `python3 build.py gcc python3 rust`）や一覧表示（`--list`）に対応します。
  - `go test -v .`（`langs/` ディレクトリ）では TOML と Go コードの整合性、追加ファイルの存在確認などを行います。
- Executor モジュール（全言語の実行検証）
  - すべての言語で A+B を実際にコンパイル・実行し、各言語イメージの健全性（ツールチェーン、実行環境、追加ファイルの扱い等）を検証します。
  - `executor/sources/aplusb/*` の各言語サンプルを使用し、`go test -v -tags=langs_all ./executor` で実行します。`langs_all` ビルドタグが付いたテストは Langs / Executor の変更時のみ CI で走ります。
- Integration（最小構成の E2E）
  - 既定では C++ を用いた E2E を行います（必要に応じて Python を追加）。
  - ここでもビルドする言語イメージは最小セット（`gcc` + `python3`）です。

## 言語イメージの構成
- `langs/`
  - `langs.toml`: 言語 ID・コンパイル / 実行コマンド・`image_name` の定義。
  - `langs.go`: TOML を読み込んで `langs.LANGS` を構成。checker / verifier / generator などの特殊言語もここで扱います。
  - `build.py`: 言語イメージをビルドする Python スクリプト（サブセット指定、JSON レポート出力に対応）。
  - `Dockerfile.*`: 各言語イメージの Dockerfile。サフィックスは `build.py` 内のキーに対応します。
- 主なマッピング（例）
  - C++ 系（`cpp`, `cpp20`, `cpp17`, `cpp-func`）→ `library-checker-images-gcc`
  - Python 系（`python3`）→ `library-checker-images-python3`、（`pypy3`）→ `library-checker-images-pypy`
  - `rust`→`…-rust`, `java`→`…-java`, `go`→`…-golang`, `haskell`→`…-haskell`, `csharp`→`…-csharp`, `swift`→`…-swift` など
  - 特殊言語 `checker` / `verifier` / `generator` は `library-checker-images-gcc` を使用（最小セットでも gcc は必須）

## CI ワークフローとトリガー
- `test-judge.yml`
  - 対象: Judge / Executor / Langs / Storage など Judge 依存モジュールの変更。
  - 内容: Docker Compose を最小構成で起動し、`python3 langs/build.py gcc python3` で最小イメージをビルド後、`go test -v .`（`judge/` ディレクトリ）を実行します。
- `test-langs.yml`
  - 対象: Langs の定義や Dockerfile の変更。
  - 内容: `go test -v .`（`langs/` ディレクトリ）。PR 時は `python langs/build.py all --output-json …` で全イメージのサイズレポートも出力します。
- `test-executor.yml`
  - 対象: Executor や Langs の変更。
  - 内容: Docker Compose を起動し、`python3 langs/build.py`（全言語）でイメージを用意してから `go test -v -tags=langs_all ./executor` を実行し、各言語の A+B をコンパイル・実行します。
- `test-integration.yml`
  - 対象: Integration モジュールや依存モジュールの変更。
  - 内容: 最小イメージをビルドし、E2E の疎通を確認します。

この構成により、通常の開発では最小セットのみをビルド・テストし、言語や実行基盤を変更する場合にのみ全言語テストを走らせることで CI の負荷を抑えます。

## 言語イメージのビルド戦略
- 最小 / 全ビルドの切り替え
  - 既定は「最小セット固定」: `gcc`（C++ + checker / verifier / generator）+ `python3`。
  - Langs / Executor の変更時のみ「全ビルド」（CI では `test-executor.yml` が担当）。
- ビルドコマンド（`langs/` ディレクトリ）
  - 全言語: `python3 build.py`
  - 最小: `python3 build.py gcc python3`
  - 任意サブセット: `python3 build.py gcc python3 rust`
  - 利用可能なキーの一覧: `python3 build.py --list`
- Python が手元にない場合の代替（一時コンテナで実行）
  - `docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v "$PWD/langs":/src -w /src python:3.12-alpine sh -lc "apk add --no-cache docker-cli >/dev/null 2>&1 && python3 build.py gcc python3"`
- ローカル開発の起動スクリプト（`launch_local.sh`）
  - 既定で最小セット（`gcc python3`）をビルドしてから Compose を起動します。
  - 全言語やサブセットに切り替える場合は環境変数で指定: `LC_LANGS=all ./launch_local.sh`、`LC_LANGS="gcc python3 rust" ./launch_local.sh`
- キャッシュの活用
  - 現時点では未導入（buildx / GHCR 等）。必要になったら将来検討します。

## Executor 側の全言語 A+B テスト
- サンプルコード: `executor/sources/aplusb/*` に各言語向けの最小 A+B サンプルを保持します。
- テスト内容（言語ごと）
  - `executor.CompileSource` でコンパイル（`cpp-func` など追加ファイルが必要な言語には `grader.cpp` 等を渡す）。
  - `lang.Exec` で実行し、`sample.in` → `sample.out` の一致を検証。加えて簡単な追加入力（`123 456` → `579`）でも確認します。
  - 実行環境・ツールチェーンの健全性は Executor の責務として担保します。

## 新規言語の追加手順
1. `langs/Dockerfile.<LANG>` を追加し、ローカルで `docker build` 動作確認。
2. `langs/langs.toml` に ID / コンパイル / 実行 / `image_name` を登録し、必要に応じて `langs/langs.go` も更新。
3. `executor/sources/aplusb/` に最小 A+B のサンプルソースを追加。
4. PR では Langs / Executor のテストを通し（`go test -v .`（`langs/` ディレクトリ）および `go test -v -tags=langs_all ./executor`）、問題ないことを確認。

## 注意点 / ベストプラクティス
- `checker` / `verifier` / `generator` が `gcc` に依存するため、最小セットでも `gcc` は必須です。
- 重量級イメージ（Haskell / Swift 等）は Langs / Executor の変更時にのみビルド・テストする方針で CI 時間を抑制します。
- Docker 画像タグは `Lang.ImageName` と一致させ、変更時は TOML も合わせて更新します。
