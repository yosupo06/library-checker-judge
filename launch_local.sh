set -e

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

# test account (name: tester / password: password)
echo "insert into users(name, passhash, admin) values ('tester', '\$2a\$10\$AqftzLHYcaGH2GxUXiGO/OzHnIMJO.PGMrLFqm7mPbpqZlQrIRrq.', true)" \
| PGPASSWORD=passwd psql -h localhost -U postgres librarychecker

cd ../library-checker-problems && ./deploy.py
