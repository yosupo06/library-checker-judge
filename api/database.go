package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

// Problem is db table
type Problem struct {
	Name      string
	Title     string
	Statement string
	Timelimit float64
	Testhash  string
}

// User is db table
type User struct {
	Name     string
	Passhash string
	Admin    bool
}

// Submission is db table
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
	UserName    sql.NullString
	User        User `gorm:"foreignkey:UserName"`
}

// SubmissionTestcaseResult is db table
type SubmissionTestcaseResult struct {
	Submission int
	Testcase   string
	Status     string
	Time       int
	Memory     int
}

// Task is db table
type Task struct {
	Submission int
}

func dbConnect() *gorm.DB {
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

	db.AutoMigrate(Problem{})
	db.AutoMigrate(User{})
	db.AutoMigrate(Submission{})
	db.AutoMigrate(SubmissionTestcaseResult{})
	db.AutoMigrate(Task{})

	return db
}
