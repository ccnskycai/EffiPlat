package utils

import (
	"github.com/gin-gonic/gin"
)

// AuditActionType defines standard audit action types
type AuditActionType string

const (
	// Audit action types
	AuditActionCreate AuditActionType = "CREATE"
	AuditActionRead   AuditActionType = "READ"
	AuditActionUpdate AuditActionType = "UPDATE"
	AuditActionDelete AuditActionType = "DELETE"
	AuditActionLogin  AuditActionType = "LOGIN"
	AuditActionLogout AuditActionType = "LOGOUT"
)

// SetAuditDetails sets operation details to be captured in audit logs
// This should be called in handlers before completing the operation
func SetAuditDetails(c *gin.Context, details interface{}) {
	c.Set("auditDetails", details)
}

// AuditChangeLog represents a simple structure for tracking changes
type AuditChangeLog struct {
	Before interface{} `json:"before,omitempty"`
	After  interface{} `json:"after,omitempty"`
}

// NewCreateAuditLog creates an audit log entry for creation operations
func NewCreateAuditLog(entity interface{}) interface{} {
	return map[string]interface{}{
		"action": "CREATE",
		"entity": entity,
	}
}

// NewUpdateAuditLog creates an audit log entry for update operations
func NewUpdateAuditLog(before, after interface{}) interface{} {
	return &AuditChangeLog{
		Before: before,
		After:  after,
	}
}

// NewDeleteAuditLog creates an audit log entry for delete operations
func NewDeleteAuditLog(entity interface{}) interface{} {
	return map[string]interface{}{
		"action":     "DELETE",
		"deletedObj": entity,
	}
}

// SetAuditDetailsWithReqRes sets operation details including request and response
func SetAuditDetailsWithReqRes(c *gin.Context, request interface{}, response interface{}) {
	details := map[string]interface{}{
		"request":  request,
		"response": response,
	}
	c.Set("auditDetails", details)
}
