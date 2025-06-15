package langs

import (
	"errors"
	"log"
	"os"
	"path"
	"time"
)

// CompileSource compiles source code and returns the volume and result
// This is the basic version for simple compilation without additional files
func CompileSource(sourcePath string, lang Lang, options []TaskInfoOption, timeout time.Duration, additionalFiles []string) (Volume, TaskResult, error) {
	// Set defaults
	if options == nil {
		options = getDefaultOptions()
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Create volume
	volume, err := CreateVolume()
	if err != nil {
		return Volume{}, TaskResult{}, err
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
		return Volume{}, TaskResult{}, err
	}

	// Copy additional files
	for _, filePath := range additionalFiles {
		if _, statErr := os.Stat(filePath); statErr == nil {
			if err = volume.CopyFile(filePath, path.Base(filePath)); err != nil {
				return Volume{}, TaskResult{}, err
			}
		} else if errors.Is(statErr, os.ErrNotExist) {
			log.Println(filePath, "is not found, skipping")
		} else {
			err = statErr
			return Volume{}, TaskResult{}, err
		}
	}

	// Create compilation task
	taskInfo, err := NewTaskInfo(lang.ImageName, append(
		options,
		WithArguments(lang.Compile...),
		WithWorkDir("/workdir"),
		WithVolume(&volume, "/workdir"),
		WithTimeout(timeout),
	)...)
	if err != nil {
		return Volume{}, TaskResult{}, err
	}

	// Run compilation
	result, err := taskInfo.Run()
	if err != nil {
		return Volume{}, TaskResult{}, err
	}

	return volume, result, nil
}

// getDefaultOptions returns the default TaskInfo options
// This is used internally when no options are provided
func getDefaultOptions() []TaskInfoOption {
	options := []TaskInfoOption{
		WithPidsLimit(100),            // DEFAULT_PID_LIMIT
		WithUnlimitedStackLimit(),     // unlimited
		WithMemoryLimitMB(1024),       // DEFAULT_MEMORY_LIMIT_MB
	}
	if c := os.Getenv("CGROUP_PARENT"); c != "" {
		options = append(options, WithCgroupParent(c))
	}
	return options
}

// CompileSourceWithFiles compiles source code with additional files from directories
// This is used by the judge system with full problem context
func CompileSourceWithFiles(sourcePath string, lang Lang, options []TaskInfoOption, timeout time.Duration, additionalFilesDir, extraFilesDir string) (Volume, TaskResult, error) {
	// Set defaults
	if options == nil {
		options = getDefaultOptions()
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Create volume
	volume, err := CreateVolume()
	if err != nil {
		return Volume{}, TaskResult{}, err
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
		return Volume{}, TaskResult{}, err
	}

	// Copy additional files specified by the language from additionalFilesDir
	if additionalFilesDir != "" {
		for _, key := range lang.AdditionalFiles {
			filePath := path.Join(additionalFilesDir, key)
			if _, statErr := os.Stat(filePath); statErr == nil {
				if err = volume.CopyFile(filePath, path.Base(filePath)); err != nil {
					return Volume{}, TaskResult{}, err
				}
			} else if errors.Is(statErr, os.ErrNotExist) {
				log.Println(filePath, "is not found, skipping")
			} else {
				err = statErr
				return Volume{}, TaskResult{}, err
			}
		}
	}

	// Copy extra files (common include files, params.h, etc.)
	if extraFilesDir != "" {
		// Copy params.h
		paramsPath := path.Join(extraFilesDir, "params.h")
		if _, statErr := os.Stat(paramsPath); statErr == nil {
			if err = volume.CopyFile(paramsPath, "params.h"); err != nil {
				return Volume{}, TaskResult{}, err
			}
		} else if !errors.Is(statErr, os.ErrNotExist) {
			err = statErr
			return Volume{}, TaskResult{}, err
		}

		// Copy common directory files
		commonDir := path.Join(extraFilesDir, "common")
		if files, readErr := os.ReadDir(commonDir); readErr == nil {
			for _, file := range files {
				filePath := path.Join(commonDir, file.Name())
				if err = volume.CopyFile(filePath, file.Name()); err != nil {
					return Volume{}, TaskResult{}, err
				}
			}
		} else if !errors.Is(readErr, os.ErrNotExist) {
			err = readErr
			return Volume{}, TaskResult{}, err
		}
	}

	// Create compilation task
	taskInfo, err := NewTaskInfo(lang.ImageName, append(
		options,
		WithArguments(lang.Compile...),
		WithWorkDir("/workdir"),
		WithVolume(&volume, "/workdir"),
		WithTimeout(timeout),
	)...)
	if err != nil {
		return Volume{}, TaskResult{}, err
	}

	// Run compilation
	result, err := taskInfo.Run()
	if err != nil {
		return Volume{}, TaskResult{}, err
	}

	return volume, result, nil
}