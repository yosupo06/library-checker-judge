package langs

import (
	"io"
	"os"
	"path"
	"testing"
	"time"
)

var (
	APLUSB_DIR                 = path.Join("testdata", "sources", "aplusb")
	CPPFUNC_GRADER_DIR        = path.Join("testdata", "sources", "aplusb", "cpp-func")
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
	defer func() { _ = outFile.Close() }()

	if _, err := io.Copy(outFile, src); err != nil {
		t.Fatal(err)
	}

	return outFile.Name()
}

func createAdditionalFilesMap(langID string) map[string]string {
	extraFiles := make(map[string]string)

	if langID == "cpp-func" {
		extraFiles["fastio.h"] = path.Join(CPPFUNC_GRADER_DIR, "fastio.h")
		extraFiles["grader.cpp"] = path.Join(CPPFUNC_GRADER_DIR, "grader.cpp")
		extraFiles["solve.hpp"] = path.Join(CPPFUNC_GRADER_DIR, "solve.hpp")
	}

	return extraFiles
}

func testLangSupport(t *testing.T, langID, srcName string) {
	t.Logf("Testing language support for %s with %s", langID, srcName)

	lang, ok := GetLang(langID)
	if !ok {
		t.Fatal("Unknown lang", langID)
	}

	srcPath := path.Join(APLUSB_DIR, srcName)
	src, err := os.Open(srcPath)
	if err != nil {
		t.Fatal("Failed to open source", err)
	}
	defer func() { _ = src.Close() }()

	srcFile := langToRealFile(src, lang.Source, t)
	defer func() { _ = os.Remove(srcFile) }()

	// Check if source file exists and is readable
	if _, err := os.Stat(srcFile); err != nil {
		t.Fatal("Source file not accessible", err)
	}

	t.Logf("Successfully created source file for %s: %s", langID, srcFile)
}

func TestLangSupport(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.langID+"_"+tt.srcName, func(t *testing.T) {
			testLangSupport(t, tt.langID, tt.srcName)
		})
	}
}

func TestAllSupportedLangs(t *testing.T) {
	for _, lang := range LANGS {
		t.Run(lang.ID, func(t *testing.T) {
			t.Logf("Language %s: name=%s, source=%s, image=%s", lang.ID, lang.Name, lang.Source, lang.ImageName)
		})
	}
}

func TestGetLang(t *testing.T) {
	// Test existing language
	lang, ok := GetLang("cpp")
	if !ok {
		t.Fatal("cpp language should exist")
	}
	if lang.ID != "cpp" {
		t.Fatalf("Expected cpp, got %s", lang.ID)
	}

	// Test non-existing language
	_, ok = GetLang("nonexistent")
	if ok {
		t.Fatal("nonexistent language should not exist")
	}
}

func TestSpecialLanguages(t *testing.T) {
	// Test that special language constants are properly initialized
	if LANG_CHECKER.ID != "checker" {
		t.Errorf("Expected checker ID, got %s", LANG_CHECKER.ID)
	}
	
	if LANG_VERIFIER.ID != "verifier" {
		t.Errorf("Expected verifier ID, got %s", LANG_VERIFIER.ID)
	}
	
	if LANG_GENERATOR.ID != "generator" {
		t.Errorf("Expected generator ID, got %s", LANG_GENERATOR.ID)
	}
	
	if LANG_MODEL_SOLUTION.ID != "cpp" {
		t.Errorf("Expected cpp for model solution, got %s", LANG_MODEL_SOLUTION.ID)
	}
}