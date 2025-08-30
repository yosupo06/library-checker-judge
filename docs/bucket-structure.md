# バケット構成とファイル

- 目的: Public と Private の各ファイル、バケット構成、オブジェクトキーの現状を整理する。

**バケット**
- Private バケット: `MINIO_BUCKET`（デフォルト `testcase`）— テストケースの tarball（非公開オブジェクト）を格納。
- Public バケット: `MINIO_PUBLIC_BUCKET`（デフォルト `testcase-public`）— 公開ファイルと例題 I/O を格納。

**オブジェクトキー構造（Object Keys）**
- Private（テストケース tarball）: `v3/{problem}/testcase/{testcase_hash}.tar.gz`
- Public（例題 I/O）:
  - 入力: `v3/{problem}/testcase/{testcase_hash}/in/example_*.in`
  - 出力: `v3/{problem}/testcase/{testcase_hash}/out/example_*.out`
- Public（バージョンごとの公開ファイル）: `v3/{problem}/files/{version}/{path}`

例（problem: `aplusb`, version: `v`, testcase hash: `h`）:
- Tarball: `v3/aplusb/testcase/h.tar.gz`
- 例題入力: `v3/aplusb/testcase/h/in/example_00.in`
- 例題出力: `v3/aplusb/testcase/h/out/example_00.out`
- 公開ファイル: `v3/aplusb/files/v/task.md`

**Private バケットの内容**
- オブジェクト:
  - テストケース tarball のみ: `v3/{problem}/testcase/{testcase_hash}.tar.gz`
- 備考:
  - 非公開なのは tarball のみ。個々のテストケースファイルは非公開ではアップされず、例題 I/O のみが Public バケットに重複配置される。
  - アップロード条件: テストケースハッシュの変更、または `-force` 指定時。
  - 利用者: ジャッジが提出ごとにこの tarball をダウンロードして展開。

**公開ファイルとしてアップロードされるもの（Public バケット）**
- リポジトリルート `common/`（必須）:
  - `common/fastio.h`
  - `common/random.h`
  - `common/testlib.h`
- 問題ディレクトリ（必須）:
  - `task.md`
  - `info.toml`
  - `checker.cpp`
  - `verifier.cpp`
  - `params.h`
  - `sol/correct.cpp`
- 問題ディレクトリ（任意・存在すればアップロード）:
  - `grader/grader.cpp`
  - `grader/solve.hpp`

これらはキー接頭辞 `v3/{problem}/files/{version}/...` の下に配置される。

**アップロードのタイミング**
- テストケース（Private バケット）:
  - トリガ: テストケースバージョン（ハッシュ）が変化、または `-force`。
  - 動作: `v3/{problem}/testcase/{hash}.tar.gz` に tar.gz をアップロード。
  - 併せて: 例題 `.in/.out` を個別に Public バケットの testcase キー配下へアップロード。
- 公開ファイル（Public バケット）:
  - トリガ: 問題バージョンが変化、または `-force`。
  - 動作: 上記の公開ファイル一式を `v3/{problem}/files/{version}/...` にアップロード。

**バージョンとハッシュの定義**
- テストケースバージョン（`TestCaseVersion`）: 問題ディレクトリ内の `hash.json` に基づく（各テストケースファイルのハッシュを結合）。
- 問題バージョン（`Version`）: 以下の内容をハッシュ化:
  - `TestCaseVersion`
  - 公開ファイル集合に含まれる全エントリの内容（必須は存在必須・任意は存在するもののみ。任意ファイルの欠如は計算を阻害しない）。

**Downloader のローカル配置（参考）**
- テストケース展開先: `<tmp>/{testcase_hash}/in|out/...`（Private バケットの tar.gz から展開）。
- 公開ファイル展開先: `<tmp>/{version}/...`（`v3/{problem}/files/{version}/...` の末尾構造をミラー）。
- アクセスヘルパ:
  - `PublicFilePath(key)` -> `<tmp>/{version}/{key}`
  - 例: `verifier.cpp`, `checker.cpp`, `sol/correct.cpp`, `params.h`, `common/*.h`, `task.md`, `info.toml`。

**関連コード**
- アップローダ本体: `uploader/main.go`
- 公開ファイル選定とアップロード: `storage/upload.go`（`fileInfos`, `UploadPublicFiles`）
- キー生成: `storage/problem.go`（`publicFileKeyPrefix`, `publicFileKey`, `publicTestCaseKey`）
- Private tarball のアップロード: `storage/problem.go`（`UploadTestCases`, `testCasesKey`）
- ジャッジによる tarball ダウンロード: `storage/download.go`（`fetchTestCases`）
- バケットと環境変数: `storage/client.go`（`MINIO_BUCKET`, `MINIO_PUBLIC_BUCKET`）
- Downloader の挙動: `storage/download.go`
