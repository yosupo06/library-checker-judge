package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

type Volume struct {
	Name string
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

func (v *Volume) CopyFile(srcPath string, dstPath string) error {
	task := TaskInfo{
		WorkDir:       "/workdir",
		WorkDirVolume: v,
		Name:          "ubuntu",
	}
	container, err := task.create()
	defer container.Remove()

	if err != nil {
		return err
	}

	args := []string{"cp", srcPath, fmt.Sprintf("%s:%s", container.containerID, path.Join("/workdir", dstPath))}

	cmd := exec.Command("docker", args...)
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err != nil {
		log.Println("copy file failed:", err.Error())
		return err
	}

	return nil
}

func (v *Volume) Remove() error {
	args := []string{"volume", "rm", v.Name}

	cmd := exec.Command("docker", args...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Println("failed to remove volume:", err)
		return err
	}

	return nil
}

type BindInfo struct {
	HostPath      string
	ContainerPath string
}

type TaskInfo struct {
	Name                string // container name e.g. ubuntu
	Argments            []string
	Timeout             time.Duration
	Cpuset              []int
	MemoryLimitMB       int
	StackLimitKB        int // -1: unlimited
	PidsLimit           int
	EnableNetwork       bool
	EnableLoggingDriver bool
	WorkDir             string
	WorkDirVolume       *Volume
	Binds               []BindInfo

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type TaskResult struct {
	ExitCode int
	Time     time.Duration
	Memory   int64
	TLE      bool
}

func (t *TaskInfo) Run() (result TaskResult, err error) {
	ci, err := t.create()
	if err != nil {
		return TaskResult{}, err
	}
	defer func() {
		if err2 := ci.Remove(); err2 != nil {
			err = err2
		}
	}()

	result, err = t.start(ci)
	if err != nil {
		return TaskResult{}, err
	}
	return result, nil
}

// docker create ... -> container ID
func (t *TaskInfo) create() (containerInfo, error) {
	args := []string{"create"}

	// enable interactive
	args = append(args, "-i")

	args = append(args, "--init")

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

	// pids limit
	if t.PidsLimit != 0 {
		args = append(args, "--pids-limit")
		args = append(args, strconv.Itoa(t.PidsLimit))
	}

	// stack size
	if t.StackLimitKB != 0 {
		args = append(args, "--ulimit")
		args = append(args, fmt.Sprintf("stack=%d:%d", t.StackLimitKB, t.StackLimitKB))
	}

	// workdir
	if t.WorkDir != "" {
		args = append(args, "-w")
		args = append(args, t.WorkDir)
	}

	// volume
	if t.WorkDirVolume != nil {
		if t.WorkDir == "" {
			return containerInfo{}, errors.New("WorkDirVolume is specified though WorkDir is empty")
		}
		args = append(args, "-v")
		args = append(args, fmt.Sprintf("%s:%s", t.WorkDirVolume.Name, t.WorkDir))
	}

	// bind
	for _, bind := range t.Binds {
		args = append(args, "--mount")
		args = append(args, fmt.Sprintf("type=bind,source=%s,target=%s", bind.HostPath, bind.ContainerPath))
	}

	// container name
	args = append(args, t.Name)

	// extra arguments
	args = append(args, t.Argments...)

	cmd := exec.Command("docker", args...)
	log.Println("arg: ", args)

	cmd.Stderr = os.Stderr

	output, err := cmd.Output()

	if err != nil {
		log.Println("create failed:", err.Error())
		return containerInfo{}, err
	}

	containerId := strings.TrimSpace(string(output))

	return containerInfo{
		containerID: containerId,
	}, nil
}

func (t *TaskInfo) start(c containerInfo) (TaskResult, error) {
	log.Println("Start: ", t.Name, t.Argments)
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

	cmd := exec.CommandContext(ctx, "docker", args...)

	cmd.Stdin = t.Stdin
	cmd.Stdout = t.Stdout
	cmd.Stderr = t.Stderr

	result := TaskResult{
		Time:   -1,
		Memory: -1,
	}

	var start time.Time
	isFirst := true
	var end time.Time

	ticker := time.NewTicker(time.Millisecond)
	doneForChild := make(chan bool)
	doneForParent := make(chan bool)
	go func() {
		for {
			select {
			case <-doneForChild:
				doneForParent <- true
				return
			case <-ticker.C:
				tasks, err := c.readCGroupTasks()
				if err == nil && len(tasks) >= 2 {
					if isFirst {
						isFirst = false
						start = time.Now()
					}
					end = time.Now()
				}
				if usedMemory, err := c.readUsedMemory(); err == nil {
					if result.Memory < usedMemory {
						result.Memory = usedMemory
					}
				}
			}
		}
	}()
	err := cmd.Run()
	ticker.Stop()
	doneForChild <- true
	<-doneForParent

	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			log.Println("execute failed:", err.Error())
			return TaskResult{}, err
		}
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.Time = t.Timeout
		result.TLE = true
		result.ExitCode = 124

		// stop docker
		cmd := exec.Command("docker", "stop", c.containerID)
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {
		if result.ExitCode, err = inspectExitCode(c.containerID); err != nil {
			log.Println("failed to load exit code: ", err)
			return TaskResult{}, err
		}
		if usedTime, err := c.readUsedTime(); err != nil {
			log.Println("failed to load used time: ", err)
		} else {
			log.Println("Time: ", usedTime)
			result.Time = usedTime
		}
		log.Println("Time2: ", end.Sub(start))
		result.Time = end.Sub(start)
	}
	return result, nil
}

type containerInfo struct {
	containerID string
}

func (c *containerInfo) Remove() error {
	args := []string{"container", "rm", c.containerID}

	cmd := exec.Command("docker", args...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Println("failed to remove container:", err)
		return err
	}

	return nil
}

func (c *containerInfo) readUsedTime() (time.Duration, error) {
	args := []string{"inspect"}

	args = append(args, c.containerID)
	args = append(args, "--format={{.State.StartedAt}},{{.State.FinishedAt}}")

	cmd := exec.Command("docker", args...)

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	arr1 := strings.Split(strings.TrimSpace(string(output)), ",")

	start, err := time.Parse(time.RFC3339Nano, arr1[0])
	if err != nil {
		return 0, err
	}
	end, err := time.Parse(time.RFC3339Nano, arr1[1])
	if err != nil {
		return 0, err
	}

	return end.Sub(start), nil
}

func readCGroupTasksFromFile(filePath string) ([]string, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return []string{}, err
	}

	return strings.Split(strings.TrimSpace(string(bytes)), "\n"), nil
}

func (c *containerInfo) readCGroupTasks() ([]string, error) {
	filePathV1 := "/sys/fs/cgroup/cpu/docker/" + c.containerID + "/tasks"
	filePathV2 := "/sys/fs/cgroup/system.slice/docker-" + c.containerID + ".scope/container/cgroup.procs"

	if result, err := readCGroupTasksFromFile(filePathV1); err == nil {
		return result, nil
	}
	if result, err := readCGroupTasksFromFile(filePathV2); err == nil {
		return result, nil
	}

	return []string{}, errors.New("failed to load cgroup tasks")
}

func readUsedMemoryFromFile(filePath string) (int64, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	result, err := strconv.ParseInt(strings.TrimSpace(string(bytes)), 10, 64)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (c *containerInfo) readUsedMemory() (int64, error) {
	filePathV1 := "/sys/fs/cgroup/memory/docker/" + c.containerID + "/memory.max_usage_in_bytes"
	filePathV2 := "/sys/fs/cgroup/system.slice/docker-" + c.containerID + ".scope/container/memory.current"

	if result, err := readUsedMemoryFromFile(filePathV1); err == nil {
		return result, nil
	}
	if result, err := readUsedMemoryFromFile(filePathV2); err == nil {
		return result, nil
	}

	return 0, errors.New("failed to load memory usage")
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
