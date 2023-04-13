package database

import (
	"database/sql"
	"testing"
)

func TestSubmission(t *testing.T) {
	db := createTestDB(t)

	problem := Problem{
		Name:      "aplusb",
		Title:     "Title",
		SourceUrl: "url",
		Statement: "statement",
		Timelimit: 123,
		Testhash:  "2345",
	}
	user := User{
		Name: "user1",
	}

	if err := SaveProblem(db, problem); err != nil {
		t.Fatal(err)
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
