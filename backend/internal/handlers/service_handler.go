package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/service"
	"EffiPlat/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ServiceHandler handles HTTP requests for Service and ServiceType resources.
// It uses the ServiceService to interact with the business logic layer.
type ServiceHandler struct {
	service service.ServiceService
	logger  *zap.Logger
}

// NewServiceHandler creates a new instance of ServiceHandler.
func NewServiceHandler(svc service.ServiceService, logger *zap.Logger) *ServiceHandler {
	return &ServiceHandler{service: svc, logger: logger}
}

// --- ServiceType Handlers ---

// CreateServiceType godoc
// @Summary Create a new service type
// @Description Creates a new service type with the provided details.
// @Tags ServiceTypes
// @Accept json
// @Produce json
// @Param serviceType body models.CreateServiceTypeRequest true "Service Type to create"
// @Success 201 {object} models.ServiceType "Successfully created service type"
// @Failure 400 {object} models.ErrorResponse "Invalid request payload or validation error"
// @Failure 409 {object} models.ErrorResponse "Service type with this name already exists"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /service-types [post]
func (h *ServiceHandler) CreateServiceType(c *gin.Context) {
	var req models.CreateServiceTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for CreateServiceType", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	serviceType, err := h.service.CreateServiceType(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create service type", zap.Error(err), zap.String("name", req.Name))
		if errors.Is(err, models.ErrServiceTypeNameExists) {
			utils.SendErrorResponse(c, http.StatusConflict, models.ErrServiceTypeNameExists.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create service type: "+err.Error())
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusCreated, serviceType)
}

// GetServiceTypeByID godoc
// @Summary Get a service type by ID
// @Description Retrieves details of a specific service type by its ID.
// @Tags ServiceTypes
// @Produce json
// @Param id path int true "Service Type ID" Format(uint)
// @Success 200 {object} models.ServiceType "Successfully retrieved service type"
// @Failure 400 {object} models.ErrorResponse "Invalid ID format"
// @Failure 404 {object} models.ErrorResponse "Service type not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /service-types/{id} [get]
func (h *ServiceHandler) GetServiceTypeByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Error("Invalid service type ID format", zap.String("idParam", idParam), zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	serviceType, err := h.service.GetServiceTypeByID(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Error in GetServiceTypeByID", zap.Error(err), zap.Uint64("id", id))
		if errors.Is(err, models.ErrServiceTypeNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, models.ErrServiceTypeNotFound.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve service type: "+err.Error())
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, serviceType)
}

// ListServiceTypes godoc
// @Summary List service types
// @Description Retrieves a paginated list of service types, optionally filtered by name.
// @Tags ServiceTypes
// @Produce json
// @Param page query int false "Page number for pagination (default: 1)"
// @Param pageSize query int false "Number of items per page (default: 10, max: 100)"
// @Param name query string false "Filter by service type name (case-insensitive, partial match)"
// @Success 200 {object} models.PaginatedData{items=[]models.ServiceType} "Successfully retrieved list of service types"
// @Failure 400 {object} models.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /service-types [get]
func (h *ServiceHandler) ListServiceTypes(c *gin.Context) {
	var params models.ServiceTypeListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Error("Failed to bind query for ListServiceTypes", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	// Service layer handles default pagination values if not provided or invalid
	_, paginatedData, err := h.service.ListServiceTypes(c.Request.Context(), params)
	if err != nil {
		h.logger.Error("Failed to list service types", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve service types")
		return
	}
	// The PaginatedData from service layer already contains the items (serviceTypes) and pagination details.
	// No need to set paginatedData.Items = serviceTypes here as service layer should do that.
	utils.SendSuccessResponse(c, http.StatusOK, paginatedData)
}

// UpdateServiceType godoc
// @Summary Update an existing service type
// @Description Updates an existing service type with the provided details. Only fields present in the request body will be updated.
// @Tags ServiceTypes
// @Accept json
// @Produce json
// @Param id path int true "Service Type ID" Format(uint)
// @Param serviceType body models.UpdateServiceTypeRequest true "Service Type details to update"
// @Success 200 {object} models.ServiceType "Successfully updated service type"
// @Failure 400 {object} models.ErrorResponse "Invalid ID format or request payload"
// @Failure 404 {object} models.ErrorResponse "Service type not found"
// @Failure 409 {object} models.ErrorResponse "Another service type with the new name already exists"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /service-types/{id} [put]
func (h *ServiceHandler) UpdateServiceType(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Error("Invalid service type ID format for update", zap.String("idParam", idParam), zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req models.UpdateServiceTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for UpdateServiceType", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	updatedServiceType, err := h.service.UpdateServiceType(c.Request.Context(), uint(id), req)
	if err != nil {
		h.logger.Error("Failed to update service type", zap.Error(err), zap.Uint64("id", id), zap.Any("request", req))
		if errors.Is(err, models.ErrServiceTypeNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, models.ErrServiceTypeNotFound.Error())
		} else if errors.Is(err, models.ErrServiceTypeNameExists) {
			utils.SendErrorResponse(c, http.StatusConflict, models.ErrServiceTypeNameExists.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update service type: "+err.Error())
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, updatedServiceType)
}

// DeleteServiceType godoc
// @Summary Delete a service type
// @Description Deletes a specific service type by its ID.
// @Tags ServiceTypes
// @Produce json
// @Param id path int true "Service Type ID" Format(uint)
// @Success 204 "Successfully deleted service type (No Content)"
// @Failure 400 {object} models.ErrorResponse "Invalid ID format"
// @Failure 404 {object} models.ErrorResponse "Service type not found"
// @Failure 409 {object} models.ErrorResponse "Service type is in use and cannot be deleted"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /service-types/{id} [delete]
func (h *ServiceHandler) DeleteServiceType(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Error("Invalid service type ID format for delete", zap.String("idParam", idParam), zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	err = h.service.DeleteServiceType(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, models.ErrServiceTypeNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, models.ErrServiceTypeNotFound.Error())
		} else if errors.Is(err, models.ErrServiceTypeInUse) {
			utils.SendErrorResponse(c, http.StatusConflict, models.ErrServiceTypeInUse.Error())
		} else {
			h.logger.Error("Failed to delete service type", zap.Error(err), zap.Uint64("id", id))
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to delete service type: "+err.Error())
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// --- Service Handlers ---

// CreateService godoc
// @Summary Create a new service
// @Description Creates a new service with the provided details.
// @Tags Services
// @Accept json
// @Produce json
// @Param service body models.CreateServiceRequest true "Service to create"
// @Success 201 {object} models.ServiceResponse "Successfully created service"
// @Failure 400 {object} models.ErrorResponse "Invalid request payload or validation error (e.g., invalid ServiceTypeID)"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /services [post]
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req models.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for CreateService", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	serviceResp, err := h.service.CreateService(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create service", zap.Error(err), zap.String("name", req.Name), zap.Any("request", req))
		if errors.Is(err, models.ErrServiceTypeNotFound) { // ServiceTypeID in request not found
			utils.SendErrorResponse(c, http.StatusBadRequest, models.ErrServiceTypeNotFound.Error()+": service_type_id in request not found")
		} else if errors.Is(err, models.ErrServiceNameExists) {
			utils.SendErrorResponse(c, http.StatusConflict, models.ErrServiceNameExists.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create service: "+err.Error())
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusCreated, serviceResp)
}

// GetServiceByID godoc
// @Summary Get a service by ID
// @Description Retrieves details of a specific service by its ID, including its ServiceType.
// @Tags Services
// @Produce json
// @Param id path int true "Service ID" Format(uint)
// @Success 200 {object} models.ServiceResponse "Successfully retrieved service"
// @Failure 400 {object} models.ErrorResponse "Invalid ID format"
// @Failure 404 {object} models.ErrorResponse "Service not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /services/{id} [get]
func (h *ServiceHandler) GetServiceByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Error("Invalid service ID format", zap.String("idParam", idParam), zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	serviceResp, err := h.service.GetServiceByID(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Error in GetServiceByID", zap.Error(err), zap.Uint64("id", id))
		if errors.Is(err, models.ErrServiceNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, models.ErrServiceNotFound.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve service: "+err.Error())
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, serviceResp)
}

// ListServices godoc
// @Summary List services
// @Description Retrieves a paginated list of services, with optional filters.
// @Tags Services
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Items per page (default: 10, max: 100)"
// @Param name query string false "Filter by service name (partial match)"
// @Param status query string false "Filter by service status (e.g., active, inactive)"
// @Param serviceTypeId query int false "Filter by service type ID" Format(uint)
// @Success 200 {object} models.PaginatedData{items=[]models.ServiceResponse} "Successfully retrieved list of services"
// @Failure 400 {object} models.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /services [get]
func (h *ServiceHandler) ListServices(c *gin.Context) {
	var params models.ServiceListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Error("Failed to bind query for ListServices", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	_, paginatedData, err := h.service.ListServices(c.Request.Context(), params) // Service responses are in paginatedData.Items
	if err != nil {
		h.logger.Error("Failed to list services", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve services")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, paginatedData)
}

// UpdateService godoc
// @Summary Update an existing service
// @Description Updates an existing service. Only fields in the request body are updated.
// @Tags Services
// @Accept json
// @Produce json
// @Param id path int true "Service ID" Format(uint)
// @Param service body models.UpdateServiceRequest true "Service details to update"
// @Success 200 {object} models.ServiceResponse "Successfully updated service"
// @Failure 400 {object} models.ErrorResponse "Invalid ID or payload (e.g., invalid ServiceTypeID)"
// @Failure 404 {object} models.ErrorResponse "Service not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /services/{id} [put]
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Error("Invalid service ID format for update", zap.String("idParam", idParam), zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req models.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for UpdateService", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	serviceResp, err := h.service.UpdateService(c.Request.Context(), uint(id), req)
	if err != nil {
		h.logger.Error("Failed to update service", zap.Error(err), zap.Uint64("id", id), zap.Any("request", req))
		if errors.Is(err, models.ErrServiceNotFound) { // The service itself to update is not found
			utils.SendErrorResponse(c, http.StatusNotFound, models.ErrServiceNotFound.Error())
		} else if errors.Is(err, models.ErrServiceTypeNotFound) { // The new ServiceTypeID in payload is not found
			utils.SendErrorResponse(c, http.StatusBadRequest, models.ErrServiceTypeNotFound.Error()+": new service_type_id in request not found")
		} else if errors.Is(err, models.ErrServiceNameExists) {
			utils.SendErrorResponse(c, http.StatusConflict, models.ErrServiceNameExists.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update service: "+err.Error())
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, serviceResp)
}

// DeleteService godoc
// @Summary Delete a service
// @Description Deletes a specific service by its ID (soft delete).
// @Tags Services
// @Produce json
// @Param id path int true "Service ID" Format(uint)
// @Success 204 "Successfully deleted service (No Content)"
// @Failure 400 {object} models.ErrorResponse "Invalid ID format"
// @Failure 404 {object} models.ErrorResponse "Service not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /services/{id} [delete]
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Error("Invalid service ID format for delete", zap.String("idParam", idParam), zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	err = h.service.DeleteService(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to delete service", zap.Error(err), zap.Uint64("id", id))
		if errors.Is(err, models.ErrServiceNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, models.ErrServiceNotFound.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to delete service: "+err.Error())
		}
		return
	}

	c.Status(http.StatusNoContent)
}
