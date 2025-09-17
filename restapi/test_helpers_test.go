package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/yosupo06/library-checker-judge/database"
	"gorm.io/gorm"
)

type fakeAuthClient struct {
	uid string
}

func (f fakeAuthClient) parseUID(_ context.Context, token string) string {
	if token == "" {
		return ""
	}
	return f.uid
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s.db?mode=memory&cache=shared&_busy_timeout=5000&_journal_mode=WAL", t.Name())
	dialector := sqlite.Open(dsn)
	db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}
