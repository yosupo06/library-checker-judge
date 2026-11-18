package executor

import (
	"bytes"
	"strings"
	"testing"
	"time"

	lc "github.com/yosupo06/library-checker-judge/langs"
)

// Compile through the configured judge command, with no network and the judge's
// normal compile timeout/memory budget. A SwiftPM build per submission would not
// fit this contract; the image must already contain the libraries.
func TestSwiftLibraries(t *testing.T) {
	lang, ok := lc.GetLang("swift")
	if !ok {
		t.Fatal("Swift language configuration is missing")
	}
	source, err := Sources.Open("sources/swift/libraries.swift")
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	srcFile := writeTempFile(t, source, lang.Source)
	opts := []TaskInfoOption{WithMemoryLimitMB(1024), WithPidsLimit(100), WithUnlimitedStackLimit()}
	vol, compRes, err := CompileSource(srcFile, executorLangFrom(lang), opts, 30*time.Second, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	defer vol.Remove()
	if compRes.ExitCode != 0 {
		t.Fatalf("compile failed: exit=%d, stderr=%s", compRes.ExitCode, compRes.Stderr)
	}
	stdout := new(bytes.Buffer)
	task, err := NewTaskInfo(lang.ImageName,
		WithArguments(lang.Exec...), WithWorkDir("/workdir"), WithVolume(&vol, "/workdir"),
		WithStdout(stdout), WithTimeout(10*time.Second), WithMemoryLimitMB(1024),
		WithPidsLimit(100), WithUnlimitedStackLimit())
	if err != nil {
		t.Fatal(err)
	}
	res, err := task.Run()
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("execution failed: exit=%d, stderr=%s", res.ExitCode, res.Stderr)
	}
	const expected = "1 4\n1 4\n3 1 2\n[3, 7, 5]\n6\n9223372036854775808"
	if got := strings.TrimSpace(stdout.String()); got != expected {
		t.Fatalf("wrong output: want %q, got %q", expected, got)
	}
}
