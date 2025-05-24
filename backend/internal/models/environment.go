package models

import (
	"time"

	"gorm.io/gorm"
)

// Environment represents a deployment or operational environment (e.g., dev, test, prod).
type Environment struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	Name        string `gorm:"type:varchar(100);uniqueIndex;not null" json:"name" binding:"required,min=2,max=100"`
	Description string `gorm:"type:text" json:"description"`
	Slug        string `gorm:"type:varchar(50);uniqueIndex;not null" json:"slug" binding:"required,min=2,max=50"`
	// SortOrder   int            `gorm:"default:100" json:"sortOrder"` // Optional: for ordering environments in UI
	// IsActive    bool           `gorm:"default:true" json:"isActive"`   // Optional: to activate/deactivate an environment
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName specifies the table name for the Environment model.
func (Environment) TableName() string {
	return "environments"
}

// --- Request/Response Structs for Environment Handler ---

// CreateEnvironmentRequest defines the structure for creating a new environment.
type CreateEnvironmentRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty"`                   // Description is optional
	Slug        string `json:"slug" validate:"required,min=2,max=50,alphanumdash"` // Slug should be URL-friendly
}

// UpdateEnvironmentRequest defines the structure for updating an existing environment.
// All fields are optional, so pointers are used for distinguishing between empty and not provided.
type UpdateEnvironmentRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty"`
	Slug        *string `json:"slug" validate:"omitempty,min=2,max=50,alphanumdash"`
}

// EnvironmentResponse defines a standard way to return environment data.
// Could be the same as Environment model itself if no transformation is needed.
// For consistency with other models, we can define it, but often it's just the model.
type EnvironmentResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Slug        string    `json:"slug"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// EnvironmentListParams defines parameters for listing environments.
type EnvironmentListParams struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=10"`
	Name     string `form:"name"` // For searching by environment name
	Slug     string `form:"slug"` // For searching by environment slug
}

// ToEnvironmentResponse converts an Environment model to an EnvironmentResponse.
func (e *Environment) ToEnvironmentResponse() EnvironmentResponse {
	return EnvironmentResponse{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		Slug:        e.Slug,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}
