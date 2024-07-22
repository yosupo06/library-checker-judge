package storage

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/minio/minio-go/v7"
)

type ProblemDir struct {
	Name string
	Root string
	Base string
}

type FileInfo struct {
	base     string
	path     string
	required bool
}

func (p ProblemDir) UploadTestcases(client Client) error {
	h, err := p.TestCaseHash()
	if err != nil {
		return err
	}
	v, err := p.Version()
	if err != nil {
		return err
	}

	tempFile, err := os.CreateTemp("", "testcase*.tar.gz")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	gzipWriter := gzip.NewWriter(tempFile)
	tarWriter := tar.NewWriter(gzipWriter)

	for _, ext := range []string{"in", "out"} {
		if err := filepath.Walk(path.Join(p.Base, ext), func(fpath string, info fs.FileInfo, err error) error {
			if strings.Contains(fpath, "example") {
				if _, err := client.client.FPutObject(context.Background(), client.publicBucket, fmt.Sprintf("v2/%s/%s/%s/%s", p.Name, v, ext, path.Base(fpath)), fpath, minio.PutObjectOptions{}); err != nil {
					return err
				}
			}

			if path.Ext(fpath) == fmt.Sprintf(".%s", ext) {
				file, err := os.Open(fpath)
				if err != nil {
					return err
				}
				defer file.Close()

				fileInfo, err := file.Stat()
				if err != nil {
					return err
				}

				header := &tar.Header{
					Name: fmt.Sprintf("%s/%s", ext, filepath.Base(fpath)),
					Size: fileInfo.Size(),
					Mode: 0600,
				}

				if err := tarWriter.WriteHeader(header); err != nil {
					return err
				}

				_, err = io.Copy(tarWriter, file)
				if err != nil {
					return err
				}

				return nil
			}
			return nil
		}); err != nil {
			return err
		}
	}

	if err := tarWriter.Close(); err != nil {
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		return err
	}

	if _, err := tempFile.Seek(0, 0); err != nil {
		return err
	}
	fileInfo, err := tempFile.Stat()
	if err != nil {
		return err
	}

	if _, err := client.client.PutObject(context.Background(), client.bucket, fmt.Sprintf("v2/%s/%s.tar.gz", p.Name, h), tempFile, fileInfo.Size(), minio.PutObjectOptions{}); err != nil {
		return err
	}

	return nil
}

func (p ProblemDir) UploadFiles(client Client) error {
	v, err := p.Version()
	if err != nil {
		log.Fatal("Failed to fetch version: ", err)
	}

	for _, info := range p.fileInfos() {
		src := path.Join(info.base, info.path)
		if _, err := os.Stat(src); err != nil {
			if info.required {
				return fmt.Errorf("required file: %s/%s not found", info.base, info.path)
			}
			continue
		}

		if _, err := client.client.FPutObject(context.Background(), client.publicBucket, fmt.Sprintf("v2/%s/%s/%s", p.Name, v, info.path), src, minio.PutObjectOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func (p ProblemDir) Version() (string, error) {
	hashes := []string{}

	if h, err := p.TestCaseHash(); err != nil {
		return "", err
	} else {
		hashes = append(hashes, h)
	}

	for _, info := range p.fileInfos() {
		path := path.Join(info.base, info.path)
		h, err := fileHash(path)
		if info.required && err != nil {
			return "", err
		}
		hashes = append(hashes, h)
	}

	return joinHashes(hashes), nil
}

func (p ProblemDir) TestCaseHash() (string, error) {
	caseHash, err := os.ReadFile(path.Join(p.Base, "hash.json"))
	if err != nil {
		return "", err
	}
	var cases map[string]string
	if err := json.Unmarshal(caseHash, &cases); err != nil {
		return "", err
	}

	hashes := make([]string, 0, len(cases))
	for _, v := range cases {
		hashes = append(hashes, v)
	}
	return joinHashes(hashes), nil
}

func joinHashes(hashes []string) string {
	arr := make([]string, len(hashes))
	copy(arr, hashes)
	sort.Strings(arr)

	h := sha256.New()
	for _, v := range arr {
		h.Write([]byte(v))
	}
	return fmt.Sprintf("%x", h.Sum([]byte{}))
}

func (p *ProblemDir) fileInfos() []FileInfo {
	return []FileInfo{
		// Common files
		// TODO: stop to manually add all common/*.h
		{
			base:     p.Root,
			path:     path.Join("common", "fastio.h"),
			required: true,
		},
		{
			base:     p.Root,
			path:     path.Join("common", "random.h"),
			required: true,
		},
		{
			base:     p.Root,
			path:     path.Join("common", "testlib.h"),
			required: true,
		},
		// Problem files
		{
			base:     p.Base,
			path:     path.Join("task.md"),
			required: true,
		},
		{
			base:     p.Base,
			path:     path.Join("info.toml"),
			required: true,
		},
		{
			base:     p.Base,
			path:     path.Join("checker.cpp"),
			required: true,
		},
		{
			base:     p.Base,
			path:     path.Join("params.h"),
			required: true,
		},
		// for C++(Function)
		{
			base:     p.Base,
			path:     path.Join("grader", "grader.cpp"),
			required: false,
		},
		{
			base:     p.Base,
			path:     path.Join("grader", "solve.hpp"),
			required: false,
		},
	}
}

func fileHash(path string) (string, error) {
	checker, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(checker)), nil
}
