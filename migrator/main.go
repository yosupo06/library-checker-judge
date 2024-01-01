package main

import (
	"os"

	"github.com/yosupo06/library-checker-judge/database"
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	pgHost := os.Getenv("POSTGRE_HOST")
	pgUser := os.Getenv("POSTGRE_USER")
	pgPass := os.Getenv("POSTGRE_PASS")
	pgTable := os.Getenv("POSTGRE_TABLE")
	pgPort := os.Getenv("POSTGRE_PORT")
	if pgPort == "" {
		pgPort = "5432"
	}

	// connect db
	db := database.Connect(
		pgHost,
		pgPort,
		pgTable,
		pgUser,
		pgPass,
		false)
	db.AutoMigrate(database.Problem{})
	db.AutoMigrate(database.User{})
	db.AutoMigrate(database.Submission{})
	db.AutoMigrate(database.SubmissionTestcaseResult{})
	db.AutoMigrate(database.SubmissionLock{})
	db.AutoMigrate(database.Task{})
	db.AutoMigrate(database.Metadata{})

}
