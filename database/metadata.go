package database

import (
	"errors"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Metadata struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

func FetchMetadata(db *gorm.DB, key string) (*string, error) {
	if key == "" {
		return nil, errors.New("empty key")
	}
	metadata := Metadata{}
	if err := db.Where("key = ?", key).Take(&metadata).Error; err != nil {
		return nil, errors.New("metadata not found")
	}
	return &metadata.Value, nil
}

func SaveMetadata(db *gorm.DB, key string, value string) error {
	if key == "" {
		return errors.New("emtpy key")
	}
	if err := db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&Metadata{
		Key:   key,
		Value: value,
	}).Error; err != nil {
		log.Print(err)
		return errors.New("metadata upsert failed")
	}
	return nil
}
