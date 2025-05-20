package handlers

import (
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/service"
	"EffiPlat/backend/pkg/response"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AssetHandler handles HTTP requests related to assets.
type AssetHandler struct {
	service service.AssetService
	logger  *zap.Logger
}

// NewAssetHandler creates a new instance of AssetHandler.
func NewAssetHandler(s service.AssetService, l *zap.Logger) *AssetHandler {
	return &AssetHandler{service: s, logger: l}
}

// CreateAsset godoc
// @Summary Create a new asset
// @Description Create a new IT asset (server, VM, etc.)
// @Tags assets
// @Accept json
// @Produce json
// @Param asset body models.CreateAssetRequest true "Asset information"
// @Success 201 {object} response.SuccessResponse{data=models.Asset}
// @Failure 400 {object} response.ErrorResponse "Invalid request payload or validation error"
// @Failure 404 {object} response.ErrorResponse "Environment not found"
// @Failure 409 {object} response.ErrorResponse "Asset with an identical hostname or IP already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /assets [post]
// @Security BearerAuth
func (h *AssetHandler) CreateAsset(c *gin.Context) {
	var req models.CreateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind JSON for CreateAsset", zap.Error(err))
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	asset, err := h.service.CreateAsset(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create asset in service", zap.Error(err), zap.Any("request", req))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "Environment not found for the asset.")
			return
		}
		response.InternalServerError(c, "Failed to create asset: "+err.Error())
		return
	}

	response.Created(c, asset)
}

// GetAssetByID godoc
// @Summary Get an asset by ID
// @Description Retrieve details of a specific asset using its ID
// @Tags assets
// @Produce json
// @Param id path uint true "Asset ID"
// @Success 200 {object} response.SuccessResponse{data=models.Asset}
// @Failure 400 {object} response.ErrorResponse "Invalid asset ID format"
// @Failure 404 {object} response.ErrorResponse "Asset not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /assets/{id} [get]
// @Security BearerAuth
func (h *AssetHandler) GetAssetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid asset ID format in GetAssetByID", zap.String("idStr", idStr), zap.Error(err))
		response.BadRequest(c, "Invalid asset ID format.")
		return
	}

	asset, err := h.service.GetAssetByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.Warn("Asset not found by ID in handler", zap.Uint64("id", id))
			response.NotFound(c, "Asset not found.")
			return
		}
		h.logger.Error("Failed to get asset by ID in service", zap.Error(err), zap.Uint64("id", id))
		response.InternalServerError(c, "Failed to retrieve asset: "+err.Error())
		return
	}

	response.OK(c, asset)
}

// ListAssets godoc
// @Summary List all assets
// @Description Retrieve a paginated list of assets, with optional filters
// @Tags assets
// @Produce json
// @Param page query int false "Page number for pagination (default: 1)"
// @Param pageSize query int false "Number of items per page (default: 10, max: 100)"
// @Param hostname query string false "Filter by hostname (supports partial match)"
// @Param ipAddress query string false "Filter by IP address (supports partial match)"
// @Param assetType query string false "Filter by asset type (e.g., physical_server, virtual_machine)"
// @Param status query string false "Filter by asset status (e.g., online, offline)"
// @Param environmentId query int false "Filter by environment ID"
// @Success 200 {object} response.PaginatedResponse{data=[]models.Asset}
// @Failure 400 {object} response.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /assets [get]
// @Security BearerAuth
func (h *AssetHandler) ListAssets(c *gin.Context) {
	var params models.AssetListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Warn("Failed to bind query for ListAssets", zap.Error(err))
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	// Set defaults if not provided or out of bounds (controller/handler level validation for query params)
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 100 {
		params.PageSize = 10 // Default page size
	}

	assets, totalCount, err := h.service.ListAssets(c.Request.Context(), params)
	if err != nil {
		h.logger.Error("Failed to list assets in service", zap.Error(err), zap.Any("params", params))
		response.InternalServerError(c, "Failed to retrieve assets: "+err.Error())
		return
	}

	response.Paginated(c, assets, totalCount, params.Page, params.PageSize)
}

// UpdateAsset godoc
// @Summary Update an existing asset
// @Description Update details of an existing asset by its ID. Only provided fields are updated.
// @Tags assets
// @Accept json
// @Produce json
// @Param id path uint true "Asset ID"
// @Param asset body models.UpdateAssetRequest true "Asset information to update (all fields optional)"
// @Success 200 {object} response.SuccessResponse{data=models.Asset}
// @Failure 400 {object} response.ErrorResponse "Invalid asset ID format or request payload"
// @Failure 404 {object} response.ErrorResponse "Asset not found or new Environment ID not found"
// @Failure 409 {object} response.ErrorResponse "Updated hostname or IP already exists for another asset"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /assets/{id} [put]
// @Security BearerAuth
func (h *AssetHandler) UpdateAsset(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid asset ID format in UpdateAsset", zap.String("idStr", idStr), zap.Error(err))
		response.BadRequest(c, "Invalid asset ID format.")
		return
	}

	var req models.UpdateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind JSON for UpdateAsset", zap.Error(err), zap.Uint64("id", id))
		response.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	asset, err := h.service.UpdateAsset(c.Request.Context(), uint(id), req)
	if err != nil {
		h.logger.Error("Failed to update asset in service", zap.Error(err), zap.Uint64("id", id), zap.Any("request", req))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "Asset or related entity not found for update.")
			return
		}
		response.InternalServerError(c, "Failed to update asset: "+err.Error())
		return
	}

	response.OK(c, asset)
}

// DeleteAsset godoc
// @Summary Delete an asset by ID
// @Description Marks an asset as deleted (soft delete)
// @Tags assets
// @Produce json
// @Param id path uint true "Asset ID"
// @Success 200 {object} response.SuccessResponse{message=string} "Asset deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid asset ID format"
// @Failure 404 {object} response.ErrorResponse "Asset not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /assets/{id} [delete]
// @Security BearerAuth
func (h *AssetHandler) DeleteAsset(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid asset ID format in DeleteAsset", zap.String("idStr", idStr), zap.Error(err))
		response.BadRequest(c, "Invalid asset ID format.")
		return
	}

	err = h.service.DeleteAsset(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.Warn("Asset not found for deletion in handler", zap.Uint64("id", id))
			response.NotFound(c, "Asset not found.")
			return
		}
		h.logger.Error("Failed to delete asset in service", zap.Error(err), zap.Uint64("id", id))
		response.InternalServerError(c, "Failed to delete asset: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "Asset deleted successfully"})
}
