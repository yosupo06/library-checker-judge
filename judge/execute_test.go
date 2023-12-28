package main

import (
	"bytes"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

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
	outFile.Seek(0, 0)
	return outFile.Name()
}

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
	task, err := NewTaskInfo("ubuntu", WithArguments("sh", "-c", "echo dummy >&2"))
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

	if strings.TrimSpace(string(result.Stderr)) != "dummy" {
		t.Errorf("Invalid Stdout: %s", result.Stderr)
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

	file := toRealFile(bytes.NewBufferString("dummy"), "dummy", t)
	defer os.Remove(file)

	if err := volume.CopyFile(file, "test.txt"); err != nil {
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

func TestNetworkDisable(t *testing.T) {
	task, err := NewTaskInfo("ibmcom/ping", WithArguments("ping", "-c", "5", "google.com"))
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

	file := toRealFile(src, "fork_bomb.sh", t)
	defer os.Remove(file)

	if err := volume.CopyFile(file, "fork_bomb.sh"); err != nil {
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

	file := toRealFile(src, "use_many_stack.cpp", t)
	defer os.Remove(file)

	if err := volume.CopyFile(file, "use_many_stack.cpp"); err != nil {
		t.Fatal(err)
	}

	compileTask, err := NewTaskInfo("gcc:12.1", WithArguments("g++", "use_many_stack.cpp"), WithWorkDir("/workdir"), WithVolume(&volume, "/workdir"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := compileTask.Run(); err != nil {
		t.Fatal(err)
	}

	task, err := NewTaskInfo("gcc:12.1", WithArguments("./a.out"), WithWorkDir("/workdir"), WithVolume(&volume, "/workdir"), WithStackLimitKB(-1))
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
