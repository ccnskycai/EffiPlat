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
	Status     string    `json:"status" gorm:"size:20;default:'active';index"` // e.g., active, inactive, pending
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

// Add other models like Responsibility if needed
