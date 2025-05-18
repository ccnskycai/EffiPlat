package repository

import (
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/pkg/utils"
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
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindAll(ctx context.Context, params models.UserListParams) (*utils.PaginatedResult[models.User], error)
	FindByID(ctx context.Context, id uint) (*models.User, error)
	Create(ctx context.Context, user *models.User, roleIDs []uint) (*models.User, error)
	Update(ctx context.Context, userID uint, updates map[string]interface{}, roleIDs *[]uint) (*models.User, error)
	Delete(ctx context.Context, id uint) error
	FindRoleByID(ctx context.Context, id uint) (*models.Role, error)
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
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
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
// For consistency with other ListParams, it's better to use a struct like models.UserListParams.
// I'll assume models.UserListParams exists or will be created.
func (r *UserRepositoryImpl) FindAll(ctx context.Context, params models.UserListParams) (*utils.PaginatedResult[models.User], error) {
	var users []models.User
	var total int64

	query := r.db.WithContext(ctx).Model(&models.User{})

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

	return &utils.PaginatedResult[models.User]{
		Items:    users,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// FindByID retrieves a user by their ID, including associated roles.
func (r *UserRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Preload("Roles").First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Create inserts a new user record into the database.
// It handles assigning roles within a transaction.
func (r *UserRepositoryImpl) Create(ctx context.Context, user *models.User, roleIDs []uint) (*models.User, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		if len(roleIDs) > 0 {
			var roles []models.Role
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
// The 'updates' map should contain fields of models.User to be updated.
// If 'roleIDs' is not nil, user's roles will be replaced with the new set.
func (r *UserRepositoryImpl) Update(ctx context.Context, userID uint, updates map[string]interface{}, roleIDs *[]uint) (*models.User, error) {
	var user models.User
	user.ID = userID

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&user, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found for update")
			}
			return err
		}

		if len(updates) > 0 {
			if err := tx.Model(&user).Updates(updates).Error; err != nil {
				return err
			}
		}

		if roleIDs != nil {
			var rolesToAssign []models.Role
			if len(*roleIDs) > 0 {
				if err := tx.Where("id IN ?", *roleIDs).Find(&rolesToAssign).Error; err != nil {
					return err
				}
				if len(rolesToAssign) != len(*roleIDs) {
					return errors.New("one or more roles not found for update")
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
		var user models.User
		if err := tx.Preload("Roles").First(&user, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found")
			}
			return err
		}

		if err := tx.Model(&user).Association("Roles").Clear(); err != nil {
			return err
		}

		result := tx.Unscoped().Delete(&models.User{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("user not found or already deleted during transaction")
		}
		return nil
	})
}

// FindRoleByID checks if a role exists by ID.
// This method might be more appropriately placed in RoleRepository.
func (r *UserRepositoryImpl) FindRoleByID(ctx context.Context, id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// AssignRolesToUser replaces all roles for a user with the given roleIDs.
func (r *UserRepositoryImpl) AssignRolesToUser(ctx context.Context, userID uint, roleIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user := models.User{ID: userID}
		// First, check if the user exists.
		if err := tx.First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				r.logger.Warn("User not found for role assignment", zap.Uint("userID", userID))
				return ErrRepoUserNotFound // Use repository-defined error
			}
			r.logger.Error("Failed to fetch user for role assignment", zap.Uint("userID", userID), zap.Error(err))
			return err
		}

		var rolesToAssign []models.Role
		if len(roleIDs) > 0 {
			if err := tx.Where("id IN ?", roleIDs).Find(&rolesToAssign).Error; err != nil {
				r.logger.Error("Failed to query roles for assignment", zap.Any("roleIDs", roleIDs), zap.Error(err))
				return err // DB error during role lookup
			}
			if len(rolesToAssign) != len(roleIDs) {
				r.logger.Warn("One or more roles not found for assignment", zap.Uint("userID", userID), zap.Any("requestedRoleIDs", roleIDs), zap.Int("foundCount", len(rolesToAssign)))
				return ErrRepoRoleNotFound // Use repository-defined error
			}
		}

		// Replace will clear existing associations and add the new ones.
		// If rolesToAssign is empty (because roleIDs was empty), it will clear all roles.
		if err := tx.Model(&user).Association("Roles").Replace(&rolesToAssign); err != nil {
			r.logger.Error("Failed to replace roles for user", zap.Uint("userID", userID), zap.Any("roleIDs", roleIDs), zap.Error(err))
			return err
		}
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
		user := models.User{ID: userID}
		// First, check if the user exists.
		if err := tx.First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				r.logger.Warn("User not found for role removal", zap.Uint("userID", userID))
				return ErrRepoUserNotFound
			}
			r.logger.Error("Failed to fetch user for role removal", zap.Uint("userID", userID), zap.Error(err))
			return err
		}

		// Validate that all roleIDs to be removed actually exist in the database.
		// This prevents attempting to remove non-existent roles, though GORM's Delete might handle it gracefully.
		// However, explicit validation provides clearer error feedback.
		var rolesToRemove []models.Role
		if len(roleIDs) > 0 { // This check is important; if roleIDs is empty, Find would fetch all roles.
			if err := tx.Where("id IN ?", roleIDs).Find(&rolesToRemove).Error; err != nil {
				r.logger.Error("Failed to query roles for removal", zap.Any("roleIDs", roleIDs), zap.Error(err))
				return err // DB error during role lookup
			}
			if len(rolesToRemove) != len(roleIDs) {
				// This means some of the provided roleIDs do not correspond to actual roles.
				// Depending on strictness, this could be an error.
				r.logger.Warn("One or more roles specified for removal not found", zap.Uint("userID", userID), zap.Any("requestedRoleIDs", roleIDs), zap.Int("foundRolesForRemoval", len(rolesToRemove)))
				return ErrRepoRoleNotFound // Or a more specific error like "ErrAttemptingToRemoveNonExistentRole"
			}
		} else {
			// If roleIDs is empty, service layer should have caught this. If it reaches here, it implies removing no roles.
			r.logger.Info("No role IDs provided for removal from user", zap.Uint("userID", userID))
			return nil // No operation to perform, success.
		}

		if err := tx.Model(&user).Association("Roles").Delete(&rolesToRemove); err != nil {
			r.logger.Error("Failed to remove roles from user", zap.Uint("userID", userID), zap.Any("roleIDs", roleIDs), zap.Error(err))
			return err
		}
		r.logger.Info("Successfully removed roles from user in transaction", zap.Uint("userID", userID), zap.Any("removedRoleIDs", roleIDs))
		return nil
	})
}
