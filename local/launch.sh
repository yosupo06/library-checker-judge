./stop.sh

docker run --name postgresql -p 5432:5432 -e POSTGRES_DB=librarychecker -e POSTGRES_PASSWORD=passwd -d postgres:11.3

until PGPASSWORD=passwd psql -c 'select 1;' -U postgres -h localhost 2>&1 > /dev/null; do
    echo 'waiting...'
    sleep 1
done

PGPASSWORD=passwd psql -h localhost -U postgres librarychecker < tables.sql

# test account (name: admin / password: password)
echo "insert into users(name, passhash, admin) values ('admin', '\$2a\$10\$AqftzLHYcaGH2GxUXiGO/OzHnIMJO.PGMrLFqm7mPbpqZlQrIRrq.', true)" \
| PGPASSWORD=passwd psql -h localhost -U postgres librarychecker

cd ../../library-checker-problems && ./deploy.py
