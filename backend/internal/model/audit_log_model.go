package model

import (
	"time"
	"gorm.io/gorm"
)

// AuditLog 表示系统中的审计日志记录
type AuditLog struct {
	ID        uint           `json:"id" gorm:"primary_key"`
	UserID    uint           `json:"userId" gorm:"index;not null"` // 执行操作的用户ID
	Username  string         `json:"username"`                     // 执行操作的用户名（冗余存储，便于查询）
	Action    string         `json:"action" gorm:"index;not null"` // 执行的操作（CREATE, UPDATE, DELETE, READ等）
	Resource  string         `json:"resource" gorm:"index;not null"` // 操作的资源类型（User, Role, Asset等）
	ResourceID uint          `json:"resourceId" gorm:"index"`      // 操作的资源ID
	Details    string         `json:"details" gorm:"type:text"`    // 操作详情（JSON格式，包含变更前后的数据）
	IPAddress  string         `json:"ipAddress"`                   // 操作者的IP地址
	UserAgent  string         `json:"userAgent"`                   // 用户代理信息
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// AuditLogResponse 是API响应中使用的审计日志数据结构
type AuditLogResponse struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"userId"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID uint      `json:"resourceId"`
	Details    string    `json:"details"`
	IPAddress  string    `json:"ipAddress"`
	UserAgent  string    `json:"userAgent"`
	CreatedAt  time.Time `json:"createdAt"`
}

// ToResponse 将AuditLog转换为AuditLogResponse
func (al *AuditLog) ToResponse() AuditLogResponse {
	return AuditLogResponse{
		ID:         al.ID,
		UserID:     al.UserID,
		Username:   al.Username,
		Action:     al.Action,
		Resource:   al.Resource,
		ResourceID: al.ResourceID,
		Details:    al.Details,
		IPAddress:  al.IPAddress,
		UserAgent:  al.UserAgent,
		CreatedAt:  al.CreatedAt,
	}
}

// AuditLogQueryParams 包含审计日志查询的所有可能参数
type AuditLogQueryParams struct {
	UserID     *uint   `form:"userId"`
	Action     *string `form:"action"`
	Resource   *string `form:"resource"`
	ResourceID *uint   `form:"resourceId"`
	StartDate  *string `form:"startDate"` // 格式：YYYY-MM-DD
	EndDate    *string `form:"endDate"`   // 格式：YYYY-MM-DD
	Page       int     `form:"page,default=1"`
	PageSize   int     `form:"pageSize,default=10"`
}
