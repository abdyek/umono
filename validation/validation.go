package validation

import (
	"regexp"

	val "github.com/go-playground/validator"
	"github.com/umono-cms/umono/utils/inarr"
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
		if fl.Field().String() == "" {
			return true
		}

		if inarr.String(fl.Field().String(), []string{"api", "admin"}) {
			return false
		}

		return regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`).MatchString(fl.Field().String())
	})

	Validator.val.RegisterValidation("numeric-screaming-snake-case", func(fl val.FieldLevel) bool {
		re := regexp.MustCompile(`^[A-Z0-9]+(?:_[A-Z0-9]+)*$`)
		return re.MatchString(fl.Field().String())
	})
}

func (v validator) Validate(data interface{}) bool {
	errs := v.val.Struct(data)
	if errs != nil {
		return false
	}

	return true
}
