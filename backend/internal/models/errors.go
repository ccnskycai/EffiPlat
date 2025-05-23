package models

import "errors"

// General errors
var (
	ErrNotFound         = errors.New("requested resource not found")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrBadRequest       = errors.New("bad request")
	ErrConflict         = errors.New("resource conflict") // General conflict
	ErrInternalServer   = errors.New("internal server error")
	ErrValidationFailed = errors.New("validation failed")
)

// User specific errors
var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user with this email already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrPasswordTooShort      = errors.New("password is too short")
	ErrRoleAssignment        = errors.New("error assigning roles to user")
	ErrRoleRemoval           = errors.New("error removing roles from user")
	ErrUserHasActiveSessions = errors.New("user has active sessions, cannot delete")
)

// Role specific errors
var (
	ErrRoleNotFound      = errors.New("role not found")
	ErrRoleAlreadyExists = errors.New("role with this name already exists")
	ErrPermissionToRole  = errors.New("error assigning permission to role")
)

// Permission specific errors
var (
	ErrPermissionNotFound      = errors.New("permission not found")
	ErrPermissionAlreadyExists = errors.New("permission with this name already exists")
)

// Responsibility specific errors
var (
	ErrResponsibilityNotFound      = errors.New("responsibility not found")
	ErrResponsibilityAlreadyExists = errors.New("responsibility with this name already exists")
)

// ResponsibilityGroup specific errors
var (
	ErrResponsibilityGroupNotFound      = errors.New("responsibility group not found")
	ErrResponsibilityGroupAlreadyExists = errors.New("responsibility group with this name already exists")
	ErrAddResponsibilityToGroup         = errors.New("error adding responsibility to group")
	ErrRemoveResponsibilityFromGroup    = errors.New("error removing responsibility from group")
)

// Environment specific errors
var (
	ErrEnvironmentNotFound      = errors.New("environment not found")
	ErrEnvironmentNameExists    = errors.New("environment name already exists")
	ErrEnvironmentSlugExists    = errors.New("environment slug already exists")
)

// Asset specific errors
var (
	ErrAssetNotFound      = errors.New("asset not found")
	ErrAssetNameExists    = errors.New("asset name already exists")
	ErrAssetIdentifierExists = errors.New("asset identifier already exists")
)

// Service specific errors
var (
	ErrServiceNotFound         = errors.New("service not found")
	ErrServiceNameExists       = errors.New("service with this name already exists")
	ErrServiceTypeNotFound     = errors.New("service type not found")
	ErrServiceTypeNameExists   = errors.New("service type with this name already exists")
	ErrInvalidServiceStatus    = errors.New("invalid service status")
	ErrServiceTypeInUse        = errors.New("service type is in use and cannot be deleted")
)

// ServiceInstance specific errors (Placeholder for future use)
var (
	ErrServiceInstanceNotFound = errors.New("service instance not found")
	// Add more as needed
)
