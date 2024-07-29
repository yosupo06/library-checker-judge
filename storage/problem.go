package storage

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/minio/minio-go/v7"
)

type Problem struct {
	Name         string
	Version      string
	TestCaseHash string
}

func (p Problem) UploadTestCases(ctx context.Context, c Client, tarGzPath string) error {
	remoteURL := p.testCasesKey()
	slog.Info("Upload test cases", "remote", remoteURL)
	if _, err := c.client.FPutObject(ctx, c.bucket, remoteURL, tarGzPath, minio.PutObjectOptions{}); err != nil {
		return err
	}
	return nil
}

func (p Problem) UploadPublicFile(ctx context.Context, c Client, localPath, key string) error {
	return p.uploadAsPublic(ctx, c, localPath, p.publicFileKey(key))
}

func (p Problem) UploadPublicTestCase(ctx context.Context, c Client, localPath, key string) error {
	return p.uploadAsPublic(ctx, c, localPath, p.publicTestCaseKey(key))
}

func (p Problem) uploadAsPublic(ctx context.Context, c Client, localPath, remoteURL string) error {
	slog.Info("Upload public file", "local", localPath, "remote", remoteURL)
	if _, err := c.client.FPutObject(ctx, c.publicBucket, remoteURL, localPath, minio.PutObjectOptions{}); err != nil {
		return err
	}
	return nil
}

func (p Problem) testCasesKey() string {
	return fmt.Sprintf("v3/%s/testcase/%s.tar.gz", p.Name, p.TestCaseHash)
}

func (p Problem) publicTestCaseKey(key string) string {
	return fmt.Sprintf("v3/%s/testcase/%s/%s", p.Name, p.TestCaseHash, key)
}

func (p Problem) publicFileKeyPrefix() string {
	return fmt.Sprintf("v3/%s/files/%s", p.Name, p.Version)
}

func (p Problem) publicFileKey(key string) string {
	return fmt.Sprintf("%s/%s", p.publicFileKeyPrefix(), key)
}

type Info struct {
	Title     string
	TimeLimit float64
	Tests     []struct {
		Name   string
		Number int
	}
}

func ParseInfo(tomlPath string) (Info, error) {
	info := Info{}
	if _, err := toml.DecodeFile(tomlPath, &info); err != nil {
		return Info{}, err
	}
	return info, nil
}

func (info Info) TestCaseNames() []string {
	names := []string{}
	for _, test := range info.Tests {
		for i := 0; i < test.Number; i++ {
			names = append(names, fmt.Sprintf("%v_%02d", strings.Split(test.Name, ".")[0], i))
		}
	}
	return names
}
