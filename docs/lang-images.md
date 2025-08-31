# Docker 言語イメージとテストの設計方針

本ドキュメントは library-checker-judge における「言語用 Docker イメージ」と、それを用いるテスト戦略（Judge/Executor/Integration）の役割分担と CI 方針をまとめます。CI 時間の短縮と責務の明確化が目的です。

## 役割分担（結論）
- Judge テスト（最小言語・ロジック検証）
  - C++ と Python のみに限定した A+B と主要判定（AC/WA/PE/TLE/RE/CE/Fail）で、判定・進捗更新・チェック連携など「ロジック」を検証する。
  - 多言語の正当性は Judge では担保しない（= 速度最優先）。
- Executor テスト（全言語の実行検証）
  - 全言語の A+B を実際にコンパイル・実行し、各言語イメージの健全性（ツールチェーン、実行環境、追加ファイルの取扱い等）を検証する。
  - `executor/sources/aplusb/*` の各言語サンプルを使用。
- Langs モジュール（定義とビルド）
  - `langs/langs.toml` に言語ID/コンパイル・実行コマンド・`image_name` を定義。
  - `Dockerfile.*` と `build.sh` で言語イメージをビルド。
  - Langs/Executor に変更が入った場合は「全言語テスト」を走らせる。

## 言語イメージの構成
- 定義ファイル: `langs/`
  - `langs.toml`: 言語ID・コンパイル/実行・`image_name` を定義
  - `langs.go`: TOML を読み込み `LANGS` を構成。特殊言語（checker/verifier/generator）も定義
  - `Dockerfile.*`: 各言語イメージ
  - `build.sh`: 言語イメージのビルドスクリプト
- 主なマッピング（例）
  - C++ 系（`cpp`, `cpp20`, `cpp17`, `cpp-func`）→ `library-checker-images-gcc`
  - Python 系（`python3`）→ `library-checker-images-python3`、（`pypy3`）→ `library-checker-images-pypy`
  - `rust`→`...-rust`, `java`→`...-java`, `go`→`...-golang`, `haskell`→`...-haskell`, `csharp`→`...-csharp`, `swift`→`...-swift`, など
  - 特殊言語: `checker`/`verifier`/`generator` は `library-checker-images-gcc` を使用（最小セットでも gcc は必須）

## CI 方針（高速化を前提）
- Judge（最小セット固定）
  - ビルドするイメージ: `gcc` + `python3`（固定）。
  - 実行: `go test ./judge -v`（C++/Python のみ）。
  - 変更: Judge 側の変更では多言語は走らない。
- Executor（全言語）
  - ビルドするイメージ: 全言語。
  - 実行: `go test ./executor -v`（各言語の A+B をコンパイル/実行して検証）。
  - トリガー: Executor または Langs の変更時（既存の `test-executor.yml` の path で実現）。
- Integration（最小 E2E）
  - 既定は C++ の E2E（必要に応じて Python を追加可）。
  - ビルドするイメージ: 最小セット固定（`gcc` + `python3`）。

## 変更検知と切り替え（簡素化）
- 追加の差分検知は不要。
  - 既存のワークフロー分割で十分（`test-judge.yml` は Judge 変更時のみ、`test-executor.yml` は Executor/Langs 変更時のみ動作）。
  - これにより通常は最小セット（Judge/Integration）、言語や実行基盤の変更時のみ全言語（Executor）という切り替えが自然に成立する。

## 言語イメージのビルド戦略
- 最小/全ビルドの切り替え
  - 既定は「最小セット固定」: `gcc`（C++ + checker/verifier/generator）+ `python3`。
  - Langs/Executor の変更時のみ「全ビルド」（`test-executor.yml` 側で全言語をビルド）。
- ビルドコマンド（Python スクリプト）
  - 全ビルド: `python3 langs/build.py`
  - 最小: `python3 langs/build.py gcc python3`
  - 利用可能なキー: `--list` で一覧表示（エイリアス: `d->ldc`, `go->golang`, `pypy3->pypy`, `cpp->gcc`）
- キャッシュの活用
  - まずは不要（buildx/GHCR は導入しない）。必要になったら将来検討する。

## Executor 側の全言語 A+B テスト設計（案）
- サンプルコード: `executor/sources/aplusb/*` を各言語向けに保持。
- テスト内容（言語ごと）
  - `executor.CompileSource` でコンパイル（`cpp-func` は grader/solve/fastio の追加ファイルを渡す）。
  - `lang.Exec` で実行し、`sample.in` → `sample.out` の一致を検証。
  - 実行環境・ツールチェーンの健全性を保証（Judge ではなく Executor の責務）。

## 新規言語の追加手順（運用）
1. `langs/Dockerfile.<LANG>` を追加し、ローカルで `docker build` 動作確認。
2. `langs/langs.toml` に ID/コンパイル/実行/`image_name` を登録。
3. `executor/sources/aplusb/` に最小 A+B のサンプルソースを追加。
4. PR では Executor 全言語テスト（またはラベル/dispatch で強制）を通し、問題ないことを確認。

## 注意点 / ベストプラクティス
- `checker`/`verifier`/`generator` が `gcc` に依存するため、最小セットでも `gcc` は必須。
- 重量級イメージ（Haskell/Swift等）は Executor/Langs の変更時にのみビルド・テストする方針で CI 時間を抑制。
- 画像タグは `Lang.ImageName` と一致させ、変更時は TOML も合わせて更新。

## ToDo（実装伴う作業）
- `langs/build.sh` のサブセットビルド対応（`gcc python3` だけの最小ビルドを可能に）。
- `test-judge.yml` と `test-integration.yml` を最小セットビルドに変更。
- Executor の全言語 A+B スモークテストの整備。
