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
	ID               int32 `gorm:"primaryKey"`
	SubmissionTime   time.Time
	ProblemName      string
	Problem          Problem `gorm:"foreignKey:ProblemName"`
	Lang             string
	Status           string
	PrevStatus       string
	Hacked           bool
	Source           string
	TestCasesVersion string
	MaxTime          int32
	MaxMemory        int64
	CompileError     []byte
	JudgePing        time.Time
	JudgeName        string
	JudgeTasked      bool
	UserName         sql.NullString
	User             User `gorm:"foreignKey:UserName"`
}

// SubmissionTestcaseResult is db table
type SubmissionTestcaseResult struct {
	Submission int32  `gorm:"primaryKey"` // TODO: should be foreign key
	Testcase   string `gorm:"primaryKey"`
	Status     string
	Time       int32
	Memory     int64
	Stderr     []byte
	CheckerOut []byte
}

func FetchSubmission(db *gorm.DB, id int32) (*Submission, error) {
	sub := Submission{
		ID: id,
	}
	if err := db.
		Preload("User").
		Preload("Problem").
		Take(&sub).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &sub, nil
}

// save submission and return id
func SaveSubmission(db *gorm.DB, submission Submission) (int32, error) {
	if submission.ID != 0 {
		return 0, errors.New("must not specify submission id")
	}
	if err := db.Save(&submission).Error; err != nil {
		return 0, err
	}

	return submission.ID, nil
}

func UpdateSubmission(db *gorm.DB, submission Submission) error {
	if submission.ID == 0 {
		return errors.New("must specify submission id")
	}
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
