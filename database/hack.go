package database

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Hack is db table
type Hack struct {
	ID           int32 `gorm:"primaryKey"`
	HackTime     time.Time
	Submission   Submission `gorm:"foreignKey:SubmissionID"`
	SubmissionID int32
	User         User `gorm:"foreignKey:UserName"`
	UserName     sql.NullString
	TestCase     []byte

	// Result
	Status     string
	Time       int32
	Memory     int64
	Stderr     []byte
	CheckerOut []byte
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

func SaveHack(db *gorm.DB, hack Hack) (int32, error) {
	if hack.ID != 0 {
		return 0, errors.New("must not specify hack id")
	}
	if err := db.Save(&hack).Error; err != nil {
		return 0, err
	}

	return hack.ID, nil
}

func UpdateHack(db *gorm.DB, hack Hack) error {
	if hack.ID == 0 {
		return errors.New("must specify hack id")
	}
	if err := db.Save(&hack).Error; err != nil {
		return err
	}
	return nil
}
