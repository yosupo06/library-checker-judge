package main

import (
	"database/sql"
	"errors"
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
	Timelimit int32
	Testhash  string
}

// User is db table
type User struct {
	Name       string
	Passhash   string
	Admin      bool
	Email      string
	LibraryURL string
}

// Submission is db table
type Submission struct {
	ID           int32
	ProblemName  string
	Problem      Problem `gorm:"foreignkey:ProblemName"`
	Lang         string
	Status       string
	PrevStatus   string
	Hacked       bool
	Source       string
	Testhash     string
	MaxTime      int32
	MaxMemory    int64
	CompileError []byte
	JudgePing    time.Time
	JudgeName    string
	JudgeTasked  bool
	UserName     sql.NullString
	User         User `gorm:"foreignkey:UserName"`
}

// SubmissionTestcaseResult is db table
type SubmissionTestcaseResult struct {
	Submission int32
	Testcase   string
	Status     string
	Time       int32
	Memory     int64
}

// Task is db table
type Task struct {
	ID         int32
	Submission int32
	Priority   int32
	Available  time.Time
}

func fetchUser(db *gorm.DB, name string) (User, error) {
	user := User{}
	if name == "" {
		return User{}, errors.New("User name is empty")
	}
	if err := db.Where("name = ?", name).Take(&user).Error; err != nil {
		return User{}, errors.New("User not found")
	}
	return user, nil
}

func updateUser(db *gorm.DB, user User) error {
	name := user.Name
	if name == "" {
		return errors.New("User name is empty")
	}
	result := db.Model(&User{}).Where("name = ?", name).Updates(
		map[string]interface{}{
			"admin":       user.Admin,
			"email":       user.Email,
			"library_url": user.LibraryURL,
		})
	if err := result.Error; err != nil {
		log.Print(err)
		return errors.New("Failed to update user")
	}
	if result.RowsAffected == 0 {
		return errors.New("User not found")
	}
	return nil
}

func fetchSubmission(id int32) (Submission, error) {
	sub := Submission{}
	if err := db.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("name")
		}).
		Preload("Problem", func(db *gorm.DB) *gorm.DB {
			return db.Select("name, title, testhash")
		}).
		Where("id = ?", id).First(&sub).Error; err != nil {
		return Submission{}, errors.New("Submission fetch failed")
	}
	return sub, nil
}

func pushTask(task Task) error {
	log.Print("Insert task:", task)
	if err := db.Create(&task).Error; err != nil {
		log.Print(err)
		return errors.New("Cannot insert into queue")
	}
	return nil
}

func popTask() (Task, error) {
	task := Task{}
	task.Submission = -1

	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Set("gorm:query_option", "FOR UPDATE").Where("available <= ?", time.Now()).Order("priority desc").First(&task).Error
		if gorm.IsRecordNotFoundError(err) {
			return nil
		}
		if err != nil {
			log.Print(err)
			return errors.New("Connection to db failed")
		}
		if tx.Delete(task).RowsAffected != 1 {
			log.Print("Failed to delete task:", task.ID)
			return errors.New("Failed to delete task")
		}
		return nil
	})
	return task, err
}

func dbConnect(logMode bool) *gorm.DB {
	host := getEnv("POSTGRE_HOST", "127.0.0.1")
	port := getEnv("POSTGRE_PORT", "5432")
	user := getEnv("POSTGRE_USER", "postgres")
	pass := getEnv("POSTGRE_PASS", "passwd")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=librarychecker password=%s sslmode=disable",
		host, port, user, pass)
	log.Printf("Try connect %s", connStr)
	for i := 0; i < 3; i++ {
		db, err := gorm.Open("postgres", connStr)
		if err != nil {
			log.Printf("Cannot connect db %d/3", i)
			time.Sleep(5 * time.Second)
			continue
		}
		db.AutoMigrate(Problem{})
		db.AutoMigrate(User{})
		db.AutoMigrate(Submission{})
		db.AutoMigrate(SubmissionTestcaseResult{})
		db.AutoMigrate(Task{})
		db.BlockGlobalUpdate(true)
		db.DB().SetMaxOpenConns(10)
		db.DB().SetConnMaxLifetime(time.Hour)

		if logMode {
			db.LogMode(true)
		}

		return db
	}
	log.Fatal("Cannot connect db 3 times")
	return nil
}
