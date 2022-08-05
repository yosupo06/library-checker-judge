package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/minio/minio-go/v6"
)

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
	zipPath := path.Join(t.casesDir, fmt.Sprintf("cases-%s.zip", version))
	dataPath := path.Join(t.casesDir, fmt.Sprintf("cases-%s", version))
	if _, err := os.Stat(zipPath); err != nil {
		zipFile, err := os.Create(zipPath)
		if err != nil {
			return TestCaseDir{}, err
		}
		object, err := t.minioClient.GetObject(t.minioBucket, version+".zip", minio.GetObjectOptions{})
		if err != nil {
			return TestCaseDir{}, err
		}
		if _, err = io.Copy(zipFile, object); err != nil {
			return TestCaseDir{}, err
		}
		if err = zipFile.Close(); err != nil {
			return TestCaseDir{}, err
		}
		cmd := exec.Command("unzip", zipPath, "-d", dataPath)
		if err := cmd.Run(); err != nil {
			return TestCaseDir{}, err
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
