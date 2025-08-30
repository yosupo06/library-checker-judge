set -e

docker --version

./run_protoc.sh

docker compose down -v
docker compose up -d --build --wait

# deploy sample problems
SCRIPT_DIR=$(cd $(dirname $0) && pwd)
PROBLEMS_PATH=$(realpath $SCRIPT_DIR/../library-checker-problems)
(cd uploader && go run ./problems -dir $PROBLEMS_PATH $PROBLEMS_PATH/sample/aplusb/info.toml $PROBLEMS_PATH/data_structure/unionfind/info.toml)

# upload categories from categories.toml
(cd uploader && go run ./categories -dir $PROBLEMS_PATH)
