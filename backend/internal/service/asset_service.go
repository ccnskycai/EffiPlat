package service

import (
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/repository"
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AssetService defines the interface for asset related business logic.
type AssetService interface {
	CreateAsset(ctx context.Context, req model.CreateAssetRequest) (*model.Asset, error)
	GetAssetByID(ctx context.Context, id uint) (*model.Asset, error)
	ListAssets(ctx context.Context, params model.AssetListParams) ([]model.Asset, int64, error)
	UpdateAsset(ctx context.Context, id uint, req model.UpdateAssetRequest) (*model.Asset, error)
	DeleteAsset(ctx context.Context, id uint) error
}

type assetServiceImpl struct {
	repo    repository.AssetRepository
	envRepo repository.EnvironmentRepository // For validating EnvironmentID
	logger  *zap.Logger
}

// NewAssetService creates a new instance of AssetService.
func NewAssetService(repo repository.AssetRepository, envRepo repository.EnvironmentRepository, logger *zap.Logger) AssetService {
	return &assetServiceImpl{repo: repo, envRepo: envRepo, logger: logger}
}

func (s *assetServiceImpl) CreateAsset(ctx context.Context, req model.CreateAssetRequest) (*model.Asset, error) {
	s.logger.Info("Attempting to create asset", zap.String("hostname", req.Hostname), zap.String("ipAddress", req.IPAddress))

	// Validate EnvironmentID exists
	_, err := s.envRepo.GetByID(ctx, req.EnvironmentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Warn("Invalid EnvironmentID during asset creation", zap.Uint("environmentId", req.EnvironmentID))
			return nil, fmt.Errorf("environment with ID %d not found: %w", req.EnvironmentID, err) // Consider a more user-friendly error or specific error type
		}
		s.logger.Error("Failed to validate EnvironmentID", zap.Error(err), zap.Uint("environmentId", req.EnvironmentID))
		return nil, fmt.Errorf("validating environment ID: %w", err)
	}

	// Check for existing hostname or IP address (optional, can also be handled by DB unique constraints)
	// existingByHostname, _ := s.repo.GetByHostname(ctx, req.Hostname)
	// if existingByHostname != nil {
	// 	 s.logger.Warn("Asset creation failed: hostname already exists", zap.String("hostname", req.Hostname))
	// 	 return nil, fmt.Errorf("hostname '%s' already exists", req.Hostname) // Consider specific error type
	// }
	// existingByIP, _ := s.repo.GetByIPAddress(ctx, req.IPAddress)
	// if existingByIP != nil {
	// 	 s.logger.Warn("Asset creation failed: IP address already exists", zap.String("ipAddress", req.IPAddress))
	// 	 return nil, fmt.Errorf("IP address '%s' already exists", req.IPAddress) // Consider specific error type
	// }

	asset := &model.Asset{
		Hostname:      req.Hostname,
		IPAddress:     req.IPAddress,
		AssetType:     req.AssetType,
		Status:        model.AssetStatusUnknown, // Default status, can be set from req if allowed
		Description:   req.Description,
		EnvironmentID: req.EnvironmentID,
	}

	if req.Status != "" { // Allow overriding default status if provided in request
		asset.Status = req.Status
	}

	if err := s.repo.Create(ctx, asset); err != nil {
		s.logger.Error("Failed to create asset in repository", zap.Error(err), zap.Any("request", req))
		return nil, fmt.Errorf("creating asset: %w", err)
	}

	s.logger.Info("Asset created successfully", zap.Uint("id", asset.ID), zap.String("hostname", asset.Hostname))
	return asset, nil
}

func (s *assetServiceImpl) GetAssetByID(ctx context.Context, id uint) (*model.Asset, error) {
	asset, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Warn("Asset not found by ID in service", zap.Uint("id", id))
			return nil, err // Or a custom error like model.ErrAssetNotFound
		}
		s.logger.Error("Failed to get asset by ID from repository", zap.Error(err), zap.Uint("id", id))
		return nil, fmt.Errorf("getting asset by ID %d: %w", id, err)
	}
	return asset, nil
}

func (s *assetServiceImpl) ListAssets(ctx context.Context, params model.AssetListParams) ([]model.Asset, int64, error) {
	assets, totalCount, err := s.repo.List(ctx, params)
	if err != nil {
		s.logger.Error("Failed to list assets from repository", zap.Error(err), zap.Any("params", params))
		return nil, 0, fmt.Errorf("listing assets: %w", err)
	}
	return assets, totalCount, nil
}

func (s *assetServiceImpl) UpdateAsset(ctx context.Context, id uint, req model.UpdateAssetRequest) (*model.Asset, error) {
	s.logger.Info("Attempting to update asset", zap.Uint("id", id), zap.Any("request", req))

	existingAsset, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Warn("Asset not found for update", zap.Uint("id", id))
			return nil, err // Or model.ErrAssetNotFound
		}
		s.logger.Error("Failed to get asset for update", zap.Error(err), zap.Uint("id", id))
		return nil, fmt.Errorf("getting asset for update: %w", err)
	}

	updated := false
	if req.Hostname != nil && *req.Hostname != existingAsset.Hostname {
		// Optional: Check for hostname uniqueness if changed
		// existingByHostname, _ := s.repo.GetByHostname(ctx, *req.Hostname)
		// if existingByHostname != nil && existingByHostname.ID != id {
		// 	 return nil, fmt.Errorf("hostname '%s' already exists", *req.Hostname)
		// }
		existingAsset.Hostname = *req.Hostname
		updated = true
	}
	if req.IPAddress != nil && *req.IPAddress != existingAsset.IPAddress {
		// Optional: Check for IP address uniqueness if changed
		// existingByIP, _ := s.repo.GetByIPAddress(ctx, *req.IPAddress)
		// if existingByIP != nil && existingByIP.ID != id {
		// 	 return nil, fmt.Errorf("IP address '%s' already exists", *req.IPAddress)
		// }
		existingAsset.IPAddress = *req.IPAddress
		updated = true
	}
	if req.AssetType != nil && *req.AssetType != existingAsset.AssetType {
		existingAsset.AssetType = *req.AssetType
		updated = true
	}
	if req.Status != nil && *req.Status != existingAsset.Status {
		existingAsset.Status = *req.Status
		updated = true
	}
	if req.Description != nil && *req.Description != existingAsset.Description {
		existingAsset.Description = *req.Description
		updated = true
	}
	if req.EnvironmentID != nil && *req.EnvironmentID != existingAsset.EnvironmentID {
		// Validate new EnvironmentID exists
		_, envErr := s.envRepo.GetByID(ctx, *req.EnvironmentID)
		if envErr != nil {
			if envErr == gorm.ErrRecordNotFound {
				s.logger.Warn("Invalid new EnvironmentID during asset update", zap.Uint("environmentId", *req.EnvironmentID))
				return nil, fmt.Errorf("new environment with ID %d not found: %w", *req.EnvironmentID, envErr)
			}
			s.logger.Error("Failed to validate new EnvironmentID for asset update", zap.Error(envErr), zap.Uint("environmentId", *req.EnvironmentID))
			return nil, fmt.Errorf("validating new environment ID: %w", envErr)
		}
		existingAsset.EnvironmentID = *req.EnvironmentID
		updated = true
	}

	if !updated {
		s.logger.Info("No changes detected for asset update", zap.Uint("id", id))
		return existingAsset, nil // No fields to update
	}

	if err := s.repo.Update(ctx, existingAsset); err != nil {
		s.logger.Error("Failed to update asset in repository", zap.Error(err), zap.Uint("id", id))
		return nil, fmt.Errorf("updating asset: %w", err)
	}

	s.logger.Info("Asset updated successfully in service", zap.Uint("id", existingAsset.ID))
	return existingAsset, nil
}

func (s *assetServiceImpl) DeleteAsset(ctx context.Context, id uint) error {
	// First, check if asset exists to provide a clearer error message
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Warn("Asset not found for deletion", zap.Uint("id", id))
			return err // Or model.ErrAssetNotFound
		}
		s.logger.Error("Failed to get asset for deletion check", zap.Error(err), zap.Uint("id", id))
		return fmt.Errorf("checking asset for deletion: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete asset from repository", zap.Error(err), zap.Uint("id", id))
		return fmt.Errorf("deleting asset: %w", err)
	}
	s.logger.Info("Asset deleted successfully from service", zap.Uint("id", id))
	return nil
}
