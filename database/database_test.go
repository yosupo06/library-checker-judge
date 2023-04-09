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
