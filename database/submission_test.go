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
	db := createTestDB(t)

	createDummyProblem(t, db)

	user := User{
		Name: "user1",
	}
	if err := SaveUser(db, user); err != nil {
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
