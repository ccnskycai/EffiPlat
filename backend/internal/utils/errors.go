package utils

import "errors"

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

// TODO: Add other common error types or helper functions if needed.
