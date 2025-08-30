package database

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

// Problem is db table
type Problem struct {
	Name             string `gorm:"primaryKey"`
	Title            string
	SourceUrl        string
	Timelimit        int32
	TestCasesVersion string
	Version          string
	OverallVersion   string
}

func FetchProblem(db *gorm.DB, name string) (Problem, error) {
	if name == "" {
		return Problem{}, errors.New("empty problem name")
	}
	problem := Problem{
		Name: name,
	}

	if err := db.Take(&problem).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return Problem{}, ErrNotExist
	} else if err != nil {
		return Problem{}, err
	}

	return problem, nil
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

type ProblemCategory struct {
	Title    string   `json:"title"`
	Problems []string `json:"problems"`
}

func FetchProblemCategories(db *gorm.DB) ([]ProblemCategory, error) {
	data, err := FetchMetadata(db, "problem_categories")
	if err != nil {
		return nil, err
	}
	var categories []ProblemCategory
	if err := json.Unmarshal([]byte(*data), &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

func SaveProblemCategories(db *gorm.DB, categories []ProblemCategory) error {
	data, err := json.Marshal(categories)
	if err != nil {
		return err
	}
	if err := SaveMetadata(db, "problem_categories", string(data)); err != nil {
		return err
	}
	return nil
}
