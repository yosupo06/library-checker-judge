module github.com/yosupo06/library-checker-frontend

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-gonic/gin v1.6.3
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2 // indirect
	github.com/improbable-eng/grpc-web v0.13.0 // indirect
	github.com/jinzhu/gorm v1.9.16 // indirect
	github.com/lib/pq v1.8.0
	github.com/yosupo06/library-checker-judge/api v0.0.0-20201104183640-c93ae6dfbe9e
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897 // indirect
	golang.org/x/net v0.0.0-20201031054903-ff519b6c9102 // indirect
	golang.org/x/sys v0.0.0-20201101102859-da207088b7d1 // indirect
	golang.org/x/text v0.3.4 // indirect
	google.golang.org/genproto v0.0.0-20201104152603-2e45c02ce95c // indirect
	google.golang.org/grpc v1.33.1
)

//replace github.com/yosupo06/library-checker-judge/api => ../library-checker-judge/api
