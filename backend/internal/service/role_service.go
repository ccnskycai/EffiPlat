package service

import (
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/repository"
	"EffiPlat/backend/internal/utils"
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RoleService defines the interface for role management operations.
type RoleService interface {
	CreateRole(ctx context.Context, roleData *models.Role, permissionIDs []uint) (*models.Role, error)
	GetRoles(ctx context.Context, params models.RoleListParams) ([]models.Role, int64, error)
	GetRoleByID(ctx context.Context, id uint) (*models.RoleDetails, error)
	UpdateRole(ctx context.Context, id uint, roleData *models.Role, permissionIDs []uint) (*models.Role, error)
	DeleteRole(ctx context.Context, id uint) error
	// TODO: Add methods for permission management on roles if needed
}

type RoleServiceImpl struct {
	roleRepo repository.RoleRepository
	logger   *zap.Logger
}

func NewRoleService(rr repository.RoleRepository, logger *zap.Logger) *RoleServiceImpl {
	return &RoleServiceImpl{
		roleRepo: rr,
		logger:   logger,
	}
}

// --- Implement RoleService methods ---

func (s *RoleServiceImpl) CreateRole(ctx context.Context, roleData *models.Role, permissionIDs []uint) (*models.Role, error) {
	s.logger.Info("RoleService: CreateRole called", zap.String("name", roleData.Name))
	// 检查重名
	roles, _, err := s.roleRepo.ListRoles(ctx, models.RoleListParams{Name: roleData.Name, Page: 1, PageSize: 1})
	if err == nil && len(roles) > 0 {
		return nil, fmt.Errorf("role name '%s' already exists: %w", roleData.Name, utils.ErrAlreadyExists)
	}
	role, err := s.roleRepo.CreateRole(ctx, roleData)
	if err != nil {
		s.logger.Error("CreateRole: repo error", zap.Error(err))
		return nil, err
	}
	// TODO: 权限分配逻辑
	return role, nil
}

func (s *RoleServiceImpl) GetRoles(ctx context.Context, params models.RoleListParams) ([]models.Role, int64, error) {
	s.logger.Info("RoleService: GetRoles called", zap.Any("params", params))
	return s.roleRepo.ListRoles(ctx, params)
}

func (s *RoleServiceImpl) GetRoleByID(ctx context.Context, id uint) (*models.RoleDetails, error) {
	s.logger.Info("RoleService: GetRoleByID called", zap.Uint("id", id))
	role, err := s.roleRepo.GetRoleByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	// TODO: 查询 userCount 和 permissions
	return &models.RoleDetails{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
		UserCount:   0,   // TODO
		Permissions: nil, // TODO
	}, nil
}

func (s *RoleServiceImpl) UpdateRole(ctx context.Context, id uint, roleData *models.Role, permissionIDs []uint) (*models.Role, error) {
	s.logger.Info("RoleService: UpdateRole called", zap.Uint("id", id), zap.String("newName", roleData.Name))
	// 检查重名
	roles, _, err := s.roleRepo.ListRoles(ctx, models.RoleListParams{Name: roleData.Name, Page: 1, PageSize: 1})
	if err == nil && len(roles) > 0 && roles[0].ID != id {
		return nil, fmt.Errorf("role name '%s' already exists: %w", roleData.Name, utils.ErrAlreadyExists)
	}
	role, err := s.roleRepo.UpdateRole(ctx, id, roleData)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		s.logger.Error("UpdateRole: repo error", zap.Error(err))
		return nil, err
	}
	// TODO: 权限分配逻辑
	return role, nil
}

func (s *RoleServiceImpl) DeleteRole(ctx context.Context, id uint) error {
	s.logger.Info("RoleService: DeleteRole called", zap.Uint("id", id))
	err := s.roleRepo.DeleteRole(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound
		}
		s.logger.Error("DeleteRole: repo error", zap.Error(err))
		return err
	}
	return nil
}

// func (s *roleService) Create(ctx context.Context, role *models.Role) (*models.Role, error) {
// 	s.logger.Info("RoleService: Create role", zap.String("name", role.Name))
// 	// return s.roleRepo.Create(ctx, role)
// 	return nil, errors.New("Create role not implemented")
// }

// func (s *roleService) GetAll(ctx context.Context) ([]models.Role, error) {
// 	s.logger.Info("RoleService: GetAll roles")
// 	// return s.roleRepo.GetAll(ctx)
// 	return nil, errors.New("GetAll roles not implemented")
// }

// Implement other methods (GetByID, Update, Delete) similarly
