module github.com/yosupo06/library-checker-frontend

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-contrib/sessions v0.0.1
	github.com/gin-gonic/gin v1.5.0
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/jinzhu/gorm v1.9.12
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/lib/pq v1.3.0
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/yosupo06/library-checker-judge/api v0.0.0-20200224201934-fed229b7d3ce
	golang.org/x/crypto v0.0.0-20200221231518-2aa609cf4a9d
	google.golang.org/grpc v1.27.1
	gopkg.in/go-playground/validator.v9 v9.31.0 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

//replace github.com/yosupo06/library-checker-judge/api => ../library-checker-judge/api
