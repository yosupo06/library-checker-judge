package database

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type UserSolvedStatus struct {
	ProblemName string
	LatestAC    bool
}

func FetchUserSolvedStatuses(db *gorm.DB, userName string) ([]UserSolvedStatus, error) {
	if userName == "" {
		return nil, errors.New("user name is empty")
	}

	var rows []UserSolvedStatus
	if err := db.
		Model(&Submission{}).
		Joins("LEFT JOIN problems ON submissions.problem_name = problems.name").
		Select("submissions.problem_name AS problem_name, SUM(CASE WHEN submissions.test_cases_version = problems.test_cases_version THEN 1 ELSE 0 END) > 0 AS latest_ac").
		Where("submissions.status = ? AND submissions.user_name = ?", "AC", userName).
		Group("submissions.problem_name").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("fetch user solved statuses: %w", err)
	}
	return rows, nil
}
