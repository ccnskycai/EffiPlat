package repository

import (
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/utils"
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PermissionRepository defines the interface for database operations on permissions.
type PermissionRepository interface {
	CreatePermission(ctx context.Context, permission *models.Permission) (*models.Permission, error)
	ListPermissions(ctx context.Context, params models.PermissionListParams) ([]models.Permission, int64, error)
	GetPermissionByID(ctx context.Context, id uint) (*models.Permission, error)
	GetPermissionsByIDs(ctx context.Context, ids []uint) ([]models.Permission, error)
	UpdatePermission(ctx context.Context, id uint, permission *models.Permission) (*models.Permission, error)
	DeletePermission(ctx context.Context, id uint) error
	// TODO: Add methods for associating/disassociating permissions with roles if needed at repository level
}

type PermissionRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPermissionRepository creates a new instance of PermissionRepository.
func NewPermissionRepository(db *gorm.DB, logger *zap.Logger) *PermissionRepositoryImpl {
	return &PermissionRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// --- Implement PermissionRepository methods below ---

func (r *PermissionRepositoryImpl) CreatePermission(ctx context.Context, permission *models.Permission) (*models.Permission, error) {
	r.logger.Info("PermissionRepository: Creating permission", zap.String("name", permission.Name))
	if err := r.db.WithContext(ctx).Create(permission).Error; err != nil {
		// TODO: Differentiate unique constraint error
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}
	return permission, nil
}

func (r *PermissionRepositoryImpl) ListPermissions(ctx context.Context, params models.PermissionListParams) ([]models.Permission, int64, error) {
	r.logger.Info("PermissionRepository: Listing permissions", zap.Any("params", params))
	var permissions []models.Permission
	var total int64
	db := r.db.WithContext(ctx).Model(&models.Permission{})

	if params.Name != "" {
		db = db.Where("name LIKE ?", "%"+params.Name+"%")
	}
	if params.Resource != "" {
		db = db.Where("resource = ?", params.Resource)
	}
	if params.Action != "" {
		db = db.Where("action = ?", params.Action)
	}

	db.Count(&total)

	err := db.Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).Find(&permissions).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list permissions: %w", err)
	}

	return permissions, total, nil
}

func (r *PermissionRepositoryImpl) GetPermissionByID(ctx context.Context, id uint) (*models.Permission, error) {
	r.logger.Info("PermissionRepository: GetPermissionByID", zap.Uint("id", id))
	var permission models.Permission
	if err := r.db.WithContext(ctx).First(&permission, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get permission by ID: %w", err)
	}
	return &permission, nil
}

func (r *PermissionRepositoryImpl) GetPermissionsByIDs(ctx context.Context, ids []uint) ([]models.Permission, error) {
	r.logger.Info("PermissionRepository: GetPermissionsByIDs", zap.Any("ids", ids))
	var permissions []models.Permission
	if len(ids) == 0 {
		return permissions, nil // Return empty slice for empty input
	}
	if err := r.db.WithContext(ctx).Find(&permissions, ids).Error; err != nil {
		return nil, fmt.Errorf("failed to get permissions by IDs: %w", err)
	}
	return permissions, nil
}

func (r *PermissionRepositoryImpl) UpdatePermission(ctx context.Context, id uint, permission *models.Permission) (*models.Permission, error) {
	r.logger.Info("PermissionRepository: UpdatePermission", zap.Uint("id", id), zap.String("newName", permission.Name))
	var existing models.Permission
	if err := r.db.WithContext(ctx).First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find permission for update: %w", err)
	}

	// Apply updates from the provided permission object, ignoring zero values for pointers
	if permission.Name != "" {
		existing.Name = permission.Name
	}
	if permission.Description != "" {
		existing.Description = permission.Description
	}
	if permission.Resource != "" {
		existing.Resource = permission.Resource
	}
	if permission.Action != "" {
		existing.Action = permission.Action
	}

	// TODO: Handle unique constraint violation on update
	if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return &existing, nil
}

func (r *PermissionRepositoryImpl) DeletePermission(ctx context.Context, id uint) error {
	r.logger.Info("PermissionRepository: DeletePermission", zap.Uint("id", id))
	// TODO: Implement logic to prevent deletion if permission is associated with roles

	result := r.db.WithContext(ctx).Delete(&models.Permission{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete permission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.ErrNotFound
	}

	return nil
}
