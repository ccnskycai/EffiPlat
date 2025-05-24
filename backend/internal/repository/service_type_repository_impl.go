package repository

import (
	"EffiPlat/backend/internal/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

// gormServiceTypeRepository implements the ServiceTypeRepository interface using GORM.
type gormServiceTypeRepository struct {
	db *gorm.DB
}

// NewGormServiceTypeRepository creates a new GormServiceTypeRepository.
func NewGormServiceTypeRepository(db *gorm.DB) ServiceTypeRepository {
	return &gormServiceTypeRepository{db: db}
}

// Create creates a new service type.
func (r *gormServiceTypeRepository) Create(ctx context.Context, serviceType *model.ServiceType) error {
	return r.db.WithContext(ctx).Create(serviceType).Error
}

// GetByID retrieves a service type by its ID.
func (r *gormServiceTypeRepository) GetByID(ctx context.Context, id uint) (*model.ServiceType, error) {
	var serviceType model.ServiceType
	err := r.db.WithContext(ctx).First(&serviceType, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.ErrServiceTypeNotFound
		}
		return nil, err
	}
	return &serviceType, nil
}

// GetByName retrieves a service type by its name.
func (r *gormServiceTypeRepository) GetByName(ctx context.Context, name string) (*model.ServiceType, error) {
	var serviceType model.ServiceType
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&serviceType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found is not an error for this specific check, service layer will handle logic
		}
		return nil, err
	}
	return &serviceType, nil
}

// List retrieves a list of service types with pagination and filtering.
func (r *gormServiceTypeRepository) List(ctx context.Context, params model.ServiceTypeListParams) ([]model.ServiceType, int64, error) {
	var serviceTypes []model.ServiceType
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&model.ServiceType{})

	if params.Name != "" {
		query = query.Where("name LIKE ?", "%"+params.Name+"%")
	}

	// Get total count for pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply ordering
	if params.OrderBy != "" && params.SortOrder != "" {
		// Ensure column name is safe to prevent SQL injection if not using enum validation strictly
		// For now, assuming params.OrderBy is validated by binding
		orderClause := params.OrderBy + " " + params.SortOrder
		query = query.Order(orderClause)
	}

	// Apply pagination
	offset := (params.Page - 1) * params.PageSize
	query = query.Offset(offset).Limit(params.PageSize)

	if err := query.Find(&serviceTypes).Error; err != nil {
		return nil, 0, err
	}

	return serviceTypes, totalCount, nil
}

// Update updates an existing service type.
func (r *gormServiceTypeRepository) Update(ctx context.Context, serviceType *model.ServiceType) error {
	return r.db.WithContext(ctx).Save(serviceType).Error
}

// Delete removes a service type by its ID.
func (r *gormServiceTypeRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.ServiceType{}, id).Error
}

// CheckExists checks if a service type with the given ID exists.
func (r *gormServiceTypeRepository) CheckExists(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ServiceType{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
