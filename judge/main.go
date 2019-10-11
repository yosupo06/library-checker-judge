package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

type Problem struct {
	Name      string
	Title     string
	Timelimit float64
	Testhash  string
}

type Submission struct {
	ID          int
	ProblemName string
	Problem     Problem `gorm:"foreignkey:ProblemName"`
	Lang        string
	Status      string
	Source      string
	Testhash    string
	MaxTime     int
	MaxMemory   int
	JudgePing   time.Time
}

type Task struct {
	Submission int
}

type SubmissionTestcaseResult struct {
	Submission int
	Testcase   string
	Status     string
	Time       int
	Memory     int
}

var workDir string

func fetchData(db *gorm.DB, problem Problem) (string, error) {
	zipPath := path.Join(workDir, fmt.Sprintf("cases-%s.zip", problem.Testhash))
	data := path.Join(workDir, fmt.Sprintf("cases-%s", problem.Testhash))
	if _, err := os.Stat(zipPath); err != nil {
		// fetch zip
		return "", err
	}

	cmd := exec.Command("unzip", zipPath)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return data, nil
}

func getCases(data string) ([]string, error) {
	// write glob code
	matches, err := filepath.Glob(path.Join(data, "in", "*.in"))
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, match := range matches {
		_, name := path.Split(match)
		name = strings.TrimSuffix(name, ".in")
		result = append(result, name)
	}
	return result, nil
}

func execJudge(db *gorm.DB, task Task) error {
	var submission Submission
	if err := db.Find("id = ?", submission.ID).First(&submission).Error; err != nil {
		return err
	}
	data, err := fetchData(db, submission.Problem)
	if err != nil {
		return err
	}

	cases, err := getCases(data)
	if err != nil {
		return err
	}
	checker, err := os.Open(path.Join(data, "checker.cpp"))
	if err != nil {
		return err
	}
	judge, err := NewJudge(submission.Lang, checker, strings.NewReader(submission.Source), submission.Problem.Timelimit)
	if err != nil {
		return err
	}

	result, err := judge.CompileChecker()
	if err != nil {
		return err
	}
	if result.ReturnCode != 0 {
		submission.Status = "ICE"
		if err = db.Save(&submission).Error; err != nil {
			return err
		}
		return nil
	}
	result, err = judge.CompileSource()
	if err != nil {
		return err
	}
	if result.ReturnCode != 0 {
		submission.Status = "CE"
		if err = db.Save(&submission).Error; err != nil {
			return err
		}
		return nil
	}

	caseResults := []CaseResult{}
	for _, caseName := range cases {
		inFile, err := os.Open(path.Join(data, "in", caseName+".in"))
		if err != nil {
			return err
		}
		outFile, err := os.Open(path.Join(data, "out", caseName+".in"))
		if err != nil {
			return err
		}
		caseResult, err := judge.TestCase(inFile, outFile)
		if err != nil {
			return err
		}
		caseResults = append(caseResults, caseResult)
	}
	caseResult := calcSummary(caseResults)
	submission.Status = caseResult.Status
	submission.MaxTime = int(caseResult.Time * 1000)
	submission.MaxMemory = caseResult.Memory
	if err = db.Save(&submission).Error; err != nil {
		return err
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func gormConnect() *gorm.DB {
	host := getEnv("POSTGRE_HOST", "127.0.0.1")
	port := getEnv("POSTGRE_PORT", "5432")
	user := getEnv("POSTGRE_USER", "postgres")
	pass := getEnv("POSTGRE_PASS", "passwd")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=librarychecker password=%s sslmode=disable",
		host, port, user, pass)

	db, err := gorm.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func main() {
	workDir, err := ioutil.TempDir("", "work")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(workDir)

	db := gormConnect()
	defer db.Close()
	db.AutoMigrate(Problem{})
	db.AutoMigrate(Submission{})
	db.AutoMigrate(Task{})

	for {
		time.Sleep(1 * time.Second)
		var task Task
		tx := db.Begin()

		err := tx.First(&task).Error
		if gorm.IsRecordNotFoundError(err) {
			log.Println("waiting...")
			tx.Rollback()
			continue
		}
		if err != nil {
			log.Println(err.Error())
			tx.Rollback()
			break
		}
		tx.Delete(task)
		err = tx.Commit().Error
		if err != nil {
			log.Println(err.Error())
			break
		}
	}
}
