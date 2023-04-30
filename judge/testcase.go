package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/minio/minio-go/v6"
	"github.com/yosupo06/library-checker-judge/database"
)

const BASE_OBJECT_PATH = "v1"

type TestCaseFetcher struct {
	minioClient       *minio.Client
	minioBucket       string
	minioPublicBucket string
	casesDir          string
}

type TestCaseDir struct {
	dir string
}

func NewTestCaseFetcher(minioEndpoint, minioID, minioKey, minioBucket, minioPublicBucket string, minioSecure bool) (TestCaseFetcher, error) {
	log.Println("init TestCaseFetcher bucket:", minioBucket, minioPublicBucket)

	// create case directory
	dir, err := ioutil.TempDir("", "case")
	if err != nil {
		log.Print("Failed to create tempdir: ", err)
		return TestCaseFetcher{}, err
	}

	// connect minio
	client, err := minio.New(
		minioEndpoint,
		minioID,
		minioKey,
		minioSecure,
	)

	if err != nil {
		log.Fatal("Cannot connect to Minio: ", err)
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
	objectPath := path.Join(BASE_OBJECT_PATH, problem.Name, problem.Testhash)
	publicObjectPath := path.Join(BASE_OBJECT_PATH, problem.Name, problem.PublicFilesHash)
	dataPath := path.Join(t.casesDir, objectPath+"-"+publicObjectPath)
	if _, err := os.Stat(dataPath); err != nil {
		if err := os.MkdirAll(dataPath, os.ModePerm); err != nil {
			return TestCaseDir{}, err
		}

		for object := range t.minioClient.ListObjects(t.minioBucket, objectPath, true, nil) {
			key := strings.TrimPrefix(object.Key, objectPath)
			log.Printf("Download: %s -> %s", object.Key, path.Join(dataPath, key))
			if err := t.minioClient.FGetObject(t.minioBucket, object.Key, path.Join(dataPath, key), minio.GetObjectOptions{}); err != nil {
				return TestCaseDir{}, err
			}
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

func (t *TestCaseDir) CheckerPath() string {
	return path.Join(t.dir, "checker.cpp")
}

func (t *TestCaseDir) CheckerFile() (*os.File, error) {
	return os.Open(t.CheckerPath())
}

func (t *TestCaseDir) IncludeDir() string {
	return path.Join(t.dir, "include")
}

func (t *TestCaseDir) IncludeFilePaths() ([]string, error) {
	files, err := os.ReadDir(t.IncludeDir())
	if err != nil {
		return nil, err
	}
	filePaths := []string{}
	for _, file := range files {
		filePaths = append(filePaths, path.Join(t.IncludeDir(), file.Name()))
	}
	return filePaths, nil
}

func (t *TestCaseDir) InFilePath(name string) string {
	return path.Join(t.dir, "testcases", "in", name+".in")
}

func (t *TestCaseDir) InFile(name string) (*os.File, error) {
	return os.Open(t.InFilePath(name))
}

func (t *TestCaseDir) OutFilePath(name string) string {
	return path.Join(t.dir, "testcases", "out", name+".out")
}

func (t *TestCaseDir) OutFile(name string) (*os.File, error) {
	return os.Open(t.OutFilePath(name))
}

func (t *TestCaseDir) PublicFilePath(key string) string {
	return path.Join(t.dir, "public", key)
}

func (t *TestCaseDir) CaseNames() ([]string, error) {
	// write glob code
	matches, err := filepath.Glob(path.Join(t.dir, "testcases", "in", "*.in"))
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
