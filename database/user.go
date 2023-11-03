package database

import (
	"errors"

	"gorm.io/gorm"
)

// User is db table
type User struct {
	Name        string `gorm:"primaryKey" validate:"username"`
	UID         string `gorm:"not null;unique"`
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

func UpdateUser(db *gorm.DB, user User) error {
	if user.Name == "" || user.UID == "" {
		return errors.New("username / uid is empty")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if user2, err := FetchUserFromUID(tx, user.UID); err != nil || user2 == nil || user2.Name != user.Name {
			if err != nil {
				return err
			} else if user2 == nil {
				return errors.New("user not found")
			} else {
				return errors.New("username is differ")
			}
		}

		// TODO skip user name validation for exising user (with invalid user name)
		name := user.Name
		user.Name = "dummy"
		if err := validate.Struct(user); err != nil {
			return err
		}
		user.Name = name

		if err := db.Save(&user).Error; err != nil {
			return err
		}

		return nil
	})
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

func FetchUserFromName(db *gorm.DB, name string) (*User, error) {
	if name == "" {
		return nil, errors.New("User name is empty")
	}

	user := User{}
	if err := db.Where(&User{Name: name}).Take(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}
