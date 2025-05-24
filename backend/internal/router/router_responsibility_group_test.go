package router_test

import (
	"EffiPlat/backend/internal/handler" // For request structs if needed
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/router" // Added import
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to create a responsibility group for testing
func createTestResponsibilityGroup(t *testing.T, router http.Handler, token string, groupReq handler.CreateResponsibilityGroupRequest) model.ResponsibilityGroup {
	jsonData, err := json.Marshal(groupReq)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/responsibility-groups", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println("[DEBUG] Create group response:", w.Body.String())
	assert.Equal(t, http.StatusCreated, w.Code, "Failed to create group, body: "+w.Body.String())

	var createdGroupResp struct {
		Data model.ResponsibilityGroup `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &createdGroupResp)
	assert.NoError(t, err)
	fmt.Println("[DEBUG] Created group ID:", createdGroupResp.Data.ID)
	assert.NotZero(t, createdGroupResp.Data.ID)
	return createdGroupResp.Data
}

func TestResponsibilityGroupRoutes(t *testing.T) {
	app := router.SetupTestApp(t)
	token := router.GetAuthTokenForTest(t, app.Router, app.DB)

	// Pre-create some responsibilities to associate with groups
	resp1 := router.CreateTestResponsibility(t, app.Router, token, &model.Responsibility{Name: "Group Test Resp 1", Description: "Desc for group test 1"})
	resp2 := router.CreateTestResponsibility(t, app.Router, token, &model.Responsibility{Name: "Group Test Resp 2", Description: "Desc for group test 2"})
	resp3 := router.CreateTestResponsibility(t, app.Router, token, &model.Responsibility{Name: "Group Test Resp 3", Description: "Desc for group test 3"})

	// 1. 创建 group
	newGroupReq := handler.CreateResponsibilityGroupRequest{
		Name:              "Test Group Alpha",
		Description:       "Alpha group for testing",
		ResponsibilityIDs: []uint{resp1.ID, resp2.ID},
	}
	createdTestGroup := createTestResponsibilityGroup(t, app.Router, token, newGroupReq)
	fmt.Println("[TEST] createdTestGroup.ID:", createdTestGroup.ID)
	assert.Equal(t, newGroupReq.Name, createdTestGroup.Name)
	assert.Equal(t, newGroupReq.Description, createdTestGroup.Description)

	// 2. 列表 group
	createTestResponsibilityGroup(t, app.Router, token, handler.CreateResponsibilityGroupRequest{Name: "Test Group Beta", Description: "Beta"})
	listReq, _ := http.NewRequest(http.MethodGet, "/api/v1/responsibility-groups?pageSize=5", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listW := httptest.NewRecorder()
	app.Router.ServeHTTP(listW, listReq)
	assert.Equal(t, http.StatusOK, listW.Code)
	var listResp struct {
		Code int `json:"code"`
		Message string `json:"message"`
		Data struct {
			Items []model.ResponsibilityGroup `json:"items"`
			Total int64 `json:"total"`
			Page int `json:"page"`
			PageSize int `json:"pageSize"`
		} `json:"data"`
	}
	err := json.Unmarshal(listW.Body.Bytes(), &listResp)
	assert.NoError(t, err)
	assert.True(t, listResp.Data.Total >= 2)
	assert.NotEmpty(t, listResp.Data.Items)

	// 3. GET by ID
	url := fmt.Sprintf("/api/v1/responsibility-groups/%d", createdTestGroup.ID)
	fmt.Println("[TEST] GET url:", url)
	getReq, _ := http.NewRequest(http.MethodGet, url, nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)
	assert.Equal(t, http.StatusOK, getW.Code)
	var getResp struct {
		Data model.ResponsibilityGroup `json:"data"`
	}
	err = json.Unmarshal(getW.Body.Bytes(), &getResp)
	assert.NoError(t, err)
	assert.Equal(t, createdTestGroup.ID, getResp.Data.ID)
	assert.Equal(t, createdTestGroup.Name, getResp.Data.Name)
	assert.Len(t, getResp.Data.Responsibilities, 2)
	foundResp1 := false
	foundResp2 := false
	for _, r := range getResp.Data.Responsibilities {
		if r.ID == resp1.ID {
			foundResp1 = true
		}
		if r.ID == resp2.ID {
			foundResp2 = true
		}
	}
	assert.True(t, foundResp1, "resp1 should be associated")
	assert.True(t, foundResp2, "resp2 should be associated")

	// 4. PUT 更新 group
	newRespIDs := []uint{resp3.ID}
	updatePayload := handler.UpdateResponsibilityGroupRequest{
		Name:              "Updated Test Group Alpha",
		Description:       "New description for Alpha.",
		ResponsibilityIDs: &newRespIDs,
	}
	jsonData, _ := json.Marshal(updatePayload)
	putURL := fmt.Sprintf("/api/v1/responsibility-groups/%d", createdTestGroup.ID)
	putReq, _ := http.NewRequest(http.MethodPut, putURL, bytes.NewBuffer(jsonData))
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("Authorization", "Bearer "+token)
	putW := httptest.NewRecorder()
	app.Router.ServeHTTP(putW, putReq)
	assert.Equal(t, http.StatusOK, putW.Code, "Update failed: "+putW.Body.String())
	var updatedResp struct {
		Data model.ResponsibilityGroup `json:"data"`
	}
	err = json.Unmarshal(putW.Body.Bytes(), &updatedResp)
	assert.NoError(t, err)
	assert.Equal(t, createdTestGroup.ID, updatedResp.Data.ID)
	assert.Equal(t, updatePayload.Name, updatedResp.Data.Name)
	assert.Len(t, updatedResp.Data.Responsibilities, 1)
	assert.Equal(t, resp3.ID, updatedResp.Data.Responsibilities[0].ID)

	// 5. POST 添加责任到 group
	addURL := fmt.Sprintf("/api/v1/responsibility-groups/%d/responsibilities/%d", createdTestGroup.ID, resp1.ID)
	addReq, _ := http.NewRequest(http.MethodPost, addURL, nil)
	addReq.Header.Set("Authorization", "Bearer "+token)
	addW := httptest.NewRecorder()
	app.Router.ServeHTTP(addW, addReq)
	assert.Equal(t, http.StatusNoContent, addW.Code, "Add resp to group failed: "+addW.Body.String())
	// 验证添加后 group
	verifyReq, _ := http.NewRequest(http.MethodGet, url, nil)
	verifyReq.Header.Set("Authorization", "Bearer "+token)
	verifyW := httptest.NewRecorder()
	app.Router.ServeHTTP(verifyW, verifyReq)
	var groupVerify struct {
		Data model.ResponsibilityGroup `json:"data"`
	}
	json.Unmarshal(verifyW.Body.Bytes(), &groupVerify)
	assert.Len(t, groupVerify.Data.Responsibilities, 2)

	// 6. DELETE 移除责任
	delURL := fmt.Sprintf("/api/v1/responsibility-groups/%d/responsibilities/%d", createdTestGroup.ID, resp3.ID)
	delReq, _ := http.NewRequest(http.MethodDelete, delURL, nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delW := httptest.NewRecorder()
	app.Router.ServeHTTP(delW, delReq)
	assert.Equal(t, http.StatusNoContent, delW.Code, "Remove resp from group failed: "+delW.Body.String())
	// 验证删除后 group
	verifyReq2, _ := http.NewRequest(http.MethodGet, url, nil)
	verifyReq2.Header.Set("Authorization", "Bearer "+token)
	verifyW2 := httptest.NewRecorder()
	app.Router.ServeHTTP(verifyW2, verifyReq2)
	var groupVerify2 struct {
		Data model.ResponsibilityGroup `json:"data"`
	}
	json.Unmarshal(verifyW2.Body.Bytes(), &groupVerify2)
	assert.Len(t, groupVerify2.Data.Responsibilities, 1)
	assert.Equal(t, resp1.ID, groupVerify2.Data.Responsibilities[0].ID)

	// 7. DELETE group
	deleteGroupURL := fmt.Sprintf("/api/v1/responsibility-groups/%d", createdTestGroup.ID)
	deleteReq, _ := http.NewRequest(http.MethodDelete, deleteGroupURL, nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)
	deleteW := httptest.NewRecorder()
	app.Router.ServeHTTP(deleteW, deleteReq)
	assert.Equal(t, http.StatusNoContent, deleteW.Code)
	// 验证 group 已删除
	getDeletedReq, _ := http.NewRequest(http.MethodGet, deleteGroupURL, nil)
	getDeletedReq.Header.Set("Authorization", "Bearer "+token)
	getDeletedW := httptest.NewRecorder()
	app.Router.ServeHTTP(getDeletedW, getDeletedReq)
	assert.Equal(t, http.StatusNotFound, getDeletedW.Code)
}
