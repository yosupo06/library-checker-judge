package langs

import (
	"io"
	"time"
)

// ExecuteSource executes compiled code with given input and returns the result
func ExecuteSource(volume Volume, lang Lang, input io.Reader, options []TaskInfoOption, timeout time.Duration) (TaskResult, error) {
	// Create execution task
	taskInfo, err := NewTaskInfo(lang.ImageName, append(
		options,
		WithArguments(lang.Exec...),
		WithWorkDir("/workdir"),
		WithVolume(&volume, "/workdir"),
		WithTimeout(timeout),
		WithStdin(input),
	)...)
	if err != nil {
		return TaskResult{}, err
	}

	// Run execution
	result, err := taskInfo.Run()
	if err != nil {
		return TaskResult{}, err
	}

	return result, nil
}

// GetLanguages returns all available languages
func GetLanguages() []Lang {
	return LANGS
}

// GetLanguage returns a specific language by ID
func GetLanguage(id string) (Lang, bool) {
	return GetLang(id)
}

// GetCheckerLanguage returns the language configuration for checkers
func GetCheckerLanguage() Lang {
	return LANG_CHECKER
}

// GetVerifierLanguage returns the language configuration for verifiers
func GetVerifierLanguage() Lang {
	return LANG_VERIFIER
}

// GetGeneratorLanguage returns the language configuration for generators
func GetGeneratorLanguage() Lang {
	return LANG_GENERATOR
}

// GetModelSolutionLanguage returns the language configuration for model solutions
func GetModelSolutionLanguage() Lang {
	return LANG_MODEL_SOLUTION
}