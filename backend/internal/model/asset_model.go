package model

import (
	"time"

	"gorm.io/gorm"
)

// AssetType represents the type of an asset (e.g., physical server, VM).
type AssetType string

const (
	AssetTypePhysicalServer AssetType = "physical_server"
	AssetTypeVM             AssetType = "virtual_machine"
	AssetTypeCloudHost      AssetType = "cloud_host"
	AssetTypeContainer      AssetType = "container"
	AssetTypeNetworkDevice  AssetType = "network_device"
	AssetTypeStorage        AssetType = "storage"
	AssetTypeOther          AssetType = "other"
)

// AssetStatus represents the status of an asset (e.g., online, offline).
type AssetStatus string

const (
	AssetStatusOnline         AssetStatus = "online"
	AssetStatusOffline        AssetStatus = "offline"
	AssetStatusMaintenance    AssetStatus = "maintenance"
	AssetStatusPending        AssetStatus = "pending"
	AssetStatusDecommissioned AssetStatus = "decommissioned"
	AssetStatusUnknown        AssetStatus = "unknown"
)

// Asset represents a server or other manageable IT asset.
type Asset struct {
	ID            uint         `gorm:"primarykey" json:"id"`
	Hostname      string       `gorm:"type:varchar(255);uniqueIndex;not null" json:"hostname" binding:"required,hostname_rfc1123"`
	IPAddress     string       `gorm:"type:varchar(100);uniqueIndex;not null" json:"ipAddress" binding:"required,ip"` // Could be IPv4 or IPv6
	AssetType     AssetType    `gorm:"type:varchar(50);not null;default:'other'" json:"assetType" binding:"required,oneof=physical_server virtual_machine cloud_host container network_device storage other"`
	Status        AssetStatus  `gorm:"type:varchar(50);not null;default:'unknown'" json:"status" binding:"required,oneof=online offline maintenance pending decommissioned unknown"`
	Description   string       `gorm:"type:text" json:"description"`
	EnvironmentID uint         `json:"environmentId" gorm:"index"` // Foreign key to Environment
	Environment   *Environment `json:"environment,omitempty"`      // Optional: for eager loading
	// More detailed fields can be added later as needed:
	// Location        string         `json:"location"`
	// OperatingSystem string         `json:"operatingSystem"`
	// CPU             string         `json:"cpu"`
	// Memory          string         `json:"memory"` // e.g., "16GB"
	// DiskSpace       string         `json:"diskSpace"` // e.g., "512GB SSD"
	// SerialNumber    string         `gorm:"type:varchar(255);uniqueIndex" json:"serialNumber"`
	// PurchaseDate    *time.Time     `json:"purchaseDate"`
	// WarrantyEndDate *time.Time     `json:"warrantyEndDate"`
	// Tags            string         `gorm:"type:varchar(255)" json:"tags"` // Comma-separated or could be a separate table for many-to-many
	// CustomFields    datatypes.JSON `gorm:"type:jsonb" json:"customFields"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Use "-" to hide from JSON by default
}

// CreateAssetRequest defines the request body for creating a new asset.
type CreateAssetRequest struct {
	Hostname      string      `json:"hostname" binding:"required,hostname_rfc1123,min=3,max=255"`
	IPAddress     string      `json:"ipAddress" binding:"required,ip"`
	AssetType     AssetType   `json:"assetType" binding:"required,oneof=physical_server virtual_machine cloud_host container network_device storage other"`
	Status        AssetStatus `json:"status" binding:"omitempty,oneof=online offline maintenance pending decommissioned unknown"` // Optional on create, defaults in model
	Description   string      `json:"description" binding:"max=1000"`
	EnvironmentID uint        `json:"environmentId" binding:"required,gt=0"` // Must belong to an environment
}

// UpdateAssetRequest defines the request body for updating an existing asset.
// All fields are optional.
type UpdateAssetRequest struct {
	Hostname      *string      `json:"hostname,omitempty" binding:"omitempty,hostname_rfc1123,min=3,max=255"`
	IPAddress     *string      `json:"ipAddress,omitempty" binding:"omitempty,ip"`
	AssetType     *AssetType   `json:"assetType,omitempty" binding:"omitempty,oneof=physical_server virtual_machine cloud_host container network_device storage other"`
	Status        *AssetStatus `json:"status,omitempty" binding:"omitempty,oneof=online offline maintenance pending decommissioned unknown"`
	Description   *string      `json:"description,omitempty" binding:"omitempty,max=1000"`
	EnvironmentID *uint        `json:"environmentId,omitempty" binding:"omitempty,gt=0"`
}

// AssetListParams defines parameters for listing assets with pagination.
type AssetListParams struct {
	Page          int    `form:"page,default=1" binding:"omitempty,gt=0"`
	PageSize      int    `form:"pageSize,default=10" binding:"omitempty,gt=0,max=100"`
	Hostname      string `form:"hostname" binding:"omitempty"`
	IPAddress     string `form:"ipAddress" binding:"omitempty"`
	AssetType     string `form:"assetType" binding:"omitempty"`
	Status        string `form:"status" binding:"omitempty"`
	EnvironmentID uint   `form:"environmentId" binding:"omitempty,gt=0"`
}
