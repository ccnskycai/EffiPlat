package models

import (
	"time"

	"gorm.io/gorm"
)

// ServiceType represents the category or type of a service.
// Examples: "API", "Database", "WebApp", "MessageQueue", "Cache", "Worker"
type ServiceType struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name" binding:"required,min=2,max=100"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TableName specifies the table name for the ServiceType model.
func (ServiceType) TableName() string {
	return "service_types"
}

// ServiceStatus represents the operational status of a service.
type ServiceStatus string

const (
	ServiceStatusActive       ServiceStatus = "active"
	ServiceStatusInactive     ServiceStatus = "inactive"
	ServiceStatusDevelopment  ServiceStatus = "development"
	ServiceStatusMaintenance  ServiceStatus = "maintenance"
	ServiceStatusDeprecated   ServiceStatus = "deprecated"
	ServiceStatusExperimental ServiceStatus = "experimental"
	ServiceStatusUnknown      ServiceStatus = "unknown"
)

// Service represents a manageable IT service or application component.
type Service struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	Name          string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"name" binding:"required,min=2,max=255"`
	Description   string         `gorm:"type:text" json:"description,omitempty"`
	Version       string         `gorm:"type:varchar(50)" json:"version,omitempty"` // e.g., "1.0.2", "beta"
	Status        ServiceStatus  `gorm:"type:varchar(50);not null;default:'unknown'" json:"status,omitempty" binding:"omitempty,oneof=active inactive development maintenance deprecated experimental unknown"`
	ExternalLink  string         `gorm:"type:varchar(2048)" json:"externalLink,omitempty"` // Link to docs, dashboard, etc.
	ServiceTypeID uint           `json:"serviceTypeId" gorm:"index;not null"`
	ServiceType   *ServiceType   `json:"serviceType,omitempty" gorm:"foreignKey:ServiceTypeID"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for the Service model.
func (Service) TableName() string {
	return "services"
}

// --- Request/Response Structs for Service Handler ---

// CreateServiceRequest defines the structure for creating a new service.
type CreateServiceRequest struct {
	Name          string        `json:"name" binding:"required,min=2,max=255"`
	Description   string        `json:"description,omitempty"`
	Version       string        `json:"version,omitempty" binding:"max=50"`
	Status        ServiceStatus `json:"status,omitempty" binding:"omitempty,oneof=active inactive development maintenance deprecated experimental unknown"`
	ExternalLink  string        `json:"externalLink,omitempty" binding:"omitempty,url,max=2048"`
	ServiceTypeID uint          `json:"serviceTypeId" binding:"required,gt=0"`
}

// UpdateServiceRequest defines the structure for updating an existing service.
// All fields are optional.
type UpdateServiceRequest struct {
	Name          *string        `json:"name,omitempty" binding:"omitempty,min=2,max=255"`
	Description   *string        `json:"description,omitempty"`
	Version       *string        `json:"version,omitempty" binding:"omitempty,max=50"`
	Status        *ServiceStatus `json:"status,omitempty" binding:"omitempty,oneof=active inactive development maintenance deprecated experimental unknown"`
	ExternalLink  *string        `json:"externalLink,omitempty" binding:"omitempty,url,max=2048"`
	ServiceTypeID *uint          `json:"serviceTypeId,omitempty" binding:"omitempty,gt=0"`
}

// ServiceResponse defines a standard way to return service data.
// It will include the nested ServiceType information.
type ServiceResponse struct {
	ID           uint          `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	Version      string        `json:"version,omitempty"`
	Status       ServiceStatus `json:"status,omitempty"`
	ExternalLink string        `json:"externalLink,omitempty"`
	ServiceType  *ServiceType  `json:"serviceType,omitempty"` // Embed ServiceType for richer response
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}

// ToServiceResponse converts a Service model to a ServiceResponse.
// It ensures that the ServiceType is also converted if available.
func (s *Service) ToServiceResponse() ServiceResponse {
	resp := ServiceResponse{
		ID:           s.ID,
		Name:         s.Name,
		Description:  s.Description,
		Version:      s.Version,
		Status:       s.Status,
		ExternalLink: s.ExternalLink,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
	if s.ServiceType != nil {
		resp.ServiceType = s.ServiceType // Assuming ServiceType doesn't need a special ToResponse() or is simple enough
	}
	return resp
}

// ServiceListParams defines parameters for listing services with pagination and filtering.
type ServiceListParams struct {
	Page          int    `form:"page,default=1" binding:"omitempty,gt=0"`
	PageSize      int    `form:"pageSize,default=10" binding:"omitempty,gt=0,max=100"`
	Name          string `form:"name" binding:"omitempty,max=255"`
	Status        string `form:"status" binding:"omitempty"` // Allows filtering by status string
	ServiceTypeID uint   `form:"serviceTypeId" binding:"omitempty,gt=0"`
	// Add SortBy and Order if needed
}

// ServiceTypeListParams defines parameters for listing service types with pagination and filtering.
type ServiceTypeListParams struct {
	Page     int    `form:"page,default=1" binding:"omitempty,gt=0"`
	PageSize int    `form:"pageSize,default=10" binding:"omitempty,gt=0,max=100"`
	Name     string `form:"name" binding:"omitempty,max=100"` // Optional: filter by name
}

// --- Request/Response Structs for ServiceType ---

// CreateServiceTypeRequest defines the structure for creating a new service type.
type CreateServiceTypeRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description,omitempty" binding:"max=1000"`
}

// UpdateServiceTypeRequest defines the structure for updating an existing service type.
type UpdateServiceTypeRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}
