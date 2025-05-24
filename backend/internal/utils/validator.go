package utils

import (
	"reflect"
	"regexp"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

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
