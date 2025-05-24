package model

import (
	"time"

	"gorm.io/gorm"
)

// BugStatusType defines the possible statuses for a bug
type BugStatusType string

const (
	BugStatusOpen       BugStatusType = "OPEN"
	BugStatusInProgress BugStatusType = "IN_PROGRESS"
	BugStatusResolved   BugStatusType = "RESOLVED"
	BugStatusClosed     BugStatusType = "CLOSED"
	BugStatusReopened   BugStatusType = "REOPENED"
)

// BugPriorityType defines the priority levels for a bug
type BugPriorityType string

const (
	BugPriorityLow    BugPriorityType = "LOW"
	BugPriorityMedium BugPriorityType = "MEDIUM"
	BugPriorityHigh   BugPriorityType = "HIGH"
	BugPriorityUrgent BugPriorityType = "URGENT"
)

// Bug represents a bug or issue reported in the system.
type Bug struct {
	ID          uint            `gorm:"primarykey" json:"id"`
	Title       string          `gorm:"type:varchar(255);not null" json:"title" binding:"required,min=5,max=255"`
	Description string          `gorm:"type:text" json:"description"`
	Status      BugStatusType   `gorm:"type:varchar(50);default:'OPEN';not null" json:"status" binding:"required"`
	Priority    BugPriorityType `gorm:"type:varchar(50);default:'MEDIUM';not null" json:"priority" binding:"required"`
	ReporterID  *uint           `json:"reporterId"` // Optional: Link to user who reported it
	AssigneeID  *uint           `json:"assigneeId"` // Optional: Link to user assigned to fix it
	// ProjectID   *uint           `json:"projectId"`  // Optional: If bugs are tied to projects
	// EnvironmentID *uint        `json:"environmentId"` // Optional: Environment where bug was found
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName specifies the table name for the Bug model.
func (Bug) TableName() string {
	return "bugs"
}

// --- Request/Response Structs for Bug Handler ---

// CreateBugRequest defines the structure for creating a new bug.
type CreateBugRequest struct {
	Title       string          `json:"title" validate:"required,min=5,max=255"`
	Description string          `json:"description" validate:"omitempty"`
	Status      BugStatusType   `json:"status" validate:"required,enum=OPEN|IN_PROGRESS|RESOLVED|CLOSED|REOPENED"`
	Priority    BugPriorityType `json:"priority" validate:"required,enum=LOW|MEDIUM|HIGH|URGENT"`
	ReporterID  *uint           `json:"reporterId" validate:"omitempty,gt=0"`
	AssigneeID  *uint           `json:"assigneeId" validate:"omitempty,gt=0"`
	// ProjectID   *uint        `json:"projectId" validate:"omitempty,gt=0"`
	// EnvironmentID *uint     `json:"environmentId" validate:"omitempty,gt=0"`
}

// UpdateBugRequest defines the structure for updating an existing bug.
type UpdateBugRequest struct {
	Title       *string          `json:"title" validate:"omitempty,min=5,max=255"`
	Description *string          `json:"description" validate:"omitempty"`
	Status      *BugStatusType   `json:"status" validate:"omitempty,enum=OPEN|IN_PROGRESS|RESOLVED|CLOSED|REOPENED"`
	Priority    *BugPriorityType `json:"priority" validate:"omitempty,enum=LOW|MEDIUM|HIGH|URGENT"`
	AssigneeID  *uint            `json:"assigneeId" validate:"omitempty,gt=0"`
	// ProjectID   *uint         `json:"projectId" validate:"omitempty,gt=0"`
	// EnvironmentID *uint      `json:"environmentId" validate:"omitempty,gt=0"`
}

// BugResponse defines a standard way to return bug data.
type BugResponse struct {
	ID          uint            `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Status      BugStatusType   `json:"status"`
	Priority    BugPriorityType `json:"priority"`
	ReporterID  *uint           `json:"reporterId,omitempty"`
	AssigneeID  *uint           `json:"assigneeId,omitempty"`
	// ProjectID   *uint        `json:"projectId,omitempty"`
	// EnvironmentID *uint     `json:"environmentId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BugListParams defines parameters for listing bugs.
type BugListParams struct {
	Page       int             `form:"page,default=1"`
	PageSize   int             `form:"pageSize,default=10"`
	Title      string          `form:"title"`      // Search by title
	Status     BugStatusType   `form:"status"`     // Filter by status
	Priority   BugPriorityType `form:"priority"`   // Filter by priority
	AssigneeID *uint           `form:"assigneeId"` // Filter by assignee
	ReporterID *uint           `form:"reporterId"` // Filter by reporter
	// ProjectID   *uint        `form:"projectId"`   // Filter by project
	// EnvironmentID *uint     `form:"environmentId"`// Filter by environment
}

// ToBugResponse converts a Bug model to a BugResponse.
func (b *Bug) ToBugResponse() BugResponse {
	return BugResponse{
		ID:          b.ID,
		Title:       b.Title,
		Description: b.Description,
		Status:      b.Status,
		Priority:    b.Priority,
		ReporterID:  b.ReporterID,
		AssigneeID:  b.AssigneeID,
		// ProjectID:   b.ProjectID,
		// EnvironmentID: b.EnvironmentID,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}

// ToBugResponses converts a slice of Bug models to a slice of BugResponse.
func ToBugResponses(bugs []*Bug) []BugResponse {
	responses := make([]BugResponse, len(bugs))
	for i, b := range bugs {
		responses[i] = b.ToBugResponse()
	}
	return responses
}
