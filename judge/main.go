package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"gorm.io/gorm"

	"github.com/yosupo06/library-checker-judge/database"
	"github.com/yosupo06/library-checker-judge/langs"
	"github.com/yosupo06/library-checker-judge/storage"
)

const POOLING_PERIOD = 3 * time.Second

func main() {
	flag.Parse()

	// connect db
	db := database.Connect(database.GetDSNFromEnv(), false)

	storageClient, err := storage.Connect(storage.GetConfigFromEnv())
	if err != nil {
		log.Fatal(err)
	}
	testCaseFetcher, err := storage.NewTestCaseFetcher(storageClient)
	if err != nil {
		log.Fatal(err)
	}
	defer testCaseFetcher.Close()

	log.Println("Start Pooling")
	for {
		taskID, taskData, err := database.PopTask(db)
		if err != nil {
			log.Print("PopJudgeTask error: ", err)
			time.Sleep(POOLING_PERIOD)
			continue
		}
		if taskID == -1 {
			time.Sleep(POOLING_PERIOD)
			continue
		}

		log.Println("Start task:", taskID)
		err = judgeSubmissionTask(db, testCaseFetcher, taskID, taskData.Submission)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		database.FinishTask(db, taskID)
	}
}

func judgeSubmissionTask(db *gorm.DB, testCaseFetcher storage.TestCaseFetcher, taskID int32, id int32) (err error) {
	log.Println("Start judge submission:", id)

	s, err := initSubmission(db, id)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if err2 := updateSubmission(db, taskID, s, "IE"); err2 != nil {
				log.Println("Deep error:", err2)
			}
		}
	}()

	lang, ok := langs.GetLang(s.Lang)
	if !ok {
		return fmt.Errorf("unknown language: %v", s.Lang)
	}

	log.Println("Fetch data")
	if err := updateSubmission(db, taskID, s, "Fetching"); err != nil {
		return err
	}

	testCases, err := testCaseFetcher.Fetch(s.Problem)
	if err != nil {
		return err
	}

	judge, err := NewJudge(lang, float64(s.Problem.Timelimit)/1000, &testCases)
	if err != nil {
		return err
	}
	defer judge.Close()

	log.Println("Compile checker")
	if err := updateSubmission(db, taskID, s, "Compiling"); err != nil {
		return err
	}

	taskResult, err := judge.CompileChecker()
	if err != nil {
		return err
	}
	if taskResult.ExitCode != 0 {
		s.CompileError = taskResult.Stderr
		return finishSubmission(db, taskID, s, "ICE")
	}

	// write source to tempfile
	tmpSourceDir, err := os.MkdirTemp("", "source")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpSourceDir)

	tmpSourceFile, err := os.Create(path.Join(tmpSourceDir, lang.Source))
	if err != nil {
		return err
	}
	if _, err := tmpSourceFile.WriteString(s.Source); err != nil {
		return err
	}
	tmpSourceFile.Close()

	log.Println("Compile source")
	result, err := judge.CompileSource(tmpSourceFile.Name())
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		s.CompileError = result.Stderr
		return finishSubmission(db, taskID, s, "CE")
	}

	log.Println("Start executing")
	cases, err := testCases.CaseNames()
	if err != nil {
		return err
	}
	caseNum := len(cases)
	caseResults := []CaseResult{}
	for idx, caseName := range cases {
		if err := updateSubmission(db, taskID, s, fmt.Sprintf("%d/%d", idx, caseNum)); err != nil {
			return err
		}

		caseResult, err := judge.TestCase(caseName)
		if err != nil {
			return err
		}
		caseResults = append(caseResults, caseResult)

		if err := database.SaveTestcaseResult(db, database.SubmissionTestcaseResult{
			Submission: id,
			Testcase:   caseName,
			Status:     caseResult.Status,
			Time:       int32(caseResult.Time.Milliseconds()),
			Memory:     caseResult.Memory,
			Stderr:     caseResult.Stderr,
			CheckerOut: caseResult.CheckerOut,
		}); err != nil {
			return err
		}
	}

	caseResult := AggregateResults(caseResults)

	s.MaxTime = int32(caseResult.Time.Milliseconds())
	s.MaxMemory = caseResult.Memory
	return finishSubmission(db, taskID, s, caseResult.Status)
}

func initSubmission(db *gorm.DB, id int32) (database.Submission, error) {
	if err := database.ClearTestcaseResult(db, id); err != nil {
		return database.Submission{}, err
	}

	s, err := database.FetchSubmission(db, id)
	if err != nil {
		return database.Submission{}, err
	}

	s.MaxTime = -1
	s.MaxMemory = -1
	s.PrevStatus = s.Status
	s.Status = "-"
	s.TestCasesVersion = s.Problem.TestCasesVersion
	s.CompileError = []byte{}

	return s, database.UpdateSubmission(db, s)
}

func updateSubmission(db *gorm.DB, taskID int32, s database.Submission, status string) error {
	if err := database.TouchTask(db, taskID); err != nil {
		return err
	}
	s.Status = status
	return database.UpdateSubmission(db, s)
}

func finishSubmission(db *gorm.DB, taskID int32, s database.Submission, status string) error {
	s.JudgedTime = time.Now()
	if err := updateSubmission(db, taskID, s, status); err != nil {
		return err
	}
	return nil
}
