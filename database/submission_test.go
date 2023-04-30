package database

import (
	"database/sql"
	"reflect"
	"testing"
)

func TestSubmission(t *testing.T) {
	db := createTestDB(t)

	createDummyProblem(t, db)

	user := User{
		Name: "user1",
	}
	if err := SaveUser(db, user); err != nil {
		t.Fatal(err)
	}

	sub := Submission{
		ID:          123,
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
	}
	if err := SaveSubmission(db, sub); err != nil {
		t.Fatal(err)
	}

	sub2, err := FetchSubmission(db, 123)

	if err != nil {
		t.Fatal(err)
	}

	if sub2.User.Name != "user1" || sub2.Problem.Name != "aplusb" {
		t.Fatal("invalid data", sub2)
	}
}

func TestSubmissionResult(t *testing.T) {
	db := createTestDB(t)

	createDummyProblem(t, db)

	user := User{
		Name: "user1",
	}
	if err := SaveUser(db, user); err != nil {
		t.Fatal(err)
	}

	sub := Submission{
		ID:          123,
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
	}
	if err := SaveSubmission(db, sub); err != nil {
		t.Fatal(err)
	}

	result := SubmissionTestcaseResult{
		Submission: sub.ID,
		Testcase:   "case1.in",
		Status:     "AC",
		Time:       123,
		Memory:     456,
		Stderr:     []byte{12, 34},
	}
	if err := SaveTestcaseResult(db, result); err != nil {
		t.Fatal(err)
	}

	actual, err := FetchTestcaseResults(db, sub.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(actual) != 1 || !reflect.DeepEqual(actual[0], result) {
		t.Fatal(actual, "!=", result)
	}
}

func TestSubmissionResultEmpty(t *testing.T) {
	db := createTestDB(t)

	createDummyProblem(t, db)

	user := User{
		Name: "user1",
	}
	if err := SaveUser(db, user); err != nil {
		t.Fatal(err)
	}

	sub := Submission{
		ID:          123,
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
	}
	if err := SaveSubmission(db, sub); err != nil {
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
