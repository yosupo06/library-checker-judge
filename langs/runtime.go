package langs

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	MAX_STDERR_LENGTH = 1 << 10
)

type VolumeMountInfo struct {
	Path   string
	Volume *Volume
}

type BindMountInfo struct {
	HostPath      string
	ContainerPath string
	ReadOnly      bool
}

type TaskInfo struct {
	Name                string // container name e.g. ubuntu
	Argments            []string
	Timeout             time.Duration
	Cpuset              []int
	MemoryLimitMB       int
	StackLimitBytes     int // -1: unlimited
	PidsLimit           int
	EnableNetwork       bool
	EnableLoggingDriver bool
	WorkDir             string
	cgroupParent        string
	VolumeMountInfo     []VolumeMountInfo
	BindMountInfo       []BindMountInfo
	monitorBuilder      ContainerMonitorBuilder

	Stdin  io.Reader
	Stdout io.Writer
}

type TaskInfoOption func(*TaskInfo) error

func WithArguments(args ...string) TaskInfoOption {
	return func(ti *TaskInfo) error {
		ti.Argments = args
		return nil
	}
}

func WithTimeout(t time.Duration) TaskInfoOption {
	return func(ti *TaskInfo) error {
		ti.Timeout = t
		return nil
	}
}

func WithCpuset(cpus ...int) TaskInfoOption {
	return func(ti *TaskInfo) error {
		ti.Cpuset = cpus
		return nil
	}
}

func WithMemoryLimitMB(limitMB int) TaskInfoOption {
	return func(ti *TaskInfo) error {
		ti.MemoryLimitMB = limitMB
		return nil
	}
}

func WithStackLimitBytes(limitBytes int) TaskInfoOption {
	return func(ti *TaskInfo) error {
		ti.StackLimitBytes = limitBytes
		return nil
	}
}

func WithUnlimitedStackLimit() TaskInfoOption {
	return func(ti *TaskInfo) error {
		ti.StackLimitBytes = -1
		return nil
	}
}

// Convenience function for common stack sizes
func WithStackLimitMB(limitMB int) TaskInfoOption {
	return func(ti *TaskInfo) error {
		ti.StackLimitBytes = limitMB * 1024 * 1024 // Convert MB to bytes
		return nil
	}
}

func WithPidsLimit(n int) TaskInfoOption {
	return func(ti *TaskInfo) error {
		ti.PidsLimit = n
		return nil
	}
}

func WithWorkDir(path string) TaskInfoOption {
	return func(ti *TaskInfo) error {
		ti.WorkDir = path
		return nil
	}
}

func WithStdin(stdin io.Reader) TaskInfoOption {
	return func(ti *TaskInfo) error {
		ti.Stdin = stdin
		return nil
	}
}

func WithStdout(stdout io.Writer) TaskInfoOption {
	return func(ti *TaskInfo) error {
		ti.Stdout = stdout
		return nil
	}
}

func WithVolume(volume *Volume, containerPath string) TaskInfoOption {
	return func(ti *TaskInfo) error {
		ti.VolumeMountInfo = append(ti.VolumeMountInfo, VolumeMountInfo{
			Path:   containerPath,
			Volume: volume,
		})
		return nil
	}
}

func WithBindMount(hostPath string, containerPath string, readOnly bool) TaskInfoOption {
	return func(ti *TaskInfo) error {
		ti.BindMountInfo = append(ti.BindMountInfo, BindMountInfo{
			HostPath:      hostPath,
			ContainerPath: containerPath,
			ReadOnly:      readOnly,
		})
		return nil
	}
}

func WithMonitorBuilder(builder ContainerMonitorBuilder) TaskInfoOption {
	return func(ti *TaskInfo) error {
		ti.monitorBuilder = builder
		return nil
	}
}

func WithCgroupParent(cgroupParent string) TaskInfoOption {
	return func(ti *TaskInfo) error {
		ti.cgroupParent = cgroupParent
		return nil
	}
}

func NewTaskInfo(name string, ops ...TaskInfoOption) (*TaskInfo, error) {
	ti := &TaskInfo{Name: name}
	for _, option := range ops {
		if err := option(ti); err != nil {
			return nil, err
		}
	}
	return ti, nil
}

type TaskResult struct {
	ExitCode int
	Time     time.Duration
	Memory   int64
	TLE      bool
	Stderr   []byte
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
	if t.StackLimitBytes != 0 {
		args = append(args, "--ulimit")
		args = append(args, fmt.Sprintf("stack=%d:%d", t.StackLimitBytes, t.StackLimitBytes))
	}

	// workdir
	if t.WorkDir != "" {
		args = append(args, "-w")
		args = append(args, t.WorkDir)
	}

	// mount volume
	for _, volumeMount := range t.VolumeMountInfo {
		args = append(args, "-v")
		args = append(args, fmt.Sprintf("%s:%s", volumeMount.Volume.Name, volumeMount.Path))
	}

	// bind mount
	for _, bindMount := range t.BindMountInfo {
		args = append(args, "--mount")
		mountOpt := fmt.Sprintf("type=bind,src=%s,dst=%s", bindMount.HostPath, bindMount.ContainerPath)
		if bindMount.ReadOnly {
			mountOpt += ",readonly"
		}
		args = append(args, mountOpt)
	}

	// cgroup parent
	if t.cgroupParent != "" {
		args = append(args, fmt.Sprintf("--cgroup-parent=%s", t.cgroupParent))
	}

	// container name
	args = append(args, t.Name)

	// extra arguments
	args = append(args, t.Argments...)

	cmd := exec.Command("docker", args...)

	cmd.Stderr = os.Stderr

	output, err := cmd.Output()

	if err != nil {
		log.Println("create failed:", err.Error())
		return containerInfo{}, err
	}

	containerId := strings.TrimSpace(string(output))

	return containerInfo{
		containerID:  containerId,
		cgroupParent: t.cgroupParent,
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

	cmd := exec.CommandContext(ctx, "docker", args...)

	cmd.Stdin = t.Stdin
	cmd.Stdout = t.Stdout
	stderr := NewLimitedWriter(MAX_STDERR_LENGTH)
	cmd.Stderr = stderr

	monitorBuilder := t.monitorBuilder
	if monitorBuilder == nil {
		monitorBuilder = DEFAULT_MONITOR_BUILDER
	}
	cm, err := monitorBuilder(&c)
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			log.Println("create monitor failed:", err.Error())
			return TaskResult{}, err
		}
	}
	cm.start()
	err = cmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			log.Println("execute failed:", err.Error())
			return TaskResult{}, err
		}
	}
	cm.stop()

	if ctx.Err() == context.DeadlineExceeded {
		// stop docker
		cmd := exec.Command("docker", "stop", c.containerID)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Println("failed to stop docker:", err)
			return TaskResult{}, err
		}

		return TaskResult{
			Time:     t.Timeout,
			Memory:   cm.maxUsedMemory(),
			TLE:      true,
			ExitCode: 124,
		}, nil
	}

	usedTime := cm.usedTime()
	tle := false

	if t.Timeout != 0 && t.Timeout < usedTime {
		usedTime = t.Timeout
		tle = true
	}

	exitCode, err := inspectExitCode(c.containerID)
	if err != nil {
		log.Println("failed to load exit code: ", err)
		return TaskResult{}, err
	}

	return TaskResult{
		Time:     usedTime,
		Memory:   cm.maxUsedMemory(),
		TLE:      tle,
		ExitCode: exitCode,
		Stderr:   stderr.Bytes(),
	}, nil
}