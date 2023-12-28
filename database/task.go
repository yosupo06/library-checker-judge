package database

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const TASK_RETRY_PERIOD = time.Minute

// Task is db table
type Task struct {
	ID         int32 `gorm:"primaryKey;autoIncrement"`
	Submission int32
	Priority   int32
	Available  time.Time
	Enqueue    time.Time
}

func PopTask(db *gorm.DB, judgeName string) (*Task, error) {
	task := Task{}
	found := false
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := db.Where("available <= ?", time.Now()).Order("priority desc").Clauses(clause.Locking{Strength: "UPDATE"}).Take(&task).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		} else if err != nil {
			return err
		}

		found = true

		task.Available = time.Now().Add(TASK_RETRY_PERIOD)
		if err := tx.Save(&task).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if !found {
		return nil, nil
	}
	return &task, nil
}

func PushTask(db *gorm.DB, subId int32, priority int32) error {
	now := time.Now()
	if err := db.Save(&Task{
		Submission: subId,
		Priority:   priority,
		Available:  now,
		Enqueue:    now,
	}).Error; err != nil {
		return err
	}
	return nil
}

func FinishTask(db *gorm.DB, taskId int32) error {
	if err := db.Delete(&Task{
		ID: taskId,
	}).Error; err != nil {
		return err
	}
	return nil
}
