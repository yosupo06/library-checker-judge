package database

import (
	"testing"
	"time"
)

func TestCanTouch(t *testing.T) {
	now := time.Now()
	if !canTouch(Task{
		Available: now,
		JudgeName: "judge1",
	}, now.Add(time.Second), "judge2") {
		t.Fatal("expected true")
	}
	if canTouch(Task{
		Available: now,
		JudgeName: "judge1",
	}, now.Add(-time.Second), "judge2") {
		t.Fatal("expected false")
	}
	if !canTouch(Task{
		Available: now,
		JudgeName: "judge1",
	}, now.Add(-time.Second), "judge1") {
		t.Fatal("expected true")
	}
}

func TestTask(t *testing.T) {
	db := createTestDB(t)

	if err := PushTask(db, 123, 1); err != nil {
		t.Fatal(err)
	}
	if err := PushTask(db, 789, 10); err != nil {
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
	db := createTestDB(t)

	if err := PushTask(db, 123, 10); err != nil {
		t.Fatal(err)
	}
	if err := PushTask(db, 124, 10); err != nil {
		t.Fatal(err)
	}
	if err := PushTask(db, 125, 10); err != nil {
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
