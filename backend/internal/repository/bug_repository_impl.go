package repository

import (
	"EffiPlat/backend/internal/model" // Corrected import path
	"context"
	"errors" // Using standard errors for now
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var ErrBugNotFound = errors.New("bug not found") // Custom error for repository

// bugRepositoryImpl implements the BugRepository interface using GORM.
type bugRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewBugRepository creates a new instance of bugRepositoryImpl.
func NewBugRepository(db *gorm.DB, logger *zap.Logger) BugRepository {
	return &bugRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// Create creates a new bug record in the database.
func (r *bugRepositoryImpl) Create(ctx context.Context, bug *model.Bug) error {
	return r.db.WithContext(ctx).Create(bug).Error
}

// GetByID retrieves a bug by its ID.
func (r *bugRepositoryImpl) GetByID(ctx context.Context, id uint) (*model.Bug, error) {
	var bug model.Bug
	if err := r.db.WithContext(ctx).First(&bug, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBugNotFound // Use custom error from this package or a common one
		}
		return nil, err
	}
	return &bug, nil
}

// Update updates an existing bug record in the database.
func (r *bugRepositoryImpl) Update(ctx context.Context, bug *model.Bug) error {
	// First check if the bug exists
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Bug{}).Where("id = ? AND deleted_at IS NULL", bug.ID).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		r.logger.Debug("Bug not found for update", zap.Uint("id", bug.ID))
		return ErrBugNotFound
	}

	// Then proceed with update
	result := r.db.WithContext(ctx).Model(&model.Bug{}).Where("id = ?", bug.ID).Updates(bug)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// Delete removes a bug record from the database by its ID.
func (r *bugRepositoryImpl) Delete(ctx context.Context, id uint) error {
	// First check if the bug exists
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Bug{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		r.logger.Debug("Bug not found for deletion", zap.Uint("id", id))
		return ErrBugNotFound
	}

	// Then proceed with deletion
	result := r.db.WithContext(ctx).Delete(&model.Bug{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// List retrieves a list of bugs based on the provided parameters, with pagination.
func (r *bugRepositoryImpl) List(ctx context.Context, params *model.BugListParams) ([]*model.Bug, int64, error) {
	r.logger.Debug("GORM: Listing bugs", zap.Any("params", params))
	var bugs []*model.Bug
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&model.Bug{})

	// Apply filters
	if params.Title != "" {
		query = query.Where("title LIKE ?", fmt.Sprintf("%%%s%%", params.Title))
	}
	if params.Status != "" { // Assuming params.Status is model.BugStatusType
		query = query.Where("status = ?", params.Status)
	}
	if params.Priority != "" { // Assuming params.Priority is model.BugPriorityType
		query = query.Where("priority = ?", params.Priority)
	}
	if params.AssigneeID != nil {
		query = query.Where("assignee_id = ?", *params.AssigneeID)
	}
	if params.ReporterID != nil {
		query = query.Where("reporter_id = ?", *params.ReporterID)
	}

	// Get total count before pagination
	if err := query.Count(&totalCount).Error; err != nil {
		r.logger.Error("GORM: Failed to count bugs", zap.Error(err))
		return nil, 0, err
	}

	// Apply pagination
	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Order("created_at DESC").Find(&bugs).Error; err != nil {
		r.logger.Error("GORM: Failed to list bugs", zap.Error(err))
		return nil, 0, err
	}

	return bugs, totalCount, nil
}

// CountBugsByAssigneeID counts the number of bugs assigned to a specific user.
func (r *bugRepositoryImpl) CountBugsByAssigneeID(ctx context.Context, assigneeID uint) (int64, error) {
	r.logger.Debug("GORM: Counting bugs by assignee ID", zap.Uint("assigneeID", assigneeID))
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Bug{}).Where("assignee_id = ?", assigneeID).Count(&count).Error; err != nil {
		r.logger.Error("GORM: Failed to count bugs by assignee ID", zap.Error(err))
		return 0, err
	}
	return count, nil
}

// CountBugsByEnvironmentID counts the number of bugs in a specific environment.
func (r *bugRepositoryImpl) CountBugsByEnvironmentID(ctx context.Context, environmentID uint) (int64, error) {
	r.logger.Debug("GORM: Counting bugs by environment ID", zap.Uint("environmentID", environmentID))
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Bug{}).Where("environment_id = ?", environmentID).Count(&count).Error; err != nil {
		r.logger.Error("GORM: Failed to count bugs by environment ID", zap.Error(err))
		return 0, err
	}
	return count, nil
}

// GetBugsByStatus retrieves bugs with a specific status.
func (r *bugRepositoryImpl) GetBugsByStatus(ctx context.Context, status model.BugStatusType, params *model.BugListParams) ([]*model.Bug, int64, error) {
	r.logger.Debug("GORM: Getting bugs by status", zap.Any("status", status), zap.Any("params", params))
	
	// Create a copy of params and override the status
	newParams := *params
	newParams.Status = status
	return r.List(ctx, &newParams)
}
