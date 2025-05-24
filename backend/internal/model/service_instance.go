package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ServiceInstanceStatusType defines the possible statuses for a service instance.
type ServiceInstanceStatusType string

const (
	ServiceInstanceStatusRunning   ServiceInstanceStatusType = "running"
	ServiceInstanceStatusStopped   ServiceInstanceStatusType = "stopped"
	ServiceInstanceStatusDeploying ServiceInstanceStatusType = "deploying"
	ServiceInstanceStatusError     ServiceInstanceStatusType = "error"
	ServiceInstanceStatusUnknown   ServiceInstanceStatusType = "unknown"
)

// ServiceInstance represents a deployed instance of a service in a specific environment.
type ServiceInstance struct {
	ID            uint                      `gorm:"primarykey" json:"id"`
	ServiceID     uint                      `json:"serviceId" gorm:"index;not null"`     // Foreign key to Service model
	EnvironmentID uint                      `json:"environmentId" gorm:"index;not null"` // Foreign key to Environment model
	Version       string                    `json:"version" gorm:"not null"`
	Status        ServiceInstanceStatusType `json:"status" gorm:"type:varchar(50);not null;default:'unknown'"`
	Hostname      *string                   `json:"hostname" gorm:"type:varchar(255)"`
	Port          *int                      `json:"port"`
	Config        datatypes.JSONMap         `json:"config" gorm:"type:json"` // Specific configuration for this instance
	DeployedAt    *time.Time                `json:"deployedAt"`              // Timestamp of when this instance was deployed/went live
	CreatedAt     time.Time                 `json:"createdAt"`
	UpdatedAt     time.Time                 `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt            `gorm:"index" json:"-"` // For soft deletes

	// Associations (optional, depending on how you want to load related data)
	// Service      Service     `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	// Environment  Environment `gorm:"foreignKey:EnvironmentID" json:"environment,omitempty"`
}

// TableName specifies the table name for the ServiceInstance model.
func (ServiceInstance) TableName() string {
	return "service_instances"
}

// IsValidStatus checks if the provided status is valid.
func (s ServiceInstanceStatusType) IsValid() bool {
	switch s {
	case ServiceInstanceStatusRunning, ServiceInstanceStatusStopped, ServiceInstanceStatusDeploying, ServiceInstanceStatusError, ServiceInstanceStatusUnknown:
		return true
	}
	return false
}
