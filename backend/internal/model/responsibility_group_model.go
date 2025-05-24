package model

import "time"

// ResponsibilityGroup is a collection of related responsibilities.
type ResponsibilityGroup struct {
	ID               uint             `json:"id" gorm:"primaryKey"`
	Name             string           `json:"name" gorm:"size:100;uniqueIndex;not null"`
	Description      string           `json:"description,omitempty" gorm:"size:255"`
	CreatedAt        time.Time        `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt        time.Time        `json:"updatedAt" gorm:"autoUpdateTime"`
	Responsibilities []Responsibility `json:"responsibilities,omitempty" gorm:"many2many:responsibility_group_responsibilities;"` // Many-to-many relationship
}

// TableName specifies the table name for the ResponsibilityGroup model.
func (ResponsibilityGroup) TableName() string {
	return "responsibility_groups"
}

// ResponsibilityGroupResponsibility is the join table for the many-to-many relationship
// between ResponsibilityGroup and Responsibility.
type ResponsibilityGroupResponsibility struct {
	ResponsibilityGroupID uint `gorm:"primaryKey"`
	ResponsibilityID      uint `gorm:"primaryKey"`
}

// TableName specifies the table name for the join model.
func (ResponsibilityGroupResponsibility) TableName() string {
	return "responsibility_group_responsibilities"
}

// ResponsibilityGroupListParams defines parameters for listing responsibility groups.
type ResponsibilityGroupListParams struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=10"`
	Name     string `form:"name"` // For searching by group name
}

// ResponsibilityGroupDetail represents a responsibility group with its associated responsibilities.
// This can be the same as ResponsibilityGroup if the Responsibilities field is always populated for detail views.
// Or it can be a dedicated struct if more/different fields are needed for the detail view.
type ResponsibilityGroupDetail struct {
	ID               uint             `json:"id"`
	Name             string           `json:"name"`
	Description      string           `json:"description,omitempty"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
	Responsibilities []Responsibility `json:"responsibilities"`
}
