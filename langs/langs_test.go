package langs

import (
	"io"
	"os"
	"path"
	"testing"
)

var (
	APLUSB_DIR     = path.Join("..", "judge", "sources", "aplusb")
	SAMPLE_IN_PATH = path.Join(APLUSB_DIR, "sample.in")
	SAMPLE_OUT_PATH = path.Join(APLUSB_DIR, "sample.out")
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

func TestCppLangSupport(t *testing.T) {
	testLangSupport(t, "cpp", "ac.cpp")
}

func TestCppAclLangSupport(t *testing.T) {
	testLangSupport(t, "cpp", "ac_acl.cpp")
}

func TestRustLangSupport(t *testing.T) {
	testLangSupport(t, "rust", "ac.rs")
}

func TestHaskellLangSupport(t *testing.T) {
	testLangSupport(t, "haskell", "ac.hs")
}

func TestHaskellCabalLangSupport(t *testing.T) {
	testLangSupport(t, "haskell", "ac_cabal.hs")
}

func TestCSharpLangSupport(t *testing.T) {
	testLangSupport(t, "csharp", "ac.cs")
}

func TestLispLangSupport(t *testing.T) {
	testLangSupport(t, "lisp", "ac.lisp")
}

func TestPython3LangSupport(t *testing.T) {
	testLangSupport(t, "python3", "ac_numpy.py")
}

func TestPyPy3LangSupport(t *testing.T) {
	testLangSupport(t, "pypy3", "ac.py")
}

func TestDLangSupport(t *testing.T) {
	testLangSupport(t, "d", "ac.d")
}

func TestJavaLangSupport(t *testing.T) {
	testLangSupport(t, "java", "ac.java")
}

func TestGoLangSupport(t *testing.T) {
	testLangSupport(t, "go", "go/ac.go")
}

func TestCrystalLangSupport(t *testing.T) {
	testLangSupport(t, "crystal", "ac.cr")
}

func TestRubyLangSupport(t *testing.T) {
	testLangSupport(t, "ruby", "ac.rb")
}

func TestAllSupportedLangs(t *testing.T) {
	expectedLangs := []string{
		"cpp", "rust", "haskell", "csharp", "lisp", 
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