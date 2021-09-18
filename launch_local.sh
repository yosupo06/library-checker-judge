set -e

./api/gen_protoc.sh

# check psql was installed
psql --version

docker-compose down -v
docker-compose up -d --build

until PGPASSWORD=passwd psql -c 'select 1;' -U postgres -h localhost 2>&1 > /dev/null; do
    echo 'waiting...'
    sleep 1
done

sleep 5 # wait to launch

cd deploy && ./gen_protoc.sh && cd ..
PYTHONPATH=../library-checker-problems ./deploy/problems_deploy.py ../library-checker-problems -p aplusb unionfind
