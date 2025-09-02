//go:build langs_all
// +build langs_all

package executor

import (
	"bytes"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	lc "github.com/yosupo06/library-checker-judge/langs"
)

func writeTempFile(t *testing.T, r io.Reader, name string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "ex-lang-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	fp := path.Join(dir, name)
	if err := os.MkdirAll(path.Dir(fp), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(fp)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(f, r); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return fp
}

// map language id -> sample source path in embed FS
func samplePathFor(langID string) (string, map[string]string) {
	// additional files map: filename -> embedded path
	add := map[string]string{}
	switch langID {
	case "cpp", "cpp20", "cpp17":
		return "sources/aplusb/ac.cpp", nil
	case "cpp-func":
		add["grader.cpp"] = "sources/aplusb/cpp-func/grader.cpp"
		add["solve.hpp"] = "sources/aplusb/cpp-func/solve.hpp"
		add["fastio.h"] = "sources/aplusb/cpp-func/fastio.h"
		return "sources/aplusb/ac_func.cpp", add
	case "rust":
		return "sources/aplusb/ac.rs", nil
	case "java":
		return "sources/aplusb/ac.java", nil
	case "go":
		return "sources/aplusb/go/ac.go", nil
	case "haskell":
		return "sources/aplusb/ac.hs", nil
	case "csharp":
		return "sources/aplusb/ac.cs", nil
	case "d":
		return "sources/aplusb/ac.d", nil
	case "crystal":
		return "sources/aplusb/ac.cr", nil
	case "ruby":
		return "sources/aplusb/ac.rb", nil
	case "lisp":
		return "sources/aplusb/ac.lisp", nil
	case "swift":
		return "sources/aplusb/ac.swift", nil
	case "python3":
		return "sources/aplusb/ac_numpy.py", nil
	case "pypy3":
		return "sources/aplusb/ac.py", nil
	default:
		return "", nil
	}
}

// Basic per-language compile+run conformance test for A+B
func TestAllLangsAplusb(t *testing.T) {
	// Load sample IO
	inBytes, err := Sources.ReadFile("sources/aplusb/sample.in")
	if err != nil {
		t.Fatal(err)
	}
	outBytes, err := Sources.ReadFile("sources/aplusb/sample.out")
	if err != nil {
		t.Fatal(err)
	}
	expected := strings.TrimSpace(string(outBytes))

	for _, lang := range lc.LANGS {
		srcPath, add := samplePathFor(lang.ID)
		if srcPath == "" {
			// No sample for this language; skip silently
			continue
		}
		t.Run(lang.ID, func(t *testing.T) {
			// Write source to temp
			src, err := Sources.Open(srcPath)
			if err != nil {
				t.Fatal(err)
			}
			defer src.Close()
			srcFile := writeTempFile(t, src, path.Base(lang.Source))

			// Prepare additional files
			extra := map[string]string{}
			for name, embedded := range add {
				r, err := Sources.Open(embedded)
				if err != nil {
					t.Fatal(err)
				}
				extra[name] = writeTempFile(t, r, name)
				_ = r.Close()
			}

			// Compile
			vol, compRes, err := CompileSource(srcFile, executorLangFrom(lang), []TaskInfoOption{}, 2*time.Minute, extra)
			if err != nil {
				t.Fatalf("compile error: %v", err)
			}
			if compRes.ExitCode != 0 {
				t.Fatalf("compile failed: exit=%d, stderr=%s", compRes.ExitCode, string(compRes.Stderr))
			}
			defer func() { _ = vol.Remove() }()

			// Execute with sample input
			stdout := new(bytes.Buffer)
			runTask, err := NewTaskInfo(lang.ImageName,
				WithArguments(lang.Exec...),
				WithWorkDir("/workdir"),
				WithVolume(&vol, "/workdir"),
				WithStdout(stdout),
				WithStdin(bytes.NewReader(inBytes)),
				WithTimeout(30*time.Second),
			)
			if err != nil {
				t.Fatal(err)
			}
			res, err := runTask.Run()
			if err != nil {
				t.Fatal(err)
			}
			if res.ExitCode != 0 {
				t.Fatalf("run failed: exit=%d, stderr=%s", res.ExitCode, string(res.Stderr))
			}
			got := strings.TrimSpace(stdout.String())
			if got != expected {
				t.Fatalf("wrong output: want %q, got %q", expected, got)
			}

			// Execute with a simple custom input (123 456 -> 579)
			stdout2 := new(bytes.Buffer)
			runTask2, err := NewTaskInfo(lang.ImageName,
				WithArguments(lang.Exec...),
				WithWorkDir("/workdir"),
				WithVolume(&vol, "/workdir"),
				WithStdout(stdout2),
				WithStdin(bytes.NewReader([]byte("123 456\n"))),
				WithTimeout(30*time.Second),
			)
			if err != nil {
				t.Fatal(err)
			}
			res2, err := runTask2.Run()
			if err != nil {
				t.Fatal(err)
			}
			if res2.ExitCode != 0 {
				t.Fatalf("custom run failed: exit=%d, stderr=%s", res2.ExitCode, string(res2.Stderr))
			}
			got2 := strings.TrimSpace(stdout2.String())
			if got2 != "579" {
				t.Fatalf("wrong custom output: want %q, got %q", "579", got2)
			}
		})
	}
}

// executorLangFrom converts langs.Lang to executor.Lang
func executorLangFrom(l lc.Lang) Lang {
	return Lang{
		ID:              l.ID,
		Name:            l.Name,
		Version:         l.Version,
		Source:          l.Source,
		Compile:         l.Compile,
		Exec:            l.Exec,
		ImageName:       l.ImageName,
		AdditionalFiles: l.AdditionalFiles,
	}
}
