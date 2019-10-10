package main

import (
	"os"
	"os/exec"
	"testing"	
)

func TestExecutorHello(t *testing.T) {
	cmd := exec.Command("echo", "Hello")
	cmd.Stdin = os.Stdin
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
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	result, err := SafeRun(cmd, 1.0, false)
	if err != nil {
		t.Fatalf("Fail Execute: %v", err)
	}
	if result.ReturnCode != 124 {
		t.Errorf("Error return code: %v", result.ReturnCode)
	}
	if result.Time < 0.9 || 1.1 < result.Time {
		t.Error("Error result time")
	}
}
