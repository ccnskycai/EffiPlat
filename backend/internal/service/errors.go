package service

import "errors"

var (
	ErrNotFound                    = errors.New("requested resource not found")
	ErrResponsibilityNotFound      = errors.New("responsibility not found")
	ErrResponsibilityGroupNotFound = errors.New("responsibility group not found")
	ErrValidationFailed            = errors.New("validation failed")
	ErrAlreadyExists               = errors.New("resource already exists")
	// Add other common service errors here
)
