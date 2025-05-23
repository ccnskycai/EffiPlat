package repository

import (
	"EffiPlat/backend/internal/models"
	"context"
)

// ServiceTypeRepository defines the interface for interacting with service type data.
type ServiceTypeRepository interface {
	Create(ctx context.Context, serviceType *models.ServiceType) error
	GetByID(ctx context.Context, id uint) (*models.ServiceType, error)
	GetByName(ctx context.Context, name string) (*models.ServiceType, error)
	List(ctx context.Context, params models.ServiceTypeListParams) ([]models.ServiceType, int64, error)
	Update(ctx context.Context, serviceType *models.ServiceType) error
	Delete(ctx context.Context, id uint) error
	CheckExists(ctx context.Context, id uint) (bool, error)
}
