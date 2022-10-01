package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const COMPILE_TIMEOUT = 30 * time.Second

type Judge struct {
	dir  string
	tl   float64
	lang Lang

	checkerVolume *Volume
	sourceVolume  *Volume
}

var defaultOptions = []TaskInfoOption{
	WithCpuset(0, 1),
	WithPidsLimit(100),
	WithStackLimitMB(-1),
	WithMemoryLimitMB(1024),
}

func NewJudge(judgedir string, lang Lang, tl float64) (*Judge, error) {
	tempdir, err := ioutil.TempDir(judgedir, "judge")
	if err != nil {
		return nil, err
	}
	log.Println("create judge dir:", tempdir)

	return &Judge{
		dir:  tempdir,
		tl:   tl,
		lang: lang,
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

func (j *Judge) CompileChecker(checkerFile io.Reader, includeFilePaths []string) (TaskResult, error) {
	volume, err := CreateVolume()
	if err != nil {
		return TaskResult{}, err
	}
	j.checkerVolume = &volume

	if err := volume.CopyFile(checkerFile, "checker.cpp"); err != nil {
		return TaskResult{}, err
	}

	for _, filePath := range includeFilePaths {
		file, err := os.Open(filePath)
		if err != nil {
			return TaskResult{}, err
		}
		defer file.Close()

		if err := volume.CopyFile(file, "testlib.h"); err != nil {
			return TaskResult{}, err
		}
	}

	taskInfo, err := NewTaskInfo(langs["checker"].ImageName, append(
		defaultOptions,
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

func (j *Judge) CompileSource(sourceFile io.Reader) (TaskResult, []byte, error) {
	// create dir for source
	volume, err := CreateVolume()
	if err != nil {
		return TaskResult{}, nil, err
	}
	j.sourceVolume = &volume

	if err := volume.CopyFile(sourceFile, j.lang.Source); err != nil {
		return TaskResult{}, nil, err
	}

	ceWriter, err := NewLimitedWriter(1 << 10)
	if err != nil {
		return TaskResult{}, nil, err
	}
	taskInfo, err := NewTaskInfo(j.lang.ImageName, append(
		defaultOptions,
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

func (j *Judge) createOutput(inFile io.Reader, outFilePath string) (TaskResult, error) {
	caseVolume, err := CreateVolume()
	if err != nil {
		return TaskResult{}, err
	}
	defer caseVolume.Remove()

	caseVolume.CopyFile(inFile, "input.in")

	// TODO: volume read only
	taskInfo, err := NewTaskInfo(j.lang.ImageName, append(
		defaultOptions,
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
		defaultOptions,
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

func (j *Judge) TestCase(inFile, expectFile io.Reader) (CaseResult, error) {
	outFile, err := ioutil.TempFile(j.dir, "output-")
	if err != nil {
		return CaseResult{}, err
	}
	defer os.Remove(outFile.Name())

	var inFile2 bytes.Buffer
	tee := io.TeeReader(inFile, &inFile2)

	result, err := j.createOutput(tee, outFile.Name())
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

	j.checkerVolume.CopyFile(&inFile2, "input.in")
	j.checkerVolume.CopyFile(expectFile, "expect.out")
	j.checkerVolume.CopyFile(outFile, "actual.out")

	// run checker
	checkerTaskInfo, err := NewTaskInfo(langs["checker"].ImageName, append(
		defaultOptions,
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
