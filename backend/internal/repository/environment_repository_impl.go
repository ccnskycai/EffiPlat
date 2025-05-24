package repository

import (
	"EffiPlat/backend/internal/model"
	apputils "EffiPlat/backend/internal/utils"
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type gormEnvironmentRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewGormEnvironmentRepository creates a new GORM-based EnvironmentRepository.
func NewGormEnvironmentRepository(db *gorm.DB, logger *zap.Logger) EnvironmentRepository {
	return &gormEnvironmentRepository{
		db:     db,
		logger: logger,
	}
}

func (r *gormEnvironmentRepository) Create(ctx context.Context, env *model.Environment) (*model.Environment, error) {
	r.logger.Debug("GORM: Creating environment", zap.Any("environment", env))
	if err := r.db.WithContext(ctx).Create(env).Error; err != nil {
		r.logger.Error("GORM: Failed to create environment", zap.Error(err))
		// TODO: Handle specific DB errors, e.g., unique constraint violation for Name or Slug
		return nil, err
	}
	return env, nil
}

func (r *gormEnvironmentRepository) List(ctx context.Context, params model.EnvironmentListParams) ([]model.Environment, int64, error) {
	r.logger.Debug("GORM: Listing environments", zap.Any("params", params))
	var environments []model.Environment
	var total int64

	tx := r.db.WithContext(ctx).Model(&model.Environment{})

	if params.Name != "" {
		tx = tx.Where("name LIKE ?", "%"+params.Name+"%")
	}
	if params.Slug != "" {
		tx = tx.Where("slug LIKE ?", "%"+params.Slug+"%")
	}

	// Get total count before pagination
	if err := tx.Count(&total).Error; err != nil {
		r.logger.Error("GORM: Failed to count environments", zap.Error(err))
		return nil, 0, err
	}

	// Apply pagination
	offset := (params.Page - 1) * params.PageSize
	if err := tx.Limit(params.PageSize).Offset(offset).Order("created_at DESC").Find(&environments).Error; err != nil {
		r.logger.Error("GORM: Failed to list environments", zap.Error(err))
		return nil, 0, err
	}

	return environments, total, nil
}

func (r *gormEnvironmentRepository) GetByID(ctx context.Context, id uint) (*model.Environment, error) {
	r.logger.Debug("GORM: Getting environment by ID", zap.Uint("id", id))
	var env model.Environment
	if err := r.db.WithContext(ctx).First(&env, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("GORM: Environment not found by ID", zap.Uint("id", id), zap.Error(err))
			return nil, gorm.ErrRecordNotFound
		}
		r.logger.Error("GORM: Failed to get environment by ID", zap.Error(err))
		return nil, err
	}
	return &env, nil
}

func (r *gormEnvironmentRepository) GetBySlug(ctx context.Context, slug string) (*model.Environment, error) {
	r.logger.Debug("GORM: Getting environment by slug", zap.String("slug", slug))
	var env model.Environment
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&env).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("GORM: Environment not found by slug", zap.String("slug", slug), zap.Error(err))
			return nil, gorm.ErrRecordNotFound
		}
		r.logger.Error("GORM: Failed to get environment by slug", zap.Error(err))
		return nil, err
	}
	return &env, nil
}

func (r *gormEnvironmentRepository) Update(ctx context.Context, env *model.Environment) (*model.Environment, error) {
	r.logger.Debug("GORM: Updating environment", zap.Any("environment", env))
	if env.ID == 0 {
		return nil, apputils.ErrMissingID
	}

	// GORM's Save will update all fields of the struct if primary key is set.
	// It will also update UpdatedAt and respect gorm.field "<-:update" or "->:false" for read-only fields.
	if err := r.db.WithContext(ctx).Save(env).Error; err != nil {
		r.logger.Error("GORM: Failed to update environment", zap.Error(err))
		// TODO: Handle specific DB errors, e.g., unique constraint violation for Name or Slug on update
		return nil, err
	}
	return env, nil
}

func (r *gormEnvironmentRepository) Delete(ctx context.Context, id uint) error {
	r.logger.Debug("GORM: Deleting environment", zap.Uint("id", id))
	// Using GORM's soft delete (updates DeletedAt field)
	result := r.db.WithContext(ctx).Delete(&model.Environment{}, id)
	if result.Error != nil {
		r.logger.Error("GORM: Failed to delete environment", zap.Error(result.Error))
		return result.Error
	}
	if result.RowsAffected == 0 {
		r.logger.Warn("GORM: Environment not found for deletion or no rows affected", zap.Uint("id", id))
		return gorm.ErrRecordNotFound
	}
	return nil
}
