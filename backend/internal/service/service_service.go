package service

import (
	"context"
	"errors"
	"fmt"

	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/repository"

	"go.uber.org/zap"
)

// ServiceService defines the interface for business logic related to services and service types.
type ServiceService interface {
	// ServiceType methods
	CreateServiceType(ctx context.Context, req model.CreateServiceTypeRequest) (*model.ServiceType, error)
	GetServiceTypeByID(ctx context.Context, id uint) (*model.ServiceType, error)
	ListServiceTypes(ctx context.Context, params model.ServiceTypeListParams) ([]model.ServiceType, *model.PaginatedData, error)
	UpdateServiceType(ctx context.Context, id uint, req model.UpdateServiceTypeRequest) (*model.ServiceType, error)
	DeleteServiceType(ctx context.Context, id uint) error

	// Service methods
	CreateService(ctx context.Context, req model.CreateServiceRequest) (*model.ServiceResponse, error)
	GetServiceByID(ctx context.Context, id uint) (*model.ServiceResponse, error)
	ListServices(ctx context.Context, params model.ServiceListParams) ([]model.ServiceResponse, *model.PaginatedData, error)
	UpdateService(ctx context.Context, id uint, req model.UpdateServiceRequest) (*model.ServiceResponse, error)
	DeleteService(ctx context.Context, id uint) error
}

type serviceService struct {
	serviceRepo     repository.ServiceRepository
	serviceTypeRepo repository.ServiceTypeRepository
	logger          *zap.Logger
}

// NewServiceService creates a new instance of ServiceService.
func NewServiceService(serviceRepo repository.ServiceRepository, serviceTypeRepo repository.ServiceTypeRepository, logger *zap.Logger) ServiceService {
	return &serviceService{
		serviceRepo:     serviceRepo,
		serviceTypeRepo: serviceTypeRepo,
		logger:          logger,
	}
}

// --- ServiceType Service Methods ---

func (s *serviceService) CreateServiceType(ctx context.Context, req model.CreateServiceTypeRequest) (*model.ServiceType, error) {
	s.logger.Info("Creating service type", zap.String("name", req.Name))

	existing, err := s.serviceTypeRepo.GetByName(ctx, req.Name)
	if err != nil {
		s.logger.Error("Error checking for existing service type by name", zap.Error(err), zap.String("name", req.Name))
		return nil, err
	}
	if existing != nil {
		s.logger.Warn("Service type with this name already exists", zap.String("name", req.Name), zap.Uint("existingID", existing.ID))
		return nil, model.ErrServiceTypeNameExists
	}

	serviceType := &model.ServiceType{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.serviceTypeRepo.Create(ctx, serviceType); err != nil {
		s.logger.Error("Failed to create service type", zap.Error(err))
		return nil, err
	}
	s.logger.Info("Service type created successfully", zap.Uint("id", serviceType.ID))
	return serviceType, nil
}

func (s *serviceService) GetServiceTypeByID(ctx context.Context, id uint) (*model.ServiceType, error) {
	s.logger.Debug("Getting service type by ID", zap.Uint("id", id))
	st, err := s.serviceTypeRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrServiceTypeNotFound) {
			s.logger.Warn("Service type not found by ID", zap.Uint("id", id))
		}
		return nil, err
	}
	return st, nil
}

func (s *serviceService) ListServiceTypes(ctx context.Context, params model.ServiceTypeListParams) ([]model.ServiceType, *model.PaginatedData, error) {
	s.logger.Debug("Listing service types", zap.Any("params", params))
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	} else if params.PageSize > 100 {
		params.PageSize = 100
	}

	serviceTypes, totalCount, err := s.serviceTypeRepo.List(ctx, params)
	if err != nil {
		s.logger.Error("Failed to list service types", zap.Error(err))
		return nil, nil, err
	}

	pData := &model.PaginatedData{
		Items:    serviceTypes,
		Total:    totalCount,
		Page:     params.Page,
		PageSize: params.PageSize,
	}
	return serviceTypes, pData, nil
}

func (s *serviceService) UpdateServiceType(ctx context.Context, id uint, req model.UpdateServiceTypeRequest) (*model.ServiceType, error) {
	s.logger.Info("Updating service type", zap.Uint("id", id))

	serviceType, err := s.serviceTypeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err // Handles ErrServiceTypeNotFound from repository
	}

	updated := false
	if req.Name != nil && *req.Name != serviceType.Name {
		existing, err := s.serviceTypeRepo.GetByName(ctx, *req.Name)
		if err != nil {
			s.logger.Error("Error checking for existing service type by name during update", zap.Error(err), zap.String("name", *req.Name))
			return nil, err
		}
		if existing != nil && existing.ID != id {
			s.logger.Warn("Another service type with this name already exists", zap.String("name", *req.Name), zap.Uint("conflictingID", existing.ID))
			return nil, model.ErrServiceTypeNameExists
		}
		serviceType.Name = *req.Name
		updated = true
	}

	if req.Description != nil && *req.Description != serviceType.Description {
		serviceType.Description = *req.Description
		updated = true
	}

	if !updated {
		s.logger.Info("No changes detected for service type update", zap.Uint("id", id))
		return serviceType, nil // No fields to update
	}

	if err := s.serviceTypeRepo.Update(ctx, serviceType); err != nil {
		s.logger.Error("Failed to update service type", zap.Error(err))
		return nil, err
	}
	s.logger.Info("Service type updated successfully", zap.Uint("id", serviceType.ID))
	return serviceType, nil
}

func (s *serviceService) DeleteServiceType(ctx context.Context, id uint) error {
	s.logger.Info("Deleting service type", zap.Uint("id", id))

	// Ensure the service type exists before attempting to delete
	_, err := s.serviceTypeRepo.GetByID(ctx, id)
	if err != nil {
		return err // Handles ErrServiceTypeNotFound
	}

	// Check if the service type is in use by any services
	count, err := s.serviceRepo.CountServicesByServiceTypeID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to check if service type is in use", zap.Error(err), zap.Uint("id", id))
		return err
	}
	if count > 0 {
		s.logger.Warn("Attempt to delete service type that is in use", zap.Uint("id", id), zap.Int64("serviceCount", count))
		return model.ErrServiceTypeInUse
	}

	if err := s.serviceTypeRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete service type", zap.Error(err))
		return err
	}
	s.logger.Info("Service type deleted successfully", zap.Uint("id", id))
	return nil
}

// --- Service Service Methods ---

func (s *serviceService) CreateService(ctx context.Context, req model.CreateServiceRequest) (*model.ServiceResponse, error) {
	s.logger.Info("Creating service", zap.String("name", req.Name))

	// Validate ServiceTypeID exists
	_, err := s.serviceTypeRepo.GetByID(ctx, req.ServiceTypeID)
	if err != nil {
		if errors.Is(err, model.ErrServiceTypeNotFound) {
			s.logger.Warn("ServiceTypeID not found during service creation", zap.Uint("serviceTypeId", req.ServiceTypeID))
			return nil, fmt.Errorf("invalid service_type_id %d: %w", req.ServiceTypeID, model.ErrServiceTypeNotFound)
		}
		s.logger.Error("Failed to validate service_type_id for service creation", zap.Error(err))
		return nil, err
	}

	// Check if service with the same name already exists
	existingService, err := s.serviceRepo.GetByName(ctx, req.Name)
	if err != nil {
		s.logger.Error("Error checking for existing service by name", zap.Error(err), zap.String("name", req.Name))
		return nil, err
	}
	if existingService != nil {
		s.logger.Warn("Service with this name already exists", zap.String("name", req.Name), zap.Uint("existingID", existingService.ID))
		return nil, model.ErrServiceNameExists
	}

	service := &model.Service{
		Name:          req.Name,
		Description:   req.Description,
		Version:       req.Version,
		Status:        req.Status,
		ExternalLink:  req.ExternalLink,
		ServiceTypeID: req.ServiceTypeID,
	}
	if service.Status == "" { // Default status if not provided
		service.Status = model.ServiceStatusUnknown
	}

	if err := s.serviceRepo.Create(ctx, service); err != nil {
		s.logger.Error("Failed to create service", zap.Error(err))
		return nil, err
	}

	// Retrieve the created service with ServiceType preloaded for the response
	createdService, err := s.serviceRepo.GetByID(ctx, service.ID)
	if err != nil {
		s.logger.Error("Failed to retrieve created service with details after creation", zap.Uint("id", service.ID), zap.Error(err))
		return nil, err // If retrieval fails, it's a problem, return the error.
	}

	resp := createdService.ToServiceResponse()
	s.logger.Info("Service created successfully", zap.Uint("id", resp.ID))
	return &resp, nil
}

func (s *serviceService) GetServiceByID(ctx context.Context, id uint) (*model.ServiceResponse, error) {
	s.logger.Debug("Getting service by ID", zap.Uint("id", id))
	service, err := s.serviceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrServiceNotFound) {
			s.logger.Warn("Service not found by ID", zap.Uint("id", id))
		}
		return nil, err
	}
	resp := service.ToServiceResponse()
	return &resp, nil
}

func (s *serviceService) ListServices(ctx context.Context, params model.ServiceListParams) ([]model.ServiceResponse, *model.PaginatedData, error) {
	s.logger.Debug("Listing services", zap.Any("params", params))
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	} else if params.PageSize > 100 {
		params.PageSize = 100
	}

	services, totalCount, err := s.serviceRepo.List(ctx, params)
	if err != nil {
		s.logger.Error("Failed to list services", zap.Error(err))
		return nil, nil, err
	}

	serviceResponses := make([]model.ServiceResponse, len(services))
	for i, svc := range services {
		serviceResponses[i] = svc.ToServiceResponse()
	}

	pData := &model.PaginatedData{
		Items:    serviceResponses,
		Total:    totalCount,
		Page:     params.Page,
		PageSize: params.PageSize,
	}
	return serviceResponses, pData, nil
}

func (s *serviceService) UpdateService(ctx context.Context, id uint, req model.UpdateServiceRequest) (*model.ServiceResponse, error) {
	s.logger.Info("Updating service", zap.Uint("id", id))

	service, err := s.serviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err // Handles ErrServiceNotFound
	}

	updated := false
	// Validate ServiceTypeID if it's being changed
	if req.ServiceTypeID != nil && *req.ServiceTypeID != service.ServiceTypeID {
		_, err := s.serviceTypeRepo.GetByID(ctx, *req.ServiceTypeID)
		if err != nil {
			if errors.Is(err, model.ErrServiceTypeNotFound) {
				s.logger.Warn("New ServiceTypeID not found during service update", zap.Uint("newServiceTypeId", *req.ServiceTypeID))
				return nil, fmt.Errorf("invalid new service_type_id %d: %w", *req.ServiceTypeID, model.ErrServiceTypeNotFound)
			}
			s.logger.Error("Failed to validate new service_type_id for service update", zap.Error(err))
			return nil, err
		}
		service.ServiceTypeID = *req.ServiceTypeID
		service.ServiceType = nil // Nullify preloaded ServiceType to ensure GORM uses the ID
		updated = true
	}

	if req.Name != nil && *req.Name != service.Name {
		// Check for name conflict if name is being changed
		existingService, err := s.serviceRepo.GetByName(ctx, *req.Name)
		if err != nil {
			s.logger.Error("Error checking for existing service by name during update", zap.Error(err), zap.String("name", *req.Name))
			return nil, err
		}
		if existingService != nil && existingService.ID != id {
			s.logger.Warn("Another service with this name already exists", zap.String("name", *req.Name), zap.Uint("conflictingID", existingService.ID))
			return nil, model.ErrServiceNameExists
		}
		service.Name = *req.Name
		updated = true
	}
	if req.Description != nil && *req.Description != service.Description {
		service.Description = *req.Description
		updated = true
	}
	if req.Version != nil && *req.Version != service.Version {
		service.Version = *req.Version
		updated = true
	}
	if req.Status != nil && *req.Status != service.Status {
		service.Status = *req.Status
		updated = true
	}
	if req.ExternalLink != nil && *req.ExternalLink != service.ExternalLink {
		service.ExternalLink = *req.ExternalLink
		updated = true
	}

	if !updated {
		s.logger.Info("No changes detected for service update", zap.Uint("id", id))
		resp := service.ToServiceResponse() // Return current state if no update
		return &resp, nil
	}

	if err := s.serviceRepo.Update(ctx, service); err != nil {
		s.logger.Error("Failed to update service", zap.Error(err))
		return nil, err
	}

	// Retrieve the updated service to get the potentially updated ServiceType association correctly preloaded.
	updatedService, err := s.serviceRepo.GetByID(ctx, service.ID)
	if err != nil {
		s.logger.Error("Failed to retrieve updated service with details after update", zap.Uint("id", service.ID), zap.Error(err))
		return nil, err
	}

	resp := updatedService.ToServiceResponse()
	s.logger.Info("Service updated successfully", zap.Uint("id", resp.ID))
	return &resp, nil
}

func (s *serviceService) DeleteService(ctx context.Context, id uint) error {
	s.logger.Info("Deleting service", zap.Uint("id", id))

	// First, ensure the service exists before attempting to delete
	_, err := s.serviceRepo.GetByID(ctx, id)
	if err != nil {
		return err // Handles ErrServiceNotFound
	}

	if err := s.serviceRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete service", zap.Error(err))
		return err
	}
	s.logger.Info("Service deleted successfully", zap.Uint("id", id))
	return nil
}
