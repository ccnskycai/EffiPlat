package handler

import (
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/service"
	"EffiPlat/backend/internal/utils" // 使用internal/utils代替pkg/response
	"errors"                          // Import errors
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ResponsibilityGroupHandler handles API requests for responsibility groups.
type ResponsibilityGroupHandler struct {
	responsibilityGroupService service.ResponsibilityGroupService
	auditService              service.AuditLogService
	logger                    *zap.Logger
}

// NewResponsibilityGroupHandler creates a new ResponsibilityGroupHandler.
func NewResponsibilityGroupHandler(rgs service.ResponsibilityGroupService, auditSvc service.AuditLogService, logger *zap.Logger) *ResponsibilityGroupHandler {
	return &ResponsibilityGroupHandler{
		responsibilityGroupService: rgs,
		auditService:              auditSvc,
		logger:                    logger,
	}
}

// Request structures for Responsibility Group

type CreateResponsibilityGroupRequest struct {
	Name              string `json:"name" binding:"required"`
	Description       string `json:"description"`
	ResponsibilityIDs []uint `json:"responsibility_ids"` // IDs of responsibilities to associate
}

type UpdateResponsibilityGroupRequest struct {
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	ResponsibilityIDs *[]uint `json:"responsibility_ids"` // Pointer to allow distinguishing between empty and not provided
}

// CreateResponsibilityGroup handles the creation of a new responsibility group.
func (h *ResponsibilityGroupHandler) CreateResponsibilityGroup(c *gin.Context) {
	var req CreateResponsibilityGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for CreateResponsibilityGroup", zap.Error(err))
		utils.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	group := &model.ResponsibilityGroup{
		Name:        req.Name,
		Description: req.Description,
	}

	createdGroup, err := h.responsibilityGroupService.CreateResponsibilityGroup(c.Request.Context(), group, req.ResponsibilityIDs)
	if err != nil {
		h.logger.Error("Failed to create responsibility group", zap.Error(err))
		// TODO: Map service errors (e.g., ErrAlreadyExists, validation errors for IDs)
		utils.InternalServerError(c, "Failed to create responsibility group: "+err.Error())
		return
	}

	// 记录审计日志
	details := map[string]interface{}{
		"id":              createdGroup.ID,
		"name":            createdGroup.Name,
		"description":     createdGroup.Description,
		"responsibilityIDs": req.ResponsibilityIDs,
	}
	_ = h.auditService.LogUserAction(c, string(utils.AuditActionCreate), "RESPONSIBILITY_GROUP", createdGroup.ID, details)

	utils.Created(c, createdGroup)
}

// GetResponsibilityGroups handles listing responsibility groups.
func (h *ResponsibilityGroupHandler) GetResponsibilityGroups(c *gin.Context) {
	var params model.ResponsibilityGroupListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Error("Failed to bind query for GetResponsibilityGroups", zap.Error(err))
		utils.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	groups, total, err := h.responsibilityGroupService.GetResponsibilityGroups(c.Request.Context(), params)
	if err != nil {
		h.logger.Error("Failed to get responsibility groups", zap.Error(err))
		utils.InternalServerError(c, "Failed to retrieve responsibility groups: "+err.Error())
		return
	}
	utils.Paginated(c, groups, total, params.Page, params.PageSize)
}

// GetResponsibilityGroupByID handles retrieving a single responsibility group by its ID.
func (h *ResponsibilityGroupHandler) GetResponsibilityGroupByID(c *gin.Context) {
	idStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid group ID format", zap.String("groupId", idStr), zap.Error(err))
		utils.BadRequest(c, "Invalid group ID format")
		return
	}

	group, err := h.responsibilityGroupService.GetResponsibilityGroupByID(c.Request.Context(), uint(groupID))
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) { // Use errors.Is with apputils.ErrNotFound
			h.logger.Warn("Responsibility group not found", zap.Uint("groupId", uint(groupID)), zap.Error(err))
			utils.NotFound(c, "Responsibility group not found")
		} else {
			h.logger.Error("Failed to get responsibility group by ID", zap.Uint("groupId", uint(groupID)), zap.Error(err))
			utils.InternalServerError(c, "Failed to retrieve responsibility group: "+err.Error())
		}
		return
	}
	utils.OK(c, group)
}

// UpdateResponsibilityGroup handles updating an existing responsibility group.
func (h *ResponsibilityGroupHandler) UpdateResponsibilityGroup(c *gin.Context) {
	idStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid group ID format for update", zap.String("groupId", idStr), zap.Error(err))
		utils.BadRequest(c, "Invalid group ID format")
		return
	}
	
	// 获取原始数据用于审计日志
	origGroup, getErr := h.responsibilityGroupService.GetResponsibilityGroupByID(c.Request.Context(), uint(groupID))
	if getErr != nil {
		h.logger.Warn("Could not get original responsibility group data for audit logging", 
			zap.Uint64("id", groupID), zap.Error(getErr))
	}

	var req UpdateResponsibilityGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for UpdateResponsibilityGroup", zap.Error(err))
		utils.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	groupUpdate := &model.ResponsibilityGroup{
		Name:        req.Name,
		Description: req.Description,
	}

	updatedGroup, err := h.responsibilityGroupService.UpdateResponsibilityGroup(c.Request.Context(), uint(groupID), groupUpdate, req.ResponsibilityIDs)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) { // Use errors.Is with apputils.ErrNotFound
			h.logger.Warn("Responsibility group or associated responsibility not found for update", zap.Uint("groupId", uint(groupID)), zap.Error(err))
			utils.NotFound(c, "Responsibility group or associated responsibility not found")
		} else {
			h.logger.Error("Failed to update responsibility group", zap.Uint("groupId", uint(groupID)), zap.Error(err))
			// TODO: Map other service errors (e.g., validation error)
			utils.InternalServerError(c, "Failed to update responsibility group: "+err.Error())
		}
		return
	}

	// 记录审计日志
	details := map[string]interface{}{
		"before": origGroup,
		"after":  updatedGroup,
		"changes": map[string]interface{}{
			"name":        req.Name,
			"description": req.Description,
			"responsibilityIDs": req.ResponsibilityIDs,
		},
	}
	_ = h.auditService.LogUserAction(c, string(utils.AuditActionUpdate), "RESPONSIBILITY_GROUP", updatedGroup.ID, details)

	utils.OK(c, updatedGroup)
}

// DeleteResponsibilityGroup handles deleting a responsibility group.
func (h *ResponsibilityGroupHandler) DeleteResponsibilityGroup(c *gin.Context) {
	idStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid group ID format for delete", zap.String("groupId", idStr), zap.Error(err))
		utils.BadRequest(c, "Invalid group ID format")
		return
	}
	
	// 获取要删除的责任组数据用于审计日志
	group, getErr := h.responsibilityGroupService.GetResponsibilityGroupByID(c.Request.Context(), uint(groupID))
	if getErr != nil {
		h.logger.Warn("Could not get responsibility group data for audit logging before deletion", 
			zap.Uint64("id", groupID), zap.Error(getErr))
	}

	err = h.responsibilityGroupService.DeleteResponsibilityGroup(c.Request.Context(), uint(groupID))
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) { // Use errors.Is with apputils.ErrNotFound
			h.logger.Warn("Responsibility group not found for delete", zap.Uint("groupId", uint(groupID)), zap.Error(err))
			utils.NotFound(c, "Responsibility group not found")
		} else {
			h.logger.Error("Failed to delete responsibility group", zap.Uint("groupId", uint(groupID)), zap.Error(err))
			utils.InternalServerError(c, "Failed to delete responsibility group: "+err.Error())
		}
		return
	}

	// 记录审计日志
	if group != nil {
		details := map[string]interface{}{
			"deletedGroup": map[string]interface{}{
				"id":          group.ID,
				"name":        group.Name,
				"description": group.Description,
			},
		}
		_ = h.auditService.LogUserAction(c, string(utils.AuditActionDelete), "RESPONSIBILITY_GROUP", uint(groupID), details)
	}

	utils.Status(c, http.StatusNoContent)
}

// AddResponsibilityToGroup handles adding a responsibility to a group.
func (h *ResponsibilityGroupHandler) AddResponsibilityToGroup(c *gin.Context) {
	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid group ID format for adding responsibility", zap.String("groupId", groupIDStr), zap.Error(err))
		utils.BadRequest(c, "Invalid group ID format")
		return
	}

	respIDStr := c.Param("responsibilityId")
	respID, err := strconv.ParseUint(respIDStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid responsibility ID format for adding to group", zap.String("responsibilityId", respIDStr), zap.Error(err))
		utils.BadRequest(c, "Invalid responsibility ID format")
		return
	}

	err = h.responsibilityGroupService.AddResponsibilityToGroup(c.Request.Context(), uint(groupID), uint(respID))
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) { // Use errors.Is with apputils.ErrNotFound
			// Here, ErrNotFound could mean group or responsibility was not found.
			// The service layer's error message (e.g., "responsibility group with id X not found") will be more specific.
			h.logger.Warn("Failed to add responsibility to group: group or responsibility not found", zap.Uint("groupID", uint(groupID)), zap.Uint("respID", uint(respID)), zap.Error(err))
			utils.NotFound(c, err.Error()) // Use the specific error from service layer
		} else {
			h.logger.Error("Failed to add responsibility to group", zap.Uint("groupID", uint(groupID)), zap.Uint("respID", uint(respID)), zap.Error(err))
			utils.InternalServerError(c, "Failed to add responsibility to group: "+err.Error())
		}
		return
	}
	utils.Status(c, http.StatusNoContent)
}

// RemoveResponsibilityFromGroup handles removing a responsibility from a group.
func (h *ResponsibilityGroupHandler) RemoveResponsibilityFromGroup(c *gin.Context) {
	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid group ID format for removing responsibility", zap.String("groupId", groupIDStr), zap.Error(err))
		utils.BadRequest(c, "Invalid group ID format")
		return
	}

	respIDStr := c.Param("responsibilityId")
	respID, err := strconv.ParseUint(respIDStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid responsibility ID format for removing from group", zap.String("responsibilityId", respIDStr), zap.Error(err))
		utils.BadRequest(c, "Invalid responsibility ID format")
		return
	}

	err = h.responsibilityGroupService.RemoveResponsibilityFromGroup(c.Request.Context(), uint(groupID), uint(respID))
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) { // Use errors.Is with apputils.ErrNotFound
			// ErrNotFound could mean group, responsibility, or the association itself was not found.
			h.logger.Warn("Failed to remove responsibility from group: entity or association not found", zap.Uint("groupID", uint(groupID)), zap.Uint("respID", uint(respID)), zap.Error(err))
			utils.NotFound(c, err.Error()) // Use the specific error from service layer
		} else {
			h.logger.Error("Failed to remove responsibility from group", zap.Uint("groupID", uint(groupID)), zap.Uint("respID", uint(respID)), zap.Error(err))
			utils.InternalServerError(c, "Failed to remove responsibility from group: "+err.Error())
		}
		return
	}
	utils.Status(c, http.StatusNoContent)
}
