package handler

import (
	"errors"
	"net/http"
	"strconv"

	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/service"
	"EffiPlat/backend/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ServiceHandler handles HTTP requests for Service and ServiceType resources.
// It uses the ServiceService to interact with the business logic layer.
type ServiceHandler struct {
	service      service.ServiceService
	auditService service.AuditLogService
	logger       *zap.Logger
}

// NewServiceHandler creates a new instance of ServiceHandler.
func NewServiceHandler(svc service.ServiceService, auditSvc service.AuditLogService, logger *zap.Logger) *ServiceHandler {
	return &ServiceHandler{
		service:      svc,
		auditService: auditSvc,
		logger:       logger,
	}
}

// --- ServiceType Handlers ---

// CreateServiceType godoc
// @Summary Create a new service type
// @Description Creates a new service type with the provided details.
// @Tags ServiceTypes
// @Accept json
// @Produce json
// @Param serviceType body model.CreateServiceTypeRequest true "Service Type to create"
// @Success 201 {object} model.ServiceType "Successfully created service type"
// @Failure 400 {object} model.ErrorResponse "Invalid request payload or validation error"
// @Failure 409 {object} model.ErrorResponse "Service type with this name already exists"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /service-types [post]
func (h *ServiceHandler) CreateServiceType(c *gin.Context) {
	var req model.CreateServiceTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for CreateServiceType", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Explicit validation for required fields
	if req.Name == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Name is required")
		return
	}

	serviceType, err := h.service.CreateServiceType(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create service type", zap.Error(err), zap.String("name", req.Name))
		if errors.Is(err, model.ErrServiceTypeNameExists) {
			utils.SendErrorResponse(c, http.StatusConflict, model.ErrServiceTypeNameExists.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create service type: "+err.Error())
		}
		return
	}
	
	// 记录审计日志
	details := map[string]interface{}{
		"name":        serviceType.Name,
		"description": serviceType.Description,
	}
	_ = h.auditService.LogUserAction(c, string(utils.AuditActionCreate), "SERVICE_TYPE", serviceType.ID, details)
	
	utils.SendSuccessResponse(c, http.StatusCreated, serviceType)
}

// GetServiceTypeByID godoc
// @Summary Get a service type by ID
// @Description Retrieves details of a specific service type by its ID.
// @Tags ServiceTypes
// @Produce json
// @Param id path int true "Service Type ID" Format(uint)
// @Success 200 {object} model.ServiceType "Successfully retrieved service type"
// @Failure 400 {object} model.ErrorResponse "Invalid ID format"
// @Failure 404 {object} model.ErrorResponse "Service type not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
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
		if errors.Is(err, model.ErrServiceTypeNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, model.ErrServiceTypeNotFound.Error())
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
// @Success 200 {object} model.PaginatedData{items=[]model.ServiceType} "Successfully retrieved list of service types"
// @Failure 400 {object} model.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /service-types [get]
func (h *ServiceHandler) ListServiceTypes(c *gin.Context) {
	var params model.ServiceTypeListParams
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
// @Param serviceType body model.UpdateServiceTypeRequest true "Service Type details to update"
// @Success 200 {object} model.ServiceType "Successfully updated service type"
// @Failure 400 {object} model.ErrorResponse "Invalid ID format or request payload"
// @Failure 404 {object} model.ErrorResponse "Service type not found"
// @Failure 409 {object} model.ErrorResponse "Another service type with the new name already exists"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /service-types/{id} [put]
func (h *ServiceHandler) UpdateServiceType(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Error("Invalid service type ID format for update", zap.String("idParam", idParam), zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req model.UpdateServiceTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for UpdateServiceType", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// 获取更新前的服务类型数据，用于审计日志
	origServiceType, getErr := h.service.GetServiceTypeByID(c.Request.Context(), uint(id))
	if getErr != nil {
		h.logger.Warn("Could not get original service type for audit logging", 
			zap.Error(getErr), zap.Uint64("id", id))
	}
	
	serviceType, err := h.service.UpdateServiceType(c.Request.Context(), uint(id), req)
	if err != nil {
		h.logger.Error("Failed to update service type", zap.Error(err), zap.Uint64("id", id))
		if errors.Is(err, model.ErrServiceTypeNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, model.ErrServiceTypeNotFound.Error())
		} else if errors.Is(err, model.ErrServiceTypeNameExists) {
			utils.SendErrorResponse(c, http.StatusConflict, model.ErrServiceTypeNameExists.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update service type: "+err.Error())
		}
		return
	}
	
	// 记录审计日志
	details := map[string]interface{}{
		"before": origServiceType,
		"after":  serviceType,
		"changes": req,
	}
	_ = h.auditService.LogUserAction(c, string(utils.AuditActionUpdate), "SERVICE_TYPE", serviceType.ID, details)
	
	utils.SendSuccessResponse(c, http.StatusOK, serviceType)
}

// DeleteServiceType godoc
// @Summary Delete a service type
// @Description Deletes a specific service type by its ID.
// @Tags ServiceTypes
// @Produce json
// @Param id path int true "Service Type ID" Format(uint)
// @Success 204 "Successfully deleted service type (No Content)"
// @Failure 400 {object} model.ErrorResponse "Invalid ID format"
// @Failure 404 {object} model.ErrorResponse "Service type not found"
// @Failure 409 {object} model.ErrorResponse "Service type is in use and cannot be deleted"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /service-types/{id} [delete]
func (h *ServiceHandler) DeleteServiceType(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Error("Invalid service type ID format for delete", zap.String("idParam", idParam), zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// 获取要删除的服务类型数据，用于审计日志
	serviceType, getErr := h.service.GetServiceTypeByID(c.Request.Context(), uint(id))
	if getErr != nil {
		h.logger.Warn("Could not get service type for audit logging before deletion", 
			zap.Error(getErr), zap.Uint64("id", id))
	}
	
	err = h.service.DeleteServiceType(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to delete service type", zap.Error(err), zap.Uint64("id", id))
		if errors.Is(err, model.ErrServiceTypeNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, model.ErrServiceTypeNotFound.Error())
		} else if errors.Is(err, model.ErrServiceTypeInUse) {
			utils.SendErrorResponse(c, http.StatusConflict, model.ErrServiceTypeInUse.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to delete service type: "+err.Error())
		}
		return
	}
	
	// 记录审计日志
	if serviceType != nil {
		details := map[string]interface{}{
			"deletedServiceType": serviceType,
		}
		_ = h.auditService.LogUserAction(c, string(utils.AuditActionDelete), "SERVICE_TYPE", uint(id), details)
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
// @Param service body model.CreateServiceRequest true "Service to create"
// @Success 201 {object} model.ServiceResponse "Successfully created service"
// @Failure 400 {object} model.ErrorResponse "Invalid request payload or validation error (e.g., invalid ServiceTypeID)"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /services [post]
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req model.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for CreateService", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Explicit validation for required fields
	if req.Name == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Name is required")
		return
	}

	if req.ServiceTypeID == 0 {
		utils.SendErrorResponse(c, http.StatusBadRequest, "ServiceTypeID is required")
		return
	}

	// Validate status if provided
	if req.Status != "" {
		valid := false
		validStatuses := []model.ServiceStatus{
			model.ServiceStatusActive,
			model.ServiceStatusInactive,
			model.ServiceStatusDevelopment,
			model.ServiceStatusMaintenance,
			model.ServiceStatusDeprecated,
			model.ServiceStatusExperimental,
			model.ServiceStatusUnknown,
		}
		for _, s := range validStatuses {
			if req.Status == s {
				valid = true
				break
			}
		}
		if !valid {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid status value")
			return
		}
	}

	serviceResp, err := h.service.CreateService(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create service", zap.Error(err), zap.Any("request", req))
		if errors.Is(err, model.ErrServiceTypeNotFound) { // ServiceTypeID in request not found
			utils.SendErrorResponse(c, http.StatusBadRequest, model.ErrServiceTypeNotFound.Error()+": service_type_id in request not found")
		} else if errors.Is(err, model.ErrServiceNameExists) {
			utils.SendErrorResponse(c, http.StatusConflict, model.ErrServiceNameExists.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create service: "+err.Error())
		}
		return
	}
	
	// 记录审计日志
	details := map[string]interface{}{
		"name":          serviceResp.Name,
		"description":   serviceResp.Description,
		"version":       serviceResp.Version,
		"serviceTypeId": serviceResp.ServiceTypeID,
		"serviceType":   serviceResp.ServiceType,
		"status":        serviceResp.Status,
		"externalLink":  serviceResp.ExternalLink,
	}
	_ = h.auditService.LogUserAction(c, string(utils.AuditActionCreate), "SERVICE", serviceResp.ID, details)
	
	utils.SendSuccessResponse(c, http.StatusCreated, serviceResp)
}

// GetServiceByID godoc
// @Summary Get a service by ID
// @Description Retrieves details of a specific service by its ID, including its ServiceType.
// @Tags Services
// @Produce json
// @Param id path int true "Service ID" Format(uint)
// @Success 200 {object} model.ServiceResponse "Successfully retrieved service"
// @Failure 400 {object} model.ErrorResponse "Invalid ID format"
// @Failure 404 {object} model.ErrorResponse "Service not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
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
		if errors.Is(err, model.ErrServiceNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, model.ErrServiceNotFound.Error())
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
// @Success 200 {object} model.PaginatedData{items=[]model.ServiceResponse} "Successfully retrieved list of services"
// @Failure 400 {object} model.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /services [get]
func (h *ServiceHandler) ListServices(c *gin.Context) {
	var params model.ServiceListParams
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
// @Param service body model.UpdateServiceRequest true "Service details to update"
// @Success 200 {object} model.ServiceResponse "Successfully updated service"
// @Failure 400 {object} model.ErrorResponse "Invalid ID or payload (e.g., invalid ServiceTypeID)"
// @Failure 404 {object} model.ErrorResponse "Service not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /services/{id} [put]
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Error("Invalid service ID format for update", zap.String("idParam", idParam), zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}
	
	// 获取原始服务数据用于审计日志
	origService, getErr := h.service.GetServiceByID(c.Request.Context(), uint(id))
	if getErr != nil {
		h.logger.Warn("Could not get original service for audit logging", 
			zap.Uint64("id", id), zap.Error(getErr))
	}

	var req model.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON for UpdateService", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Validate fields if provided
	if req.Name != nil && *req.Name == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Name cannot be empty")
		return
	}

	if req.Status != nil {
		valid := false
		validStatuses := []model.ServiceStatus{
			model.ServiceStatusActive,
			model.ServiceStatusInactive,
			model.ServiceStatusDevelopment,
			model.ServiceStatusMaintenance,
			model.ServiceStatusDeprecated,
			model.ServiceStatusExperimental,
			model.ServiceStatusUnknown,
		}
		for _, s := range validStatuses {
			if *req.Status == s {
				valid = true
				break
			}
		}
		if !valid {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid status value")
			return
		}
	}

	serviceResp, err := h.service.UpdateService(c.Request.Context(), uint(id), req)
	if err != nil {
		h.logger.Error("Failed to update service", zap.Error(err), zap.Uint64("id", id), zap.Any("request", req))
		if errors.Is(err, model.ErrServiceNotFound) { // The service itself to update is not found
			utils.SendErrorResponse(c, http.StatusNotFound, model.ErrServiceNotFound.Error())
		} else if errors.Is(err, model.ErrServiceTypeNotFound) { // The new ServiceTypeID in payload is not found
			utils.SendErrorResponse(c, http.StatusBadRequest, model.ErrServiceTypeNotFound.Error()+": new service_type_id in request not found")
		} else if errors.Is(err, model.ErrServiceNameExists) {
			utils.SendErrorResponse(c, http.StatusConflict, model.ErrServiceNameExists.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update service: "+err.Error())
		}
		return
	}
	
	// 记录审计日志
	details := map[string]interface{}{
		"before":  origService,
		"after":   serviceResp,
		"changes": req,
	}
	_ = h.auditService.LogUserAction(c, string(utils.AuditActionUpdate), "SERVICE", serviceResp.ID, details)

	utils.SendSuccessResponse(c, http.StatusOK, serviceResp)
}

// DeleteService godoc
// @Summary Delete a service
// @Description Deletes a specific service by its ID (soft delete).
// @Tags Services
// @Produce json
// @Param id path int true "Service ID" Format(uint)
// @Success 204 "Successfully deleted service (No Content)"
// @Failure 400 {object} model.ErrorResponse "Invalid ID format"
// @Failure 404 {object} model.ErrorResponse "Service not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /services/{id} [delete]
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Error("Invalid service ID format for delete", zap.String("idParam", idParam), zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// 获取要删除的服务数据用于审计日志
	service, getErr := h.service.GetServiceByID(c.Request.Context(), uint(id))
	if getErr != nil {
		h.logger.Warn("Could not get service for audit logging before deletion", 
			zap.Uint64("id", id), zap.Error(getErr))
	}
	
	err = h.service.DeleteService(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to delete service", zap.Error(err), zap.Uint64("id", id))
		if errors.Is(err, model.ErrServiceNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, model.ErrServiceNotFound.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to delete service: "+err.Error())
		}
		return
	}
	
	// 记录审计日志
	if service != nil {
		details := map[string]interface{}{
			"deletedService": service,
		}
		_ = h.auditService.LogUserAction(c, string(utils.AuditActionDelete), "SERVICE", uint(id), details)
	}

	c.Status(http.StatusNoContent)
}
