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

type ResponsibilityHandler struct {
	responsibilityService service.ResponsibilityService
	auditService          service.AuditLogService
	logger                *zap.Logger
}

func NewResponsibilityHandler(rs service.ResponsibilityService, auditSvc service.AuditLogService, logger *zap.Logger) *ResponsibilityHandler {
	return &ResponsibilityHandler{
		responsibilityService: rs,
		auditService:          auditSvc,
		logger:                logger,
	}
}

// CreateResponsibility handles the creation of a new responsibility.
func (h *ResponsibilityHandler) CreateResponsibility(c *gin.Context) {
	var req model.Responsibility
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for CreateResponsibility", zap.Error(err))
		utils.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	// TODO: Add more specific validation if needed for req fields

	createdResp, err := h.responsibilityService.CreateResponsibility(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create responsibility", zap.Error(err))
		// TODO: Map service errors to appropriate HTTP responses (e.g., if already exists)
		utils.InternalServerError(c, "Failed to create responsibility: "+err.Error())
		return
	}

	// 记录审计日志
	details := map[string]interface{}{
		"id":          createdResp.ID,
		"name":        createdResp.Name,
		"description": createdResp.Description,
	}
	_ = h.auditService.LogUserAction(c, string(utils.AuditActionCreate), "RESPONSIBILITY", createdResp.ID, details)

	utils.Created(c, createdResp)
}

// GetResponsibilities handles listing responsibilities.
func (h *ResponsibilityHandler) GetResponsibilities(c *gin.Context) {
	var params model.ResponsibilityListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Error("Failed to bind query for GetResponsibilities", zap.Error(err))
		utils.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	// Set default pagination if not provided
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10 // Default page size
	}

	responsibilities, total, err := h.responsibilityService.GetResponsibilities(c.Request.Context(), params)
	if err != nil {
		h.logger.Error("Failed to get responsibilities", zap.Error(err))
		utils.InternalServerError(c, "Failed to retrieve responsibilities: "+err.Error())
		return
	}

	utils.Paginated(c, responsibilities, total, params.Page, params.PageSize)
}

// GetResponsibilityByID handles retrieving a single responsibility by its ID.
func (h *ResponsibilityHandler) GetResponsibilityByID(c *gin.Context) {
	idStr := c.Param("responsibilityId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid responsibility ID format", zap.String("responsibilityId", idStr), zap.Error(err))
		utils.BadRequest(c, "Invalid responsibility ID format")
		return
	}

	resp, err := h.responsibilityService.GetResponsibilityByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) { // Use errors.Is with apputils.ErrNotFound
			h.logger.Warn("Responsibility not found", zap.Uint("responsibilityId", uint(id)), zap.Error(err))
			utils.NotFound(c, "Responsibility not found")
		} else {
			h.logger.Error("Failed to get responsibility by ID", zap.Uint("responsibilityId", uint(id)), zap.Error(err))
			utils.InternalServerError(c, "Failed to retrieve responsibility: "+err.Error())
		}
		return
	}
	utils.OK(c, resp)
}

// UpdateResponsibility handles updating an existing responsibility.
func (h *ResponsibilityHandler) UpdateResponsibility(c *gin.Context) {
	idStr := c.Param("responsibilityId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid responsibility ID format for update", zap.String("responsibilityId", idStr), zap.Error(err))
		utils.BadRequest(c, "Invalid responsibility ID format")
		return
	}

	var req model.Responsibility
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for UpdateResponsibility", zap.Error(err))
		utils.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	// 获取原始数据用于审计日志
	origResp, getErr := h.responsibilityService.GetResponsibilityByID(c.Request.Context(), uint(id))
	if getErr != nil {
		h.logger.Warn("Could not get original responsibility data for audit logging", 
			zap.Uint64("id", id), zap.Error(getErr))
	}

	// req.ID will be ignored by the service layer, it uses the path parameter `id`
	updatedResp, err := h.responsibilityService.UpdateResponsibility(c.Request.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) { // Use errors.Is with apputils.ErrNotFound
			h.logger.Warn("Responsibility not found for update", zap.Uint("responsibilityId", uint(id)), zap.Error(err))
			utils.NotFound(c, "Responsibility not found")
			return
		}
		h.logger.Error("Failed to update responsibility", zap.Error(err))
		utils.InternalServerError(c, "Failed to update responsibility: "+err.Error())
		return
	}

	// 记录审计日志
	details := map[string]interface{}{
		"before": origResp,
		"after":  updatedResp,
		"changes": req,
	}
	_ = h.auditService.LogUserAction(c, string(utils.AuditActionUpdate), "RESPONSIBILITY", updatedResp.ID, details)

	utils.OK(c, updatedResp)
}

// DeleteResponsibility handles deleting a responsibility.
func (h *ResponsibilityHandler) DeleteResponsibility(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("responsibilityId"), 10, 32)
	if err != nil {
		h.logger.Error("Failed to parse responsibility ID", zap.Error(err))
		utils.BadRequest(c, "Invalid responsibility ID format")
		return
	}
	
	// 获取要删除的责任数据用于审计日志
	resp, getErr := h.responsibilityService.GetResponsibilityByID(c.Request.Context(), uint(id))
	if getErr != nil {
		h.logger.Warn("Could not get responsibility data for audit logging before deletion", 
			zap.Uint64("id", id), zap.Error(getErr))
	}

	err = h.responsibilityService.DeleteResponsibility(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) { // Use errors.Is with apputils.ErrNotFound
			h.logger.Warn("Responsibility not found for delete", zap.Uint("responsibilityId", uint(id)), zap.Error(err))
			utils.NotFound(c, "Responsibility not found")
		} else {
			h.logger.Error("Failed to delete responsibility", zap.Uint("responsibilityId", uint(id)), zap.Error(err))
			utils.InternalServerError(c, "Failed to delete responsibility: "+err.Error())
		}
		return
	}

	// 记录审计日志
	if resp != nil {
		details := map[string]interface{}{
			"deletedResponsibility": map[string]interface{}{
				"id":          resp.ID,
				"name":        resp.Name,
				"description": resp.Description,
			},
		}
		_ = h.auditService.LogUserAction(c, string(utils.AuditActionDelete), "RESPONSIBILITY", uint(id), details)
	}

	utils.Status(c, http.StatusNoContent)
}
