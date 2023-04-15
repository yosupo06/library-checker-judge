package main

import (
	"flag"
	"log"
	"os"
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/yosupo06/library-checker-judge/database"
)

var testCaseFetcher TestCaseFetcher
var cgroupParent string

func main() {
	langsTomlPath := flag.String("langs", "../langs/langs.toml", "toml path of langs.toml")
	judgedir := flag.String("judgedir", "", "temporary directory of judge")

	prod := flag.Bool("prod", false, "production mode")

	pgHost := flag.String("pghost", "localhost", "postgre host")
	pgUser := flag.String("pguser", "postgres", "postgre user")
	pgPass := flag.String("pgpass", "passwd", "postgre password")
	pgTable := flag.String("pgtable", "librarychecker", "postgre table name")

	minioHost := flag.String("miniohost", "localhost:9000", "minio host")
	minioID := flag.String("minioid", "minio", "minio ID")
	minioKey := flag.String("miniokey", "miniopass", "minio access key")
	minioBucket := flag.String("miniobucket", "testcase", "minio bucket")

	tmpCgroupParent := flag.String("cgroup-parent", "", "cgroup parent")

	flag.Parse()

	cgroupParent = *tmpCgroupParent

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

		log.Println("Start Judge:", task.Submission)
		err = execTask(db, *judgedir, judgeName, *task)
		if err != nil {
			log.Println(err.Error())
			continue
		}
	}
}

func execTask(db *gorm.DB, judgedir, judgeName string, task database.Task) error {
	subID := task.Submission
	submission, err := database.FetchSubmission(db, subID)
	if err != nil {
		return err
	}
	version := submission.Problem.Testhash

	log.Println("Submission info:", subID, submission.Problem.Title)
	submission.MaxTime = -1
	submission.MaxMemory = -1
	submission.PrevStatus = submission.Status
	submission.Testhash = version
	submission.Status = "Judging"

	if err = database.SaveSubmission(db, submission); err != nil {
		return err
	}
	if err = database.ClearTestcaseResult(db, subID); err != nil {
		return err
	}
	if err = database.TouchTask(db, task.ID, judgeName); err != nil {
		return err
	}

	if err := judgeSubmission(db, judgedir, judgeName, task, submission); err != nil {
		// error detected, try to change status into IE
		submission.Status = "IE"
		if err2 := database.SaveSubmission(db, submission); err2 != nil {
			log.Println("deep error:", err2)
		}
		if err2 := database.FinishTask(db, task.ID); err2 != nil {
			log.Println("deep error:", err2)
		}
		return err
	}
	return nil
}

func judgeSubmission(db *gorm.DB, judgedir, judgeName string, task database.Task, submission database.Submission) error {
	subID := submission.ID
	version := submission.Problem.Testhash

	submission.MaxTime = -1
	submission.MaxMemory = -1
	submission.PrevStatus = submission.Status
	submission.Testhash = version

	log.Println("Fetch data")
	submission.Status = "Fetching"
	if err := database.SaveSubmission(db, submission); err != nil {
		return err
	}
	if err := database.TouchTask(db, task.ID, judgeName); err != nil {
		return err
	}

	testCases, err := testCaseFetcher.Fetch(submission.ProblemName, version)
	if err != nil {
		return err
	}
	log.Print("Fetched :", version)

	judge, err := NewJudge(judgedir, langs[submission.Lang], float64(submission.Problem.Timelimit)/1000, cgroupParent)
	if err != nil {
		return err
	}
	defer judge.Close()

	log.Println("Compile start")
	submission.Status = "Compiling"
	if err := database.SaveSubmission(db, submission); err != nil {
		return err
	}
	if err := database.TouchTask(db, task.ID, judgeName); err != nil {
		return err
	}

	checkerFile, err := testCases.CheckerFile()
	if err != nil {
		return err
	}
	defer checkerFile.Close()

	includeFilePaths, err := testCases.IncludeFilePaths()
	if err != nil {
		return err
	}

	taskResult, err := judge.CompileChecker(checkerFile, includeFilePaths)
	if err != nil {
		return err
	}
	if taskResult.ExitCode != 0 {
		submission.Status = "ICE"
		if err = database.SaveSubmission(db, submission); err != nil {
			return err
		}
		return database.FinishTask(db, task.ID)
	}

	tmpSourceFile, err := os.CreateTemp("", "output-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpSourceFile.Name())

	if _, err := tmpSourceFile.WriteString(submission.Source); err != nil {
		return err
	}
	tmpSourceFile.Close()

	tmpSourceFile2, err := os.Open(tmpSourceFile.Name())
	if err != nil {
		return err
	}
	defer tmpSourceFile2.Close()

	result, compileError, err := judge.CompileSource(tmpSourceFile2)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		submission.Status = "CE"
		submission.CompileError = compileError
		if err = database.SaveSubmission(db, submission); err != nil {
			return err
		}
		return database.FinishTask(db, task.ID)
	}

	log.Println("Start executing")
	submission.Status = "Executing"
	if err := database.SaveSubmission(db, submission); err != nil {
		return err
	}
	if err := database.TouchTask(db, task.ID, judgeName); err != nil {
		return err
	}

	cases, err := testCases.CaseNames()
	if err != nil {
		return err
	}

	caseResults := []CaseResult{}
	for _, caseName := range cases {
		inFile, err := testCases.InFile(caseName)
		if err != nil {
			return err
		}
		outFile, err := testCases.OutFile(caseName)
		if err != nil {
			return err
		}
		caseResult, err := judge.TestCase(inFile, outFile)
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
		}); err != nil {
			return err
		}
	}
	caseResult := AggregateResults(caseResults)

	submission.Status = caseResult.Status
	submission.MaxTime = int32(caseResult.Time.Milliseconds())
	submission.MaxMemory = caseResult.Memory

	log.Println(submission)
	if err := database.SaveSubmission(db, submission); err != nil {
		return err
	}
	return database.FinishTask(db, task.ID)
}
