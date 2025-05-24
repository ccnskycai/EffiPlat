package repository

import (
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/utils"
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Repository-specific errors
var (
	ErrRepoUserNotFound = errors.New("repository: user not found")
	ErrRepoRoleNotFound = errors.New("repository: one or more roles not found")
	// Add other repo-specific errors if they arise and need to be distinct
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindAll(ctx context.Context, params model.UserListParams) (*utils.PaginatedResult[model.User], error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
	Create(ctx context.Context, user *model.User, roleIDs []uint) (*model.User, error)
	Update(ctx context.Context, userID uint, updates map[string]interface{}, roleIDs *[]uint) (*model.User, error)
	Delete(ctx context.Context, id uint) error
	FindRoleByID(ctx context.Context, id uint) (*model.Role, error)
	AssignRolesToUser(ctx context.Context, userID uint, roleIDs []uint) error
	RemoveRolesFromUser(ctx context.Context, userID uint, roleIDs []uint) error
}

// UserRepositoryImpl implements the UserRepository interface.
// Renamed from userRepositoryImpl to UserRepositoryImpl for export.
type UserRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *gorm.DB, logger *zap.Logger) UserRepository {
	return &UserRepositoryImpl{db: db, logger: logger}
}

// FindByEmail retrieves a user by their email.
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).Preload("Roles").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("User not found by email", zap.String("email", email))
			return nil, err
		}
		r.logger.Error("Failed to find user by email", zap.String("email", email), zap.Error(err))
		return nil, err
	}
	return &user, nil
}

// FindAll retrieves a paginated list of users based on filters.
// Note: The original FindAll took params map[string]string, page, pageSize.
// For consistency with other ListParams, it's better to use a struct like model.UserListParams.
// I'll assume model.UserListParams exists or will be created.
func (r *UserRepositoryImpl) FindAll(ctx context.Context, params model.UserListParams) (*utils.PaginatedResult[model.User], error) {
	var users []model.User
	var total int64

	query := r.db.WithContext(ctx).Model(&model.User{})

	if params.Name != "" {
		query = query.Where("name LIKE ?", "%"+params.Name+"%")
	}
	if params.Email != "" {
		query = query.Where("email = ?", params.Email)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Default sorting if not provided in params
	sortBy := "created_at"
	order := "desc"
	if params.SortBy != "" {
		sortBy = params.SortBy
	}
	if params.Order != "" {
		order = params.Order
	}
	query = query.Order(sortBy + " " + order)

	offset := (params.Page - 1) * params.PageSize
	err := query.Preload("Roles").Offset(offset).Limit(params.PageSize).Find(&users).Error
	if err != nil {
		return nil, err
	}

	return &utils.PaginatedResult[model.User]{
		Items:    users,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// FindByID retrieves a user by their ID, including associated roles.
func (r *UserRepositoryImpl) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Preload("Roles").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create inserts a new user record into the database.
// It handles assigning roles within a transaction.
func (r *UserRepositoryImpl) Create(ctx context.Context, user *model.User, roleIDs []uint) (*model.User, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		if len(roleIDs) > 0 {
			var roles []model.Role
			if err := tx.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
				return err
			}
			if len(roles) != len(roleIDs) {
				return ErrRepoRoleNotFound // Use predefined repository error
			}
			if err := tx.Model(user).Association("Roles").Append(&roles); err != nil {
				return err
			}
		}
		return tx.Preload("Roles").First(user, user.ID).Error
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update modifies an existing user record.
// It handles updating role associations within a transaction.
// The 'updates' map should contain fields of model.User to be updated.
// If 'roleIDs' is not nil, user's roles will be replaced with the new set.
func (r *UserRepositoryImpl) Update(ctx context.Context, userID uint, updates map[string]interface{}, roleIDs *[]uint) (*model.User, error) {
	var user model.User
	user.ID = userID

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&user, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRepoUserNotFound // Use defined error
			}
			return err
		}

		if len(updates) > 0 {
			if err := tx.Model(&user).Updates(updates).Error; err != nil {
				return err
			}
		}

		if roleIDs != nil {
			var rolesToAssign []model.Role
			if len(*roleIDs) > 0 {
				if err := tx.Where("id IN ?", *roleIDs).Find(&rolesToAssign).Error; err != nil {
					return err
				}
				if len(rolesToAssign) != len(*roleIDs) {
					return ErrRepoRoleNotFound // Use defined error
				}
			}
			if err := tx.Model(&user).Association("Roles").Replace(&rolesToAssign); err != nil {
				return err
			}
		}
		return tx.Preload("Roles").First(&user, userID).Error
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Delete performs a hard delete of a user by ID.
// It also clears associations in a transaction.
func (r *UserRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.Preload("Roles").First(&user, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRepoUserNotFound // Use defined error
			}
			return err
		}

		if err := tx.Model(&user).Association("Roles").Clear(); err != nil {
			return err
		}

		result := tx.Unscoped().Delete(&model.User{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrRepoUserNotFound // Use defined error (covers "not found or already deleted")
		}
		return nil
	})
}

// FindRoleByID checks if a role exists by ID.
// This method might be more appropriately placed in RoleRepository.
func (r *UserRepositoryImpl) FindRoleByID(ctx context.Context, id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// AssignRolesToUser replaces all roles for a user with the given roleIDs.
func (r *UserRepositoryImpl) AssignRolesToUser(ctx context.Context, userID uint, roleIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user := model.User{ID: userID}
		// First, check if the user exists.
		if err := tx.First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				r.logger.Warn("User not found for role assignment", zap.Uint("userID", userID))
				return ErrRepoUserNotFound
			}
			r.logger.Error("Failed to fetch user for role assignment", zap.Uint("userID", userID), zap.Error(err))
			return err
		}

		// Validate all the roles exist - this is a secondary check after service layer validation
		var roles []model.Role
		if len(roleIDs) > 0 { // Important check; if roleIDs is empty, Find would get all roles
			if err := tx.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
				r.logger.Error("Failed to query roles for assignment", zap.Any("roleIDs", roleIDs), zap.Error(err))
				return err // Database error during role query
			}

			// Strict check: Ensure all requested role IDs exist
			if len(roles) != len(roleIDs) {
				// This means some of the provided roleIDs don't correspond to actual roles
				r.logger.Debug("UserRepository.AssignRolesToUser: Condition len(roles) != len(roleIDs) is TRUE. Returning ErrRepoRoleNotFound",
					zap.Int("rolesLen", len(roles)),
					zap.Int("roleIDsLen", len(roleIDs)))
				r.logger.Warn("One or more roles specified for assignment not found",
					zap.Uint("userID", userID),
					zap.Any("requestedRoleIDs", roleIDs),
					zap.Int("foundRolesForAssignment", len(roles)))

				// Find which role IDs don't exist
				foundRoleIDs := make(map[uint]bool)
				for _, role := range roles {
					foundRoleIDs[role.ID] = true
				}

				for _, requestedID := range roleIDs {
					if !foundRoleIDs[requestedID] {
						r.logger.Warn("Specific role ID not found", zap.Uint("missingRoleID", requestedID))
					}
				}

				return ErrRepoRoleNotFound
			}
		} else {
			// Empty roleIDs means clear all roles
			r.logger.Info("Empty role IDs provided, clearing all roles for user", zap.Uint("userID", userID))
		}

		// Use 'Replace' to replace all roles for the user in one operation
		r.logger.Debug("UserRepository.AssignRolesToUser: Using Replace Association method",
			zap.Uint("userID", userID),
			zap.Any("roleIDs", roleIDs))

		// Use Replace instead of Clear+Append to handle the operation in one step
		if err := tx.Model(&user).Association("Roles").Replace(&roles); err != nil {
			r.logger.Error("Failed to replace roles for user", zap.Uint("userID", userID), zap.Any("roleIDs", roleIDs), zap.Error(err))
			return err
		}

		r.logger.Debug("UserRepository.AssignRolesToUser: Transaction completed successfully. Returning nil.")
		r.logger.Info("Successfully assigned roles to user in transaction", zap.Uint("userID", userID), zap.Any("assignedRoleIDs", roleIDs))
		return nil
	})
}

// RemoveRolesFromUser removes specified roles from a user.
func (r *UserRepositoryImpl) RemoveRolesFromUser(ctx context.Context, userID uint, roleIDs []uint) error {
	// If roleIDs is empty, it means no roles to remove. This can be treated as a success or a bad request.
	// The service layer currently has a check for empty roleIDs and returns ErrInvalidRoleIDs.
	// If this check is removed from the service, this repo method would attempt to remove nothing, which is fine.
	// For now, we assume roleIDs is not empty due to service layer validation.

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user := model.User{ID: userID}
		// First, check if the user exists.
		if err := tx.First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				r.logger.Warn("User not found for role removal", zap.Uint("userID", userID))
				return ErrRepoUserNotFound
			}
			r.logger.Error("Failed to fetch user for role removal", zap.Uint("userID", userID), zap.Error(err))
			return err
		}

		// 验证所有要移除的角色ID是否实际存在于数据库中
		// 这可以防止尝试移除不存在的角色
		var rolesToRemove []model.Role
		if len(roleIDs) > 0 { // 这个检查很重要；如果roleIDs为空，Find将获取所有角色
			if err := tx.Where("id IN ?", roleIDs).Find(&rolesToRemove).Error; err != nil {
				r.logger.Error("Failed to query roles for removal", zap.Any("roleIDs", roleIDs), zap.Error(err))
				return err // 角色查询期间的数据库错误
			}

			// 严格检查：确保所有请求的角色ID都存在
			if len(rolesToRemove) != len(roleIDs) {
				// 这意味着提供的roleIDs中有些不对应实际角色
				r.logger.Debug("UserRepository.RemoveRolesFromUser: 条件 len(rolesToRemove) != len(roleIDs) 为TRUE。返回 ErrRepoRoleNotFound",
					zap.Int("rolesToRemoveLen", len(rolesToRemove)),
					zap.Int("roleIDsLen", len(roleIDs)))
				r.logger.Warn("One or more roles specified for removal not found",
					zap.Uint("userID", userID),
					zap.Any("requestedRoleIDs", roleIDs),
					zap.Int("foundRolesForRemoval", len(rolesToRemove)))

				// 找出哪些角色ID不存在
				foundRoleIDs := make(map[uint]bool)
				for _, role := range rolesToRemove {
					foundRoleIDs[role.ID] = true
				}

				for _, requestedID := range roleIDs {
					if !foundRoleIDs[requestedID] {
						r.logger.Warn("Specific role ID not found", zap.Uint("missingRoleID", requestedID))
					}
				}

				return ErrRepoRoleNotFound // 或更具体的错误，如"ErrAttemptingToRemoveNonExistentRole"
			}
		} else {
			// If roleIDs is empty, service layer should have caught this. If it reaches here, it implies removing no roles.
			r.logger.Info("No role IDs provided for removal from user", zap.Uint("userID", userID))
			r.logger.Debug("UserRepository.RemoveRolesFromUser: roleIDs is empty. Transaction completed successfully. Returning nil.")
			return nil // No operation to perform, success.
		}

		if err := tx.Model(&user).Association("Roles").Delete(&rolesToRemove); err != nil {
			r.logger.Error("Failed to remove roles from user", zap.Uint("userID", userID), zap.Any("roleIDs", roleIDs), zap.Error(err))
			return err
		}
		r.logger.Debug("UserRepository.RemoveRolesFromUser: Transaction completed successfully. Returning nil.")
		r.logger.Info("Successfully removed roles from user in transaction", zap.Uint("userID", userID), zap.Any("removedRoleIDs", roleIDs))
		return nil
	})
}
