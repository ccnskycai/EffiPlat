package factories

import (
	"EffiPlat/backend/internal/models"
	"fmt"

	"gorm.io/gorm"
)

// RoleFactory helps create Role instances for seeding/testing.
// For simpler direct creation, a CreateRole function is also provided.
type RoleFactory struct {
	Name        string
	Description string
	Permissions []models.Permission // Optional: for creating roles with predefined permissions
}

// NewRoleFactory creates a RoleFactory with default values.
func NewRoleFactory() *RoleFactory {
	return &RoleFactory{
		Name:        "Test Role",
		Description: "A role for testing purposes",
		Permissions: []models.Permission{},
	}
}

// WithName sets a custom name for the role.
func (f *RoleFactory) WithName(name string) *RoleFactory {
	f.Name = name
	return f
}

// WithDescription sets a custom description for the role.
func (f *RoleFactory) WithDescription(desc string) *RoleFactory {
	f.Description = desc
	return f
}

// WithPermissions sets permissions for the role.
func (f *RoleFactory) WithPermissions(permissions []models.Permission) *RoleFactory {
	f.Permissions = permissions
	return f
}

// Create builds and saves the Role model to the database.
func (f *RoleFactory) Create(db *gorm.DB) (*models.Role, error) {
	role := &models.Role{
		Name:        f.Name,
		Description: f.Description,
		// CreatedAt and UpdatedAt are handled by default values or GORM hooks
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(role).Error; err != nil {
			return fmt.Errorf("failed to create role: %w", err)
		}
		if len(f.Permissions) > 0 {
			if err := tx.Model(role).Association("Permissions").Append(f.Permissions); err != nil {
				return fmt.Errorf("failed to assign permissions to role: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return role, nil
}

// CreateRole is a helper function to quickly create and save a role.
// It takes role details (Name is required, Description is optional).
// Permissions in roleDetails.Permissions will be associated.
func CreateRole(db *gorm.DB, roleDetails *models.Role) (*models.Role, error) {
	if roleDetails.Name == "" {
		return nil, fmt.Errorf("role name cannot be empty")
	}

	factory := NewRoleFactory().
		WithName(roleDetails.Name).
		WithDescription(roleDetails.Description)

	if len(roleDetails.Permissions) > 0 {
		factory.WithPermissions(roleDetails.Permissions)
	}

	return factory.Create(db)
}
