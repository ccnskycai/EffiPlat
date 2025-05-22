package service

import (
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/repository"
	"EffiPlat/backend/internal/utils" // Import apputils
	"context"
	"errors" // For error checking
	"fmt"    // Import fmt for error wrapping

	"go.uber.org/zap"
	"gorm.io/gorm" // For gorm.ErrRecordNotFound
)

// ResponsibilityGroupService defines the interface for responsibility group operations.
type ResponsibilityGroupService interface {
	CreateResponsibilityGroup(ctx context.Context, group *models.ResponsibilityGroup, responsibilityIDs []uint) (*models.ResponsibilityGroup, error)
	GetResponsibilityGroups(ctx context.Context, params models.ResponsibilityGroupListParams) ([]models.ResponsibilityGroup, int64, error)
	GetResponsibilityGroupByID(ctx context.Context, id uint) (*models.ResponsibilityGroup, error) // Changed to return *models.ResponsibilityGroup as repo preloads
	UpdateResponsibilityGroup(ctx context.Context, id uint, groupUpdate *models.ResponsibilityGroup, responsibilityIDs *[]uint) (*models.ResponsibilityGroup, error)
	DeleteResponsibilityGroup(ctx context.Context, id uint) error
	AddResponsibilityToGroup(ctx context.Context, groupID uint, responsibilityID uint) error
	RemoveResponsibilityFromGroup(ctx context.Context, groupID uint, responsibilityID uint) error
}

type responsibilityGroupServiceImpl struct {
	groupRepo          repository.ResponsibilityGroupRepository
	responsibilityRepo repository.ResponsibilityRepository // For validation if needed
	logger             *zap.Logger
}

// NewResponsibilityGroupService creates a new instance of ResponsibilityGroupService.
func NewResponsibilityGroupService(groupRepo repository.ResponsibilityGroupRepository, respRepo repository.ResponsibilityRepository, logger *zap.Logger) ResponsibilityGroupService {
	return &responsibilityGroupServiceImpl{
		groupRepo:          groupRepo,
		responsibilityRepo: respRepo,
		logger:             logger,
	}
}

func (s *responsibilityGroupServiceImpl) CreateResponsibilityGroup(ctx context.Context, group *models.ResponsibilityGroup, responsibilityIDs []uint) (*models.ResponsibilityGroup, error) {
	s.logger.Info("Service: Creating new responsibility group", zap.String("name", group.Name), zap.Uints("responsibilityIDs", responsibilityIDs))
	// Optional: Validate that responsibilityIDs exist using s.responsibilityRepo.GetByIDs if implemented
	// For now, assuming repository Create handles this or relies on DB constraints.
	createdGroup, err := s.groupRepo.Create(ctx, group, responsibilityIDs)
	if err != nil {
		// TODO: Handle specific errors from repo, e.g., duplicate name
		s.logger.Error("Service: Failed to create responsibility group", zap.Error(err))
		return nil, err
	}
	return createdGroup, nil
}

func (s *responsibilityGroupServiceImpl) GetResponsibilityGroups(ctx context.Context, params models.ResponsibilityGroupListParams) ([]models.ResponsibilityGroup, int64, error) {
	s.logger.Info("Service: Fetching responsibility groups", zap.Any("params", params))
	return s.groupRepo.List(ctx, params)
}

func (s *responsibilityGroupServiceImpl) GetResponsibilityGroupByID(ctx context.Context, id uint) (*models.ResponsibilityGroup, error) {
	s.logger.Info("Service: Fetching responsibility group by ID", zap.Uint("id", id))
	group, err := s.groupRepo.GetByID(ctx, id) // Repo's GetByID preloads Responsibilities
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service: Responsibility group not found by ID", zap.Uint("id", id))
			return nil, fmt.Errorf("responsibility group with id %d not found: %w", id, utils.ErrNotFound)
		}
		s.logger.Error("Service: Error fetching responsibility group by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return group, nil
}

func (s *responsibilityGroupServiceImpl) UpdateResponsibilityGroup(ctx context.Context, id uint, groupUpdate *models.ResponsibilityGroup, responsibilityIDs *[]uint) (*models.ResponsibilityGroup, error) {
	s.logger.Info("Service: Updating responsibility group", zap.Uint("id", id), zap.Any("responsibilityIDs_ptr", responsibilityIDs != nil))

	groupUpdate.ID = id // Ensure ID is set for the update in the repo layer
	// Optional: Validate that responsibilityIDs (if provided) exist.

	updatedGroup, err := s.groupRepo.Update(ctx, groupUpdate, responsibilityIDs)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service: Responsibility group not found for update", zap.Uint("id", id))
			return nil, fmt.Errorf("responsibility group with id %d not found for update: %w", id, utils.ErrNotFound)
		}
		// TODO: Handle other specific errors like duplicate name after update
		s.logger.Error("Service: Failed to update responsibility group", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return updatedGroup, nil
}

func (s *responsibilityGroupServiceImpl) DeleteResponsibilityGroup(ctx context.Context, id uint) error {
	s.logger.Info("Service: Deleting responsibility group", zap.Uint("id", id))
	err := s.groupRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service: Responsibility group not found for deletion", zap.Uint("id", id))
			return fmt.Errorf("responsibility group with id %d not found for deletion: %w", id, utils.ErrNotFound)
		}
		s.logger.Error("Service: Error deleting responsibility group", zap.Uint("id", id), zap.Error(err))
		return err
	}
	return nil
}

func (s *responsibilityGroupServiceImpl) AddResponsibilityToGroup(ctx context.Context, groupID uint, responsibilityID uint) error {
	s.logger.Info("Service: Adding responsibility to group", zap.Uint("groupID", groupID), zap.Uint("responsibilityID", responsibilityID))
	// 1. Check if group exists
	_, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("responsibility group with id %d not found: %w", groupID, utils.ErrNotFound)
		}
		return err // Other error
	}
	// 2. Check if responsibility exists
	_, err = s.responsibilityRepo.GetByID(ctx, responsibilityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("responsibility with id %d not found: %w", responsibilityID, utils.ErrNotFound)
		}
		return err // Other error
	}

	// 3. Add association
	err = s.groupRepo.AddResponsibilityToGroup(ctx, groupID, responsibilityID)
	if err != nil {
		// TODO: Handle specific errors from repo, e.g., duplicate association if not handled by repo
		s.logger.Error("Service: Failed to add responsibility to group", zap.Error(err))
		return err
	}
	return nil
}

func (s *responsibilityGroupServiceImpl) RemoveResponsibilityFromGroup(ctx context.Context, groupID uint, responsibilityID uint) error {
	s.logger.Info("Service: Removing responsibility from group", zap.Uint("groupID", groupID), zap.Uint("responsibilityID", responsibilityID))
	// Optional: Check if group and responsibility exist before attempting removal for clearer error messages.
	// However, the repo method might already handle this or return gorm.ErrRecordNotFound if the association doesn't exist or if group/responsibility is invalid.
	err := s.groupRepo.RemoveResponsibilityFromGroup(ctx, groupID, responsibilityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Service: Association not found for removal", zap.Uint("groupID", groupID), zap.Uint("responsibilityID", responsibilityID), zap.Error(err))
			return fmt.Errorf("association not found for removal (group %d, responsibility %d): %w", groupID, responsibilityID, utils.ErrNotFound) // Map to service.ErrNotFound
		}
		s.logger.Error("Service: Failed to remove responsibility from group", zap.Uint("groupID", groupID), zap.Uint("responsibilityID", responsibilityID), zap.Error(err))
		return err
	}
	return nil
}
