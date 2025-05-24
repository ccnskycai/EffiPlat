package service

import (
	"context"
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/repository"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strings"
)

//go:generate mockgen -source=audit_log_service.go -destination=mocks/mock_audit_log_service.go -package=mocks AuditLogService

// AuditLogService 定义了审计日志服务的接口
type AuditLogService interface {
	// CreateLog 创建一条审计日志
	CreateLog(ctx context.Context, log *model.AuditLog) error
	
	// FindLogs 根据查询参数查找审计日志
	FindLogs(ctx context.Context, params model.AuditLogQueryParams) ([]model.AuditLog, int64, error)
	
	// FindLogByID 根据ID查找一条审计日志
	FindLogByID(ctx context.Context, id uint) (*model.AuditLog, error)
	
	// LogUserAction 记录用户操作，从Gin上下文中获取用户信息
	LogUserAction(c *gin.Context, action, resource string, resourceID uint, details interface{}) error
}

// AuditLogServiceImpl 实现了AuditLogService接口
type AuditLogServiceImpl struct {
	repo   repository.AuditLogRepository
	logger *zap.Logger
}

// NewAuditLogService 创建一个新的AuditLogService实例
func NewAuditLogService(repo repository.AuditLogRepository, logger *zap.Logger) AuditLogService {
	return &AuditLogServiceImpl{
		repo:   repo,
		logger: logger,
	}
}

// CreateLog 实现了AuditLogService接口的CreateLog方法
func (s *AuditLogServiceImpl) CreateLog(ctx context.Context, log *model.AuditLog) error {
	return s.repo.CreateLog(ctx, log)
}

// FindLogs 实现了AuditLogService接口的FindLogs方法
func (s *AuditLogServiceImpl) FindLogs(ctx context.Context, params model.AuditLogQueryParams) ([]model.AuditLog, int64, error) {
	return s.repo.FindLogs(ctx, params)
}

// FindLogByID 实现了AuditLogService接口的FindLogByID方法
func (s *AuditLogServiceImpl) FindLogByID(ctx context.Context, id uint) (*model.AuditLog, error) {
	return s.repo.FindLogByID(ctx, id)
}

// LogUserAction 实现了AuditLogService接口的LogUserAction方法
func (s *AuditLogServiceImpl) LogUserAction(c *gin.Context, action, resource string, resourceID uint, details interface{}) error {
	// 从上下文中获取用户信息
	userID, exists := c.Get("userID")
	if !exists {
		s.logger.Warn("User ID not found in context when logging action",
			zap.String("action", action),
			zap.String("resource", resource),
			zap.Uint("resourceID", resourceID))
		return nil // 不阻止主要操作，即使审计日志记录失败
	}
	
	username, _ := c.Get("username")
	usernameStr, ok := username.(string)
	if !ok {
		usernameStr = "unknown"
	}
	
	// 将详情转换为JSON字符串
	var detailsJSON string
	if details != nil {
		detailsBytes, err := json.Marshal(details)
		if err != nil {
			s.logger.Warn("Failed to marshal details to JSON",
				zap.Error(err),
				zap.String("action", action),
				zap.String("resource", resource))
			detailsJSON = "{}"
		} else {
			detailsJSON = string(detailsBytes)
		}
	} else {
		detailsJSON = "{}"
	}
	
	// 获取IP地址
	ipAddress := c.ClientIP()
	
	// 获取用户代理
	userAgent := c.Request.UserAgent()
	
	// 创建审计日志记录
	log := &model.AuditLog{
		UserID:     userID.(uint),
		Username:   usernameStr,
		Action:     strings.ToUpper(action),
		Resource:   strings.ToUpper(resource),
		ResourceID: resourceID,
		Details:    detailsJSON,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}
	
	// 异步保存日志，不阻塞主要操作
	go func(logEntry *model.AuditLog) {
		if err := s.repo.CreateLog(context.Background(), logEntry); err != nil {
			s.logger.Error("Failed to save audit log",
				zap.Error(err),
				zap.String("action", logEntry.Action),
				zap.String("resource", logEntry.Resource),
				zap.Uint("userID", logEntry.UserID))
		}
	}(log)
	
	return nil
}
