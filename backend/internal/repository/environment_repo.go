package repository

import (
	"EffiPlat/backend/internal/models"
	"context"
)

// EnvironmentRepository defines the interface for database operations on environments.
type EnvironmentRepository interface {
	Create(ctx context.Context, environment *models.Environment) (*models.Environment, error)
	List(ctx context.Context, params models.EnvironmentListParams) ([]models.Environment, int64, error)
	GetByID(ctx context.Context, id uint) (*models.Environment, error)
	GetBySlug(ctx context.Context, slug string) (*models.Environment, error) // Useful for checking uniqueness or fetching by slug
	Update(ctx context.Context, environment *models.Environment) (*models.Environment, error)
	Delete(ctx context.Context, id uint) error
	// TODO: Consider if a GetByName method is needed in the future.
}
