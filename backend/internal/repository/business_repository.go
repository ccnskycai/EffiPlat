package repository

import (
	"EffiPlat/backend/internal/model"
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ListBusinessesParams 定义了列出业务时的过滤和分页参数
type ListBusinessesParams struct {
	Page     int
	PageSize int
	Name     *string                   // 按名称过滤 (可选)
	Status   *model.BusinessStatusType // 按状态过滤 (可选)
	Owner    *string                   // 按负责人过滤 (可选)
	SortBy   string                    // 排序字段 (e.g., "name", "createdAt")
	Order    string                    // 排序顺序 ("asc", "desc")
}

// BusinessRepository 定义业务相关的数据库操作接口
type BusinessRepository interface {
	Create(ctx context.Context, business *model.Business) error
	GetByID(ctx context.Context, id uint) (*model.Business, error)
	GetByName(ctx context.Context, name string) (*model.Business, error)
	List(ctx context.Context, params *ListBusinessesParams) ([]model.Business, int64, error)
	Update(ctx context.Context, business *model.Business) error
	Delete(ctx context.Context, id uint) error
	CheckExists(ctx context.Context, name string, excludeID uint) (bool, error) // 检查同名业务是否存在，可排除特定ID
}

type businessRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewBusinessRepository 创建一个新的 BusinessRepository 实例
func NewBusinessRepository(db *gorm.DB, logger *zap.Logger) BusinessRepository {
	return &businessRepositoryImpl{db: db, logger: logger}
}

// Create 创建一个新的业务记录
func (r *businessRepositoryImpl) Create(ctx context.Context, business *model.Business) error {
	r.logger.Debug("Creating business", zap.Any("business", business))
	if err := r.db.WithContext(ctx).Create(business).Error; err != nil {
		r.logger.Error("Failed to create business", zap.Error(err))
		return fmt.Errorf("repository.Create: %w", err)
	}
	r.logger.Info("Business created successfully", zap.Uint("id", business.ID))
	return nil
}

// GetByID 通过ID获取业务记录
func (r *businessRepositoryImpl) GetByID(ctx context.Context, id uint) (*model.Business, error) {
	r.logger.Debug("Getting business by ID", zap.Uint("id", id))
	var business model.Business
	if err := r.db.WithContext(ctx).First(&business, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Warn("Business not found by ID", zap.Uint("id", id))
			return nil, fmt.Errorf("repository.GetByID: %w", gorm.ErrRecordNotFound)
		}
		r.logger.Error("Failed to get business by ID", zap.Uint("id", id), zap.Error(err))
		return nil, fmt.Errorf("repository.GetByID: %w", err)
	}
	return &business, nil
}

// GetByName 通过名称获取业务记录
func (r *businessRepositoryImpl) GetByName(ctx context.Context, name string) (*model.Business, error) {
	r.logger.Debug("Getting business by name", zap.String("name", name))
	var business model.Business
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&business).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Warn("Business not found by name", zap.String("name", name))
			return nil, fmt.Errorf("repository.GetByName: %w", gorm.ErrRecordNotFound)
		}
		r.logger.Error("Failed to get business by name", zap.String("name", name), zap.Error(err))
		return nil, fmt.Errorf("repository.GetByName: %w", err)
	}
	return &business, nil
}

// List 列出业务记录，支持过滤和分页
func (r *businessRepositoryImpl) List(ctx context.Context, params *ListBusinessesParams) ([]model.Business, int64, error) {
	r.logger.Debug("Listing businesses", zap.Any("params", params))
	var businesses []model.Business
	var total int64

	dbQuery := r.db.WithContext(ctx).Model(&model.Business{})

	if params.Name != nil && *params.Name != "" {
		dbQuery = dbQuery.Where("name LIKE ?", "%"+*params.Name+"%")
	}
	if params.Status != nil {
		dbQuery = dbQuery.Where("status = ?", *params.Status)
	}
	if params.Owner != nil && *params.Owner != "" {
		dbQuery = dbQuery.Where("owner LIKE ?", "%"+*params.Owner+"%")
	}

	// Count total records for pagination
	if err := dbQuery.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count businesses", zap.Error(err))
		return nil, 0, fmt.Errorf("repository.List.Count: %w", err)
	}

	// Apply sorting
	sortBy := "created_at"
	if params.SortBy != "" {
		sortBy = params.SortBy
	}
	order := "desc"
	if params.Order != "" && (params.Order == "asc" || params.Order == "desc") {
		order = params.Order
	}
	dbQuery = dbQuery.Order(fmt.Sprintf("%s %s", sortBy, order))

	// Apply pagination
	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		dbQuery = dbQuery.Offset(offset).Limit(params.PageSize)
	}

	if err := dbQuery.Find(&businesses).Error; err != nil {
		r.logger.Error("Failed to list businesses", zap.Error(err))
		return nil, 0, fmt.Errorf("repository.List.Find: %w", err)
	}

	return businesses, total, nil
}

// Update 更新现有的业务记录
func (r *businessRepositoryImpl) Update(ctx context.Context, business *model.Business) error {
	r.logger.Debug("Updating business", zap.Uint("id", business.ID), zap.Any("businessChanges", business))
	if business.ID == 0 {
		return fmt.Errorf("repository.Update: business ID is required")
	}

	// GORM's Updates method only updates non-zero fields.
	// If you want to allow clearing fields (e.g., setting Owner to ""), use Select to specify all fields
	// or use a map for updates.
	// For now, we assume standard behavior of Updates.
	result := r.db.WithContext(ctx).Model(&model.Business{}).Where("id = ?", business.ID).Updates(business)
	if result.Error != nil {
		r.logger.Error("Failed to update business", zap.Uint("id", business.ID), zap.Error(result.Error))
		return fmt.Errorf("repository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		// Check if the record actually exists, as RowsAffected can be 0 if no fields were changed OR if the record doesn't exist.
		var count int64
		if err := r.db.Model(&model.Business{}).Where("id = ?", business.ID).Count(&count).Error; err != nil {
			r.logger.Error("Failed to count business for update check", zap.Uint("id", business.ID), zap.Error(err))
			return fmt.Errorf("repository.Update.Count: %w", err)
		}
		if count == 0 {
			r.logger.Warn("Business not found for update", zap.Uint("id", business.ID))
			return fmt.Errorf("repository.Update: %w", gorm.ErrRecordNotFound)
		}
		// If count > 0 and RowsAffected == 0, it means no fields were actually changed. This is not an error.
		r.logger.Debug("No fields changed for business update", zap.Uint("id", business.ID))
	}

	r.logger.Info("Business updated successfully", zap.Uint("id", business.ID))
	return nil
}

// Delete (soft)删除一个业务记录
func (r *businessRepositoryImpl) Delete(ctx context.Context, id uint) error {
	r.logger.Debug("Deleting business", zap.Uint("id", id))
	if id == 0 {
		return fmt.Errorf("repository.Delete: business ID is required")
	}

	result := r.db.WithContext(ctx).Delete(&model.Business{}, id)
	if result.Error != nil {
		r.logger.Error("Failed to delete business", zap.Uint("id", id), zap.Error(result.Error))
		return fmt.Errorf("repository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		r.logger.Warn("Business not found for deletion or already deleted", zap.Uint("id", id))
		return fmt.Errorf("repository.Delete: %w", gorm.ErrRecordNotFound) // GORM itself returns this if RowsAffected is 0 for delete
	}

	r.logger.Info("Business deleted successfully", zap.Uint("id", id))
	return nil
}

// CheckExists 检查具有给定名称的业务是否已存在 (可选择排除特定ID)
func (r *businessRepositoryImpl) CheckExists(ctx context.Context, name string, excludeID uint) (bool, error) {
	r.logger.Debug("Checking if business exists", zap.String("name", name), zap.Uint("excludeID", excludeID))
	var count int64
	dbQuery := r.db.WithContext(ctx).Model(&model.Business{}).Where("name = ?", name)
	if excludeID > 0 {
		dbQuery = dbQuery.Where("id <> ?", excludeID)
	}
	if err := dbQuery.Count(&count).Error; err != nil {
		r.logger.Error("Failed to count businesses for existence check", zap.Error(err))
		return false, fmt.Errorf("repository.CheckExists: %w", err)
	}
	return count > 0, nil
}
