package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

type Volume struct {
	Name string
}

type TaskInfo struct {
	Name                string // container name e.g. ubuntu
	Argments            []string
	Timeout             time.Duration
	Cpuset              []int
	MemoryLimitMB       int
	EnableNetwork       bool
	EnableLoggingDriver bool

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type containerInfo struct {
	containerID string
	cgroupName  string
}

type TaskResult struct {
	ExitCode int
	Time     time.Duration
	Memory   int64
	TLE      bool
}

func CreateVolume() (Volume, error) {
	volumeName := "volume-" + uuid.New().String()

	args := []string{"volume", "create"}
	args = append(args, "--name", volumeName)

	cmd := exec.Command("docker", args...)

	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		log.Println("volume create failed:", err.Error())
		return Volume{}, err
	}

	return Volume{
		Name: volumeName,
	}, nil
}

func (v *Volume) CopyFile(file *io.Reader, dstPath string) (Volume, error) {
	volumeName := "volume-" + uuid.New().String()

	args := []string{"volume", "create"}
	args = append(args, "--name", volumeName)

	cmd := exec.Command("docker", args...)

	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		log.Println("volume create failed:", err.Error())
		return Volume{}, err
	}

	return Volume{
		Name: volumeName,
	}, nil
}

func (t *TaskInfo) Run() (TaskResult, error) {
	ci, err := t.create()
	if err != nil {
		return TaskResult{}, err
	}
	result, err := t.start(ci)
	if err != nil {
		return TaskResult{}, err
	}
	return result, nil
}

func readUsedTime(cgroupName string) (time.Duration, error) {
	fileName := "/sys/fs/cgroup/cpuacct/" + cgroupName + "/cpuacct.usage"
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return 0, err
	}

	result, err := strconv.ParseInt(strings.TrimSpace(string(bytes)), 10, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(result) * time.Nanosecond, nil
}

func readUsedMemory(cgroupName string) (int64, error) {
	fileName := "/sys/fs/cgroup/memory/" + cgroupName + "/memory.max_usage_in_bytes"
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return 0, err
	}

	result, err := strconv.ParseInt(strings.TrimSpace(string(bytes)), 10, 64)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func inspectExitCode(containerId string) (int, error) {
	args := []string{"inspect"}

	args = append(args, containerId)
	args = append(args, "--format={{.State.ExitCode}}")

	cmd := exec.Command("docker", args...)

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	code, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 32)
	if err != nil {
		return 0, err
	}

	return int(code), nil
}

// docker create ... -> container ID
func (t *TaskInfo) create() (containerInfo, error) {
	cgroupName := "container-" + uuid.New().String()

	args := []string{"create"}

	// enable interactive
	args = append(args, "-i")

	// cgroup
	args = append(args, "--cgroup-parent="+cgroupName)

	// cpuset
	if len(t.Cpuset) != 0 {
		cpus := []string{}
		for c := range t.Cpuset {
			cpus = append(cpus, strconv.Itoa(c))
		}
		args = append(args, "--cpuset-cpus="+strings.Join(cpus, ","))
	}

	// network
	if !t.EnableNetwork {
		args = append(args, "--net=none")
	}

	// logging driver
	if !t.EnableLoggingDriver {
		args = append(args, "--log-driver=none")
	}

	// memory limit
	if t.MemoryLimitMB != 0 {
		args = append(args, fmt.Sprintf("--memory=%dm", t.MemoryLimitMB))
		args = append(args, fmt.Sprintf("--memory-swap=%dm", t.MemoryLimitMB))
	}

	// container name
	args = append(args, t.Name)

	// extra arguments
	args = append(args, t.Argments...)

	log.Printf("create docker args:%s\n", args)

	cmd := exec.Command("docker", args...)

	cmd.Stderr = os.Stderr

	output, err := cmd.Output()

	if err != nil {
		log.Println("create failed:", err.Error())
		return containerInfo{}, err
	}

	containerId := strings.TrimSpace(string(output))

	return containerInfo{
		containerID: containerId,
		cgroupName:  cgroupName,
	}, nil
}

func (t *TaskInfo) start(c containerInfo) (TaskResult, error) {
	ctx := context.Background()
	if t.Timeout != 0 {
		ctx2, cancel := context.WithTimeout(context.Background(), t.Timeout+500*time.Millisecond)
		ctx = ctx2
		defer cancel()
	}

	args := []string{"start"}

	// enable interactive
	args = append(args, "-i")

	args = append(args, c.containerID)

	log.Printf("execute docker args:%s\n", args)

	cmd := exec.CommandContext(ctx, "docker", args...)

	cmd.Stdin = t.Stdin
	cmd.Stdout = t.Stdout
	cmd.Stderr = t.Stderr

	err := cmd.Run()

	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			log.Println("execute failed:", err.Error())
			return TaskResult{}, err
		}
	}

	result := TaskResult{
		Time:   -1,
		Memory: -1,
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.Time = t.Timeout
		result.TLE = true
		result.ExitCode = 124
	} else {
		if result.ExitCode, err = inspectExitCode(c.containerID); err != nil {
			log.Println("failed to load exit code: ", err)
			return TaskResult{}, err
		}
		if usedTime, err := readUsedTime(c.cgroupName); err != nil {
			log.Println("failed to load used time: ", err)
		} else {
			result.Time = usedTime
		}
	}

	if usedMemory, err := readUsedMemory(c.cgroupName); err != nil {
		log.Println("failed to load used memory: ", err)
	} else {
		result.Memory = usedMemory
	}

	return result, nil
}
