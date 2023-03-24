module github.com/yosupo06/library-checker-judge/judge

go 1.18

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/google/uuid v1.1.2
	github.com/lib/pq v1.10.3
	github.com/minio/minio-go/v6 v6.0.57
	github.com/yosupo06/library-checker-judge/api v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.45.0
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/klauspost/cpuid v1.3.1 // indirect
	github.com/minio/md5-simd v1.1.0 // indirect
	github.com/minio/sha256-simd v0.1.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/net v0.0.0-20220325170049-de3da57026de // indirect
	golang.org/x/sys v0.0.0-20220328115105-d36c6a25d886 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220405205423-9d709892a2bf // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
)

replace github.com/yosupo06/library-checker-judge/api => ../api
