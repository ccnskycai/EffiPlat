package repository

import (
	"EffiPlat/backend/internal/model"
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AssetRepository defines the interface for asset data operations.
type AssetRepository interface {
	Create(ctx context.Context, asset *model.Asset) error
	GetByID(ctx context.Context, id uint) (*model.Asset, error)
	GetByHostname(ctx context.Context, hostname string) (*model.Asset, error)
	GetByIPAddress(ctx context.Context, ipAddress string) (*model.Asset, error)
	List(ctx context.Context, params model.AssetListParams) ([]model.Asset, int64, error)
	Update(ctx context.Context, asset *model.Asset) error
	Delete(ctx context.Context, id uint) error
	// Add other specific query methods if needed, e.g., ListByEnvironmentID
}

type gormAssetRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewGormAssetRepository creates a new instance of AssetRepository using GORM.
func NewGormAssetRepository(db *gorm.DB, logger *zap.Logger) AssetRepository {
	return &gormAssetRepository{db: db, logger: logger}
}

func (r *gormAssetRepository) Create(ctx context.Context, asset *model.Asset) error {
	if err := r.db.WithContext(ctx).Create(asset).Error; err != nil {
		r.logger.Error("Failed to create asset", zap.Error(err), zap.String("hostname", asset.Hostname))
		return fmt.Errorf("creating asset: %w", err)
	}
	r.logger.Info("Asset created successfully", zap.Uint("id", asset.ID), zap.String("hostname", asset.Hostname))
	return nil
}

func (r *gormAssetRepository) GetByID(ctx context.Context, id uint) (*model.Asset, error) {
	var asset model.Asset
	if err := r.db.WithContext(ctx).Preload("Environment").First(&asset, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Warn("Asset not found by ID", zap.Uint("id", id))
			return nil, err // Or a custom not found error
		}
		r.logger.Error("Failed to get asset by ID", zap.Error(err), zap.Uint("id", id))
		return nil, fmt.Errorf("getting asset by ID %d: %w", id, err)
	}
	return &asset, nil
}

func (r *gormAssetRepository) GetByHostname(ctx context.Context, hostname string) (*model.Asset, error) {
	var asset model.Asset
	if err := r.db.WithContext(ctx).Preload("Environment").Where("hostname = ?", hostname).First(&asset).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Warn("Asset not found by hostname", zap.String("hostname", hostname))
			return nil, err
		}
		r.logger.Error("Failed to get asset by hostname", zap.Error(err), zap.String("hostname", hostname))
		return nil, fmt.Errorf("getting asset by hostname %s: %w", hostname, err)
	}
	return &asset, nil
}

func (r *gormAssetRepository) GetByIPAddress(ctx context.Context, ipAddress string) (*model.Asset, error) {
	var asset model.Asset
	if err := r.db.WithContext(ctx).Preload("Environment").Where("ip_address = ?", ipAddress).First(&asset).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Warn("Asset not found by IP address", zap.String("ipAddress", ipAddress))
			return nil, err
		}
		r.logger.Error("Failed to get asset by IP address", zap.Error(err), zap.String("ipAddress", ipAddress))
		return nil, fmt.Errorf("getting asset by IP address %s: %w", ipAddress, err)
	}
	return &asset, nil
}

func (r *gormAssetRepository) List(ctx context.Context, params model.AssetListParams) ([]model.Asset, int64, error) {
	var assets []model.Asset
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&model.Asset{}).Preload("Environment")

	// Apply filters from params
	if params.Hostname != "" {
		query = query.Where("hostname LIKE ?", "%"+params.Hostname+"%")
	}
	if params.IPAddress != "" {
		query = query.Where("ip_address LIKE ?", "%"+params.IPAddress+"%")
	}
	if params.AssetType != "" {
		query = query.Where("asset_type = ?", params.AssetType)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.EnvironmentID > 0 {
		query = query.Where("environment_id = ?", params.EnvironmentID)
	}

	// Count total records for pagination
	if err := query.Count(&totalCount).Error; err != nil {
		r.logger.Error("Failed to count assets", zap.Error(err))
		return nil, 0, fmt.Errorf("counting assets: %w", err)
	}

	// Apply pagination
	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Order("hostname ASC").Find(&assets).Error; err != nil {
		r.logger.Error("Failed to list assets", zap.Error(err))
		return nil, 0, fmt.Errorf("listing assets: %w", err)
	}

	return assets, totalCount, nil
}

func (r *gormAssetRepository) Update(ctx context.Context, asset *model.Asset) error {
	// Ensure ID is set for update
	if asset.ID == 0 {
		r.logger.Error("Attempted to update asset with zero ID")
		return fmt.Errorf("cannot update asset with zero ID")
	}
	// Use .Model(&model.Asset{}).Where("id = ?", asset.ID).Updates(asset) to ensure hooks are run and only non-zero fields are updated if that's the desired behavior for partial updates.
	// For a full replacement (excluding zero values of primitive types by default in .Save), .Save is okay.
	if err := r.db.WithContext(ctx).Save(asset).Error; err != nil {
		r.logger.Error("Failed to update asset", zap.Error(err), zap.Uint("id", asset.ID))
		return fmt.Errorf("updating asset ID %d: %w", asset.ID, err)
	}
	r.logger.Info("Asset updated successfully", zap.Uint("id", asset.ID))
	return nil
}

func (r *gormAssetRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.Asset{}, id).Error; err != nil {
		r.logger.Error("Failed to delete asset", zap.Error(err), zap.Uint("id", id))
		return fmt.Errorf("deleting asset ID %d: %w", id, err)
	}
	r.logger.Info("Asset deleted successfully (soft delete)", zap.Uint("id", id))
	return nil
}
