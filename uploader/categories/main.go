package main

import (
	"flag"
	"log/slog"
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/yosupo06/library-checker-judge/database"
	"gorm.io/gorm"
)

func main() {
	problemsDir := flag.String("dir", "../../library-checker-problems", "directory of library-checker-problems")
	flag.Parse()

	db := database.Connect(database.GetDSNFromEnv(), false)

	if err := uploadCategories(*problemsDir, db); err != nil {
		slog.Error("Failed to update categories", "err", err)
		os.Exit(1)
	}
}

func uploadCategories(dir string, db *gorm.DB) error {
	var data struct {
		Categories []struct {
			Name     string
			Problems []string
		}
	}
	if _, err := toml.DecodeFile(path.Join(dir, "categories.toml"), &data); err != nil {
		return err
	}

	cs := []database.ProblemCategory{}
	for _, c := range data.Categories {
		cs = append(cs, database.ProblemCategory{Title: c.Name, Problems: c.Problems})
	}
	return database.SaveProblemCategories(db, cs)
}
