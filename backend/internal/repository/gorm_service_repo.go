package repository

import (
	"EffiPlat/backend/internal/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

// gormServiceRepository implements the ServiceRepository interface using GORM.
type gormServiceRepository struct {
	db *gorm.DB
}

// NewGormServiceRepository creates a new GormServiceRepository.
func NewGormServiceRepository(db *gorm.DB) ServiceRepository {
	return &gormServiceRepository{db: db}
}

// Create creates a new service.
func (r *gormServiceRepository) Create(ctx context.Context, service *models.Service) error {
	return r.db.WithContext(ctx).Create(service).Error
}

// GetByID retrieves a service by its ID, preloading its ServiceType.
func (r *gormServiceRepository) GetByID(ctx context.Context, id uint) (*models.Service, error) {
	var service models.Service
	err := r.db.WithContext(ctx).Preload("ServiceType").First(&service, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, models.ErrServiceNotFound // Assuming ErrServiceNotFound is defined in models
		}
		return nil, err
	}
	return &service, nil
}

// GetByName retrieves a service by its name.
// This is typically used for uniqueness checks, so not preloading ServiceType.
func (r *gormServiceRepository) GetByName(ctx context.Context, name string) (*models.Service, error) {
	var service models.Service
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&service).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found is not an error for this specific check
		}
		return nil, err
	}
	return &service, nil
}

// List retrieves a list of services with pagination and filtering, preloading ServiceType.
func (r *gormServiceRepository) List(ctx context.Context, params models.ServiceListParams) ([]models.Service, int64, error) {
	var services []models.Service
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&models.Service{}).Preload("ServiceType")

	if params.Name != "" {
		query = query.Where("services.name LIKE ?", "%"+params.Name+"%") // Disambiguate 'name' if ServiceType also has it
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.ServiceTypeID > 0 {
		query = query.Where("service_type_id = ?", params.ServiceTypeID)
	}

	// Get total count for pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (params.Page - 1) * params.PageSize
	query = query.Offset(offset).Limit(params.PageSize)

	// Apply ordering
	if params.OrderBy != "" && params.SortOrder != "" {
		// Ensure column name is safe to prevent SQL injection if not using enum validation strictly
		// For now, assuming params.OrderBy is validated by binding
		// Need to handle potential ambiguity if OrderBy refers to a field in ServiceType
		// For simplicity, assume OrderBy fields are unique to 'services' table or qualified if necessary
		orderClause := params.OrderBy + " " + params.SortOrder
		// If ordering by 'serviceTypeId', it's clear. If by 'name', 'status', etc., it's also from 'services'.
		// If a joined table field was allowed for sorting, qualification (e.g., "service_types.name") would be needed.
		query = query.Order(orderClause)
	} else {
		// Default order if nothing specified, e.g., by primary key or creation date
		query = query.Order("services.created_at DESC") // Default to services.created_at to be explicit
	}

	if err := query.Find(&services).Error; err != nil {
		return nil, 0, err
	}

	return services, totalCount, nil
}

// Update updates an existing service.
func (r *gormServiceRepository) Update(ctx context.Context, service *models.Service) error {
	// Ensure ServiceType is not inadvertently cleared if not provided or only ID is set
	// GORM's Save behavior with associations can be tricky. 
	// If service.ServiceType is nil but ServiceTypeID is set, GORM might nullify the association.
	// It's often safer to update specific fields or handle associations carefully.
	// For simplicity here, we assume the service object is correctly populated by the service layer.
	return r.db.WithContext(ctx).Save(service).Error
}

// Delete removes a service by its ID (soft delete if GORM model is configured).
func (r *gormServiceRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Service{}, id).Error
}

// CountServicesByServiceTypeID counts the number of services associated with a given serviceTypeID.
func (r *gormServiceRepository) CountServicesByServiceTypeID(ctx context.Context, serviceTypeID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Service{}).Where("service_type_id = ?", serviceTypeID).Count(&count).Error
	return count, err
}
