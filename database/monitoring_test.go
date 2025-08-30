package database

import (
	"database/sql"
	"testing"
)

func TestFetchMonitoringData(t *testing.T) {
	db := CreateTestDB(t)

	// Test with empty database
	t.Run("EmptyDatabase", func(t *testing.T) {
		data, err := FetchMonitoringData(db)
		if err != nil {
			t.Fatal("FetchMonitoringData failed:", err)
		}
		if data == nil {
			t.Fatal("data is nil")
		}
		if data.TotalUsers != 0 {
			t.Errorf("Expected TotalUsers = 0, got %d", data.TotalUsers)
		}
		if data.TotalSubmissions != 0 {
			t.Errorf("Expected TotalSubmissions = 0, got %d", data.TotalSubmissions)
		}
		if data.TaskQueue.PendingTasks != 0 {
			t.Errorf("Expected PendingTasks = 0, got %d", data.TaskQueue.PendingTasks)
		}
		if data.TaskQueue.RunningTasks != 0 {
			t.Errorf("Expected RunningTasks = 0, got %d", data.TaskQueue.RunningTasks)
		}
		if data.TaskQueue.TotalTasks != 0 {
			t.Errorf("Expected TotalTasks = 0, got %d", data.TaskQueue.TotalTasks)
		}
	})

	// Test with sample data
	t.Run("WithSampleData", func(t *testing.T) {
		// Clean up any existing data using GORM
		db.Where("1 = 1").Delete(&User{})
		db.Where("1 = 1").Delete(&Submission{})
		db.Where("1 = 1").Delete(&Task{})
		db.Where("1 = 1").Delete(&Problem{})

		// Create test problems first
		problems := []Problem{
			{Name: "problem1", Title: "Problem 1", Timelimit: 2000},
			{Name: "problem2", Title: "Problem 2", Timelimit: 2000},
			{Name: "problem3", Title: "Problem 3", Timelimit: 2000},
		}
		for _, problem := range problems {
			if err := SaveProblem(db, problem); err != nil {
				t.Fatal("Failed to create problem:", err)
			}
		}

		// Create test users
		users := []User{
			{Name: "user1", UID: "uid1"},
			{Name: "user2", UID: "uid2"},
			{Name: "user3", UID: "uid3"},
		}
		for _, user := range users {
			if err := db.Create(&user).Error; err != nil {
				t.Fatal("Failed to create user:", err)
			}
		}

		// Create test submissions using sql.NullString for UserName
		submissions := []Submission{
			{
				ID:          1,
				ProblemName: "problem1",
				UserName:    sql.NullString{String: "user1", Valid: true},
				Status:      "AC",
			},
			{
				ID:          2,
				ProblemName: "problem2",
				UserName:    sql.NullString{String: "user1", Valid: true},
				Status:      "WA",
			},
			{
				ID:          3,
				ProblemName: "problem1",
				UserName:    sql.NullString{String: "user2", Valid: true},
				Status:      "AC",
			},
			{
				ID:          4,
				ProblemName: "problem3",
				UserName:    sql.NullString{String: "user3", Valid: true},
				Status:      "TLE",
			},
		}
		for _, submission := range submissions {
			if err := db.Create(&submission).Error; err != nil {
				t.Fatal("Failed to create submission:", err)
			}
		}

		// Create test tasks using PushSubmissionTask function
		submissionData1 := SubmissionData{ID: 1, TleKnockout: false}
		submissionData2 := SubmissionData{ID: 2, TleKnockout: false}
		submissionData3 := SubmissionData{ID: 3, TleKnockout: false}

		if err := PushSubmissionTask(db, submissionData1, 1); err != nil {
			t.Fatal("Failed to push submission task:", err)
		}
		if err := PushSubmissionTask(db, submissionData2, 1); err != nil {
			t.Fatal("Failed to push submission task:", err)
		}
		if err := PushSubmissionTask(db, submissionData3, 1); err != nil {
			t.Fatal("Failed to push submission task:", err)
		}

		// Fetch monitoring data
		data, err := FetchMonitoringData(db)
		if err != nil {
			t.Fatal("FetchMonitoringData failed:", err)
		}
		if data == nil {
			t.Fatal("data is nil")
		}

		// Verify results
		if data.TotalUsers != 3 {
			t.Errorf("Expected TotalUsers = 3, got %d", data.TotalUsers)
		}
		if data.TotalSubmissions != 4 {
			t.Errorf("Expected TotalSubmissions = 4, got %d", data.TotalSubmissions)
		}
		if data.TaskQueue.TotalTasks != 3 {
			t.Errorf("Expected TotalTasks = 3, got %d", data.TaskQueue.TotalTasks)
		}
	})
}

func TestFetchTaskQueueInfo(t *testing.T) {
	db := CreateTestDB(t)

	// Clean up any existing tasks using GORM
	db.Where("1 = 1").Delete(&Task{})

	t.Run("EmptyTaskQueue", func(t *testing.T) {
		info, err := fetchTaskQueueInfo(db)
		if err != nil {
			t.Fatal("fetchTaskQueueInfo failed:", err)
		}
		if info == nil {
			t.Fatal("info is nil")
		}
		if info.PendingTasks != 0 {
			t.Errorf("Expected PendingTasks = 0, got %d", info.PendingTasks)
		}
		if info.RunningTasks != 0 {
			t.Errorf("Expected RunningTasks = 0, got %d", info.RunningTasks)
		}
		if info.TotalTasks != 0 {
			t.Errorf("Expected TotalTasks = 0, got %d", info.TotalTasks)
		}
	})

	t.Run("WithTasks", func(t *testing.T) {
		// Clean up using GORM
		db.Where("1 = 1").Delete(&Task{})

		// Create tasks using the PushSubmissionTask function
		submissionData1 := SubmissionData{ID: 1, TleKnockout: false}
		submissionData2 := SubmissionData{ID: 2, TleKnockout: false}
		submissionData3 := SubmissionData{ID: 3, TleKnockout: false}

		if err := PushSubmissionTask(db, submissionData1, 1); err != nil {
			t.Fatal("Failed to push submission task:", err)
		}
		if err := PushSubmissionTask(db, submissionData2, 1); err != nil {
			t.Fatal("Failed to push submission task:", err)
		}
		if err := PushSubmissionTask(db, submissionData3, 1); err != nil {
			t.Fatal("Failed to push submission task:", err)
		}

		info, err := fetchTaskQueueInfo(db)
		if err != nil {
			t.Fatal("fetchTaskQueueInfo failed:", err)
		}
		if info == nil {
			t.Fatal("info is nil")
		}
		if info.TotalTasks != 3 {
			t.Errorf("Expected TotalTasks = 3, got %d", info.TotalTasks)
		}
		// Since tasks are newly created, they should all be waiting (pending)
		if info.PendingTasks != 3 {
			t.Errorf("Expected PendingTasks = 3, got %d", info.PendingTasks)
		}
		if info.RunningTasks != 0 {
			t.Errorf("Expected RunningTasks = 0, got %d", info.RunningTasks)
		}
	})
}

func TestMonitoringDataIntegration(t *testing.T) {
	db := CreateTestDB(t)

	// Clean up all tables using GORM
	db.Where("1 = 1").Delete(&User{})
	db.Where("1 = 1").Delete(&Submission{})
	db.Where("1 = 1").Delete(&Task{})
	db.Where("1 = 1").Delete(&Problem{})

	t.Run("Integration", func(t *testing.T) {
		// Create test problems first
		problems := []Problem{
			{Name: "aplusb", Title: "A + B", Timelimit: 2000},
			{Name: "unionfind", Title: "Union Find", Timelimit: 5000},
		}
		for _, problem := range problems {
			if err := SaveProblem(db, problem); err != nil {
				t.Fatal("Failed to create problem:", err)
			}
		}

		// Create test users
		users := []User{
			{Name: "alice", UID: "alice_uid"},
			{Name: "bob", UID: "bob_uid"},
			{Name: "charlie", UID: "charlie_uid"},
		}
		for _, user := range users {
			if err := db.Create(&user).Error; err != nil {
				t.Fatal("Failed to create user:", err)
			}
		}

		// Create test submissions
		submissions := []Submission{
			{
				ID:          1,
				ProblemName: "aplusb",
				UserName:    sql.NullString{String: "alice", Valid: true},
				Status:      "AC",
			},
			{
				ID:          2,
				ProblemName: "unionfind",
				UserName:    sql.NullString{String: "bob", Valid: true},
				Status:      "WA",
			},
			{
				ID:          3,
				ProblemName: "aplusb",
				UserName:    sql.NullString{String: "charlie", Valid: true},
				Status:      "AC",
			},
		}
		for _, submission := range submissions {
			if err := db.Create(&submission).Error; err != nil {
				t.Fatal("Failed to create submission:", err)
			}
		}

		// Create test tasks
		submissionData1 := SubmissionData{ID: 1, TleKnockout: false}
		submissionData2 := SubmissionData{ID: 2, TleKnockout: false}

		if err := PushSubmissionTask(db, submissionData1, 1); err != nil {
			t.Fatal("Failed to push submission task:", err)
		}
		if err := PushSubmissionTask(db, submissionData2, 2); err != nil {
			t.Fatal("Failed to push submission task:", err)
		}

		// Fetch monitoring data
		data, err := FetchMonitoringData(db)
		if err != nil {
			t.Fatal("FetchMonitoringData failed:", err)
		}
		if data == nil {
			t.Fatal("data is nil")
		}

		// Verify counts
		if data.TotalUsers != 3 {
			t.Errorf("Expected TotalUsers = 3, got %d", data.TotalUsers)
		}
		if data.TotalSubmissions != 3 {
			t.Errorf("Expected TotalSubmissions = 3, got %d", data.TotalSubmissions)
		}
		if data.TaskQueue.TotalTasks != 2 {
			t.Errorf("Expected TotalTasks = 2, got %d", data.TaskQueue.TotalTasks)
		}

		// Verify queue metrics are consistent
		total := data.TaskQueue.PendingTasks + data.TaskQueue.RunningTasks
		if total > data.TaskQueue.TotalTasks {
			t.Errorf("Pending + Running (%d) should not exceed TotalTasks (%d)", total, data.TaskQueue.TotalTasks)
		}
	})
}
