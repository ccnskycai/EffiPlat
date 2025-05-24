package handler

import (
	"EffiPlat/backend/internal/service" // Added to access UserService interface and error variables
	"EffiPlat/backend/internal/utils"   // Import apputils
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"EffiPlat/backend/internal/model"

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
// The actual `model.User` structure is already suitable if GORM tags don't interfere
// and password is correctly omitted (`json:"-"` tag in model).
// We can directly use model.User or define a specific DTO if more transformation is needed.
// type UserResponse struct {
// 	ID         uint          `json:"id"`
// 	Name       string        `json:"name"`
// 	Email      string        `json:"email"`
// 	Department string        `json:"department,omitempty"`
// 	Status     string        `json:"status"`
// 	CreatedAt  time.Time     `json:"createdAt"`
// 	UpdatedAt  time.Time     `json:"updatedAt"`
// 	Roles      []model.Role `json:"roles,omitempty"`
// }

// GetUsers handles GET /users request
func (h *UserHandler) GetUsers(c *gin.Context) {
	var params model.UserListParams
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

	// 准备审计日志信息
	c.Set("auditAction", string(utils.AuditActionCreate))
	c.Set("auditResource", "USER")

	user, err := h.userService.CreateUser(c.Request.Context(), req.Name, req.Email, req.Password, req.Department, req.Roles)
	if err != nil {
		statusCode := http.StatusInternalServerError // Default
		errMsg := fmt.Sprintf("Failed to create user: %v", err)

		// Check for specific wrapped errors from the service layer
		if errors.Is(err, utils.ErrAlreadyExists) { // Was service.ErrEmailExists
			statusCode = http.StatusBadRequest
			errMsg = err.Error() // Use the specific message from service
		} else if errors.Is(err, utils.ErrNotFound) { // Was service.ErrRoleNotFound
			statusCode = http.StatusBadRequest
			errMsg = err.Error() // Use the specific message from service
		} else if errors.Is(err, service.ErrPasswordHashing) { // This one is still defined in user_service
			statusCode = http.StatusInternalServerError
			// errMsg can remain generic or be made specific for password hashing
		} else if errors.Is(err, utils.ErrBadRequest) { // General bad request from service
			statusCode = http.StatusBadRequest
			errMsg = err.Error()
		}
		RespondWithError(c, statusCode, errMsg)
		return
	}

	// 设置详细的审计日志信息 - 不包含敏感信息如密码
	auditDetails := utils.NewCreateAuditLog(map[string]interface{}{
		"id":         user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"department": user.Department,
		"roles":      req.Roles,
	})
	utils.SetAuditDetails(c, auditDetails)
	// 设置资源ID以便中间件获取
	c.Set("auditResourceID", user.ID)

	RespondWithSuccess(c, http.StatusCreated, "User created successfully", user)
}

// GetUserByID handles GET /users/{userId} request
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	// 设置审计日志操作类型和资源
	c.Set("auditAction", string(utils.AuditActionRead))
	c.Set("auditResource", "USER")
	c.Set("auditResourceID", uint(userID))

	user, err := h.userService.GetUserByID(c.Request.Context(), uint(userID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := fmt.Sprintf("Failed to retrieve user: %v", err)

		if errors.Is(err, utils.ErrNotFound) { // Was service.ErrUserNotFound
			statusCode = http.StatusNotFound
			errMsg = fmt.Sprintf("User with ID %d not found", userID)
		}

		RespondWithError(c, statusCode, errMsg)
		return
	}

	// Note: The user object already includes roles by default from the service

	// 记录查询详情 (对于READ操作，只需记录最基本的信息)
	utils.SetAuditDetails(c, map[string]interface{}{
		"action": "READ",
		"userQueried": uint(userID),
	})

	// In real-world scenarios, you might apply additional authorization checks here
	// Check if the requesting user can access the target user's details

	// Use the user directly as response with omitted fields
	// or map to a response DTO if needed for more control
	RespondWithSuccess(c, http.StatusOK, "User retrieved successfully", user)
}

// UpdateUser handles PUT /users/{userId} request
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid user ID format")
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

	// 设置审计日志操作类型和资源
	c.Set("auditAction", string(utils.AuditActionUpdate))
	c.Set("auditResource", "USER")
	c.Set("auditResourceID", uint(userID))

	// Fetch the current user for comparison (and to verify it exists)
	existingUser, err := h.userService.GetUserByID(c.Request.Context(), uint(userID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := fmt.Sprintf("Failed to retrieve user: %v", err)

		if errors.Is(err, utils.ErrNotFound) { // Was service.ErrUserNotFound
			statusCode = http.StatusNotFound
			errMsg = fmt.Sprintf("User with ID %d not found", userID)
		}

		RespondWithError(c, statusCode, errMsg)
		return
	}

	// Optional: Track role changes separately if needed
	var roleUpdate bool
	if req.Roles != nil {
		roleUpdate = true
	}

	// 记录更新前的状态用于审计日志
	beforeUpdate := map[string]interface{}{
		"id": existingUser.ID,
		"name": existingUser.Name,
		"email": existingUser.Email,
		"department": existingUser.Department,
		"status": existingUser.Status,
	}

	// Perform the update
	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), uint(userID), req.Name, req.Department, req.Status, req.Roles)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := fmt.Sprintf("Failed to update user: %v", err)

		if errors.Is(err, utils.ErrNotFound) {
			statusCode = http.StatusNotFound
			errMsg = fmt.Sprintf("User with ID %d not found", userID)
		} else if errors.Is(err, utils.ErrBadRequest) {
			statusCode = http.StatusBadRequest
			errMsg = err.Error() // Use specific message from service
		}

		RespondWithError(c, statusCode, errMsg)
		return
	}

	// 创建审计日志信息
	afterUpdate := map[string]interface{}{
		"id": updatedUser.ID,
		"name": updatedUser.Name,
		"email": updatedUser.Email,
		"department": updatedUser.Department,
		"status": updatedUser.Status,
	}

	// 如果角色被更新，添加到审计信息中
	if roleUpdate && req.Roles != nil {
		afterUpdate["roles"] = *req.Roles
	}

	auditDetails := utils.NewUpdateAuditLog(beforeUpdate, afterUpdate)
	utils.SetAuditDetails(c, auditDetails)

	RespondWithSuccess(c, http.StatusOK, "User updated successfully", updatedUser)
}

// DeleteUser handles DELETE /users/{userId} request
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	// 设置审计日志操作类型和资源
	c.Set("auditAction", string(utils.AuditActionDelete))
	c.Set("auditResource", "USER")
	c.Set("auditResourceID", uint(userID))

	// Check if user exists before attempting to delete
	existingUser, err := h.userService.GetUserByID(c.Request.Context(), uint(userID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := fmt.Sprintf("Error checking user: %v", err)

		if errors.Is(err, utils.ErrNotFound) {
			statusCode = http.StatusNotFound
			errMsg = fmt.Sprintf("User with ID %d not found", userID)
		}

		RespondWithError(c, statusCode, errMsg)
		return
	}

	// Check if we're trying to delete our own admin account
	// This might need more sophisticated checks in real implementation
	if existingUser.ID == 1 {
		RespondWithError(c, http.StatusForbidden, "Cannot delete the primary admin account")
		return
	}

	// 记录被删除的用户信息用于审计日志
	deleteDetails := map[string]interface{}{
		"id": existingUser.ID,
		"name": existingUser.Name,
		"email": existingUser.Email,
		"department": existingUser.Department,
		"status": existingUser.Status,
	}
	utils.SetAuditDetails(c, utils.NewDeleteAuditLog(deleteDetails))

	err = h.userService.DeleteUser(c.Request.Context(), uint(userID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := fmt.Sprintf("Failed to delete user: %v", err)

		if errors.Is(err, utils.ErrNotFound) {
			statusCode = http.StatusNotFound
			errMsg = fmt.Sprintf("User with ID %d not found", userID)
		} else if errors.Is(err, utils.ErrForbidden) {
			statusCode = http.StatusForbidden
			errMsg = err.Error()
		}

		RespondWithError(c, statusCode, errMsg)
		return
	}

	RespondWithSuccess(c, http.StatusNoContent, "User deleted successfully", nil)
}

// hasAdminRole (placeholder - needs proper implementation based on how roles are stored/checked)
// This function should ideally be in a common utility or auth-related package if used across handler.
func hasAdminRole(claims *model.Claims) bool {
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
		RespondWithError(c, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	var req model.AssignRemoveRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
		return
	}

	err = h.userService.AssignRolesToUser(c.Request.Context(), uint(userID), req.RoleIDs)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := fmt.Sprintf("Failed to assign roles to user: %v", err)

		if errors.Is(err, utils.ErrNotFound) {
			statusCode = http.StatusNotFound
			errMsg = err.Error()
		} else if errors.Is(err, utils.ErrBadRequest) {
			statusCode = http.StatusBadRequest
			errMsg = err.Error()
		} else if errors.Is(err, utils.ErrInternalServer) {
			statusCode = http.StatusInternalServerError
			errMsg = err.Error()
		}
		RespondWithError(c, statusCode, errMsg)
		return
	}

	RespondWithSuccess(c, http.StatusOK, "Roles assigned successfully to user", nil)
}

// RemoveRolesFromUser handles DELETE /users/{userId}/roles request
func (h *UserHandler) RemoveRolesFromUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		RespondWithError(c, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	var req model.AssignRemoveRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
		return
	}

	err = h.userService.RemoveRolesFromUser(c.Request.Context(), uint(userID), req.RoleIDs)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := fmt.Sprintf("Failed to remove roles from user: %v", err)
		if errors.Is(err, utils.ErrNotFound) {
			statusCode = http.StatusNotFound
			errMsg = err.Error()
		} else if errors.Is(err, utils.ErrInternalServer) {
			statusCode = http.StatusInternalServerError
			errMsg = err.Error()
		} else if errors.Is(err, utils.ErrBadRequest) {
			statusCode = http.StatusBadRequest
			errMsg = err.Error()
		}
		RespondWithError(c, statusCode, errMsg)
		return
	}

	RespondWithSuccess(c, http.StatusOK, "Roles removed successfully from user", nil)
}
