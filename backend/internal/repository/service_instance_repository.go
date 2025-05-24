//go:generate mockgen -destination=mocks/mock_service_instance_repository.go -package=mocks EffiPlat/backend/internal/repository ServiceInstanceRepository
package repository

import (
	"context"
	"fmt"

	"EffiPlat/backend/internal/model"
	// "EffiPlat/backend/internal/utils/dbutil" // Removed this incorrect import

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ListServiceInstancesParams defines parameters for listing service instances.
type ListServiceInstancesParams struct {
	// PaginationParams (fields directly included instead of embedding)
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=10"`
	SortBy   string `form:"sortBy,default=createdAt"` // Field to sort by
	Order    string `form:"order,default=desc"`       // Sort order (asc, desc)

	// Filter Params
	ServiceID     *uint   `form:"serviceId"`
	EnvironmentID *uint   `form:"environmentId"`
	Status        *string `form:"status"`
	Hostname      *string `form:"hostname"`
	Version       *string `form:"version"`
}

// ServiceInstanceRepository defines the interface for service instance data operations.
type ServiceInstanceRepository interface {
	Create(ctx context.Context, instance *model.ServiceInstance) error
	GetByID(ctx context.Context, id uint) (*model.ServiceInstance, error)
	List(ctx context.Context, params *ListServiceInstancesParams) ([]*model.ServiceInstance, int64, error)
	Update(ctx context.Context, instance *model.ServiceInstance) error
	Delete(ctx context.Context, id uint) error
	// CheckExists checks if a service instance with the given serviceId, environmentId, and version already exists.
	CheckExists(ctx context.Context, serviceID, environmentID uint, version string, excludeID uint) (bool, error)
}

// serviceInstanceRepositoryImpl implements ServiceInstanceRepository.
type serviceInstanceRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewServiceInstanceRepository creates a new ServiceInstanceRepository.
func NewServiceInstanceRepository(db *gorm.DB, logger *zap.Logger) ServiceInstanceRepository {
	return &serviceInstanceRepositoryImpl{db: db, logger: logger}
}

// Create creates a new service instance record.
func (r *serviceInstanceRepositoryImpl) Create(ctx context.Context, instance *model.ServiceInstance) error {
	r.logger.Debug("Creating service instance", zap.Any("instance", instance))
	if err := r.db.WithContext(ctx).Create(instance).Error; err != nil {
		r.logger.Error("Failed to create service instance", zap.Error(err))
		return fmt.Errorf("repository.Create: %w", err)
	}
	return nil
}

// GetByID retrieves a service instance by its ID.
func (r *serviceInstanceRepositoryImpl) GetByID(ctx context.Context, id uint) (*model.ServiceInstance, error) {
	var instance model.ServiceInstance
	r.logger.Debug("Getting service instance by ID", zap.Uint("id", id))
	if err := r.db.WithContext(ctx).First(&instance, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Warn("Service instance not found by ID", zap.Uint("id", id))
			return nil, err // Propagate gorm.ErrRecordNotFound
		}
		r.logger.Error("Failed to get service instance by ID", zap.Uint("id", id), zap.Error(err))
		return nil, fmt.Errorf("repository.GetByID: %w", err)
	}
	return &instance, nil
}

// List retrieves a list of service instances based on parameters.
func (r *serviceInstanceRepositoryImpl) List(ctx context.Context, params *ListServiceInstancesParams) ([]*model.ServiceInstance, int64, error) {
	var instances []*model.ServiceInstance
	var total int64

	r.logger.Debug("Listing service instances", zap.Any("params", params))

	tx := r.db.WithContext(ctx).Model(&model.ServiceInstance{})

	if params.ServiceID != nil {
		tx = tx.Where("service_id = ?", *params.ServiceID)
	}
	if params.EnvironmentID != nil {
		tx = tx.Where("environment_id = ?", *params.EnvironmentID)
	}
	if params.Status != nil {
		if status := model.ServiceInstanceStatusType(*params.Status); status.IsValid() {
			tx = tx.Where("status = ?", status)
		} else {
			r.logger.Warn("Invalid status filter provided for listing service instances", zap.String("status", *params.Status))
			// Optionally, return an error or simply don't apply this filter
		}
	}
	if params.Hostname != nil && *params.Hostname != "" {
		tx = tx.Where("hostname LIKE ?", "%"+*params.Hostname+"%")
	}
	if params.Version != nil && *params.Version != "" {
		tx = tx.Where("version = ?", *params.Version)
	}

	// Count total records before pagination
	if err := tx.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count service instances", zap.Error(err))
		return nil, 0, fmt.Errorf("repository.List.Count: %w", err)
	}

	// Handle pagination defaults if not provided or invalid
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10 // Default page size
	}
	offset := (params.Page - 1) * params.PageSize
	tx = tx.Limit(params.PageSize).Offset(offset)

	// Apply sorting
	sortByField := "created_at" // Default sort field
	allowedSorts := map[string]string{
		"id":            "id",
		"serviceId":     "service_id",
		"environmentId": "environment_id",
		"version":       "version",
		"status":        "status",
		"hostname":      "hostname",
		"deployedAt":    "deployed_at",
		"createdAt":     "created_at",
		"updatedAt":     "updated_at",
	}
	if col, ok := allowedSorts[params.SortBy]; ok {
		sortByField = col
	} else if params.SortBy != "" {
		r.logger.Warn("Invalid sortBy parameter", zap.String("sortBy", params.SortBy))
	}

	sortOrder := "DESC" // Default sort order
	if params.Order == "asc" || params.Order == "ASC" {
		sortOrder = "ASC"
	} else if params.Order != "" && !(params.Order == "desc" || params.Order == "DESC") {
		r.logger.Warn("Invalid order parameter, defaulting to DESC", zap.String("order", params.Order))
	}

	tx = tx.Order(fmt.Sprintf("%s %s", sortByField, sortOrder))

	if err := tx.Find(&instances).Error; err != nil {
		r.logger.Error("Failed to list service instances", zap.Error(err))
		return nil, 0, fmt.Errorf("repository.List.Find: %w", err)
	}

	return instances, total, nil
}

// Update updates an existing service instance.
func (r *serviceInstanceRepositoryImpl) Update(ctx context.Context, instance *model.ServiceInstance) error {
	r.logger.Debug("Updating service instance", zap.Uint("id", instance.ID), zap.Any("instance", instance))
	if instance.ID == 0 {
		return fmt.Errorf("repository.Update: instance ID is required")
	}

	// For testability, use Session to disable GORM's automatic transaction creation
	db := r.db.Session(&gorm.Session{SkipDefaultTransaction: true})
	result := db.WithContext(ctx).Model(instance).Updates(instance) // .Model(instance) will use instance.ID for WHERE clause
	if result.Error != nil {
		r.logger.Error("Failed to update service instance", zap.Uint("id", instance.ID), zap.Error(result.Error))
		return fmt.Errorf("repository.Update: %w", result.Error)
	}

	// If no rows affected, it could be:
	// 1. Record doesn't exist
	// 2. No fields were changed
	// In either case, let's consider it as "not found" to maintain the original behavior
	if result.RowsAffected == 0 {
		r.logger.Warn("Service instance update resulted in no rows affected, record not found or no changes", zap.Uint("id", instance.ID))
		return gorm.ErrRecordNotFound
	}

	return nil
}

// Delete removes a service instance by ID (soft delete if gorm.DeletedAt is used).
func (r *serviceInstanceRepositoryImpl) Delete(ctx context.Context, id uint) error {
	r.logger.Debug("Deleting service instance by ID", zap.Uint("id", id))
	result := r.db.WithContext(ctx).Delete(&model.ServiceInstance{}, id)
	if result.Error != nil {
		r.logger.Error("Failed to delete service instance", zap.Uint("id", id), zap.Error(result.Error))
		return fmt.Errorf("repository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		r.logger.Warn("No service instance found to delete by ID", zap.Uint("id", id))
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CheckExists checks if a service instance with the given serviceId, environmentId, and version already exists.
// excludeID is used during updates to exclude the current instance being updated.
func (r *serviceInstanceRepositoryImpl) CheckExists(ctx context.Context, serviceID, environmentID uint, version string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.ServiceInstance{}).
		Where("service_id = ? AND environment_id = ? AND version = ?", serviceID, environmentID, version)

	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	r.logger.Debug("Checking if service instance exists",
		zap.Uint("serviceID", serviceID),
		zap.Uint("environmentID", environmentID),
		zap.String("version", version),
		zap.Uint("excludeID", excludeID),
	)

	if err := query.Count(&count).Error; err != nil {
		r.logger.Error("Failed to check service instance existence",
			zap.Uint("serviceID", serviceID),
			zap.Uint("environmentID", environmentID),
			zap.String("version", version),
			zap.Uint("excludeID", excludeID),
			zap.Error(err),
		)
		return false, fmt.Errorf("repository.CheckExists: %w", err)
	}
	return count > 0, nil
}
