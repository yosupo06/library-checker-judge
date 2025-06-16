package langs

import (
	"errors"
	"log"
	"os"
	"time"
	
	"github.com/yosupo06/library-checker-judge/executor"
)

// CompileSource compiles source code and returns the volume and result
// extraFilePaths is a map from filename to full path for additional files
func CompileSource(sourcePath string, lang Lang, options []executor.TaskInfoOption, timeout time.Duration, extraFilePaths map[string]string) (executor.Volume, executor.TaskResult, error) {
	// Validate arguments
	if sourcePath == "" {
		return executor.Volume{}, executor.TaskResult{}, errors.New("sourcePath cannot be empty")
	}
	if options == nil {
		return executor.Volume{}, executor.TaskResult{}, errors.New("options cannot be nil")
	}
	if timeout <= 0 {
		return executor.Volume{}, executor.TaskResult{}, errors.New("timeout must be positive")
	}
	if extraFilePaths == nil {
		return executor.Volume{}, executor.TaskResult{}, errors.New("extraFilePaths cannot be nil")
	}

	// Create volume
	volume, err := executor.CreateVolume()
	if err != nil {
		return executor.Volume{}, executor.TaskResult{}, err
	}

	// Cleanup on error
	defer func() {
		if err != nil {
			if removeErr := volume.Remove(); removeErr != nil {
				log.Println("Volume remove failed:", removeErr)
			}
		}
	}()

	// Copy source file
	if err = volume.CopyFile(sourcePath, lang.Source); err != nil {
		return executor.Volume{}, executor.TaskResult{}, err
	}

	// Validate that all required additional files are provided
	for _, filename := range lang.AdditionalFiles {
		if _, exists := extraFilePaths[filename]; !exists {
			err = errors.New("required additional file not provided: " + filename)
			return executor.Volume{}, executor.TaskResult{}, err
		}
	}

	// Copy additional files specified by the language
	for _, filename := range lang.AdditionalFiles {
		filePath := extraFilePaths[filename]
		if _, statErr := os.Stat(filePath); statErr == nil {
			if err = volume.CopyFile(filePath, filename); err != nil {
				return executor.Volume{}, executor.TaskResult{}, err
			}
		} else if errors.Is(statErr, os.ErrNotExist) {
			log.Println(filePath, "is not found, skipping")
		} else {
			err = statErr
			return executor.Volume{}, executor.TaskResult{}, err
		}
	}

	// Copy other extra files (params.h, common files, etc.) if provided
	for filename, filePath := range extraFilePaths {
		// Skip files already handled as additional files
		isAdditionalFile := false
		for _, additionalFile := range lang.AdditionalFiles {
			if additionalFile == filename {
				isAdditionalFile = true
				break
			}
		}
		if isAdditionalFile {
			continue
		}

		if _, statErr := os.Stat(filePath); statErr == nil {
			if err = volume.CopyFile(filePath, filename); err != nil {
				return executor.Volume{}, executor.TaskResult{}, err
			}
		} else if errors.Is(statErr, os.ErrNotExist) {
			log.Println(filePath, "is not found, skipping")
		} else {
			err = statErr
			return executor.Volume{}, executor.TaskResult{}, err
		}
	}

	// Create compilation task
	taskInfo, err := executor.NewTaskInfo(lang.ImageName, append(
		options,
		executor.WithArguments(lang.Compile...),
		executor.WithWorkDir("/workdir"),
		executor.WithVolume(&volume, "/workdir"),
		executor.WithTimeout(timeout),
	)...)
	if err != nil {
		return executor.Volume{}, executor.TaskResult{}, err
	}

	// Run compilation
	result, err := taskInfo.Run()
	if err != nil {
		return executor.Volume{}, executor.TaskResult{}, err
	}

	return volume, result, nil
}