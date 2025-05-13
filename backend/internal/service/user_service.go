package service

import (
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/repository"
	"EffiPlat/backend/pkg/utils"
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailExists       = errors.New("email already exists")
	ErrRoleNotFound      = errors.New("one or more roles not found")
	ErrUpdateFailed      = errors.New("user update failed")
	ErrDeleteFailed      = errors.New("user delete failed")
	ErrPasswordHashing = errors.New("failed to hash password")
)

// UserService defines the interface for user-related business logic.
type UserService interface {
	GetUsers(params map[string]string, page, pageSize int) (*utils.PaginatedResult[models.User], error)
	CreateUser(ctx context.Context, name, email, password, department string, roleIDs []uint) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	UpdateUser(id uint, name, department, status *string, roleIDs *[]uint) (*models.User, error)
	DeleteUser(id uint) error
}

// userServiceImpl implements the UserService interface.
type userServiceImpl struct {
	userRepo *repository.UserRepository
}

// NewUserService creates a new instance of UserService.
func NewUserService(userRepo *repository.UserRepository) UserService {
	return &userServiceImpl{userRepo: userRepo}
}

// GetUsers retrieves a list of users.
func (s *userServiceImpl) GetUsers(params map[string]string, page, pageSize int) (*utils.PaginatedResult[models.User], error) {
	// Basic validation for pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 { // Set a reasonable max page size
		pageSize = 10
	}
	return s.userRepo.FindAll(params, page, pageSize)
}

// CreateUser creates a new user.
func (s *userServiceImpl) CreateUser(ctx context.Context, name, email, password, department string, roleIDs []uint) (*models.User, error) {
	// Validate input (basic example)
	if name == "" || email == "" || password == "" {
		return nil, errors.New("name, email, and password are required")
	}
	// Check if email already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPasswordHashing, err)
	}

	user := &models.User{
		Name:       name,
		Email:      email,
		Password:   string(hashedPassword),
		Department: department,
		Status:     "active", // Default status
	}

	createdUser, err := s.userRepo.Create(user, roleIDs)
	if err != nil {
		// Check for specific GORM errors or custom repo errors if needed
		if errors.Is(err, errors.New("one or more roles not found")) { // Match repo error
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

// GetUserByID retrieves a user by their ID.
func (s *userServiceImpl) GetUserByID(id uint) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateUser updates an existing user's information.
func (s *userServiceImpl) UpdateUser(id uint, name, department, status *string, roleIDs *[]uint) (*models.User, error) {
	// First, check if the user exists
	existingUser, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence for update: %w", err)
	}
	if existingUser == nil {
		return nil, ErrUserNotFound
	}

	// Build the updates map dynamically
	updates := make(map[string]interface{})
	if name != nil && *name != "" {
		updates["name"] = *name
	}
	if department != nil { // Allow setting empty department
		updates["department"] = *department
	}
	if status != nil && *status != "" {
		// Add validation for allowed status values if necessary
		updates["status"] = *status
	}

	// Update the user record and optionally roles
	updatedUser, err := s.userRepo.Update(existingUser, updates, roleIDs)
	if err != nil {
		if errors.Is(err, errors.New("one or more roles not found for update")) { // Match repo error
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrUpdateFailed, err)
	}

	return updatedUser, nil
}

// DeleteUser deletes a user by their ID.
func (s *userServiceImpl) DeleteUser(id uint) error {
	err := s.userRepo.Delete(id)
	if err != nil {
		if errors.Is(err, errors.New("user not found")) { // Match repo error
			return ErrUserNotFound
		}
		return fmt.Errorf("%w: %v", ErrDeleteFailed, err)
	}
	return nil
} 