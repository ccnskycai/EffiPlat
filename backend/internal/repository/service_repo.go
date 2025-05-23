package repository

import (
	"EffiPlat/backend/internal/models"
	"context"
)

// ServiceRepository defines the interface for interacting with service data.
type ServiceRepository interface {
	Create(ctx context.Context, service *models.Service) error
	GetByID(ctx context.Context, id uint) (*models.Service, error)
	GetByName(ctx context.Context, name string) (*models.Service, error)
	List(ctx context.Context, params models.ServiceListParams) ([]models.Service, int64, error)
	Update(ctx context.Context, service *models.Service) error
	Delete(ctx context.Context, id uint) error
	// CheckServiceTypeExists is useful for validating ServiceTypeID when creating/updating services.
	// It might be better placed in ServiceTypeRepository and called by the service layer.
	// For now, let's assume the service layer handles this validation using ServiceTypeRepository.
	CountServicesByServiceTypeID(ctx context.Context, serviceTypeID uint) (int64, error)
}
