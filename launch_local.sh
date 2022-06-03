set -e

docker --version

./api/gen_protoc.sh

docker compose down -v
docker compose up -d --build --wait

cd deploy && ./gen_protoc.sh && cd ..
PYTHONPATH=../library-checker-problems ./deploy/problems_deploy.py ../library-checker-problems -p aplusb unionfind
