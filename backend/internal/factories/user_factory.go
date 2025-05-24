package factories

import (
	"EffiPlat/backend/internal/model"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserFactory helps create User instances for seeding/testing.
type UserFactory struct {
	Name       string
	Email      string
	Department *string
	Password   string // Plain password, will be hashed
	Status     string
	Roles      []model.Role // Added roles field
}

// NewUserFactory creates a UserFactory with default values.
func NewUserFactory() *UserFactory {
	rand.Seed(time.Now().UnixNano())
	defaultEmail := fmt.Sprintf("user_%d@example.com", rand.Intn(100000))
	return &UserFactory{
		Name:       "Test User",
		Email:      defaultEmail,
		Department: nil,
		Password:   "password",
		Status:     "active",
		Roles:      []model.Role{}, // Initialize empty roles slice
	}
}

// WithName sets a custom name for the user.
func (f *UserFactory) WithName(name string) *UserFactory {
	f.Name = name
	return f
}

// WithEmail sets a custom email for the user.
func (f *UserFactory) WithEmail(email string) *UserFactory {
	f.Email = email
	return f
}

// WithDepartment sets a custom department for the user.
func (f *UserFactory) WithDepartment(dept string) *UserFactory {
	f.Department = &dept
	return f
}

// WithPassword sets a custom plain text password.
func (f *UserFactory) WithPassword(password string) *UserFactory {
	f.Password = password
	return f
}

// WithStatus sets a custom status for the user.
func (f *UserFactory) WithStatus(status string) *UserFactory {
	f.Status = status
	return f
}

// WithRoles sets roles for the user.
func (f *UserFactory) WithRoles(roles []model.Role) *UserFactory {
	f.Roles = roles
	return f
}

// Create builds and saves the User model to the database.
func (f *UserFactory) Create(db *gorm.DB) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(f.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Name:     f.Name,
		Email:    f.Email,
		Password: string(hashedPassword),
		Status:   f.Status,
	}
	if f.Department != nil {
		user.Department = *f.Department
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		if len(f.Roles) > 0 {
			if err := tx.Model(user).Association("Roles").Append(f.Roles); err != nil {
				return fmt.Errorf("failed to assign roles: %w", err)
			}
		}
		// Reload user to get ID and potentially preloaded associations if needed by caller
		// For now, just returning the user as created.
		return nil
	})

	if err != nil {
		return nil, err
	}
	return user, nil
}

// CreateUser is a helper function to quickly create and save a user.
// It takes a user model (Password field should be plain text, it will be hashed).
// Roles defined in user.Roles will be associated.
func CreateUser(db *gorm.DB, userDetails *model.User) (*model.User, error) {
	factory := NewUserFactory().
		WithName(userDetails.Name).
		WithEmail(userDetails.Email).
		WithStatus(userDetails.Status)

	if userDetails.Password != "" { // Allow creating user without password for factory if needed, though real users need it
		factory.WithPassword(userDetails.Password)
	} else {
		factory.WithPassword("testpassword") // Default if not provided in model
	}

	if userDetails.Department != "" {
		factory.WithDepartment(userDetails.Department)
	}

	if len(userDetails.Roles) > 0 {
		factory.WithRoles(userDetails.Roles)
	}

	return factory.Create(db)
}
