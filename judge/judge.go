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
	COMPILE_TIMEOUT    = 30 * time.Second
	MAX_MESSAGE_LENGTH = 1 << 10
)

type Judge struct {
	dir          string
	tl           float64
	lang         Lang
	cgroupParent string

	checkerVolume *Volume
	sourceVolume  *Volume

	caseDir *TestCaseDir
}

func NewJudge(judgedir string, lang Lang, tl float64, cgroupParent string, caseDir *TestCaseDir) (*Judge, error) {
	tempdir, err := ioutil.TempDir(judgedir, "judge")
	if err != nil {
		return nil, err
	}
	log.Println("create judge dir:", tempdir)

	return &Judge{
		dir:          tempdir,
		tl:           tl,
		lang:         lang,
		cgroupParent: cgroupParent,
		caseDir:      caseDir,
	}, nil
}

func (j *Judge) defaultOptions() []TaskInfoOption {
	options := []TaskInfoOption{
		WithPidsLimit(100),
		WithStackLimitKB(-1),
		WithMemoryLimitMB(1024),
	}
	if j.cgroupParent != "" {
		options = append(options, WithCgroupParent(j.cgroupParent))
	}
	return options
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
	volume, err := CreateVolume()
	if err != nil {
		return TaskResult{}, err
	}
	j.checkerVolume = &volume

	if err := volume.CopyFile(j.caseDir.CheckerPath(), "checker.cpp"); err != nil {
		return TaskResult{}, err
	}

	includeFilePaths, err := j.caseDir.IncludeFilePaths()
	if err != nil {
		return TaskResult{}, err
	}
	for _, filePath := range includeFilePaths {
		if err := volume.CopyFile(filePath, path.Base(filePath)); err != nil {
			return TaskResult{}, err
		}
	}

	taskInfo, err := NewTaskInfo(langs["checker"].ImageName, append(
		j.defaultOptions(),
		WithArguments(langs["checker"].Compile...),
		WithWorkDir("/workdir"),
		WithVolume(&volume, "/workdir"),
		WithTimeout(COMPILE_TIMEOUT),
		WithStdin(os.Stdout),
		WithStderr(os.Stderr),
	)...)
	if err != nil {
		return TaskResult{}, err
	}
	result, err := taskInfo.Run()
	if err != nil {
		return TaskResult{}, err
	}

	return result, err
}

func (j *Judge) CompileSource(sourcePath string) (TaskResult, []byte, error) {
	// create dir for source
	volume, err := CreateVolume()
	if err != nil {
		return TaskResult{}, nil, err
	}
	j.sourceVolume = &volume

	if err := volume.CopyFile(sourcePath, j.lang.Source); err != nil {
		return TaskResult{}, nil, err
	}

	for _, key := range j.lang.AdditionalFiles {
		filePath := j.caseDir.PublicFilePath(key)
		if _, err := os.Stat(filePath); err != nil {
			continue
		}
		if err := volume.CopyFile(j.caseDir.PublicFilePath(key), path.Base(key)); err != nil {
			return TaskResult{}, nil, err
		}
	}

	ceWriter, err := NewLimitedWriter(MAX_MESSAGE_LENGTH)
	if err != nil {
		return TaskResult{}, nil, err
	}
	taskInfo, err := NewTaskInfo(j.lang.ImageName, append(
		j.defaultOptions(),
		WithArguments(j.lang.Compile...),
		WithWorkDir("/workdir"),
		WithVolume(&volume, "/workdir"),
		WithTimeout(COMPILE_TIMEOUT),
		WithStderr(ceWriter),
	)...)
	if err != nil {
		return TaskResult{}, nil, err
	}
	result, err := taskInfo.Run()
	if err != nil {
		return TaskResult{}, nil, err
	}

	return result, ceWriter.Bytes(), nil
}

func (j *Judge) createOutput(caseName string, outFilePath string) (TaskResult, []byte, error) {
	caseVolume, err := CreateVolume()
	if err != nil {
		return TaskResult{}, nil, err
	}
	defer caseVolume.Remove()

	if err := caseVolume.CopyFile(j.caseDir.InFilePath(caseName), "input.in"); err != nil {
		return TaskResult{}, nil, err
	}

	stderrWriter, err := NewLimitedWriter(MAX_MESSAGE_LENGTH)
	if err != nil {
		return TaskResult{}, nil, err
	}
	// TODO: volume read only
	taskInfo, err := NewTaskInfo(j.lang.ImageName, append(
		j.defaultOptions(),
		WithArguments(append([]string{"library-checker-init", "/casedir/input.in", "/casedir/actual.out"}, j.lang.Exec...)...),
		WithWorkDir("/workdir"),
		WithVolume(j.sourceVolume, "/workdir"),
		WithVolume(&caseVolume, "/casedir"),
		WithTimeout(time.Duration(j.tl*1000*1000*1000)*time.Nanosecond),
		WithStderr(stderrWriter),
	)...)
	if err != nil {
		return TaskResult{}, nil, err
	}

	result, err := taskInfo.Run()
	if err != nil {
		return TaskResult{}, nil, err
	}

	outFile, err := os.Create(outFilePath)
	if err != nil {
		return TaskResult{}, nil, err
	}
	defer outFile.Close()

	genOutputFileTaskInfo, err := NewTaskInfo("ubuntu", append(
		j.defaultOptions(),
		WithArguments("cat", "/casedir/actual.out"),
		WithTimeout(COMPILE_TIMEOUT),
		WithVolume(&caseVolume, "/casedir"),
		WithStdout(outFile),
	)...)
	if err != nil {
		return TaskResult{}, nil, err
	}

	_, err = genOutputFileTaskInfo.Run()
	if err != nil {
		return TaskResult{}, nil, err
	}

	return result, stderrWriter.Bytes(), err
}

func (j *Judge) TestCase(caseName string) (CaseResult, error) {
	log.Println("start to judge case:", caseName)
	outFile, err := ioutil.TempFile(j.dir, "output-")
	if err != nil {
		return CaseResult{}, err
	}
	defer os.Remove(outFile.Name())

	result, stderr, err := j.createOutput(caseName, outFile.Name())
	if err != nil {
		return CaseResult{}, err
	}
	checkerOutWriter, err := NewLimitedWriter(MAX_MESSAGE_LENGTH)
	if err != nil {
		return CaseResult{}, err
	}

	baseResult := CaseResult{Time: result.Time, Memory: result.Memory, TLE: result.TLE, Stderr: stderr, CheckerOut: []byte{}}
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
		j.defaultOptions(),
		WithArguments(langs["checker"].Exec...),
		WithWorkDir("/workdir"),
		WithTimeout(COMPILE_TIMEOUT),
		WithVolume(j.checkerVolume, "/workdir"),
		WithStderr(checkerOutWriter),
	)...)
	if err != nil {
		return CaseResult{}, err
	}

	checkerResult, err := checkerTaskInfo.Run()
	if err != nil {
		return CaseResult{}, err
	}

	baseResult.CheckerOut = checkerOutWriter.Bytes()

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
