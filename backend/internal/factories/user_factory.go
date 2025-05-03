package factories

import (
	"EffiPlat/backend/internal/models"
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

// Create builds and saves the User model to the database.
func (f *UserFactory) Create(db *gorm.DB) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(f.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Name:         f.Name,
		Email:        f.Email,
		Department:   f.Department,
		PasswordHash: string(hashedPassword),
		Status:       f.Status,
		// CreatedAt and UpdatedAt are handled by default values or GORM hooks
	}

	result := db.Create(user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %w", result.Error)
	}
	return user, nil
}
