package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
)

const COMPILE_TIMEOUT = 30 * time.Second
const MEMORY_LIMIT_MB = 1024
const PIDS_LIMIT = 100

type Judge struct {
	dir  string
	tl   float64
	lang Lang

	checkerVolume *Volume
	sourceVolume  *Volume
}

func NewJudge(judgedir, lang string, tl float64) (*Judge, error) {
	tempdir, err := ioutil.TempDir(judgedir, "judge")
	if err != nil {
		return nil, err
	}

	log.Println("create judge dir:", tempdir)

	judge := new(Judge)

	judge.lang = langs[lang]
	judge.tl = tl
	judge.dir = tempdir

	return judge, nil
}

func (j *Judge) Close() error {
	if err := os.RemoveAll(j.dir); err != nil {
		return err
	}
	if j.checkerVolume != nil {
		if err := j.checkerVolume.Remove(); err != nil {
			return err
		}
	}
	if j.sourceVolume != nil {
		if err := j.sourceVolume.Remove(); err != nil {
			return err
		}
	}
	return nil
}

func (j *Judge) CompileChecker(checkerPath string) (TaskResult, error) {
	volume, err := CreateVolume()
	if err != nil {
		return TaskResult{}, err
	}
	j.checkerVolume = &volume

	if err := volume.CopyFile(checkerPath, ""); err != nil {
		return TaskResult{}, err
	}

	testlibPath, err := filepath.Abs("testlib.h")
	if err != nil {
		return TaskResult{}, err
	}
	if err := volume.CopyFile(testlibPath, ""); err != nil {
		return TaskResult{}, err
	}

	taskInfo := TaskInfo{
		Name:          langs["checker"].ImageName,
		Argments:      langs["checker"].Compile,
		WorkDir:       "/workdir",
		WorkDirVolume: &volume,
		Stdout:        os.Stdout,
		Stderr:        os.Stderr,
		Timeout:       COMPILE_TIMEOUT,
		PidsLimit:     PIDS_LIMIT,
		MemoryLimitMB: MEMORY_LIMIT_MB,
	}
	result, err := taskInfo.Run()
	if err != nil {
		return TaskResult{}, err
	}

	return result, err
}

func (j *Judge) CompileSource(sourcePath string) (TaskResult, error) {
	// create dir for source
	volume, err := CreateVolume()
	if err != nil {
		return TaskResult{}, err
	}
	j.sourceVolume = &volume

	if err := volume.CopyFile(sourcePath, j.lang.Source); err != nil {
		return TaskResult{}, err
	}

	taskInfo := TaskInfo{
		Name:          j.lang.ImageName,
		Argments:      j.lang.Compile,
		WorkDir:       "/workdir",
		WorkDirVolume: &volume,
		Timeout:       COMPILE_TIMEOUT,
		Stderr:        os.Stderr,
		PidsLimit:     PIDS_LIMIT,
		MemoryLimitMB: MEMORY_LIMIT_MB,
	}
	result, err := taskInfo.Run()
	if err != nil {
		return TaskResult{}, err
	}

	return result, err
}

func fileCopy(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}
	return nil
}

func (j *Judge) createOutput(inFilePath string, outFilePath string) (TaskResult, error) {
	casedir, err := ioutil.TempDir(j.dir, "judge")
	if err != nil {
		return TaskResult{}, err
	}
	defer os.RemoveAll(casedir)

	fileCopy(inFilePath, filepath.Join(casedir, "input.in"))

	// TODO: volume read only
	taskInfo := TaskInfo{
		Name:          j.lang.ImageName,
		Argments:      append([]string{"library-checker-init", "/casedir/input.in", "/casedir/actual.out"}, j.lang.Exec...),
		WorkDir:       "/workdir",
		WorkDirVolume: j.sourceVolume,
		Timeout:       time.Duration(j.tl*1000*1000*1000) * time.Nanosecond,
		PidsLimit:     PIDS_LIMIT,
		MemoryLimitMB: MEMORY_LIMIT_MB,
		Binds: []BindInfo{
			{
				HostPath:      casedir,
				ContainerPath: "/casedir",
			},
		},
	}

	result, err := taskInfo.Run()
	if err != nil {
		return TaskResult{}, err
	}

	outFile, err := os.Create(outFilePath)
	if err != nil {
		return TaskResult{}, err
	}
	defer outFile.Close()

	genOutputFileTaskInfo := TaskInfo{
		Name:          "ubuntu",
		Argments:      []string{"cat", "/casedir/actual.out"},
		Timeout:       time.Duration(j.tl*1000*1000*1000) * time.Nanosecond,
		PidsLimit:     PIDS_LIMIT,
		MemoryLimitMB: MEMORY_LIMIT_MB,
		Binds: []BindInfo{
			{
				HostPath:      casedir,
				ContainerPath: "/casedir",
			},
		},
		Stdout: outFile,
	}

	_, err = genOutputFileTaskInfo.Run()
	if err != nil {
		return TaskResult{}, err
	}

	return result, err
}

func (j *Judge) TestCase(inFilePath string, expectFilePath string) (CaseResult, error) {
	outFilePath, err := ioutil.TempFile(j.dir, "output-")
	if err != nil {
		return CaseResult{}, err
	}
	defer os.Remove(outFilePath.Name())

	result, err := j.createOutput(inFilePath, outFilePath.Name())
	if err != nil {
		return CaseResult{}, err
	}
	if result.TLE {
		//timeout
		return CaseResult{Status: "TLE", Time: result.Time, Memory: result.Memory, TLE: result.TLE}, nil
	}

	if result.ExitCode != 0 {
		//runtime error
		return CaseResult{Status: "RE", Time: result.Time, Memory: result.Memory, TLE: result.TLE}, nil
	}

	j.checkerVolume.CopyFile(inFilePath, "input.in")
	j.checkerVolume.CopyFile(expectFilePath, "expect.out")
	j.checkerVolume.CopyFile(outFilePath.Name(), "actual.out")

	// run checker
	checkerTaskInfo := TaskInfo{
		Name:          langs["checker"].ImageName,
		Argments:      langs["checker"].Exec,
		WorkDir:       "/workdir",
		WorkDirVolume: j.checkerVolume,
		Timeout:       COMPILE_TIMEOUT,
		PidsLimit:     PIDS_LIMIT,
		MemoryLimitMB: MEMORY_LIMIT_MB,
	}

	checkerResult, err := checkerTaskInfo.Run()
	if err != nil {
		return CaseResult{}, err
	}
	if checkerResult.TLE {
		return CaseResult{Status: "ITLE", Time: result.Time, Memory: result.Memory, TLE: result.TLE}, nil
	}
	if checkerResult.ExitCode == 1 {
		return CaseResult{Status: "WA", Time: result.Time, Memory: result.Memory, TLE: result.TLE}, nil
	}
	if checkerResult.ExitCode == 2 {
		return CaseResult{Status: "PE", Time: result.Time, Memory: result.Memory, TLE: result.TLE}, nil
	}
	if checkerResult.ExitCode == 3 {
		return CaseResult{Status: "Fail", Time: result.Time, Memory: result.Memory, TLE: result.TLE}, nil
	}
	if checkerResult.ExitCode != 0 {
		return CaseResult{Status: "Unknown", Time: result.Time, Memory: result.Memory, TLE: result.TLE}, nil
	}
	return CaseResult{Status: "AC", Time: result.Time, Memory: result.Memory, TLE: result.TLE}, nil
}

type CaseResult struct {
	CaseName string
	Status   string
	Time     time.Duration
	Memory   int64
	TLE      bool
}

func AggregateResults(results []CaseResult) CaseResult {
	ans := CaseResult{
		Status: "AC",
		Time:   0,
		Memory: -1,
	}
	for _, res := range results {
		if res.Status != "AC" {
			ans.Status = res.Status
		}
		if ans.Time < res.Time {
			ans.Time = res.Time
		}
		if ans.Memory < res.Memory {
			ans.Memory = res.Memory
		}
	}
	return ans
}
