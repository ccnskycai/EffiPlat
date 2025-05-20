package service

import (
	"EffiPlat/backend/internal/models"     // Assuming models.Responsibility and models.ResponsibilityListParams exist or will be created
	"EffiPlat/backend/internal/repository" // Assuming repository.ResponsibilityRepository will exist
	"context"
	"errors" // Import errors package

	"go.uber.org/zap"
	"gorm.io/gorm" // Import gorm
)

// ResponsibilityService defines the interface for responsibility-related operations.
// It's good practice to define an interface for the service layer to allow for easier testing and DI.
type ResponsibilityService interface {
	CreateResponsibility(ctx context.Context, responsibility *models.Responsibility) (*models.Responsibility, error)
	GetResponsibilities(ctx context.Context, params models.ResponsibilityListParams) ([]models.Responsibility, int64, error) // Returns items, total count, error
	GetResponsibilityByID(ctx context.Context, id uint) (*models.Responsibility, error)
	UpdateResponsibility(ctx context.Context, id uint, responsibilityUpdate *models.Responsibility) (*models.Responsibility, error)
	DeleteResponsibility(ctx context.Context, id uint) error
}

type responsibilityServiceImpl struct {
	repo   repository.ResponsibilityRepository // Placeholder for actual repository
	logger *zap.Logger
}

// NewResponsibilityService creates a new instance of ResponsibilityService.
func NewResponsibilityService(repo repository.ResponsibilityRepository, logger *zap.Logger) ResponsibilityService {
	return &responsibilityServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

func (s *responsibilityServiceImpl) CreateResponsibility(ctx context.Context, r *models.Responsibility) (*models.Responsibility, error) {
	s.logger.Info("Service: Creating new responsibility", zap.String("name", r.Name))
	// TODO: Add validation - e.g., check for existing name if it should be unique beyond DB constraint
	return s.repo.Create(ctx, r)
}

func (s *responsibilityServiceImpl) GetResponsibilities(ctx context.Context, params models.ResponsibilityListParams) ([]models.Responsibility, int64, error) {
	s.logger.Info("Service: Fetching responsibilities", zap.Any("params", params))
	return s.repo.List(ctx, params)
}

func (s *responsibilityServiceImpl) GetResponsibilityByID(ctx context.Context, id uint) (*models.Responsibility, error) {
	s.logger.Info("Service: Fetching responsibility by ID", zap.Uint("id", id))
	resp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service: Responsibility not found by ID", zap.Uint("id", id))
			return nil, ErrResponsibilityNotFound
		}
		s.logger.Error("Service: Error fetching responsibility by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func (s *responsibilityServiceImpl) UpdateResponsibility(ctx context.Context, id uint, responsibilityUpdate *models.Responsibility) (*models.Responsibility, error) {
	s.logger.Info("Service: Updating responsibility", zap.Uint("id", id), zap.String("newName", responsibilityUpdate.Name))

	// First, check if the responsibility exists
	existingResp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service: Responsibility not found for update", zap.Uint("id", id))
			return nil, ErrResponsibilityNotFound
		}
		s.logger.Error("Service: Error fetching responsibility for update", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}

	// Apply updates from responsibilityUpdate to existingResp
	// This allows for partial updates if responsibilityUpdate doesn't contain all fields
	existingResp.Name = responsibilityUpdate.Name
	existingResp.Description = responsibilityUpdate.Description
	// Note: ID, CreatedAt, UpdatedAt should be handled by GORM or are not typically updated this way.

	return s.repo.Update(ctx, existingResp)
}

func (s *responsibilityServiceImpl) DeleteResponsibility(ctx context.Context, id uint) error {
	s.logger.Info("Service: Deleting responsibility", zap.Uint("id", id))
	// Optional: Check if responsibility is in use before deleting
	// For example, query if any ResponsibilityGroup uses this ResponsibilityID

	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Depending on desired behavior, deleting a non-existent item might not be an error.
			// For now, we'll treat it as "not found" similar to GetByID.
			s.logger.Warn("Service: Responsibility not found for deletion", zap.Uint("id", id))
			return ErrResponsibilityNotFound
		}
		s.logger.Error("Service: Error deleting responsibility", zap.Uint("id", id), zap.Error(err))
		return err
	}
	return nil
}
