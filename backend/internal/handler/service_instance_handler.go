package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"EffiPlat/backend/internal/repository" // For ListServiceInstancesParams
	"EffiPlat/backend/internal/service"
	apputils "EffiPlat/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ServiceInstanceHandler handles HTTP requests for service instances.
type ServiceInstanceHandler struct {
	svc    service.ServiceInstanceService
	logger *zap.Logger
}

// NewServiceInstanceHandler creates a new ServiceInstanceHandler.
func NewServiceInstanceHandler(svc service.ServiceInstanceService, logger *zap.Logger) *ServiceInstanceHandler {
	return &ServiceInstanceHandler{svc: svc, logger: logger}
}

// CreateServiceInstance handles the creation of a new service instance.
// POST /service-instances
func (h *ServiceInstanceHandler) CreateServiceInstance(c *gin.Context) {
	var input service.ServiceInstanceInputDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("Failed to bind JSON for create service instance", zap.Error(err))
		apputils.SendErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
		return
	}

	createdInstance, err := h.svc.CreateServiceInstance(c.Request.Context(), &input)
	if err != nil {
		h.logger.Error("Failed to create service instance", zap.Error(err), zap.Any("input", input))
		if errors.Is(err, apputils.ErrBadRequest) {
			apputils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		} else if errors.Is(err, apputils.ErrAlreadyExists) {
			apputils.SendErrorResponse(c, http.StatusConflict, err.Error())
		} else {
			apputils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create service instance")
		}
		return
	}

	apputils.SendSuccessResponse(c, http.StatusCreated, createdInstance)
}

// GetServiceInstance handles fetching a service instance by ID.
// GET /service-instances/:instanceId
func (h *ServiceInstanceHandler) GetServiceInstance(c *gin.Context) {
	idStr := c.Param("instanceId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid instance ID format", zap.String("instanceId", idStr), zap.Error(err))
		apputils.SendErrorResponse(c, http.StatusBadRequest, "Invalid instance ID format")
		return
	}

	instance, err := h.svc.GetServiceInstanceByID(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get service instance by ID", zap.Uint64("id", id), zap.Error(err))
		if errors.Is(err, apputils.ErrNotFound) {
			apputils.SendErrorResponse(c, http.StatusNotFound, "Service instance not found")
		} else {
			apputils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve service instance")
		}
		return
	}

	apputils.SendSuccessResponse(c, http.StatusOK, instance)
}

// ListServiceInstances handles listing service instances with pagination and filtering.
// GET /service-instances
func (h *ServiceInstanceHandler) ListServiceInstances(c *gin.Context) {
	var params repository.ListServiceInstancesParams

	// Bind query parameters
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Error("Failed to bind query params for list service instances", zap.Error(err))
		apputils.SendErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid query parameters: %v", err))
		return
	}

	// Manual binding for Page and PageSize as ShouldBindQuery might not handle defaults well for int if not provided
	// Or ensure form tags in ListServiceInstancesParams have default values that Gin respects for ints.
	// For simplicity, we can rely on the service/repo layer to apply defaults if values are zero/negative.
	// Let's assume params struct tags `form:"page,default=1"` and `form:"pageSize,default=10"` work or are handled later.

	pagedResult, err := h.svc.ListServiceInstances(c.Request.Context(), &params)
	if err != nil {
		h.logger.Error("Failed to list service instances", zap.Error(err), zap.Any("params", params))
		if errors.Is(err, apputils.ErrBadRequest) {
			apputils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		} else {
			apputils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to list service instances")
		}
		return
	}

	apputils.SendSuccessResponse(c, http.StatusOK, pagedResult)
}

// UpdateServiceInstance handles updating an existing service instance.
// PUT /service-instances/:instanceId
func (h *ServiceInstanceHandler) UpdateServiceInstance(c *gin.Context) {
	idStr := c.Param("instanceId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid instance ID format for update", zap.String("instanceId", idStr), zap.Error(err))
		apputils.SendErrorResponse(c, http.StatusBadRequest, "Invalid instance ID format")
		return
	}

	var input service.ServiceInstanceInputDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("Failed to bind JSON for update service instance", zap.Error(err))
		apputils.SendErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request payload: %v", err))
		return
	}

	updatedInstance, err := h.svc.UpdateServiceInstance(c.Request.Context(), uint(id), &input)
	if err != nil {
		h.logger.Error("Failed to update service instance", zap.Uint64("id", id), zap.Any("input", input), zap.Error(err))
		if errors.Is(err, apputils.ErrNotFound) {
			apputils.SendErrorResponse(c, http.StatusNotFound, "Service instance not found")
		} else if errors.Is(err, apputils.ErrBadRequest) {
			apputils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		} else if errors.Is(err, apputils.ErrAlreadyExists) {
			apputils.SendErrorResponse(c, http.StatusConflict, err.Error())
		} else {
			apputils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update service instance")
		}
		return
	}

	apputils.SendSuccessResponse(c, http.StatusOK, updatedInstance)
}

// DeleteServiceInstance handles deleting a service instance by ID.
// DELETE /service-instances/:instanceId
func (h *ServiceInstanceHandler) DeleteServiceInstance(c *gin.Context) {
	idStr := c.Param("instanceId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid instance ID format for delete", zap.String("instanceId", idStr), zap.Error(err))
		apputils.SendErrorResponse(c, http.StatusBadRequest, "Invalid instance ID format")
		return
	}

	err = h.svc.DeleteServiceInstance(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to delete service instance", zap.Uint64("id", id), zap.Error(err))
		if errors.Is(err, apputils.ErrNotFound) {
			apputils.SendErrorResponse(c, http.StatusNotFound, "Service instance not found or already deleted")
		} else {
			apputils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to delete service instance")
		}
		return
	}

	apputils.SendSuccessResponse(c, http.StatusNoContent, nil) // 204 No Content for successful deletion
}
