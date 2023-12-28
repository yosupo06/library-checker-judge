package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/yosupo06/library-checker-judge/database"
)

var testCaseFetcher TestCaseFetcher

func main() {
	langsTomlPath := flag.String("langs", "../langs/langs.toml", "toml path of langs.toml")

	prod := flag.Bool("prod", false, "production mode")

	pgHost := flag.String("pghost", "localhost", "postgre host")
	pgUser := flag.String("pguser", "postgres", "postgre user")
	pgPass := flag.String("pgpass", "passwd", "postgre password")
	pgTable := flag.String("pgtable", "librarychecker", "postgre table name")

	minioHost := flag.String("miniohost", "localhost:9000", "minio host")
	minioID := flag.String("minioid", "minio", "minio ID")
	minioKey := flag.String("miniokey", "miniopass", "minio access key")
	minioBucket := flag.String("miniobucket", "testcase", "minio bucket")
	minioPublicBucket := flag.String("miniopublicbucket", "testcase-public", "minio public bucket")

	flag.Parse()

	judgeName, err := os.Hostname()
	if err != nil {
		log.Fatal("Cannot get hostname:", err)
	}
	judgeName = judgeName + "-" + uuid.New().String()

	log.Print("JudgeName: ", judgeName)

	// connect db
	db := database.Connect(
		*pgHost,
		"5432",
		*pgTable,
		*pgUser,
		*pgPass,
		false)

	ReadLangs(*langsTomlPath)

	testCaseFetcher, err = NewTestCaseFetcher(
		*minioHost,
		*minioID,
		*minioKey,
		*minioBucket,
		*minioPublicBucket,
		*prod,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer testCaseFetcher.Close()

	log.Println("Start Pooling")
	for {
		task, err := database.PopTask(db, judgeName)
		if err != nil {
			time.Sleep(3 * time.Second)
			log.Print("PopJudgeTask error: ", err)
			continue
		}
		if task == nil {
			time.Sleep(3 * time.Second)
			continue
		}

		log.Println("Start Task:", task.Submission)
		err = execTask(db, judgeName, *task)
		if err != nil {
			log.Println(err.Error())
			continue
		}
	}
}

func execTask(db *gorm.DB, judgeName string, task database.Task) error {
	subID := task.Submission
	submission, err := database.FetchSubmission(db, subID)
	if err != nil {
		return err
	}
	problem, err := database.FetchProblem(db, submission.ProblemName)
	if problem == nil {
		return fmt.Errorf("somehow problem is not found: %s", submission.ProblemName)
	}
	if err != nil {
		return err
	}

	log.Println("Submission info:", subID, problem.Title)
	submission.MaxTime = -1
	submission.MaxMemory = -1
	submission.PrevStatus = submission.Status
	submission.Status = "Judging"

	if err = database.UpdateSubmission(db, *submission); err != nil {
		return err
	}
	if err = database.ClearTestcaseResult(db, subID); err != nil {
		return err
	}
	if err = database.TouchTask(db, task.ID, judgeName); err != nil {
		return err
	}

	if err := judgeSubmission(db, judgeName, task, *submission, *problem); err != nil {
		// error detected, try to change status into IE
		submission.Status = "IE"
		if err2 := finishSubmission(db, judgeName, submission, task.ID); err2 != nil {
			log.Println("deep error:", err2)
		}
		return err
	}
	return nil
}

func judgeSubmission(db *gorm.DB, judgeName string, task database.Task, submission database.Submission, problem database.Problem) error {
	subID := submission.ID
	taskID := task.ID

	initSubmission(&submission, problem)

	log.Println("Fetch data")
	if err := updateSubmission(db, judgeName, &submission, "Fetching", taskID); err != nil {
		return err
	}

	testCases, err := testCaseFetcher.Fetch(submission.Problem)
	if err != nil {
		return err
	}

	judge, err := NewJudge(langs[submission.Lang], float64(problem.Timelimit)/1000, &testCases)
	if err != nil {
		return err
	}
	defer judge.Close()

	log.Println("Compile checker")
	if err := updateSubmission(db, judgeName, &submission, "Compiling", taskID); err != nil {
		return err
	}

	taskResult, err := judge.CompileChecker()
	if err != nil {
		return err
	}
	if taskResult.ExitCode != 0 {
		submission.Status = "ICE"
		return finishSubmission(db, judgeName, &submission, taskID)
	}

	// write source to tempfile
	tmpSourceDir, err := os.MkdirTemp("", "source")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpSourceDir)
	tmpSourceFile, err := os.Create(path.Join(tmpSourceDir, langs[submission.Lang].Source))
	if err != nil {
		return err
	}
	if _, err := tmpSourceFile.WriteString(submission.Source); err != nil {
		return err
	}
	tmpSourceFile.Close()

	log.Println("Compile source")
	result, err := judge.CompileSource(tmpSourceFile.Name())
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		submission.Status = "CE"
		submission.CompileError = result.Stderr
		return finishSubmission(db, judgeName, &submission, taskID)
	}

	log.Println("Start executing")
	cases, err := testCases.CaseNames()
	if err != nil {
		return err
	}
	caseNum := len(cases)
	caseResults := []CaseResult{}
	for idx, caseName := range cases {
		if err := updateSubmission(db, judgeName, &submission, fmt.Sprintf("%d/%d", idx, caseNum), taskID); err != nil {
			return err
		}

		caseResult, err := judge.TestCase(caseName)
		if err != nil {
			return err
		}
		caseResults = append(caseResults, caseResult)

		if err := database.SaveTestcaseResult(db, database.SubmissionTestcaseResult{
			Submission: subID,
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

	submission.Status = caseResult.Status
	submission.MaxTime = int32(caseResult.Time.Milliseconds())
	submission.MaxMemory = caseResult.Memory
	return finishSubmission(db, judgeName, &submission, taskID)
}

func initSubmission(s *database.Submission, p database.Problem) {
	s.MaxTime = -1
	s.MaxMemory = -1
	s.PrevStatus = s.Status
	s.Status = "-"
	s.TestCasesVersion = p.TestCasesVersion
	s.CompileError = []byte{}
}

func updateSubmission(db *gorm.DB, judgeName string, s *database.Submission, status string, taskID int32) error {
	s.Status = status
	if err := database.UpdateSubmission(db, *s); err != nil {
		return err
	}
	return database.TouchTask(db, taskID, judgeName)
}

func finishSubmission(db *gorm.DB, judgeName string, s *database.Submission, taskID int32) error {
	if err := database.UpdateSubmission(db, *s); err != nil {
		return err
	}
	return database.FinishTask(db, taskID)
}
