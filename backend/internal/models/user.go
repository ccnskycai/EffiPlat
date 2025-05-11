package models

import (
	"time"
)

// User represents the users table in the database.
// We use pointers for fields that can be NULL in the DB or for zero values
// that need distinction (e.g., differentiating between 0 and not set).
// GORM tags define column names, types, constraints etc.
type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Email        string    `gorm:"unique;not null" json:"email"`
	Department   *string   `json:"department,omitempty"` // Pointer because it can be NULL, omitempty if nil
	PasswordHash string    `gorm:"not null" json:"-"`         // Do not include PasswordHash in JSON responses
	Status       string    `gorm:"not null;default:'active'" json:"status"`
	CreatedAt    time.Time `json:"createdAt"`              // GORM will handle this automatically
	UpdatedAt    time.Time `json:"updatedAt"`              // GORM will handle this automatically
	// Add associations later if needed, e.g.:
	// Roles        []Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// TableName explicitly sets the table name for GORM.
func (User) TableName() string {
	return "users"
}
