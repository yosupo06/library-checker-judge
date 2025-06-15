package langs

import (
	"strings"
	"testing"
)

func TestGetLang(t *testing.T) {
	testCases := []struct {
		langID   string
		expected bool
	}{
		{"cpp", true},
		{"rust", true},
		{"java", true},
		{"python3", true},
		{"pypy3", true},
		{"haskell", true},
		{"csharp", true},
		{"lisp", true},
		{"d", true},
		{"go", true},
		{"crystal", true},
		{"ruby", true},
		{"nonexistent", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.langID, func(t *testing.T) {
			lang, ok := GetLang(tc.langID)
			if ok != tc.expected {
				t.Errorf("GetLang(%s) returned ok=%v, expected %v", tc.langID, ok, tc.expected)
			}
			
			if ok && lang.ID != tc.langID {
				t.Errorf("GetLang(%s) returned lang with ID=%s, expected %s", tc.langID, lang.ID, tc.langID)
			}
		})
	}
}

func TestLangProperties(t *testing.T) {
	// Test that all languages have required properties
	for _, langID := range []string{"cpp", "rust", "java", "python3", "go"} {
		t.Run(langID, func(t *testing.T) {
			lang, ok := GetLang(langID)
			if !ok {
				t.Fatalf("Language %s not found", langID)
			}

			// Check ID is set
			if lang.ID == "" {
				t.Error("Language ID is empty")
			}

			// Check source extension is set
			if lang.Source == "" {
				t.Error("Source extension is empty")
			}

			// Check image name is set and follows convention
			if lang.ImageName == "" {
				t.Error("ImageName is empty")
			}
			if !strings.HasPrefix(lang.ImageName, "library-checker-images-") {
				t.Errorf("ImageName should start with 'library-checker-images-', got %s", lang.ImageName)
			}

			// Check compile commands exist
			if len(lang.Compile) == 0 {
				t.Error("Compile commands are empty")
			}

			// Check exec commands exist
			if len(lang.Exec) == 0 {
				t.Error("Exec commands are empty")
			}

			t.Logf("Language %s: Source=%s, ImageName=%s", langID, lang.Source, lang.ImageName)
		})
	}
}

func TestSpecialLangConstants(t *testing.T) {
	// Test special language constants used by judge
	specialLangs := []Lang{LANG_CHECKER, LANG_VERIFIER, LANG_MODEL_SOLUTION}
	
	for i, lang := range specialLangs {
		t.Run(lang.ID, func(t *testing.T) {
			if lang.ID == "" {
				t.Errorf("Special language %d has empty ID", i)
			}
			
			if lang.ImageName == "" {
				t.Errorf("Special language %s has empty ImageName", lang.ID)
			}
			
			if len(lang.Compile) == 0 {
				t.Errorf("Special language %s has empty compile commands", lang.ID)
			}
		})
	}
}

func TestLanguageFileExtensions(t *testing.T) {
	expectedExtensions := map[string]string{
		"cpp":     "main.cpp",
		"rust":    "main.rs", 
		"java":    "Main.java",
		"python3": "main.py",
		"pypy3":   "main.py",
		"haskell": "main.hs",
		"csharp":  "Program.cs",
		"go":      "main.go",
		"crystal": "main.cr",
		"ruby":    "main.rb",
		"d":       "main.d",
		"lisp":    "main.lisp",
	}

	for langID, expectedSource := range expectedExtensions {
		t.Run(langID, func(t *testing.T) {
			lang, ok := GetLang(langID)
			if !ok {
				t.Fatalf("Language %s not found", langID)
			}

			if lang.Source != expectedSource {
				t.Errorf("Language %s source file expected %s, got %s", langID, expectedSource, lang.Source)
			}
		})
	}
}

func TestDockerImageNaming(t *testing.T) {
	expectedImages := map[string]string{
		"cpp":     "library-checker-images-gcc",
		"rust":    "library-checker-images-rust",
		"java":    "library-checker-images-java",
		"python3": "library-checker-images-python3",
		"pypy3":   "library-checker-images-pypy",
		"haskell": "library-checker-images-haskell",
		"csharp":  "library-checker-images-csharp",
		"go":      "library-checker-images-golang",
		"crystal": "library-checker-images-crystal",
		"ruby":    "library-checker-images-ruby",
		"d":       "library-checker-images-ldc",
		"lisp":    "library-checker-images-lisp",
	}

	for langID, expectedImage := range expectedImages {
		t.Run(langID, func(t *testing.T) {
			lang, ok := GetLang(langID)
			if !ok {
				t.Fatalf("Language %s not found", langID)
			}

			if lang.ImageName != expectedImage {
				t.Errorf("Language %s image expected %s, got %s", langID, expectedImage, lang.ImageName)
			}
		})
	}
}

func TestCompileCommands(t *testing.T) {
	// Test that compile commands are reasonable
	testCases := []struct {
		langID           string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			langID:        "cpp",
			shouldContain: []string{"g++"},
		},
		{
			langID:        "rust",
			shouldContain: []string{"rustc"},
		},
		{
			langID:        "java",
			shouldContain: []string{"javac"},
		},
		{
			langID:        "go",
			shouldContain: []string{"go", "build"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.langID, func(t *testing.T) {
			lang, ok := GetLang(tc.langID)
			if !ok {
				t.Fatalf("Language %s not found", tc.langID)
			}

			compileStr := strings.Join(lang.Compile, " ")
			
			for _, should := range tc.shouldContain {
				if !strings.Contains(compileStr, should) {
					t.Errorf("Language %s compile command should contain %s, got %s", tc.langID, should, compileStr)
				}
			}
			
			for _, shouldNot := range tc.shouldNotContain {
				if strings.Contains(compileStr, shouldNot) {
					t.Errorf("Language %s compile command should not contain %s, got %s", tc.langID, shouldNot, compileStr)
				}
			}
		})
	}
}

// Test that all languages used in integration tests are properly configured
func TestIntegrationTestLanguages(t *testing.T) {
	integrationLangs := []string{
		"cpp", "rust", "haskell", "csharp", "lisp", 
		"python3", "pypy3", "d", "java", "go", "crystal", "ruby",
	}

	for _, langID := range integrationLangs {
		t.Run(langID, func(t *testing.T) {
			lang, ok := GetLang(langID)
			if !ok {
				t.Errorf("Integration test language %s is not configured", langID)
				return
			}

			// Basic validation
			if lang.ID != langID {
				t.Errorf("Language ID mismatch: expected %s, got %s", langID, lang.ID)
			}

			if lang.ImageName == "" {
				t.Errorf("Language %s has no Docker image configured", langID)
			}

			if len(lang.Compile) == 0 {
				t.Errorf("Language %s has no compile commands", langID)
			}

			if len(lang.Exec) == 0 {
				t.Errorf("Language %s has no exec commands", langID)
			}
		})
	}
}