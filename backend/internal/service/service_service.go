package service

import (
	"context"
	"errors" // Required for custom error checking if any, or remove if not used directly
	"fmt"    // Added for fmt.Errorf

	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/repository"
	"EffiPlat/backend/internal/utils" // For ErrNotFound

	"go.uber.org/zap"
	// "gorm.io/gorm" // Not directly used in service layer, repository handles db interaction
)

// ServiceService defines the interface for business logic related to services and service types.
// It orchestrates operations using the ServiceRepository and can include additional business rules,
// validations, and transformations between request/response DTOs and domain models.
type ServiceService interface {
	// ServiceType methods
	CreateServiceType(ctx context.Context, req models.CreateServiceTypeRequest) (*models.ServiceType, error)
	GetServiceTypeByID(ctx context.Context, id uint) (*models.ServiceType, error)
	GetServiceTypeByName(ctx context.Context, name string) (*models.ServiceType, error)
	ListServiceTypes(ctx context.Context, params models.ServiceTypeListParams) ([]models.ServiceType, *models.PaginatedData, error)
	UpdateServiceType(ctx context.Context, id uint, req models.UpdateServiceTypeRequest) (*models.ServiceType, error)
	DeleteServiceType(ctx context.Context, id uint) error

	// Service methods
	CreateService(ctx context.Context, req models.CreateServiceRequest) (*models.ServiceResponse, error)
	GetServiceByID(ctx context.Context, id uint) (*models.ServiceResponse, error)
	ListServices(ctx context.Context, params models.ServiceListParams) ([]models.ServiceResponse, *models.PaginatedData, error)
	UpdateService(ctx context.Context, id uint, req models.UpdateServiceRequest) (*models.ServiceResponse, error)
	DeleteService(ctx context.Context, id uint) error
}

type serviceService struct {
	repo   repository.ServiceRepository
	logger *zap.Logger
}

// NewServiceService creates a new instance of ServiceService.
func NewServiceService(repo repository.ServiceRepository, logger *zap.Logger) ServiceService {
	return &serviceService{repo: repo, logger: logger}
}

// --- ServiceType Service Methods ---

// CreateServiceType handles the business logic for creating a new service type.
func (s *serviceService) CreateServiceType(ctx context.Context, req models.CreateServiceTypeRequest) (*models.ServiceType, error) {
	s.logger.Info("Service layer: Creating service type", zap.String("name", req.Name))

	// Optional: Add any specific business validation here before creating
	// For example, check for naming conventions, restricted names, etc.

	// Check if a service type with the same name already exists
	existing, err := s.repo.GetServiceTypeByName(ctx, req.Name)
	if err != nil && !errors.Is(err, utils.ErrNotFound) {
		s.logger.Error("Error checking for existing service type by name", zap.Error(err), zap.String("name", req.Name))
		return nil, err // Propagate repository error
	}
	if existing != nil {
		s.logger.Warn("Service type with this name already exists", zap.String("name", req.Name), zap.Uint("existingID", existing.ID))
		return nil, fmt.Errorf("service type with name '%s' already exists: %w", req.Name, utils.ErrAlreadyExists)
	}

	serviceType := &models.ServiceType{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.repo.CreateServiceType(ctx, serviceType); err != nil {
		// Error already logged by repository
		return nil, err
	}
	s.logger.Info("Service layer: Service type created successfully", zap.Uint("id", serviceType.ID))
	return serviceType, nil
}

// GetServiceTypeByID retrieves a service type by ID.
func (s *serviceService) GetServiceTypeByID(ctx context.Context, id uint) (*models.ServiceType, error) {
	s.logger.Debug("Service layer: Getting service type by ID", zap.Uint("id", id))
	return s.repo.GetServiceTypeByID(ctx, id)
}

// GetServiceTypeByName retrieves a service type by name.
func (s *serviceService) GetServiceTypeByName(ctx context.Context, name string) (*models.ServiceType, error) {
	s.logger.Debug("Service layer: Getting service type by name", zap.String("name", name))
	return s.repo.GetServiceTypeByName(ctx, name)
}

// ListServiceTypes retrieves a list of service types.
func (s *serviceService) ListServiceTypes(ctx context.Context, params models.ServiceTypeListParams) ([]models.ServiceType, *models.PaginatedData, error) {
	s.logger.Debug("Service layer: Listing service types", zap.Any("params", params))
	// Basic validation for pagination params
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 100 { // Max page size cap
		params.PageSize = 10
	}
	return s.repo.ListServiceTypes(ctx, params)
}

// UpdateServiceType handles the business logic for updating a service type.
func (s *serviceService) UpdateServiceType(ctx context.Context, id uint, req models.UpdateServiceTypeRequest) (*models.ServiceType, error) {
	s.logger.Info("Service layer: Updating service type", zap.Uint("id", id))

	serviceType, err := s.repo.GetServiceTypeByID(ctx, id)
	if err != nil {
		// Error already logged by repository or utils.ErrNotFound
		return nil, err
	}

	// Check for name conflict if name is being changed
	if req.Name != nil && *req.Name != serviceType.Name {
		existing, err := s.repo.GetServiceTypeByName(ctx, *req.Name)
		if err != nil && !errors.Is(err, utils.ErrNotFound) {
			s.logger.Error("Error checking for existing service type by name during update", zap.Error(err), zap.String("newName", *req.Name))
			return nil, err
		}
		if existing != nil && existing.ID != id {
			s.logger.Warn("Another service type with this name already exists", zap.String("newName", *req.Name), zap.Uint("conflictingID", existing.ID))
			return nil, fmt.Errorf("another service type with name '%s' already exists: %w", *req.Name, utils.ErrAlreadyExists)
		}
		serviceType.Name = *req.Name
	}

	if req.Description != nil {
		serviceType.Description = *req.Description
	}

	if err := s.repo.UpdateServiceType(ctx, serviceType); err != nil {
		return nil, err
	}
	s.logger.Info("Service layer: Service type updated successfully", zap.Uint("id", serviceType.ID))
	return serviceType, nil
}

// DeleteServiceType handles the business logic for deleting a service type.
func (s *serviceService) DeleteServiceType(ctx context.Context, id uint) error {
	s.logger.Info("Service layer: Deleting service type", zap.Uint("id", id))

	// Optional: Add business logic here, e.g., check if any Service entities are still using this ServiceType.
	// If the foreign key in the `services` table has `ON DELETE RESTRICT`, the database will prevent this.
	// However, providing a user-friendly error from the service layer is better.
	// Example check (requires a method in repository like CountServicesByServiceTypeID):
	// count, err := s.repo.CountServicesByServiceTypeID(ctx, id)
	// if err != nil { return err }
	// if count > 0 { return errors.New("cannot delete service type as it is currently in use by services") }

	return s.repo.DeleteServiceType(ctx, id)
}

// --- Service Service Methods ---

// CreateService handles the business logic for creating a new service.
func (s *serviceService) CreateService(ctx context.Context, req models.CreateServiceRequest) (*models.ServiceResponse, error) {
	s.logger.Info("Service layer: Creating service", zap.String("name", req.Name))

	// Validate ServiceTypeID exists
	_, err := s.repo.GetServiceTypeByID(ctx, req.ServiceTypeID)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			s.logger.Warn("ServiceTypeID not found during service creation", zap.Uint("serviceTypeId", req.ServiceTypeID))
			return nil, fmt.Errorf("invalid service_type_id %d: %w", req.ServiceTypeID, utils.ErrNotFound)
		}
		s.logger.Error("Failed to validate service_type_id", zap.Error(err))
		return nil, err
	}

	// Optional: Check for existing service with the same name (if names should be unique)
	// Implement GetServiceByName in repository if needed for this check.

	service := &models.Service{
		Name:          req.Name,
		Description:   req.Description,
		Version:       req.Version,
		Status:        req.Status,
		ExternalLink:  req.ExternalLink,
		ServiceTypeID: req.ServiceTypeID,
	}
	// If status is not provided in request, GORM default 'unknown' will be used.
	if req.Status == "" {
		service.Status = models.ServiceStatusUnknown // Explicitly set if not provided
	}

	if err := s.repo.CreateService(ctx, service); err != nil {
		return nil, err
	}

	// Retrieve the service again to get it with preloaded ServiceType for the response
	createdService, err := s.repo.GetServiceByID(ctx, service.ID)
	if err != nil {
		s.logger.Error("Failed to retrieve created service with details", zap.Uint("id", service.ID), zap.Error(err))
		// Even if retrieval fails, the service was created. Decide on error handling.
		// For now, return the error, but client might get a confusing state.
		// Alternatively, return a simpler response or the initial service object (without preload).
		return nil, err
	}

	resp := createdService.ToServiceResponse()
	s.logger.Info("Service layer: Service created successfully", zap.Uint("id", resp.ID))
	return &resp, nil
}

// GetServiceByID retrieves a service by its ID and converts it to a ServiceResponse.
func (s *serviceService) GetServiceByID(ctx context.Context, id uint) (*models.ServiceResponse, error) {
	s.logger.Debug("Service layer: Getting service by ID", zap.Uint("id", id))
	service, err := s.repo.GetServiceByID(ctx, id) // Repository already preloads ServiceType
	if err != nil {
		return nil, err
	}
	resp := service.ToServiceResponse()
	return &resp, nil
}

// ListServices retrieves a list of services, converting them to ServiceResponse.
func (s *serviceService) ListServices(ctx context.Context, params models.ServiceListParams) ([]models.ServiceResponse, *models.PaginatedData, error) {
	s.logger.Debug("Service layer: Listing services", zap.Any("params", params))
	// Basic validation for pagination params
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 100 {
		params.PageSize = 10
	}

	services, paginatedData, err := s.repo.ListServices(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	serviceResponses := make([]models.ServiceResponse, len(services))
	for i, service := range services {
		serviceResponses[i] = service.ToServiceResponse()
	}

	// The Items in PaginatedData should be the list of ServiceResponses
	paginatedData.Items = serviceResponses

	return serviceResponses, paginatedData, nil
}

// UpdateService handles the business logic for updating an existing service.
func (s *serviceService) UpdateService(ctx context.Context, id uint, req models.UpdateServiceRequest) (*models.ServiceResponse, error) {
	s.logger.Info("Service layer: Updating service", zap.Uint("id", id))

	service, err := s.repo.GetServiceByID(ctx, id) // Gets service with preloaded ServiceType
	if err != nil {
		return nil, err
	}

	// Validate ServiceTypeID if it's being changed
	if req.ServiceTypeID != nil {
		// Validate new ServiceTypeID exists
		_, err := s.repo.GetServiceTypeByID(ctx, *req.ServiceTypeID)
		if err != nil {
			if errors.Is(err, utils.ErrNotFound) {
				s.logger.Warn("New ServiceTypeID not found during service update", zap.Uint("newServiceTypeId", *req.ServiceTypeID))
				return nil, fmt.Errorf("invalid new service_type_id %d: %w", *req.ServiceTypeID, utils.ErrNotFound)
			}
			s.logger.Error("Failed to validate new service_type_id for service update", zap.Error(err))
			return nil, err
		}
		service.ServiceTypeID = *req.ServiceTypeID
		service.ServiceType = nil // Important: Nullify the loaded ServiceType so GORM re-fetches or uses ID correctly on Save.
		// Otherwise, GORM might try to update the existing preloaded ServiceType based on the new ID,
		// or create a new one if service.ServiceType struct fields were also changed, which is not intended here.
	}

	if req.Name != nil {
		service.Name = *req.Name
	}
	if req.Description != nil {
		service.Description = *req.Description
	}
	if req.Version != nil {
		service.Version = *req.Version
	}
	if req.Status != nil {
		service.Status = *req.Status
	}
	if req.ExternalLink != nil {
		service.ExternalLink = *req.ExternalLink
	}

	if err := s.repo.UpdateService(ctx, service); err != nil {
		return nil, err
	}

	// Retrieve the updated service to get the potentially updated ServiceType association correctly preloaded.
	updatedService, err := s.repo.GetServiceByID(ctx, service.ID)
	if err != nil {
		s.logger.Error("Failed to retrieve updated service with details", zap.Uint("id", service.ID), zap.Error(err))
		return nil, err
	}

	resp := updatedService.ToServiceResponse()
	s.logger.Info("Service layer: Service updated successfully", zap.Uint("id", resp.ID))
	return &resp, nil
}

// DeleteService handles the business logic for deleting a service.
func (s *serviceService) DeleteService(ctx context.Context, id uint) error {
	s.logger.Info("Service layer: Deleting service", zap.Uint("id", id))

	// Optional: Add business logic, e.g., check for dependent ServiceInstances before deleting a Service.
	// This depends on how strict the deletion policy should be and if cascading deletes are handled elsewhere.

	// First, ensure the service exists before attempting to delete
	_, err := s.repo.GetServiceByID(ctx, id)
	if err != nil {
		// This will return utils.ErrNotFound if it doesn't exist, which is appropriate.
		return err
	}

	return s.repo.DeleteService(ctx, id)
}
