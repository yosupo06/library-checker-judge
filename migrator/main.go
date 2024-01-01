package main

import (
	"log/slog"
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

	if err := database.AutoMigrate(db); err != nil {
		slog.Error("Migration failed:", err)
	}
}
