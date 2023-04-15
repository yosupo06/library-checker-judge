package database

import (
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Task is db table
type Task struct {
	ID         int32 `gorm:"primaryKey;autoIncrement"`
	JudgeName  string
	Submission int32 `gorm:"not null;unique"`
	Priority   int32
	Available  time.Time
}

func canTouch(task Task, now time.Time, judgeName string) bool {
	return now.After(task.Available) || task.JudgeName == judgeName
}

func TouchTask(db *gorm.DB, taskID int32, judgeName string) error {
	now := time.Now()

	return db.Transaction(func(tx *gorm.DB) error {
		task := Task{
			ID: taskID,
		}
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&task).Error; err != nil {
			return err
		}
		if !canTouch(task, now, judgeName) {
			return errors.New("cannot touch to this task")
		}

		task.Available = time.Now().Add(time.Minute)
		task.JudgeName = judgeName

		if err := tx.Save(&task).Error; err != nil {
			return err
		}

		return nil
	})
}

func PopTask(db *gorm.DB, judgeName string) (*Task, error) {
	task := Task{}

	if err := db.Where("available <= ?", time.Now()).Order("priority desc").Take(&task).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if err := TouchTask(db, task.ID, judgeName); err != nil {
		log.Print("failed to touch, maybe data race")
		return nil, nil
	}

	return &task, nil
}

func PushTask(db *gorm.DB, subId int32, priority int32) error {
	if err := db.Save(&Task{
		Submission: subId,
		Priority:   priority,
		Available:  time.Now(),
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
