package models

import (
	"time"
)

// User represents the users table in the database.
// We use pointers for fields that can be NULL in the DB or for zero values
// that need distinction (e.g., differentiating between 0 and not set).
// GORM tags define column names, types, constraints etc.
type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	Name         string    `gorm:"not null"`
	Email        string    `gorm:"unique;not null"`
	Department   *string   // Pointer because it can be NULL
	PasswordHash string    `gorm:"not null"`
	Status       string    `gorm:"not null;default:'active'"`
	CreatedAt    time.Time // GORM will handle this automatically
	UpdatedAt    time.Time // GORM will handle this automatically
	// Add associations later if needed, e.g.:
	// Roles        []Role `gorm:"many2many:user_roles;"`
}

// TableName explicitly sets the table name for GORM.
func (User) TableName() string {
	return "users"
}
