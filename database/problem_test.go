package database

import (
	"testing"
)

func TestProblemInfoFail(t *testing.T) {
	db := createTestDB(t)

	_, err := FetchProblem(db, "invalid")

	if err == nil {
		t.Fatal("fetch succeeded")
	}

	t.Log("expected failure:", err)
}

func TestProblemInfo(t *testing.T) {
	db := createTestDB(t)

	problem := Problem{
		Name:      "aplusb",
		Title:     "Title",
		SourceUrl: "url",
		Statement: "statement",
		Timelimit: 123,
		Testhash:  "2345",
	}
	if err := SaveProblem(db, problem); err != nil {
		t.Fatal(err)
	}

	problem2, err := FetchProblem(db, "aplusb")
	if err != nil {
		t.Fatal(err)
	}
	if problem != *problem2 {
		t.Fatal(problem, "!=", problem2)
	}
}
