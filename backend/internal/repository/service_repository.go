//go:generate mockgen -destination=mocks/mock_service_repository.go -package=mocks EffiPlat/backend/internal/repository ServiceRepository
package repository

import (
	"EffiPlat/backend/internal/model"
	"context"
)

// ServiceRepository defines the interface for interacting with service data.
type ServiceRepository interface {
	Create(ctx context.Context, service *model.Service) error
	GetByID(ctx context.Context, id uint) (*model.Service, error)
	GetByName(ctx context.Context, name string) (*model.Service, error)
	List(ctx context.Context, params model.ServiceListParams) ([]model.Service, int64, error)
	Update(ctx context.Context, service *model.Service) error
	Delete(ctx context.Context, id uint) error
	// CheckServiceTypeExists is useful for validating ServiceTypeID when creating/updating services.
	// It might be better placed in ServiceTypeRepository and called by the service layer.
	// For now, let's assume the service layer handles this validation using ServiceTypeRepository.
	CountServicesByServiceTypeID(ctx context.Context, serviceTypeID uint) (int64, error)
}
