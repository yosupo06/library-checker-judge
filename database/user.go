package database

import (
	"errors"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User is db table
type User struct {
	Name        string `gorm:"primaryKey"`
	Passhash    string
	Admin       bool
	Email       string
	LibraryURL  string
	IsDeveloper bool
}

func RegisterUser(db *gorm.DB, name string, password string, isAdmin bool) error {
	type UserParam struct {
		User     string `validate:"username"`
		Password string `validate:"required"`
	}

	userParam := &UserParam{
		User:     name,
		Password: password,
	}
	if err := validate.Struct(userParam); err != nil {
		return err
	}

	passHash, err := generatePasswordHash(password)
	if err != nil {
		return err
	}
	user := User{
		Name:     name,
		Passhash: string(passHash),
		Admin:    isAdmin,
	}
	if err := db.Create(&user).Error; err != nil {
		return errors.New("this username is already registered")
	}
	return nil
}

func VerifyUserPassword(db *gorm.DB, name string, password string) error {
	user, err := FetchUser(db, name)
	if user == nil {
		return fmt.Errorf("User %s not found", name)
	}
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Passhash), []byte(password)); err != nil {
		return errors.New("password invalid")
	}

	return nil
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
	result := db.Model(&User{}).Where("name = ?", name).Updates(
		map[string]interface{}{
			"admin":        user.Admin,
			"email":        user.Email,
			"library_url":  user.LibraryURL,
			"is_developer": user.IsDeveloper,
		})
	if err := result.Error; err != nil {
		log.Print(err)
		return errors.New("failed to update user")
	}
	if result.RowsAffected == 0 {
		return errors.New("User not found")
	}
	return nil
}
