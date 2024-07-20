module github.com/yosupo06/library-checker-judge/judge

go 1.21

require (
	github.com/BurntSushi/toml v1.4.0
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	github.com/minio/minio-go/v6 v6.0.57
	github.com/yosupo06/library-checker-judge/database v0.0.0-20240720160437-32b501e50d46
	gorm.io/gorm v1.25.11
)

require (
	github.com/gabriel-vasile/mimetype v1.4.4 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.22.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/yosupo06/library-checker-judge/langs v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/crypto v0.25.0 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gorm.io/driver/postgres v1.5.9 // indirect
)

replace github.com/yosupo06/library-checker-judge/database => ../database

replace github.com/yosupo06/library-checker-judge/langs => ../langs
