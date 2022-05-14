package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRunHelloWorld(t *testing.T) {
	task := TaskInfo{}
	task.Name = "ubuntu"
	task.Argments = []string{"echo", "hello-world"}

	result, err := task.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("task result: %v\n", result)

	if result.ExitCode != 0 {
		t.Errorf("Invalid exit code (not 0): %v", result.ExitCode)
	}
}

func TestExitCode(t *testing.T) {
	task := TaskInfo{}
	task.Name = "ubuntu"
	task.Argments = []string{"sh", "-c", "exit 123"}

	result, err := task.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("task result: %v\n", result)

	if result.ExitCode != 123 {
		t.Errorf("Invalid exit code (not 123): %v", result.ExitCode)
	}
}

func TestStdin(t *testing.T) {
	task := TaskInfo{}
	task.Name = "ubuntu"
	task.Argments = []string{"sh", "-c", "read input; test $input = dummy"}
	task.Stdin = strings.NewReader("dummy")

	result, err := task.Run()

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("task result: %v\n", result)

	if result.ExitCode != 0 {
		t.Errorf("Invalid exit code (not 0): %v", result.ExitCode)
	}
}

func TestStdout(t *testing.T) {
	task := TaskInfo{}
	task.Name = "ubuntu"
	task.Argments = []string{"echo", "dummy"}
	output := new(bytes.Buffer)
	task.Stdout = output

	result, err := task.Run()

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("task result: %v\n", result)

	if result.ExitCode != 0 {
		t.Errorf("Invalid exit code (not 0): %v", result.ExitCode)
	}

	if strings.TrimSpace(output.String()) != "dummy" {
		t.Errorf("Invalid Stdout: %s", output.String())
	}
}

func TestStderr(t *testing.T) {
	task := TaskInfo{}
	task.Name = "ubuntu"
	task.Argments = []string{"sh", "-c", "echo dummy >&2"}
	output := new(bytes.Buffer)
	task.Stderr = output

	result, err := task.Run()

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("task result: %v\n", result)

	if result.ExitCode != 0 {
		t.Errorf("Invalid exit code (not 0): %v", result.ExitCode)
	}

	if strings.TrimSpace(output.String()) != "dummy" {
		t.Errorf("Invalid Stdout: %s", output.String())
	}
}

func TestTimeout(t *testing.T) {
	task := TaskInfo{}
	task.Name = "ubuntu"
	task.Argments = []string{"sleep", "5"}
	task.Timeout = 3 * time.Second

	result, err := task.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("task result: %v\n", result)

	if !result.TLE {
		t.Error("TLE is not detected")
	}

	if result.ExitCode != 124 {
		t.Errorf("Exit code is not 124: %v", result.ExitCode)
	}
}

func TestMemoryLimit(t *testing.T) {
	task := TaskInfo{}
	task.Name = "ubuntu"
	// this command consumes 800M memory
	task.Argments = []string{"dd", "if=/dev/zero", "of=/dev/null", "bs=800M"}
	task.MemoryLimitMB = 500
	task.Timeout = 3 * time.Second

	result, err := task.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("task result: %v\n", result)

	if result.ExitCode == 0 {
		t.Errorf("Command succeeded: %v", result.ExitCode)
	}
	if result.TLE {
		t.Errorf("TLE is detected")
	}
}
