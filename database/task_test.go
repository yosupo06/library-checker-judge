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

	id1, data1, err := PopTask(db)
	if id1 == -1 || data1.Submission != 789 || err != nil {
		t.Fatal(id1, data1, err)
	}

	id2, data2, err := PopTask(db)
	if id2 == -1 || data2.Submission != 123 || err != nil {
		t.Fatal(id2, data2, err)
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

	id1, data1, err := PopTask(db)
	if id1 == -1 || data1.Submission != 123 || err != nil {
		t.Fatal(id1, data1, err)
	}

	id2, data2, err := PopTask(db)
	if id2 == -1 || data2.Submission != 124 || err != nil {
		t.Fatal(id2, data2, err)
	}

	id3, data3, err := PopTask(db)
	if id3 == -1 || data3.Submission != 125 || err != nil {
		t.Fatal(data3, err)
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
