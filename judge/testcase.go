package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/minio/minio-go/v6"
	"github.com/yosupo06/library-checker-judge/database"
)

const BASE_OBJECT_PATH = "v2"

type TestCaseFetcher struct {
	minioClient       *minio.Client
	minioBucket       string
	minioPublicBucket string
	casesDir          string
}

type TestCaseDir struct {
	dir string
}

func (t *TestCaseDir) PublicFileDir() string {
	return path.Join(t.dir, "public")
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
	return path.Join(t.dir, "in")
}

func (t *TestCaseDir) InFilePath(name string) string {
	return path.Join(t.InFilesDir(), name+".in")
}

func (t *TestCaseDir) InFile(name string) (*os.File, error) {
	return os.Open(t.InFilePath(name))
}

func (t *TestCaseDir) OutFilePath(name string) string {
	return path.Join(t.dir, "out", name+".out")
}

func (t *TestCaseDir) OutFile(name string) (*os.File, error) {
	return os.Open(t.OutFilePath(name))
}

func NewTestCaseFetcher(minioEndpoint, minioID, minioKey, minioBucket, minioPublicBucket string, minioSecure bool) (TestCaseFetcher, error) {
	log.Println("Init TestCaseFetcher bucket:", minioBucket, minioPublicBucket)

	// create case directory
	dir, err := ioutil.TempDir("", "case")
	if err != nil {
		log.Println("Failed to create tempdir:", err)
		return TestCaseFetcher{}, err
	}
	log.Println("TestCaseFetcher data dir:", dir)

	// connect minio
	client, err := minio.New(
		minioEndpoint,
		minioID,
		minioKey,
		minioSecure,
	)

	if err != nil {
		log.Fatalln("Cannot connect to Minio:", err)
		return TestCaseFetcher{}, err
	}

	return TestCaseFetcher{
		minioClient:       client,
		minioBucket:       minioBucket,
		minioPublicBucket: minioPublicBucket,
		casesDir:          dir,
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

		for object := range t.minioClient.ListObjects(t.minioPublicBucket, publicObjectPath, true, nil) {
			key := strings.TrimPrefix(object.Key, publicObjectPath)
			dstPath := path.Join(dataPath, "public", key)
			log.Printf("Download: %s -> %s", object.Key, dstPath)
			if err := t.minioClient.FGetObject(t.minioPublicBucket, object.Key, dstPath, minio.GetObjectOptions{}); err != nil {
				return TestCaseDir{}, err
			}
		}
	}
	return TestCaseDir{dir: dataPath}, nil
}

func (t *TestCaseFetcher) downloadTestCases(problem database.Problem) error {
	s3TestCasesPath := path.Join(BASE_OBJECT_PATH, problem.Name, fmt.Sprintf("%s.tar.gz", problem.TestCasesVersion))

	if _, err := os.Stat(t.testCasesPath(problem)); err != nil {
		if err := t.minioClient.FGetObject(t.minioBucket, s3TestCasesPath, t.testCasesPath(problem), minio.GetObjectOptions{}); err != nil {
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
