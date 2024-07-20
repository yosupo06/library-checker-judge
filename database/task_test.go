package database

import (
	"testing"
)

func TestTask(t *testing.T) {
	db := CreateTestDB(t)

	if err := PushTask(db, TaskData{
		TaskType:   JUDGE_SUBMISSION,
		Submission: 123,
	}, 1); err != nil {
		t.Fatal(err)
	}
	if err := PushTask(db, TaskData{
		TaskType:   JUDGE_SUBMISSION,
		Submission: 789,
	}, 10); err != nil {
		t.Fatal(err)
	}

	task1, err := PopTask(db, "judge")
	if task1 == nil || task1.Submission != 789 || err != nil {
		t.Fatal(task1, err)
	}

	task2, err := PopTask(db, "judge")
	if task2 == nil || task2.Submission != 123 || err != nil {
		t.Fatal(task2, err)
	}

	if task, err := PopTask(db, "judge"); task != nil || err != nil {
		t.Fatal(task, err)
	}

	if err := FinishTask(db, task1.ID); err != nil {
		t.Fatal(err)
	}
	if err := FinishTask(db, task2.ID); err != nil {
		t.Fatal(err)
	}
}

func TestTaskSamePriority(t *testing.T) {
	db := CreateTestDB(t)

	if err := PushTask(db, TaskData{
		TaskType:   JUDGE_SUBMISSION,
		Submission: 123,
	}, 10); err != nil {
		t.Fatal(err)
	}
	if err := PushTask(db, TaskData{
		TaskType:   JUDGE_SUBMISSION,
		Submission: 124,
	}, 10); err != nil {
		t.Fatal(err)
	}
	if err := PushTask(db, TaskData{
		TaskType:   JUDGE_SUBMISSION,
		Submission: 125,
	}, 10); err != nil {
		t.Fatal(err)
	}

	task1, err := PopTask(db, "judge")
	if task1 == nil || task1.Submission != 123 || err != nil {
		t.Fatal(task1, err)
	}

	task2, err := PopTask(db, "judge")
	if task2 == nil || task2.Submission != 124 || err != nil {
		t.Fatal(task2, err)
	}

	task3, err := PopTask(db, "judge")
	if task3 == nil || task3.Submission != 125 || err != nil {
		t.Fatal(task3, err)
	}
}

func TestTaskDataSerialize(t *testing.T) {
	task := TaskData{
		TaskType:   JUDGE_SUBMISSION,
		Submission: 123,
	}
	buf, err := encode(task)
	if err != nil {
		t.Fatal(err)
	}
	task2, err := decode(buf)
	if err != nil {
		t.Fatal(err)
	}
	if task != task2 {
		t.Fatal(task, task2)
	}
}
