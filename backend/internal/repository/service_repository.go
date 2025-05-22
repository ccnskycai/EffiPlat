package repository

import (
	"context"
	"errors"

	"EffiPlat/backend/internal/models"
	apputils "EffiPlat/backend/internal/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ServiceRepository defines the interface for interacting with service and service_type data.
// It includes methods for both Service and ServiceType entities.
type ServiceRepository interface {
	// ServiceType methods
	CreateServiceType(ctx context.Context, serviceType *models.ServiceType) error
	GetServiceTypeByID(ctx context.Context, id uint) (*models.ServiceType, error)
	GetServiceTypeByName(ctx context.Context, name string) (*models.ServiceType, error)
	ListServiceTypes(ctx context.Context, params models.ServiceTypeListParams) ([]models.ServiceType, *models.PaginatedData, error)
	UpdateServiceType(ctx context.Context, serviceType *models.ServiceType) error
	DeleteServiceType(ctx context.Context, id uint) error

	// Service methods (to be implemented next)
	CreateService(ctx context.Context, service *models.Service) error
	GetServiceByID(ctx context.Context, id uint) (*models.Service, error)
	ListServices(ctx context.Context, params models.ServiceListParams) ([]models.Service, *models.PaginatedData, error)
	UpdateService(ctx context.Context, service *models.Service) error
	DeleteService(ctx context.Context, id uint) error
}

type serviceRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewServiceRepository creates a new instance of ServiceRepository.
func NewServiceRepository(db *gorm.DB, logger *zap.Logger) ServiceRepository {
	return &serviceRepository{db: db, logger: logger}
}

// --- ServiceType Repository Methods ---

// CreateServiceType creates a new service type in the database.
func (r *serviceRepository) CreateServiceType(ctx context.Context, serviceType *models.ServiceType) error {
	r.logger.Info("Creating service type", zap.String("name", serviceType.Name))
	if err := r.db.WithContext(ctx).Create(serviceType).Error; err != nil {
		r.logger.Error("Failed to create service type", zap.Error(err))
		return err
	}
	r.logger.Info("Service type created successfully", zap.Uint("id", serviceType.ID))
	return nil
}

// GetServiceTypeByID retrieves a service type by its ID.
func (r *serviceRepository) GetServiceTypeByID(ctx context.Context, id uint) (*models.ServiceType, error) {
	var serviceType models.ServiceType
	r.logger.Debug("Getting service type by ID", zap.Uint("id", id))
	if err := r.db.WithContext(ctx).First(&serviceType, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("Service type not found by ID", zap.Uint("id", id))
			return nil, apputils.ErrNotFound // Consider a shared error for not found
		}
		r.logger.Error("Failed to get service type by ID", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}
	return &serviceType, nil
}

// GetServiceTypeByName retrieves a service type by its name.
func (r *serviceRepository) GetServiceTypeByName(ctx context.Context, name string) (*models.ServiceType, error) {
	var serviceType models.ServiceType
	r.logger.Debug("Getting service type by name", zap.String("name", name))
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&serviceType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("Service type not found by name", zap.String("name", name))
			return nil, apputils.ErrNotFound
		}
		r.logger.Error("Failed to get service type by name", zap.Error(err), zap.String("name", name))
		return nil, err
	}
	return &serviceType, nil
}

// ListServiceTypes retrieves a paginated list of service types, optionally filtered by name.
func (r *serviceRepository) ListServiceTypes(ctx context.Context, params models.ServiceTypeListParams) ([]models.ServiceType, *models.PaginatedData, error) {
	var serviceTypes []models.ServiceType
	var totalCount int64

	r.logger.Debug("Listing service types", zap.Any("params", params))

	dbQuery := r.db.WithContext(ctx).Model(&models.ServiceType{})

	if params.Name != "" {
		dbQuery = dbQuery.Where("name LIKE ?", "%"+params.Name+"%")
	}

	// Get total count for pagination
	if err := dbQuery.Count(&totalCount).Error; err != nil {
		r.logger.Error("Failed to count service types", zap.Error(err))
		return nil, nil, err
	}

	// Apply pagination
	offset := (params.Page - 1) * params.PageSize
	dbQuery = dbQuery.Offset(offset).Limit(params.PageSize)

	// Fetch records
	if err := dbQuery.Order("name ASC").Find(&serviceTypes).Error; err != nil {
		r.logger.Error("Failed to list service types", zap.Error(err))
		return nil, nil, err
	}

	paginationDetails := &models.PaginatedData{
		Items:    serviceTypes,
		Total:    totalCount,
		Page:     params.Page,
		PageSize: params.PageSize,
	}

	r.logger.Info("Service types listed successfully", zap.Int("count", len(serviceTypes)), zap.Int64("total", totalCount))
	return serviceTypes, paginationDetails, nil
}

// UpdateServiceType updates an existing service type.
func (r *serviceRepository) UpdateServiceType(ctx context.Context, serviceType *models.ServiceType) error {
	r.logger.Info("Updating service type", zap.Uint("id", serviceType.ID))
	// Ensure ID is set for update
	if serviceType.ID == 0 {
		return apputils.ErrMissingID
	}

	// Using .Model(&models.ServiceType{}).Where("id = ?", serviceType.ID).Updates(serviceType) is safer for partial updates
	// as it only updates non-zero fields. If a full overwrite is intended, .Save(serviceType) can be used.
	// For simplicity, we use Save here, assuming the service layer prepares the full object or handles partial updates.
	if err := r.db.WithContext(ctx).Save(serviceType).Error; err != nil {
		r.logger.Error("Failed to update service type", zap.Error(err), zap.Uint("id", serviceType.ID))
		return err
	}
	r.logger.Info("Service type updated successfully", zap.Uint("id", serviceType.ID))
	return nil
}

// DeleteServiceType deletes a service type by its ID.
func (r *serviceRepository) DeleteServiceType(ctx context.Context, id uint) error {
	r.logger.Info("Deleting service type", zap.Uint("id", id))
	if err := r.db.WithContext(ctx).Delete(&models.ServiceType{}, id).Error; err != nil {
		r.logger.Error("Failed to delete service type", zap.Error(err), zap.Uint("id", id))
		return err
	}
	// Check if any rows were affected to confirm deletion, GORM returns error if record not found for Delete
	// However, if it returns no error and no rows affected, it might mean the record wasn't there to begin with.
	// For now, we assume no error means success or record didn't exist.
	r.logger.Info("Service type deleted successfully or did not exist", zap.Uint("id", id))
	return nil
}

// --- Service Repository Methods ---

// CreateService creates a new service in the database.
func (r *serviceRepository) CreateService(ctx context.Context, service *models.Service) error {
	r.logger.Info("Creating service", zap.String("name", service.Name))
	if err := r.db.WithContext(ctx).Create(service).Error; err != nil {
		r.logger.Error("Failed to create service", zap.Error(err))
		return err
	}
	r.logger.Info("Service created successfully", zap.Uint("id", service.ID))
	return nil
}

// GetServiceByID retrieves a service by its ID, preloading its ServiceType.
func (r *serviceRepository) GetServiceByID(ctx context.Context, id uint) (*models.Service, error) {
	var service models.Service
	r.logger.Debug("Getting service by ID", zap.Uint("id", id))
	// Preload ServiceType for richer data
	if err := r.db.WithContext(ctx).Preload("ServiceType").First(&service, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("Service not found by ID", zap.Uint("id", id))
			return nil, apputils.ErrNotFound
		}
		r.logger.Error("Failed to get service by ID", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}
	return &service, nil
}

// ListServices retrieves a paginated list of services, optionally filtered by parameters.
// It also preloads the ServiceType for each service.
func (r *serviceRepository) ListServices(ctx context.Context, params models.ServiceListParams) ([]models.Service, *models.PaginatedData, error) {
	var services []models.Service
	var totalCount int64

	r.logger.Debug("Listing services", zap.Any("params", params))

	dbQuery := r.db.WithContext(ctx).Model(&models.Service{}).Preload("ServiceType")

	if params.Name != "" {
		dbQuery = dbQuery.Where("services.name LIKE ?", "%"+params.Name+"%") // Qualify column name
	}
	if params.Status != "" {
		dbQuery = dbQuery.Where("services.status = ?", params.Status) // Qualify column name
	}
	if params.ServiceTypeID > 0 {
		dbQuery = dbQuery.Where("services.service_type_id = ?", params.ServiceTypeID) // Qualify column name
	}

	// Get total count for pagination
	if err := dbQuery.Count(&totalCount).Error; err != nil {
		r.logger.Error("Failed to count services", zap.Error(err))
		return nil, nil, err
	}

	// Apply pagination
	offset := (params.Page - 1) * params.PageSize
	dbQuery = dbQuery.Offset(offset).Limit(params.PageSize)

	// Define default order, e.g., by name or creation date
	// Make sure to qualify column names if there are joins or ambiguity, e.g., "services.name ASC"
	if err := dbQuery.Order("services.name ASC").Find(&services).Error; err != nil {
		r.logger.Error("Failed to list services", zap.Error(err))
		return nil, nil, err
	}

	paginationDetails := &models.PaginatedData{
		Items:    services,
		Total:    totalCount,
		Page:     params.Page,
		PageSize: params.PageSize,
	}

	r.logger.Info("Services listed successfully", zap.Int("count", len(services)), zap.Int64("total", totalCount))
	return services, paginationDetails, nil
}

// UpdateService updates an existing service.
// It's important to handle associations like ServiceType carefully if they can be changed.
// GORM's Save method might attempt to re-create associations if not handled properly.
// For updating fields of Service itself, .Updates is often safer for partial updates.
func (r *serviceRepository) UpdateService(ctx context.Context, service *models.Service) error {
	r.logger.Info("Updating service", zap.Uint("id", service.ID))
	if service.ID == 0 {
		return apputils.ErrMissingID
	}

	// If ServiceTypeID is being updated, ensure the ServiceType association is handled correctly.
	// GORM's Save will attempt to save associations. If only ID is changing, it might be fine.
	// If the ServiceType object itself is nested and changed, Save will try to update/create it.
	// For updating specific fields, including foreign keys like ServiceTypeID:
	// result := r.db.WithContext(ctx).Model(&models.Service{ID: service.ID}).Updates(models.Service{
	// Name: service.Name, Description: service.Description, Version: service.Version,
	// Status: service.Status, ExternalLink: service.ExternalLink, ServiceTypeID: service.ServiceTypeID,
	// })
	// For this implementation, we'll use Save, assuming the service layer provides the correct structure.
	if err := r.db.WithContext(ctx).Save(service).Error; err != nil {
		r.logger.Error("Failed to update service", zap.Error(err), zap.Uint("id", service.ID))
		return err
	}
	r.logger.Info("Service updated successfully", zap.Uint("id", service.ID))
	return nil
}

// DeleteService deletes a service by its ID (soft delete due to gorm.DeletedAt in model).
func (r *serviceRepository) DeleteService(ctx context.Context, id uint) error {
	r.logger.Info("Deleting service", zap.Uint("id", id))
	// GORM will perform a soft delete because Service model has gorm.DeletedAt field
	if err := r.db.WithContext(ctx).Delete(&models.Service{}, id).Error; err != nil {
		r.logger.Error("Failed to delete service", zap.Error(err), zap.Uint("id", id))
		return err
	}
	r.logger.Info("Service marked as deleted successfully or did not exist", zap.Uint("id", id))
	return nil
}
