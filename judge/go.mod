module github.com/yosupo06/library-checker-judge/judge

go 1.13

require (
	9fans.net/go v0.0.2
	cloud.google.com/go v0.37.4 // indirect
	github.com/BurntSushi/toml v0.3.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-gonic/gin v1.6.2
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/jinzhu/gorm v1.9.12
	github.com/kr/pretty v0.2.0 // indirect
	github.com/lib/pq v1.4.0
	github.com/yosupo06/library-checker-judge/api v0.0.0-20200425150722-92dad6d5d37a
	golang.org/x/net v0.0.0-20200425230154-ff2c4b7c35a0 // indirect
	golang.org/x/sys v0.0.0-20200428200454-593003d681fa // indirect
	google.golang.org/genproto v0.0.0-20200429120912-1f37eeb960b2 // indirect
	google.golang.org/grpc v1.29.1
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/yosupo06/library-checker-judge/api => ../api
