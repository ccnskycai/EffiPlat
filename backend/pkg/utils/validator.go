package utils

import (
	"fmt"
	"strings"

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