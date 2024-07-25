package storage

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

type Problem struct {
	Name         string
	Version      string
	TestCaseHash string
}

func (p Problem) UploadTestCase(ctx context.Context, c Client, tarGzPath string) error {
	if _, err := c.client.FPutObject(ctx, c.bucket, fmt.Sprintf("v2/%s/%s.tar.gz", p.Name, p.TestCaseHash), tarGzPath, minio.PutObjectOptions{}); err != nil {
		return err
	}
	return nil
}

func (p Problem) UploadPublicFile(ctx context.Context, c Client, localPath, remotePath string) error {
	if _, err := c.client.FPutObject(ctx, c.publicBucket, fmt.Sprintf("v2/%s/%s/%s", p.Name, p.Version, remotePath), localPath, minio.PutObjectOptions{}); err != nil {
		return err
	}
	return nil
}
