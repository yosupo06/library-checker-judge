package database

import (
	"database/sql"
	"errors"
	"sort"
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
	Submission int32  `gorm:"primaryKey"`
	Testcase   string `gorm:"primaryKey"`
	Status     string
	Time       int32
	Memory     int64
	Stderr     []byte
	CheckerOut []byte
}

func FetchSubmission(db *gorm.DB, id int32) (Submission, error) {
	sub := Submission{}
	if err := db.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("name")
		}).
		Preload("Problem", func(db *gorm.DB) *gorm.DB {
			return db.Select("name, title, testhash, public_files_hash")
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

func ClearTestcaseResult(db *gorm.DB, subID int32) error {
	if err := db.Where("submission = ?", subID).Delete(&SubmissionTestcaseResult{}).Error; err != nil {
		return err
	}
	return nil
}

func SaveTestcaseResult(db *gorm.DB, result SubmissionTestcaseResult) error {
	if err := db.Save(&result).Error; err != nil {
		return err
	}

	return nil
}

func FetchTestcaseResults(db *gorm.DB, id int32) ([]SubmissionTestcaseResult, error) {
	var cases []SubmissionTestcaseResult
	if err := db.Where("submission = ?", id).Find(&cases).Error; err != nil {
		return nil, err
	}

	// TODO: to DB query
	sort.Slice(cases, func(i, j int) bool {
		return cases[i].Testcase < cases[j].Testcase
	})

	return cases, nil
}
