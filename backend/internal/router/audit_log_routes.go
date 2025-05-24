package router

import (
	"EffiPlat/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

// auditLogRoutes 注册审计日志相关的路由
func auditLogRoutes(rg *gin.RouterGroup, auditLogHdlr *handler.AuditLogHandler) {
	{
		rg.GET("", auditLogHdlr.GetLogs)            // GET /api/v1/audit-logs
		rg.GET("/:id", auditLogHdlr.GetLogByID)     // GET /api/v1/audit-logs/{id}
	}
}
