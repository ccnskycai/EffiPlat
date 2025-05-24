package model

import (
	"time"

	"gorm.io/gorm"
)

// BusinessStatusType 定义业务状态的类型
type BusinessStatusType string

const (
	BusinessStatusActive   BusinessStatusType = "active"
	BusinessStatusInactive BusinessStatusType = "inactive"
	// 可以根据需要添加其他状态
)

// IsValid 检查状态是否有效
func (bst BusinessStatusType) IsValid() bool {
	switch bst {
	case BusinessStatusActive, BusinessStatusInactive:
		return true
	}
	return false
}

// Business 表示一个业务或产品线
type Business struct {
	ID          uint               `gorm:"primarykey" json:"id"`
	Name        string             `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`        // 业务/产品线名称
	Description string             `gorm:"type:text" json:"description,omitempty"`                    // 详细描述
	Owner       string             `gorm:"type:varchar(100)" json:"owner,omitempty"`                  // 业务负责人/团队
	Status      BusinessStatusType `gorm:"type:varchar(50);default:'active'" json:"status,omitempty"` // 业务状态 (e.g., active, inactive)
	CreatedAt   time.Time          `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time          `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt     `gorm:"index" json:"-"` // 使用json:"-"来在API响应中隐藏它
}

// TableName 指定 Business 模型在数据库中的表名
func (Business) TableName() string {
	return "businesses"
}
