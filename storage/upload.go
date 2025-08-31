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
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type UploadTarget struct {
	Base    string
	Root    string
	Problem Problem
}

type FileInfo struct {
	base     string
	path     string
	required bool
}

func fileInfos(base, root string) []FileInfo {
	return []FileInfo{
		// Common files
		// TODO: stop to manually add all common/*.h
		{
			base:     root,
			path:     path.Join("common", "fastio.h"),
			required: true,
		},
		{
			base:     root,
			path:     path.Join("common", "random.h"),
			required: true,
		},
		{
			base:     root,
			path:     path.Join("common", "testlib.h"),
			required: true,
		},
		// Problem files
		{
			base:     base,
			path:     path.Join("task.md"),
			required: true,
		},
		{
			base:     base,
			path:     path.Join("info.toml"),
			required: true,
		},
		{
			base:     base,
			path:     path.Join("checker.cpp"),
			required: true,
		},
		{
			base:     base,
			path:     path.Join("verifier.cpp"),
			required: true,
		},
		{
			base:     base,
			path:     path.Join("params.h"),
			required: true,
		},
		{
			base:     base,
			path:     path.Join("sol", "correct.cpp"),
			required: true,
		},
		// for C++(Function)
		{
			base:     base,
			path:     path.Join("grader", "grader.cpp"),
			required: false,
		},
		{
			base:     base,
			path:     path.Join("grader", "solve.hpp"),
			required: false,
		},
	}
}

func NewUploadTarget(base, root string) (UploadTarget, error) {
	// Normalize to absolute paths to avoid Rel errors when root is relative and base is absolute.
	base, err := filepath.Abs(base)
	if err != nil {
		return UploadTarget{}, err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return UploadTarget{}, err
	}

	h, err := testCaseHash(base)
	if err != nil {
		return UploadTarget{}, err
	}
	v, err := version(base, root)
	if err != nil {
		return UploadTarget{}, err
	}
	ov, err := overallVersion(base, root)
	if err != nil {
		return UploadTarget{}, err
	}
	return UploadTarget{
		Root: root,
		Base: base,
		Problem: Problem{
			Name:            filepath.Base(base),
			TestCaseVersion: h,
			Version:         v,
			OverallVersion:  ov,
		},
	}, nil
}

func testCaseHash(base string) (string, error) {
	caseHash, err := os.ReadFile(path.Join(base, "hash.json"))
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

func version(base, root string) (string, error) {
	hashes := []string{}

	if h, err := testCaseHash(base); err != nil {
		return "", err
	} else {
		hashes = append(hashes, h)
	}

	for _, info := range fileInfos(base, root) {
		path := path.Join(info.base, info.path)
		h, err := fileHash(path)
		if info.required && err != nil {
			return "", err
		}
		hashes = append(hashes, h)
	}

	return joinHashes(hashes), nil
}

// overallVersion computes the v4 OverallVersion.
// It includes TestCaseVersion and the content hash of all files under:
//   - root/common/** (files only)
//   - base/** (the entire problem dir; files only)
func overallVersion(base, root string) (string, error) {
	hashes := []string{}

	if h, err := testCaseHash(base); err != nil {
		return "", err
	} else {
		hashes = append(hashes, h)
	}

	commonFiles, problemFiles, err := gitTrackedFiles(base, root)
	if err != nil {
		return "", err
	}

	for _, rel := range commonFiles {
		fp := path.Join(root, "common", rel)
		h, err := fileHash(fp)
		if err != nil {
			return "", err
		}
		hashes = append(hashes, h)
	}
	for _, rel := range problemFiles {
		fp := path.Join(root, rel)
		h, err := fileHash(fp)
		if err != nil {
			return "", err
		}
		hashes = append(hashes, h)
	}

	return joinHashes(hashes), nil
}

func (p UploadTarget) UploadTestcases(client Client) error {
	tarGz, err := p.BuildTestCaseTarGz()
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tarGz) }()

	// v3 upload
	if err := p.Problem.UploadTestCases(context.Background(), client, tarGz); err != nil {
		return err
	}

	// v4 upload (private)
	if err := p.Problem.UploadTestCasesV4(context.Background(), client, tarGz); err != nil {
		return err
	}

	// upload examples to the public bucket
	for _, ext := range []string{"in", "out"} {
		if err := filepath.Walk(path.Join(p.Base, ext), func(fpath string, info fs.FileInfo, err error) error {
			if strings.Contains(fpath, "example") {
				// v3
				if err := p.Problem.UploadPublicTestCase(context.Background(), client, fpath, path.Join(ext, path.Base(fpath))); err != nil {
					return err
				}
				// v4
				if err := p.Problem.UploadPublicTestCaseV4(context.Background(), client, fpath, path.Join(ext, path.Base(fpath))); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p UploadTarget) BuildTestCaseTarGz() (string, error) {
	tempFile, err := os.CreateTemp("", "testcase*.tar.gz")
	if err != nil {
		return "", err
	}
	defer func() { _ = tempFile.Close() }()

	gzipWriter := gzip.NewWriter(tempFile)
	tarWriter := tar.NewWriter(gzipWriter)

	for _, ext := range []string{"in", "out"} {
		if err := filepath.Walk(path.Join(p.Base, ext), func(fpath string, info fs.FileInfo, err error) error {
			if path.Ext(fpath) == fmt.Sprintf(".%s", ext) {
				file, err := os.Open(fpath)
				if err != nil {
					return err
				}
				defer func() { _ = file.Close() }()

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
			return "", err
		}
	}

	if err := tarWriter.Close(); err != nil {
		return "", err
	}
	if err := gzipWriter.Close(); err != nil {
		return "", err
	}
	return tempFile.Name(), nil
}

func (p UploadTarget) UploadPublicFilesV3(client Client) error {
	for _, info := range fileInfos(p.Base, p.Root) {
		src := path.Join(info.base, info.path)
		if _, err := os.Stat(src); err != nil {
			if info.required {
				return fmt.Errorf("required file: %s/%s not found", info.base, info.path)
			}
			continue
		}

		if err := p.Problem.UploadPublicFile(context.Background(), client, src, info.path); err != nil {
			return err
		}
	}
	return nil
}

func (p UploadTarget) UploadPublicFilesV4(client Client) error {
	commonFiles, problemFiles, err := gitTrackedFiles(p.Base, p.Root)
	if err != nil {
		return err
	}

	for _, relCommon := range commonFiles {
		local := path.Join(p.Root, "common", relCommon)
		remote := p.Problem.v4FilesCommonKey(relCommon)
		if err := p.Problem.UploadPublicFileTo(context.Background(), client, local, remote); err != nil {
			return err
		}
	}

	relProblemDir, err := filepath.Rel(p.Root, p.Base)
	if err != nil {
		return err
	}
	relProblemDir = filepath.ToSlash(relProblemDir)
	prefix := relProblemDir + "/"

    for _, rel := range problemFiles {
        local := path.Join(p.Root, rel)
        sub := strings.TrimPrefix(filepath.ToSlash(rel), prefix)
        remote := p.Problem.v4FilesProblemKey(sub)
        if err := p.Problem.UploadPublicFileTo(context.Background(), client, local, remote); err != nil {
            return err
        }
    }
    // Also upload params.h (generated, not always tracked by git)
    paramsLocal := path.Join(p.Base, "params.h")
    if _, err := os.Stat(paramsLocal); err == nil {
        paramsRemote := p.Problem.v4FilesProblemKey("params.h")
        if err := p.Problem.UploadPublicFileTo(context.Background(), client, paramsLocal, paramsRemote); err != nil {
            return err
        }
    }
    return nil
}

func gitTrackedFiles(base, root string) ([]string, []string, error) {
	relProblemDir, err := filepath.Rel(root, base)
	if err != nil {
		return nil, nil, err
	}
	relProblemDir = filepath.ToSlash(relProblemDir)

	cmd := exec.Command("git", "-C", root, "ls-files")
	out, err := cmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("git ls-files failed: %w", err)
	}
	lines := strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n")
	commonFiles := []string{}
	problemFiles := []string{}
	for _, line := range lines {
		if line == "" {
			continue
		}
		p := filepath.ToSlash(line)
		if strings.HasPrefix(p, "common/") {
			commonFiles = append(commonFiles, strings.TrimPrefix(p, "common/"))
			continue
		}
		if p == relProblemDir || strings.HasPrefix(p, relProblemDir+"/") {
			problemFiles = append(problemFiles, p)
			continue
		}
	}
	return commonFiles, problemFiles, nil
}

func fileHash(path string) (string, error) {
	checker, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(checker)), nil
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
