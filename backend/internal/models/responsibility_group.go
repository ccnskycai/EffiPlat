package models

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
