package storage

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/minio/minio-go/v7"
)

type TestCaseDownloader struct {
	client   Client
	localDir string
}

func NewTestCaseDownloader(client Client) (TestCaseDownloader, error) {
	dir, err := os.MkdirTemp("", "case")
	if err != nil {
		slog.Error("Failed to create tempdir", "err", err)
		return TestCaseDownloader{}, err
	}
	slog.Info("TestCaseDownloader created", "dir", dir)
	return TestCaseDownloader{
		client:   client,
		localDir: dir,
	}, nil
}
func (t TestCaseDownloader) Close() error {
	if err := os.RemoveAll(t.localDir); err != nil {
		return err
	}
	return nil
}

type ProblemFiles struct {
	TestCases   string
	PublicFiles string
}

func (t TestCaseDownloader) Fetch(problem Problem) (ProblemFiles, error) {
	testCases, err := t.fetchTestCases(problem)
	if err != nil {
		return ProblemFiles{}, err
	}
	publicFiles, err := t.fetchPublicFiles(problem)
	if err != nil {
		return ProblemFiles{}, err
	}

	return ProblemFiles{
		TestCases:   testCases,
		PublicFiles: publicFiles,
	}, nil
}

func (t TestCaseDownloader) fetchTestCases(problem Problem) (string, error) {
	slog.Info("Download test cases", "name", problem.Name, "hash", problem.TestCaseHash)

	tarGzPath := path.Join(t.localDir, problem.TestCaseHash+".tar.gz")
	localDir := path.Join(t.localDir, problem.TestCaseHash)
	key := problem.testCasesKey()

	if _, err := os.Stat(tarGzPath); err != nil {
		slog.Info("Download test cases", "remote", key)
		if err := t.client.client.FGetObject(context.Background(), t.client.bucket, key, tarGzPath, minio.GetObjectOptions{}); err != nil {
			return "", err
		}
		if err := os.MkdirAll(localDir, os.ModePerm); err != nil {
			return "", err
		}
		cmd := exec.Command("tar", "-xf", tarGzPath, "-C", localDir)
		if err := cmd.Run(); err != nil {
			slog.Error("failed to expand tar.gz")
			return "", err
		}
	}

	return localDir, nil
}

func (t TestCaseDownloader) fetchPublicFiles(problem Problem) (string, error) {
	prefix := problem.publicFileKeyPrefix()

	destDir := path.Join(t.localDir, problem.Version)
	if _, err := os.Stat(destDir); err != nil {
		slog.Info("Download public files", "name", problem.Name, "version", problem.Version)
		for object := range t.client.client.ListObjects(context.Background(), t.client.publicBucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
			destPath := path.Join(destDir, strings.TrimPrefix(object.Key, prefix))
			slog.Info("Download public file", "key", object.Key, "to", destPath)
			if err := t.client.client.FGetObject(context.Background(), t.client.publicBucket, object.Key, destPath, minio.GetObjectOptions{}); err != nil {
				return "", err
			}
		}
	}
	return destDir, nil
}

func (p ProblemFiles) PublicFilePath(key string) string {
	return path.Join(p.PublicFiles, key)
}

func (p ProblemFiles) CheckerPath() string {
	return p.PublicFilePath("checker.cpp")
}

func (p ProblemFiles) SolutionPath() string {
	return p.PublicFilePath(path.Join("sol", "correct.cpp"))
}

func (p ProblemFiles) IncludeFilePaths() ([]string, error) {
	filePaths := []string{
		p.PublicFilePath("params.h"),
	}

	files, err := os.ReadDir(p.PublicFilePath("common"))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		filePaths = append(filePaths, p.PublicFilePath(path.Join("common", file.Name())))
	}

	return filePaths, nil
}

func (p ProblemFiles) InfoTomlPath() string {
	return p.PublicFilePath("info.toml")
}

func (p ProblemFiles) InFilePath(testCase string) string {
	return path.Join(p.TestCases, "in", testCase+".in")
}

func (p ProblemFiles) OutFilePath(testCase string) string {
	return path.Join(p.TestCases, "out", testCase+".out")
}
