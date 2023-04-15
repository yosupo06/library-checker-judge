set -e

docker --version

./api/gen_protoc.sh

docker compose down -v
docker compose up -d --build --wait

(cd deploy && ./gen_protoc.sh)

(cd uploader && go run . ../../library-checker-problems/sample/aplusb/info.toml ../../library-checker-problems/datastructure/unionfind/info.toml)
