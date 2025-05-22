package models

import (
	"time"
)

// User represents the user model in the database
type User struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name" gorm:"size:100;not null"`
	Email      string    `json:"email" gorm:"size:100;uniqueIndex;not null"`
	Password   string    `json:"-" gorm:"column:password_hash;size:255;not null"` // Changed tag to map to password_hash
	Department string    `json:"department,omitempty" gorm:"size:100"`
	Status     string    `json:"status,omitempty" gorm:"size:20;default:'pending'"` // e.g., active, inactive, pending
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
	Roles      []Role    `json:"roles,omitempty" gorm:"many2many:user_roles;"` // Many-to-many relationship with Role
	// AssignedResponsibilities []Responsibility `json:"assignedResponsibilities,omitempty" gorm:"-"` // Placeholder, implementation depends on Responsibility model and join table
}

// Role represents a user role
type Role struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"size:50;uniqueIndex;not null"`
	Description string       `json:"description,omitempty" gorm:"size:255"`
	Users       []User       `json:"-" gorm:"many2many:user_roles;"`                           // Many-to-many relationship with User
	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:role_permissions;"` // Many-to-many relationship with Permission
	CreatedAt   time.Time    `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for the User model.
func (User) TableName() string {
	return "users"
}

// TableName specifies the table name for the Role model.
func (Role) TableName() string {
	return "roles"
}

// UserRole represents the join table for users and roles.
type UserRole struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

// TableName specifies the table name for the UserRole model.
func (UserRole) TableName() string {
	return "user_roles"
}

// AssignRemoveRolesRequest defines the structure for assigning or removing roles for/from a user.
type AssignRemoveRolesRequest struct {
	RoleIDs []uint `json:"role_ids" binding:"required,dive,gte=1"` // dive ensures each uint in slice is >= 1
}

// UserListParams defines parameters for listing users with pagination and filtering.
type UserListParams struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Name     string `form:"name,omitempty"`
	Email    string `form:"email,omitempty"`
	Status   string `form:"status,omitempty"`
	SortBy   string `form:"sortBy,omitempty"` // e.g., "name", "email", "createdAt"
	Order    string `form:"order,omitempty"`  // e.g., "asc", "desc"
}

// Add other models like Responsibility if needed
