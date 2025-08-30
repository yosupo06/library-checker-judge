package database

import (
	"gorm.io/gorm"
)

// UserStatistics represents a user's ranking statistics
type UserStatistics struct {
	UserName string
	AcCount  int
}

// FetchRanking retrieves paginated user ranking data
func FetchRanking(db *gorm.DB, skip, limit int) ([]UserStatistics, int64, error) {
	// Get total count for pagination
	var totalCount int64
	if err := db.
		Model(&Submission{}).
		Select("count(distinct user_name)").
		Where("status = 'AC' and user_name is not null").
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with sorting
	var results []UserStatistics
	if err := db.
		Model(&Submission{}).
		Select("user_name, count(distinct problem_name) as ac_count").
		Where("status = 'AC' and user_name is not null").
		Group("user_name").
		Order("ac_count desc, user_name asc").
		Limit(limit).
		Offset(skip).
		Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, totalCount, nil
}
