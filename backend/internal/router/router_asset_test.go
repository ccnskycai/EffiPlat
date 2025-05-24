package router_test

import (
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/router"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAssetManagementRoutes covers tests for asset CRUD operations.
func TestAssetManagementRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var createdAssetID uint
	var testEnvironmentID uint // To store the ID of the environment created for tests

	// --- Setup: Create a test environment first ---
	t.Run("Setup_CreateTestEnvironment", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		envName := "Test Env for Assets " + uuid.NewString()[:8]
		envSlug := "test-assets-" + uuid.NewString()[:8]
		createEnvReq := model.CreateEnvironmentRequest{
			Name:        envName,
			Slug:        envSlug,
			Description: "Temporary environment for asset tests",
		}
		payloadBytes, _ := json.Marshal(createEnvReq)
		req, _ := http.NewRequest("POST", "/api/v1/environments", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code, "Failed to create test environment")
		var envRespBody struct {
			Data model.Environment `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &envRespBody)
		require.NoError(t, err)
		require.NotZero(t, envRespBody.Data.ID, "Test environment ID should not be zero")
		testEnvironmentID = envRespBody.Data.ID
		t.Logf("Test environment created with ID: %d", testEnvironmentID)
	})

	require.NotZero(t, testEnvironmentID, "Test environment ID must be set for asset tests")

	// --- Test Case: Create Asset - Success ---
	t.Run("CreateAsset_Success", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		assetHostname := "test-asset-" + uuid.NewString()[:8]
		assetIP := fmt.Sprintf("192.168.1.%d", uuid.New().ID()%250+1) // Generate a pseudo-random IP

		createAssetReq := model.CreateAssetRequest{
			Hostname:      assetHostname,
			IPAddress:     assetIP,
			AssetType:     model.AssetTypeVM,
			Status:        model.AssetStatusOnline,
			Description:   "A test virtual machine",
			EnvironmentID: testEnvironmentID,
		}
		payloadBytes, _ := json.Marshal(createAssetReq)
		req, _ := http.NewRequest("POST", "/api/v1/assets", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Expected status 201 Created for asset creation")
		var respBody struct {
			Data model.Asset `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err, "Failed to unmarshal response body")

		assert.NotZero(t, respBody.Data.ID, "Created asset ID should not be zero")
		createdAssetID = respBody.Data.ID // Save for later tests
		assert.Equal(t, assetHostname, respBody.Data.Hostname)
		assert.Equal(t, assetIP, respBody.Data.IPAddress)
		assert.Equal(t, model.AssetTypeVM, respBody.Data.AssetType)
		assert.Equal(t, model.AssetStatusOnline, respBody.Data.Status) // Ensure status from request is used
		assert.Equal(t, testEnvironmentID, respBody.Data.EnvironmentID)
		t.Logf("Asset created successfully with ID: %d, Hostname: %s", createdAssetID, assetHostname)
	})

	// --- Test Case: Create Asset - Fail - Invalid Payload (Missing Hostname) ---
	t.Run("CreateAsset_Fail_InvalidPayload_MissingHostname", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		assetIP := fmt.Sprintf("192.168.1.%d", uuid.New().ID()%250+1)

		createAssetReq := model.CreateAssetRequest{
			// Hostname: "missing-hostname-asset", // Hostname is missing
			IPAddress:     assetIP,
			AssetType:     model.AssetTypePhysicalServer,
			EnvironmentID: testEnvironmentID,
		}
		payloadBytes, _ := json.Marshal(createAssetReq)
		req, _ := http.NewRequest("POST", "/api/v1/assets", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status 400 Bad Request for missing hostname")
		// Optionally, check error message structure if your response.Error function provides a consistent format
		// For example:
		// var errResp response.ErrorResponse
		// err := json.Unmarshal(w.Body.Bytes(), &errResp)
		// assert.NoError(t, err)
		// assert.Contains(t, strings.ToLower(errResp.Message), "hostname") // Check if error message mentions hostname
	})

	// --- Test Case: Create Asset - Fail - Non-existent Environment ID ---
	t.Run("CreateAsset_Fail_NonExistentEnvironment", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		nonExistentEnvID := uint(999999)           // An ID that is very unlikely to exist
		if testEnvironmentID == nonExistentEnvID { // Just in case, highly improbable
			nonExistentEnvID++
		}

		createAssetReq := model.CreateAssetRequest{
			Hostname:      "asset-for-nonexistent-env-" + uuid.NewString()[:6],
			IPAddress:     fmt.Sprintf("192.168.1.%d", uuid.New().ID()%250+1),
			AssetType:     model.AssetTypeCloudHost,
			EnvironmentID: nonExistentEnvID,
		}
		payloadBytes, _ := json.Marshal(createAssetReq)
		req, _ := http.NewRequest("POST", "/api/v1/assets", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// The service layer should return an error that the handler translates to 404 (or potentially 400/422 depending on design)
		// Assuming service.CreateAsset checks envRepo.GetByID and that returns gorm.ErrRecordNotFound,
		// and the handler translates this to a 404.
		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status 404 Not Found for non-existent environment ID")
	})

	// TODO: Add more test cases:
	// - CreateAsset_Fail_DuplicateHostname (if business logic prevents it before DB)
	// - CreateAsset_Fail_DuplicateIPAddress (if business logic prevents it before DB)
	// - ListAssets_Success_Empty
	// - ListAssets_Success_WithDataAndPagination (after creating a few assets)
	// - ListAssets_Success_WithFilters
	// - GetAssetByID_Success (using createdAssetID)
	// - GetAssetByID_Fail_NotFound
	// - UpdateAsset_Success (using createdAssetID)
	// - UpdateAsset_Fail_NotFound
	// - UpdateAsset_Fail_InvalidPayload
	// - UpdateAsset_Fail_ChangeToExistingHostnameOrIP (if applicable)
	// - DeleteAsset_Success (using createdAssetID)
	// - DeleteAsset_Fail_NotFound

	// --- Teardown: Delete the test environment ---
	// This should ideally run even if asset tests fail. Consider using t.Cleanup or a separate teardown test.
	t.Run("Teardown_DeleteTestEnvironment", func(t *testing.T) {
		if testEnvironmentID == 0 {
			t.Log("Skipping test environment deletion as ID is zero.")
			return
		}
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		deleteEnvURL := fmt.Sprintf("/api/v1/environments/%d", testEnvironmentID)
		req, _ := http.NewRequest("DELETE", deleteEnvURL, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// Environment deletion should return 204 No Content or 200 OK with a message
		// Adjust based on your EnvironmentHandler's DeleteEnvironment implementation
		if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
			t.Errorf("Failed to delete test environment. Status: %d, Body: %s", w.Code, w.Body.String())
		} else {
			t.Logf("Test environment %d deleted successfully or was already gone.", testEnvironmentID)
		}
	})
}
