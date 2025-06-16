package langs

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

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