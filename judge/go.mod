module github.com/yosupo06/library-checker-judge/judge

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/lib/pq v1.5.2
	github.com/minio/minio-go/v6 v6.0.55
	github.com/yosupo06/library-checker-judge/api v0.0.0-20200518154354-1e7d1562e900
	google.golang.org/grpc v1.29.1
)

//replace github.com/yosupo06/library-checker-judge/api => ../api
