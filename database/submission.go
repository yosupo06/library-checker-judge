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
	UserName         sql.NullString `gorm:"index"`
	User             User           `gorm:"foreignKey:UserName"`
	JudgedTime       time.Time
}

// SubmissionOverview is smart select table
type SubmissionOverView struct {
	ID               int32
	SubmissionTime   time.Time
	ProblemName      string
	Problem          Problem
	Lang             string
	Status           string
	TestCasesVersion string
	MaxTime          int32
	MaxMemory        int64
	UserName         sql.NullString
	User             User
}

func ToSubmissionOverView(s Submission) SubmissionOverView {
	return SubmissionOverView{
		ID:               s.ID,
		SubmissionTime:   s.SubmissionTime,
		ProblemName:      s.ProblemName,
		Problem:          s.Problem,
		Lang:             s.Lang,
		Status:           s.Status,
		TestCasesVersion: s.TestCasesVersion,
		MaxTime:          s.MaxTime,
		MaxMemory:        s.MaxMemory,
		UserName:         s.UserName,
		User:             s.User,
	}
}

type SubmissionOrder int

const (
	ID_DESC SubmissionOrder = iota
	MAX_TIME_ASC
)

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

func FetchSubmission(db *gorm.DB, id int32) (Submission, error) {
	sub := Submission{
		ID: id,
	}
	if err := db.
		Preload("User").
		Preload("Problem").
		Take(&sub).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return Submission{}, ErrNotExist
	} else if err != nil {
		return Submission{}, err
	}

	return sub, nil
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

func UpdateSubmissionStatus(db *gorm.DB, id int32, status string) error {
	if err := db.Updates(Submission{
		ID:     id,
		Status: status,
	}).Error; err != nil {
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

func applyOrder(query *gorm.DB, order []SubmissionOrder) *gorm.DB {
	for _, o := range order {
		switch o {
		case ID_DESC:
			query.Order("id desc")
		case MAX_TIME_ASC:
			query.Order("max_time asc")
		}
	}
	return query
}

func FetchSubmissionList(db *gorm.DB, problem, status, lang, user string, dedupUser bool, order []SubmissionOrder, offset, limit int) ([]SubmissionOverView, int64, error) {
	filter := &Submission{
		ProblemName: problem,
		Status:      status,
		Lang:        lang,
		UserName:    sql.NullString{String: user, Valid: (user != "")},
	}

	query := db.Model(&Submission{}).Where(filter)
	query.Session(&gorm.Session{})

	if dedupUser {
		query.Order("user_name desc")
		query = applyOrder(query, order)
		query = db.Model(&Submission{}).Where("id IN (?)", query.Select("DISTINCT ON (user_name) id"))
	}

	count := int64(0)
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, errors.New("count query failed")
	}

	query = applyOrder(query, order)

	var submissions = make([]SubmissionOverView, 0)
	if err := query.Limit(limit).Offset(offset).
		Preload("User").Preload("Problem").
		Find(&submissions).Error; err != nil {
		return nil, 0, err
	}

	return submissions, count, nil
}
