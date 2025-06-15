set -e

docker --version

./run_protoc.sh

docker compose down -v
docker compose up -d --build --wait

# deploy sample problems
SCRIPT_DIR=$(cd $(dirname $0) && pwd)
PROBLEMS_PATH=$(realpath $SCRIPT_DIR/../library-checker-problems)
(cd uploader && go run . -dir $PROBLEMS_PATH $PROBLEMS_PATH/sample/aplusb/info.toml $PROBLEMS_PATH/data_structure/unionfind/info.toml)
