package repository

import (
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/utils" // Added for utils.ErrMissingID
	"context"
	"errors" // For placeholder errors initially

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type gormResponsibilityRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewGormResponsibilityRepository creates a new GORM-based ResponsibilityRepository.
func NewGormResponsibilityRepository(db *gorm.DB, logger *zap.Logger) ResponsibilityRepository {
	return &gormResponsibilityRepository{
		db:     db,
		logger: logger,
	}
}

func (r *gormResponsibilityRepository) Create(ctx context.Context, resp *models.Responsibility) (*models.Responsibility, error) {
	r.logger.Debug("GORM: Creating responsibility", zap.Any("responsibility", resp))
	if err := r.db.WithContext(ctx).Create(resp).Error; err != nil {
		r.logger.Error("GORM: Failed to create responsibility", zap.Error(err))
		// TODO: Handle specific DB errors, e.g., unique constraint violation
		return nil, err
	}
	return resp, nil
}

func (r *gormResponsibilityRepository) List(ctx context.Context, params models.ResponsibilityListParams) ([]models.Responsibility, int64, error) {
	r.logger.Debug("GORM: Listing responsibilities", zap.Any("params", params))
	var responsibilities []models.Responsibility
	var total int64

	tx := r.db.WithContext(ctx).Model(&models.Responsibility{})

	if params.Name != "" {
		tx = tx.Where("name LIKE ?", "%"+params.Name+"%")
	}

	// Get total count before pagination
	if err := tx.Count(&total).Error; err != nil {
		r.logger.Error("GORM: Failed to count responsibilities", zap.Error(err))
		return nil, 0, err
	}

	// Apply pagination
	offset := (params.Page - 1) * params.PageSize
	if err := tx.Limit(params.PageSize).Offset(offset).Find(&responsibilities).Error; err != nil {
		r.logger.Error("GORM: Failed to list responsibilities", zap.Error(err))
		return nil, 0, err
	}

	return responsibilities, total, nil
}

func (r *gormResponsibilityRepository) GetByID(ctx context.Context, id uint) (*models.Responsibility, error) {
	r.logger.Debug("GORM: Getting responsibility by ID", zap.Uint("id", id))
	var resp models.Responsibility
	if err := r.db.WithContext(ctx).First(&resp, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("GORM: Responsibility not found", zap.Uint("id", id), zap.Error(err))
			return nil, gorm.ErrRecordNotFound // Return gorm.ErrRecordNotFound directly
		}
		r.logger.Error("GORM: Failed to get responsibility by ID", zap.Error(err))
		return nil, err
	}
	return &resp, nil
}

func (r *gormResponsibilityRepository) Update(ctx context.Context, resp *models.Responsibility) (*models.Responsibility, error) {
	r.logger.Debug("GORM: Updating responsibility", zap.Any("responsibility", resp))
	// Ensure ID is present for update, GORM uses primary key from the struct
	if resp.ID == 0 {
		return nil, utils.ErrMissingID
	}
	// Save will update all fields, or create if not found (but we should ensure it exists first)
	// Consider using .Model(&models.Responsibility{}).Where("id = ?", resp.ID).Updates(resp) for partial updates
	if err := r.db.WithContext(ctx).Save(resp).Error; err != nil {
		r.logger.Error("GORM: Failed to update responsibility", zap.Error(err))
		// TODO: Handle specific DB errors
		return nil, err
	}
	return resp, nil
}

func (r *gormResponsibilityRepository) Delete(ctx context.Context, id uint) error {
	r.logger.Debug("GORM: Deleting responsibility", zap.Uint("id", id))
	result := r.db.WithContext(ctx).Delete(&models.Responsibility{}, id)
	if result.Error != nil {
		r.logger.Error("GORM: Failed to delete responsibility", zap.Error(result.Error))
		return result.Error
	}
	if result.RowsAffected == 0 {
		r.logger.Warn("GORM: Responsibility not found for deletion or no rows affected", zap.Uint("id", id))
		return gorm.ErrRecordNotFound
	}
	return nil
}
