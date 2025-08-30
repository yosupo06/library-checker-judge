package database

import (
	"gorm.io/gorm"
)

type TaskQueueInfo struct {
	PendingTasks int `json:"pending_tasks"`
	RunningTasks int `json:"running_tasks"`
	TotalTasks   int `json:"total_tasks"`
}

type MonitoringData struct {
	TotalUsers       int           `json:"total_users"`
	TotalSubmissions int           `json:"total_submissions"`
	TaskQueue        TaskQueueInfo `json:"task_queue"`
}

func FetchMonitoringData(db *gorm.DB) (*MonitoringData, error) {
	var totalUsers int64
	var totalSubmissions int64

	// Get total number of users
	if err := db.Model(&User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}

	// Get total number of submissions
	if err := db.Model(&Submission{}).Count(&totalSubmissions).Error; err != nil {
		return nil, err
	}

	// Get task queue information
	taskQueue, err := fetchTaskQueueInfo(db)
	if err != nil {
		return nil, err
	}

	return &MonitoringData{
		TotalUsers:       int(totalUsers),
		TotalSubmissions: int(totalSubmissions),
		TaskQueue:        *taskQueue,
	}, nil
}

func fetchTaskQueueInfo(db *gorm.DB) (*TaskQueueInfo, error) {
	var pendingTasks int64
	var runningTasks int64
	var totalTasks int64

	// Count pending tasks (available <= now, which means they are waiting to be processed)
	// Tasks that are available for processing are considered "pending"
	if err := db.Model(&Task{}).Where("available <= ?", "NOW()").Count(&pendingTasks).Error; err != nil {
		return nil, err
	}

	// Count running tasks (available > now, which means they are currently being processed)
	// Tasks with future availability are currently being processed (running)
	if err := db.Model(&Task{}).Where("available > ?", "NOW()").Count(&runningTasks).Error; err != nil {
		return nil, err
	}

	// Count total tasks
	if err := db.Model(&Task{}).Count(&totalTasks).Error; err != nil {
		return nil, err
	}

	return &TaskQueueInfo{
		PendingTasks: int(pendingTasks),
		RunningTasks: int(runningTasks),
		TotalTasks:   int(totalTasks),
	}, nil
}
