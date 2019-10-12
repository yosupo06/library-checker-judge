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
	Testzip   []byte
}

type Submission struct {
	ID          int
	ProblemName string
	Problem     Problem `gorm:"foreignkey:ProblemName"`
	Lang        string
	UserName	string
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

var casesDir string

func fetchData(db *gorm.DB, problem Problem) (string, error) {
	zipPath := path.Join(casesDir, fmt.Sprintf("cases-%s.zip", problem.Testhash))
	data := path.Join(casesDir, fmt.Sprintf("cases-%s", problem.Testhash))
	if _, err := os.Stat(zipPath); err != nil {
		// fetch zip
		zipFile, err := os.Create(zipPath)
		if err != nil {
			return "", err
		}
		if _, err = zipFile.Write(problem.Testzip); err != nil {
			return "", err
		}
		if err = zipFile.Close(); err != nil {
			return "", err
		}
		cmd := exec.Command("unzip", zipPath, "-d", data)
		if err := cmd.Run(); err != nil {
			return "", err
		}
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
	if err := db.
		Preload("Problem", func(db *gorm.DB) *gorm.DB {
			return db.Select("name, title, timelimit, testhash, testzip")
		}).
		Where("id = ?", task.Submission).First(&submission).Error; err != nil {
		return err
	}

	//set testhash
	if err := db.Model(&submission).Select("testhash").Update("testhash", submission.Problem.Testhash).Error; err != nil {
		return err
	}

	
	caseDir, err := fetchData(db, submission.Problem)
	workDir, err := ioutil.TempDir("", "work")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)
	if err != nil {
		log.Println("Fail to fetchData")
		return err
	}



	cases, err := getCases(caseDir)
	if err != nil {
		return err
	}
	checker, err := os.Open(path.Join(caseDir, "checker.cpp"))
	if err != nil {
		return err
	}
	judge, err := NewJudge(submission.Lang, checker, strings.NewReader(submission.Source), submission.Problem.Timelimit / 1000)
	if err != nil {
		return err
	}

	result, err := judge.CompileChecker()
	if err != nil {
		return err
	}
	if result.ReturnCode != 0 {
		submission.Status = "ICE"
		return db.Model(&submission).Select("status").Update("status", "ICE").Error
	}
	result, err = judge.CompileSource()
	if err != nil {
		return err
	}
	if result.ReturnCode != 0 {
		submission.Status = "CE"
		return db.Model(&submission).Select("status").Update("status", "CE").Error
	}

	caseResults := []CaseResult{}
	for _, caseName := range cases {
		inFile, err := os.Open(path.Join(caseDir, "in", caseName+".in"))
		if err != nil {
			return err
		}
		outFile, err := os.Open(path.Join(caseDir, "out", caseName+".out"))
		if err != nil {
			return err
		}
		caseResult, err := judge.TestCase(inFile, outFile)
		if err != nil {
			return err
		}		
		caseResults = append(caseResults, caseResult)

		sqlRes := SubmissionTestcaseResult{
			Submission: submission.ID,
			Testcase: caseName,
			Status: caseResult.Status,
			Time: int(caseResult.Time * 1000),
			Memory: caseResult.Memory,
		}
		if err = db.Save(&sqlRes).Error; err != nil {
			return err
		}
	}
	log.Println(caseResults)
	caseResult := AggregateResults(caseResults)
	if err = db.Model(&submission).Select("status", "max_time", "max_memory").Updates(
		Submission{Status: caseResult.Status, MaxTime: int(caseResult.Time * 1000), MaxMemory: caseResult.Memory}).Error; err != nil {
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
	myCasesDir, err := ioutil.TempDir("", "case")
	casesDir = myCasesDir
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(casesDir)
	log.Println("Case Pool Directory =", myCasesDir)

	db := gormConnect()
	defer db.Close()
	db.AutoMigrate(Problem{})
	db.AutoMigrate(Submission{})
	db.AutoMigrate(Task{})
	db.AutoMigrate(SubmissionTestcaseResult{})
	//db.LogMode(true)

	for {
		time.Sleep(1 * time.Second)
		var task Task
		tx := db.Begin()

		err := tx.First(&task).Error
		if gorm.IsRecordNotFoundError(err) {
			log.Println("waiting... ", err)
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
		log.Println("Start Judge")
		err = execJudge(db, task)
		if err != nil {
			log.Println(err.Error())
			continue
		}
	}
}
