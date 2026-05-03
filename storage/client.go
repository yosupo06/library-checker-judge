package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
)

type Config struct {
	Bucket       string
	PublicBucket string
}

var DEFAULT_CONFIG = Config{
	Bucket:       "testcase",
	PublicBucket: "testcase-public",
}

type Client struct {
	client       *storage.Client
	bucket       string
	publicBucket string
}

func GetConfigFromEnv() Config {
	config := DEFAULT_CONFIG
	if bucket := os.Getenv("STORAGE_PRIVATE_BUCKET"); bucket != "" {
		config.Bucket = bucket
	}
	if publicBucket := os.Getenv("STORAGE_PUBLIC_BUCKET"); publicBucket != "" {
		config.PublicBucket = publicBucket
	}
	return config
}

func Connect(ctx context.Context, config Config) (Client, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return Client{}, err
	}

	return Client{
		client:       client,
		bucket:       config.Bucket,
		publicBucket: config.PublicBucket,
	}, nil
}

func (c Client) Close() error {
	return c.client.Close()
}

func (c Client) downloadToFile(ctx context.Context, bucketName, objectName, destPath string) (err error) {
	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return err
	}
	reader, err := c.client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if _, err := io.Copy(file, reader); err != nil {
		return err
	}
	return nil
}

func (c Client) uploadFile(ctx context.Context, bucketName, objectName, srcPath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	w := c.client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	if _, err := io.Copy(w, file); err != nil {
		_ = w.Close()
		return fmt.Errorf("upload copy failed: %w", err)
	}
	return w.Close()
}

func EnsureBuckets(ctx context.Context, config Config, projectID string) error {
	if projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Close()
	}()

	if err := ensureBucketExists(ctx, client, config.Bucket, projectID); err != nil {
		return err
	}
	if err := ensureBucketExists(ctx, client, config.PublicBucket, projectID); err != nil {
		return err
	}
	return nil
}

func ensureBucketExists(ctx context.Context, client *storage.Client, bucketName, projectID string) error {
	if bucketName == "" {
		return nil
	}
	if _, err := client.Bucket(bucketName).Attrs(ctx); err != nil {
		if !errors.Is(err, storage.ErrBucketNotExist) {
			return err
		}
		if err := client.Bucket(bucketName).Create(ctx, projectID, nil); err != nil {
			var gErr *googleapi.Error
			if errors.As(err, &gErr) && gErr.Code == 409 {
				return nil
			}
			return fmt.Errorf("create bucket %s: %w", bucketName, err)
		}
	}
	return nil
}
