package database

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	MAX_TRY_TIMES = 3
)

func Connect(host, port, dbname, user, pass string, enableLogger bool) *gorm.DB {
	connStr := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		host, port, dbname, user, pass)
	log.Printf("try to connect db, host=%s port=%s dbname=%s user=%s", host, port, dbname, user)
	for i := 0; i < MAX_TRY_TIMES; i++ {
		newLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		)
		config := gorm.Config{
			Logger: newLogger,
		}
		if enableLogger {
			config.Logger.LogMode(logger.Info)
		}
		db, err := gorm.Open(postgres.Open(connStr), &config)
		if err != nil {
			log.Printf("cannot connect db %d/%d", i, MAX_TRY_TIMES)
			time.Sleep(5 * time.Second)
			continue
		}
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatal("db.DB() failed")
		}

		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetConnMaxLifetime(time.Hour)

		return db
	}
	log.Fatalf("cannot connect db %d times", MAX_TRY_TIMES)
	return nil
}

func CreateTestDB(t *testing.T) *gorm.DB {
	dbName := uuid.New().String()
	t.Log("create DB: ", dbName)

	createCmd := exec.Command("createdb",
		"-h", "localhost",
		"-U", "postgres",
		"-p", "5432",
		dbName)
	createCmd.Env = append(os.Environ(), "PGPASSWORD=passwd")
	if err := createCmd.Run(); err != nil {
		t.Fatal("createdb failed: ", err.Error())
	}

	db := Connect("localhost", "5432", dbName, "postgres", "passwd", os.Getenv("API_DB_LOG") != "")
	if err := AutoMigrate(db); err != nil {
		t.Fatal("Migration failed:", err)
	}

	t.Cleanup(func() {
		db2, err := db.DB()
		if err != nil {
			t.Fatal("db.DB() failed:", err)
		}
		if err := db2.Close(); err != nil {
			t.Fatal("db.Close() failed:", err)
		}
		createCmd := exec.Command("dropdb",
			"-h", "localhost",
			"-U", "postgres",
			"-p", "5432",
			dbName)
		createCmd.Env = append(os.Environ(), "PGPASSWORD=passwd")
		createCmd.Stderr = os.Stderr
		createCmd.Stdin = os.Stdin
		if err := createCmd.Run(); err != nil {
			t.Fatal("dropdb failed:", err)
		}
	})

	return db
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(Problem{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(User{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(Submission{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(SubmissionTestcaseResult{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(SubmissionLock{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(Task{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(Metadata{}); err != nil {
		return err
	}
	return nil
}
