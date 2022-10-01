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
)

const BASE_OBJECT_PATH = "v1"

type TestCaseFetcher struct {
	minioClient *minio.Client
	minioBucket string
	casesDir    string
}

type TestCaseDir struct {
	dir string
}

func NewTestCaseFetcher(minioEndpoint, minioID, minioKey, minioBucket string, minioSecure bool) (TestCaseFetcher, error) {
	// create case directory
	dir, err := ioutil.TempDir("", "case")
	if err != nil {
		log.Print("Failed to create tempdir: ", err)
		return TestCaseFetcher{}, nil
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
		return TestCaseFetcher{}, nil
	}

	return TestCaseFetcher{
		minioClient: client,
		minioBucket: minioBucket,
		casesDir:    dir,
	}, nil
}

func (t *TestCaseFetcher) Close() error {
	if err := os.RemoveAll(t.casesDir); err != nil {
		return err
	}
	return nil
}

func (t *TestCaseFetcher) Fetch(problem string, version string) (TestCaseDir, error) {
	objectPath := path.Join(problem, version)
	dataPath := path.Join(t.casesDir, objectPath)
	if _, err := os.Stat(dataPath); err != nil {
		if err := os.MkdirAll(dataPath, os.ModePerm); err != nil {
			return TestCaseDir{}, err
		}

		for object := range t.minioClient.ListObjects(t.minioBucket, path.Join(BASE_OBJECT_PATH, objectPath), true, nil) {
			key := strings.TrimPrefix(object.Key, path.Join(BASE_OBJECT_PATH, objectPath))
			log.Printf("Download: %s -> %s", object.Key, path.Join(dataPath, key))
			if err := t.minioClient.FGetObject(t.minioBucket, object.Key, path.Join(dataPath, key), minio.GetObjectOptions{}); err != nil {
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
	return os.Open(path.Join(t.dir, "checker.cpp"))
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
	return path.Join(t.dir, "in", name+".in")
}

func (t *TestCaseDir) InFile(name string) (*os.File, error) {
	return os.Open(path.Join(t.dir, "in", name+".in"))
}

func (t *TestCaseDir) OutFilePath(name string) string {
	return path.Join(t.dir, "out", name+".out")
}

func (t *TestCaseDir) OutFile(name string) (*os.File, error) {
	return os.Open(path.Join(t.dir, "out", name+".out"))
}

func (t *TestCaseDir) CaseNames() ([]string, error) {
	// write glob code
	matches, err := filepath.Glob(path.Join(t.dir, "in", "*.in"))
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
