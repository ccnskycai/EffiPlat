//go:generate mockgen -destination=mocks/mock_environment_repository.go -package=mocks EffiPlat/backend/internal/repository EnvironmentRepository
package repository

import (
	"EffiPlat/backend/internal/model"
	"context"
)

// EnvironmentRepository defines the interface for database operations on environments.
type EnvironmentRepository interface {
	Create(ctx context.Context, environment *model.Environment) (*model.Environment, error)
	List(ctx context.Context, params model.EnvironmentListParams) ([]model.Environment, int64, error)
	GetByID(ctx context.Context, id uint) (*model.Environment, error)
	GetBySlug(ctx context.Context, slug string) (*model.Environment, error) // Useful for checking uniqueness or fetching by slug
	Update(ctx context.Context, environment *model.Environment) (*model.Environment, error)
	Delete(ctx context.Context, id uint) error
	// TODO: Consider if a GetByName method is needed in the future.
}
