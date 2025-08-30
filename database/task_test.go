package database

import (
	"testing"
)

func TestTask(t *testing.T) {
	db := CreateTestDB(t)

	submissionData := SubmissionData{
		ID: 123,
	}
	hackData := HackData{
		ID: 456,
	}

	if err := PushSubmissionTask(db, submissionData, 1); err != nil {
		t.Fatal(err)
	}
	if err := PushHackTask(db, hackData, 10); err != nil {
		t.Fatal(err)
	}

	id1, data1, err := PopTask(db)
	if id1 == -1 || err != nil {
		t.Fatal(id1, data1, err)
	}
	if hackResult, ok := data1.Data.(HackData); !ok || hackResult.ID != 456 {
		t.Fatal("Expected HackData with ID 456, got:", data1.Data)
	}

	id2, data2, err := PopTask(db)
	if id2 == -1 || err != nil {
		t.Fatal(id2, data2, err)
	}
	if submissionResult, ok := data2.Data.(SubmissionData); !ok || submissionResult.ID != 123 {
		t.Fatal("Expected SubmissionData with ID 123, got:", data2.Data)
	}

	id3, data3, err := PopTask(db)
	if id3 != -1 || err != nil {
		t.Fatal(id3, data3, err)
	}

	if err := FinishTask(db, id1); err != nil {
		t.Fatal(err)
	}
	if err := FinishTask(db, id2); err != nil {
		t.Fatal(err)
	}
}

func TestTaskSamePriority(t *testing.T) {
	db := CreateTestDB(t)

	submission1 := SubmissionData{ID: 123}
	submission2 := SubmissionData{ID: 124}
	submission3 := SubmissionData{ID: 125}

	if err := PushSubmissionTask(db, submission1, 10); err != nil {
		t.Fatal(err)
	}
	if err := PushSubmissionTask(db, submission2, 10); err != nil {
		t.Fatal(err)
	}
	if err := PushSubmissionTask(db, submission3, 10); err != nil {
		t.Fatal(err)
	}

	id1, data1, err := PopTask(db)
	if id1 == -1 || err != nil {
		t.Fatal(id1, data1, err)
	}
	if submissionResult, ok := data1.Data.(SubmissionData); !ok || submissionResult.ID != 123 {
		t.Fatal("Expected SubmissionData with ID 123, got:", data1.Data)
	}

	id2, data2, err := PopTask(db)
	if id2 == -1 || err != nil {
		t.Fatal(id2, data2, err)
	}
	if submissionResult, ok := data2.Data.(SubmissionData); !ok || submissionResult.ID != 124 {
		t.Fatal("Expected SubmissionData with ID 124, got:", data2.Data)
	}

	id3, data3, err := PopTask(db)
	if id3 == -1 || err != nil {
		t.Fatal(id3, data3, err)
	}
	if submissionResult, ok := data3.Data.(SubmissionData); !ok || submissionResult.ID != 125 {
		t.Fatal("Expected SubmissionData with ID 125, got:", data3.Data)
	}
}

func TestTaskDataSerialize(t *testing.T) {
	submissionData := SubmissionData{
		ID: 123,
	}
	task := TaskData{
		TaskType: JudgeSubmission,
		Data:     submissionData,
	}
	buf, err := encode(task)
	if err != nil {
		t.Fatal(err)
	}
	task2, err := decode(buf)
	if err != nil {
		t.Fatal(err)
	}

	// Compare the decoded data
	if task2.TaskType != task.TaskType {
		t.Fatal("TaskType mismatch:", task2.TaskType, "!=", task.TaskType)
	}

	decodedSubmission, ok := task2.Data.(SubmissionData)
	if !ok {
		t.Fatal("Failed to decode SubmissionData")
	}

	if decodedSubmission != submissionData {
		t.Fatal("SubmissionData mismatch:", decodedSubmission, "!=", submissionData)
	}
}

func TestTaskWithRealDatabaseRecords(t *testing.T) {
	db := CreateTestDB(t)

	// Create test problem first
	problem := Problem{
		Name:             "test_problem",
		Title:            "Test Problem",
		Timelimit:        2000,
		TestCasesVersion: "v1.0",
		Version:          "1.0",
	}
	if err := db.Save(&problem).Error; err != nil {
		t.Fatal("Failed to save problem:", err)
	}

	// Create test submission first
	submission := Submission{
		ProblemName:      "test_problem",
		Lang:             "cpp",
		Status:           "WJ",
		Source:           "int main(){}",
		TestCasesVersion: "v1.0",
	}
	submissionID, err := SaveSubmission(db, submission)
	if err != nil {
		t.Fatal("Failed to save submission:", err)
	}

	// Create test hack first
	hack := Hack{
		SubmissionID: submissionID,
		TestCaseTxt:  []byte("test input"),
		Status:       "WJ",
	}
	hackID, err := SaveHack(db, hack)
	if err != nil {
		t.Fatal("Failed to save hack:", err)
	}

	// Test PushSubmissionTask with real submission ID
	submissionData := SubmissionData{ID: submissionID}
	if err := PushSubmissionTask(db, submissionData, 5); err != nil {
		t.Fatal("Failed to push submission task:", err)
	}

	// Test PushHackTask with real hack ID
	hackData := HackData{ID: hackID}
	if err := PushHackTask(db, hackData, 10); err != nil {
		t.Fatal("Failed to push hack task:", err)
	}

	// Pop hack task (higher priority)
	id1, data1, err := PopTask(db)
	if id1 == -1 || err != nil {
		t.Fatal("Failed to pop hack task:", id1, data1, err)
	}
	if hackResult, ok := data1.Data.(HackData); !ok || hackResult.ID != hackID {
		t.Fatal("Expected HackData with correct ID, got:", data1.Data)
	}

	// Pop submission task
	id2, data2, err := PopTask(db)
	if id2 == -1 || err != nil {
		t.Fatal("Failed to pop submission task:", id2, data2, err)
	}
	if submissionResult, ok := data2.Data.(SubmissionData); !ok || submissionResult.ID != submissionID {
		t.Fatal("Expected SubmissionData with correct ID, got:", data2.Data)
	}

	// Clean up
	if err := FinishTask(db, id1); err != nil {
		t.Fatal(err)
	}
	if err := FinishTask(db, id2); err != nil {
		t.Fatal(err)
	}
}
