//go:generate mockgen -destination=mocks/mock_service_type_repository.go -package=mocks EffiPlat/backend/internal/repository ServiceTypeRepository
package repository

import (
	"EffiPlat/backend/internal/model"
	"context"
)

// ServiceTypeRepository defines the interface for interacting with service type data.
type ServiceTypeRepository interface {
	Create(ctx context.Context, serviceType *model.ServiceType) error
	GetByID(ctx context.Context, id uint) (*model.ServiceType, error)
	GetByName(ctx context.Context, name string) (*model.ServiceType, error)
	List(ctx context.Context, params model.ServiceTypeListParams) ([]model.ServiceType, int64, error)
	Update(ctx context.Context, serviceType *model.ServiceType) error
	Delete(ctx context.Context, id uint) error
	CheckExists(ctx context.Context, id uint) (bool, error)
}
