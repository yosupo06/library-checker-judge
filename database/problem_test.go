package database

import (
	"reflect"
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

	if problem3, err := FetchProblem(db, "aplusc"); problem3 != nil || err != nil {
		t.Fatal(problem3, err)
	}
}

func TestProblemCategory(t *testing.T) {
	db := createTestDB(t)

	categories := []ProblemCategory{
		{
			Title:    "Sample",
			Problems: []string{"aplusb", "many_aplusb"},
		},
		{
			Title:    "Data Structure",
			Problems: []string{"unionfind"},
		},
	}

	if err := SaveProblemCategories(db, categories); err != nil {
		t.Fatal(err)
	}

	categories2, err := FetchProblemCategories(db)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(categories, categories2) {
		t.Fatal(categories, "!=", categories2)
	}
}
