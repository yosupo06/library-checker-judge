set -e

docker --version

./gen_protoc.sh

docker compose down -v
docker compose up -d --build --wait

# deploy sample problems
PROBLEMS_PATH=$(realpath $PROBLEMS_PATH)
(cd uploader && go run . -dir $PROBLEMS_PATH $PROBLEMS_PATH/sample/aplusb/info.toml $PROBLEMS_PATH/data_structure/unionfind/info.toml)
