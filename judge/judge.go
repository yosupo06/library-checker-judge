package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/yosupo06/library-checker-judge/executor"
	"github.com/yosupo06/library-checker-judge/langs"
	"github.com/yosupo06/library-checker-judge/storage"
)

const (
	DEFAULT_PID_LIMIT       = 100
	DEFAULT_MEMORY_LIMIT_MB = 1024
	COMPILE_TIMEOUT         = 30 * time.Second
	CHECKER_TIMEOUT         = 10 * time.Second
	VERIFIER_TIMEOUT        = 10 * time.Second
	GENERATOR_TIMEOUT       = 10 * time.Second
)

var DEFAULT_OPTIONS []executor.TaskInfoOption

func init() {
	DEFAULT_OPTIONS = []executor.TaskInfoOption{
		executor.WithPidsLimit(DEFAULT_PID_LIMIT),
		executor.WithUnlimitedStackLimit(),
		executor.WithMemoryLimitMB(DEFAULT_MEMORY_LIMIT_MB),
	}
	if c := os.Getenv("CGROUP_PARENT"); c != "" {
		DEFAULT_OPTIONS = append(DEFAULT_OPTIONS, executor.WithCgroupParent(c))
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

func compileChecker(dir storage.ProblemFiles) (executor.Volume, executor.TaskResult, error) {
	return compile(dir, dir.CheckerPath(), langs.LANG_CHECKER)
}

func compileVerifier(dir storage.ProblemFiles) (executor.Volume, executor.TaskResult, error) {
	return compile(dir, dir.VerifierPath(), langs.LANG_VERIFIER)
}

func compileModelSolution(dir storage.ProblemFiles) (executor.Volume, executor.TaskResult, error) {
	return compile(dir, dir.SolutionPath(), langs.LANG_MODEL_SOLUTION)
}

func compile(dir storage.ProblemFiles, srcPath string, l langs.Lang) (v executor.Volume, t executor.TaskResult, err error) {
	slog.Info("Compile", "lang", l.ID, "src", srcPath)

	publicRoot := dir.PublicFiles
	if info, statErr := os.Stat(publicRoot); statErr != nil || !info.IsDir() {
		if statErr != nil {
			return executor.Volume{}, executor.TaskResult{}, statErr
		}
		return executor.Volume{}, executor.TaskResult{}, fmt.Errorf("public files directory is not a directory: %s", publicRoot)
	}

	extraFilePaths := map[string]string{
		"fastio.h":   dir.PublicFilePath("common/fastio.h"),
		"grader.cpp": dir.PublicFilePath("grader/grader.cpp"),
		"solve.hpp":  dir.PublicFilePath("grader/solve.hpp"),
	}

	langForCompile := executor.Lang{
		ID:              l.ID,
		Name:            l.Name,
		Version:         l.Version,
		Source:          l.Source,
		Compile:         l.Compile,
		Exec:            l.Exec,
		ImageName:       l.ImageName,
		AdditionalFiles: l.AdditionalFiles,
	}

	options := append([]executor.TaskInfoOption{}, DEFAULT_OPTIONS...)
	options = append(options, executor.WithBindMount(publicRoot, "/problem", true))

	return executor.CompileSource(srcPath, langForCompile, options, COMPILE_TIMEOUT, extraFilePaths)
}

func runTestCase(sourceVolume, checkerVolume executor.Volume, lang langs.Lang, timeLimit float64, inFilePath, expectFilePath string) (CaseResult, error) {
	slog.Info("TestCase", "lang", lang.ID, "in", inFilePath, "expect", expectFilePath)
	outFilePath, result, err := runSource(sourceVolume, lang, timeLimit, inFilePath)
	if err != nil {
		return CaseResult{}, err
	}
	defer func() { _ = os.Remove(outFilePath) }()

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

func runSource(volume executor.Volume, lang langs.Lang, timeLimit float64, inFilePath string) (string, executor.TaskResult, error) {
	caseVolume, err := executor.CreateVolume()
	if err != nil {
		return "", executor.TaskResult{}, err
	}
	defer func() {
		if err := caseVolume.Remove(); err != nil {
			log.Println("Failed to remove caseVolume:", err)
		}
	}()

	if err := caseVolume.CopyFile(inFilePath, "input.in"); err != nil {
		return "", executor.TaskResult{}, err
	}

	// TODO: make volume read only
	taskInfo, err := executor.NewTaskInfo(lang.ImageName, append(
		DEFAULT_OPTIONS,
		executor.WithArguments(append([]string{"library-checker-init", "/casedir/input.in", "/casedir/actual.out"}, lang.Exec...)...),
		executor.WithWorkDir("/workdir"),
		executor.WithVolume(&volume, "/workdir"),
		executor.WithVolume(&caseVolume, "/casedir"),
		executor.WithTimeout(time.Duration(timeLimit*1000*1000*1000)*time.Nanosecond),
	)...)
	if err != nil {
		return "", executor.TaskResult{}, err
	}

	result, err := taskInfo.Run()
	if err != nil {
		return "", executor.TaskResult{}, err
	}

	outFile, err := os.CreateTemp("", "")
	if err != nil {
		return "", executor.TaskResult{}, err
	}
	defer func() { _ = outFile.Close() }()

	// TODO: find faster way to copy actual.out
	genOutputFileTaskInfo, err := executor.NewTaskInfo("ubuntu", append(
		DEFAULT_OPTIONS,
		executor.WithArguments("cat", "/casedir/actual.out"),
		executor.WithTimeout(COMPILE_TIMEOUT),
		executor.WithVolume(&caseVolume, "/casedir"),
		executor.WithStdout(outFile),
	)...)
	if err != nil {
		return "", executor.TaskResult{}, err
	}

	if _, err := genOutputFileTaskInfo.Run(); err != nil {
		return "", executor.TaskResult{}, err
	}

	return outFile.Name(), result, err
}

func runChecker(volume executor.Volume, inFilePath, expectFilePath, actualFilePath string) (executor.TaskResult, error) {
	checkerTaskInfo, err := executor.NewTaskInfo(langs.LANG_CHECKER.ImageName, append(
		DEFAULT_OPTIONS,
		executor.WithArguments(langs.LANG_CHECKER.Exec...),
		executor.WithWorkDir("/workdir"),
		executor.WithTimeout(CHECKER_TIMEOUT),
		executor.WithVolume(&volume, "/workdir"),
		executor.WithBindMount(inFilePath, "/workdir/input.in", true),
		executor.WithBindMount(expectFilePath, "/workdir/expect.out", true),
		executor.WithBindMount(actualFilePath, "/workdir/actual.out", true),
	)...)
	if err != nil {
		return executor.TaskResult{}, err
	}

	return checkerTaskInfo.Run()
}

func runGenerator(v executor.Volume) (string, executor.TaskResult, error) {
	outFile, err := os.CreateTemp("", "")
	if err != nil {
		return "", executor.TaskResult{}, err
	}
	defer func() { _ = outFile.Close() }()

	ti, err := executor.NewTaskInfo(langs.LANG_GENERATOR.ImageName, append(
		DEFAULT_OPTIONS,
		executor.WithArguments(langs.LANG_GENERATOR.Exec...),
		executor.WithWorkDir("/workdir"),
		executor.WithTimeout(VERIFIER_TIMEOUT),
		executor.WithVolume(&v, "/workdir"),
		executor.WithStdout(outFile),
	)...)
	if err != nil {
		return "", executor.TaskResult{}, err
	}

	result, err := ti.Run()
	if err != nil {
		return "", executor.TaskResult{}, err
	}
	return outFile.Name(), result, nil
}
