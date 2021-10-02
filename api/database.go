package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	pb "github.com/yosupo06/library-checker-judge/api/proto"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// Problem is db table
type Problem struct {
	Name      string `gorm:"primaryKey"`
	Title     string
	SourceUrl string
	Statement string
	Timelimit int32
	Testhash  string
}

// User is db table
type User struct {
	Name       string `gorm:"primaryKey"`
	Passhash   string
	Admin      bool
	Email      string
	LibraryURL string
}

// Submission is db table
type Submission struct {
	ID           int32 `gorm:"primaryKey"`
	ProblemName  string
	Problem      Problem `gorm:"foreignKey:ProblemName"`
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
	User         User `gorm:"foreignKey:UserName"`
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
	ID         int32 `gorm:"primaryKey"`
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
		return errors.New("failed to update user")
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

func fetchUserStatistics(userName string) (map[string]pb.SolvedStatus, error) {
	type Result struct {
		ProblemName string
		LatestAC    bool
	}
	var results = make([]Result, 0)
	if err := db.
		Model(&Submission{}).
		Joins("left join problems on submissions.problem_name = problems.name").
		Select("problem_name, bool_or(submissions.testhash=problems.testhash) as latest_ac").
		Where("status = 'AC' and user_name = ?", userName).
		Group("problem_name").
		Find(&results).Error; err != nil {
		log.Print(err)
		return nil, errors.New("failed sql query")
	}
	stats := make(map[string]pb.SolvedStatus)
	for _, result := range results {
		if result.LatestAC {
			stats[result.ProblemName] = pb.SolvedStatus_LATEST_AC
		} else {
			stats[result.ProblemName] = pb.SolvedStatus_AC
		}
	}
	return stats, nil
}

func pushTask(task Task) error {
	log.Print("Insert task:", task)
	if err := db.Create(&task).Error; err != nil {
		log.Print(err)
		return errors.New("cannot insert into queue")
	}
	return nil
}

func popTask() (Task, error) {
	task := Task{}
	task.Submission = -1

	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("available <= ?", time.Now()).Order("priority desc").First(&task).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			log.Print(err)
			return errors.New("connection to db failed")
		}
		if tx.Delete(&task).RowsAffected != 1 {
			log.Print("Failed to delete task:", task.ID)
			return errors.New("failed to delete task")
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
		config := gorm.Config{}
		if logMode {
			config.Logger = logger.Default.LogMode(logger.Info)
		}
		db, err := gorm.Open(postgres.Open(connStr), &config)
		if err != nil {
			log.Printf("Cannot connect db %d/3", i)
			time.Sleep(5 * time.Second)
			continue
		}
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatal("db.DB() failed")
		}
		db.AutoMigrate(Problem{})
		db.AutoMigrate(User{})
		db.AutoMigrate(Submission{})
		db.AutoMigrate(SubmissionTestcaseResult{})
		db.AutoMigrate(Task{})

		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetConnMaxLifetime(time.Hour)

		return db
	}
	log.Fatal("Cannot connect db 3 times")
	return nil
}
