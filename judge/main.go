package main

import (
	"time"
	"fmt"
	"os"
	"log"
	

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
