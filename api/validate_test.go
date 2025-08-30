package main

import (
	"strings"
	"testing"
)

// TestSourceValidate tests the 'source' validation rule defined in api/validate.go
func TestSourceValidate(t *testing.T) {
	type Param struct {
		Source string `validate:"source"`
	}

	// Valid cases
	validSources := []string{
		"a",                                // Minimum valid length (not empty)
		strings.Repeat("a", maxSourceSize), // Maximum valid length
	}
	for _, source := range validSources {
		// Use a slice of the source for error messages if it's long
		displaySource := source
		if len(source) > 50 {
			displaySource = source[:50] + "..."
		}
		if err := validate.Struct(&Param{Source: source}); err != nil {
			t.Errorf("Source starting with '%s' should be valid but got error: %v", displaySource, err)
		}
	}

	// Invalid cases
	invalidSources := []string{
		"",                                   // Empty source
		strings.Repeat("a", maxSourceSize+1), // Exceeds max length
	}
	for _, source := range invalidSources {
		displaySource := source
		if len(source) > 50 {
			displaySource = source[:50] + "..."
		}
		if err := validate.Struct(&Param{Source: source}); err == nil {
			t.Errorf("Source starting with '%s' should be invalid but passed validation", displaySource)
		}
	}
}

// TestLangValidate tests the 'lang' validation rule defined in api/validate.go
func TestLangValidate(t *testing.T) {
	type Param struct {
		Lang string `validate:"lang"`
	}

	// Valid cases - These should correspond to keys in langs/langs.toml
	// Assuming common languages are present based on project structure
	validLangs := []string{
		"cpp",
		"go",
		"java",
		"python3",
		"pypy3", // Correct ID from langs.toml
		"rust",
		"csharp",
		"haskell",
		"crystal",
		"d",
		"ruby",
		"lisp",
	}
	for _, lang := range validLangs {
		if err := validate.Struct(&Param{Lang: lang}); err != nil {
			t.Errorf("Lang '%s' should be valid but got error: %v", lang, err)
		}
	}

	// Invalid cases
	invalidLangs := []string{
		"",             // Empty lang
		"invalid-lang", // Non-existent lang
		"CPP",          // Incorrect case (assuming case-sensitive)
		" c++ ",        // Lang with spaces
	}
	for _, lang := range invalidLangs {
		if err := validate.Struct(&Param{Lang: lang}); err == nil {
			t.Errorf("Lang '%s' should be invalid but passed validation", lang)
		}
	}
}
