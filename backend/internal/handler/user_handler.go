package handler

import (
	"EffiPlat/backend/internal/service" // Added to access UserService interface and error variables
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"EffiPlat/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// UserHandler handles HTTP requests for user management.
type UserHandler struct {
	userService service.UserService // Changed to service.UserService
	validate    *validator.Validate
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(userService service.UserService) *UserHandler { // Changed to service.UserService
	return &UserHandler{
		userService: userService,
		validate:    validator.New(),
	}
}

// Request and Response Structures (DTOs)
type CreateUserRequest struct {
	Name       string `json:"name" validate:"required,min=2,max=100"`
	Email      string `json:"email" validate:"required,email,max=100"`
	Password   string `json:"password" validate:"required,min=8,max=72"`
	Department string `json:"department,omitempty" validate:"max=100"`
	Roles      []uint `json:"roles,omitempty"` // List of role IDs
}

type UpdateUserRequest struct {
	Name       *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Department *string `json:"department,omitempty" validate:"omitempty,max=100"`
	Status     *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive pending"`
	Roles      *[]uint `json:"roles,omitempty"` // Pointer to allow explicit null vs empty array for roles update
}

// UserResponse is a generic user response, omitting sensitive data like password.
// Roles are included as per the API design.
// The actual `models.User` structure is already suitable if GORM tags don't interfere
// and password is correctly omitted (`json:"-"` tag in model).
// We can directly use models.User or define a specific DTO if more transformation is needed.
// type UserResponse struct {
// 	ID         uint          `json:"id"`
// 	Name       string        `json:"name"`
// 	Email      string        `json:"email"`
// 	Department string        `json:"department,omitempty"`
// 	Status     string        `json:"status"`
// 	CreatedAt  time.Time     `json:"createdAt"`
// 	UpdatedAt  time.Time     `json:"updatedAt"`
// 	Roles      []models.Role `json:"roles,omitempty"`
// }

// GetUsers handles GET /users request
func (h *UserHandler) GetUsers(c *gin.Context) {
	var params models.UserListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid query parameters: %v", err))
		return
	}

	// Default pagination if not provided
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	paginatedResult, err := h.userService.GetUsers(c.Request.Context(), params)
	if err != nil {
		RespondWithError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve users: %v", err))
		return
	}

	RespondWithSuccess(c, http.StatusOK, "Users retrieved successfully", paginatedResult)
}

// CreateUser handles POST /users request
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("Validation failed: %s", FormatValidationErrors(err)))
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), req.Name, req.Email, req.Password, req.Department, req.Roles)
	if err != nil {
		// Determine appropriate status code based on error type
		statusCode := http.StatusInternalServerError // Default
		errMsg := fmt.Sprintf("Failed to create user: %v", err)

		// DEBUGGING: Log the error from service layer
		fmt.Printf("DEBUG CreateUser Handler: error from userService.CreateUser: %[1]T - %#[2]v\n", err, err)
		if errors.Unwrap(err) != nil {
			fmt.Printf("DEBUG CreateUser Handler: unwrapped error: %[1]T - %#[2]v\n", errors.Unwrap(err), errors.Unwrap(err))
		}

		if errors.Is(err, service.ErrEmailExists) || errors.Is(err, service.ErrRoleNotFound) {
			statusCode = http.StatusBadRequest
		} else if errors.Is(err, service.ErrPasswordHashing) {
			statusCode = http.StatusInternalServerError
		}
		RespondWithError(c, statusCode, errMsg)
		return
	}

	RespondWithSuccess(c, http.StatusCreated, "User created successfully", user)
}

// GetUserByID handles GET /users/{userId} request
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid user ID format: %v", err))
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), uint(userID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := fmt.Sprintf("Failed to retrieve user: %v", err)
		if errors.Is(err, service.ErrUserNotFound) {
			statusCode = http.StatusNotFound
		}
		RespondWithError(c, statusCode, errMsg)
		return
	}

	claimsValue, exists := c.Get("user")
	if !exists {
		RespondWithError(c, http.StatusInternalServerError, "User claims not found in context")
		return
	}
	claims, ok := claimsValue.(*models.Claims)
	if !ok {
		RespondWithError(c, http.StatusInternalServerError, "Invalid claims format in context")
		return
	}

	if claims.UserID != uint(userID) && !hasAdminRole(claims) { // Assuming hasAdminRole helper exists
		RespondWithError(c, http.StatusForbidden, "Permission denied: User can only retrieve their own information or requires admin role")
		return
	}

	RespondWithSuccess(c, http.StatusOK, "User retrieved successfully", user)
}

// UpdateUser handles PUT /users/{userId} request
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid user ID format: %v", err))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("Validation failed: %s", FormatValidationErrors(err)))
		return
	}

	claimsValue, exists := c.Get("user")
	if !exists {
		RespondWithError(c, http.StatusInternalServerError, "User claims not found in context")
		return
	}
	claims, ok := claimsValue.(*models.Claims)
	if !ok {
		RespondWithError(c, http.StatusInternalServerError, "Invalid claims format in context")
		return
	}

	if claims.UserID != uint(userID) && !hasAdminRole(claims) {
		RespondWithError(c, http.StatusForbidden, "Permission denied: User can only update their own information or requires admin role")
		return
	}
	if !hasAdminRole(claims) {
		if req.Status != nil || req.Roles != nil {
			RespondWithError(c, http.StatusForbidden, "Permission denied: Only admins can change status or roles")
			return
		}
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), uint(userID), req.Name, req.Department, req.Status, req.Roles)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := fmt.Sprintf("Failed to update user: %v", err)
		if errors.Is(err, service.ErrUserNotFound) {
			statusCode = http.StatusNotFound
		} else if errors.Is(err, service.ErrRoleNotFound) {
			statusCode = http.StatusBadRequest
		} else if errors.Is(err, service.ErrUpdateFailed) { // Assuming ErrUpdateFailed is defined in service
			statusCode = http.StatusInternalServerError
		}
		RespondWithError(c, statusCode, errMsg)
		return
	}

	RespondWithSuccess(c, http.StatusOK, "User updated successfully", user)
}

// DeleteUser handles DELETE /users/{userId} request
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid user ID format: %v", err))
		return
	}

	claimsValue, exists := c.Get("user")
	if !exists {
		RespondWithError(c, http.StatusInternalServerError, "User claims not found in context")
		return
	}
	claims, ok := claimsValue.(*models.Claims)
	if !ok {
		RespondWithError(c, http.StatusInternalServerError, "Invalid claims format in context")
		return
	}

	// Add check: Admin can delete anyone, user can delete themselves (optional, based on policy)
	if claims.UserID != uint(userID) && !hasAdminRole(claims) {
		RespondWithError(c, http.StatusForbidden, "Permission denied: User can only delete their own account or requires admin role")
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), uint(userID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := fmt.Sprintf("Failed to delete user: %v", err)
		if errors.Is(err, service.ErrUserNotFound) { // Assuming ErrUserNotFound is defined in service
			statusCode = http.StatusNotFound
		} else if errors.Is(err, service.ErrDeleteFailed) { // Assuming ErrDeleteFailed is defined in service
			statusCode = http.StatusInternalServerError // Or perhaps a more specific error if available
		}
		RespondWithError(c, statusCode, errMsg)
		return
	}

	// RespondWithSuccess(c, http.StatusOK, "User deleted successfully", nil) // No data to return on successful delete typically
	c.Status(http.StatusNoContent) // Return 204 No Content on successful deletion
}

// hasAdminRole (placeholder - needs proper implementation based on how roles are stored/checked)
// This function should ideally be in a common utility or auth-related package if used across handlers.
func hasAdminRole(claims *models.Claims) bool {
	// This is a placeholder. In a real application, you would check
	// if claims.Roles (or similar field) contains an admin role.
	// For example:
	// for _, role := range claims.Roles { // Assuming claims has a Roles []string field
	// if role == "admin" {
	// return true
	// }
	// }
	// return false
	return true // TEMPORARY: Assume admin for now for development
}

// FormatValidationErrors formats validation errors from validator.v10 to a string.
func FormatValidationErrors(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errorMsgs []string
		for _, e := range validationErrors {
			// Example: "Field 'Name' failed on 'required' tag"
			errorMsgs = append(errorMsgs, fmt.Sprintf("Field '%s' failed on '%s' tag", e.Field(), e.Tag()))
		}
		return fmt.Sprintf("Validation errors: [%s]", string(JoinStrings(errorMsgs, ", ")))
	}
	return err.Error()
}

// JoinStrings is a helper function, you might have this in a utility package.
// If not, define it or use strings.Join.
func JoinStrings(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	if len(elems) == 1 {
		return elems[0]
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i])
	}

	var b []byte
	if n > 0 {
		b = make([]byte, 0, n)
	}
	for i, s := range elems {
		if i > 0 {
			b = append(b, sep...)
		}
		b = append(b, s...)
	}
	return string(b)
}

// AssignRolesToUser handles POST /users/{userId}/roles - Assigns/Replaces roles for a user
func (h *UserHandler) AssignRolesToUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid user ID format: %v", err))
		return
	}

	var req models.AssignRemoveRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
		return
	}

	// Basic validation for RoleIDs presence, though service layer might also check
	// For Assign, an empty list might mean "remove all roles", so service handles that logic.
	// Here, we ensure the binding itself worked for the overall structure.

	// Authorization check: Who can assign roles? (e.g., Admin only)
	claimsValue, _ := c.Get("user") // Assume middleware has set this
	claims, ok := claimsValue.(*models.Claims)
	if !ok || !hasAdminRole(claims) { // hasAdminRole is a placeholder
		RespondWithError(c, http.StatusForbidden, "Permission denied: Only admins can assign roles to users")
		return
	}

	err = h.userService.AssignRolesToUser(c.Request.Context(), uint(userID), req.RoleIDs)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := fmt.Sprintf("Failed to assign roles: %v", err)
		if errors.Is(err, service.ErrUserNotFound) {
			statusCode = http.StatusNotFound
		} else if errors.Is(err, service.ErrAssignRolesFailed) || errors.Is(err, service.ErrInvalidRoleIDs) || errors.Is(err, service.ErrRoleNotFound) {
			// ErrRoleNotFound could also come from service if it tries to validate role IDs exist
			statusCode = http.StatusBadRequest
		}
		RespondWithError(c, statusCode, errMsg)
		return
	}

	RespondWithSuccess(c, http.StatusOK, "Roles assigned successfully to user", nil)
}

// RemoveRolesFromUser handles DELETE /users/{userId}/roles - Removes specified roles from a user
func (h *UserHandler) RemoveRolesFromUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid user ID format: %v", err))
		return
	}

	var req models.AssignRemoveRolesRequest // Using the same request DTO for simplicity
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
		return
	}

	// Authorization check: Who can remove roles? (e.g., Admin only)
	claimsValue, _ := c.Get("user")
	claims, ok := claimsValue.(*models.Claims)
	if !ok || !hasAdminRole(claims) { // hasAdminRole is a placeholder
		RespondWithError(c, http.StatusForbidden, "Permission denied: Only admins can remove roles from users")
		return
	}

	err = h.userService.RemoveRolesFromUser(c.Request.Context(), uint(userID), req.RoleIDs)
	if err != nil {
		statusCode := http.StatusInternalServerError // Default
		errMsg := fmt.Sprintf("Failed to remove roles from user: %v", err)

		// DEBUGGING: Log the error from service layer in Handler
		fmt.Printf("DEBUG RemoveRoles Handler: error from userService.RemoveRolesFromUser: %[1]T - %#[2]v\n", err, err)
		if errors.Unwrap(err) != nil {
			fmt.Printf("DEBUG RemoveRoles Handler: unwrapped error: %[1]T - %#[2]v\n", errors.Unwrap(err), errors.Unwrap(err))
		}

		if errors.Is(err, service.ErrUserNotFound) {
			statusCode = http.StatusNotFound
		} else if errors.Is(err, service.ErrRoleNotFound) { // THIS IS THE KEY CHECK
			statusCode = http.StatusBadRequest
		} else if errors.Is(err, service.ErrInvalidRoleIDs) { // This was for empty roleIDs, now handled as success by service
			statusCode = http.StatusBadRequest
		}
		RespondWithError(c, statusCode, errMsg)
		return
	}

	RespondWithSuccess(c, http.StatusOK, "Roles removed successfully from user", nil)
}
