package handler

import (
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/service"
	"EffiPlat/backend/internal/utils"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AssetHandler handles HTTP requests related to assets.
type AssetHandler struct {
	service      service.AssetService
	auditService service.AuditLogService
	logger       *zap.Logger
}

// NewAssetHandler creates a new instance of AssetHandler.
func NewAssetHandler(s service.AssetService, a service.AuditLogService, l *zap.Logger) *AssetHandler {
	return &AssetHandler{service: s, auditService: a, logger: l}
}

// CreateAsset godoc
// @Summary Create a new asset
// @Description Create a new IT asset (server, VM, etc.)
// @Tags assets
// @Accept json
// @Produce json
// @Param asset body model.CreateAssetRequest true "Asset information"
// @Success 201 {object} utils.SuccessResponse{data=model.Asset}
// @Failure 400 {object} utils.ErrorResponse "Invalid request payload or validation error"
// @Failure 404 {object} utils.ErrorResponse "Environment not found"
// @Failure 409 {object} utils.ErrorResponse "Asset with an identical hostname or IP already exists"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /assets [post]
// @Security BearerAuth
func (h *AssetHandler) CreateAsset(c *gin.Context) {
	var req model.CreateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind JSON for CreateAsset", zap.Error(err))
		utils.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	// Explicit validation for required fields
	if req.Hostname == "" {
		utils.BadRequest(c, "Hostname is required")
		return
	}

	if req.EnvironmentID == 0 {
		utils.BadRequest(c, "EnvironmentID is required")
		return
	}

	asset, err := h.service.CreateAsset(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create asset in service", zap.Error(err), zap.Any("request", req))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.NotFound(c, "Environment not found for the asset.")
			return
		}
		utils.InternalServerError(c, "Failed to create asset: "+err.Error())
		return
	}
	
	// 记录审计日志
	details := map[string]interface{}{
		"hostname":      asset.Hostname,
		"ipAddress":     asset.IPAddress,
		"assetType":     asset.AssetType,
		"status":        asset.Status,
		"environmentID": asset.EnvironmentID,
		"description":   asset.Description,
	}
	_ = h.auditService.LogUserAction(c, string(utils.AuditActionCreate), "ASSET", asset.ID, details)
	
	utils.Created(c, asset)
}

// GetAssetByID godoc
// @Summary Get an asset by ID
// @Description Retrieve details of a specific asset using its ID
// @Tags assets
// @Produce json
// @Param id path uint true "Asset ID"
// @Success 200 {object} utils.SuccessResponse{data=model.Asset}
// @Failure 400 {object} utils.ErrorResponse "Invalid asset ID format"
// @Failure 404 {object} utils.ErrorResponse "Asset not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /assets/{id} [get]
// @Security BearerAuth
func (h *AssetHandler) GetAssetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid asset ID format in GetAssetByID", zap.String("idStr", idStr), zap.Error(err))
		utils.BadRequest(c, "Invalid asset ID format.")
		return
	}

	asset, err := h.service.GetAssetByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.Warn("Asset not found by ID in handler", zap.Uint64("id", id))
			utils.NotFound(c, "Asset not found.")
			return
		}
		h.logger.Error("Failed to get asset by ID in service", zap.Error(err), zap.Uint64("id", id))
		utils.InternalServerError(c, "Failed to retrieve asset: "+err.Error())
		return
	}

	utils.OK(c, asset)
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
// @Success 200 {object} utils.PaginatedResponse{data=[]model.Asset}
// @Failure 400 {object} utils.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /assets [get]
// @Security BearerAuth
func (h *AssetHandler) ListAssets(c *gin.Context) {
	var params model.AssetListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Warn("Failed to bind query for ListAssets", zap.Error(err))
		utils.BadRequest(c, "Invalid query parameters: "+err.Error())
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
		utils.InternalServerError(c, "Failed to retrieve assets: "+err.Error())
		return
	}

	utils.Paginated(c, assets, totalCount, params.Page, params.PageSize)
}

// UpdateAsset godoc
// @Summary Update an existing asset
// @Description Update details of an existing asset by its ID. Only provided fields are updated.
// @Tags assets
// @Accept json
// @Produce json
// @Param id path uint true "Asset ID"
// @Param asset body model.UpdateAssetRequest true "Asset information to update (all fields optional)"
// @Success 200 {object} utils.SuccessResponse{data=model.Asset}
// @Failure 400 {object} utils.ErrorResponse "Invalid asset ID format or request payload"
// @Failure 404 {object} utils.ErrorResponse "Asset not found or new Environment ID not found"
// @Failure 409 {object} utils.ErrorResponse "Updated hostname or IP already exists for another asset"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /assets/{id} [put]
// @Security BearerAuth
func (h *AssetHandler) UpdateAsset(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid asset ID format in UpdateAsset", zap.String("idStr", idStr), zap.Error(err))
		utils.BadRequest(c, "Invalid asset ID format.")
		return
	}

	var req model.UpdateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind JSON for UpdateAsset", zap.Error(err), zap.Uint64("id", id))
		utils.BadRequest(c, "Invalid request payload: "+err.Error())
		return
	}

	// 先获取资产原始数据，用于审计日志
	origAsset, getErr := h.service.GetAssetByID(c.Request.Context(), uint(id))
	if getErr != nil {
		// 如果找不到原始资产，不阻止更新操作
		h.logger.Warn("Could not find original asset for audit logging", 
			zap.Uint64("id", id), zap.Error(getErr))
	}
	
	asset, err := h.service.UpdateAsset(c.Request.Context(), uint(id), req)
	if err != nil {
		h.logger.Error("Failed to update asset in service", zap.Error(err), zap.Uint64("id", id), zap.Any("request", req))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.NotFound(c, "Asset or related entity not found for update.")
			return
		}
		utils.InternalServerError(c, "Failed to update asset: "+err.Error())
		return
	}
	
	// 记录审计日志
	details := map[string]interface{}{
		"before": origAsset,
		"after":  asset,
		"changes": req,
	}
	_ = h.auditService.LogUserAction(c, string(utils.AuditActionUpdate), "ASSET", asset.ID, details)
	
	utils.OK(c, asset)
}

// DeleteAsset godoc
// @Summary Delete an asset by ID
// @Description Marks an asset as deleted (soft delete)
// @Tags assets
// @Produce json
// @Param id path uint true "Asset ID"
// @Success 200 {object} utils.SuccessResponse{message=string} "Asset deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid asset ID format"
// @Failure 404 {object} utils.ErrorResponse "Asset not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /assets/{id} [delete]
// @Security BearerAuth
func (h *AssetHandler) DeleteAsset(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid asset ID format in DeleteAsset", zap.String("idStr", idStr), zap.Error(err))
		utils.BadRequest(c, "Invalid asset ID format.")
		return
	}

	// 先获取资产数据，用于审计日志
	asset, getErr := h.service.GetAssetByID(c.Request.Context(), uint(id))
	if getErr != nil {
		// 如果找不到资产，不阻止删除操作
		h.logger.Warn("Could not find asset for audit logging before deletion", 
			zap.Uint64("id", id), zap.Error(getErr))
	}
	
	err = h.service.DeleteAsset(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.Warn("Asset not found for deletion in handler", zap.Uint64("id", id))
			utils.NotFound(c, "Asset not found.")
			return
		}
		h.logger.Error("Failed to delete asset in service", zap.Error(err), zap.Uint64("id", id))
		utils.InternalServerError(c, "Failed to delete asset: "+err.Error())
		return
	}
	
	// 记录审计日志
	if asset != nil {
		details := map[string]interface{}{
			"deletedAsset": asset,
		}
		_ = h.auditService.LogUserAction(c, string(utils.AuditActionDelete), "ASSET", uint(id), details)
	}
	
	utils.OK(c, gin.H{"message": "Asset deleted successfully"})
}
