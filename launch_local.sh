set -e

docker --version

./run_protoc.sh

# Build language images (minimal by default: gcc + python3). Override with LC_LANGS.
# Examples: LC_LANGS=all ./launch_local.sh
echo "Building language images: ${LC_LANGS:-gcc python3}"
(cd langs && python3 ./build.py ${LC_LANGS:-gcc python3})

docker compose down -v
docker compose up -d --build --wait

# deploy sample problems
SCRIPT_DIR=$(cd $(dirname $0) && pwd)
PROBLEMS_PATH=$(realpath $SCRIPT_DIR/../library-checker-problems)
(cd uploader && go run ./problems -dir $PROBLEMS_PATH $PROBLEMS_PATH/sample/aplusb/info.toml $PROBLEMS_PATH/data_structure/unionfind/info.toml)

# upload categories from categories.toml
(cd uploader && go run ./categories -dir $PROBLEMS_PATH)
