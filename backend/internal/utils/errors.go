package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ErrNotFound is a common error for when a resource is not found.
var ErrNotFound = errors.New("resource not found")

// Standard application errors
var (
	ErrAlreadyExists      = errors.New("resource already exists")
	ErrValidationFailed   = errors.New("validation failed")
	ErrInvalidCredentials = errors.New("invalid credentials") // Changed from "invalid email or password" for broader use
	ErrTokenGeneration    = errors.New("could not generate token")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrBadRequest         = errors.New("bad request")
	ErrInternalServer     = errors.New("internal server error")
	ErrUpdateFailed       = errors.New("update failed")         // General update error
	ErrDeleteFailed       = errors.New("delete failed")         // General delete error
	ErrMissingID          = errors.New("identifier is missing") // For operations requiring an ID
)

// FormatValidationError converts validator.ValidationErrors into a user-friendly string.
func FormatValidationError(err error) string {
	if err == nil {
		return ""
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		var errMsgs []string
		for _, fieldErr := range validationErrors {
			// You can customize the error messages based on fieldErr.Tag(), fieldErr.Field(), etc.
			errMsgs = append(errMsgs, fmt.Sprintf("Field '%s' failed on the '%s' tag", fieldErr.StructNamespace(), fieldErr.Tag()))
		}
		return strings.Join(errMsgs, "; ")
	}

	// If it's not a validator.ValidationErrors, return the original error message.
	return err.Error()
}

// TODO: Add other common error types or helper functions if needed.
