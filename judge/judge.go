package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	_ "github.com/lib/pq"
)

const (
	COMPILE_TIMEOUT = 30 * time.Second
)

var DEFAULT_OPTIONS []TaskInfoOption

func init() {
	DEFAULT_OPTIONS = []TaskInfoOption{
		WithPidsLimit(100),
		WithStackLimitKB(-1),
		WithMemoryLimitMB(1024),
	}
	if c := os.Getenv("CGROUP_PARENT"); c != "" {
		DEFAULT_OPTIONS = append(DEFAULT_OPTIONS, WithCgroupParent(c))
	}
}

type Judge struct {
	dir  string
	tl   float64
	lang Lang

	checkerVolume *Volume
	sourceVolume  *Volume

	caseDir *TestCaseDir
}

func NewJudge(judgedir string, lang Lang, tl float64, caseDir *TestCaseDir) (*Judge, error) {
	tempdir, err := ioutil.TempDir(judgedir, "judge")
	if err != nil {
		return nil, err
	}
	log.Println("create judge dir:", tempdir)

	return &Judge{
		dir:     tempdir,
		tl:      tl,
		lang:    lang,
		caseDir: caseDir,
	}, nil
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

func (j *Judge) CompileChecker() (TaskResult, error) {
	paths := []string{}
	paths = append(paths, j.caseDir.CheckerPath())
	includeFilePaths, err := j.caseDir.IncludeFilePaths()
	if err != nil {
		return TaskResult{}, err
	}
	paths = append(paths, includeFilePaths...)
	v, t, err := compile(paths, langs["checker"].ImageName, langs["checker"].Compile)
	if err != nil {
		return TaskResult{}, err
	}
	j.checkerVolume = &v
	return t, err
}

func (j *Judge) CompileSource(sourcePath string) (TaskResult, error) {
	paths := []string{}
	paths = append(paths, sourcePath)
	for _, key := range j.lang.AdditionalFiles {
		paths = append(paths, j.caseDir.PublicFilePath(key))
	}
	v, t, err := compile(paths, j.lang.ImageName, j.lang.Compile)
	if err != nil {
		return TaskResult{}, err
	}
	j.sourceVolume = &v
	return t, err
}

func (j *Judge) createOutput(caseName string, outFilePath string) (TaskResult, error) {
	caseVolume, err := CreateVolume()
	if err != nil {
		return TaskResult{}, err
	}
	defer caseVolume.Remove()

	if err := caseVolume.CopyFile(j.caseDir.InFilePath(caseName), "input.in"); err != nil {
		return TaskResult{}, err
	}

	// TODO: volume read only
	taskInfo, err := NewTaskInfo(j.lang.ImageName, append(
		DEFAULT_OPTIONS,
		WithArguments(append([]string{"library-checker-init", "/casedir/input.in", "/casedir/actual.out"}, j.lang.Exec...)...),
		WithWorkDir("/workdir"),
		WithVolume(j.sourceVolume, "/workdir"),
		WithVolume(&caseVolume, "/casedir"),
		WithTimeout(time.Duration(j.tl*1000*1000*1000)*time.Nanosecond),
	)...)
	if err != nil {
		return TaskResult{}, err
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

	genOutputFileTaskInfo, err := NewTaskInfo("ubuntu", append(
		DEFAULT_OPTIONS,
		WithArguments("cat", "/casedir/actual.out"),
		WithTimeout(COMPILE_TIMEOUT),
		WithVolume(&caseVolume, "/casedir"),
		WithStdout(outFile),
	)...)
	if err != nil {
		return TaskResult{}, err
	}

	_, err = genOutputFileTaskInfo.Run()
	if err != nil {
		return TaskResult{}, err
	}

	return result, err
}

func (j *Judge) TestCase(caseName string) (CaseResult, error) {
	log.Println("start to judge case:", caseName)
	outFile, err := ioutil.TempFile(j.dir, "output-")
	if err != nil {
		return CaseResult{}, err
	}
	defer os.Remove(outFile.Name())

	result, err := j.createOutput(caseName, outFile.Name())
	if err != nil {
		return CaseResult{}, err
	}

	baseResult := CaseResult{Time: result.Time, Memory: result.Memory, TLE: result.TLE, Stderr: result.Stderr, CheckerOut: []byte{}}
	if result.TLE {
		//timeout
		baseResult.Status = "TLE"
		return baseResult, nil
	}

	if result.ExitCode != 0 {
		//runtime error
		baseResult.Status = "RE"
		return baseResult, nil
	}

	if err := j.checkerVolume.CopyFile(j.caseDir.InFilePath(caseName), "input.in"); err != nil {
		return CaseResult{}, err
	}
	if err := j.checkerVolume.CopyFile(j.caseDir.OutFilePath(caseName), "expect.out"); err != nil {
		return CaseResult{}, err
	}
	if err := j.checkerVolume.CopyFile(outFile.Name(), "actual.out"); err != nil {
		return CaseResult{}, err
	}

	// run checker
	checkerTaskInfo, err := NewTaskInfo(langs["checker"].ImageName, append(
		DEFAULT_OPTIONS,
		WithArguments(langs["checker"].Exec...),
		WithWorkDir("/workdir"),
		WithTimeout(COMPILE_TIMEOUT),
		WithVolume(j.checkerVolume, "/workdir"),
	)...)
	if err != nil {
		return CaseResult{}, err
	}

	checkerResult, err := checkerTaskInfo.Run()
	if err != nil {
		return CaseResult{}, err
	}

	baseResult.CheckerOut = checkerResult.Stderr

	if checkerResult.TLE {
		baseResult.Status = "ITLE"
	} else if checkerResult.ExitCode == 1 {
		baseResult.Status = "WA"
	} else if checkerResult.ExitCode == 2 {
		baseResult.Status = "PE"
	} else if checkerResult.ExitCode == 3 {
		baseResult.Status = "Fail"
	} else if checkerResult.ExitCode != 0 {
		baseResult.Status = "Unknown"
	} else {
		baseResult.Status = "AC"
	}
	return baseResult, nil
}

type CaseResult struct {
	CaseName   string
	Status     string
	Time       time.Duration
	Memory     int64
	TLE        bool
	Stderr     []byte
	CheckerOut []byte
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

func compile(srcPaths []string, imageName string, cmd []string) (v Volume, t TaskResult, err error) {
	log.Println("Compile:", srcPaths, imageName, cmd)
	v, err = CreateVolume()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			if err := v.Remove(); err != nil {
				log.Println("Volume remove failed:", err)
			}
		}
	}()

	for _, p := range srcPaths {
		if err = v.CopyFile(p, path.Base(p)); err != nil {
			return
		}
	}

	ti, err := NewTaskInfo(imageName, append(
		DEFAULT_OPTIONS,
		WithArguments(cmd...),
		WithWorkDir("/workdir"),
		WithVolume(&v, "/workdir"),
		WithTimeout(COMPILE_TIMEOUT),
	)...)
	if err != nil {
		return
	}
	t, err = ti.Run()
	return
}
