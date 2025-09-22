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
	"google.golang.org/api/option"
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
	var clientOptions []option.ClientOption
	emulatorHost := os.Getenv("STORAGE_EMULATOR_HOST")
	if emulatorHost != "" {
		clientOptions = append(clientOptions, option.WithoutAuthentication())
	}

	client, err := storage.NewClient(ctx, clientOptions...)
	if err != nil {
		return Client{}, err
	}

	if emulatorHost != "" {
		projectID := os.Getenv("STORAGE_PROJECT_ID")
		if projectID == "" {
			projectID = "dev-library-checker-project"
		}
		if err := ensureBucketExists(ctx, client, config.Bucket, projectID); err != nil {
			return Client{}, err
		}
		if err := ensureBucketExists(ctx, client, config.PublicBucket, projectID); err != nil {
			return Client{}, err
		}
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

func (c Client) downloadToFile(ctx context.Context, bucketName, objectName, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return err
	}
	reader, err := c.client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = reader.Close()
	}()

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
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

func ensureBucketExists(ctx context.Context, client *storage.Client, bucketName, projectID string) error {
	if bucketName == "" {
		return nil
	}
	if _, err := client.Bucket(bucketName).Attrs(ctx); err != nil {
		if !errors.Is(err, storage.ErrBucketNotExist) {
			return err
		}
		if err := client.Bucket(bucketName).Create(ctx, projectID, nil); err != nil {
			if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 409 {
				return nil
			}
			return fmt.Errorf("create bucket %s: %w", bucketName, err)
		}
	}
	return nil
}
