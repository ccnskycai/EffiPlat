package handler_test

import (
	"EffiPlat/backend/internal/handler"
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

// 模拟审计日志服务
type mockAuditLogService struct {
	createLogFunc func(ctx context.Context, log *model.AuditLog) error
	findLogsFunc func(ctx context.Context, params model.AuditLogQueryParams) ([]model.AuditLog, int64, error)
	findLogByIDFunc func(ctx context.Context, id uint) (*model.AuditLog, error)
	logUserActionFunc func(c *gin.Context, action, resource string, resourceID uint, details interface{}) error
	// 记录调用方法的参数
	calledActions []struct {
		action     string
		resource   string
		resourceID uint
		details    interface{}
	}
}

// CreateLog 实现AuditLogService接口
func (m *mockAuditLogService) CreateLog(ctx context.Context, log *model.AuditLog) error {
	if m.createLogFunc != nil {
		return m.createLogFunc(ctx, log)
	}
	return nil
}

// FindLogs 实现AuditLogService接口
func (m *mockAuditLogService) FindLogs(ctx context.Context, params model.AuditLogQueryParams) ([]model.AuditLog, int64, error) {
	if m.findLogsFunc != nil {
		return m.findLogsFunc(ctx, params)
	}
	return []model.AuditLog{}, 0, nil
}

// FindLogByID 实现AuditLogService接口
func (m *mockAuditLogService) FindLogByID(ctx context.Context, id uint) (*model.AuditLog, error) {
	if m.findLogByIDFunc != nil {
		return m.findLogByIDFunc(ctx, id)
	}
	return nil, nil
}

// LogUserAction 实现AuditLogService接口
func (m *mockAuditLogService) LogUserAction(c *gin.Context, action, resource string, resourceID uint, details interface{}) error {
	// 记录调用
	if m.calledActions == nil {
		m.calledActions = make([]struct {
			action     string
			resource   string
			resourceID uint
			details    interface{}
		}, 0)
	}
	m.calledActions = append(m.calledActions, struct {
		action     string
		resource   string
		resourceID uint
		details    interface{}
	}{
		action:     action,
		resource:   resource,
		resourceID: resourceID,
		details:    details,
	})
	
	if m.logUserActionFunc != nil {
		return m.logUserActionFunc(c, action, resource, resourceID, details)
	}
	return nil
}

// 创建模拟审计日志服务
func setupMockAuditLogService(t *testing.T) (*mockAuditLogService, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockService := &mockAuditLogService{}
	return mockService, ctrl
}

// 创建测试上下文
func setupTestContext(t *testing.T) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	
	// 模拟已登录用户
	ctx.Set("userID", uint(1))
	ctx.Set("username", "testuser")
	
	// 设置请求
	ctx.Request = httptest.NewRequest("POST", "/", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	
	return ctx, w
}

// 测试PermissionHandler的CreatePermission是否正确记录审计日志
func TestPermissionHandler_CreatePermission_AuditLog(t *testing.T) {
	// 创建模拟服务和上下文
	mockAuditService, ctrl := setupMockAuditLogService(t)
	defer ctrl.Finish()
	
	ctx, _ := setupTestContext(t)
	
	// 创建测试数据
	createReq := model.CreatePermissionRequest{
		Name:        "TestPermission",
		Description: "Test Description",
		Resource:    "TEST_RESOURCE",
		Action:      "READ",
	}
	
	reqBody, _ := json.Marshal(createReq)
	ctx.Request = httptest.NewRequest("POST", "/permissions", bytes.NewBuffer(reqBody))
	ctx.Request.Header.Set("Content-Type", "application/json")
	
	// 创建返回的模拟权限
	createdPermission := &model.Permission{
		ID:          1,
		Name:        createReq.Name,
		Description: createReq.Description,
		Resource:    createReq.Resource,
		Action:      createReq.Action,
	}
	
	// 创建模拟权限服务
	mockPermissionService := &mockPermissionService{
		createFunc: func(ctx context.Context, permission *model.Permission) (*model.Permission, error) {
			return createdPermission, nil
		},
	}
	
	// 创建handler
	logger := zaptest.NewLogger(t)
	permissionHandler := handler.NewPermissionHandler(mockPermissionService, mockAuditService, logger)
	
	// 执行请求
	permissionHandler.CreatePermission(ctx)
	
	// 验证审计日志调用
	assert.Len(t, mockAuditService.calledActions, 1, "应该有一次审计日志调用")
	if len(mockAuditService.calledActions) > 0 {
		action := mockAuditService.calledActions[0]
		assert.Equal(t, string(utils.AuditActionCreate), action.action, "应为创建操作")
		assert.Equal(t, "PERMISSION", action.resource, "资源类型应为权限")
		assert.Equal(t, createdPermission.ID, action.resourceID, "资源ID应匹配")
	}
}

// 测试ResponsibilityHandler的CreateResponsibility是否正确记录审计日志
func TestResponsibilityHandler_CreateResponsibility_AuditLog(t *testing.T) {
	// 创建模拟服务和上下文
	mockAuditService, ctrl := setupMockAuditLogService(t)
	defer ctrl.Finish()
	
	ctx, _ := setupTestContext(t)
	
	// 创建测试数据
	createReq := model.Responsibility{
		Name:        "TestResponsibility",
		Description: "Test Description",
	}
	
	reqBody, _ := json.Marshal(createReq)
	ctx.Request = httptest.NewRequest("POST", "/responsibilities", bytes.NewBuffer(reqBody))
	ctx.Request.Header.Set("Content-Type", "application/json")
	
	// 创建返回的模拟责任
	createdResponsibility := &model.Responsibility{
		ID:          1,
		Name:        createReq.Name,
		Description: createReq.Description,
	}
	
	// 创建模拟责任服务
	mockResponsibilityService := &mockResponsibilityService{
		createFunc: func(ctx context.Context, resp *model.Responsibility) (*model.Responsibility, error) {
			return createdResponsibility, nil
		},
	}
	
	// 创建handler
	logger := zaptest.NewLogger(t)
	responsibilityHandler := handler.NewResponsibilityHandler(mockResponsibilityService, mockAuditService, logger)
	
	// 执行请求
	responsibilityHandler.CreateResponsibility(ctx)
	
	// 验证审计日志调用
	assert.Len(t, mockAuditService.calledActions, 1, "应该有一次审计日志调用")
	if len(mockAuditService.calledActions) > 0 {
		action := mockAuditService.calledActions[0]
		assert.Equal(t, string(utils.AuditActionCreate), action.action, "应为创建操作")
		assert.Equal(t, "RESPONSIBILITY", action.resource, "资源类型应为责任")
		assert.Equal(t, createdResponsibility.ID, action.resourceID, "资源ID应匹配")
	}
}

// 测试ResponsibilityGroupHandler的CreateResponsibilityGroup是否正确记录审计日志
func TestResponsibilityGroupHandler_CreateResponsibilityGroup_AuditLog(t *testing.T) {
	// 创建模拟服务和上下文
	mockAuditService, ctrl := setupMockAuditLogService(t)
	defer ctrl.Finish()
	
	ctx, _ := setupTestContext(t)
	
	// 创建测试数据
	createReq := handler.CreateResponsibilityGroupRequest{
		Name:        "TestGroup",
		Description: "Test Group Description",
		ResponsibilityIDs: []uint{1, 2, 3},
	}
	
	reqBody, _ := json.Marshal(createReq)
	ctx.Request = httptest.NewRequest("POST", "/responsibility-groups", bytes.NewBuffer(reqBody))
	ctx.Request.Header.Set("Content-Type", "application/json")
	
	// 创建返回的模拟责任组
	createdGroup := &model.ResponsibilityGroup{
		ID:          1,
		Name:        createReq.Name,
		Description: createReq.Description,
	}
	
	// 创建模拟责任组服务
	mockResponsibilityGroupService := &mockResponsibilityGroupService{
		createFunc: func(ctx context.Context, group *model.ResponsibilityGroup, responsibilityIDs []uint) (*model.ResponsibilityGroup, error) {
			return createdGroup, nil
		},
	}
	
	// 创建handler
	logger := zaptest.NewLogger(t)
	groupHandler := handler.NewResponsibilityGroupHandler(mockResponsibilityGroupService, mockAuditService, logger)
	
	// 执行请求
	groupHandler.CreateResponsibilityGroup(ctx)
	
	// 验证审计日志调用
	assert.Len(t, mockAuditService.calledActions, 1, "应该有一次审计日志调用")
	if len(mockAuditService.calledActions) > 0 {
		action := mockAuditService.calledActions[0]
		assert.Equal(t, string(utils.AuditActionCreate), action.action, "应为创建操作")
		assert.Equal(t, "RESPONSIBILITY_GROUP", action.resource, "资源类型应为责任组")
		assert.Equal(t, createdGroup.ID, action.resourceID, "资源ID应匹配")
	}
}

// 模拟的权限服务
type mockPermissionService struct {
	createFunc func(ctx context.Context, permission *model.Permission) (*model.Permission, error)
}

func (m *mockPermissionService) CreatePermission(ctx context.Context, permission *model.Permission) (*model.Permission, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, permission)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockPermissionService) GetPermissions(ctx context.Context, params model.PermissionListParams) ([]model.Permission, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (m *mockPermissionService) GetPermissionByID(ctx context.Context, id uint) (*model.Permission, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockPermissionService) UpdatePermission(ctx context.Context, id uint, permission *model.Permission) (*model.Permission, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockPermissionService) DeletePermission(ctx context.Context, id uint) error {
	return fmt.Errorf("not implemented")
}

func (m *mockPermissionService) AddPermissionsToRole(ctx context.Context, roleID uint, permissionIDs []uint) error {
	return fmt.Errorf("not implemented")
}

func (m *mockPermissionService) RemovePermissionsFromRole(ctx context.Context, roleID uint, permissionIDs []uint) error {
	return fmt.Errorf("not implemented")
}

func (m *mockPermissionService) GetPermissionsByRoleID(ctx context.Context, roleID uint) ([]model.Permission, error) {
	return nil, fmt.Errorf("not implemented")
}

// 模拟的责任服务
type mockResponsibilityService struct {
	createFunc func(ctx context.Context, resp *model.Responsibility) (*model.Responsibility, error)
	getByIDFunc func(ctx context.Context, id uint) (*model.Responsibility, error)
	updateFunc func(ctx context.Context, id uint, resp *model.Responsibility) (*model.Responsibility, error)
	deleteFunc func(ctx context.Context, id uint) error
}

func (m *mockResponsibilityService) CreateResponsibility(ctx context.Context, resp *model.Responsibility) (*model.Responsibility, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, resp)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockResponsibilityService) GetResponsibilities(ctx context.Context, params model.ResponsibilityListParams) ([]model.Responsibility, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (m *mockResponsibilityService) GetResponsibilityByID(ctx context.Context, id uint) (*model.Responsibility, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockResponsibilityService) UpdateResponsibility(ctx context.Context, id uint, resp *model.Responsibility) (*model.Responsibility, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, resp)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockResponsibilityService) DeleteResponsibility(ctx context.Context, id uint) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return fmt.Errorf("not implemented")
}
// 模拟的责任组服务
type mockResponsibilityGroupService struct {
	createFunc func(ctx context.Context, group *model.ResponsibilityGroup, respIDs []uint) (*model.ResponsibilityGroup, error)
	getByIDFunc func(ctx context.Context, id uint) (*model.ResponsibilityGroup, error)
	getAllFunc func(ctx context.Context, params model.ResponsibilityGroupListParams) ([]model.ResponsibilityGroup, int64, error)
	updateFunc func(ctx context.Context, id uint, group *model.ResponsibilityGroup, respIDs *[]uint) (*model.ResponsibilityGroup, error)
	deleteFunc func(ctx context.Context, id uint) error
	addRespFunc func(ctx context.Context, groupID, respID uint) error
	removeRespFunc func(ctx context.Context, groupID, respID uint) error
}

func (m *mockResponsibilityGroupService) CreateResponsibilityGroup(ctx context.Context, group *model.ResponsibilityGroup, respIDs []uint) (*model.ResponsibilityGroup, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, group, respIDs)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockResponsibilityGroupService) GetResponsibilityGroups(ctx context.Context, params model.ResponsibilityGroupListParams) ([]model.ResponsibilityGroup, int64, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc(ctx, params)
	}
	return nil, 0, fmt.Errorf("not implemented")
}

func (m *mockResponsibilityGroupService) GetResponsibilityGroupByID(ctx context.Context, id uint) (*model.ResponsibilityGroup, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockResponsibilityGroupService) UpdateResponsibilityGroup(ctx context.Context, id uint, group *model.ResponsibilityGroup, respIDs *[]uint) (*model.ResponsibilityGroup, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, group, respIDs)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockResponsibilityGroupService) DeleteResponsibilityGroup(ctx context.Context, id uint) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return fmt.Errorf("not implemented")
}

func (m *mockResponsibilityGroupService) AddResponsibilityToGroup(ctx context.Context, groupID, respID uint) error {
	if m.addRespFunc != nil {
		return m.addRespFunc(ctx, groupID, respID)
	}
	return fmt.Errorf("not implemented")
}

func (m *mockResponsibilityGroupService) RemoveResponsibilityFromGroup(ctx context.Context, groupID, respID uint) error {
	if m.removeRespFunc != nil {
		return m.removeRespFunc(ctx, groupID, respID)
	}
	return fmt.Errorf("not implemented")
}
