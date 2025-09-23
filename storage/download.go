package storage

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
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
	slog.Info("Download test cases", "name", problem.Name, "hash", problem.TestCaseVersion)

	tarGzPath := path.Join(t.localDir, problem.TestCaseVersion+".tar.gz")
	localDir := path.Join(t.localDir, problem.TestCaseVersion)
	// Phase 2: use v4 path for private testcases tarball
	key := problem.v4TestCasesKey()

	if _, err := os.Stat(tarGzPath); err != nil {
		slog.Info("Download test cases", "remote", key)
		if err := t.client.downloadToFile(context.Background(), t.client.bucket, key, tarGzPath); err != nil {
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
	// Phase 2: switch to v4 public files
	prefix := problem.v4PublicFilesKeyPrefix()
	// ensure trailing slash to avoid absolute-join surprises
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	destDir := path.Join(t.localDir, problem.OverallVersion)
	if _, err := os.Stat(destDir); err != nil {
		slog.Info("Download public files", "name", problem.Name, "overall_version", problem.OverallVersion)
		it := t.client.client.Bucket(t.client.publicBucket).Objects(context.Background(), &storage.Query{Prefix: prefix})
		for {
			attrs, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return "", err
			}
			rel := strings.TrimPrefix(attrs.Name, prefix)
			// v4 layout includes either "common/..." or "{problem}/..."; flatten the latter
			if strings.HasPrefix(rel, problem.Name+"/") {
				rel = strings.TrimPrefix(rel, problem.Name+"/")
			}
			// guard: strip any leading slashes
			rel = strings.TrimLeft(rel, "/")
			if rel == "" {
				continue
			}
			destPath := path.Join(destDir, rel)
			slog.Info("Download public file", "key", attrs.Name, "to", destPath)
			if err := t.client.downloadToFile(context.Background(), t.client.publicBucket, attrs.Name, destPath); err != nil {
				return "", err
			}
		}
	}
	return destDir, nil
}

func (p ProblemFiles) PublicFilePath(key string) string {
	return path.Join(p.PublicFiles, key)
}

func (p ProblemFiles) VerifierPath() string {
	return p.PublicFilePath("verifier.cpp")
}

func (p ProblemFiles) CheckerPath() string {
	return p.PublicFilePath("checker.cpp")
}

func (p ProblemFiles) SolutionPath() string {
	return p.PublicFilePath(path.Join("sol", "correct.cpp"))
}

func (p ProblemFiles) GetAdditionalFilePaths(additionalFiles []string) []string {
	paths := []string{}
	for _, key := range additionalFiles {
		paths = append(paths, p.PublicFilePath(key))
	}
	return paths
}

func (p ProblemFiles) GetIncludeFilePaths() ([]string, error) {
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

// IncludeFilePaths is kept for backward compatibility
func (p ProblemFiles) IncludeFilePaths() ([]string, error) {
	return p.GetIncludeFilePaths()
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
