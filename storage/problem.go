package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/minio/minio-go/v7"
)

type Problem struct {
	Name         string
	Version      string
	TestCaseHash string
}

func (c Client) UploadTestCase(ctx context.Context, problem Problem, testcase *os.File) error {
	fileInfo, err := testcase.Stat()
	if err != nil {
		return err
	}
	if _, err := c.client.PutObject(ctx, c.bucket, fmt.Sprintf("v2/%s/%s.tar.gz", problem.Name, problem.TestCaseHash), testcase, fileInfo.Size(), minio.PutObjectOptions{}); err != nil {
		return err
	}
	return nil
}
