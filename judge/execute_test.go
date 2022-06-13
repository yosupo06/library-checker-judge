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

func TestSleepTime(t *testing.T) {
	task := TaskInfo{}
	task.Name = "ubuntu"
	task.Argments = []string{"sleep", "3"}

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

func TestVolume(t *testing.T) {
	volume, err := CreateVolume()
	if err != nil {
		t.Fatal(err)
	}

	if err := volume.CopyFile(bytes.NewBufferString("dummy"), "test.txt"); err != nil {
		t.Fatal(err)
	}
	output := new(bytes.Buffer)

	task := TaskInfo{
		Name:     "ubuntu",
		Argments: []string{"cat", "/workdir/test.txt"},
		WorkDir:  "/workdir",
		VolumeMountInfo: []VolumeMountInfo{
			{
				Path:   "/workdir",
				Volume: &volume,
			},
		},
		Stdout: output,
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

	task := TaskInfo{
		Name:     "ubuntu",
		Argments: []string{"sh", "-c", "echo dummy > /bind/a.txt"},
		Binds: []BindInfo{
			{
				HostPath:      tmpdir,
				ContainerPath: "/bind",
			},
		},
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

func TestNetworkEnable(t *testing.T) {
	task := TaskInfo{
		Name:          "ibmcom/ping",
		EnableNetwork: true,
		Stderr:        os.Stderr,
	}
	task.Argments = []string{"ping", "-c", "5", "google.com"}

	result, err := task.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("task result: %v\n", result)

	if result.ExitCode != 0 {
		t.Errorf("ping failed")
	}
}
func TestNetworkDisable(t *testing.T) {
	task := TaskInfo{
		Name:          "ibmcom/ping",
		EnableNetwork: false,
		Stderr:        os.Stderr,
	}
	task.Argments = []string{"ping", "-c", "5", "google.com"}

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

	task := TaskInfo{
		Name:      "ubuntu",
		PidsLimit: 100,
		WorkDir:   "/workdir",
		VolumeMountInfo: []VolumeMountInfo{
			{
				Path:   "/workdir",
				Volume: &volume,
			},
		},
		Timeout: 3 * time.Second,
	}
	task.Argments = []string{"./fork_bomb.sh"}

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

	compileTask := TaskInfo{
		Name:     "gcc:12.1",
		Argments: []string{"g++", "use_many_stack.cpp"},
		WorkDir:  "/workdir",
		VolumeMountInfo: []VolumeMountInfo{
			{
				Path:   "/workdir",
				Volume: &volume,
			},
		},
	}
	if _, err := compileTask.Run(); err != nil {
		t.Fatal(err)
	}

	task := TaskInfo{
		Name:     "gcc:12.1",
		Argments: []string{"./a.out"},
		WorkDir:  "/workdir",
		VolumeMountInfo: []VolumeMountInfo{
			{
				Path:   "/workdir",
				Volume: &volume,
			},
		},
		StackLimitKB: -1,
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
