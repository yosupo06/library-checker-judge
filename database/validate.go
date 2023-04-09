package database

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

const userNameRegexString = `^[A-Za-z0-9-_]+$`

var (
	userNameRegex = regexp.MustCompile(userNameRegexString)
)

var validate = validator.New()

func init() {
	validate.RegisterValidation("username", userNameValidator)
}

func userNameValidator(fl validator.FieldLevel) bool {
	return userNameRegex.MatchString(fl.Field().String())
}
