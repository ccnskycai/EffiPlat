package models

import "time"

// Responsibility represents a specific task or duty.
type Responsibility struct {
	ID          uint                  `json:"id" gorm:"primaryKey"`
	Name        string                `json:"name" gorm:"size:100;uniqueIndex;not null"`
	Description string                `json:"description,omitempty" gorm:"size:255"`
	CreatedAt   time.Time             `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time             `json:"updatedAt" gorm:"autoUpdateTime"`
	Groups      []ResponsibilityGroup `json:"-" gorm:"many2many:responsibility_group_responsibilities;"` // Many-to-many relationship
}

// TableName specifies the table name for the Responsibility model.
func (Responsibility) TableName() string {
	return "responsibilities"
}
