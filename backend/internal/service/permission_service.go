package service

import (
	"context"
	"fmt"

	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/repository"

	"go.uber.org/zap"
)

type PermissionService interface {
	CreatePermission(ctx context.Context, permission *model.Permission) (*model.Permission, error)
	GetPermissions(ctx context.Context, params model.PermissionListParams) ([]model.Permission, int64, error)
	GetPermissionByID(ctx context.Context, id uint) (*model.Permission, error)
	UpdatePermission(ctx context.Context, id uint, permission *model.Permission) (*model.Permission, error)
	DeletePermission(ctx context.Context, id uint) error
	AddPermissionsToRole(ctx context.Context, roleID uint, permissionIDs []uint) error
	RemovePermissionsFromRole(ctx context.Context, roleID uint, permissionIDs []uint) error
	GetPermissionsByRoleID(ctx context.Context, roleID uint) ([]model.Permission, error)
}

type PermissionServiceImpl struct {
	permissionRepo repository.PermissionRepository
	roleRepo       repository.RoleRepository
	logger         *zap.Logger
}

// NewPermissionService creates a new instance of PermissionService.
func NewPermissionService(permissionRepo repository.PermissionRepository, roleRepo repository.RoleRepository, logger *zap.Logger) PermissionService {
	return &PermissionServiceImpl{
		permissionRepo: permissionRepo,
		roleRepo:       roleRepo,
		logger:         logger,
	}
}

// --- Permission CRUD Methods ---

func (s *PermissionServiceImpl) CreatePermission(ctx context.Context, permission *model.Permission) (*model.Permission, error) {
	s.logger.Info("PermissionService: CreatePermission called")
	// TODO: Add validation (e.g., check for existing name/resource/action combination)
	return s.permissionRepo.CreatePermission(ctx, permission)
}

func (s *PermissionServiceImpl) GetPermissions(ctx context.Context, params model.PermissionListParams) ([]model.Permission, int64, error) {
	s.logger.Info("PermissionService: GetPermissions called", zap.Any("params", params))
	return s.permissionRepo.ListPermissions(ctx, params)
}

func (s *PermissionServiceImpl) GetPermissionByID(ctx context.Context, id uint) (*model.Permission, error) {
	s.logger.Info("PermissionService: GetPermissionByID called", zap.Uint("id", id))
	return s.permissionRepo.GetPermissionByID(ctx, id)
}

func (s *PermissionServiceImpl) UpdatePermission(ctx context.Context, id uint, permission *model.Permission) (*model.Permission, error) {
	s.logger.Info("PermissionService: UpdatePermission called", zap.Uint("id", id))
	// TODO: Add validation (e.g., check for existing name/resource/action combination if changed)
	return s.permissionRepo.UpdatePermission(ctx, id, permission)
}

func (s *PermissionServiceImpl) DeletePermission(ctx context.Context, id uint) error {
	s.logger.Info("PermissionService: DeletePermission called", zap.Uint("id", id))
	// TODO: Check if permission is associated with any roles before deleting

	err := s.permissionRepo.DeletePermission(ctx, id)
	if err != nil {
		s.logger.Error("DeletePermission: repo error", zap.Error(err))
		return err
	}

	return nil
}

// --- Role-Permission Association Methods ---

// AddPermissionsToRole adds a list of permissions to a role.
func (s *PermissionServiceImpl) AddPermissionsToRole(ctx context.Context, roleID uint, permissionIDs []uint) error {
	s.logger.Info("PermissionService: AddPermissionsToRole called", zap.Uint("roleID", roleID), zap.Any("permissionIDs", permissionIDs))

	// Check if role exists
	role, err := s.roleRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		// TODO: Differentiate not found error
		return fmt.Errorf("failed to get role %d: %w", roleID, err)
	}

	// Check if permissions exist
	permissions, err := s.permissionRepo.GetPermissionsByIDs(ctx, permissionIDs)
	if err != nil {
		return fmt.Errorf("failed to get permissions by IDs: %w", err)
	}
	// Verify all requested permissions were found
	if len(permissions) != len(permissionIDs) {
		// Find which IDs were not found
		foundIDs := make(map[uint]bool)
		for _, p := range permissions {
			foundIDs[p.ID] = true
		}
		var notFoundIDs []uint
		for _, id := range permissionIDs {
			if !foundIDs[id] {
				notFoundIDs = append(notFoundIDs, id)
			}
		}
		return fmt.Errorf("one or more permissions not found: %v", notFoundIDs)
	}

	// Associate permissions with the role using the RoleRepository method
	if err := s.roleRepo.AddPermissionsToRole(ctx, role, permissions); err != nil {
		return fmt.Errorf("failed to add permissions to role via repo: %w", err)
	}

	return nil
}

// RemovePermissionsFromRole removes a list of permissions from a role.
func (s *PermissionServiceImpl) RemovePermissionsFromRole(ctx context.Context, roleID uint, permissionIDs []uint) error {
	s.logger.Info("PermissionService: RemovePermissionsFromRole called", zap.Uint("roleID", roleID), zap.Any("permissionIDs", permissionIDs))

	// Check if role exists
	role, err := s.roleRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		// TODO: Differentiate not found error
		return fmt.Errorf("failed to get role %d: %w", roleID, err)
	}

	// Check if permissions exist (optional but good practice)
	permissions, err := s.permissionRepo.GetPermissionsByIDs(ctx, permissionIDs)
	if err != nil {
		return fmt.Errorf("failed to get permissions by IDs: %w", err)
	}
	// Note: Repository Delete association might not return error if permissionID doesn't exist
	// If stricter validation is needed (e.g., ensure all specified permissions were actually associated),
	// we might need to query existing associations first.

	// Disassociate permissions from the role using the RoleRepository method
	if err := s.roleRepo.RemovePermissionsFromRole(ctx, role, permissions); err != nil {
		return fmt.Errorf("failed to remove permissions from role via repo: %w", err)
	}

	return nil
}

// GetPermissionsByRoleID gets all permissions associated with a role.
func (s *PermissionServiceImpl) GetPermissionsByRoleID(ctx context.Context, roleID uint) ([]model.Permission, error) {
	s.logger.Info("PermissionService: GetPermissionsByRoleID called", zap.Uint("roleID", roleID))

	// Get the role with permissions preloaded using the RoleRepository method
	role, err := s.roleRepo.GetRoleWithPermissions(ctx, roleID)
	if err != nil {
		// TODO: Differentiate not found error
		return nil, fmt.Errorf("failed to get role with permissions via repo: %w", err)
	}

	// Return the preloaded permissions
	return role.Permissions, nil
}
