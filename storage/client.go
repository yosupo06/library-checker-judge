package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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
	if emulatorHost := os.Getenv("STORAGE_EMULATOR_HOST"); emulatorHost != "" {
		return c.downloadFromEmulator(ctx, emulatorHost, bucketName, objectName, destPath)
	}

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

func (c Client) downloadFromEmulator(ctx context.Context, host, bucketName, objectName, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return err
	}

	baseURL, err := parseEmulatorHost(host)
	if err != nil {
		return err
	}

	basePath := baseURL.Path
	if basePath != "" && !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}
	endpoint := fmt.Sprintf("%s://%s/%sdownload/storage/v1/b/%s/o/%s?alt=media",
		baseURL.Scheme,
		baseURL.Host,
		strings.TrimPrefix(basePath, "/"),
		url.PathEscape(bucketName),
		url.PathEscape(objectName),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("emulator download %s: status=%d body=%s", objectName, resp.StatusCode, string(body))
	}

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return err
	}
	return nil
}

func parseEmulatorHost(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, errors.New("empty emulator host")
	}
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil {
			return nil, err
		}
		return u, nil
	}
	return &url.URL{Scheme: "http", Host: raw}, nil
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
