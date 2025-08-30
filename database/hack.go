package database

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Hack is db table
type Hack struct {
	ID           int32      `gorm:"primaryKey"`
	HackTime     time.Time  `gorm:"not null"`
	Submission   Submission `gorm:"foreignKey:SubmissionID"`
	SubmissionID int32
	User         *User `gorm:"foreignKey:UserName"`
	UserName     sql.NullString

	// Test case, exactly one item is not nil
	TestCaseTxt []byte
	TestCaseCpp []byte

	// Result
	Status      string
	Time        sql.NullInt32
	Memory      sql.NullInt64
	Stderr      []byte
	JudgeOutput []byte
}

func FetchHack(db *gorm.DB, id int32) (Hack, error) {
	hack := Hack{
		ID: id,
	}
	if err := db.
		Preload("User").
		Preload("Submission").
		Take(&hack).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return Hack{}, ErrNotExist
	} else if err != nil {
		return Hack{}, err
	}

	return hack, nil
}

func SaveHack(db *gorm.DB, h Hack) (int32, error) {
	if h.ID != 0 {
		return 0, errors.New("must not specify hack id")
	}
	if err := h.valid(); err != nil {
		return 0, err
	}
	if err := db.Save(&h).Error; err != nil {
		return 0, err
	}
	return h.ID, nil
}

func UpdateHack(db *gorm.DB, h Hack) error {
	if h.ID == 0 {
		return errors.New("must specify hack id")
	}
	if err := h.valid(); err != nil {
		return err
	}
	if err := db.Save(&h).Error; err != nil {
		return err
	}
	return nil
}

func (h *Hack) valid() error {
	if h.TestCaseCpp == nil && h.TestCaseTxt == nil {
		return errors.New("must contain test case")
	}
	if h.TestCaseCpp != nil && h.TestCaseTxt != nil {
		return errors.New("must contain at most one test case")
	}
	return nil
}

func FetchHackList(db *gorm.DB, skip, limit int, user, status, order string) ([]Hack, error) {
	var hacks []Hack

	query := db.Preload("User").Preload("Submission")

	if user != "" {
		query = query.Where("user_name = ?", user)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if order == "" {
		order = "-id"
	}
	switch order {
	case "-id":
		query = query.Order("id DESC")
	case "id":
		query = query.Order("id ASC")
	case "time":
		query = query.Order("hack_time ASC")
	case "-time":
		query = query.Order("hack_time DESC")
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if skip > 0 {
		query = query.Offset(skip)
	}

	if err := query.Find(&hacks).Error; err != nil {
		return nil, err
	}

	return hacks, nil
}

func CountHacks(db *gorm.DB, user, status string) (int, error) {
	var count int64

	query := db.Model(&Hack{})

	if user != "" {
		query = query.Where("user_name = ?", user)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}
