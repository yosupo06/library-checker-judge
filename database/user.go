package database

import (
	"errors"

	"gorm.io/gorm"
)

// User is db table
type User struct {
	Name        string `gorm:"primaryKey" validate:"username"`
	UID         string
	Passhash    string
	LibraryURL  string `validate:"libraryURL"`
	IsDeveloper bool
}

func RegisterUser(db *gorm.DB, name string, uid string) error {
	if uid == "" {
		return errors.New("UID is empty")
	}

	user := User{
		Name: name,
		UID:  uid,
	}
	if err := validate.Struct(user); err != nil {
		return err
	}

	if err := db.Create(&user).Error; err != nil {
		return errors.New("this username / uid is already registered")
	}
	return nil
}

func FetchUserFromUID(db *gorm.DB, uid string) (*User, error) {
	if uid == "" {
		return nil, errors.New("UID is empty")
	}

	user := User{}
	if err := db.Where(&User{UID: uid}).Take(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func FetchUser(db *gorm.DB, name string) (*User, error) {
	if name == "" {
		return nil, errors.New("User name is empty")
	}

	user := User{
		Name: name,
	}
	if err := db.Take(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func SaveUser(db *gorm.DB, user User) error {
	if err := db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

func UpdateUser(db *gorm.DB, user User) error {
	name := user.Name
	if name == "" {
		return errors.New("User name is empty")
	}

	// TODO skip user name validation for exising user (with invalid user name)
	user.Name = "dummy"
	if err := validate.Struct(user); err != nil {
		return err
	}
	user.Name = name

	result := db.Model(&user).Updates(
		map[string]interface{}{
			"library_url":  user.LibraryURL,
			"is_developer": user.IsDeveloper,
		})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return errors.New("User not found")
	}
	return nil
}
