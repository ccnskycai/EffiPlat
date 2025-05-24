package repository

import (
	"context"
	"time"

	"EffiPlat/backend/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuditLogRepository 定义了审计日志存储和检索的接口
type AuditLogRepository interface {
	// CreateLog 创建一条新的审计日志记录
	CreateLog(ctx context.Context, log *model.AuditLog) error
	
	// FindLogs 根据查询参数查找审计日志
	FindLogs(ctx context.Context, params model.AuditLogQueryParams) ([]model.AuditLog, int64, error)
	
	// FindLogByID 根据ID查找一条审计日志
	FindLogByID(ctx context.Context, id uint) (*model.AuditLog, error)
}

// AuditLogRepositoryImpl 实现了AuditLogRepository接口
type AuditLogRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAuditLogRepository 创建一个新的AuditLogRepository实例
func NewAuditLogRepository(db *gorm.DB, logger *zap.Logger) AuditLogRepository {
	return &AuditLogRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// CreateLog 实现了AuditLogRepository接口的CreateLog方法
func (r *AuditLogRepositoryImpl) CreateLog(ctx context.Context, log *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindLogs 实现了AuditLogRepository接口的FindLogs方法
func (r *AuditLogRepositoryImpl) FindLogs(ctx context.Context, params model.AuditLogQueryParams) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var count int64
	
	query := r.db.WithContext(ctx).Model(&model.AuditLog{})
	
	// 应用筛选条件
	if params.UserID != nil {
		query = query.Where("user_id = ?", *params.UserID)
	}
	
	if params.Action != nil && *params.Action != "" {
		query = query.Where("action = ?", *params.Action)
	}
	
	if params.Resource != nil && *params.Resource != "" {
		query = query.Where("resource = ?", *params.Resource)
	}
	
	if params.ResourceID != nil {
		query = query.Where("resource_id = ?", *params.ResourceID)
	}
	
	// 日期范围筛选
	if params.StartDate != nil && *params.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", *params.StartDate)
		if err == nil {
			query = query.Where("created_at >= ?", startDate)
		} else {
			r.logger.Warn("Invalid start date format", zap.String("startDate", *params.StartDate), zap.Error(err))
		}
	}
	
	if params.EndDate != nil && *params.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", *params.EndDate)
		if err == nil {
			// 将结束日期设为当天的最后一刻
			endDate = endDate.Add(24*time.Hour - time.Second)
			query = query.Where("created_at <= ?", endDate)
		} else {
			r.logger.Warn("Invalid end date format", zap.String("endDate", *params.EndDate), zap.Error(err))
		}
	}
	
	// 计算总数
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	
	// 分页
	offset := (params.Page - 1) * params.PageSize
	if offset < 0 {
		offset = 0
	}
	
	// 获取记录
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(params.PageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	
	return logs, count, nil
}

// FindLogByID 实现了AuditLogRepository接口的FindLogByID方法
func (r *AuditLogRepositoryImpl) FindLogByID(ctx context.Context, id uint) (*model.AuditLog, error) {
	var log model.AuditLog
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}
