package models

import (
	"time"
)

// Permission represents a permission in the system.
type Permission struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;uniqueIndex;not null"` // e.g., "user:read", "role:write"
	Description string    `json:"description,omitempty" gorm:"size:255"`
	Resource    string    `json:"resource,omitempty" gorm:"size:50;uniqueIndex:idx_resource_action;not null"` // e.g., "user", "role"
	Action      string    `json:"action,omitempty" gorm:"size:50;uniqueIndex:idx_resource_action;not null"`   // e.g., "list", "get", "create", "update", "delete"
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
	Roles       []Role    `json:"-" gorm:"many2many:role_permissions;"` // Many-to-many relationship with Role
}

// TableName specifies the table name for the Permission model.
func (Permission) TableName() string {
	return "permissions"
}

// RolePermission represents the join table for roles and permissions.
type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
}

// TableName specifies the table name for the RolePermission model.
func (RolePermission) TableName() string {
	return "role_permissions"
}

// Request and response structures for permissions

// CreatePermissionRequest is the request body for creating a permission.
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
}

// UpdatePermissionRequest is the request body for updating a permission.
type UpdatePermissionRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Resource    *string `json:"resource,omitempty"`
	Action      *string `json:"action,omitempty"`
}

// PermissionListParams defines parameters for listing/searching permissions.
type PermissionListParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Name     string `json:"name"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}
