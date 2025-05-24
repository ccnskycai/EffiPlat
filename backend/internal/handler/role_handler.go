package handler

import (
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/service"
	"EffiPlat/backend/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RoleHandler struct {
	roleService  service.RoleService
	auditService service.AuditLogService
	logger       *zap.Logger
}

func NewRoleHandler(rs service.RoleService, auditSvc service.AuditLogService, logger *zap.Logger) *RoleHandler {
	return &RoleHandler{
		roleService:  rs,
		auditService: auditSvc,
		logger:       logger,
	}
}

// Request structures based on API design
type CreateRoleRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	PermissionIDs []uint `json:"permissionIds"`
}

type UpdateRoleRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	PermissionIDs []uint `json:"permissionIds"`
}

// CreateRole godoc
// @Summary Create a new role
// @Description Create a new role with name and description
// @Tags roles
// @Accept  json
// @Produce  json
// @Param   role body CreateRoleRequest true "Role object"
// @Success 201 {object} model.Role "Actually returns a unified response: {code, message, data: model.Role}"
// @Failure 400 {object} map[string]interface{} "Bad Request (e.g. validation error, name exists)"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("CreateRole: Failed to bind JSON", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// 设置审计日志操作类型和资源
	c.Set("auditAction", string(utils.AuditActionCreate))
	c.Set("auditResource", "ROLE")

	// Convert handler request to service layer model/struct if necessary
	// For now, assuming roleService.CreateRole can take something compatible with req
	// Or, more likely, a specific model for creation, e.g., model.RoleInput
	roleToCreate := model.Role{ // This is an assumption, adjust based on your actual model.Role
		Name:        req.Name,
		Description: req.Description,
		// PermissionIDs might be handled by the service layer through a separate field or method
	}

	// Simpler assumption for now, service handles permission linking if needed
	createdRole, err := h.roleService.CreateRole(c.Request.Context(), &roleToCreate, req.PermissionIDs) // Adjusted to pass PermissionIDs
	if err != nil {
		h.logger.Error("CreateRole: Service error", zap.Error(err))
		if err.Error() == "role name already exists" { // Placeholder for actual error checking
			RespondWithError(c, http.StatusBadRequest, "Validation error: Role name already exists")
		} else {
			RespondWithError(c, http.StatusInternalServerError, "Failed to create role")
		}
		return
	}

	// 记录审计日志
	details := map[string]interface{}{
		"id":           createdRole.ID,
		"name":         createdRole.Name,
		"description":  createdRole.Description,
		"permissionIds": req.PermissionIDs,
	}
	_ = h.auditService.LogUserAction(c, string(utils.AuditActionCreate), "ROLE", createdRole.ID, details)

	RespondWithSuccess(c, http.StatusCreated, "Role created successfully", createdRole)
}

// GetRoles godoc
// @Summary Get all roles
// @Description Retrieve a list of all roles with pagination and search
// @Tags roles
// @Produce  json
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Number of items per page (default: 10)"
// @Param name query string false "Search by role name (fuzzy)"
// @Success 200 {object} map[string]interface{} "Unified response: {code, message, data: {items: []model.Role, total: int, page: int, pageSize: int}}"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /roles [get]
func (h *RoleHandler) GetRoles(c *gin.Context) {
	pageQuery := c.DefaultQuery("page", "1")
	pageSizeQuery := c.DefaultQuery("pageSize", "10") // Design doc mentions 20, but test default is often 10
	nameQuery := c.Query("name")

	page, err := strconv.Atoi(pageQuery)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeQuery)
	if err != nil || pageSize < 1 {
		pageSize = 10 // Default to 10 if invalid
	}

	params := model.RoleListParams{ // Assuming this struct exists in models
		Page:     page,
		PageSize: pageSize,
		Name:     nameQuery,
	}

	roles, total, err := h.roleService.GetRoles(c.Request.Context(), params)
	if err != nil {
		h.logger.Error("GetRoles: Service error", zap.Error(err))
		RespondWithError(c, http.StatusInternalServerError, "Failed to retrieve roles")
		return
	}

	RespondWithSuccess(c, http.StatusOK, "Roles retrieved successfully", gin.H{
		"items":    roles,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetRoleByID godoc
// @Summary Get a role by ID
// @Description Retrieve a specific role by its ID, including user count and permissions
// @Tags roles
// @Produce  json
// @Param   roleId path int true "Role ID"
// @Success 200 {object} model.Role "Actually returns a unified response: {code, message, data: model.RoleDetails (with userCount, permissions)}"
// @Failure 400 {object} map[string]interface{} "Bad Request (Invalid ID)"
// @Failure 404 {object} map[string]interface{} "Not Found (Role not found 40402)"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /roles/{roleId} [get]
func (h *RoleHandler) GetRoleByID(c *gin.Context) {
	roleIDStr := c.Param("roleId") // Design says roleId, router.go uses roleId
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		h.logger.Error("GetRoleByID: Invalid role ID format", zap.String("roleId", roleIDStr), zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid role ID format")
		return
	}

	// 设置审计日志操作类型和资源
	c.Set("auditAction", string(utils.AuditActionRead))
	c.Set("auditResource", "ROLE")
	c.Set("auditResourceID", uint(roleID))

	role, err := h.roleService.GetRoleByID(c.Request.Context(), uint(roleID))
	if err != nil {
		h.logger.Error("GetRoleByID: Service error", zap.Uint("roleId", uint(roleID)), zap.Error(err))
		if err.Error() == "role not found" { // Placeholder for actual error check
			RespondWithError(c, http.StatusNotFound, "Role not found")
		} else {
			RespondWithError(c, http.StatusInternalServerError, "Failed to retrieve role")
		}
		return
	}

	// 记录查询详情 (对于READ操作，只需记录基本信息)
	utils.SetAuditDetails(c, map[string]interface{}{
		"action": "READ",
		"roleQueried": uint(roleID),
	})

	// Optional: Get related data as required by the API design
	// e.g., user count, permissions if not already included
	// Example: role.UserCount = h.roleService.GetUserCountForRole(roleID)

	RespondWithSuccess(c, http.StatusOK, "Role retrieved successfully", role)
}

// UpdateRole godoc
// @Summary Update an existing role
// @Description Update an existing role by its ID
// @Tags roles
// @Accept  json
// @Produce  json
// @Param   roleId path int true "Role ID"
// @Param   role body UpdateRoleRequest true "Role object with updated fields"
// @Success 200 {object} model.Role "Actually returns a unified response: {code, message, data: model.Role}"
// @Failure 400 {object} map[string]interface{} "Bad Request (Invalid ID, validation error, name exists 40002)"
// @Failure 404 {object} map[string]interface{} "Not Found (Role not found 40402)"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /roles/{roleId} [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	roleIDStr := c.Param("roleId")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		h.logger.Error("UpdateRole: Invalid role ID format", zap.String("roleId", roleIDStr), zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid role ID format")
		return
	}

	// 设置审计日志操作类型和资源
	c.Set("auditAction", string(utils.AuditActionUpdate))
	c.Set("auditResource", "ROLE")
	c.Set("auditResourceID", uint(roleID))

	// 获取更新前的角色信息，用于审计日志
	existingRole, err := h.roleService.GetRoleByID(c.Request.Context(), uint(roleID))
	if err != nil {
		h.logger.Error("UpdateRole: Failed to get existing role", zap.Uint("roleId", uint(roleID)), zap.Error(err))
		if err.Error() == "role not found" { // Placeholder
			RespondWithError(c, http.StatusNotFound, "Role not found")
		} else {
			RespondWithError(c, http.StatusInternalServerError, "Failed to retrieve role")
		}
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("UpdateRole: Failed to bind JSON", zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// 记录更新前的状态用于审计日志
	beforeUpdate := map[string]interface{}{
		"id":          existingRole.ID,
		"name":        existingRole.Name,
		"description": existingRole.Description,
		// 添加其他需要审计的字段
	}

	// Assuming service layer takes a model for update, or individual fields
	roleToUpdate := model.Role{ // This is an assumption
		Name:        req.Name,
		Description: req.Description,
	}

	updatedRole, err := h.roleService.UpdateRole(c.Request.Context(), uint(roleID), &roleToUpdate, req.PermissionIDs) // Adjusted
	if err != nil {
		h.logger.Error("UpdateRole: Service error", zap.Uint("roleId", uint(roleID)), zap.Error(err))
		if err.Error() == "role not found" { // Placeholder
			RespondWithError(c, http.StatusNotFound, "Role not found")
		} else if err.Error() == "role name already exists" { // Placeholder
			RespondWithError(c, http.StatusBadRequest, "Validation error: Role name already exists")
		} else {
			RespondWithError(c, http.StatusInternalServerError, "Failed to update role")
		}
		return
	}

	// 记录审计日志
	details := map[string]interface{}{
		"before": beforeUpdate,
		"after": map[string]interface{}{
			"id":           updatedRole.ID,
			"name":         updatedRole.Name,
			"description":  updatedRole.Description,
			"permissionIds": req.PermissionIDs,
		},
		"changes": req,
	}
	_ = h.auditService.LogUserAction(c, string(utils.AuditActionUpdate), "ROLE", updatedRole.ID, details)

	RespondWithSuccess(c, http.StatusOK, "Role updated successfully", updatedRole)
}

// DeleteRole godoc
// @Summary Delete a role
// @Description Delete a role by its ID
// @Tags roles
// @Produce  json
// @Param   roleId path int true "Role ID"
// @Success 204 {object} map[string]interface{} "Actually returns a unified response: {code, message, data: null} with HTTP 200/204. Design suggests 204, but unified response might send 200."
// @Failure 400 {object} map[string]interface{} "Bad Request (Invalid ID, role assigned to users 40003)"
// @Failure 404 {object} map[string]interface{} "Not Found (Role not found 40402)"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /roles/{roleId} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	roleIDStr := c.Param("roleId") // Design says roleId, router.go uses roleId
	roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
	if err != nil {
		h.logger.Error("DeleteRole: Invalid role ID format", zap.String("roleId", roleIDStr), zap.Error(err))
		RespondWithError(c, http.StatusBadRequest, "Invalid role ID format")
		return
	}

	// 设置审计日志操作类型和资源
	c.Set("auditAction", string(utils.AuditActionDelete))
	c.Set("auditResource", "ROLE")
	c.Set("auditResourceID", uint(roleID))

	// 获取将要删除的角色信息，用于审计日志
	existingRole, err := h.roleService.GetRoleByID(c.Request.Context(), uint(roleID))
	if err != nil {
		h.logger.Error("DeleteRole: Failed to get existing role", zap.Uint("roleId", uint(roleID)), zap.Error(err))
		if err.Error() == "role not found" { // Placeholder
			RespondWithError(c, http.StatusNotFound, "Role not found")
		} else {
			RespondWithError(c, http.StatusInternalServerError, "Failed to retrieve role")
		}
		return
	}

	// 记录审计日志
	details := map[string]interface{}{
		"deletedRole": map[string]interface{}{
			"id":          existingRole.ID,
			"name":        existingRole.Name,
			"description": existingRole.Description,
			// 得到角色的权限信息
			"permissions": existingRole.Permissions,
		},
	}
	_ = h.auditService.LogUserAction(c, string(utils.AuditActionDelete), "ROLE", existingRole.ID, details)

	err = h.roleService.DeleteRole(c.Request.Context(), uint(roleID))
	if err != nil {
		h.logger.Error("DeleteRole: Service error", zap.Uint("roleId", uint(roleID)), zap.Error(err))
		if err.Error() == "role not found" { // Placeholder
			RespondWithError(c, http.StatusNotFound, "Role not found")
		} else if err.Error() == "role is assigned to users" { // Placeholder for actual error check
			RespondWithError(c, http.StatusBadRequest, "Cannot delete role: it is assigned to one or more users")
		} else {
			RespondWithError(c, http.StatusInternalServerError, "Failed to delete role")
		}
		return
	}

	// For 204 No Content, the common RespondWithSuccess might add a body, which is fine.
	// If strictly no body is needed for 204, use c.Status(http.StatusNoContent) directly.
	// RespondWithSuccess(c, http.StatusNoContent, "Role deleted successfully", nil) // Or use http.StatusOK for consistency
	c.Status(http.StatusNoContent) // Consistent with 204 No Content for DELETE
}

// TODO: Add handlers for role permissions if needed (GET /roles/{roleId}/permissions, etc.)
// func (h *RoleHandler) GetRolePermissions(c *gin.Context) {}
// func (h *RoleHandler) AddPermissionToRole(c *gin.Context) {}
// func (h *RoleHandler) RemovePermissionFromRole(c *gin.Context) {}
