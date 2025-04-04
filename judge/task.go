package main

import (
	"time"

	"github.com/yosupo06/library-checker-judge/database"
	"gorm.io/gorm"
)

const TASK_TOUCH_INTERVAL = 1 * time.Minute

type TaskData struct {
	db            *gorm.DB
	taskID        int32
	lastTouchTime time.Time
}

func NewTaskData(db *gorm.DB, taskID int32) TaskData {
	return TaskData{
		db:     db,
		taskID: taskID,
	}
}

func (t *TaskData) TouchIfNeeded() error {
	now := time.Now()
	if t.lastTouchTime.IsZero() || now.Sub(t.lastTouchTime) >= TASK_TOUCH_INTERVAL {
		if err := database.TouchTask(t.db, t.taskID); err != nil {
			return err
		}
		t.lastTouchTime = now
	}
	return nil
}
