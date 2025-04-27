package database

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

const (
	MAX_SOURCE_SIZE        = 1024 * 1024
	MAX_USER_NAME_LENGTH   = 30
	USER_NAME_REGEX_STRING = `^[A-Za-z0-9-_]+$`
)

var (
	userNameRegex = regexp.MustCompile(USER_NAME_REGEX_STRING)
)

var validate = validator.New()

func init() {
	validate.RegisterValidation("username", userNameValidator)
	validate.RegisterAlias("libraryURL", "omitempty,url,lt=200")
	validate.RegisterValidation("source", sourceValidator)
}

func userNameValidator(fl validator.FieldLevel) bool {
	name := fl.Field().String()

	if len(name) > MAX_USER_NAME_LENGTH {
		return false
	}

	return userNameRegex.MatchString(name)
}

func sourceValidator(fl validator.FieldLevel) bool {
	source := fl.Field().String()

	if source == "" {
		return false
	}

	return len(source) <= MAX_SOURCE_SIZE
}
