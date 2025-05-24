package router_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/router"
	"EffiPlat/backend/internal/utils"
	
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuditLogRoutes 测试审计日志API端点
func TestAuditLogRoutes(t *testing.T) {
	// 设置测试环境
	components := router.SetupTestApp(t)
	rtr := components.Router
	db := components.DB
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 创建测试用户并获取管理员令牌
	adminUser, err := router.CreateTestUser(db, "admin@example.com", "password")
	require.NoError(t, err)
	adminToken := router.GetAdminToken(t, components)

	// 创建测试审计日志数据
	createTestAuditLogs(t, db, adminUser.ID)

	// 子测试
	t.Run("GetAuditLogs", func(t *testing.T) {
		testGetAuditLogs(t, rtr, adminToken)
	})

	t.Run("GetAuditLogsWithFilters", func(t *testing.T) {
		testGetAuditLogsWithFilters(t, rtr, adminToken, adminUser.ID)
	})

	t.Run("GetAuditLogByID", func(t *testing.T) {
		testGetAuditLogByID(t, rtr, adminToken)
	})

	t.Run("GetAuditLogByID_NotFound", func(t *testing.T) {
		testGetAuditLogByIDNotFound(t, rtr, adminToken)
	})
}

// createTestAuditLogs 创建测试审计日志记录
func createTestAuditLogs(t *testing.T, db *gorm.DB, userID uint) {
	// 创建5条审计日志记录
	for i := 1; i <= 5; i++ {
		var action string
		var resource string
		var resourceID uint

		// 创建不同类型的审计日志
		switch i % 3 {
		case 0:
			action = string(utils.AuditActionCreate)
			resource = "USER"
			resourceID = uint(10 + i)
		case 1:
			action = string(utils.AuditActionUpdate)
			resource = "ROLE"
			resourceID = uint(20 + i)
		case 2:
			action = string(utils.AuditActionDelete)
			resource = "PERMISSION"
			resourceID = uint(30 + i)
		}

		// 创建审计日志记录
		auditLog := model.AuditLog{
			UserID:     userID,
			Username:   "admin@example.com",
			Action:     action,
			Resource:   resource,
			ResourceID: resourceID,
			Details:    fmt.Sprintf(`{"test": "details %d"}`, i),
			IPAddress:  "127.0.0.1",
			UserAgent:  "Test Agent",
			CreatedAt:  time.Now().Add(-time.Duration(i) * time.Hour), // 不同时间创建
		}

		err := db.Create(&auditLog).Error
		require.NoError(t, err, "Failed to create test audit log")
	}
}

// testGetAuditLogs 测试获取审计日志列表
func testGetAuditLogs(t *testing.T, rtr http.Handler, adminToken string) {
	req, _ := http.NewRequest("GET", "/api/v1/audit-logs", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Items     []model.AuditLogResponse `json:"items"`
			Total     int                      `json:"total"`
			Page      int                      `json:"page"`
			PageSize  int                      `json:"pageSize"`
		} `json:"data"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Code)
	assert.NotEmpty(t, response.Data.Items)
	assert.GreaterOrEqual(t, response.Data.Total, 5) // 至少5条记录
	assert.Equal(t, 1, response.Data.Page)           // 默认第1页
}

// testGetAuditLogsWithFilters 测试带过滤条件的审计日志查询
func testGetAuditLogsWithFilters(t *testing.T, rtr http.Handler, adminToken string, userID uint) {
	// 测试按用户ID过滤
	testUserFilter := func(t *testing.T) {
		url := fmt.Sprintf("/api/v1/audit-logs?userId=%d", userID)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Items []model.AuditLogResponse `json:"items"`
				Total int                      `json:"total"`
			} `json:"data"`
		}

		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.Code)
		assert.NotEmpty(t, response.Data.Items)

		// 验证所有记录的用户ID都匹配
		for _, item := range response.Data.Items {
			assert.Equal(t, userID, item.UserID)
		}
	}

	// 测试按操作类型过滤
	testActionFilter := func(t *testing.T) {
		action := string(utils.AuditActionCreate)
		url := fmt.Sprintf("/api/v1/audit-logs?action=%s", action)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Items []model.AuditLogResponse `json:"items"`
			} `json:"data"`
		}

		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.Code)

		// 验证所有记录的操作类型都匹配
		for _, item := range response.Data.Items {
			assert.Equal(t, action, item.Action)
		}
	}

	// 测试按资源类型过滤
	testResourceFilter := func(t *testing.T) {
		resource := "USER"
		url := fmt.Sprintf("/api/v1/audit-logs?resource=%s", resource)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Items []model.AuditLogResponse `json:"items"`
			} `json:"data"`
		}

		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.Code)

		// 验证所有记录的资源类型都匹配
		for _, item := range response.Data.Items {
			assert.Equal(t, resource, item.Resource)
		}
	}

	// 执行子测试
	t.Run("UserFilter", testUserFilter)
	t.Run("ActionFilter", testActionFilter)
	t.Run("ResourceFilter", testResourceFilter)
}

// testGetAuditLogByID 测试获取单个审计日志详情
func testGetAuditLogByID(t *testing.T, rtr http.Handler, adminToken string) {
	// 先获取审计日志列表，得到一个ID
	req, _ := http.NewRequest("GET", "/api/v1/audit-logs?pageSize=1", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var listResponse struct {
		Data struct {
			Items []struct {
				ID uint `json:"id"`
			} `json:"items"`
		} `json:"data"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &listResponse)
	assert.NoError(t, err)
	assert.NotEmpty(t, listResponse.Data.Items)

	auditLogID := listResponse.Data.Items[0].ID

	// 获取单个审计日志详情
	detailReq, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/audit-logs/%d", auditLogID), nil)
	detailReq.Header.Set("Authorization", "Bearer "+adminToken)

	detailW := httptest.NewRecorder()
	rtr.ServeHTTP(detailW, detailReq)

	assert.Equal(t, http.StatusOK, detailW.Code)

	var detailResponse struct {
		Code    int                     `json:"code"`
		Message string                  `json:"message"`
		Data    model.AuditLogResponse `json:"data"`
	}

	err = json.Unmarshal(detailW.Body.Bytes(), &detailResponse)
	assert.NoError(t, err)
	assert.Equal(t, 0, detailResponse.Code)
	assert.Equal(t, auditLogID, detailResponse.Data.ID)
}

// testGetAuditLogByIDNotFound 测试获取不存在的审计日志
func testGetAuditLogByIDNotFound(t *testing.T, rtr http.Handler, adminToken string) {
	req, _ := http.NewRequest("GET", "/api/v1/audit-logs/999999", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, response.Code) // 错误码不为0
	assert.Contains(t, response.Message, "not found")
}
