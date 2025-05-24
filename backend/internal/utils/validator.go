package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// FieldError wraps a validation error for a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FormatValidationErrors converts validator.ValidationErrors into a more user-friendly format.
// It can return a slice of FieldError or a map, depending on preference.
// Returning a string for simplicity here, but a slice of FieldError is often better for APIs.
func FormatValidationErrors(err error) string {
	if verr, ok := err.(validator.ValidationErrors); ok {
		var errs []string
		for _, fe := range verr {
			// Customize error messages based on fe.Tag()
			var msg string
			switch fe.Tag() {
			case "required":
				msg = fmt.Sprintf("Field '%s' is required", fe.Field())
			case "email":
				msg = fmt.Sprintf("Field '%s' must be a valid email address", fe.Field())
			case "min":
				msg = fmt.Sprintf("Field '%s' must be at least %s characters long", fe.Field(), fe.Param())
			case "max":
				msg = fmt.Sprintf("Field '%s' must be at most %s characters long", fe.Field(), fe.Param())
			case "oneof":
				msg = fmt.Sprintf("Field '%s' must be one of [%s]", fe.Field(), fe.Param())
			default:
				msg = fmt.Sprintf("Field '%s' validation failed on tag '%s'", fe.Field(), fe.Tag())
			}
			errs = append(errs, msg)
		}
		return strings.Join(errs, "; ")
	}
	return err.Error() // Fallback for non-validator errors
}

var (
	validateInstance *validator.Validate
	once             sync.Once
	alphanumDashRegex = regexp.MustCompile("^[a-zA-Z0-9-]+$")
)

// defaultValidator implements the gin.StructValidator interface.
type defaultValidator struct {
	validate *validator.Validate
}

// ValidateStruct receives Struct level validations.
func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	if obj == nil {
		return nil
	}

	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		return v.ValidateStruct(value.Elem().Interface())
	case reflect.Struct:
		return v.validate.Struct(obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		for i := 0; i < count; i++ {
			if err := v.ValidateStruct(value.Index(i).Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

// Engine returns the underlying validator engine.
func (v *defaultValidator) Engine() interface{} {
	return v.validate
}

// validateAlphanumDash validates that a string contains only alphanumeric characters and dashes
func validateAlphanumDash(fl validator.FieldLevel) bool {
	return alphanumDashRegex.MatchString(fl.Field().String())
}

// GetValidator returns a singleton instance of the validator.Validate engine.
// It also configures gin to use this validator for binding.
func GetValidator() *validator.Validate {
	once.Do(func() {
		validateInstance = validator.New()

		// Configure gin to use this validator instance
		binding.Validator = &defaultValidator{validate: validateInstance}

		// Register custom validators
		_ = validateInstance.RegisterValidation("alphanumdash", validateAlphanumDash)

		// Other custom validations (like is_business_status) should be registered
		// where GetValidator() is first called and where those validation funcs are defined,
		// typically in the service constructor that uses them.
	})
	return validateInstance
}
