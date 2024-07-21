package storage

import (
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Host         string
	ID           string
	Secret       string
	Bucket       string
	PublicBucket string
	useTLS       bool
}

var DEFAULT_CONFIG = Config{
	Host:         "localhost:9000",
	ID:           "minio",
	Secret:       "miniopass",
	Bucket:       "testcase",
	PublicBucket: "testcase-public",
	useTLS:       false,
}

func GetConfigFromEnv() Config {
	config := DEFAULT_CONFIG
	if host := os.Getenv("MINIO_HOST"); host != "" {
		config.Host = host
	}
	if id := os.Getenv("MINIO_ID"); id != "" {
		config.ID = id
	}
	if secret := os.Getenv("MINIO_SECRET"); secret != "" {
		config.Secret = secret
	}
	if bucket := os.Getenv("MINIO_BUCKET"); bucket != "" {
		config.Bucket = bucket
	}
	if publicBucket := os.Getenv("MINIO_PUBLIC_BUCKET"); publicBucket != "" {
		config.PublicBucket = publicBucket
	}
	if useTLS := os.Getenv("MINIO_USE_TLS"); useTLS != "" {
		config.useTLS = true
	}
	return config
}

func Connect(config Config) (*minio.Client, error) {
	return minio.New(
		config.Host, &minio.Options{
			Creds:  credentials.NewStaticV4(config.ID, config.Secret, ""),
			Secure: config.useTLS,
		},
	)
}
