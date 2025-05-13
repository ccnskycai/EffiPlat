package handler

import (
	"EffiPlat/backend/internal/service" // Added to access UserService interface and error variables
	"EffiPlat/backend/pkg/utils"
	"errors"
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
	// Parse query parameters for pagination and filtering
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// Collect filter params (example)
	filterParams := map[string]string{}
	if name := c.Query("name"); name != "" {
		filterParams["name"] = name
	}
	if email := c.Query("email"); email != "" {
		filterParams["email"] = email
	}
	if status := c.Query("status"); status != "" {
		filterParams["status"] = status
	}
	sortBy := c.Query("sortBy")
	order := c.Query("order")
	if sortBy != "" {
		filterParams["sortBy"] = sortBy
	}
	if order != "" {
		filterParams["order"] = order
	}

	paginatedResult, err := h.userService.GetUsers(filterParams, page, pageSize)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve users", err.Error())
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "Users retrieved successfully", paginatedResult)
}

// CreateUser handles POST /users request
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
		return
	}

	// Pass the request context to the service layer
	user, err := h.userService.CreateUser(c.Request.Context(), req.Name, req.Email, req.Password, req.Department, req.Roles)
	if err != nil {
		if errors.Is(err, service.ErrEmailExists) {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Failed to create user", err.Error())
		} else if errors.Is(err, service.ErrRoleNotFound) {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Failed to create user", err.Error())
		} else if errors.Is(err, service.ErrPasswordHashing) {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create user", err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create user", err.Error())
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusCreated, "User created successfully", user)
}

// GetUserByID handles GET /users/{userId} request
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", err.Error())
		return
	}

	user, err := h.userService.GetUserByID(uint(userID))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, "User not found", err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user", err.Error())
		}
		return
	}

	// 从上下文中获取用户信息
	claimsValue, exists := c.Get("user")
	if !exists {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "User claims not found", "User claims not found in context")
		return
	}

	claims, ok := claimsValue.(*models.Claims)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid claims format", "Invalid claims format in context")
		return
	}

	// 检查权限：用户只能获取自己的信息，或者管理员可以获取任何人的信息
	if claims.UserID != uint(userID) && !hasAdminRole(claims) {
		utils.SendErrorResponse(c, http.StatusForbidden, "Permission denied", "User can only retrieve their own information or requires admin role")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "User retrieved successfully", user)
}

// 辅助函数：检查用户是否具有管理员角色
func hasAdminRole(claims *models.Claims) bool {
	// TODO: 实现角色检查逻辑
	// 目前简单返回 true，后续需要根据实际角色系统实现
	return true
}

// UpdateUser handles PUT /users/{userId} request
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", err.Error())
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Validation failed", utils.FormatValidationErrors(err))
		return
	}

	// 从上下文中获取用户信息
	claimsValue, exists := c.Get("user")
	if !exists {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "User claims not found", "User claims not found in context")
		return
	}

	claims, ok := claimsValue.(*models.Claims)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid claims format", "Invalid claims format in context")
		return
	}

	// 检查权限：用户只能更新自己的信息，或者管理员可以更新任何人的信息
	if claims.UserID != uint(userID) && !hasAdminRole(claims) {
		utils.SendErrorResponse(c, http.StatusForbidden, "Permission denied", "User can only update their own information or requires admin role")
		return
	}

	// 如果不是管理员，限制可更新的字段
	if !hasAdminRole(claims) {
		if req.Status != nil || req.Roles != nil {
			utils.SendErrorResponse(c, http.StatusForbidden, "Permission denied", "Only admins can change status or roles")
			return
		}
	}

	user, err := h.userService.UpdateUser(uint(userID), req.Name, req.Department, req.Status, req.Roles)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, "User not found for update", err.Error())
		} else if errors.Is(err, service.ErrRoleNotFound) {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Failed to update user", err.Error())
		} else if errors.Is(err, service.ErrUpdateFailed) {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update user", err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update user", err.Error())
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, "User updated successfully", user)
}

// DeleteUser handles DELETE /users/{userId} request
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", err.Error())
		return
	}

	// 从上下文中获取用户信息
	claimsValue, exists := c.Get("user")
	if !exists {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "User claims not found", "User claims not found in context")
		return
	}

	claims, ok := claimsValue.(*models.Claims)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid claims format", "Invalid claims format in context")
		return
	}

	// 检查权限：只有管理员可以删除用户
	if !hasAdminRole(claims) {
		utils.SendErrorResponse(c, http.StatusForbidden, "Permission denied", "Only admins can delete users")
		return
	}

	err = h.userService.DeleteUser(uint(userID))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, "User not found for deletion", err.Error())
		} else if errors.Is(err, service.ErrDeleteFailed) {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to delete user", err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to delete user", err.Error())
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusNoContent, "User deleted successfully", nil)
} 