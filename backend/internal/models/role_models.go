package models

import (
	"time"
	// "gorm.io/gorm"
)

// PermissionBasic provides essential permission information.
// Used in RoleDetails.
type PermissionBasic struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RoleDetails extends Role with additional information like user count and detailed permissions.
// This structure is based on the API design for GET /roles/{roleId}.
// It assumes the base Role fields (ID, Name, Description, CreatedAt, UpdatedAt) will be populated
// from the existing models.Role defined in user.go or similar.
type RoleDetails struct {
	ID          uint              `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	UserCount   int               `json:"userCount"`   // As per API design
	Permissions []PermissionBasic `json:"permissions"` // As per API design
}

// RoleListParams defines parameters for listing/searching roles.
type RoleListParams struct {
	Page     int    `json:"page"`     // Added struct tags for consistency, though not strictly necessary for query params
	PageSize int    `json:"pageSize"` // Added struct tags
	Name     string `json:"name"`     // Added struct tags
}

// Note: The actual Permission model would be more complex,
// potentially with 'Resource' and 'Action' fields.
// For now, PermissionBasic is for the GetRoleByID response.
// type Permission struct {
//  gorm.Model
//  Name        string `json:"name" gorm:"uniqueIndex:idx_resource_action"` // e.g., "manage_users", "view_services"
//  Resource    string `json:"resource" gorm:"uniqueIndex:idx_resource_action"` // e.g., "users", "services"
//  Action      string `json:"action" gorm:"uniqueIndex:idx_resource_action"`   // e.g., "create", "read", "update", "delete"
//  Description string `json:"description"`
// }

// // TableName overrides the table name for Role
// func (Role) TableName() string {
// 	return "roles"
// }
