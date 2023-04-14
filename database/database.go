package database

import (
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(host, port, dbname, user, pass string, enableLogger bool) *gorm.DB {
	connStr := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		host, port, dbname, user, pass)
	log.Printf("Try connect %s", connStr)
	for i := 0; i < 3; i++ {
		config := gorm.Config{}
		if enableLogger {
			config.Logger = logger.Default.LogMode(logger.Info)
		}
		db, err := gorm.Open(postgres.Open(connStr), &config)
		if err != nil {
			log.Printf("Cannot connect db %d/3", i)
			time.Sleep(5 * time.Second)
			continue
		}
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatal("db.DB() failed")
		}
		db.AutoMigrate(Problem{})
		db.AutoMigrate(User{})
		db.AutoMigrate(Submission{})
		db.AutoMigrate(SubmissionTestcaseResult{})
		db.AutoMigrate(Task{})
		db.AutoMigrate(Metadata{})

		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetConnMaxLifetime(time.Hour)

		return db
	}
	log.Fatal("Cannot connect db 3 times")
	return nil
}

func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", errors.New("bcrypt broken")
	}
	return string(hash), nil
}
