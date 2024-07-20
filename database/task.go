package database

import (
	"bytes"
	"encoding/gob"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const TASK_RETRY_PERIOD = time.Minute

type TaskType = int

const (
	JUDGE_SUBMISSION TaskType = 1
)

type TaskData struct {
	TaskType   TaskType
	Submission int32
}

// Task is db table
type Task struct {
	ID        int32 `gorm:"primaryKey;autoIncrement"`
	Priority  int32
	Available time.Time
	Enqueue   time.Time
	TaskData  []byte
}

func encode(data TaskData) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := gob.NewEncoder(buf).Encode(&data)
	return buf.Bytes(), err
}

func decode(data []byte) (TaskData, error) {
	var taskData TaskData
	err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(&taskData)
	return taskData, err
}

func PushTask(db *gorm.DB, taskData TaskData, priority int32) error {
	now := time.Now()
	binTaskData, err := encode(taskData)
	if err != nil {
		return err
	}
	if err := db.Save(&Task{
		Priority:  priority,
		Available: now,
		Enqueue:   now,
		TaskData:  binTaskData,
	}).Error; err != nil {
		return err
	}
	return nil
}

func PopTask(db *gorm.DB) (int32, TaskData, error) {
	task := Task{}
	found := false
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("available <= ?", time.Now()).Order("priority desc, id asc").Clauses(clause.Locking{Strength: "UPDATE"}).Take(&task).Error; errors.Is(err, gorm.ErrRecordNotFound) {
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
		return -1, TaskData{}, err
	}

	if !found {
		return -1, TaskData{}, nil
	}
	taskData, err := decode(task.TaskData)
	if err != nil {
		return -1, TaskData{}, err
	}
	return task.ID, taskData, nil
}

func TouchTask(db *gorm.DB, id int32) error {
	if err := db.Transaction(func(tx *gorm.DB) error {
		task := Task{
			ID: id,
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&task).Error; err != nil {
			return err
		}

		if task.Available.Before(time.Now()) {
			return errors.New("task.Available is not order than now")
		}
		task.Available = time.Now().Add(TASK_RETRY_PERIOD)
		if err := tx.Save(&task).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
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
