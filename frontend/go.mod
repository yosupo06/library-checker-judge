module github.com/yosupo06/library-checker-frontend

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/gin-contrib/sessions v0.0.1
	github.com/gin-gonic/gin v1.4.0
	github.com/jinzhu/gorm v1.9.11
	github.com/lib/pq v1.2.0
	github.com/yosupo06/library-checker-judge/api v0.0.0-20191103214956-e3671c285e76
	golang.org/x/crypto v0.0.0-20190911031432-227b76d455e7
	google.golang.org/grpc v1.24.0
)

// replace github.com/yosupo06/library-checker-judge/api => ../library-checker-judge/api
