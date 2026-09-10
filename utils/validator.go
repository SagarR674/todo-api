package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()
	// Report errors using the json tag name rather than the Go field name.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return v
}

// ValidateStruct runs validation and returns a field -> message map, or nil
// when the input is valid.
func ValidateStruct(s interface{}) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var invalid *validator.InvalidValidationError
	if reflect.TypeOf(err) == reflect.TypeOf(invalid) {
		return map[string]string{"_": "invalid validation target"}
	}

	out := make(map[string]string)
	for _, fe := range err.(validator.ValidationErrors) {
		out[fe.Field()] = messageFor(fe)
	}
	return out
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", strings.ReplaceAll(fe.Param(), " ", ", "))
	case "datetime":
		return fmt.Sprintf("must be a valid date in %s format", datePattern(fe.Param()))
	case "gte":
		return fmt.Sprintf("must be %s or greater", fe.Param())
	case "lte":
		return fmt.Sprintf("must be %s or less", fe.Param())
	default:
		return fmt.Sprintf("failed the %q rule", fe.Tag())
	}
}

// layoutPatterns rewrites a Go reference-time layout (the parameter of the
// validator's `datetime` rule) into the conventional placeholder form used in
// the API docs, so error messages never leak Go's reference date.
var layoutPatterns = strings.NewReplacer(
	"2006", "YYYY",
	"01", "MM",
	"02", "DD",
	"15", "HH",
	"04", "mm",
	"05", "ss",
)

// datePattern turns a layout such as "2006-01-02" into "YYYY-MM-DD".
func datePattern(layout string) string {
	return layoutPatterns.Replace(layout)
}
