package database

import (
	"errors"

	"gorm.io/gorm"
)

type Metadata struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

func FetchMetadata(db *gorm.DB, key string) (*string, error) {
	if key == "" {
		return nil, errors.New("empty key")
	}
	metadata := Metadata{Key: key}
	if err := db.First(&metadata).Error; err != nil {
		return nil, err
	}
	return &metadata.Value, nil
}

func SaveMetadata(db *gorm.DB, key string, value string) error {
	if key == "" {
		return errors.New("emtpy key")
	}
	if err := db.Save(&Metadata{
		Key:   key,
		Value: value,
	}).Error; err != nil {
		return err
	}
	return nil
}
