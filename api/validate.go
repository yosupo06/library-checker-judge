package main

import (
	"github.com/go-playground/validator/v10"
	"github.com/yosupo06/library-checker-judge/langs"
)

const (
	maxSourceSize = 1024 * 1024 // Use lowerCamelCase for unexported const if preferred, or keep UpperCamelCase
)

var validate *validator.Validate // Make it a pointer to initialize in init

func init() {
	validate = validator.New()
	validate.RegisterValidation("source", sourceValidator)
	validate.RegisterValidation("lang", langValidator)
}

func sourceValidator(fl validator.FieldLevel) bool {
	source := fl.Field().String()
	if source == "" {
		return false
	}
	return len(source) <= maxSourceSize
}

func langValidator(fl validator.FieldLevel) bool {
	lang := fl.Field().String()
	_, ok := langs.GetLang(lang)
	return ok
}
