package repository

import (
	"EffiPlat/backend/internal/models"
	"context"

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
