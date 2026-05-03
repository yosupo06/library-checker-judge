#!/bin/bash
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
cd "$SCRIPT_DIR"

export LOCAL_STORAGE_EMULATOR_HOST=${LOCAL_STORAGE_EMULATOR_HOST:-http://localhost:4443}
export LOCAL_STORAGE_PROJECT_ID=${LOCAL_STORAGE_PROJECT_ID:-dev-library-checker-project}
export LOCAL_STORAGE_PRIVATE_BUCKET=${LOCAL_STORAGE_PRIVATE_BUCKET:-testcase}
export LOCAL_STORAGE_PUBLIC_BUCKET=${LOCAL_STORAGE_PUBLIC_BUCKET:-testcase-public}

run_local_uploader() {
    (
        cd uploader
        STORAGE_EMULATOR_HOST="$LOCAL_STORAGE_EMULATOR_HOST" \
        STORAGE_PROJECT_ID="$LOCAL_STORAGE_PROJECT_ID" \
        STORAGE_PRIVATE_BUCKET="$LOCAL_STORAGE_PRIVATE_BUCKET" \
        STORAGE_PUBLIC_BUCKET="$LOCAL_STORAGE_PUBLIC_BUCKET" \
            "$@"
    )
}

docker --version

# Build language images (minimal by default: gcc + python3). Override with LC_LANGS.
# Examples: LC_LANGS=all ./launch_local.sh
echo "Building language images: ${LC_LANGS:-gcc python3}"
(cd langs && python3 ./build.py ${LC_LANGS:-gcc python3})

docker compose down -v
docker compose up -d --build --wait

# deploy sample problems
PROBLEMS_PATH=$(realpath "$SCRIPT_DIR/../library-checker-problems")
echo "Using local storage emulator: $LOCAL_STORAGE_EMULATOR_HOST"
run_local_uploader go run ./problems -dir "$PROBLEMS_PATH" "$PROBLEMS_PATH/sample/aplusb/info.toml" "$PROBLEMS_PATH/data_structure/unionfind/info.toml"

# upload categories from categories.toml
run_local_uploader go run ./categories -dir "$PROBLEMS_PATH"
