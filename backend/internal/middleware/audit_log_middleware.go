package middleware

import (
	"EffiPlat/backend/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

// AuditOperation 表示需要记录的审计操作类型
type AuditOperation struct {
	Action   string // 操作类型：CREATE, UPDATE, DELETE, READ
	Resource string // 资源类型：USER, ROLE, ASSET 等
}

// AuditLogMiddleware 创建一个用于记录审计日志的中间件
func AuditLogMiddleware(auditLogService service.AuditLogService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过不需要审计的路径
		if shouldSkipAudit(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 解析操作类型和资源类型
		op := parseRequestForAuditLog(c)
		if op == nil {
			// 无法识别的操作或资源，跳过审计
			c.Next()
			return
		}

		// 从路径中提取资源ID
		resourceID := extractResourceIDFromPath(c.Request.URL.Path)

		// 如果上下文中已经设置了资源ID，则优先使用
		if resourceIDFromCtx, exists := c.Get("auditResourceID"); exists {
			if id, ok := resourceIDFromCtx.(uint); ok && id > 0 {
				resourceID = id
			}
		}

		// 如果上下文中已经设置了资源类型，则优先使用
		if resourceFromCtx, exists := c.Get("auditResource"); exists {
			if res, ok := resourceFromCtx.(string); ok && res != "" {
				op.Resource = strings.ToUpper(res)
			}
		}

		// 如果上下文中已经设置了操作类型，则优先使用
		if actionFromCtx, exists := c.Get("auditAction"); exists {
			if action, ok := actionFromCtx.(string); ok && action != "" {
				op.Action = strings.ToUpper(action)
			}
		}

		// 先执行请求处理链
		c.Next()

		// 只记录成功的操作（2xx状态码）
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			// 从上下文中获取操作详情（由处理器设置）
			var details interface{}
			if d, exists := c.Get("auditDetails"); exists {
				details = d
			} else {
				// 如果没有设置详情，可以创建一个基本详情
				details = map[string]interface{}{
					"requestPath": c.Request.URL.Path,
					"method": c.Request.Method,
					"status": c.Writer.Status(),
				}
			}

			// 记录审计日志
			if err := auditLogService.LogUserAction(c, op.Action, op.Resource, resourceID, details); err != nil {
				logger.Error("Failed to log audit entry", zap.Error(err))
			}
		}
	}
}

// parseRequestForAuditLog 根据HTTP方法和路径解析出操作类型和资源类型
func parseRequestForAuditLog(c *gin.Context) *AuditOperation {
	method := c.Request.Method
	path := c.Request.URL.Path

	// 忽略不需要审计的路径
	if strings.HasPrefix(path, "/api/v1/auth") || 
	   strings.HasPrefix(path, "/api/v1/docs") ||
	   strings.HasPrefix(path, "/healthz") {
		return nil
	}

	// 解析资源类型
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) < 3 {
		return nil // 不符合 /api/v1/resource 格式
	}

	resource := pathParts[2]
	// 将多词资源转换为单一标识符 (例如：responsibility-groups -> RESPONSIBILITY_GROUP)
	resource = strings.ToUpper(strings.ReplaceAll(resource, "-", "_"))
	// 转换复数形式为单数
	if strings.HasSuffix(resource, "S") && resource != "BUSINESS" { // 特殊处理"BUSINESS"
		resource = resource[:len(resource)-1]
	}

	// 根据HTTP方法确定操作类型
	var action string
	switch method {
	case "POST":
		action = "CREATE"
	case "PUT", "PATCH":
		action = "UPDATE"
	case "DELETE":
		action = "DELETE"
	case "GET":
		action = "READ"
	default:
		return nil // 不记录其他HTTP方法
	}

	return &AuditOperation{
		Action:   action,
		Resource: resource,
	}
}

// extractResourceIDFromPath 从路径中提取资源ID
func extractResourceIDFromPath(path string) uint {
	parts := strings.Split(path, "/")
	// 尝试查找路径中的ID，通常在资源名称之后
	for i, part := range parts {
		if i > 0 && isResourceName(parts[i-1]) {
			if id, err := strconv.ParseUint(part, 10, 32); err == nil {
				return uint(id)
			}
		}
	}
	return 0 // 无法确定资源ID
}

// isResourceName 判断路径部分是否为资源名称
func isResourceName(part string) bool {
	resources := []string{
		"users", "roles", "permissions", "responsibilities", "responsibility-groups",
		"environments", "assets", "services", "service-types", "service-instances", 
		"businesses", "bugs", "audit-logs", "projects", "teams", "settings",
	}
	
	// 检查是否为资源名称
	for _, r := range resources {
		if part == r {
			return true
		}
	}
	
	return false
}

// shouldSkipAudit 判断是否应该跳过审计记录
func shouldSkipAudit(path string) bool {
	// 跳过不需要审计的路径
	skipPrefixes := []string{
		"/api/v1/auth/login",   // 登录路径由专门的审计处理
		"/api/v1/docs",        // 文档路径
		"/healthz",            // 健康检查
		"/metrics",           // 指标路径
		"/api/v1/audit-logs", // 审计日志查询本身不需要被审计
	}

	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}
