package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/yosupo06/library-checker-judge/database"
	"github.com/yosupo06/library-checker-judge/langs"
	"github.com/yosupo06/library-checker-judge/storage"
	"gorm.io/gorm"
)

func execHackTask(db *gorm.DB, downloader storage.TestCaseDownloader, taskID int32, hackID int32) error {
	slog.Info("Start hack judge", "hackID", hackID)

	hack, err := database.FetchHack(db, hackID)
	if err != nil {
		return err
	}
	s, err := database.FetchSubmission(db, hack.SubmissionID)
	if err != nil {
		return err
	}
	lang, ok := langs.GetLang(s.Lang)
	if !ok {
		return fmt.Errorf("unknown language: %v", lang)
	}
	p := storage.Problem{
		Name:            s.Problem.Name,
		Version:         s.Problem.Version,
		TestCaseVersion: s.Problem.TestCasesVersion,
	}
	files, err := downloader.Fetch(p)
	if err != nil {
		return err
	}

	info, err := storage.ParseInfo(files.InfoTomlPath())
	if err != nil {
		return err
	}

	data := HackTaskData{
		task:  NewTaskData(db, taskID),
		files: files,
		info:  info,
		h:     hack,
		lang:  lang,
	}
	if err := data.judge(); err != nil {
		data.h.Status = "IE"
		if err := data.updateHack(); err != nil {
			slog.Error("Deep error", "err", err)
		}
		return err
	}

	return nil
}

type HackTaskData struct {
	task  TaskData
	files storage.ProblemFiles
	info  storage.Info
	h     database.Hack
	lang  langs.Lang
}

func (data *HackTaskData) judge() error {
	if err := data.updateHackStatus("Generating"); err != nil {
		return err
	}
	inFilePath, err := data.generateTestCase()
	if err != nil {
		return err
	}
	if inFilePath == "" {
		slog.Info("Failed to generate test case")
		return nil
	}
	defer os.Remove(inFilePath)

	if err := data.updateHackStatus("Compiling"); err != nil {
		return err
	}
	slog.Info("Compile source")
	sourceVolume, taskResult, err := data.compileSource()
	if err != nil {
		return err
	}
	defer sourceVolume.Remove()
	if taskResult.ExitCode != 0 {
		return data.updateHackStatus("CE")
	}
	slog.Info("Compile checker")
	checkerVolume, taskResult, err := compileChecker(data.files)
	if err != nil {
		return err
	}
	defer checkerVolume.Remove()
	if taskResult.ExitCode != 0 {
		return data.updateHackStatus("ICE")
	}
	slog.Info("Compile solution")
	solutionVolume, err := data.compileSolution()
	if err != nil {
		return err
	}
	defer solutionVolume.Remove()
	slog.Info("Compile verifier")
	verifierVolume, err := data.compileVerifier()
	if err != nil {
		return err
	}
	defer verifierVolume.Remove()

	slog.Info("Verify input")
	if err := data.updateHackStatus("Verifying"); err != nil {
		return err
	}
	path, r, err := runSource(verifierVolume, langs.LANG_VERIFIER, VERIFIER_TIMEOUT.Seconds(), inFilePath)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	if r.ExitCode != 0 {
		data.h.JudgeOutput = r.Stderr
		return data.updateHackStatus("Invalid")
	}

	slog.Info("Generate model output")
	expectedFilePath, err := data.runModelSolution(solutionVolume, inFilePath)
	if err != nil {
		return err
	}
	defer os.Remove(expectedFilePath)

	slog.Info("Start executing")
	result, err := runTestCase(sourceVolume, checkerVolume, data.lang, data.info.TimeLimit, inFilePath, expectedFilePath)
	if err != nil {
		return err
	}
	data.h.Status = result.Status
	data.h.Time = sql.NullInt32{Valid: true, Int32: int32(result.Time.Milliseconds())}
	data.h.Memory = sql.NullInt64{Valid: true, Int64: result.Memory}
	data.h.Stderr = result.Stderr
	data.h.JudgeOutput = result.CheckerOut
	return data.updateHack()
}

func (data *HackTaskData) compileSource() (Volume, TaskResult, error) {
	// write source to tempfile
	sourceDir, err := os.MkdirTemp("", "source")
	if err != nil {
		return Volume{}, TaskResult{}, err
	}
	defer os.RemoveAll(sourceDir)

	sourceFile, err := os.Create(path.Join(sourceDir, data.lang.Source))
	if err != nil {
		return Volume{}, TaskResult{}, err
	}
	if _, err := sourceFile.WriteString(data.h.Submission.Source); err != nil {
		return Volume{}, TaskResult{}, err
	}
	sourceFile.Close()

	return compile(data.files, sourceFile.Name(), data.lang)
}

func (data *HackTaskData) compileSolution() (Volume, error) {
	slog.Info("Compile solution")
	v, r, err := compileModelSolution(data.files)
	if err != nil {
		return Volume{}, err
	}
	if r.ExitCode != 0 {
		if err := v.Remove(); err != nil {
			return Volume{}, err
		}
		return Volume{}, fmt.Errorf("compile failed of model solution")
	}
	return v, nil
}

func (data *HackTaskData) compileVerifier() (Volume, error) {
	slog.Info("Compile verifier")
	v, r, err := compileVerifier(data.files)
	if err != nil {
		return Volume{}, err
	}
	if r.ExitCode != 0 {
		if err := v.Remove(); err != nil {
			return Volume{}, err
		}
		return Volume{}, fmt.Errorf("compile failed of verifier")
	}
	return v, nil
}

func (data *HackTaskData) generateTestCase() (string, error) {
	slog.Info("Generate TestCase")
	if data.h.TestCaseCpp != nil {
		tempFile, err := os.CreateTemp("", "")
		if err != nil {
			return "", err
		}
		if _, err := tempFile.Write(data.h.TestCaseCpp); err != nil {
			return "", err
		}
		if err := tempFile.Close(); err != nil {
			return "", err
		}
		defer os.Remove(tempFile.Name())

		v, r, err := compile(data.files, tempFile.Name(), langs.LANG_GENERATOR)
		if err != nil {
			return "", err
		}
		if r.ExitCode != 0 {
			data.h.JudgeOutput = r.Stderr
			return "", data.updateHackStatus("GCE")
		}
		path, r, err := runGenerator(v)
		if err != nil {
			return "", err
		}
		if r.ExitCode != 0 {
			data.h.JudgeOutput = r.Stderr
			return "", data.updateHackStatus("GE")
		}

		return path, nil
	} else if data.h.TestCaseTxt != nil {
		tempFile, err := os.CreateTemp("", "")
		if err != nil {
			return "", err
		}
		if _, err := tempFile.Write(data.h.TestCaseTxt); err != nil {
			return "", err
		}
		if err := tempFile.Close(); err != nil {
			return "", err
		}
		inFilePath := tempFile.Name()
		return inFilePath, nil
	} else {
		return "", errors.New("data source is not found")
	}
}

func (data *HackTaskData) runModelSolution(v Volume, inFilePath string) (string, error) {
	slog.Info("Generate model output")
	path, r, err := runSource(v, langs.LANG_MODEL_SOLUTION, data.info.TimeLimit, inFilePath)
	if err != nil {
		return "", err
	}
	if r.ExitCode != 0 {
		if err := os.Remove(path); err != nil {
			return "", err
		}
		return "", fmt.Errorf("model solution run failed, exits code %d", r.ExitCode)
	}
	return path, nil
}

func (data *HackTaskData) updateHackStatus(status string) error {
	data.h.Status = status

	if err := data.task.TouchIfNeeded(); err != nil {
		return err
	}

	if err := database.UpdateHack(data.task.db, data.h); err != nil {
		return err
	}
	return nil
}

func (data *HackTaskData) updateHack() error {
	if err := data.task.TouchIfNeeded(); err != nil {
		return err
	}

	if err := database.UpdateHack(data.task.db, data.h); err != nil {
		return err
	}
	return nil
}
