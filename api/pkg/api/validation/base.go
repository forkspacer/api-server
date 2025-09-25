package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	Validate *validator.Validate
)

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())

	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func validationErrors2Map(ve validator.ValidationErrors) map[string]string {
	errors := make(map[string]string)
	for _, err := range ve {
		errors[err.Field()] = err.Error()
	}
	return errors
}
