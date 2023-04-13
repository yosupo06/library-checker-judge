package database

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
)

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

func FetchSubmission(db *gorm.DB, id int32) (Submission, error) {
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

func SaveSubmission(db *gorm.DB, submission Submission) error {
	if err := db.Save(&submission).Error; err != nil {
		return err
	}

	return nil
}
