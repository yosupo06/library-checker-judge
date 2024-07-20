package main

import (
	"log/slog"

	"github.com/yosupo06/library-checker-judge/database"
)

func main() {
	db := database.Connect(database.GetDSNFromEnv(), false)

	if err := database.AutoMigrate(db); err != nil {
		slog.Error("Migration failed:", slog.Any("err", err))
	}
}
