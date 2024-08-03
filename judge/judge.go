package main

import (
	"errors"
	"log"
	"os"
	"path"
	"time"

	"github.com/yosupo06/library-checker-judge/langs"
	"github.com/yosupo06/library-checker-judge/storage"
)

const (
	COMPILE_TIMEOUT   = 30 * time.Second
	CHECKER_TIMEOUT   = 30 * time.Second
	VERIFIER_TIMEOUT  = 30 * time.Second
	GENERATOR_TIMEOUT = 10 * time.Second
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

func compileChecker(dir storage.ProblemFiles) (Volume, TaskResult, error) {
	includeFiles, err := dir.IncludeFilePaths()
	if err != nil {
		return Volume{}, TaskResult{}, err
	}

	v, t, err := compile(dir.CheckerPath(), includeFiles, langs.LANG_CHECKER)
	if err != nil {
		return Volume{}, TaskResult{}, err
	}
	return v, t, nil
}

func compileVerifier(dir storage.ProblemFiles) (Volume, TaskResult, error) {
	includeFiles, err := dir.IncludeFilePaths()
	if err != nil {
		return Volume{}, TaskResult{}, err
	}

	v, t, err := compile(dir.VerifierPath(), includeFiles, langs.LANG_VERIFIER)
	if err != nil {
		return Volume{}, TaskResult{}, err
	}
	return v, t, nil
}

func compileSolution(dir storage.ProblemFiles) (Volume, TaskResult, error) {
	v, t, err := compile(dir.SolutionPath(), []string{}, langs.LANG_SOLUTION)
	if err != nil {
		return Volume{}, TaskResult{}, err
	}
	return v, t, err
}

func compileSource(dir storage.ProblemFiles, sourcePath string, lang langs.Lang) (Volume, TaskResult, error) {
	paths := []string{}
	for _, key := range lang.AdditionalFiles {
		paths = append(paths, dir.PublicFilePath(key))
	}
	v, t, err := compile(sourcePath, paths, lang)
	if err != nil {
		return Volume{}, TaskResult{}, err
	}
	return v, t, err
}

func compile(srcPath string, includePaths []string, lang langs.Lang) (v Volume, t TaskResult, err error) {
	log.Println("Compile:", srcPath, lang)
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

	if err = v.CopyFile(srcPath, lang.Source); err != nil {
		return
	}
	for _, p := range includePaths {
		if _, err = os.Stat(p); err == nil {
			if err = v.CopyFile(p, path.Base(p)); err != nil {
				return
			}
		} else if errors.Is(err, os.ErrNotExist) {
			log.Println(p, "is not found, skip")
		} else {
			return
		}
	}

	ti, err := NewTaskInfo(lang.ImageName, append(
		DEFAULT_OPTIONS,
		WithArguments(lang.Compile...),
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

func testCase(sourceVolume, checkerVolume Volume, lang langs.Lang, timeLimit float64, inFilePath, expectFilePath string) (CaseResult, error) {
	outFilePath, result, err := runSource(sourceVolume, lang, timeLimit, inFilePath)
	if err != nil {
		return CaseResult{}, err
	}
	defer os.Remove(outFilePath)

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

	checkerResult, err := runChecker(checkerVolume, inFilePath, expectFilePath, outFilePath)
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

func runSource(volume Volume, lang langs.Lang, timeLimit float64, inFilePath string) (string, TaskResult, error) {
	caseVolume, err := CreateVolume()
	if err != nil {
		return "", TaskResult{}, err
	}
	defer func() {
		if err := caseVolume.Remove(); err != nil {
			log.Println("Failed to remove caseVolume:", err)
		}
	}()

	if err := caseVolume.CopyFile(inFilePath, "input.in"); err != nil {
		return "", TaskResult{}, err
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
		return "", TaskResult{}, err
	}

	result, err := taskInfo.Run()
	if err != nil {
		return "", TaskResult{}, err
	}

	outFile, err := os.CreateTemp("", "")
	if err != nil {
		return "", TaskResult{}, err
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
		return "", TaskResult{}, err
	}

	if _, err := genOutputFileTaskInfo.Run(); err != nil {
		return "", TaskResult{}, err
	}

	return outFile.Name(), result, err
}

func runChecker(volume Volume, inFilePath, expectFilePath, actualFilePath string) (TaskResult, error) {
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
	checkerTaskInfo, err := NewTaskInfo(langs.LANG_CHECKER.ImageName, append(
		DEFAULT_OPTIONS,
		WithArguments(langs.LANG_CHECKER.Exec...),
		WithWorkDir("/workdir"),
		WithTimeout(CHECKER_TIMEOUT),
		WithVolume(&volume, "/workdir"),
	)...)
	if err != nil {
		return TaskResult{}, err
	}

	return checkerTaskInfo.Run()
}

func runGenerator(v Volume) (string, TaskResult, error) {
	outFile, err := os.CreateTemp("", "")
	if err != nil {
		return "", TaskResult{}, err
	}
	defer outFile.Close()

	ti, err := NewTaskInfo(langs.LANG_GENERATOR.ImageName, append(
		DEFAULT_OPTIONS,
		WithArguments(langs.LANG_GENERATOR.Exec...),
		WithWorkDir("/workdir"),
		WithTimeout(VERIFIER_TIMEOUT),
		WithVolume(&v, "/workdir"),
		WithStdout(outFile),
	)...)
	if err != nil {
		return "", TaskResult{}, err
	}

	result, err := ti.Run()
	if err != nil {
		return "", TaskResult{}, err
	}
	return outFile.Name(), result, nil
}
