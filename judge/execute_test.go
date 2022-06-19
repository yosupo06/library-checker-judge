package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunHelloWorld(t *testing.T) {
	task, err := NewTaskInfo("ubuntu", WithArguments("echo", "hello-world"))
	if err != nil {
		t.Fatal(err)
	}

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
	task, err := NewTaskInfo("ubuntu", WithArguments("sh", "-c", "exit 123"))
	if err != nil {
		t.Fatal(err)
	}

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
	task, err := NewTaskInfo("ubuntu", WithArguments("sh", "-c", "read input; test $input = dummy"), WithStdin(strings.NewReader("dummy")))
	if err != nil {
		t.Fatal(err)
	}

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
	output := new(bytes.Buffer)
	task, err := NewTaskInfo("ubuntu", WithArguments("echo", "dummy"), WithStdout(output))
	if err != nil {
		t.Fatal(err)
	}

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
	output := new(bytes.Buffer)
	task, err := NewTaskInfo("ubuntu", WithArguments("sh", "-c", "echo dummy >&2"), WithStderr(output))
	if err != nil {
		t.Fatal(err)
	}

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

func TestSleepTime(t *testing.T) {
	task, err := NewTaskInfo("ubuntu", WithArguments("sleep", "3"))
	if err != nil {
		t.Fatal(err)
	}

	result, err := task.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("task result: %v\n", result)

	if !(2*time.Second < result.Time && result.Time < 4*time.Second) {
		t.Error("invalid consumed\n")
	}
}

func TestTimeout(t *testing.T) {
	task, err := NewTaskInfo("ubuntu", WithArguments("sleep", "5"), WithTimeout(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}

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
	// this command consumes 800M memory
	task, err := NewTaskInfo("ubuntu", WithArguments("dd", "if=/dev/zero", "of=/dev/null", "bs=800M"), WithTimeout(3*time.Second), WithMemoryLimitMB(500))
	if err != nil {
		t.Fatal(err)
	}

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

func TestVolume(t *testing.T) {
	volume, err := CreateVolume()
	if err != nil {
		t.Fatal(err)
	}

	if err := volume.CopyFile(bytes.NewBufferString("dummy"), "test.txt"); err != nil {
		t.Fatal(err)
	}
	output := new(bytes.Buffer)

	task, err := NewTaskInfo("ubuntu", WithArguments("cat", "/workdir/test.txt"), WithWorkDir("/workdir"), WithVolume(&volume, "/workdir"), WithStdout(output))
	if err != nil {
		t.Fatal(err)
	}

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

func TestBind(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "tmp-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	task, err := NewTaskInfo("ubuntu", WithArguments("sh", "-c", "echo dummy > /bind/a.txt"), WithBind(tmpdir, "/bind"))
	if err != nil {
		t.Fatal(err)
	}

	result, err := task.Run()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("task result: %v\n", result)
	if result.ExitCode != 0 {
		t.Errorf("Invalid exit code (not 0): %v", result.ExitCode)
	}

	output, err := os.ReadFile(filepath.Join(tmpdir, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(string(output)) != "dummy" {
		t.Errorf("Invalid Stdout: %s", string(output))
	}
}

func TestNetworkDisable(t *testing.T) {
	task, err := NewTaskInfo("ibmcom/ping", WithArguments("ping", "-c", "5", "google.com"), WithStderr(os.Stderr))
	if err != nil {
		t.Fatal(err)
	}

	result, err := task.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("task result: %v\n", result)

	if result.ExitCode == 0 {
		t.Errorf("ping succeeded")
	}
}

func TestForkBomb(t *testing.T) {
	volume, err := CreateVolume()
	if err != nil {
		t.Fatal(err)
	}
	defer volume.Remove()

	src, err := sources.Open("sources/badcode/fork_bomb.sh")
	if err != nil {
		t.Fatal(err)
	}
	if err := volume.CopyFile(src, "fork_bomb.sh"); err != nil {
		t.Fatal(err)
	}

	task, err := NewTaskInfo("ubuntu", WithArguments("./fork_bomb.sh"), WithPidsLimit(100), WithWorkDir("/workdir"), WithVolume(&volume, "/workdir"), WithTimeout(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	result, err := task.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("task result: %v\n", result)
}

func TestUseManyStack(t *testing.T) {
	volume, err := CreateVolume()
	if err != nil {
		t.Fatal(err)
	}
	defer volume.Remove()

	src, err := sources.Open("sources/badcode/use_many_stack.cpp")
	if err != nil {
		t.Fatal(err)
	}
	if err := volume.CopyFile(src, "use_many_stack.cpp"); err != nil {
		t.Fatal(err)
	}

	compileTask, err := NewTaskInfo("gcc:12.1", WithArguments("g++", "use_many_stack.cpp"), WithWorkDir("/workdir"), WithVolume(&volume, "/workdir"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := compileTask.Run(); err != nil {
		t.Fatal(err)
	}

	task, err := NewTaskInfo("gcc:12.1", WithArguments("./a.out"), WithWorkDir("/workdir"), WithVolume(&volume, "/workdir"), WithStackLimitMB(-1))
	if err != nil {
		t.Fatal(err)
	}

	result, err := task.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("task result: %v\n", result)
	if result.ExitCode != 0 {
		t.Error("exec failed")
	}
}
