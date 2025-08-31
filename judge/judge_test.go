package main

import (
	"embed"
	"flag"
	"io"
	"os"
	"path"
	"testing"

	"github.com/yosupo06/library-checker-judge/langs"
	"github.com/yosupo06/library-checker-judge/storage"
)

var (
	TESTLIB_PATH       = path.Join("sources", "testlib.h")
	APLUSB_DIR         = path.Join("sources", "aplusb")
	CHECKER_PATH       = path.Join(APLUSB_DIR, "checker.cpp")
	PARAMS_H_PATH      = path.Join(APLUSB_DIR, "params.h")
	SAMPLE_IN_PATH     = path.Join(APLUSB_DIR, "sample.in")
	SAMPLE_OUT_PATH    = path.Join(APLUSB_DIR, "sample.out")
	SAMPLE_WA_OUT_PATH = path.Join(APLUSB_DIR, "sample_wa.out")
	DUMMY_CASE_NAME    = "case_00"
)

//go:embed sources/*
var sources embed.FS

func toRealFile(src io.Reader, name string, t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatal(err)
		}
	})

	outFile, err := os.Create(path.Join(tmpDir, name))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(outFile, src); err != nil {
		t.Fatal(err)
	}
	if err := outFile.Close(); err != nil {
		t.Fatal(err)
	}
	return outFile.Name()
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func prepareProblemFiles(t *testing.T, inFilePath, outFilePath string) storage.ProblemFiles {
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal("Failed to create tempDir: ", tempDir)
	}

	dir := storage.ProblemFiles{
		PublicFiles: tempDir,
		TestCases:   tempDir,
	}

	type Info struct {
		src string
		dst string
	}
	for _, info := range []Info{
		{src: CHECKER_PATH, dst: dir.CheckerPath()},
		{src: TESTLIB_PATH, dst: dir.PublicFilePath(path.Join("common", "testlib.h"))},
		{src: PARAMS_H_PATH, dst: dir.PublicFilePath("params.h")},
		{src: inFilePath, dst: dir.InFilePath(DUMMY_CASE_NAME)},
		{src: outFilePath, dst: dir.OutFilePath(DUMMY_CASE_NAME)},
	} {
		checker, err := sources.ReadFile(info.src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(path.Dir(info.dst), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(info.dst, checker, 0644); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

func testAplusB(t *testing.T, langID, srcName, inFilePath, outFilePath, expectedStatus string) {
	t.Log("Start", langID, srcName)

	files := prepareProblemFiles(t, inFilePath, outFilePath)

	src, err := sources.Open(path.Join(APLUSB_DIR, srcName))
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer func() { _ = src.Close() }()

	lang, ok := langs.GetLang(langID)
	if !ok {
		t.Fatal("Unknown lang", langID)
	}
	srcFile := toRealFile(src, lang.Source, t)
	defer func() { _ = os.Remove(srcFile) }()

	checkerVolume, checkerResult, err := compileChecker(files)
	if err != nil || checkerResult.ExitCode != 0 {
		t.Fatal("Error CompileChecker", err)
	}
	t.Cleanup(func() { _ = checkerVolume.Remove() })

	sourceVolume, sourceResult, err := compile(files, srcFile, lang)
	if err != nil || sourceResult.ExitCode != 0 {
		t.Fatal("Error CompileSource", err)
	}
	t.Cleanup(func() { _ = sourceVolume.Remove() })

	result, err := runTestCase(sourceVolume, checkerVolume, lang, 2.0, files.InFilePath(DUMMY_CASE_NAME), files.OutFilePath(DUMMY_CASE_NAME))
	if err != nil {
		t.Fatal("Error to eval testCase", err)
	}
	t.Log("Result:", result)

	if result.Status != expectedStatus {
		t.Fatal("Error Status", result, string(result.Stderr), string(result.CheckerOut))
	}
}

func testAplusBAC(t *testing.T, langID, srcName string) {
	testAplusB(t, langID, srcName, SAMPLE_IN_PATH, SAMPLE_OUT_PATH, "AC")
}

func TestCppAplusBAC(t *testing.T) {
	testAplusBAC(t, "cpp", "ac.cpp")
}

func TestCppAclAplusBAC(t *testing.T) {
	testAplusBAC(t, "cpp", "ac_acl.cpp")
}

func TestPython3AplusBAC(t *testing.T) {
	testAplusBAC(t, "python3", "ac_numpy.py")
}

func TestCppAplusBWA(t *testing.T) {
	testAplusB(t, "cpp", "wa.cpp", SAMPLE_IN_PATH, SAMPLE_OUT_PATH, "WA")
}

func TestCppAplusBPE(t *testing.T) {
	testAplusB(t, "cpp", "pe.cpp", SAMPLE_IN_PATH, SAMPLE_OUT_PATH, "PE")
}

func TestCppAplusBTLE(t *testing.T) {
	testAplusB(t, "cpp", "tle.cpp", SAMPLE_IN_PATH, SAMPLE_OUT_PATH, "TLE")
}

func TestCppAplusBRE(t *testing.T) {
	testAplusB(t, "cpp", "re.cpp", SAMPLE_IN_PATH, SAMPLE_OUT_PATH, "RE")
}

func TestCppAplusBFail(t *testing.T) {
	testAplusB(t, "cpp", "ac.cpp", SAMPLE_IN_PATH, SAMPLE_WA_OUT_PATH, "Fail")
}

func TestAplusBCE(t *testing.T) {
	files := prepareProblemFiles(t, SAMPLE_IN_PATH, SAMPLE_OUT_PATH)

	src, err := sources.Open(path.Join(APLUSB_DIR, "ce.cpp"))
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer func() { _ = src.Close() }()

	lang, ok := langs.GetLang("cpp")
	if !ok {
		t.Fatal("Unknown lang cpp")
	}
	srcFile := toRealFile(src, lang.Source, t)
	defer func() { _ = os.Remove(srcFile) }()

	volume, result, err := compile(files, srcFile, lang)
	if err != nil {
		t.Fatal("Failed CompileChecker", err, result)
	}
	if result.ExitCode == 0 {
		t.Fatal("Success CompileChecker", result)
	}
	t.Cleanup(func() { _ = volume.Remove() })
}
