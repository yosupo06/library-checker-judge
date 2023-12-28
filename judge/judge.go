package main

import (
	"log"
	"os"
	"path"
	"time"

	_ "github.com/lib/pq"
)

const (
	COMPILE_TIMEOUT = 30 * time.Second
	CHECKER_TIMEOUT = 30 * time.Second
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
	tl   float64
	lang Lang

	checkerVolume *Volume
	sourceVolume  *Volume

	caseDir *TestCaseDir
}

func NewJudge(lang Lang, tl float64, caseDir *TestCaseDir) (*Judge, error) {
	return &Judge{
		tl:      tl,
		lang:    lang,
		caseDir: caseDir,
	}, nil
}

func (j *Judge) Close() error {
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

func (j *Judge) TestCase(caseName string) (CaseResult, error) {
	log.Println("Start to judge case:", caseName)

	inFilePath := j.caseDir.InFilePath(caseName)
	expectFilePath := j.caseDir.OutFilePath(caseName)

	outFile, result, err := runSource(*j.sourceVolume, j.lang, j.tl, inFilePath)
	if err != nil {
		return CaseResult{}, err
	}
	defer os.ReadDir(outFile.Name())

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

	checkerResult, err := runChecker(j.checkerVolume, inFilePath, expectFilePath, outFile.Name())
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

func runSource(volume Volume, lang Lang, timeLimit float64, inFilePath string) (*os.File, TaskResult, error) {
	caseVolume, err := CreateVolume()
	if err != nil {
		return nil, TaskResult{}, err
	}
	defer func() {
		if err := caseVolume.Remove(); err != nil {
			log.Println("Failed to remove caseVolume:", err)
		}
	}()

	if err := caseVolume.CopyFile(inFilePath, "input.in"); err != nil {
		return nil, TaskResult{}, err
	}

	// TODO: make volume read only
	taskInfo, err := NewTaskInfo(lang.ImageName, append(
		DEFAULT_OPTIONS,
		WithArguments(append([]string{"library-checker-init", "/casedir/input.in", "/casedir/actual.out"}, lang.Exec...)...),
		WithWorkDir("/workdir"),
		WithVolume(&volume, "/workdir"),
		WithVolume(&caseVolume, "/casedir"),
		WithTimeout(time.Duration(timeLimit*1000*1000*1000)*time.Nanosecond),
	)...)
	if err != nil {
		return nil, TaskResult{}, err
	}

	result, err := taskInfo.Run()
	if err != nil {
		return nil, TaskResult{}, err
	}

	outFile, err := os.CreateTemp("", "")
	if err != nil {
		return nil, TaskResult{}, err
	}
	defer outFile.Close()

	// TODO: find faster way to copy actual.out
	genOutputFileTaskInfo, err := NewTaskInfo("ubuntu", append(
		DEFAULT_OPTIONS,
		WithArguments("cat", "/casedir/actual.out"),
		WithTimeout(COMPILE_TIMEOUT),
		WithVolume(&caseVolume, "/casedir"),
		WithStdout(outFile),
	)...)
	if err != nil {
		return nil, TaskResult{}, err
	}

	if _, err := genOutputFileTaskInfo.Run(); err != nil {
		return nil, TaskResult{}, err
	}

	return outFile, result, err
}

func runChecker(volume *Volume, inFilePath, expectFilePath, actualFilePath string) (TaskResult, error) {
	if err := volume.CopyFile(inFilePath, "input.in"); err != nil {
		return TaskResult{}, err
	}
	if err := volume.CopyFile(expectFilePath, "expect.out"); err != nil {
		return TaskResult{}, err
	}
	if err := volume.CopyFile(actualFilePath, "actual.out"); err != nil {
		return TaskResult{}, err
	}

	// TODO: make volume read only?
	checkerTaskInfo, err := NewTaskInfo(langs["checker"].ImageName, append(
		DEFAULT_OPTIONS,
		WithArguments(langs["checker"].Exec...),
		WithWorkDir("/workdir"),
		WithTimeout(CHECKER_TIMEOUT),
		WithVolume(volume, "/workdir"),
	)...)
	if err != nil {
		return TaskResult{}, err
	}

	return checkerTaskInfo.Run()
}
