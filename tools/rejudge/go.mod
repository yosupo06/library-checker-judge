module github.com/yosupo06/library-checker-judge/tools/rejudge

go 1.25.0

require github.com/alecthomas/kingpin/v2 v2.4.0

require (
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.29.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.9.2 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lib/pq v1.12.3 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	gorm.io/driver/postgres v1.6.0 // indirect
	gorm.io/gorm v1.31.1 // indirect
)

require (
	github.com/alecthomas/units v0.0.0-20240626203959-61d1e3462e30 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	github.com/yosupo06/library-checker-judge/database v0.0.0-20240720194232-699a76c34e8c
)

replace github.com/yosupo06/library-checker-judge/database => ../../database
