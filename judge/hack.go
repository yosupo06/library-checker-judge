package main

import (
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
		db:     db,
		taskID: taskID,
		files:  files,
		info:   info,
		h:      hack,
		lang:   lang,
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
	db     *gorm.DB
	taskID int32
	files  storage.ProblemFiles
	info   storage.Info
	h      database.Hack
	lang   langs.Lang
}

func (data *HackTaskData) judge() error {
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

	// write input to tempfle
	tempFile, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}
	if _, err := tempFile.Write(data.h.TestCase); err != nil {
		return err
	}
	tempFile.Close()
	inFilePath := tempFile.Name()

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
		data.h.CheckerOut = r.Stderr
		return data.updateHackStatus("Invalid")
	}

	slog.Info("Generate model output")
	expectedFilePath, err := data.runModelSolution(solutionVolume, inFilePath)
	if err != nil {
		return err
	}
	defer os.Remove(expectedFilePath)

	slog.Info("Start executing")
	result, err := testCase(sourceVolume, checkerVolume, data.lang, data.info.TimeLimit, inFilePath, expectedFilePath)
	if err != nil {
		return err
	}

	data.h.Status = result.Status
	data.h.Time = int32(result.Time.Milliseconds())
	data.h.Memory = result.Memory
	data.h.Stderr = result.Stderr
	data.h.CheckerOut = result.CheckerOut
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

	return compileSource(data.files, sourceFile.Name(), data.lang)
}

func (data *HackTaskData) compileSolution() (Volume, error) {
	slog.Info("Compile solution")
	v, r, err := compileSolution(data.files)
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

func (data *HackTaskData) runModelSolution(v Volume, inFilePath string) (string, error) {
	slog.Info("Generate model output")
	path, r, err := runSource(v, langs.LANG_SOLUTION, data.info.TimeLimit, inFilePath)
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
	if err := database.TouchTask(data.db, data.taskID); err != nil {
		return err
	}
	if err := database.UpdateHack(data.db, data.h); err != nil {
		return err
	}
	return nil
}

func (data *HackTaskData) updateHack() error {
	if err := database.TouchTask(data.db, data.taskID); err != nil {
		return err
	}
	if err := database.UpdateHack(data.db, data.h); err != nil {
		return err
	}
	return nil
}
