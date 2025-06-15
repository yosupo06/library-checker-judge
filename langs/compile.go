package langs

import (
	"errors"
	"log"
	"os"
	"path"
	"time"
)

// CompileSource compiles source code and returns the volume and result
// This is a shared function used by both judge and test code
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