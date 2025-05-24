package handler

import (
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/service"
	"EffiPlat/backend/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

// AuditLogHandler 处理与审计日志相关的HTTP请求
type AuditLogHandler struct {
	service service.AuditLogService
	logger  *zap.Logger
}

// NewAuditLogHandler 创建一个新的AuditLogHandler实例
func NewAuditLogHandler(service service.AuditLogService, logger *zap.Logger) *AuditLogHandler {
	return &AuditLogHandler{
		service: service,
		logger:  logger,
	}
}

// GetLogs 获取审计日志列表
// @Summary 获取审计日志列表
// @Description 根据查询参数获取审计日志列表
// @Tags audit-logs
// @Accept json
// @Produce json
// @Param userId query int false "用户ID"
// @Param action query string false "操作类型 (CREATE, UPDATE, DELETE, READ)"
// @Param resource query string false "资源类型 (USER, ROLE, ASSET等)"
// @Param resourceId query int false "资源ID"
// @Param startDate query string false "开始日期 (YYYY-MM-DD)"
// @Param endDate query string false "结束日期 (YYYY-MM-DD)"
// @Param page query int false "页码 (默认: 1)"
// @Param pageSize query int false "每页数量 (默认: 10)"
// @Success 200 {object} model.SuccessResponse{data=model.PaginatedData{items=[]model.AuditLogResponse}}
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /audit-logs [get]
func (h *AuditLogHandler) GetLogs(c *gin.Context) {
	var params model.AuditLogQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Error("Failed to bind query params", zap.Error(err))
		utils.SendStandardErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}
	
	logs, total, err := h.service.FindLogs(c.Request.Context(), params)
	if err != nil {
		h.logger.Error("Failed to fetch audit logs", zap.Error(err))
		utils.SendStandardErrorResponse(c, http.StatusInternalServerError, "Failed to fetch audit logs")
		return
	}
	
	// 转换为响应格式
	logResponses := make([]model.AuditLogResponse, len(logs))
	for i, log := range logs {
		logResponses[i] = log.ToResponse()
	}
	
	utils.SendPaginatedSuccessResponse(c, http.StatusOK, "Audit logs retrieved successfully", logResponses, params.Page, params.PageSize, total)
}

// GetLogByID 根据ID获取审计日志
// @Summary 获取单个审计日志
// @Description 根据ID获取单个审计日志的详细信息
// @Tags audit-logs
// @Accept json
// @Produce json
// @Param id path int true "审计日志ID"
// @Success 200 {object} model.SuccessResponse{data=model.AuditLogResponse}
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /audit-logs/{id} [get]
func (h *AuditLogHandler) GetLogByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid audit log ID", zap.String("id", idStr), zap.Error(err))
		utils.SendStandardErrorResponse(c, http.StatusBadRequest, "Invalid audit log ID")
		return
	}
	
	log, err := h.service.FindLogByID(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to fetch audit log", zap.Uint("id", uint(id)), zap.Error(err))
		utils.SendStandardErrorResponse(c, http.StatusNotFound, "Audit log not found")
		return
	}
	
	utils.SendStandardSuccessResponse(c, http.StatusOK, "Audit log retrieved successfully", log.ToResponse())
}
