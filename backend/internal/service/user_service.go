package service

import (
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/repository"
	"EffiPlat/backend/pkg/utils"
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailExists       = errors.New("email already exists")
	ErrRoleNotFound      = errors.New("one or more roles not found")
	ErrUpdateFailed      = errors.New("user update failed")
	ErrDeleteFailed      = errors.New("user delete failed")
	ErrPasswordHashing   = errors.New("failed to hash password")
	ErrAssignRolesFailed = errors.New("failed to assign roles to user")
	ErrRemoveRolesFailed = errors.New("failed to remove roles from user")
	ErrInvalidRoleIDs    = errors.New("role IDs list cannot be empty for assignment or removal")
)

// UserService defines the interface for user-related business logic.
type UserService interface {
	GetUsers(ctx context.Context, params models.UserListParams) (*utils.PaginatedResult[models.User], error)
	CreateUser(ctx context.Context, name, email, password, department string, roleIDs []uint) (*models.User, error)
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	UpdateUser(ctx context.Context, id uint, name, department, status *string, roleIDs *[]uint) (*models.User, error)
	DeleteUser(ctx context.Context, id uint) error
	AssignRolesToUser(ctx context.Context, userID uint, roleIDs []uint) error
	RemoveRolesFromUser(ctx context.Context, userID uint, roleIDs []uint) error
}

// userServiceImpl implements the UserService interface.
type userServiceImpl struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new instance of UserService.
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userServiceImpl{userRepo: userRepo}
}

// GetUsers retrieves a list of users.
func (s *userServiceImpl) GetUsers(ctx context.Context, params models.UserListParams) (*utils.PaginatedResult[models.User], error) {
	// Basic validation for pagination parameters can be done here or in handler
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 { // Set a reasonable max page size
		params.PageSize = 10
	}
	return s.userRepo.FindAll(ctx, params) // Pass ctx and params struct
}

// CreateUser creates a new user.
func (s *userServiceImpl) CreateUser(ctx context.Context, name, email, password, department string, roleIDs []uint) (*models.User, error) {
	// Validate input (basic example)
	if name == "" || email == "" || password == "" {
		return nil, errors.New("name, email, and password are required")
	}
	// Check if email already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, email) // Pass ctx
	if err != nil {
		// If the error is specifically gorm.ErrRecordNotFound, it means email is NOT taken, so proceed.
		// Any other error during email check is a genuine problem.
		if !errors.Is(err, gorm.ErrRecordNotFound) { // Assuming gorm.ErrRecordNotFound is the error for not found
			return nil, fmt.Errorf("failed to check email existence: %w", err)
		}
		// If errors.Is(err, gorm.ErrRecordNotFound), then existingUser will be nil, so err can be cleared for next step.
		err = nil
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

	createdUser, err := s.userRepo.Create(ctx, user, roleIDs) // Pass ctx
	if err != nil {
		// Check for specific GORM errors or custom repo errors if needed
		if errors.Is(err, repository.ErrRepoRoleNotFound) || (errors.Unwrap(err) != nil && errors.Is(errors.Unwrap(err), repository.ErrRepoRoleNotFound)) {
			return nil, ErrRoleNotFound // Convert to service layer error
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

// GetUserByID retrieves a user by their ID.
func (s *userServiceImpl) GetUserByID(ctx context.Context, id uint) (*models.User, error) { // Added ctx to signature
	user, err := s.userRepo.FindByID(ctx, id) // Pass ctx
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateUser updates an existing user's information.
func (s *userServiceImpl) UpdateUser(ctx context.Context, id uint, name, department, status *string, roleIDs *[]uint) (*models.User, error) { // Added ctx to signature
	// First, check if the user exists (repo's Update now also does this, but good practice for service layer)
	_, err := s.userRepo.FindByID(ctx, id) // Pass ctx
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence for update: %w", err)
	}
	// if existingUser == nil { // This check is now implicitly handled by repo.Update or FindByID above
	// 	return nil, ErrUserNotFound
	// }

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
	updatedUser, err := s.userRepo.Update(ctx, id, updates, roleIDs) // Pass ctx and id directly
	if err != nil {
		if errors.Is(err, errors.New("one or more roles not found for update")) { // Match repo error
			return nil, ErrRoleNotFound
		} else if errors.Is(err, errors.New("user not found for update")) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrUpdateFailed, err)
	}

	return updatedUser, nil
}

// DeleteUser deletes a user by their ID.
func (s *userServiceImpl) DeleteUser(ctx context.Context, id uint) error { // Added ctx to signature
	err := s.userRepo.Delete(ctx, id) // Pass ctx
	if err != nil {
		if errors.Is(err, errors.New("user not found")) { // Match repo error
			return ErrUserNotFound
		}
		return fmt.Errorf("%w: %v", ErrDeleteFailed, err)
	}
	return nil
}

// AssignRolesToUser assigns a list of roles to a user.
// It replaces all existing roles of the user with the new list.
func (s *userServiceImpl) AssignRolesToUser(ctx context.Context, userID uint, roleIDs []uint) error {
	// User existence is now checked within the repository's transaction for this specific method.
	err := s.userRepo.AssignRolesToUser(ctx, userID, roleIDs)
	if err != nil {
		// Check for wrapped errors first, then direct errors if GORM Transaction wraps them.
		if errors.Is(err, repository.ErrRepoUserNotFound) || (errors.Unwrap(err) != nil && errors.Is(errors.Unwrap(err), repository.ErrRepoUserNotFound)) {
			return ErrUserNotFound // Convert to service layer error
		}
		if errors.Is(err, repository.ErrRepoRoleNotFound) || (errors.Unwrap(err) != nil && errors.Is(errors.Unwrap(err), repository.ErrRepoRoleNotFound)) {
			return ErrRoleNotFound // Convert to service layer error
		}
		// For other errors, wrap them as ErrAssignRolesFailed
		return fmt.Errorf("%w: %v", ErrAssignRolesFailed, err)
	}
	return nil
}

// RemoveRolesFromUser removes a list of roles from a user.
func (s *userServiceImpl) RemoveRolesFromUser(ctx context.Context, userID uint, roleIDs []uint) error {
	if len(roleIDs) == 0 {
		// If roleIDs is empty, it's a no-op, which is successful.
		return nil
	}

	err := s.userRepo.RemoveRolesFromUser(ctx, userID, roleIDs)
	if err != nil {
		// Log the raw error from repository layer for diagnosis
		fmt.Printf("DEBUG: Raw error from userRepo.RemoveRolesFromUser: %[1]T - %#[2]v\n", err, err)
		if errors.Unwrap(err) != nil {
			fmt.Printf("DEBUG: Unwrapped error: %[1]T - %#[2]v\n", errors.Unwrap(err), errors.Unwrap(err))
		}

		// Check for wrapped errors first, then direct errors if GORM Transaction wraps them.
		if errors.Is(err, repository.ErrRepoUserNotFound) || (errors.Unwrap(err) != nil && errors.Is(errors.Unwrap(err), repository.ErrRepoUserNotFound)) {
			return ErrUserNotFound // Convert to service layer error
		}
		if errors.Is(err, repository.ErrRepoRoleNotFound) || (errors.Unwrap(err) != nil && errors.Is(errors.Unwrap(err), repository.ErrRepoRoleNotFound)) {
			return ErrRoleNotFound // Convert to service layer error
		}
		return fmt.Errorf("%w: %v", ErrRemoveRolesFailed, err)
	}
	return nil
}
