package main

import (
	"errors"
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

const POOLING_PERIOD = 3 * time.Second

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
			log.Print("PopJudgeTask error: ", err)
			time.Sleep(POOLING_PERIOD)
			continue
		}
		if task == nil {
			time.Sleep(POOLING_PERIOD)
			continue
		}

		log.Println("Start task:", task)
		err = judgeSubmissionTask(db, judgeName, task.Submission, task.Enqueue)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		database.FinishTask(db, task.ID)
	}
}

func judgeSubmissionTask(db *gorm.DB, judgeName string, id int32, enqueue time.Time) (err error) {
	log.Println("Start judge submission:", id)

	s, err := initSubmission(db, judgeName, id, enqueue)
	if err != nil {
		return err
	}
	if s == nil {
		return nil
	}

	defer func() {
		if err != nil {
			if err2 := updateSubmission(db, judgeName, s, "IE"); err2 != nil {
				log.Println("Deep error:", err2)
			}
		}
	}()

	log.Println("Fetch data")
	if err := updateSubmission(db, judgeName, s, "Fetching"); err != nil {
		return err
	}

	testCases, err := testCaseFetcher.Fetch(s.Problem)
	if err != nil {
		return err
	}

	judge, err := NewJudge(langs[s.Lang], float64(s.Problem.Timelimit)/1000, &testCases)
	if err != nil {
		return err
	}
	defer judge.Close()

	log.Println("Compile checker")
	if err := updateSubmission(db, judgeName, s, "Compiling"); err != nil {
		return err
	}

	taskResult, err := judge.CompileChecker()
	if err != nil {
		return err
	}
	if taskResult.ExitCode != 0 {
		return finishSubmission(db, judgeName, s, "ICE")
	}

	// write source to tempfile
	tmpSourceDir, err := os.MkdirTemp("", "source")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpSourceDir)
	tmpSourceFile, err := os.Create(path.Join(tmpSourceDir, langs[s.Lang].Source))
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
		return finishSubmission(db, judgeName, s, "CE")
	}

	log.Println("Start executing")
	cases, err := testCases.CaseNames()
	if err != nil {
		return err
	}
	caseNum := len(cases)
	caseResults := []CaseResult{}
	for idx, caseName := range cases {
		if err := updateSubmission(db, judgeName, s, fmt.Sprintf("%d/%d", idx, caseNum)); err != nil {
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
	return finishSubmission(db, judgeName, s, caseResult.Status)
}

func initSubmission(db *gorm.DB, name string, id int32, enqueue time.Time) (*database.Submission, error) {
	if ok, err := database.TryLockSubmission(db, id, name); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("failed to lock submission: %d", id)
	}

	s, err := database.FetchSubmission(db, id)
	if err != nil {
		return nil, err
	}

	if s.JudgedTime.After(enqueue) {
		log.Println("Already judged:", id)
		return nil, err
	}

	s.MaxTime = -1
	s.MaxMemory = -1
	s.PrevStatus = s.Status
	s.Status = "-"
	s.TestCasesVersion = s.Problem.TestCasesVersion
	s.CompileError = []byte{}

	return s, database.UpdateSubmission(db, *s)
}

func updateSubmission(db *gorm.DB, judgeName string, s *database.Submission, status string) error {
	if err := lockSubmission(db, s.ID, judgeName); err != nil {
		return err
	}
	s.Status = status
	return database.UpdateSubmission(db, *s)
}

func finishSubmission(db *gorm.DB, judgeName string, s *database.Submission, status string) error {
	s.JudgedTime = time.Now()
	if err := updateSubmission(db, judgeName, s, status); err != nil {
		return err
	}
	return database.UnlockSubmission(db, s.ID, judgeName)
}

func lockSubmission(db *gorm.DB, id int32, name string) error {
	if ok, err := database.TryLockSubmission(db, id, name); err != nil {
		return err
	} else if !ok {
		return errors.New("lock failed")
	}
	return nil
}
