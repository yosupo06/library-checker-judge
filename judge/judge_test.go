package main

import (
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

func TestExecutorHello(t *testing.T) {
	cmd := exec.Command("echo", "Hello")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	result, err := SafeRun(cmd, 1.0, false)
	if err != nil {
		t.Fatalf("Fail Execute: %v", err)
	}
	t.Log(result)
	if result.ReturnCode != 0 {
		t.Errorf("Error return code: %v", result.ReturnCode)
	}
	if 0.5 < result.Time {
		t.Error("Comsume too long time for Hello")
	}
}

func TestExecutorTimeOut(t *testing.T) {
	cmd := exec.Command("sleep", "5")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	result, err := SafeRun(cmd, 1.0, false)
	if err != nil {
		t.Fatal("Error ", err)
	}
	if cmd.ProcessState.ExitCode() != 124 {
		t.Fatal("Error return code = ", result.ReturnCode)
	}
	if result.Time < 0.9 || 1.1 < result.Time {
		t.Fatal("Error result time = ", result.Time)
	}
}

func generateAplusB(t *testing.T, lang, srcName string) *Judge {
	checker, err := os.Open("./test_src/aplusb/checker.cpp")
	if err != nil {
		t.Fatal("Failed: Checker", err)
	}
	src, err := os.Open(path.Join("test_src/aplusb", srcName))
	if err != nil {
		t.Fatal("Failed: Source", err)
	}

	judge, err := NewJudge(lang, checker, src, 2.0)
	if err != nil {
		t.Fatal("Failed: NewJudge", err)
	}

	result, err := judge.CompileChecker()
	if err != nil || result.ReturnCode != 0 {
		t.Fatal("error CompileChecker", err, result.ReturnCode)
	}
	result, err = judge.CompileSource()
	if err != nil || result.ReturnCode != 0 {
		t.Fatal("error CompileSource", err, result.ReturnCode)
	}

	return judge
}

func TestAplusbAC(t *testing.T) {
	judge := generateAplusB(t, "cpp", "ac.cpp")
	in := strings.NewReader("1 1")
	expect := strings.NewReader("2")
	result, err := judge.TestCase(in, expect)
	log.Println(judge.dir)
	if err != nil {
		t.Fatal("error Run Test", err)
	}
	if result.Status != "AC" {
		t.Fatal("error Status", result)
	}
}

func TestAplusbRustAC(t *testing.T) {
	judge := generateAplusB(t, "rust", "ac.rs")
	in := strings.NewReader("1 1")
	expect := strings.NewReader("2")
	result, err := judge.TestCase(in, expect)
	log.Println(judge.dir)
	if err != nil {
		t.Fatal("error Run Test", err)
	}
	if result.Status != "AC" {
		t.Fatal("error Status", result)
	}
}

func TestAplusbWA(t *testing.T) {
	judge := generateAplusB(t, "cpp", "wa.cpp")
	in := strings.NewReader("1 1")
	expect := strings.NewReader("2")
	result, err := judge.TestCase(in, expect)
	log.Println(judge.dir)
	if err != nil {
		t.Fatal("error Run Test", err)
	}
	if result.Status != "WA" {
		t.Fatal("error Status", result)
	}
}

func TestAplusbPE(t *testing.T) {
	judge := generateAplusB(t, "cpp", "pe.cpp")
	in := strings.NewReader("1 1")
	expect := strings.NewReader("2")
	result, err := judge.TestCase(in, expect)
	log.Println(judge.dir)
	if err != nil {
		t.Fatal("error Run Test", err)
	}
	if result.Status != "PE" {
		t.Fatal("error Status", result)
	}
}

func TestAplusbFail(t *testing.T) {
	judge := generateAplusB(t, "cpp", "ac.cpp")
	in := strings.NewReader("1 1")
	expect := strings.NewReader("3") // !?
	result, err := judge.TestCase(in, expect)
	log.Println(judge.dir)
	if err != nil {
		t.Fatal("error Run Test", err)
	}
	if result.Status != "Fail" {
		t.Fatal("error Status", result)
	}
}
