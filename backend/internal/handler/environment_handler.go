package handler

import (
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/service"
	"EffiPlat/backend/internal/utils"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EnvironmentHandler handles HTTP requests related to environments.
type EnvironmentHandler struct {
	service      service.EnvironmentService
	auditService service.AuditLogService
	logger       *zap.Logger
}

// NewEnvironmentHandler creates a new instance of EnvironmentHandler.
func NewEnvironmentHandler(s service.EnvironmentService, a service.AuditLogService, l *zap.Logger) *EnvironmentHandler {
	return &EnvironmentHandler{
		service:      s,
		auditService: a,
		logger:       l,
	}
}

// CreateEnvironment godoc
// @Summary Create a new environment
// @Description Creates a new environment with the given details.
// @Tags environments
// @Accept json
// @Produce json
// @Param environment body model.CreateEnvironmentRequest true "Environment creation request"
// @Success 201 {object} utils.SuccessResponse{data=model.EnvironmentResponse} "Environment created successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request payload or validation error"
// @Failure 409 {object} utils.ErrorResponse "Environment with the same slug already exists"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /environments [post]
// @Security BearerAuth
func (h *EnvironmentHandler) CreateEnvironment(c *gin.Context) {
	var req model.CreateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for CreateEnvironment", zap.Error(err))
		utils.Error(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Basic validation (more can be added in service layer)
	if req.Name == "" || req.Slug == "" {
		utils.Error(c, http.StatusBadRequest, "Name and Slug are required")
		return
	}

	env, err := h.service.CreateEnvironment(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create environment", zap.Error(err))
		if errors.Is(err, utils.ErrAlreadyExists) {
			utils.Error(c, http.StatusConflict, err.Error())
		} else if errors.Is(err, utils.ErrBadRequest) {
			utils.Error(c, http.StatusBadRequest, err.Error())
		} else {
			utils.Error(c, http.StatusInternalServerError, "Failed to create environment: "+err.Error())
		}
		return
	}
	
	// 记录审计日志
	details := map[string]interface{}{
		"name":        env.Name,
		"slug":        env.Slug,
		"description": env.Description,
	}
	_ = h.auditService.LogUserAction(c, string(utils.AuditActionCreate), "ENVIRONMENT", env.ID, details)
	
	utils.Created(c, env)
}

// GetEnvironments godoc
// @Summary Get all environments
// @Description Retrieves a list of all environments, with optional pagination.
// @Tags environments
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination" default(1)
// @Param pageSize query int false "Number of items per page for pagination" default(10)
// @Success 200 {object} utils.SuccessResponse{data=utils.PaginatedData{items=[]model.EnvironmentResponse}} "Environments retrieved successfully"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /environments [get]
// @Security BearerAuth
func (h *EnvironmentHandler) GetEnvironments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	listParams := model.EnvironmentListParams{
		Page:     page,
		PageSize: pageSize,
	}

	environments, total, err := h.service.GetEnvironments(c.Request.Context(), listParams)
	if err != nil {
		h.logger.Error("Failed to list environments", zap.Error(err))
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve environments: "+err.Error())
		return
	}

	utils.Paginated(c, environments, total, page, pageSize)
}

// GetEnvironmentByID godoc
// @Summary Get an environment by ID
// @Description Retrieves a specific environment by its unique ID.
// @Tags environments
// @Accept json
// @Produce json
// @Param id path string true "Environment ID"
// @Success 200 {object} utils.SuccessResponse{data=model.EnvironmentResponse} "Environment retrieved successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid environment ID format"
// @Failure 404 {object} utils.ErrorResponse "Environment not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /environments/{id} [get]
// @Security BearerAuth
func (h *EnvironmentHandler) GetEnvironmentByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid environment ID format")
		return
	}

	env, err := h.service.GetEnvironmentByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			utils.Error(c, http.StatusNotFound, err.Error())
		} else {
			h.logger.Error("Failed to get environment by ID", zap.String("id", idStr), zap.Error(err))
			utils.Error(c, http.StatusInternalServerError, "Failed to retrieve environment: "+err.Error())
		}
		return
	}
	utils.OK(c, env)
}

// GetEnvironmentBySlug godoc
// @Summary Get an environment by slug
// @Description Retrieves a specific environment by its unique slug.
// @Tags environments
// @Accept json
// @Produce json
// @Param slug path string true "Environment Slug"
// @Success 200 {object} utils.SuccessResponse{data=model.EnvironmentResponse} "Environment retrieved successfully"
// @Failure 404 {object} utils.ErrorResponse "Environment not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /environments/slug/{slug} [get]
// @Security BearerAuth
func (h *EnvironmentHandler) GetEnvironmentBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.Error(c, http.StatusBadRequest, "Slug parameter is required")
		return
	}

	env, err := h.service.GetEnvironmentBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			utils.Error(c, http.StatusNotFound, err.Error())
		} else {
			h.logger.Error("Failed to get environment by slug", zap.String("slug", slug), zap.Error(err))
			utils.Error(c, http.StatusInternalServerError, "Failed to retrieve environment: "+err.Error())
		}
		return
	}
	utils.OK(c, env)
}

// UpdateEnvironment godoc
// @Summary Update an existing environment
// @Description Updates an existing environment with the given details.
// @Tags environments
// @Accept json
// @Produce json
// @Param id path string true "Environment ID"
// @Param environment body model.UpdateEnvironmentRequest true "Environment update request"
// @Success 200 {object} utils.SuccessResponse{data=model.EnvironmentResponse} "Environment updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid environment ID format or request payload"
// @Failure 404 {object} utils.ErrorResponse "Environment not found"
// @Failure 409 {object} utils.ErrorResponse "Environment with the same slug already exists (if slug is changed)"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /environments/{id} [put]
// @Security BearerAuth
func (h *EnvironmentHandler) UpdateEnvironment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid environment ID format")
		return
	}

	var req model.UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for UpdateEnvironment", zap.Error(err))
		utils.Error(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// 先获取环境原始数据，用于审计日志
	origEnv, getErr := h.service.GetEnvironmentByID(c.Request.Context(), uint(id))
	if getErr != nil {
		// 如果找不到原始环境，不阻止更新操作
		h.logger.Warn("Could not find original environment for audit logging", 
			zap.String("id", idStr), zap.Error(getErr))
	}
	
	env, err := h.service.UpdateEnvironment(c.Request.Context(), uint(id), req)
	if err != nil {
		h.logger.Error("Failed to update environment", zap.String("id", idStr), zap.Error(err))
		if errors.Is(err, utils.ErrNotFound) {
			utils.Error(c, http.StatusNotFound, err.Error())
		} else if errors.Is(err, utils.ErrAlreadyExists) {
			utils.Error(c, http.StatusConflict, err.Error())
		} else if errors.Is(err, utils.ErrBadRequest) {
			utils.Error(c, http.StatusBadRequest, err.Error())
		} else {
			utils.Error(c, http.StatusInternalServerError, "Failed to update environment: "+err.Error())
		}
		return
	}
	
	// 记录审计日志
	details := map[string]interface{}{
		"before": origEnv,
		"after":  env,
	}
	_ = h.auditService.LogUserAction(c, string(utils.AuditActionUpdate), "ENVIRONMENT", env.ID, details)
	
	utils.OK(c, env)
}

// DeleteEnvironment godoc
// @Summary Delete an environment
// @Description Deletes an existing environment by its ID.
// @Tags environments
// @Accept json
// @Produce json
// @Param id path string true "Environment ID"
// @Success 204 "Environment deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid environment ID format"
// @Failure 404 {object} utils.ErrorResponse "Environment not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /environments/{id} [delete]
// @Security BearerAuth
func (h *EnvironmentHandler) DeleteEnvironment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid environment ID format")
		utils.Error(c, http.StatusBadRequest, "Invalid environment ID format")
		return
	}

	// 先获取环境数据，用于审计日志
	env, getErr := h.service.GetEnvironmentByID(c.Request.Context(), uint(id))
	if getErr != nil {
		// 如果找不到原始环境，不阻止删除操作
		h.logger.Warn("Could not find environment for audit logging before deletion", 
			zap.String("id", idStr), zap.Error(getErr))
	}
	
	err = h.service.DeleteEnvironment(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			utils.Error(c, http.StatusNotFound, err.Error())
		} else {
			h.logger.Error("Failed to delete environment", zap.String("id", idStr), zap.Error(err))
			utils.Error(c, http.StatusInternalServerError, "Failed to delete environment: "+err.Error())
		}
		return
	}
	
	// 记录审计日志
	if env != nil {
		details := map[string]interface{}{
			"deletedEnvironment": env,
		}
		_ = h.auditService.LogUserAction(c, string(utils.AuditActionDelete), "ENVIRONMENT", uint(id), details)
	}
	
	c.Status(http.StatusNoContent)
}
