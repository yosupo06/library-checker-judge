set -e

# check psql was installed
psql --version

docker-compose down -v
docker-compose up -d --build

until PGPASSWORD=passwd psql -c 'select 1;' -U postgres -h localhost 2>&1 > /dev/null; do
    echo 'waiting...'
    sleep 1
done

PGPASSWORD=passwd psql -h localhost -U postgres librarychecker < tables.sql

# test account (name: admin / password: password / admin)
echo "insert into users(name, passhash, admin) values ('admin', '\$2a\$10\$AqftzLHYcaGH2GxUXiGO/OzHnIMJO.PGMrLFqm7mPbpqZlQrIRrq.', true)" \
| PGPASSWORD=passwd psql -h localhost -U postgres librarychecker

# test account (name: judge / password: password / admin)
echo "insert into users(name, passhash, admin) values ('judge', '\$2a\$10\$AqftzLHYcaGH2GxUXiGO/OzHnIMJO.PGMrLFqm7mPbpqZlQrIRrq.', true)" \
| PGPASSWORD=passwd psql -h localhost -U postgres librarychecker

# test account (name: upload / password: password / admin)
echo "insert into users(name, passhash, admin) values ('upload', '\$2a\$10\$AqftzLHYcaGH2GxUXiGO/OzHnIMJO.PGMrLFqm7mPbpqZlQrIRrq.', true)" \
| PGPASSWORD=passwd psql -h localhost -U postgres librarychecker

# test account (name: tester / password: password)
echo "insert into users(name, passhash, admin) values ('tester', '\$2a\$10\$AqftzLHYcaGH2GxUXiGO/OzHnIMJO.PGMrLFqm7mPbpqZlQrIRrq.', false)" \
| PGPASSWORD=passwd psql -h localhost -U postgres librarychecker

sleep 5 # wait to launch

cd deploy && ./gen_protoc.sh && cd ..
PYTHONPATH=../library-checker-problems ./deploy/problems_deploy.py ../library-checker-problems -p aplusb unionfind
