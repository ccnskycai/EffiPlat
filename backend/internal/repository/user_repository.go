package repository

import (
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/pkg/utils"
	"context"
	"errors"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindAll retrieves a paginated list of users based on filters.
func (r *UserRepository) FindAll(params map[string]string, page, pageSize int) (*utils.PaginatedResult[models.User], error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	// Apply filters dynamically based on params map
	if name, ok := params["name"]; ok && name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if email, ok := params["email"]; ok && email != "" {
		query = query.Where("email = ?", email)
	}
	if status, ok := params["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total records before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply sorting
	sortBy := params["sortBy"]
	order := params["order"]
	if sortBy == "" {
		sortBy = "created_at" // Default sort
	}
	if order == "" {
		order = "desc" // Default order
	}
	query = query.Order(sortBy + " " + order)

	// Apply pagination
	offset := (page - 1) * pageSize
	err := query.Preload("Roles").Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		return nil, err
	}

	result := &utils.PaginatedResult[models.User]{
		Items:    users,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	return result, nil
}

// FindByID retrieves a user by their ID, including associated roles.
func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Roles").First(&user, id).Error
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
func (r *UserRepository) Create(user *models.User, roleIDs []uint) (*models.User, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		if len(roleIDs) > 0 {
			var roles []models.Role
			if err := tx.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
				return err
			}
			if len(roles) != len(roleIDs) {
				return errors.New("one or more roles not found")
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

// Update modifies an existing user record. Does not update password or email here.
// Handles updating role associations.
func (r *UserRepository) Update(user *models.User, updates map[string]interface{}, roleIDs *[]uint) (*models.User, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(user).Updates(updates).Error; err != nil {
			return err
		}
		if roleIDs != nil {
			var roles []models.Role
			if len(*roleIDs) > 0 {
				if err := tx.Where("id IN ?", *roleIDs).Find(&roles).Error; err != nil {
					return err
				}
				if len(roles) != len(*roleIDs) {
					return errors.New("one or more roles not found for update")
				}
			}
			if err := tx.Model(user).Association("Roles").Replace(&roles); err != nil {
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

// Delete performs a hard delete.
func (r *UserRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.First(&user, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found")
			}
			return err
		}
		if err := tx.Model(&user).Association("Roles").Clear(); err != nil {
			return err
		}
		// Use Select to ensure associations are handled if GORM is configured for it,
		// though explicit clearing (above) is safer for many2many.
		// For hard delete, ensure it's not trying to soft delete by default.
		result := tx.Select("Roles").Delete(&user) // GORM handles many2many deletion with .Select("Roles") or .Association("Roles").Clear() and then Delete(). Explicit is safer.
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("user not found or already deleted")
		}
		return nil
	})
}

// FindRoleByID checks if a role exists by ID.
func (r *UserRepository) FindRoleByID(id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.First(&role, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}
