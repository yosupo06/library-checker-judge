package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/yosupo06/library-checker-judge/database"
)

const BASE_OBJECT_PATH = "v2"

type TestCaseFetcher struct {
	client   Client
	casesDir string
}

type TestCaseDir struct {
	Dir string
}

func (t *TestCaseDir) PublicFileDir() string {
	return path.Join(t.Dir, "public")
}

func (t *TestCaseDir) PublicFilePath(key string) string {
	return path.Join(t.PublicFileDir(), key)
}

func (t *TestCaseDir) CheckerPath() string {
	return t.PublicFilePath("checker.cpp")
}

func (t *TestCaseDir) CheckerFile() (*os.File, error) {
	return os.Open(t.CheckerPath())
}

func (t *TestCaseDir) IncludeFilePaths() ([]string, error) {
	filePaths := []string{
		t.PublicFilePath("params.h"),
	}

	files, err := os.ReadDir(path.Join(t.PublicFileDir(), "common"))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		filePaths = append(filePaths, path.Join(t.PublicFileDir(), "common", file.Name()))
	}

	return filePaths, nil
}

func (t *TestCaseDir) InFilesDir() string {
	return path.Join(t.Dir, "in")
}

func (t *TestCaseDir) InFilePath(name string) string {
	return path.Join(t.InFilesDir(), name+".in")
}

func (t *TestCaseDir) InFile(name string) (*os.File, error) {
	return os.Open(t.InFilePath(name))
}

func (t *TestCaseDir) OutFilePath(name string) string {
	return path.Join(t.Dir, "out", name+".out")
}

func (t *TestCaseDir) OutFile(name string) (*os.File, error) {
	return os.Open(t.OutFilePath(name))
}

func NewTestCaseFetcher(client Client) (TestCaseFetcher, error) {
	// create case directory
	dir, err := os.MkdirTemp("", "case")
	if err != nil {
		log.Println("Failed to create tempdir:", err)
		return TestCaseFetcher{}, err
	}
	log.Println("TestCaseFetcher data dir:", dir)

	return TestCaseFetcher{
		client:   client,
		casesDir: dir,
	}, nil
}

func (t *TestCaseFetcher) Close() error {
	if err := os.RemoveAll(t.casesDir); err != nil {
		return err
	}
	return nil
}

func (t *TestCaseFetcher) Fetch(problem database.Problem) (TestCaseDir, error) {
	publicObjectPath := path.Join(BASE_OBJECT_PATH, problem.Name, problem.Version)
	dataPath := path.Join(t.casesDir, path.Join(BASE_OBJECT_PATH, problem.Name, problem.Version))

	if err := t.downloadTestCases(problem); err != nil {
		return TestCaseDir{}, err
	}

	if _, err := os.Stat(dataPath); err != nil {
		if err := os.MkdirAll(dataPath, os.ModePerm); err != nil {
			return TestCaseDir{}, err
		}

		cmd := exec.Command("tar", "-xf", t.testCasesPath(problem), "-C", dataPath)
		if err := cmd.Run(); err != nil {
			return TestCaseDir{}, err
		}

		log.Printf("Download public files: %s//%s", t.client.publicBucket, publicObjectPath)
		for object := range t.client.client.ListObjects(context.Background(), t.client.publicBucket, minio.ListObjectsOptions{Prefix: publicObjectPath, Recursive: true}) {
			key := strings.TrimPrefix(object.Key, publicObjectPath)
			dstPath := path.Join(dataPath, "public", key)
			log.Printf("Download: %s -> %s", object.Key, dstPath)
			if err := t.client.client.FGetObject(context.Background(), t.client.publicBucket, object.Key, dstPath, minio.GetObjectOptions{}); err != nil {
				return TestCaseDir{}, err
			}
		}
	}
	return TestCaseDir{Dir: dataPath}, nil
}

func (t *TestCaseFetcher) downloadTestCases(problem database.Problem) error {
	s3TestCasesPath := path.Join(BASE_OBJECT_PATH, problem.Name, fmt.Sprintf("%s.tar.gz", problem.TestCasesVersion))

	if _, err := os.Stat(t.testCasesPath(problem)); err != nil {
		if err := t.client.client.FGetObject(context.Background(), t.client.bucket, s3TestCasesPath, t.testCasesPath(problem), minio.GetObjectOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func (t *TestCaseFetcher) testCasesPath(problem database.Problem) string {
	return path.Join(t.casesDir, fmt.Sprintf("%s-%s.tar.gz", problem.Name, problem.TestCasesVersion))
}

func (t *TestCaseDir) CaseNames() ([]string, error) {
	// write glob code
	matches, err := filepath.Glob(path.Join(t.InFilesDir(), "*.in"))
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, match := range matches {
		_, name := path.Split(match)
		name = strings.TrimSuffix(name, ".in")
		result = append(result, name)
	}
	sort.Strings(result)
	return result, nil
}
