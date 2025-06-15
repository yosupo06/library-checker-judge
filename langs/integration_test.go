package langs

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

//go:embed testdata/*
var testdata embed.FS

// Test cases for each language verification
type TestCase struct {
	LangID     string
	SourceFile string
	Expected   string // Expected result: "AC", "WA", "PE", "TLE", "RE"
}

// AC (Accepted) test cases for each language
var acTestCases = []TestCase{
	{"cpp", "ac.cpp", "AC"},
	{"cpp", "ac_acl.cpp", "AC"},
	{"rust", "ac.rs", "AC"},
	{"haskell", "ac.hs", "AC"},
	{"haskell", "ac_cabal.hs", "AC"},
	{"csharp", "ac.cs", "AC"},
	{"lisp", "ac.lisp", "AC"},
	{"python3", "ac_numpy.py", "AC"},
	{"pypy3", "ac.py", "AC"},
	{"d", "ac.d", "AC"},
	{"java", "ac.java", "AC"},
	{"go", "go/ac.go", "AC"},
	{"crystal", "ac.cr", "AC"},
	{"ruby", "ac.rb", "AC"},
}

// Error test cases
var errorTestCases = []TestCase{
	{"cpp", "wa.cpp", "WA"},
	{"cpp", "pe.cpp", "PE"},
	{"cpp", "tle.cpp", "TLE"},
	{"cpp", "re.cpp", "RE"},
}

func TestLanguagesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Check if Docker is available
	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping integration tests")
	}

	// Ensure Docker images are built
	if err := ensureDockerImages(); err != nil {
		t.Fatalf("Failed to ensure Docker images are built: %v", err)
	}

	// Test AC cases
	for _, tc := range acTestCases {
		t.Run(fmt.Sprintf("%s_%s", tc.LangID, strings.TrimSuffix(tc.SourceFile, filepath.Ext(tc.SourceFile))), func(t *testing.T) {
			testLanguageVerification(t, tc)
		})
	}

	// Test error cases
	for _, tc := range errorTestCases {
		t.Run(fmt.Sprintf("%s_%s_error", tc.LangID, strings.TrimSuffix(tc.SourceFile, filepath.Ext(tc.SourceFile))), func(t *testing.T) {
			testLanguageVerification(t, tc)
		})
	}
}

func TestCompilationError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	if !isDockerAvailable() {
		t.Skip("Docker is not available, skipping integration tests")
	}

	// Test compilation error
	tc := TestCase{"cpp", "ce.cpp", "CE"}
	testLanguageCompilation(t, tc, true) // expectFailure = true
}

func testLanguageVerification(t *testing.T, tc TestCase) {
	lang, ok := GetLang(tc.LangID)
	if !ok {
		t.Fatalf("Language %s not found", tc.LangID)
	}

	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "langs_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Copy test files to temp directory
	if err := setupTestFiles(tempDir); err != nil {
		t.Fatalf("Failed to setup test files: %v", err)
	}

	// Copy source file
	sourceContent, err := testdata.ReadFile(fmt.Sprintf("testdata/aplusb/sources/%s", tc.SourceFile))
	if err != nil {
		t.Fatalf("Failed to read source file %s: %v", tc.SourceFile, err)
	}

	sourcePath := filepath.Join(tempDir, lang.Source)
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	if err := os.WriteFile(sourcePath, sourceContent, 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Compile checker
	checkerResult, err := compileChecker(tempDir)
	if err != nil || checkerResult.ExitCode != 0 {
		t.Fatalf("Failed to compile checker: %v, exit code: %d", err, checkerResult.ExitCode)
	}

	// Compile source
	compileResult, err := compileSource(tempDir, lang)
	if err != nil {
		t.Fatalf("Failed to compile source: %v", err)
	}

	// For compilation error test case
	if tc.Expected == "CE" {
		if compileResult.ExitCode == 0 {
			t.Fatalf("Expected compilation to fail, but it succeeded")
		}
		return
	}

	if compileResult.ExitCode != 0 {
		t.Fatalf("Compilation failed: exit code %d, stderr: %s", compileResult.ExitCode, string(compileResult.Stderr))
	}

	// Run test case
	result, err := runTestCase(tempDir, lang)
	if err != nil {
		t.Fatalf("Failed to run test case: %v", err)
	}

	// Verify result
	if result.Status != tc.Expected {
		t.Fatalf("Expected status %s, got %s. Stderr: %s, CheckerOut: %s", 
			tc.Expected, result.Status, string(result.Stderr), string(result.CheckerOut))
	}

	t.Logf("Language %s test %s passed with status %s", tc.LangID, tc.SourceFile, result.Status)
}

func testLanguageCompilation(t *testing.T, tc TestCase, expectFailure bool) {
	lang, ok := GetLang(tc.LangID)
	if !ok {
		t.Fatalf("Language %s not found", tc.LangID)
	}

	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "langs_compile_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Copy source file
	sourceContent, err := testdata.ReadFile(fmt.Sprintf("testdata/aplusb/sources/%s", tc.SourceFile))
	if err != nil {
		t.Fatalf("Failed to read source file %s: %v", tc.SourceFile, err)
	}

	sourcePath := filepath.Join(tempDir, lang.Source)
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	if err := os.WriteFile(sourcePath, sourceContent, 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Compile source
	compileResult, err := compileSource(tempDir, lang)
	if err != nil {
		t.Fatalf("Failed to run compile: %v", err)
	}

	if expectFailure {
		if compileResult.ExitCode == 0 {
			t.Fatalf("Expected compilation to fail, but it succeeded")
		}
		t.Logf("Compilation correctly failed with exit code %d", compileResult.ExitCode)
	} else {
		if compileResult.ExitCode != 0 {
			t.Fatalf("Compilation failed: exit code %d, stderr: %s", compileResult.ExitCode, string(compileResult.Stderr))
		}
		t.Logf("Compilation succeeded")
	}
}

// Docker helper functions

type ExecResult struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
	Time     time.Duration
}

type CaseResult struct {
	Status     string
	Time       time.Duration
	Memory     int64
	Stderr     []byte
	CheckerOut []byte
}

func isDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	return cmd.Run() == nil
}

func ensureDockerImages() error {
	// Check if build script exists and run it
	buildScript := "../build.sh"
	if _, err := os.Stat(buildScript); err == nil {
		cmd := exec.Command("bash", buildScript)
		cmd.Dir = ".."
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to build Docker images: %v", err)
		}
	}
	return nil
}

func setupTestFiles(tempDir string) error {
	// Copy test files from embedded filesystem
	files := map[string]string{
		"testdata/aplusb/sources/checker.cpp": "checker.cpp",
		"testdata/aplusb/sources/params.h":    "params.h",
		"testdata/aplusb/input/sample.in":     "sample.in",
		"testdata/aplusb/output/sample.out":   "sample.out",
	}

	for src, dst := range files {
		content, err := testdata.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read %s: %v", src, err)
		}
		
		dstPath := filepath.Join(tempDir, dst)
		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %v", dstPath, err)
		}
	}
	return nil
}

func compileChecker(tempDir string) (ExecResult, error) {
	// Compile checker using GCC Docker image
	cmd := []string{
		"docker", "run", "--rm",
		"-v", tempDir + ":/workdir",
		"-w", "/workdir",
		"library-checker-images-gcc",
		"g++", "-O2", "-std=c++23", "-o", "checker", "checker.cpp",
	}
	
	return runDockerCommand(cmd, 30*time.Second)
}

func compileSource(tempDir string, lang Lang) (ExecResult, error) {
	// Compile source using language-specific Docker image
	cmd := []string{"docker", "run", "--rm", "-v", tempDir + ":/workdir", "-w", "/workdir", lang.ImageName}
	cmd = append(cmd, lang.Compile...)
	
	return runDockerCommand(cmd, 30*time.Second)
}

func runTestCase(tempDir string, lang Lang) (CaseResult, error) {
	// Run the compiled program
	execCmd := strings.Join(lang.Exec, " ")
	cmd := []string{
		"bash", "-c", 
		fmt.Sprintf("docker run --rm -v %s:/workdir -w /workdir %s %s < /workdir/sample.in > /workdir/output.txt", 
			tempDir, lang.ImageName, execCmd),
	}
	
	start := time.Now()
	result, err := runDockerCommand(cmd, 5*time.Second)
	duration := time.Since(start)
	
	if err != nil {
		return CaseResult{}, err
	}
	
	baseResult := CaseResult{
		Time:   duration,
		Stderr: result.Stderr,
	}
	
	if duration > 2*time.Second {
		baseResult.Status = "TLE"
		return baseResult, nil
	}
	
	if result.ExitCode != 0 {
		baseResult.Status = "RE"
		return baseResult, nil
	}
	
	// Run checker
	checkerCmd := []string{
		"docker", "run", "--rm",
		"-v", tempDir + ":/workdir",
		"-w", "/workdir",
		"library-checker-images-gcc",
		"./checker", "sample.in", "sample.out", "output.txt",
	}
	
	checkerResult, err := runDockerCommand(checkerCmd, 10*time.Second)
	if err != nil {
		return CaseResult{}, err
	}
	
	baseResult.CheckerOut = checkerResult.Stderr
	
	if checkerResult.ExitCode == 1 {
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

func runDockerCommand(cmd []string, timeout time.Duration) (ExecResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	execCmd := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	
	start := time.Now()
	stdout, err := execCmd.Output()
	duration := time.Since(start)
	
	var stderr []byte
	if exitError, ok := err.(*exec.ExitError); ok {
		stderr = exitError.Stderr
	}
	
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return ExecResult{}, err
		}
	}
	
	return ExecResult{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
		Time:     duration,
	}, nil
}