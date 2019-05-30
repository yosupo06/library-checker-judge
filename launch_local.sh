docker run --name mysql -p 3306:3306 -e MYSQL_DATABASE=librarychecker -e MYSQL_ROOT_PASSWORD=passwd -d mysql:8 --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci

until mysqladmin ping -h 127.0.0.1 --silent; do
    echo 'waiting...'
    sleep 1
done

mysql -h 127.0.0.1 -uroot --port 3306 -ppasswd librarychecker < tables.sql
