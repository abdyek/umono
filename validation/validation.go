package validation

import (
	"regexp"

	val "github.com/go-playground/validator"
)

var Validator *validator

type validator struct {
	val *val.Validate
}

func Init() {
	Validator = &validator{
		val: val.New(),
	}

	Validator.val.RegisterValidation("slug", func(fl val.FieldLevel) bool {
		return regexp.MustCompile(`^[a-z]+(-[a-z]+)*$`).MatchString(fl.Field().String())
	})
}

func (v validator) Validate(data interface{}) bool {
	errs := v.val.Struct(data)
	if errs != nil {
		return false
	}

	return true
}
