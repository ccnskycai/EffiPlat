package handler

import (
	"EffiPlat/backend/internal/service"
	"EffiPlat/backend/internal/utils"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// BusinessHandler 负责处理业务相关的 HTTP 请求
type BusinessHandler struct {
	service service.BusinessService
	logger  *zap.Logger
}

// NewBusinessHandler 创建一个新的 BusinessHandler 实例
func NewBusinessHandler(service service.BusinessService, logger *zap.Logger) *BusinessHandler {
	return &BusinessHandler{service: service, logger: logger}
}

// CreateBusiness godoc
// @Summary Create a new business
// @Description Create a new business with the given details.
// @Tags businesses
// @Accept json
// @Produce json
// @Param business body service.BusinessInputDTO true "Business information"
// @Success 201 {object} models.SuccessResponse{data=service.BusinessOutputDTO} "Business created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request payload or validation error"
// @Failure 409 {object} models.ErrorResponse "Business with this name already exists"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /businesses [post]
func (h *BusinessHandler) CreateBusiness(c *gin.Context) {
	var input service.BusinessInputDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Warn("Failed to bind JSON for CreateBusiness", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	createdBusiness, err := h.service.CreateBusiness(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("Failed to create business", zap.Error(err), zap.Any("input", input))
		if errors.Is(err, utils.ErrBadRequest) {
			utils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		} else if errors.Is(err, utils.ErrAlreadyExists) {
			utils.SendErrorResponse(c, http.StatusConflict, err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create business")
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusCreated, createdBusiness, "Business created successfully")
}

// GetBusinessByID godoc
// @Summary Get a business by its ID
// @Description Retrieve details of a specific business using its ID.
// @Tags businesses
// @Produce json
// @Param businessId path int true "Business ID"
// @Success 200 {object} models.SuccessResponse{data=service.BusinessOutputDTO} "Business details"
// @Failure 400 {object} models.ErrorResponse "Invalid business ID format"
// @Failure 404 {object} models.ErrorResponse "Business not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /businesses/{businessId} [get]
func (h *BusinessHandler) GetBusinessByID(c *gin.Context) {
	idStr := c.Param("businessId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid business ID format", zap.String("businessId", idStr), zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid business ID format")
		return
	}

	business, err := h.service.GetBusinessByID(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get business by ID", zap.Uint64("id", id), zap.Error(err))
		if errors.Is(err, utils.ErrNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, "Business not found")
		} else if errors.Is(err, utils.ErrBadRequest) {
			utils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve business")
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, business)
}

// ListBusinesses godoc
// @Summary List businesses
// @Description Get a list of businesses with optional filters and pagination.
// @Tags businesses
// @Produce json
// @Param page query int false "Page number for pagination (default: 1)"
// @Param pageSize query int false "Number of items per page (default: 10)"
// @Param name query string false "Filter by business name (partial match)"
// @Param status query string false "Filter by business status (e.g., active, inactive)"
// @Param owner query string false "Filter by business owner (partial match)"
// @Param sortBy query string false "Field to sort by (e.g., name, createdAt, status, owner). Default: createdAt"
// @Param order query string false "Sort order (asc or desc). Default: desc"
// @Success 200 {object} models.PaginatedResponse{data=[]service.BusinessOutputDTO} "List of businesses"
// @Failure 400 {object} models.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /businesses [get]
func (h *BusinessHandler) ListBusinesses(c *gin.Context) {
	var params service.ListBusinessesParamsDTO
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Warn("Failed to bind query parameters for ListBusinesses", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	businesses, total, err := h.service.ListBusinesses(c.Request.Context(), params)
	if err != nil {
		h.logger.Error("Failed to list businesses", zap.Any("params", params), zap.Error(err))
		if errors.Is(err, utils.ErrBadRequest) {
			utils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to list businesses")
		}
		return
	}

	utils.SendPaginatedSuccessResponse(c, http.StatusOK, "Businesses listed successfully", businesses, params.Page, params.PageSize, total)
}

// UpdateBusiness godoc
// @Summary Update an existing business
// @Description Update details of an existing business by its ID.
// @Tags businesses
// @Accept json
// @Produce json
// @Param businessId path int true "Business ID"
// @Param business body service.BusinessInputDTO true "Business information to update"
// @Success 200 {object} models.SuccessResponse{data=service.BusinessOutputDTO} "Business updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid business ID format or request payload"
// @Failure 404 {object} models.ErrorResponse "Business not found"
// @Failure 409 {object} models.ErrorResponse "Another business with the new name already exists"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /businesses/{businessId} [put]
func (h *BusinessHandler) UpdateBusiness(c *gin.Context) {
	idStr := c.Param("businessId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid business ID format for UpdateBusiness", zap.String("businessId", idStr), zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid business ID format")
		return
	}

	var input service.BusinessInputDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Warn("Failed to bind JSON for UpdateBusiness", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	updatedBusiness, err := h.service.UpdateBusiness(c.Request.Context(), uint(id), input)
	if err != nil {
		h.logger.Error("Failed to update business", zap.Uint64("id", id), zap.Any("input", input), zap.Error(err))
		if errors.Is(err, utils.ErrNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, "Business not found")
		} else if errors.Is(err, utils.ErrBadRequest) {
			utils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		} else if errors.Is(err, utils.ErrAlreadyExists) {
			utils.SendErrorResponse(c, http.StatusConflict, err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update business")
		}
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, updatedBusiness, "Business updated successfully")
}

// DeleteBusiness godoc
// @Summary Delete a business
// @Description Delete a specific business by its ID.
// @Tags businesses
// @Produce json
// @Param businessId path int true "Business ID"
// @Success 204 {object} models.SuccessResponse "Business deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid business ID format"
// @Failure 404 {object} models.ErrorResponse "Business not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /businesses/{businessId} [delete]
func (h *BusinessHandler) DeleteBusiness(c *gin.Context) {
	idStr := c.Param("businessId")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid business ID format for DeleteBusiness", zap.String("businessId", idStr), zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid business ID format")
		return
	}

	err = h.service.DeleteBusiness(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to delete business", zap.Uint64("id", id), zap.Error(err))
		if errors.Is(err, utils.ErrNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, "Business not found")
		} else if errors.Is(err, utils.ErrBadRequest) {
			utils.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to delete business")
		}
		return
	}

	c.Status(http.StatusNoContent)
}
