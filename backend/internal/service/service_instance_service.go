package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/repository"
	apputils "EffiPlat/backend/internal/utils"

	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ServiceInstanceInputDTO is a common structure for creating and updating service instances.
// It helps in bundling request data and applying validation.
// We will map this to model.ServiceInstance before repository calls.
type ServiceInstanceInputDTO struct {
	ServiceID     uint              `json:"serviceId" binding:"required"`
	EnvironmentID uint              `json:"environmentId" binding:"required"`
	Version       string            `json:"version" binding:"required,min=1,max=100"`
	Status        string            `json:"status" binding:"required,oneof=running stopped deploying error unknown"`
	Hostname      *string           `json:"hostname" binding:"omitempty,max=255"`
	Port          *int              `json:"port" binding:"omitempty,min=1,max=65535"`
	Config        datatypes.JSONMap `json:"config"` // No specific binding here, handled as raw JSON
	DeployedAt    *time.Time        `json:"deployedAt"`
}

// ServiceInstanceOutputDTO is used for presenting service instance data to the client.
// It mirrors the model.ServiceInstance but can be adjusted for API responses.
type ServiceInstanceOutputDTO struct {
	ID            uint              `json:"id"`
	ServiceID     uint              `json:"serviceId"`
	EnvironmentID uint              `json:"environmentId"`
	Version       string            `json:"version"`
	Status        string            `json:"status"`
	Hostname      *string           `json:"hostname,omitempty"`
	Port          *int              `json:"port,omitempty"`
	Config        datatypes.JSONMap `json:"config,omitempty"`
	DeployedAt    *time.Time        `json:"deployedAt,omitempty"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
}

// ListServiceInstancesResponseDTO wraps the paginated list of service instances.
type ListServiceInstancesResponseDTO struct {
	Items []*ServiceInstanceOutputDTO `json:"items"`
	Total int64                       `json:"total"`
	Page  int                         `json:"page"`
	Size  int                         `json:"pageSize"`
}

// ServiceInstanceService defines the interface for service instance business logic.
//
//go:generate mockgen -source=service_instance_service.go -destination=mocks/mock_service_instance_service.go -package=mocks ServiceInstanceService
type ServiceInstanceService interface {
	CreateServiceInstance(ctx context.Context, input *ServiceInstanceInputDTO) (*ServiceInstanceOutputDTO, error)
	GetServiceInstanceByID(ctx context.Context, id uint) (*ServiceInstanceOutputDTO, error)
	ListServiceInstances(ctx context.Context, params *repository.ListServiceInstancesParams) (*ListServiceInstancesResponseDTO, error)
	UpdateServiceInstance(ctx context.Context, id uint, input *ServiceInstanceInputDTO) (*ServiceInstanceOutputDTO, error)
	DeleteServiceInstance(ctx context.Context, id uint) error
}

// serviceInstanceServiceImpl implements ServiceInstanceService.
type serviceInstanceServiceImpl struct {
	repo        repository.ServiceInstanceRepository
	serviceRepo repository.ServiceRepository     // For validating ServiceID
	envRepo     repository.EnvironmentRepository // For validating EnvironmentID
	logger      *zap.Logger
}

// NewServiceInstanceService creates a new ServiceInstanceService.
func NewServiceInstanceService(
	repo repository.ServiceInstanceRepository,
	serviceRepo repository.ServiceRepository,
	envRepo repository.EnvironmentRepository,
	logger *zap.Logger,
) ServiceInstanceService {
	return &serviceInstanceServiceImpl{
		repo:        repo,
		serviceRepo: serviceRepo,
		envRepo:     envRepo,
		logger:      logger,
	}
}

func convertModelToOutputDTO(instance *model.ServiceInstance) *ServiceInstanceOutputDTO {
	if instance == nil {
		return nil
	}
	return &ServiceInstanceOutputDTO{
		ID:            instance.ID,
		ServiceID:     instance.ServiceID,
		EnvironmentID: instance.EnvironmentID,
		Version:       instance.Version,
		Status:        string(instance.Status),
		Hostname:      instance.Hostname,
		Port:          instance.Port,
		Config:        instance.Config,
		DeployedAt:    instance.DeployedAt,
		CreatedAt:     instance.CreatedAt,
		UpdatedAt:     instance.UpdatedAt,
	}
}

func convertModelsToOutputDTOs(instances []*model.ServiceInstance) []*ServiceInstanceOutputDTO {
	out := make([]*ServiceInstanceOutputDTO, len(instances))
	for i, inst := range instances {
		out[i] = convertModelToOutputDTO(inst)
	}
	return out
}

// CreateServiceInstance creates a new service instance.
func (s *serviceInstanceServiceImpl) CreateServiceInstance(ctx context.Context, input *ServiceInstanceInputDTO) (*ServiceInstanceOutputDTO, error) {
	s.logger.Info("Attempting to create service instance", zap.Any("input", input))

	// Validate ServiceID
	_, err := s.serviceRepo.GetByID(ctx, input.ServiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service not found during instance creation", zap.Uint("serviceId", input.ServiceID))
			return nil, fmt.Errorf("%w: service with ID %d not found", apputils.ErrBadRequest, input.ServiceID)
		}
		s.logger.Error("Failed to get service during instance creation", zap.Uint("serviceId", input.ServiceID), zap.Error(err))
		return nil, fmt.Errorf("failed to validate service: %w", err)
	}

	// Validate EnvironmentID
	_, err = s.envRepo.GetByID(ctx, input.EnvironmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Environment not found during instance creation", zap.Uint("environmentId", input.EnvironmentID))
			return nil, fmt.Errorf("%w: environment with ID %d not found", apputils.ErrBadRequest, input.EnvironmentID)
		}
		s.logger.Error("Failed to get environment during instance creation", zap.Uint("environmentId", input.EnvironmentID), zap.Error(err))
		return nil, fmt.Errorf("failed to validate environment: %w", err)
	}

	// Check for existing instance
	exists, err := s.repo.CheckExists(ctx, input.ServiceID, input.EnvironmentID, input.Version, 0)
	if err != nil {
		s.logger.Error("Failed to check for existing service instance", zap.Error(err))
		return nil, fmt.Errorf("failed to check for existing instance: %w", err)
	}
	if exists {
		msg := fmt.Sprintf("service instance with service ID %d, environment ID %d, and version '%s' already exists", input.ServiceID, input.EnvironmentID, input.Version)
		s.logger.Warn(msg)
		return nil, fmt.Errorf("%w: %s", apputils.ErrAlreadyExists, msg)
	}

	instance := &model.ServiceInstance{
		ServiceID:     input.ServiceID,
		EnvironmentID: input.EnvironmentID,
		Version:       input.Version,
		Status:        model.ServiceInstanceStatusType(input.Status),
		Hostname:      input.Hostname,
		Port:          input.Port,
		Config:        input.Config,
		DeployedAt:    input.DeployedAt,
	}

	if err := s.repo.Create(ctx, instance); err != nil {
		s.logger.Error("Failed to create service instance in repository", zap.Error(err))
		return nil, fmt.Errorf("failed to create service instance: %w", err)
	}

	s.logger.Info("Service instance created successfully", zap.Uint("instanceId", instance.ID))
	return convertModelToOutputDTO(instance), nil
}

// GetServiceInstanceByID retrieves a service instance by its ID.
func (s *serviceInstanceServiceImpl) GetServiceInstanceByID(ctx context.Context, id uint) (*ServiceInstanceOutputDTO, error) {
	s.logger.Info("Getting service instance by ID", zap.Uint("id", id))
	instance, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service instance not found by ID", zap.Uint("id", id))
			return nil, apputils.ErrNotFound
		}
		s.logger.Error("Failed to get service instance by ID from repository", zap.Uint("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get service instance: %w", err)
	}
	return convertModelToOutputDTO(instance), nil
}

// ListServiceInstances retrieves a paginated list of service instances.
func (s *serviceInstanceServiceImpl) ListServiceInstances(ctx context.Context, params *repository.ListServiceInstancesParams) (*ListServiceInstancesResponseDTO, error) {
	s.logger.Info("Listing service instances", zap.Any("params", params))

	if params.Status != nil {
		if !model.ServiceInstanceStatusType(*params.Status).IsValid() {
			msg := fmt.Sprintf("invalid status value '%s' for listing service instances", *params.Status)
			s.logger.Warn(msg)
			return nil, fmt.Errorf("%w: %s", apputils.ErrBadRequest, msg)
		}
	}

	instances, total, err := s.repo.List(ctx, params)
	if err != nil {
		s.logger.Error("Failed to list service instances from repository", zap.Error(err))
		return nil, fmt.Errorf("failed to list service instances: %w", err)
	}

	return &ListServiceInstancesResponseDTO{
		Items: convertModelsToOutputDTOs(instances),
		Total: total,
		Page:  params.Page,
		Size:  params.PageSize,
	}, nil
}

// UpdateServiceInstance updates an existing service instance.
func (s *serviceInstanceServiceImpl) UpdateServiceInstance(ctx context.Context, id uint, input *ServiceInstanceInputDTO) (*ServiceInstanceOutputDTO, error) {
	s.logger.Info("Updating service instance", zap.Uint("id", id), zap.Any("input", input))

	instance, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service instance not found for update", zap.Uint("id", id))
			return nil, apputils.ErrNotFound
		}
		s.logger.Error("Failed to get service instance for update", zap.Uint("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve instance for update: %w", err)
	}

	// For this service, we prevent changing ServiceID and EnvironmentID after creation.
	if input.ServiceID != instance.ServiceID {
		s.logger.Warn("Attempt to change ServiceID during update, which is not allowed.", zap.Uint("id", id), zap.Uint("currentServiceID", instance.ServiceID), zap.Uint("inputServiceID", input.ServiceID))
		return nil, fmt.Errorf("%w: serviceId cannot be changed after creation", apputils.ErrBadRequest)
	}
	if input.EnvironmentID != instance.EnvironmentID {
		s.logger.Warn("Attempt to change EnvironmentID during update, which is not allowed.", zap.Uint("id", id), zap.Uint("currentEnvID", instance.EnvironmentID), zap.Uint("inputEnvID", input.EnvironmentID))
		return nil, fmt.Errorf("%w: environmentId cannot be changed after creation", apputils.ErrBadRequest)
	}

	if input.Version != instance.Version {
		exists, err := s.repo.CheckExists(ctx, instance.ServiceID, instance.EnvironmentID, input.Version, id)
		if err != nil {
			s.logger.Error("Failed to check for existing service instance during update", zap.Error(err))
			return nil, fmt.Errorf("failed to check for existing instance during update: %w", err)
		}
		if exists {
			msg := fmt.Sprintf("service instance with service ID %d, environment ID %d, and version '%s' already exists", instance.ServiceID, instance.EnvironmentID, input.Version)
			s.logger.Warn(msg)
			return nil, fmt.Errorf("%w: %s", apputils.ErrAlreadyExists, msg)
		}
		instance.Version = input.Version
	}

	instance.Status = model.ServiceInstanceStatusType(input.Status)
	instance.Hostname = input.Hostname
	instance.Port = input.Port
	instance.Config = input.Config
	instance.DeployedAt = input.DeployedAt

	if err := s.repo.Update(ctx, instance); err != nil {
		s.logger.Error("Failed to update service instance in repository", zap.Uint("id", id), zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apputils.ErrNotFound
		}
		return nil, fmt.Errorf("failed to update service instance: %w", err)
	}

	s.logger.Info("Service instance updated successfully", zap.Uint("id", instance.ID))
	return convertModelToOutputDTO(instance), nil
}

// DeleteServiceInstance deletes a service instance by its ID.
func (s *serviceInstanceServiceImpl) DeleteServiceInstance(ctx context.Context, id uint) error {
	s.logger.Info("Deleting service instance", zap.Uint("id", id))

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service instance not found for deletion", zap.Uint("id", id))
			return apputils.ErrNotFound
		}
		s.logger.Error("Failed to get service instance before deletion", zap.Uint("id", id), zap.Error(err))
		return fmt.Errorf("failed to retrieve instance for deletion: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete service instance from repository", zap.Uint("id", id), zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apputils.ErrNotFound
		}
		return fmt.Errorf("failed to delete service instance: %w", err)
	}

	s.logger.Info("Service instance deleted successfully", zap.Uint("id", id))
	return nil
}
