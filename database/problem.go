package database

import (
	"errors"

	"gorm.io/gorm"
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

func FetchProblem(db *gorm.DB, name string) (*Problem, error) {
	if name == "" {
		return nil, errors.New("empty problem name")
	}
	problem := Problem{
		Name: name,
	}
	if err := db.First(&problem).Error; err != nil {
		return nil, err
	}

	return &problem, nil
}

func SaveProblem(db *gorm.DB, problem Problem) error {
	name := problem.Name
	if name == "" {
		return errors.New("empty problem name")
	}

	if err := db.Save(&problem).Error; err != nil {
		return err
	}
	return nil
}
