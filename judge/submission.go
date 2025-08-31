package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path"
	"time" // 追加

	"gorm.io/gorm"

	"github.com/yosupo06/library-checker-judge/database"
	"github.com/yosupo06/library-checker-judge/executor"
	"github.com/yosupo06/library-checker-judge/langs"
	"github.com/yosupo06/library-checker-judge/storage"
)

func execSubmissionTask(db *gorm.DB, downloader storage.TestCaseDownloader, taskID int32, submissionData database.SubmissionData) error {
	slog.Info("Start to judge submission", "taskID", taskID, "submissionID", submissionData.ID)

	s, err := database.FetchSubmission(db, submissionData.ID)
	if err != nil {
		return err
	}

	lang, ok := langs.GetLang(s.Lang)
	if !ok {
		return fmt.Errorf("unknown language: %v", s.Lang)
	}

	problem := storage.Problem{
		Name:            s.Problem.Name,
		Version:         s.Problem.Version,
		OverallVersion:  s.Problem.OverallVersion,
		TestCaseVersion: s.Problem.TestCasesVersion,
	}
	files, err := downloader.Fetch(problem)
	if err != nil {
		return err
	}
	data := SubmissionTaskData{
		task:           NewTaskData(db, taskID),
		files:          files,
		s:              s,
		submissionData: submissionData,
		lang:           lang,
	}

	if err := data.init(); err != nil {
		return err
	}
	if err := data.judge(); err != nil {
		data.s.Status = "IE"
		if err := data.updateSubmission(); err != nil {
			slog.Error("Deep error", "taskID", taskID, "err", err)
		}
		return err
	}

	return nil
}

type SubmissionTaskData struct {
	task           TaskData
	files          storage.ProblemFiles
	s              database.Submission
	submissionData database.SubmissionData
	lang           langs.Lang
	results        []database.SubmissionTestcaseResult
	resultsToSave  []database.SubmissionTestcaseResult
	lastUpdate     time.Time
	info           storage.Info
}

func (data *SubmissionTaskData) init() error {
	info, err := storage.ParseInfo(data.files.InfoTomlPath())
	if err != nil {
		return err
	}
	data.info = info

	testCases := data.info.TestCaseNames()
	data.results = make([]database.SubmissionTestcaseResult, len(testCases))
	for i := range data.results {
		data.results[i].Submission = data.s.ID
		data.results[i].Testcase = testCases[i]
		data.results[i].Status = "-"
		data.results[i].DisplayOrder = int32(i)
	}

	data.s.MaxTime = -1
	data.s.MaxMemory = -1
	data.s.PrevStatus = data.s.Status
	data.s.Status = "-"
	data.s.TestCasesVersion = data.s.Problem.TestCasesVersion
	data.s.CompileError = []byte{}
	data.lastUpdate = time.Now()
	if err := data.updateSubmission(); err != nil {
		return err
	}
	if err := database.ClearTestcaseResult(data.task.db, data.s.ID); err != nil {
		return err
	}
	if err := database.SaveTestcaseResults(data.task.db, data.results); err != nil {
		return err
	}

	return nil
}

func (data *SubmissionTaskData) judge() error {
	slog.Info("Fetch data")
	data.s.Status = "Fetching"
	if err := data.syncStatusAndResults(false); err != nil {
		return err
	}

	slog.Info("Compile checker")
	data.s.Status = "Compiling"
	if err := data.syncStatusAndResults(false); err != nil {
		return err
	}
	checkerVolume, taskResult, err := compileChecker(data.files)
	if err != nil {
		return err
	}
	defer func() {
		if err := checkerVolume.Remove(); err != nil {
			log.Printf("Failed to remove checker volume: %v", err)
		}
	}()
	if taskResult.ExitCode != 0 {
		data.s.Status = "ICE"
		data.s.CompileError = taskResult.Stderr
		return data.updateSubmission()
	}

	sourceVolume, taskResult, err := data.compileSource()
	if err != nil {
		return err
	}
	defer func() {
		if err := sourceVolume.Remove(); err != nil {
			log.Printf("Failed to remove source volume: %v", err)
		}
	}()
	if taskResult.ExitCode != 0 {
		data.s.Status = "CE"
		data.s.CompileError = taskResult.Stderr
		return data.updateSubmission()
	}

	slog.Info("Start executing")
	testCaseNum := len(data.results)
	caseResults := []CaseResult{}
	for idx := 0; idx < testCaseNum; idx++ {
		testCaseName := data.results[idx].Testcase
		data.s.Status = fmt.Sprintf("%d/%d", idx, testCaseNum)

		if err := data.syncStatusAndResults(false); err != nil {
			return err
		}

		inFilePath := data.files.InFilePath(testCaseName)
		expectFilePath := data.files.OutFilePath(testCaseName)

		result, err := runTestCase(sourceVolume, checkerVolume, data.lang, data.info.TimeLimit, inFilePath, expectFilePath)
		if err != nil {
			return err
		}
		data.results[idx].Status = result.Status
		data.results[idx].Time = int32(result.Time.Milliseconds())
		data.results[idx].Memory = result.Memory
		data.results[idx].Stderr = result.Stderr
		data.results[idx].CheckerOut = result.CheckerOut

		data.resultsToSave = append(data.resultsToSave, data.results[idx])
		caseResults = append(caseResults, result)

		// Check if we should stop on TLE
		if data.submissionData.TleKnockout && result.Status == "TLE" {
			slog.Info("Stopping execution due to TLE (tle_knockout=true)", "testCase", testCaseName, "caseIndex", idx)
			// Mark remaining test cases as not executed
			for remainingIdx := idx + 1; remainingIdx < testCaseNum; remainingIdx++ {
				data.results[remainingIdx].Status = "-"
				data.results[remainingIdx].Time = 0
				data.results[remainingIdx].Memory = 0
				data.results[remainingIdx].Stderr = []byte{}
				data.results[remainingIdx].CheckerOut = []byte{}
				data.resultsToSave = append(data.resultsToSave, data.results[remainingIdx])
			}
			break
		}
	}

	// Final sync to save all results
	if err := data.syncStatusAndResults(true); err != nil {
		return err
	}

	totalResult := aggregateResults(caseResults)

	data.s.Status = totalResult.Status
	data.s.MaxTime = int32(totalResult.Time.Milliseconds())
	data.s.MaxMemory = totalResult.Memory
	return data.updateSubmission()
}

func (data *SubmissionTaskData) syncStatusAndResults(force bool) error {
	now := time.Now()
	if !force && now.Sub(data.lastUpdate) < 3*time.Second {
		return nil
	}

	if err := data.task.TouchIfNeeded(); err != nil {
		return err
	}

	if err := database.UpdateSubmissionStatus(data.task.db, data.s.ID, data.s.Status); err != nil {
		return err
	}

	if len(data.resultsToSave) != 0 {
		if err := database.SaveTestcaseResults(data.task.db, data.resultsToSave); err != nil {
			return err
		}
		data.resultsToSave = []database.SubmissionTestcaseResult{}
	}

	data.lastUpdate = now
	return nil
}

func (data *SubmissionTaskData) updateSubmission() error {
	if err := data.task.TouchIfNeeded(); err != nil {
		return err
	}

	if err := database.UpdateSubmission(data.task.db, data.s); err != nil {
		return err
	}
	return nil
}

func (data *SubmissionTaskData) compileSource() (executor.Volume, executor.TaskResult, error) {
	// write source to tempfile
	sourceDir, err := os.MkdirTemp("", "source")
	if err != nil {
		return executor.Volume{}, executor.TaskResult{}, err
	}
	defer func() {
		if err := os.RemoveAll(sourceDir); err != nil {
			log.Printf("Failed to remove source directory: %v", err)
		}
	}()

	sourceFile, err := os.Create(path.Join(sourceDir, data.lang.Source))
	if err != nil {
		return executor.Volume{}, executor.TaskResult{}, err
	}
	if _, err := sourceFile.WriteString(data.s.Source); err != nil {
		return executor.Volume{}, executor.TaskResult{}, err
	}
	if err := sourceFile.Close(); err != nil {
		return executor.Volume{}, executor.TaskResult{}, err
	}

	return compile(data.files, sourceFile.Name(), data.lang)
}

func aggregateResults(results []CaseResult) CaseResult {
	ans := CaseResult{
		Status: "AC",
		Time:   0,
		Memory: -1,
	}
	for _, res := range results {
		if res.Status != "AC" {
			ans.Status = res.Status
		}
		if ans.Time < res.Time {
			ans.Time = res.Time
		}
		if ans.Memory < res.Memory {
			ans.Memory = res.Memory
		}
	}
	return ans
}
