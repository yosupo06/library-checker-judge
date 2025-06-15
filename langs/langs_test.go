package langs

import (
	"io"
	"os"
	"path"
	"testing"
	"time"
)

var (
	APLUSB_DIR                 = path.Join("sources", "aplusb")
	CPPFUNC_GRADER_DIR        = path.Join("sources", "aplusb", "cpp-func")
	DEFAULT_PID_LIMIT          = 100
	DEFAULT_MEMORY_LIMIT_MB    = 1024
	COMPILE_TIMEOUT            = 30 * time.Second
)

func langToRealFile(src io.Reader, name string, t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatal(err)
		}
	})

	outFile, err := os.Create(path.Join(tmpDir, name))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(outFile, src); err != nil {
		t.Fatal(err)
	}
	if err := outFile.Close(); err != nil {
		t.Fatal(err)
	}
	return outFile.Name()
}

func createAdditionalFilesMap(langID string) map[string]string {
	extraFiles := make(map[string]string)
	
	if langID == "cpp-func" {
		// Add the grader files for cpp-func
		extraFiles["grader.cpp"] = path.Join(CPPFUNC_GRADER_DIR, "grader.cpp")
		extraFiles["solve.hpp"] = path.Join(CPPFUNC_GRADER_DIR, "solve.hpp")
		extraFiles["fastio.h"] = path.Join(CPPFUNC_GRADER_DIR, "fastio.h")
	}
	
	return extraFiles
}

func testLangSupport(t *testing.T, langID, srcName string) {
	t.Log("Testing language support for", langID, "with", srcName)

	lang, ok := GetLang(langID)
	if !ok {
		t.Fatal("Unknown lang", langID)
	}

	srcPath := path.Join(APLUSB_DIR, srcName)
	src, err := os.Open(srcPath)
	if err != nil {
		t.Fatal("Failed to open source", err)
	}
	defer src.Close()

	srcFile := langToRealFile(src, lang.Source, t)
	defer os.Remove(srcFile)

	// Check if source file exists and is readable
	if _, err := os.Stat(srcFile); err != nil {
		t.Fatal("Source file not accessible", err)
	}

	t.Logf("Successfully created source file for %s: %s", langID, srcFile)
}

func getTestDefaultOptions() []TaskInfoOption {
	options := []TaskInfoOption{
		WithPidsLimit(DEFAULT_PID_LIMIT),
		WithUnlimitedStackLimit(),
		WithMemoryLimitMB(DEFAULT_MEMORY_LIMIT_MB),
	}
	if c := os.Getenv("CGROUP_PARENT"); c != "" {
		options = append(options, WithCgroupParent(c))
	}
	return options
}

func compileSource(srcFile string, lang Lang, t *testing.T) (Volume, TaskResult) {
	t.Logf("Compiling %s source: %s", lang.ID, srcFile)

	extraFiles := createAdditionalFilesMap(lang.ID)
	volume, result, err := CompileSource(srcFile, lang, getTestDefaultOptions(), COMPILE_TIMEOUT, extraFiles)
	if err != nil {
		if volume.Name != "" {
			volume.Remove()
		}
		t.Fatal("Failed to compile source:", err)
	}

	return volume, result
}

func runSource(volume Volume, lang Lang, timeLimit float64, inputContent string, t *testing.T) (string, TaskResult) {
	t.Logf("Running %s source with input: %s", lang.ID, inputContent)

	caseVolume, err := CreateVolume()
	if err != nil {
		t.Fatal("Failed to create case volume:", err)
	}
	defer func() {
		if err := caseVolume.Remove(); err != nil {
			t.Logf("Warning: Failed to remove case volume: %v", err)
		}
	}()

	// Create input file
	inputFile, err := os.CreateTemp("", "input*.in")
	if err != nil {
		t.Fatal("Failed to create input file:", err)
	}
	defer os.Remove(inputFile.Name())

	if _, err := inputFile.WriteString(inputContent); err != nil {
		t.Fatal("Failed to write input content:", err)
	}
	inputFile.Close()

	if err := caseVolume.CopyFile(inputFile.Name(), "input.in"); err != nil {
		t.Fatal("Failed to copy input file to volume:", err)
	}

	taskInfo, err := NewTaskInfo(lang.ImageName, append(
		getTestDefaultOptions(),
		WithArguments(append([]string{"library-checker-init", "/casedir/input.in", "/casedir/actual.out"}, lang.Exec...)...),
		WithWorkDir("/workdir"),
		WithVolume(&volume, "/workdir"),
		WithVolume(&caseVolume, "/casedir"),
		WithTimeout(time.Duration(timeLimit*1000*1000*1000)*time.Nanosecond),
	)...)
	if err != nil {
		t.Fatal("Failed to create run task:", err)
	}

	result, err := taskInfo.Run()
	if err != nil {
		t.Fatal("Failed to run source:", err)
	}

	// Extract output
	outFile, err := os.CreateTemp("", "output*.out")
	if err != nil {
		t.Fatal("Failed to create output file:", err)
	}
	defer outFile.Close()

	genOutputFileTaskInfo, err := NewTaskInfo("ubuntu", append(
		getTestDefaultOptions(),
		WithArguments("cat", "/casedir/actual.out"),
		WithTimeout(COMPILE_TIMEOUT),
		WithVolume(&caseVolume, "/casedir"),
		WithStdout(outFile),
	)...)
	if err != nil {
		t.Fatal("Failed to create output extraction task:", err)
	}

	if _, err := genOutputFileTaskInfo.Run(); err != nil {
		t.Fatal("Failed to extract output:", err)
	}

	return outFile.Name(), result
}

func testCompileAndRun(t *testing.T, langID, srcName, expectedOutput string) {
	t.Logf("Testing compile and run for %s with %s", langID, srcName)

	lang, ok := GetLang(langID)
	if !ok {
		t.Fatal("Unknown lang", langID)
	}

	srcPath := path.Join(APLUSB_DIR, srcName)
	src, err := os.Open(srcPath)
	if err != nil {
		t.Fatal("Failed to open source", err)
	}
	defer src.Close()

	srcFile := langToRealFile(src, lang.Source, t)
	defer os.Remove(srcFile)

	volume, compileResult := compileSource(srcFile, lang, t)
	defer volume.Remove()

	if compileResult.ExitCode != 0 {
		t.Fatalf("Compilation failed with exit code %d: %s", compileResult.ExitCode, string(compileResult.Stderr))
	}
	t.Logf("Compilation successful for %s", langID)

	inputContent := "1 2\n"
	outFile, runResult := runSource(volume, lang, 2.0, inputContent, t)
	defer os.Remove(outFile)

	if runResult.TLE {
		t.Fatal("Execution timed out")
	}
	if runResult.ExitCode != 0 {
		t.Fatalf("Execution failed with exit code %d: %s", runResult.ExitCode, string(runResult.Stderr))
	}

	// Read output
	outputBytes, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal("Failed to read output file:", err)
	}
	output := string(outputBytes)

	if output != expectedOutput {
		t.Fatalf("Output mismatch: expected %q, got %q", expectedOutput, output)
	}

	t.Logf("Execution successful for %s: input=%q, output=%q", langID, inputContent, output)
}

func TestLangSupport(t *testing.T) {
	testCases := []struct {
		langID  string
		srcName string
	}{
		{"cpp", "ac.cpp"},
		{"cpp", "ac_acl.cpp"},
		{"cpp-func", "ac_func.cpp"},
		{"rust", "ac.rs"},
		{"haskell", "ac.hs"},
		{"haskell", "ac_cabal.hs"},
		{"csharp", "ac.cs"},
		{"lisp", "ac.lisp"},
		{"python3", "ac_numpy.py"},
		{"pypy3", "ac.py"},
		{"d", "ac.d"},
		{"java", "ac.java"},
		{"go", "go/ac.go"},
		{"crystal", "ac.cr"},
		{"ruby", "ac.rb"},
	}

	for _, tc := range testCases {
		t.Run(tc.langID+"_"+path.Base(tc.srcName), func(t *testing.T) {
			testLangSupport(t, tc.langID, tc.srcName)
		})
	}
}

func TestAllSupportedLangs(t *testing.T) {
	expectedLangs := []string{
		"cpp", "cpp-func", "rust", "haskell", "csharp", "lisp", 
		"python3", "pypy3", "d", "java", "go", "crystal", "ruby",
	}

	for _, langID := range expectedLangs {
		t.Run(langID, func(t *testing.T) {
			lang, ok := GetLang(langID)
			if !ok {
				t.Errorf("Language %s not found in LANGS", langID)
				return
			}
			
			if lang.ID != langID {
				t.Errorf("Language ID mismatch: expected %s, got %s", langID, lang.ID)
			}
			
			if lang.Name == "" {
				t.Errorf("Language %s has empty name", langID)
			}
			
			if lang.Source == "" {
				t.Errorf("Language %s has empty source file extension", langID)
			}
			
			if len(lang.Exec) == 0 {
				t.Errorf("Language %s has no execution command", langID)
			}
			
			t.Logf("Language %s: name=%s, source=%s, image=%s", 
				lang.ID, lang.Name, lang.Source, lang.ImageName)
		})
	}
}

// Compile and run tests
func TestCompileAndRun(t *testing.T) {
	testCases := []struct {
		langID         string
		srcName        string
		expectedOutput string
	}{
		{"cpp", "ac.cpp", "3\n"},
		{"cpp-func", "ac_func.cpp", "3\n"},
		{"rust", "ac.rs", "3\n"},
		{"python3", "ac_numpy.py", "3\n"},
		{"go", "go/ac.go", "3\n"},
		{"java", "ac.java", "3\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.langID+"_"+path.Base(tc.srcName), func(t *testing.T) {
			testCompileAndRun(t, tc.langID, tc.srcName, tc.expectedOutput)
		})
	}
}