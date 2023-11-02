package database

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

const (
	maxUserNameLength   = 30
	userNameRegexString = `^[A-Za-z0-9-_]+$`
)

var (
	userNameRegex = regexp.MustCompile(userNameRegexString)
)

var validate = validator.New()

func init() {
	validate.RegisterValidation("username", userNameValidator)
	validate.RegisterAlias("libraryURL", "omitempty,url,lt=200")
}

func userNameValidator(fl validator.FieldLevel) bool {
	name := fl.Field().String()

	if maxUserNameLength < len(name) {
		return false
	}

	return userNameRegex.MatchString(name)
}
