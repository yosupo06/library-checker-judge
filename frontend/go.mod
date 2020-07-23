module github.com/yosupo06/library-checker-frontend

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-gonic/gin v1.6.2
	github.com/kr/pretty v0.2.0 // indirect
	github.com/lib/pq v1.4.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/yosupo06/library-checker-judge/api v0.0.0-20200723031016-4c021e50b7cc
	google.golang.org/grpc v1.29.1
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

//replace github.com/yosupo06/library-checker-judge/api => ../library-checker-judge/api
