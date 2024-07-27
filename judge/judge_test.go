package main

import (
	"embed"
	"flag"
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

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func prepareProblemFiles(t *testing.T) storage.ProblemFiles {
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
		{src: SAMPLE_IN_PATH, dst: dir.InFilePath(DUMMY_CASE_NAME)},
		{src: SAMPLE_OUT_PATH, dst: dir.OutFilePath(DUMMY_CASE_NAME)},
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

func testAplusB(t *testing.T, langID, srcName, expectedStatus string) {
	t.Log("Start", langID, srcName)

	files := prepareProblemFiles(t)

	src, err := sources.Open(path.Join(APLUSB_DIR, srcName))
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer src.Close()

	lang, ok := langs.GetLang(langID)
	if !ok {
		t.Fatal("Unknown lang", langID)
	}
	srcFile := toRealFile(src, lang.Source, t)
	defer os.Remove(srcFile)

	checkerVolume, checkerResult, err := compileChecker(files)
	if err != nil || checkerResult.ExitCode != 0 {
		t.Fatal("Error CompileChecker", err)
	}
	t.Cleanup(func() { checkerVolume.Remove() })

	sourceVolume, sourceResult, err := compileSource(files, srcFile, lang)
	if err != nil || sourceResult.ExitCode != 0 {
		t.Fatal("Error CompileSource", err)
	}
	t.Cleanup(func() { sourceVolume.Remove() })

	result, err := testCase(sourceVolume, checkerVolume, lang, 2.0, files.InFilePath(DUMMY_CASE_NAME), files.OutFilePath(DUMMY_CASE_NAME))
	if err != nil {
		t.Fatal("Error to eval testCase", err)
	}
	t.Log("Result:", result)

	if result.Status != expectedStatus {
		t.Fatal("Error Status", result, string(result.Stderr), string(result.CheckerOut))
	}
}

func testAplusBAC(t *testing.T, langID, srcName string) {
	testAplusB(t, langID, srcName, "AC")
}

func TestCppAplusBAC(t *testing.T) {
	testAplusBAC(t, "cpp", "ac.cpp")
}

func TestCppAclAplusBAC(t *testing.T) {
	testAplusBAC(t, "cpp", "ac_acl.cpp")
}
func TestRustAplusBAC(t *testing.T) {
	testAplusBAC(t, "rust", "ac.rs")
}

func TestHaskellAplusBAC(t *testing.T) {
	testAplusBAC(t, "haskell", "ac.hs")
}

func TestHaskellCabalAplusBAC(t *testing.T) {
	testAplusBAC(t, "haskell", "ac_cabal.hs")
}

func TestCSharpAplusBAC(t *testing.T) {
	testAplusBAC(t, "csharp", "ac.cs")
}

func TestLispAplusBAC(t *testing.T) {
	testAplusBAC(t, "lisp", "ac.lisp")
}

func TestPython3AplusBAC(t *testing.T) {
	testAplusBAC(t, "python3", "ac_numpy.py")
}

func TestPyPy3AplusBAC(t *testing.T) {
	testAplusBAC(t, "pypy3", "ac.py")
}
func TestDAplusBAC(t *testing.T) {
	testAplusBAC(t, "d", "ac.d")
}
func TestJavaAplusBAC(t *testing.T) {
	testAplusBAC(t, "java", "ac.java")
}
func TestGoAplusBAC(t *testing.T) {
	testAplusBAC(t, "go", "ac.go")
}
func TestCrystalAplusBAC(t *testing.T) {
	testAplusBAC(t, "crystal", "ac.cr")
}
func TestRubyAplusBAC(t *testing.T) {
	testAplusBAC(t, "ruby", "ac.rb")
}

func TestCppAplusBWA(t *testing.T) {
	testAplusB(t, "cpp", "wa.cpp", "WA")
}

func TestCppAplusBPE(t *testing.T) {
	testAplusB(t, "cpp", "pe.cpp", "PE")
}

func TestCppAplusBTLE(t *testing.T) {
	testAplusB(t, "cpp", "tle.cpp", "TLE")
}

func TestCppAplusBRE(t *testing.T) {
	testAplusB(t, "cpp", "re.cpp", "RE")
}

func TestAplusBCE(t *testing.T) {
	files := prepareProblemFiles(t)

	src, err := sources.Open(path.Join(APLUSB_DIR, "ce.cpp"))
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer src.Close()

	lang, ok := langs.GetLang("cpp")
	if !ok {
		t.Fatal("Unknown lang cpp")
	}
	srcFile := toRealFile(src, lang.Source, t)
	defer os.Remove(srcFile)

	volume, result, err := compileSource(files, srcFile, lang)
	if err != nil {
		t.Fatal("Failed CompileChecker", err, result)
	}
	if result.ExitCode == 0 {
		t.Fatal("Success CompileChecker", result)
	}
	t.Cleanup(func() { volume.Remove() })
}
