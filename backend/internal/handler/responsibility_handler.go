package handler

import (
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/service"
	"EffiPlat/backend/internal/utils" // Import apputils
	"EffiPlat/backend/pkg/response"   // Assuming you have a response package
	"errors"                          // Import errors
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ResponsibilityHandler struct {
	service service.ResponsibilityService
	logger  *zap.Logger
}

func NewResponsibilityHandler(service service.ResponsibilityService, logger *zap.Logger) *ResponsibilityHandler {
	return &ResponsibilityHandler{
		service: service,
		logger:  logger,
	}
}

// CreateResponsibility handles the creation of a new responsibility.
func (h *ResponsibilityHandler) CreateResponsibility(c *gin.Context) {
	var req models.Responsibility
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for CreateResponsibility", zap.Error(err))
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	// TODO: Add more specific validation if needed for req fields

	createdResp, err := h.service.CreateResponsibility(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create responsibility", zap.Error(err))
		// TODO: Map service errors to appropriate HTTP responses (e.g., if already exists)
		response.InternalServerError(c, "Failed to create responsibility: "+err.Error())
		return
	}
	response.Created(c, createdResp)
}

// GetResponsibilities handles listing responsibilities.
func (h *ResponsibilityHandler) GetResponsibilities(c *gin.Context) {
	var params models.ResponsibilityListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Error("Failed to bind query for GetResponsibilities", zap.Error(err))
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	// Set default pagination if not provided
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10 // Default page size
	}

	responsibilities, total, err := h.service.GetResponsibilities(c.Request.Context(), params)
	if err != nil {
		h.logger.Error("Failed to get responsibilities", zap.Error(err))
		response.InternalServerError(c, "Failed to retrieve responsibilities: "+err.Error())
		return
	}

	response.Paginated(c, responsibilities, total, params.Page, params.PageSize)
}

// GetResponsibilityByID handles retrieving a single responsibility by its ID.
func (h *ResponsibilityHandler) GetResponsibilityByID(c *gin.Context) {
	idStr := c.Param("responsibilityId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid responsibility ID format", zap.String("responsibilityId", idStr), zap.Error(err))
		response.BadRequest(c, "Invalid responsibility ID format")
		return
	}

	resp, err := h.service.GetResponsibilityByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) { // Use errors.Is with apputils.ErrNotFound
			h.logger.Warn("Responsibility not found", zap.Uint("responsibilityId", uint(id)), zap.Error(err))
			response.NotFound(c, "Responsibility not found")
		} else {
			h.logger.Error("Failed to get responsibility by ID", zap.Uint("responsibilityId", uint(id)), zap.Error(err))
			response.InternalServerError(c, "Failed to retrieve responsibility: "+err.Error())
		}
		return
	}
	response.OK(c, resp)
}

// UpdateResponsibility handles updating an existing responsibility.
func (h *ResponsibilityHandler) UpdateResponsibility(c *gin.Context) {
	idStr := c.Param("responsibilityId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid responsibility ID format for update", zap.String("responsibilityId", idStr), zap.Error(err))
		response.BadRequest(c, "Invalid responsibility ID format")
		return
	}

	var req models.Responsibility
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for UpdateResponsibility", zap.Error(err))
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	// req.ID will be ignored by the service layer, it uses the path parameter `id`
	updatedResp, err := h.service.UpdateResponsibility(c.Request.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) { // Use errors.Is with apputils.ErrNotFound
			h.logger.Warn("Responsibility not found for update", zap.Uint("responsibilityId", uint(id)), zap.Error(err))
			response.NotFound(c, "Responsibility not found")
		} else {
			h.logger.Error("Failed to update responsibility", zap.Uint("responsibilityId", uint(id)), zap.Error(err))
			// TODO: Map other service errors (e.g., validation error)
			response.InternalServerError(c, "Failed to update responsibility: "+err.Error())
		}
		return
	}
	response.OK(c, updatedResp)
}

// DeleteResponsibility handles deleting a responsibility.
func (h *ResponsibilityHandler) DeleteResponsibility(c *gin.Context) {
	idStr := c.Param("responsibilityId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid responsibility ID format for delete", zap.String("responsibilityId", idStr), zap.Error(err))
		response.BadRequest(c, "Invalid responsibility ID format")
		return
	}

	err = h.service.DeleteResponsibility(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) { // Use errors.Is with apputils.ErrNotFound
			h.logger.Warn("Responsibility not found for delete", zap.Uint("responsibilityId", uint(id)), zap.Error(err))
			response.NotFound(c, "Responsibility not found")
		} else {
			h.logger.Error("Failed to delete responsibility", zap.Uint("responsibilityId", uint(id)), zap.Error(err))
			response.InternalServerError(c, "Failed to delete responsibility: "+err.Error())
		}
		return
	}
	response.Status(c, http.StatusNoContent)
}
