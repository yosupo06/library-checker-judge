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
    Name            string
    Version         string
    OverallVersion  string
    TestCaseVersion string
}

func (p Problem) UploadTestCases(ctx context.Context, c Client, tarGzPath string) error {
    remoteURL := p.testCasesKey()
    slog.Info("Upload test cases", "remote", remoteURL)
    if _, err := c.client.FPutObject(ctx, c.bucket, remoteURL, tarGzPath, minio.PutObjectOptions{}); err != nil {
        return err
    }
    return nil
}

// UploadTestCasesV4 uploads testcases tarball also to v4 private path for Phase 1 dual-write.
func (p Problem) UploadTestCasesV4(ctx context.Context, c Client, tarGzPath string) error {
    remoteURL := p.v4TestCasesKey()
    slog.Info("Upload test cases (v4)", "remote", remoteURL)
    if _, err := c.client.FPutObject(ctx, c.bucket, remoteURL, tarGzPath, minio.PutObjectOptions{}); err != nil {
        return err
    }
    return nil
}

func (p Problem) UploadPublicFile(ctx context.Context, c Client, localPath, key string) error {
    return p.uploadAsPublic(ctx, c, localPath, p.publicFileKey(key))
}

// UploadPublicFileTo uploads a local file to an explicitly specified public remote URL.
// This is useful for transitional migrations where multiple public schemas coexist.
func (p Problem) UploadPublicFileTo(ctx context.Context, c Client, localPath, remoteURL string) error {
    return p.uploadAsPublic(ctx, c, localPath, remoteURL)
}

func (p Problem) UploadPublicTestCase(ctx context.Context, c Client, localPath, key string) error {
    return p.uploadAsPublic(ctx, c, localPath, p.publicTestCaseKey(key))
}

// UploadPublicTestCaseV4 uploads example I/O to v4 examples path for Phase 1 dual-write.
func (p Problem) UploadPublicTestCaseV4(ctx context.Context, c Client, localPath, key string) error {
    return p.uploadAsPublic(ctx, c, localPath, p.v4ExamplesKey(key))
}

func (p Problem) uploadAsPublic(ctx context.Context, c Client, localPath, remoteURL string) error {
	slog.Info("Upload public file", "local", localPath, "remote", remoteURL)
	if _, err := c.client.FPutObject(ctx, c.publicBucket, remoteURL, localPath, minio.PutObjectOptions{}); err != nil {
		return err
	}
	return nil
}

func (p Problem) testCasesKey() string {
    return fmt.Sprintf("%s.tar.gz", p.testCaseKeyPrefix())
}

func (p Problem) publicTestCaseKey(key string) string {
	return fmt.Sprintf("%s/%s", p.testCaseKeyPrefix(), key)
}

func (p Problem) testCaseKeyPrefix() string {
    return fmt.Sprintf("v3/%s/testcase/%s", p.Name, p.TestCaseVersion)
}

func (p Problem) publicFileKeyPrefix() string {
    return fmt.Sprintf("v3/%s/files/%s", p.Name, p.Version)
}

func (p Problem) publicFileKey(key string) string {
    return fmt.Sprintf("%s/%s", p.publicFileKeyPrefix(), key)
}

// v4 paths (Phase 1 dual-write)
func (p Problem) v4TestCasesKey() string {
    return fmt.Sprintf("v4/testcase/%s/%s.tar.gz", p.Name, p.TestCaseVersion)
}

func (p Problem) v4ExamplesKey(key string) string {
    return fmt.Sprintf("v4/examples/%s/%s/%s", p.Name, p.TestCaseVersion, key)
}

func (p Problem) v4FilesCommonKey(key string) string {
    // key is like "common/xxx" or just file under common. We accept both.
    commonPath := key
    if strings.HasPrefix(commonPath, "common/") {
        commonPath = strings.TrimPrefix(commonPath, "common/")
    }
    return fmt.Sprintf("v4/files/%s/%s/common/%s", p.Name, p.OverallVersion, commonPath)
}

func (p Problem) v4FilesProblemKey(rel string) string {
    // rel is path relative to the problem dir (e.g., "task.md", "sol/correct.cpp")
    return fmt.Sprintf("v4/files/%s/%s/%s/%s", p.Name, p.OverallVersion, p.Name, rel)
}

// publicCommonV4Key returns the v4 path for common/* files only.
// v4/files/{problem}/{version}/common/{path_under_common}
func (p Problem) publicCommonV4Key(key string) string {
    // key is expected like "common/fastio.h"
    trimmed := strings.TrimPrefix(key, "common/")
    return fmt.Sprintf("v4/files/%s/%s/common/%s", p.Name, p.Version, trimmed)
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
