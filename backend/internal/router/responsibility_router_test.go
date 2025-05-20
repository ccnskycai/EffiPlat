package router_test

import (
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/router"
	"EffiPlat/backend/pkg/response"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponsibilityRoutes(t *testing.T) {
	app := router.SetupTestApp(t)
	sqlDB, _ := app.DB.DB()
	defer sqlDB.Close()

	// Get a token using the helper from router_test_helper.go
	authToken := router.GetAuthTokenForTest(t, app.Router, app.DB)

	var createdTestResp models.Responsibility

	t.Run("POST /responsibilities - Create Responsibility", func(t *testing.T) {
		newResp := models.Responsibility{
			Name:        "Test Responsibility",
			Description: "This is a test responsibility.",
		}
		createdTestResp = router.CreateTestResponsibility(t, app.Router, authToken, &newResp)
		assert.Equal(t, newResp.Name, createdTestResp.Name)
		assert.Equal(t, newResp.Description, createdTestResp.Description)
	})

	t.Run("GET /responsibilities - List Responsibilities", func(t *testing.T) {
		// Create a second responsibility to test listing
		router.CreateTestResponsibility(t, app.Router, authToken, &models.Responsibility{Name: "Another Test Resp", Description: "Desc 2"})

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/responsibilities?pageSize=5", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var listResp struct {
			Data response.PaginatedData `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &listResp)
		assert.NoError(t, err)
		assert.True(t, listResp.Data.Total >= 2) // At least the two we created
		assert.NotEmpty(t, listResp.Data.Items)

		// Check if our createdTestResp is in the list
		foundInList := false
		if items, ok := listResp.Data.Items.([]interface{}); ok {
			for _, itemMap := range items {
				if item, ok := itemMap.(map[string]interface{}); ok {
					if idFloat, ok := item["id"].(float64); ok && uint(idFloat) == createdTestResp.ID {
						foundInList = true
						break
					}
				}
			}
		}
		assert.True(t, foundInList, "Created responsibility should be in the list")
	})

	t.Run("GET /responsibilities/:id - Get Responsibility By ID", func(t *testing.T) {
		url := fmt.Sprintf("/api/v1/responsibilities/%d", createdTestResp.ID)
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var getResp struct {
			Data models.Responsibility `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &getResp)
		assert.NoError(t, err)
		assert.Equal(t, createdTestResp.ID, getResp.Data.ID)
		assert.Equal(t, createdTestResp.Name, getResp.Data.Name)
	})

	t.Run("GET /responsibilities/:id - Not Found", func(t *testing.T) {
		url := "/api/v1/responsibilities/99999" // Non-existent ID
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("PUT /responsibilities/:id - Update Responsibility", func(t *testing.T) {
		updatePayload := models.Responsibility{
			Name:        "Updated Test Responsibility",
			Description: "This description has been updated.",
		}
		jsonData, _ := json.Marshal(updatePayload)
		url := fmt.Sprintf("/api/v1/responsibilities/%d", createdTestResp.ID)
		req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken)

		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var updatedResp struct {
			Data models.Responsibility `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &updatedResp)
		assert.NoError(t, err)
		assert.Equal(t, createdTestResp.ID, updatedResp.Data.ID)
		assert.Equal(t, updatePayload.Name, updatedResp.Data.Name)
		assert.Equal(t, updatePayload.Description, updatedResp.Data.Description)
	})

	t.Run("DELETE /responsibilities/:id - Delete Responsibility", func(t *testing.T) {
		url := fmt.Sprintf("/api/v1/responsibilities/%d", createdTestResp.ID)
		req, _ := http.NewRequest(http.MethodDelete, url, nil)
		req.Header.Set("Authorization", "Bearer "+authToken)

		w := httptest.NewRecorder()
		app.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify it's deleted
		reqGet, _ := http.NewRequest(http.MethodGet, url, nil)
		reqGet.Header.Set("Authorization", "Bearer "+authToken)
		wGet := httptest.NewRecorder()
		app.Router.ServeHTTP(wGet, reqGet)
		assert.Equal(t, http.StatusNotFound, wGet.Code)
	})
}
