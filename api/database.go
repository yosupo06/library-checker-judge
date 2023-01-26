package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"

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
	Name        string `gorm:"primaryKey"`
	Passhash    string
	Admin       bool
	Email       string
	LibraryURL  string
	IsDeveloper bool
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

type Metadata struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", errors.New("bcrypt broken")
	}
	return string(hash), nil
}

func registerUser(db *gorm.DB, name string, password string, isAdmin bool) error {
	if name == "" {
		return errors.New("empty user name")
	}
	if password == "" {
		return errors.New("empty password")
	}

	passHash, err := generatePasswordHash(password)
	if err != nil {
		return err
	}
	user := User{
		Name:     name,
		Passhash: string(passHash),
		Admin:    isAdmin,
	}
	if err := db.Create(&user).Error; err != nil {
		return errors.New("this username is already registered")
	}
	return nil
}

func verifyUserPassword(db *gorm.DB, name string, password string) error {
	user, err := fetchUser(db, name)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Passhash), []byte(password)); err != nil {
		return errors.New("password invalid")
	}

	return nil
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
			"admin":        user.Admin,
			"email":        user.Email,
			"library_url":  user.LibraryURL,
			"is_developer": user.IsDeveloper,
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

func fetchMetadata(db *gorm.DB, key string) (string, error) {
	metadata := Metadata{}
	if key == "" {
		return "", errors.New("key is empty")
	}
	if err := db.Where("key = ?", key).Take(&metadata).Error; err != nil {
		return "", errors.New("metadata not found")
	}
	return metadata.Value, nil

}

func setMetadata(db *gorm.DB, key string, value string) error {
	if key == "" {
		return errors.New("key is empty")
	}
	if err := db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&Metadata{
		Key:   key,
		Value: value,
	}).Error; err != nil {
		log.Print(err)
		return errors.New("metadata upsert failed")
	}
	return nil
}

func fetchSubmission(db *gorm.DB, id int32) (Submission, error) {
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

func fetchUserStatistics(db *gorm.DB, userName string) (map[string]pb.SolvedStatus, error) {
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

func pushTask(db *gorm.DB, task Task) error {
	log.Print("Insert task:", task)
	if err := db.Create(&task).Error; err != nil {
		log.Print(err)
		return errors.New("cannot insert into queue")
	}
	return nil
}

func popTask(db *gorm.DB) (Task, error) {
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

func dbConnect(host, port, dbname, user, pass string, enableLogger bool) *gorm.DB {
	connStr := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		host, port, dbname, user, pass)
	log.Printf("Try connect %s", connStr)
	for i := 0; i < 3; i++ {
		config := gorm.Config{}
		if enableLogger {
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
		db.AutoMigrate(Metadata{})

		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetConnMaxLifetime(time.Hour)

		return db
	}
	log.Fatal("Cannot connect db 3 times")
	return nil
}
