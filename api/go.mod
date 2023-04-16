module github.com/yosupo06/library-checker-judge/api

go 1.18

require (
	github.com/BurntSushi/toml v1.2.1
	github.com/go-playground/validator/v10 v10.12.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/improbable-eng/grpc-web v0.15.0
	github.com/lib/pq v1.10.7
	github.com/yosupo06/library-checker-judge/database v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.54.0
	google.golang.org/protobuf v1.30.0
	gorm.io/gorm v1.24.7-0.20230306060331-85eaf9eeda11
)

require (
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.3.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.16.4 // indirect
	github.com/leodido/go-urn v1.2.3 // indirect
	github.com/rs/cors v1.8.3 // indirect
	golang.org/x/crypto v0.8.0 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230403163135-c38d8f061ccd // indirect
	gorm.io/driver/postgres v1.5.0 // indirect
	nhooyr.io/websocket v1.8.7 // indirect
)

replace github.com/yosupo06/library-checker-judge/database => ../database
