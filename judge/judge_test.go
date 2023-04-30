package main

import (
	"embed"
	"flag"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

var (
	TESTLIB_PATH       = path.Join("sources", "testlib.h")
	APLUSB_DIR         = path.Join("sources", "aplusb")
	CHECKER_PATH       = path.Join(APLUSB_DIR, "checker.cpp")
	SAMPLE_IN_PATH     = path.Join(APLUSB_DIR, "sample.in")
	SAMPLE_OUT_PATH    = path.Join(APLUSB_DIR, "sample.out")
	SAMPLE_WA_OUT_PATH = path.Join(APLUSB_DIR, "sample_wa.out")
	DUMMY_CASE_NAME    = "case_00"
)

//go:embed sources/*
var sources embed.FS

func TestMain(m *testing.M) {
	langsTomlPath := flag.String("langs", "../langs/langs.toml", "toml path of langs.toml")

	flag.Parse()

	ReadLangs(*langsTomlPath)
	os.Exit(m.Run())
}

func generateAplusBJudge(t *testing.T, lang, srcName, inFilePath, outFilePath string) *Judge {

	src, err := sources.Open(path.Join(APLUSB_DIR, srcName))
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	defer src.Close()

	srcFile := toRealFile(src, t)
	defer os.Remove(srcFile)

	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal("Failed to create tempDir: ", tempDir)
	}

	caseDir := TestCaseDir{
		dir: tempDir,
	}

	checker, err := sources.ReadFile(CHECKER_PATH)
	if err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(caseDir.CheckerPath(), checker, 0644); err != nil {
		t.Fatal(err)
	}

	inFile, err := sources.ReadFile(inFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(path.Dir(caseDir.InFilePath(DUMMY_CASE_NAME)), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(caseDir.InFilePath(DUMMY_CASE_NAME), inFile, 0644); err != nil {
		t.Fatal(err)
	}

	outFile, err := sources.ReadFile(outFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(path.Dir(caseDir.OutFilePath(DUMMY_CASE_NAME)), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(caseDir.OutFilePath(DUMMY_CASE_NAME), outFile, 0644); err != nil {
		t.Fatal(err)
	}

	judge, err := NewJudge("", langs[lang], 2.0, "", &caseDir)
	if err != nil {
		t.Fatal("Failed to create Judge", err)
	}

	testLibRaw, err := sources.ReadFile(TESTLIB_PATH)
	if err != nil {
		t.Fatal("Failed to open: testlib.h", err)
	}
	if err := os.MkdirAll(path.Join(caseDir.dir, "include"), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(path.Join(caseDir.dir, "include", "testlib.h"), testLibRaw, 0644); err != nil {
		t.Fatal(err)
	}

	checkerResult, err := judge.CompileChecker()
	if err != nil || checkerResult.ExitCode != 0 {
		t.Fatal("error CompileChecker", err)
	}
	sourceResult, _, err := judge.CompileSource(srcFile)
	if err != nil || sourceResult.ExitCode != 0 {
		t.Fatal("error CompileSource", err)
	}

	return judge
}

func testAplusBAC(t *testing.T, lang, srcName string) {
	t.Logf("Start %s test: %s", lang, srcName)
	judge := generateAplusBJudge(t, lang, srcName, SAMPLE_IN_PATH, SAMPLE_OUT_PATH)
	defer judge.Close()

	result, err := judge.TestCase(DUMMY_CASE_NAME)
	if err != nil {
		t.Fatal("error Run Test", err)
	}
	t.Log("Result:", result)

	if result.Status != "AC" {
		t.Fatal("error Status", result, string(result.Stderr), string(result.CheckerOut))
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

func TestAplusBWA(t *testing.T) {
	judge := generateAplusBJudge(t, "cpp", "wa.cpp", SAMPLE_IN_PATH, SAMPLE_OUT_PATH)
	defer judge.Close()

	result, err := judge.TestCase(DUMMY_CASE_NAME)

	if err != nil {
		t.Fatal("error Run Test", err)
	}
	t.Log("Result:", result)

	if result.Status != "WA" {
		t.Fatal("error Status", result)
	}
}

func TestAplusbPE(t *testing.T) {
	judge := generateAplusBJudge(t, "cpp", "pe.cpp", SAMPLE_IN_PATH, SAMPLE_OUT_PATH)
	defer judge.Close()

	result, err := judge.TestCase(DUMMY_CASE_NAME)

	if err != nil {
		t.Fatal("error Run Test", err)
	}
	t.Log("Result:", result)

	if result.Status != "PE" {
		t.Fatal("error Status", result)
	}
}

func TestAplusbFail(t *testing.T) {
	judge := generateAplusBJudge(t, "cpp", "ac.cpp", SAMPLE_IN_PATH, SAMPLE_WA_OUT_PATH)
	defer judge.Close()

	result, err := judge.TestCase(DUMMY_CASE_NAME)

	if err != nil {
		t.Fatal("error Run Test", err)
	}
	t.Log("Result:", result)

	if result.Status != "Fail" {
		t.Fatal("error Status", result)
	}
}

func TestAplusbTLE(t *testing.T) {
	judge := generateAplusBJudge(t, "cpp", "tle.cpp", SAMPLE_IN_PATH, SAMPLE_OUT_PATH)
	defer judge.Close()

	result, err := judge.TestCase(DUMMY_CASE_NAME)
	if err != nil {
		t.Fatal("error Run Test", err)
	}
	t.Log("Result:", result)

	if result.Status != "TLE" {
		t.Fatal("error Status", result)
	}
}

func TestAplusbRE(t *testing.T) {
	judge := generateAplusBJudge(t, "cpp", "re.cpp", SAMPLE_IN_PATH, SAMPLE_OUT_PATH)
	defer judge.Close()

	result, err := judge.TestCase(DUMMY_CASE_NAME)
	if err != nil {
		t.Fatal("error Run Test", err)
	}
	t.Log("Result:", result)

	if result.Status != "RE" {
		t.Fatal("error Status", result)
	}
}

/*
func TestAplusbCE(t *testing.T) {
	src, err := sources.Open(path.Join(APLUSB_DIR, "ce.cpp"))
	if err != nil {
		t.Fatal("Failed: Source", err)
	}
	srcPath := toRealFile(src, t)

	judge, err := NewJudge("", langs["cpp"], 2.0, "")
	if err != nil {
		t.Fatal("Failed to create Judge", err)
	}
	defer judge.Close()

	sourceResult, compileError, err := judge.CompileSource(srcPath)
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
*/
