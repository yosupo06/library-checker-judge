package langs

import (
	"testing"
)

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
