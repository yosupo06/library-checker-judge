package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
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

var ErrNotExist = errors.New("not exist")

type DSN struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

var DEFAULT_DSN = DSN{
	Host:     "localhost",
	Port:     5432,
	Database: "librarychecker",
	User:     "postgres",
	Password: "lcdummypassword",
}

func GetDSNFromEnv() DSN {
	dsn := DEFAULT_DSN
	if host := os.Getenv("PGHOST"); host != "" {
		dsn.Host = host
	}
	if portStr := os.Getenv("PGPORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err != nil {
			log.Println("Parse PGPORT failed:", portStr)
		} else {
			dsn.Port = port
		}
	}
	if database := os.Getenv("PGDATABASE"); database != "" {
		dsn.Database = database
	}
	if user := os.Getenv("PGUSER"); user != "" {
		dsn.User = user
	}
	if password := os.Getenv("PGPASSWORD"); password != "" {
		dsn.Password = password
	}
	return dsn
}

func Connect(dsn DSN, enableLogger bool) *gorm.DB {
	connStr := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		dsn.Host, dsn.Port, dsn.Database, dsn.User, dsn.Password)
	log.Printf("try to connect db, host=%s port=%d dbname=%s user=%s", dsn.Host, dsn.Port, dsn.Database, dsn.User)
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

	dsn := GetDSNFromEnv()
	dsn.Database = dbName

	createCmd := exec.Command("createdb",
		"-h", dsn.Host,
		"-U", dsn.User,
		"-p", strconv.Itoa(dsn.Port),
		dbName)
	createCmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", dsn.Password))
	if err := createCmd.Run(); err != nil {
		t.Fatal("createdb failed: ", err)
	}

	db := Connect(dsn, os.Getenv("API_DB_LOG") != "")

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
			"-h", dsn.Host,
			"-U", dsn.User,
			"-p", strconv.Itoa(dsn.Port),
			dbName)
		createCmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", dsn.Password))
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
	if err := db.AutoMigrate(Task{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(Metadata{}); err != nil {
		return err
	}
	return nil
}
