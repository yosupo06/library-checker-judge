package main

import (
	"log"
	"log/slog"
	"os"
	"path"
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

	// Create map of extra files - always include these 3 files for all languages
	extraFilePaths := map[string]string{
		"fastio.h":   dir.PublicFilePath("common/fastio.h"),
		"grader.cpp": dir.PublicFilePath("grader/grader.cpp"),
		"solve.hpp":  dir.PublicFilePath("grader/solve.hpp"),
	}

	// Add include files (params.h and common directory files) using existing method
	includeFiles, err := dir.GetIncludeFilePaths()
	if err != nil {
		return executor.Volume{}, executor.TaskResult{}, err
	}
	for _, filePath := range includeFiles {
		filename := path.Base(filePath)
		extraFilePaths[filename] = filePath
	}

	// Use shared CompileSource function with file map
	return executor.CompileSource(srcPath, executor.Lang{ID: l.ID, Name: l.Name, Version: l.Version, Source: l.Source, Compile: l.Compile, Exec: l.Exec, ImageName: l.ImageName, AdditionalFiles: l.AdditionalFiles}, DEFAULT_OPTIONS, COMPILE_TIMEOUT, extraFilePaths)
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

func runSource(volume langs.Volume, lang langs.Lang, timeLimit float64, inFilePath string) (string, langs.TaskResult, error) {
	caseVolume, err := langs.CreateVolume()
	if err != nil {
		return "", langs.TaskResult{}, err
	}
	defer func() {
		if err := caseVolume.Remove(); err != nil {
			log.Println("Failed to remove caseVolume:", err)
		}
	}()

	if err := caseVolume.CopyFile(inFilePath, "input.in"); err != nil {
		return "", langs.TaskResult{}, err
	}

	// TODO: make volume read only
	taskInfo, err := langs.NewTaskInfo(lang.ImageName, append(
		DEFAULT_OPTIONS,
		langs.WithArguments(append([]string{"library-checker-init", "/casedir/input.in", "/casedir/actual.out"}, lang.Exec...)...),
		langs.WithWorkDir("/workdir"),
		langs.WithVolume(&volume, "/workdir"),
		langs.WithVolume(&caseVolume, "/casedir"),
		langs.WithTimeout(time.Duration(timeLimit*1000*1000*1000)*time.Nanosecond),
	)...)
	if err != nil {
		return "", langs.TaskResult{}, err
	}

	result, err := taskInfo.Run()
	if err != nil {
		return "", langs.TaskResult{}, err
	}

	outFile, err := os.CreateTemp("", "")
	if err != nil {
		return "", langs.TaskResult{}, err
	}
	defer func() { _ = outFile.Close() }()

	// TODO: find faster way to copy actual.out
	genOutputFileTaskInfo, err := langs.NewTaskInfo("ubuntu", append(
		DEFAULT_OPTIONS,
		langs.WithArguments("cat", "/casedir/actual.out"),
		langs.WithTimeout(COMPILE_TIMEOUT),
		langs.WithVolume(&caseVolume, "/casedir"),
		langs.WithStdout(outFile),
	)...)
	if err != nil {
		return "", langs.TaskResult{}, err
	}

	if _, err := genOutputFileTaskInfo.Run(); err != nil {
		return "", langs.TaskResult{}, err
	}

	return outFile.Name(), result, err
}

func runChecker(volume langs.Volume, inFilePath, expectFilePath, actualFilePath string) (langs.TaskResult, error) {
	checkerTaskInfo, err := langs.NewTaskInfo(langs.LANG_CHECKER.ImageName, append(
		DEFAULT_OPTIONS,
		langs.WithArguments(langs.LANG_CHECKER.Exec...),
		langs.WithWorkDir("/workdir"),
		langs.WithTimeout(CHECKER_TIMEOUT),
		langs.WithVolume(&volume, "/workdir"),
		langs.WithBindMount(inFilePath, "/workdir/input.in", true),
		langs.WithBindMount(expectFilePath, "/workdir/expect.out", true),
		langs.WithBindMount(actualFilePath, "/workdir/actual.out", true),
	)...)
	if err != nil {
		return langs.TaskResult{}, err
	}

	return checkerTaskInfo.Run()
}

func runGenerator(v langs.Volume) (string, langs.TaskResult, error) {
	outFile, err := os.CreateTemp("", "")
	if err != nil {
		return "", langs.TaskResult{}, err
	}
	defer func() { _ = outFile.Close() }()

	ti, err := langs.NewTaskInfo(langs.LANG_GENERATOR.ImageName, append(
		DEFAULT_OPTIONS,
		langs.WithArguments(langs.LANG_GENERATOR.Exec...),
		langs.WithWorkDir("/workdir"),
		langs.WithTimeout(VERIFIER_TIMEOUT),
		langs.WithVolume(&v, "/workdir"),
		langs.WithStdout(outFile),
	)...)
	if err != nil {
		return "", langs.TaskResult{}, err
	}

	result, err := ti.Run()
	if err != nil {
		return "", langs.TaskResult{}, err
	}
	return outFile.Name(), result, nil
}
