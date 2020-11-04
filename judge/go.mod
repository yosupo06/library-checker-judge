module github.com/yosupo06/library-checker-judge/judge

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/klauspost/cpuid v1.3.1 // indirect
	github.com/lib/pq v1.8.0
	github.com/minio/minio-go/v6 v6.0.57
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/yosupo06/library-checker-judge/api v0.0.0-20201104183640-c93ae6dfbe9e
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897 // indirect
	golang.org/x/net v0.0.0-20201031054903-ff519b6c9102 // indirect
	golang.org/x/sys v0.0.0-20201101102859-da207088b7d1 // indirect
	golang.org/x/text v0.3.4 // indirect
	google.golang.org/genproto v0.0.0-20201104152603-2e45c02ce95c // indirect
	google.golang.org/grpc v1.33.1
	gopkg.in/ini.v1 v1.62.0 // indirect
)

//replace github.com/yosupo06/library-checker-judge/api => ../api
