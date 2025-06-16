package executor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MAX_STDERR_LENGTH = 1 << 10
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
	log.Printf("Copy file to %v:%v", v.Name, dstPath)

	task := TaskInfo{
		VolumeMountInfo: []VolumeMountInfo{
			{
				Path:   "/workdir",
				Volume: v,
			},
		},
		Name: "ubuntu",
	}
	ci, err := task.create()
	if err != nil {
		return err
	}
	defer func() {
		if err := ci.Remove(); err != nil {
			log.Printf("Failed to remove container: %v", err)
		}
	}()

	return ci.CopyFile(srcPath, path.Join("/workdir", dstPath))
}

func (v *Volume) Remove() error {
	args := []string{"volume", "rm", v.Name}

	cmd := exec.Command("docker", args...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

type VolumeMountInfo struct {
	Path   string
	Volume *Volume
}

type BindMountInfo struct {
	HostPath      string
	ContainerPath string
	ReadOnly      bool
}

type ContainerMonitorBuilder func(c *containerInfo) (containerMonitor, error)

var DEFAULT_MONITOR_BUILDER ContainerMonitorBuilder

func init() {
	if _, ok := os.LookupEnv("LIBRARY_CHECKER_JUDGE"); ok {
		log.Println("Started in judge server, use HighPrecisionContainerMonitor")
		DEFAULT_MONITOR_BUILDER = NewHighPrecisionContainerMonitor
	} else {
		log.Println("Started in local, use LowPrecisionContainerMonitor")
		DEFAULT_MONITOR_BUILDER = NewLowPrecisionContainerMonitor
	}
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

type containerMonitor interface {
	start()
	stop()

	usedTime() time.Duration
	maxUsedMemory() int64
}

// A highPrecisionContainerMonitor measures used time in high precision.
// Because this monitor uses hacky trick it won't work in all environments.
type highPrecisionContainerMonitor struct {
	c *containerInfo

	ticker        *time.Ticker
	doneForChild  chan bool
	doneForParent chan bool

	isStarted bool
	startTime time.Time
	endTime   time.Time

	maxMemory int64
}

func NewHighPrecisionContainerMonitor(c *containerInfo) (containerMonitor, error) {
	cm := highPrecisionContainerMonitor{
		c:             c,
		isStarted:     false,
		ticker:        time.NewTicker(time.Millisecond),
		doneForChild:  make(chan bool),
		doneForParent: make(chan bool),
	}

	return &cm, nil
}

func (cm *highPrecisionContainerMonitor) start() {
	go func() {
		for {
			select {
			case <-cm.doneForChild:
				cm.doneForParent <- true
				return
			case <-cm.ticker.C:
				tasks, err := cm.c.readCGroupTasks()
				if err == nil && len(tasks) >= 2 {
					if !cm.isStarted {
						cm.isStarted = true
						cm.startTime = time.Now()
					}
					cm.endTime = time.Now()
				}
				if usedMemory, err := cm.c.readUsedMemory(); err == nil {
					if cm.maxMemory < usedMemory {
						cm.maxMemory = usedMemory
					}
				}
			}
		}
	}()
}

func (cm *highPrecisionContainerMonitor) stop() {
	cm.ticker.Stop()
	cm.doneForChild <- true
	<-cm.doneForParent
}

func (cm *highPrecisionContainerMonitor) usedTime() time.Duration {
	return cm.endTime.Sub(cm.startTime)
}

func (cm *highPrecisionContainerMonitor) maxUsedMemory() int64 {
	return cm.maxMemory
}

type lowPrecisionContainerMonitor struct {
	c   *containerInfo
	hcm containerMonitor
}

func NewLowPrecisionContainerMonitor(c *containerInfo) (containerMonitor, error) {
	hcm, _ := NewHighPrecisionContainerMonitor(c)
	cm := lowPrecisionContainerMonitor{
		c:   c,
		hcm: hcm,
	}

	return &cm, nil
}
func (cm *lowPrecisionContainerMonitor) start() {
	cm.hcm.start()
}
func (cm *lowPrecisionContainerMonitor) stop() {
	cm.hcm.stop()
}
func (cm *lowPrecisionContainerMonitor) usedTime() time.Duration {
	var startedAt, finishedAt time.Time

	output, err := readInspect(cm.c.containerID, "--format={{.State.StartedAt}}")
	if err != nil {
		return 0
	}
	startedAt, err = cm.parseDate(output)
	if err != nil {
		return 0
	}
	output, err = readInspect(cm.c.containerID, "--format={{.State.FinishedAt}}")
	if err != nil {
		return 0
	}
	finishedAt, err = cm.parseDate(output)
	if err != nil {
		return 0
	}
	return finishedAt.Sub(startedAt)
}

func (cm *lowPrecisionContainerMonitor) maxUsedMemory() int64 {
	return cm.hcm.maxUsedMemory()
}

func (cm *lowPrecisionContainerMonitor) parseDate(output []byte) (time.Time, error) {
	date, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(string(output)))
	if err != nil {
		log.Println("failed to parse date: ", err.Error())
		return time.Unix(0, 0), err
	}
	return date, nil
}

type containerInfo struct {
	containerID  string
	cgroupParent string
}

func (c *containerInfo) Remove() error {
	args := []string{"container", "rm", c.containerID}

	cmd := exec.Command("docker", args...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func readCGroupTasksFromFile(filePath string) ([]string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return []string{}, err
	}

	return strings.Split(strings.TrimSpace(string(bytes)), "\n"), nil
}

func (c *containerInfo) CopyFile(src string, dst string) error {
	args := []string{"cp", src, c.containerID + ":" + dst}

	cmd := exec.Command("docker", args...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (c *containerInfo) cgroupDirs() []string {
	cgroupParent := c.cgroupParent
	if cgroupParent == "" {
		cgroupParent = "system.slice"
	}

	return []string{
		// cgroup-v2 cgroupdriver=systemd
		path.Join("/sys/fs/cgroup", cgroupParent, "docker-"+c.containerID+".scope", "container"),
		// cgroup-v2 cgroupdriver=cgroupfs
		path.Join("/sys/fs/cgroup", cgroupParent, c.containerID),
	}
}

func (c *containerInfo) readCGroupTasks() ([]string, error) {
	for _, dir := range c.cgroupDirs() {
		if result, err := readCGroupTasksFromFile(path.Join(dir, "cgroup.procs")); err == nil {
			return result, nil
		}
	}

	return []string{}, errors.New("failed to load cgroup tasks")
}

func readUsedMemoryFromFile(filePath string) (int64, error) {
	bytes, err := os.ReadFile(filePath)
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
	for _, dir := range c.cgroupDirs() {
		if result, err := readUsedMemoryFromFile(path.Join(dir, "memory.current")); err == nil {
			return result, nil
		}
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

func readInspect(containerId string, args ...string) ([]byte, error) {
	args = append([]string{
		"inspect",
		containerId,
	}, args...)
	cmd := exec.Command("docker", args...)
	output, err := cmd.Output()
	if err != nil {
		log.Println("failed to read inspect:", err.Error())
	}
	return output, err
}