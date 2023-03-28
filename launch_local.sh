set -e

docker --version

./api/gen_protoc.sh

docker compose down -v
docker compose up -d --build --wait

(cd deploy && ./gen_protoc.sh)

../library-checker-problems/generate.py --only-html -p aplusb unionfind
(cd uploader && go run . --toml ../../library-checker-problems/sample/aplusb/info.toml)
(cd uploader && go run . --toml ../../library-checker-problems/datastructure/unionfind/info.toml)
