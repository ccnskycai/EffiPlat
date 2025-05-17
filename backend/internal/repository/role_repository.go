package repository

import (
	"EffiPlat/backend/internal/models"
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RoleRepository defines the interface for database operations on roles.
type RoleRepository interface {
	CreateRole(ctx context.Context, role *models.Role) (*models.Role, error)
	ListRoles(ctx context.Context, params models.RoleListParams) ([]models.Role, int64, error)
	GetRoleByID(ctx context.Context, id uint) (*models.Role, error)
	UpdateRole(ctx context.Context, id uint, role *models.Role) (*models.Role, error)
	DeleteRole(ctx context.Context, id uint) error

	// Role-Permission Association Methods
	AddPermissionsToRole(ctx context.Context, role *models.Role, permissions []models.Permission) error
	RemovePermissionsFromRole(ctx context.Context, role *models.Role, permissions []models.Permission) error
	GetRoleWithPermissions(ctx context.Context, roleID uint) (*models.Role, error)
}

type RoleRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewRoleRepository creates a new instance of RoleRepository.
func NewRoleRepository(db *gorm.DB, logger *zap.Logger) *RoleRepositoryImpl {
	return &RoleRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// Implement RoleRepository methods below

func (r *RoleRepositoryImpl) CreateRole(ctx context.Context, role *models.Role) (*models.Role, error) {
	r.logger.Info("RoleRepository: Creating role", zap.String("name", role.Name))
	if err := r.db.WithContext(ctx).Create(role).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).First(role, role.ID).Error; err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleRepositoryImpl) ListRoles(ctx context.Context, params models.RoleListParams) ([]models.Role, int64, error) {
	r.logger.Info("RoleRepository: Listing roles", zap.Any("params", params))
	var roles []models.Role
	var total int64
	db := r.db.WithContext(ctx).Model(&models.Role{})
	if params.Name != "" {
		db = db.Where("name LIKE ?", "%"+params.Name+"%")
	}
	db.Count(&total)
	err := db.Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).Find(&roles).Error
	if err != nil {
		return nil, 0, err
	}
	return roles, total, nil
}

func (r *RoleRepositoryImpl) GetRoleByID(ctx context.Context, id uint) (*models.Role, error) {
	r.logger.Info("RoleRepository: GetRoleByID", zap.Uint("id", id))
	var role models.Role
	if err := r.db.WithContext(ctx).First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepositoryImpl) UpdateRole(ctx context.Context, id uint, role *models.Role) (*models.Role, error) {
	r.logger.Info("RoleRepository: UpdateRole", zap.Uint("id", id))
	var existing models.Role
	if err := r.db.WithContext(ctx).First(&existing, id).Error; err != nil {
		return nil, err
	}
	existing.Name = role.Name
	existing.Description = role.Description
	if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (r *RoleRepositoryImpl) DeleteRole(ctx context.Context, id uint) error {
	r.logger.Info("RoleRepository: DeleteRole", zap.Uint("id", id))
	return r.db.WithContext(ctx).Delete(&models.Role{}, id).Error
}

// AddPermissionsToRole adds a list of permissions to a role.
func (r *RoleRepositoryImpl) AddPermissionsToRole(ctx context.Context, role *models.Role, permissions []models.Permission) error {
	r.logger.Info("RoleRepository: Adding permissions to role", zap.Uint("roleID", role.ID), zap.Int("permissionCount", len(permissions)))
	// Use Append to only add new associations, ignoring existing ones
	if err := r.db.WithContext(ctx).Model(role).Association("Permissions").Append(permissions); err != nil {
		return fmt.Errorf("failed to add permissions to role: %w", err)
	}
	return nil
}

// RemovePermissionsFromRole removes a list of permissions from a role.
func (r *RoleRepositoryImpl) RemovePermissionsFromRole(ctx context.Context, role *models.Role, permissions []models.Permission) error {
	r.logger.Info("RoleRepository: Removing permissions from role", zap.Uint("roleID", role.ID), zap.Int("permissionCount", len(permissions)))
	// Use Delete to remove the associations
	if err := r.db.WithContext(ctx).Model(role).Association("Permissions").Delete(permissions); err != nil {
		return fmt.Errorf("failed to remove permissions from role: %w", err)
	}
	return nil
}

// GetRoleWithPermissions gets a role by ID and preloads its associated permissions.
func (r *RoleRepositoryImpl) GetRoleWithPermissions(ctx context.Context, roleID uint) (*models.Role, error) {
	r.logger.Info("RoleRepository: GetRoleWithPermissions", zap.Uint("roleID", roleID))
	var role models.Role
	// Preload the Permissions association
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&role, roleID).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
