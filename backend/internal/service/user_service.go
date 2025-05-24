package service

import (
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/repository"
	"EffiPlat/backend/internal/utils"
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Removed most specific error variables, will use wrapped apputils errors instead.
var (
	// ErrUserNotFound      = errors.New("user not found")
	// ErrEmailExists       = errors.New("email already exists") // Already handled
	// ErrRoleNotFound      = errors.New("one or more roles not found")
	// ErrUpdateFailed      = errors.New("user update failed") // Already handled
	// ErrDeleteFailed      = errors.New("user delete failed") // Already handled
	ErrPasswordHashing = errors.New("failed to hash password") // Keeping this specific one for now
	// ErrAssignRolesFailed = errors.New("failed to assign roles to user")
	// ErrRemoveRolesFailed = errors.New("failed to remove roles from user")
	// ErrInvalidRoleIDs    = errors.New("role IDs list cannot be empty for assignment or removal")
)

// UserService defines the interface for user-related business logic.
type UserService interface {
	GetUsers(ctx context.Context, params model.UserListParams) (*utils.PaginatedResult[model.User], error)
	CreateUser(ctx context.Context, name, email, password, department string, roleIDs []uint) (*model.User, error)
	GetUserByID(ctx context.Context, id uint) (*model.User, error)
	UpdateUser(ctx context.Context, id uint, name, department, status *string, roleIDs *[]uint) (*model.User, error)
	DeleteUser(ctx context.Context, id uint) error
	AssignRolesToUser(ctx context.Context, userID uint, roleIDs []uint) error
	RemoveRolesFromUser(ctx context.Context, userID uint, roleIDs []uint) error
}

// userServiceImpl implements the UserService interface.
type userServiceImpl struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
	logger   *zap.Logger
}

// NewUserService creates a new instance of UserService.
func NewUserService(ur repository.UserRepository, rr repository.RoleRepository, logger *zap.Logger) UserService {
	return &userServiceImpl{
		userRepo: ur,
		roleRepo: rr,
		logger:   logger,
	}
}

// GetUsers retrieves a list of users.
func (s *userServiceImpl) GetUsers(ctx context.Context, params model.UserListParams) (*utils.PaginatedResult[model.User], error) {
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
func (s *userServiceImpl) CreateUser(ctx context.Context, name, email, password, department string, roleIDs []uint) (*model.User, error) {
	// Validate input (basic example)
	if name == "" || email == "" || password == "" {
		return nil, fmt.Errorf("name, email, and password are required: %w", utils.ErrBadRequest)
	}
	// Check if email already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, email) // Pass ctx
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check email existence: %w", err)
		}
		err = nil
	}
	if existingUser != nil {
		return nil, fmt.Errorf("email '%s' already exists: %w", email, utils.ErrAlreadyExists)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPasswordHashing, err)
	}

	user := &model.User{
		Name:       name,
		Email:      email,
		Password:   string(hashedPassword),
		Department: department,
		Status:     "active", // Default status
	}

	createdUser, err := s.userRepo.Create(ctx, user, roleIDs) // Pass ctx
	if err != nil {
		if errors.Is(err, repository.ErrRepoRoleNotFound) || (errors.Unwrap(err) != nil && errors.Is(errors.Unwrap(err), repository.ErrRepoRoleNotFound)) {
			return nil, fmt.Errorf("one or more roles not found during user creation: %w", utils.ErrNotFound) // Use wrapped ErrNotFound
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

// GetUserByID retrieves a user by their ID.
func (s *userServiceImpl) GetUserByID(ctx context.Context, id uint) (*model.User, error) { // Added ctx to signature
	user, err := s.userRepo.FindByID(ctx, id) // Pass ctx
	if err != nil {
		// Assuming FindByID returns gorm.ErrRecordNotFound or a wrapped version of it (like ErrRepoUserNotFound)
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, repository.ErrRepoUserNotFound) {
			return nil, fmt.Errorf("user with id %d not found: %w", id, utils.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to find user by ID %d: %w", id, err)
	}
	if user == nil { // Should ideally be covered by err check above if repo returns error for nil user
		return nil, fmt.Errorf("user with id %d not found: %w", id, utils.ErrNotFound)
	}
	return user, nil
}

// UpdateUser updates an existing user's information.
func (s *userServiceImpl) UpdateUser(ctx context.Context, id uint, name, department, status *string, roleIDs *[]uint) (*model.User, error) {
	// ... (user existence check can remain or be removed if repo.Update handles it robustly)
	updates := make(map[string]interface{})
	if name != nil && *name != "" {
		updates["name"] = *name
	}
	if department != nil {
		updates["department"] = *department
	}
	if status != nil && *status != "" {
		updates["status"] = *status
	}

	updatedUser, err := s.userRepo.Update(ctx, id, updates, roleIDs)
	if err != nil {
		if errors.Is(err, repository.ErrRepoRoleNotFound) {
			return nil, fmt.Errorf("one or more roles not found during user update: %w", utils.ErrNotFound)
		} else if errors.Is(err, repository.ErrRepoUserNotFound) {
			return nil, fmt.Errorf("user with id %d not found for update: %w", id, utils.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to update user %d: %w", id, utils.ErrUpdateFailed)
	}

	return updatedUser, nil
}

// DeleteUser deletes a user by their ID.
func (s *userServiceImpl) DeleteUser(ctx context.Context, id uint) error {
	err := s.userRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRepoUserNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user with id %d not found for deletion: %w", id, utils.ErrNotFound)
		}
		return fmt.Errorf("failed to delete user %d: %w", id, utils.ErrDeleteFailed)
	}
	return nil
}

// AssignRolesToUser assigns a list of roles to a user.
func (s *userServiceImpl) AssignRolesToUser(ctx context.Context, userID uint, roleIDs []uint) error {
	// Check if user exists first
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, repository.ErrRepoUserNotFound) {
			return fmt.Errorf("user with id %d not found for role assignment: %w", userID, utils.ErrNotFound)
		}
		return fmt.Errorf("failed to verify user existence for role assignment (user_id: %d): %w", userID, utils.ErrInternalServer)
	}
	if user == nil { // Explicitly check if user object is nil even if error is nil
		return fmt.Errorf("user with id %d not found for role assignment (user is nil): %w", userID, utils.ErrNotFound)
	}

	if len(roleIDs) == 0 {
		// Allow assigning empty list as a successful no-op, aligns with test expectations
		return nil
	}

	// Validate all roleIDs exist before attempting to assign
	s.logger.Debug("UserService.AssignRolesToUser: Validating role IDs before calling repository", zap.Any("roleIDs", roleIDs))
	for _, roleID := range roleIDs {
		role, err := s.roleRepo.GetRoleByID(ctx, roleID)
		s.logger.Debug("UserService.AssignRolesToUser: Role lookup result",
			zap.Uint("roleID", roleID),
			zap.Any("role_obj_exists", role != nil),
			zap.Error(err),
			zap.Bool("is_gorm_ErrRecordNotFound", errors.Is(err, gorm.ErrRecordNotFound)),
		)

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.logger.Warn("UserService.AssignRolesToUser: Non-existent role ID found", zap.Uint("roleID", roleID))
				return fmt.Errorf("one or more specified role IDs do not exist (role ID %d not found): %w", roleID, utils.ErrBadRequest)
			}
			s.logger.Error("UserService.AssignRolesToUser: Error verifying role existence", zap.Uint("roleID", roleID), zap.Error(err))
			return fmt.Errorf("failed to verify role with id %d for assignment: %w", roleID, utils.ErrInternalServer)
		}
		if role == nil {
			s.logger.Error("UserService.AssignRolesToUser: Role object is nil but no error from GetRoleByID (unexpected state)", zap.Uint("roleID", roleID))
			return fmt.Errorf("one or more specified role IDs do not exist (role ID %d not found, nil role object with no error): %w", roleID, utils.ErrBadRequest)
		}
	}

	s.logger.Debug("UserService.AssignRolesToUser: Role ID validation complete. Calling userRepo.AssignRolesToUser.")
	err = s.userRepo.AssignRolesToUser(ctx, userID, roleIDs)
	s.logger.Debug("UserService.AssignRolesToUser: Received error from userRepo.AssignRolesToUser", zap.Error(err))
	if err != nil {
		// User existence already checked. Role existence also checked.
		// If userRepo.AssignRolesToUser still returns ErrRepoUserNotFound or ErrRepoRoleNotFound,
		// it could indicate a deeper issue (race condition, transactional problem in repo, etc.)
		if errors.Is(err, repository.ErrRepoUserNotFound) {
			return fmt.Errorf("user with id %d became unavailable during role assignment: %w", userID, utils.ErrNotFound) // More specific error
		}
		if errors.Is(err, repository.ErrRepoRoleNotFound) {
			// This should ideally not be hit if pre-validation is correct, but kept for robustness
			return fmt.Errorf("one or more roles became unavailable for assignment to user %d: %w", userID, utils.ErrBadRequest)
		}
		return fmt.Errorf("failed to assign roles to user %d: %w", userID, utils.ErrInternalServer)
	}
	return nil
}

// RemoveRolesFromUser removes a list of roles from a user.
func (s *userServiceImpl) RemoveRolesFromUser(ctx context.Context, userID uint, roleIDs []uint) error {
	// Check if user exists first
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, repository.ErrRepoUserNotFound) {
			return fmt.Errorf("user with id %d not found for role removal: %w", userID, utils.ErrNotFound)
		}
		return fmt.Errorf("failed to verify user existence for role removal (user_id: %d): %w", userID, utils.ErrInternalServer)
	}
	if user == nil { // Explicitly check if user object is nil even if error is nil
		return fmt.Errorf("user with id %d not found for role removal (user is nil): %w", userID, utils.ErrNotFound)
	}

	if len(roleIDs) == 0 {
		return nil // Allow removing empty list as a successful no-op, aligns with test expectations.
	}

	s.logger.Debug("UserService.RemoveRolesFromUser: Validating role IDs before calling repository", zap.Any("roleIDs", roleIDs))
	// Validate all roleIDs exist before attempting to remove
	// This is important because the repository's RemoveRolesFromUser might not error out if a roleID is non-existent,
	// but the test expects a BadRequest in such cases.

	// First, check if all roles exist
	for _, roleID := range roleIDs {
		role, err := s.roleRepo.GetRoleByID(ctx, roleID)
		s.logger.Debug("UserService.RemoveRolesFromUser: Role lookup result",
			zap.Uint("roleID", roleID),
			zap.Any("role_obj_exists", role != nil), // Log if role object is non-nil
			zap.Error(err),
			zap.Bool("is_gorm_ErrRecordNotFound", errors.Is(err, gorm.ErrRecordNotFound)),
		)

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.logger.Warn("UserService.RemoveRolesFromUser: Non-existent role ID found", zap.Uint("roleID", roleID))
				return fmt.Errorf("one or more specified role IDs do not exist (role ID %d not found): %w", roleID, utils.ErrBadRequest)
			}
			s.logger.Error("UserService.RemoveRolesFromUser: Error verifying role existence", zap.Uint("roleID", roleID), zap.Error(err))
			return fmt.Errorf("failed to verify role with id %d for removal: %w", roleID, utils.ErrInternalServer)
		}
		if role == nil { // This case should ideally not be hit if err != nil handles ErrRecordNotFound correctly
			s.logger.Error("UserService.RemoveRolesFromUser: Role object is nil but no error from GetRoleByID (unexpected state)", zap.Uint("roleID", roleID))
			return fmt.Errorf("one or more specified role IDs do not exist (role ID %d not found, nil role object with no error): %w", roleID, utils.ErrBadRequest)
		}
	}

	// Before calling the repository layer, we've already verified all role IDs exist
	// Now we can safely call the repository method
	s.logger.Debug("UserService.RemoveRolesFromUser: Role ID validation complete. Calling userRepo.RemoveRolesFromUser.")
	err = s.userRepo.RemoveRolesFromUser(ctx, userID, roleIDs)
	s.logger.Debug("UserService.RemoveRolesFromUser: Received error from userRepo.RemoveRolesFromUser", zap.Error(err))
	if err != nil {
		// User and Role existence already checked.
		if errors.Is(err, repository.ErrRepoUserNotFound) {
			return fmt.Errorf("user with id %d became unavailable during role removal: %w", userID, utils.ErrNotFound)
		}
		if errors.Is(err, repository.ErrRepoRoleNotFound) {
			// If the repository layer returns a role not found error, we should return a clear error message
			return fmt.Errorf("one or more specified role IDs do not exist: %w", utils.ErrBadRequest)
		}
		return fmt.Errorf("failed to remove roles from user %d: %w", userID, utils.ErrInternalServer)
	}
	return nil
}
