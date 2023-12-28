package database

import (
	"database/sql"
	"reflect"
	"testing"
	"time"
)

func TestSubmission(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	if err := RegisterUser(db, "user1", "id1"); err != nil {
		t.Fatal(err)
	}

	sub := Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
	}

	id, err := SaveSubmission(db, sub)
	if err != nil {
		t.Fatal(err)
	}

	sub2, err := FetchSubmission(db, id)

	if err != nil {
		t.Fatal(err)
	}

	if sub2.User.Name != "user1" || sub2.Problem.Name != "aplusb" {
		t.Fatal("invalid data", sub2)
	}
}

func TestFetchInvalidSubmission(t *testing.T) {
	db := CreateTestDB(t)

	sub, err := FetchSubmission(db, 123)

	if err != nil {
		t.Fatal(err)
	}
	if sub != nil {
		t.Fatal("result should be null", sub)
	}
}

func TestSubmissionResult(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	if err := RegisterUser(db, "user1", "id1"); err != nil {
		t.Fatal(err)
	}

	sub := Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
	}

	id, err := SaveSubmission(db, sub)
	if err != nil {
		t.Fatal(err)
	}

	result := SubmissionTestcaseResult{
		Submission: id,
		Testcase:   "case1.in",
		Status:     "AC",
		Time:       123,
		Memory:     456,
		Stderr:     []byte{12, 34},
	}
	if err := SaveTestcaseResult(db, result); err != nil {
		t.Fatal(err)
	}

	actual, err := FetchTestcaseResults(db, id)
	if err != nil {
		t.Fatal(err)
	}

	if len(actual) != 1 || !reflect.DeepEqual(actual[0], result) {
		t.Fatal(actual, "!=", result)
	}
}

func TestSubmissionResultEmpty(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	if err := RegisterUser(db, "user1", "id1"); err != nil {
		t.Fatal(err)
	}

	sub := Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
	}
	if _, err := SaveSubmission(db, sub); err != nil {
		t.Fatal(err)
	}

	actual, err := FetchTestcaseResults(db, sub.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(actual) != 0 {
		t.Fatal(actual, "is not empty")
	}
}

func TestSubmissionLock(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	sub := Submission{
		ProblemName: "aplusb",
	}
	id, err := SaveSubmission(db, sub)
	if err != nil {
		t.Fatal(err)
	}

	if ok, err := TryLockSubmission(db, id, "judge"); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("lock failed")
	}
}

func TestSubmissionLockTwice(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	sub := Submission{
		ProblemName: "aplusb",
	}
	id, err := SaveSubmission(db, sub)
	if err != nil {
		t.Fatal(err)
	}

	if ok, err := TryLockSubmission(db, id, "judge"); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("lock failed")
	}

	if ok, err := TryLockSubmission(db, id, "judge"); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("lock failed")
	}
}

func TestSubmissionLockFailed(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	sub := Submission{
		ProblemName: "aplusb",
	}
	id, err := SaveSubmission(db, sub)
	if err != nil {
		t.Fatal(err)
	}

	if ok, err := TryLockSubmission(db, id, "judge1"); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("lock failed")
	}

	if ok, err := TryLockSubmission(db, id, "judge2"); err != nil {
		t.Fatal(err)
	} else if ok {
		t.Fatal("lock succeeded")
	}
}

func TestLeavedSubmissionLock(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	sub := Submission{
		ProblemName: "aplusb",
	}
	id, err := SaveSubmission(db, sub)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Save(&SubmissionLock{
		ID:   id,
		Name: "judge1",
		Ping: time.Now().Add(-time.Hour),
	}).Error; err != nil {
		t.Fatal(err)
	}

	if ok, err := TryLockSubmission(db, id, "judge2"); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("lock failed")
	}
}

func TestSubmissionLockUnlock(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	sub := Submission{
		ProblemName: "aplusb",
	}
	id, err := SaveSubmission(db, sub)
	if err != nil {
		t.Fatal(err)
	}

	if ok, err := TryLockSubmission(db, id, "judge1"); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("lock failed")
	}

	if err := UnlockSubmission(db, id, "judge1"); err != nil {
		t.Fatal(err)
	}

	if ok, err := TryLockSubmission(db, id, "judge2"); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("lock failed")
	}
}
