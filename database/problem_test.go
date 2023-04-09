package database

import (
	"os"
	"os/exec"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func createTestDB(t *testing.T) *gorm.DB {
	dbName := uuid.New().String()
	t.Log("create DB: ", dbName)

	createCmd := exec.Command("createdb",
		"-h", "localhost",
		"-U", "postgres",
		"-p", "5432",
		dbName)
	createCmd.Env = append(os.Environ(), "PGPASSWORD=passwd")
	if err := createCmd.Run(); err != nil {
		t.Fatal("exec failed: ", err.Error())
	}

	db := Connect("localhost", "5432", dbName, "postgres", "passwd", false)

	return db
}

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
