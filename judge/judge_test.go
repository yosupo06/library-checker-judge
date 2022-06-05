package main

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	ReadLangs("../langs/langs.toml")
	os.Exit(m.Run())
}

func generateAplusBJudge(t *testing.T, lang, srcName string) *Judge {
	checkerPath, err := filepath.Abs("./aplusb/checker.cpp")
	if err != nil {
		t.Fatal(err)
	}
	srcPath, err := filepath.Abs(path.Join("aplusb", srcName))
	if err != nil {
		t.Fatal("Failed: Source", err)
	}

	judge, err := NewJudge("", lang, 2.0)
	if err != nil {
		t.Fatal("Failed to create Judge", err)
	}

	checkerResult, err := judge.CompileChecker(checkerPath, "testlib.h")
	if err != nil || checkerResult.ExitCode != 0 {
		t.Fatal("error CompileChecker", err)
	}
	sourceResult, err := judge.CompileSource(srcPath)
	if err != nil || sourceResult.ExitCode != 0 {
		t.Fatal("error CompileSource", err)
	}

	return judge
}

func testAplusBAC(t *testing.T, lang, srcName string) {
	t.Logf("Start %s test: %s", lang, srcName)
	judge := generateAplusBJudge(t, lang, srcName)
	defer judge.Close()

	result, err := judge.TestCase("./aplusb/sample.in", "./aplusb/sample.out")
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

func TestAplusBWA(t *testing.T) {
	judge := generateAplusBJudge(t, "cpp", "wa.cpp")
	defer judge.Close()

	result, err := judge.TestCase("./aplusb/sample.in", "./aplusb/sample.out")
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

	result, err := judge.TestCase("./aplusb/sample.in", "./aplusb/sample.out")
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

	result, err := judge.TestCase("./aplusb/sample.in", "./aplusb/sample_wa.out")
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

	result, err := judge.TestCase("./aplusb/sample.in", "./aplusb/sample.out")
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

	result, err := judge.TestCase("./aplusb/sample.in", "./aplusb/sample.out")
	if err != nil {
		t.Fatal("error Run Test", err)
	}
	t.Log("Result:", result)

	if result.Status != "RE" {
		t.Fatal("error Status", result)
	}
}

func TestAplusbCE(t *testing.T) {
	srcPath, err := filepath.Abs(path.Join("aplusb", "ce.cpp"))
	if err != nil {
		t.Fatal("Failed: Source", err)
	}

	judge, err := NewJudge("", "cpp", 2.0)
	if err != nil {
		t.Fatal("Failed to create Judge", err)
	}
	defer judge.Close()

	sourceResult, err := judge.CompileSource(srcPath)
	if err != nil {
		t.Fatal("error CompileSource", err)
	}
	if sourceResult.ExitCode == 0 {
		t.Fatal("compile succeeded")
	}
}
