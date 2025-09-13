module github.com/yosupo06/library-checker-judge/api

go 1.24

require (
	firebase.google.com/go/v4 v4.14.1
	github.com/go-playground/validator/v10 v10.22.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/improbable-eng/grpc-web v0.15.0
	github.com/lib/pq v1.10.9
	github.com/yosupo06/library-checker-judge/database v0.0.0-20240720183423-bc39d9998377
	github.com/yosupo06/library-checker-judge/langs v0.0.0-20240720194232-699a76c34e8c
	google.golang.org/grpc v1.65.0
	google.golang.org/protobuf v1.34.2
	gorm.io/gorm v1.25.11
)

require (
	cloud.google.com/go v0.115.0 // indirect
	cloud.google.com/go/auth v0.7.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.3 // indirect
	cloud.google.com/go/compute/metadata v0.5.0 // indirect
	cloud.google.com/go/firestore v1.15.0 // indirect
	cloud.google.com/go/iam v1.1.11 // indirect
	cloud.google.com/go/longrunning v0.5.10 // indirect
	cloud.google.com/go/storage v1.43.0 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/MicahParks/keyfunc v1.9.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/desertbit/timer v1.0.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/gabriel-vasile/mimetype v1.4.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.5 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/rs/cors v1.11.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.53.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.53.0 // indirect
	go.opentelemetry.io/otel v1.28.0 // indirect
	go.opentelemetry.io/otel/metric v1.28.0 // indirect
	go.opentelemetry.io/otel/trace v1.28.0 // indirect
	golang.org/x/crypto v0.25.0 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/oauth2 v0.21.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/api v0.188.0 // indirect
	google.golang.org/appengine/v2 v2.0.6 // indirect
	google.golang.org/genproto v0.0.0-20240711142825-46eb208f015d // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240711142825-46eb208f015d // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240711142825-46eb208f015d // indirect
	gorm.io/driver/postgres v1.5.9 // indirect
	nhooyr.io/websocket v1.8.11 // indirect
)

replace github.com/yosupo06/library-checker-judge/database => ../database

replace github.com/yosupo06/library-checker-judge/langs => ../langs
