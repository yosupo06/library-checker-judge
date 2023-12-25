package database

import (
	"reflect"
	"testing"

	"gorm.io/gorm"
)

func createDummyProblem(t *testing.T, db *gorm.DB) {
	problem := Problem{
		Name:             "aplusb",
		Title:            "Title",
		SourceUrl:        "url",
		Statement:        "statement",
		Timelimit:        123,
		TestCasesVersion: "tversion123",
		Version:          "version456",
	}
	if err := SaveProblem(db, problem); err != nil {
		t.Fatal(err)
	}
}

func TestProblemInfo(t *testing.T) {
	db := CreateTestDB(t)
	createDummyProblem(t, db)

	problem, err := FetchProblem(db, "aplusb")
	if err != nil {
		t.Fatal(err)
	}
	expect := Problem{
		Name:             "aplusb",
		Title:            "Title",
		SourceUrl:        "url",
		Statement:        "statement",
		Timelimit:        123,
		TestCasesVersion: "tversion123",
		Version:          "version456",
	}
	if *problem != expect {
		t.Fatal(problem, "!=", expect)
	}

	if problem3, err := FetchProblem(db, "aplusc"); problem3 != nil || err != nil {
		t.Fatal(problem3, err)
	}
}

func TestProblemCategory(t *testing.T) {
	db := CreateTestDB(t)

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
