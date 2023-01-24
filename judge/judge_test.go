package main

import (
	"embed"
	"flag"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"
)

var (
	TESTLIB_PATH       = path.Join("sources", "testlib.h")
	APLUSB_DIR         = path.Join("sources", "aplusb")
	CHECKER_PATH       = path.Join(APLUSB_DIR, "checker.cpp")
	SAMPLE_IN_PATH     = path.Join(APLUSB_DIR, "sample.in")
	SAMPLE_OUT_PATH    = path.Join(APLUSB_DIR, "sample.out")
	SAMPLE_WA_OUT_PATH = path.Join(APLUSB_DIR, "sample_wa.out")
)

//go:embed sources/*
var sources embed.FS

func TestMain(m *testing.M) {
	langsTomlPath := flag.String("langs", "../langs/langs.toml", "toml path of langs.toml")

	flag.Parse()

	ReadLangs(*langsTomlPath)
	os.Exit(m.Run())
}

func generateAplusBJudge(t *testing.T, lang, srcName string) *Judge {
	checker, err := sources.Open(CHECKER_PATH)
	if err != nil {
		t.Fatal(err)
	}
	defer checker.Close()

	src, err := sources.Open(path.Join(APLUSB_DIR, srcName))
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer src.Close()

	judge, err := NewJudge("", langs[lang], 2.0, "")
	if err != nil {
		t.Fatal("Failed to create Judge", err)
	}

	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal("Failed to create tempDir: ", tempDir)
	}
	defer os.RemoveAll(tempDir)
	testlibFilePath := filepath.Join(tempDir, "testlib.h")

	testlibFile, err := os.Create(testlibFilePath)
	if err != nil {
		t.Fatal(err)
	}

	testLibRaw, err := sources.Open(TESTLIB_PATH)
	if err != nil {
		t.Fatal("Failed to open: testlib.h", err)
	}
	io.Copy(testlibFile, testLibRaw)

	checkerResult, err := judge.CompileChecker(checker, []string{testlibFilePath})
	if err != nil || checkerResult.ExitCode != 0 {
		t.Fatal("error CompileChecker", err)
	}
	sourceResult, _, err := judge.CompileSource(src)
	if err != nil || sourceResult.ExitCode != 0 {
		t.Fatal("error CompileSource", err)
	}

	return judge
}

func testAplusBAC(t *testing.T, lang, srcName string) {
	t.Logf("Start %s test: %s", lang, srcName)
	judge := generateAplusBJudge(t, lang, srcName)
	defer judge.Close()

	in, err := sources.Open(SAMPLE_IN_PATH)
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer in.Close()

	expect, err := sources.Open(SAMPLE_OUT_PATH)
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer expect.Close()

	result, err := judge.TestCase(in, expect)
	if err != nil {
		t.Fatal("error Run Test", err)
	}
	t.Log("Result:", result)

	if result.Status != "AC" {
		t.Fatal("error Status", result)
	}
}

func TestCppAplusBAC(t *testing.T) {
	testAplusBAC(t, "cpp", "ac.cpp")
}
func TestCppAclAplusBAC(t *testing.T) {
	testAplusBAC(t, "cpp-acl", "ac_acl.cpp")
}
func TestRustAplusBAC(t *testing.T) {
	testAplusBAC(t, "rust", "ac.rs")
}

func TestHaskellAplusBAC(t *testing.T) {
	testAplusBAC(t, "haskell", "ac.hs")
}
func TestHaskellStackAplusBAC(t *testing.T) {
	testAplusBAC(t, "haskell", "ac_stack.hs")
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

func TestAplusBWA(t *testing.T) {
	judge := generateAplusBJudge(t, "cpp", "wa.cpp")
	defer judge.Close()

	in, err := sources.Open(SAMPLE_IN_PATH)
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer in.Close()

	expect, err := sources.Open(SAMPLE_OUT_PATH)
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer expect.Close()

	result, err := judge.TestCase(in, expect)

	if err != nil {
		t.Fatal("error Run Test", err)
	}
	t.Log("Result:", result)

	if result.Status != "WA" {
		t.Fatal("error Status", result)
	}
}

func TestAplusbPE(t *testing.T) {
	judge := generateAplusBJudge(t, "cpp", "pe.cpp")
	defer judge.Close()

	in, err := sources.Open(SAMPLE_IN_PATH)
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer in.Close()

	expect, err := sources.Open(SAMPLE_OUT_PATH)
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer expect.Close()

	result, err := judge.TestCase(in, expect)

	if err != nil {
		t.Fatal("error Run Test", err)
	}
	t.Log("Result:", result)

	if result.Status != "PE" {
		t.Fatal("error Status", result)
	}
}

func TestAplusbFail(t *testing.T) {
	judge := generateAplusBJudge(t, "cpp", "ac.cpp")
	defer judge.Close()

	in, err := sources.Open(SAMPLE_IN_PATH)
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer in.Close()

	expect, err := sources.Open(SAMPLE_WA_OUT_PATH)
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer expect.Close()

	result, err := judge.TestCase(in, expect)

	if err != nil {
		t.Fatal("error Run Test", err)
	}
	t.Log("Result:", result)

	if result.Status != "Fail" {
		t.Fatal("error Status", result)
	}
}

func TestAplusbTLE(t *testing.T) {
	judge := generateAplusBJudge(t, "cpp", "tle.cpp")
	defer judge.Close()

	in, err := sources.Open(SAMPLE_IN_PATH)
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer in.Close()

	expect, err := sources.Open(SAMPLE_OUT_PATH)
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer expect.Close()

	result, err := judge.TestCase(in, expect)
	if err != nil {
		t.Fatal("error Run Test", err)
	}
	t.Log("Result:", result)

	if result.Status != "TLE" {
		t.Fatal("error Status", result)
	}
}

func TestAplusbRE(t *testing.T) {
	judge := generateAplusBJudge(t, "cpp", "re.cpp")
	defer judge.Close()

	in, err := sources.Open(SAMPLE_IN_PATH)
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer in.Close()

	expect, err := sources.Open(SAMPLE_OUT_PATH)
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer in.Close()

	result, err := judge.TestCase(in, expect)
	if err != nil {
		t.Fatal("error Run Test", err)
	}
	t.Log("Result:", result)

	if result.Status != "RE" {
		t.Fatal("error Status", result)
	}
}

func TestAplusbCE(t *testing.T) {
	src, err := sources.Open(path.Join(APLUSB_DIR, "ce.cpp"))
	if err != nil {
		t.Fatal("Failed: Source", err)
	}

	judge, err := NewJudge("", langs["cpp"], 2.0, "")
	if err != nil {
		t.Fatal("Failed to create Judge", err)
	}
	defer judge.Close()

	sourceResult, compileError, err := judge.CompileSource(src)
	if err != nil {
		t.Fatal("error CompileSource", err)
	}
	if sourceResult.ExitCode == 0 {
		t.Fatal("compile succeeded")
	}
	if len(compileError) == 0 {
		t.Fatal("compile error is empty")
	}
	t.Log("Compile error:", compileError)
}
