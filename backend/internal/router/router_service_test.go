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

// TestServiceAndTypeManagementRoutes covers tests for service and service type CRUD operations.
func TestServiceAndTypeManagementRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var (
		createdServiceTypeID uint
		createdServiceID     uint // Used specifically for DeleteServiceType_Fail_InUse and cleaned up locally
		testServiceTypeName  string
		testServiceID        uint // Will store the ID of the service created for general service tests
	)

	// --- Setup: Create a test service type first ---
	t.Run("Setup_CreateTestServiceType", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		stName := "Test ST " + uuid.NewString()[:8]
		createSTReq := model.CreateServiceTypeRequest{
			Name:        stName,
			Description: "Temporary service type for API tests",
		}
		payloadBytes, _ := json.Marshal(createSTReq)
		req, _ := http.NewRequest("POST", "/api/v1/service-types", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code, "Failed to create test service type")
		var stRespBody struct {
			Data model.ServiceType `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &stRespBody)
		require.NoError(t, err)
		require.NotZero(t, stRespBody.Data.ID, "Test service type ID should not be zero")
		createdServiceTypeID = stRespBody.Data.ID
		testServiceTypeName = stName // Assign to package-level variable
		t.Logf("Test service type created with ID: %d, Name: %s", createdServiceTypeID, stName)
	})

	require.NotZero(t, createdServiceTypeID, "Test service type ID must be set for subsequent tests")

	// --- ServiceType API Test Cases ---
	// The package-level 'testServiceTypeName' (used in Setup_CreateTestServiceType)
	// is already available and holds the name of the initially created service type.
	// No need to redeclare or re-fetch it here.

	t.Run("CreateServiceType_Fail_DuplicateName", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		createSTReq := model.CreateServiceTypeRequest{
			Name:        testServiceTypeName, // Use the same name as the one created in setup
			Description: "Attempt to create duplicate service type",
		}
		payloadBytes, _ := json.Marshal(createSTReq)
		req, _ := http.NewRequest("POST", "/api/v1/service-types", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code, "Expected status 409 Conflict for duplicate service type name")
	})

	t.Run("CreateServiceType_Fail_InvalidPayload_MissingName", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		createSTReq := model.CreateServiceTypeRequest{
			Description: "Service type with missing name",
		}
		payloadBytes, _ := json.Marshal(createSTReq)
		req, _ := http.NewRequest("POST", "/api/v1/service-types", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status 400 Bad Request for missing name")
	})

	t.Run("GetServiceTypeByID_Success", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		getURL := fmt.Sprintf("/api/v1/service-types/%d", createdServiceTypeID)
		req, _ := http.NewRequest("GET", getURL, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK for getting service type by ID")
		var respBody struct {
			Data model.ServiceType `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err, "Failed to unmarshal response body")
		assert.Equal(t, createdServiceTypeID, respBody.Data.ID)
		assert.Equal(t, testServiceTypeName, respBody.Data.Name) // Verify name matches
	})

	t.Run("GetServiceTypeByID_Fail_NotFound", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		nonExistentID := uint(999999)
		getURL := fmt.Sprintf("/api/v1/service-types/%d", nonExistentID)
		req, _ := http.NewRequest("GET", getURL, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status 404 Not Found for non-existent service type ID")
	})

	t.Run("UpdateServiceType_Success", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		updatedName := "Updated ST Name " + uuid.NewString()[:8]
		updatedDesc := "Updated description for test ST " + uuid.NewString()[:8]
		updateReq := model.UpdateServiceTypeRequest{
			Name:        &updatedName,
			Description: &updatedDesc,
		}
		payloadBytes, _ := json.Marshal(updateReq)
		updateURL := fmt.Sprintf("/api/v1/service-types/%d", createdServiceTypeID)
		req, _ := http.NewRequest("PUT", updateURL, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK for updating service type")
		var respBody struct {
			Data model.ServiceType `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, createdServiceTypeID, respBody.Data.ID)
		assert.Equal(t, updatedDesc, respBody.Data.Description)
		assert.Equal(t, updatedName, respBody.Data.Name) // Ensure name is updated as expected
	})

	t.Run("UpdateServiceType_Fail_NotFound", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		nonExistentID := uint(999998)
		updatedDesc := "Attempt to update non-existent ST"
		updateReq := model.UpdateServiceTypeRequest{
			Name:        model.StringPtr("NonExistentName"),
			Description: &updatedDesc,
		}
		payloadBytes, _ := json.Marshal(updateReq)
		updateURL := fmt.Sprintf("/api/v1/service-types/%d", nonExistentID)
		req, _ := http.NewRequest("PUT", updateURL, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status 404 Not Found for updating non-existent service type")
	})

	t.Run("UpdateServiceType_Fail_DuplicateName", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		// 1. Create an auxiliary service type
		auxName := "Aux ST For Update Conflict " + uuid.NewString()[:8]
		createAuxSTReq := model.CreateServiceTypeRequest{
			Name:        auxName,
			Description: "Auxiliary ST",
		}
		auxPayloadBytes, _ := json.Marshal(createAuxSTReq)
		auxReq, _ := http.NewRequest("POST", "/api/v1/service-types", bytes.NewBuffer(auxPayloadBytes))
		auxReq.Header.Set("Content-Type", "application/json")
		auxReq.Header.Set("Authorization", "Bearer "+adminToken)
		auxW := httptest.NewRecorder()
		rtr.ServeHTTP(auxW, auxReq)
		require.Equal(t, http.StatusCreated, auxW.Code, "Failed to create auxiliary service type for conflict test")
		var auxRespBody struct {
			Data model.ServiceType `json:"data"`
		}
		err := json.Unmarshal(auxW.Body.Bytes(), &auxRespBody)
		require.NoError(t, err)
		auxServiceTypeID := auxRespBody.Data.ID
		require.NotZero(t, auxServiceTypeID, "Auxiliary service type ID should not be zero")

		// Teardown for auxiliary service type
		t.Cleanup(func() {
			deleteAuxURL := fmt.Sprintf("/api/v1/service-types/%d", auxServiceTypeID)
			cleanupReq, _ := http.NewRequest("DELETE", deleteAuxURL, nil)
			cleanupReq.Header.Set("Authorization", "Bearer "+adminToken)
			cleanupW := httptest.NewRecorder()
			rtr.ServeHTTP(cleanupW, cleanupReq)
			if cleanupW.Code != http.StatusNoContent && cleanupW.Code != http.StatusOK {
				t.Logf("Failed to cleanup auxiliary service type %d. Status: %d", auxServiceTypeID, cleanupW.Code)
			}
		})

		// 2. Attempt to update the main test service type (createdServiceTypeID) to use auxName
		updateReq := model.UpdateServiceTypeRequest{
			Name: model.StringPtr(auxName), // Attempt to use the name of the auxiliary ST
		}
		payloadBytes, _ := json.Marshal(updateReq)
		updateURL := fmt.Sprintf("/api/v1/service-types/%d", createdServiceTypeID)
		req, _ := http.NewRequest("PUT", updateURL, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code, "Expected status 409 Conflict when updating service type to an existing name")
	})

	t.Run("ListServiceTypes_Success_WithDataAndPagination", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		// Ensure there's at least one service type (createdServiceTypeID exists)
		// Create a couple more for pagination testing
		stNamesToCreate := []string{
			"Paging ST A " + uuid.NewString()[:7],
			"Paging ST B " + uuid.NewString()[:7],
		}
		createdAuxIDs := []uint{}

		for _, name := range stNamesToCreate {
			createReq := model.CreateServiceTypeRequest{Name: name, Description: "Paging test ST"}
			payload, _ := json.Marshal(createReq)
			req, _ := http.NewRequest("POST", "/api/v1/service-types", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+adminToken)
			w := httptest.NewRecorder()
			rtr.ServeHTTP(w, req)
			require.Equal(t, http.StatusCreated, w.Code, "Failed to create aux service type for list test: "+name)
			var resp struct{ Data model.ServiceType `json:"data"` }
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			createdAuxIDs = append(createdAuxIDs, resp.Data.ID)
		}

		t.Cleanup(func() {
			for _, id := range createdAuxIDs {
				deleteURL := fmt.Sprintf("/api/v1/service-types/%d", id)
				cleanupReq, _ := http.NewRequest("DELETE", deleteURL, nil)
				cleanupReq.Header.Set("Authorization", "Bearer "+adminToken)
				cleanupW := httptest.NewRecorder()
				rtr.ServeHTTP(cleanupW, cleanupReq)
				if cleanupW.Code != http.StatusNoContent && cleanupW.Code != http.StatusOK {
					t.Logf("Failed to cleanup aux service type %d for list test. Status: %d", id, cleanupW.Code)
				}
			}
		})

		// Assuming at least 3 service types now (original + 2 aux)
		// Test fetching page 1 with page size 2
		listURL := "/api/v1/service-types?page=1&pageSize=2&orderBy=name&sortOrder=asc"
		reqList, _ := http.NewRequest("GET", listURL, nil)
		reqList.Header.Set("Authorization", "Bearer "+adminToken)
		wList := httptest.NewRecorder()
		rtr.ServeHTTP(wList, reqList)

		assert.Equal(t, http.StatusOK, wList.Code, "Expected status 200 OK for listing service types")

		var listResp struct {
			Data struct {
				Items    []model.ServiceType `json:"items"`
				Total    int64                `json:"total"`
				Page     int                  `json:"page"`
				PageSize int                  `json:"pageSize"`
			} `json:"data"`
		}
		err := json.Unmarshal(wList.Body.Bytes(), &listResp)
		assert.NoError(t, err, "Failed to unmarshal list response")

		assert.Len(t, listResp.Data.Items, 2, "Expected 2 service types on page 1 with pageSize 2")
		assert.GreaterOrEqual(t, listResp.Data.Total, int64(3), "Expected total count to be at least 3")
		assert.Equal(t, 1, listResp.Data.Page, "Expected current page to be 1")         // Corrected to int
		assert.Equal(t, 2, listResp.Data.PageSize, "Expected page size to be 2")      // Corrected to int
		// Check if items are sorted by name ascending (first item's name should be alphabetically <= second item's name)
		if len(listResp.Data.Items) == 2 {
			assert.LessOrEqual(t, listResp.Data.Items[0].Name, listResp.Data.Items[1].Name, "Service types should be sorted by name ascending")
		}
	})

	t.Run("DeleteServiceType_Success", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		// 1. Create a temporary service type for deletion test
		tempSTName := "Temp ST For Delete " + uuid.NewString()[:8]
		createTempSTReq := model.CreateServiceTypeRequest{Name: tempSTName}
		payload, _ := json.Marshal(createTempSTReq)
		reqCreate, _ := http.NewRequest("POST", "/api/v1/service-types", bytes.NewBuffer(payload))
		reqCreate.Header.Set("Content-Type", "application/json")
		reqCreate.Header.Set("Authorization", "Bearer "+adminToken)
		wCreate := httptest.NewRecorder()
		rtr.ServeHTTP(wCreate, reqCreate)
		require.Equal(t, http.StatusCreated, wCreate.Code)
		var respCreate struct{ Data model.ServiceType `json:"data"` }
		err := json.Unmarshal(wCreate.Body.Bytes(), &respCreate)
		require.NoError(t, err)
		tempSTID := respCreate.Data.ID
		require.NotZero(t, tempSTID)

		// 2. Delete the temporary service type
		deleteURL := fmt.Sprintf("/api/v1/service-types/%d", tempSTID)
		reqDelete, _ := http.NewRequest("DELETE", deleteURL, nil)
		reqDelete.Header.Set("Authorization", "Bearer "+adminToken)
		wDelete := httptest.NewRecorder()
		rtr.ServeHTTP(wDelete, reqDelete)

		assert.Equal(t, http.StatusNoContent, wDelete.Code, "Expected status 204 No Content for successful service type deletion")
	})

	t.Run("DeleteServiceType_Fail_NotFound", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		nonExistentID := uint(999997)
		deleteURL := fmt.Sprintf("/api/v1/service-types/%d", nonExistentID)
		req, _ := http.NewRequest("DELETE", deleteURL, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status 404 Not Found for deleting non-existent service type")
	})

	t.Run("DeleteServiceType_Fail_InUse", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		// 1. Create a Service that uses createdServiceTypeID
		serviceName := "Test Svc For ST Delete Conflict " + uuid.NewString()[:8]
		createServiceReq := model.CreateServiceRequest{
			Name:          serviceName,
			ServiceTypeID: createdServiceTypeID, // Link to the main test service type
			Status:        model.ServiceStatusActive,
		}
		sPayloadBytes, _ := json.Marshal(createServiceReq)
		sReq, _ := http.NewRequest("POST", "/api/v1/services", bytes.NewBuffer(sPayloadBytes))
		sReq.Header.Set("Content-Type", "application/json")
		sReq.Header.Set("Authorization", "Bearer "+adminToken)
		sW := httptest.NewRecorder()
		rtr.ServeHTTP(sW, sReq)
		require.Equal(t, http.StatusCreated, sW.Code, "Failed to create service for ST delete conflict test")
		var sRespBody struct {
			Data model.Service `json:"data"`
		}
		err := json.Unmarshal(sW.Body.Bytes(), &sRespBody)
		require.NoError(t, err)
		createdServiceID = sRespBody.Data.ID // Use the package-level var
		require.NotZero(t, createdServiceID, "Service ID for conflict test should not be zero")

		// Teardown for the created service
		t.Cleanup(func() {
			if createdServiceID != 0 {
				deleteServiceURL := fmt.Sprintf("/api/v1/services/%d", createdServiceID)
				cleanupReq, _ := http.NewRequest("DELETE", deleteServiceURL, nil)
				cleanupReq.Header.Set("Authorization", "Bearer "+adminToken)
				cleanupW := httptest.NewRecorder()
				rtr.ServeHTTP(cleanupW, cleanupReq)
				if cleanupW.Code != http.StatusNoContent && cleanupW.Code != http.StatusOK {
					t.Logf("Failed to cleanup service %d for ST delete conflict test. Status: %d", createdServiceID, cleanupW.Code)
				}
				createdServiceID = 0 // Reset after deletion attempt
			}
		})

		// 2. Attempt to delete the service type that is now in use
		deleteSTURL := fmt.Sprintf("/api/v1/service-types/%d", createdServiceTypeID)
		stReq, _ := http.NewRequest("DELETE", deleteSTURL, nil)
		stReq.Header.Set("Authorization", "Bearer "+adminToken)
		stW := httptest.NewRecorder()
		rtr.ServeHTTP(stW, stReq)

		assert.Equal(t, http.StatusConflict, stW.Code, "Expected status 409 Conflict when deleting a service type in use")
	})

	// --- Service API Test Cases ---

t.Run("CreateService_Success", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		serviceName := "Test Service " + uuid.NewString()[:8]
		createReq := model.CreateServiceRequest{
			Name:          serviceName,
			Description:   "A test service description",
			Version:       "1.0.0",
			Status:        model.ServiceStatusActive,
			ExternalLink:  "http://example.com/test-service",
			ServiceTypeID: createdServiceTypeID, // Use the ST created in TestMain/setup
		}
		payloadBytes, _ := json.Marshal(createReq)
		req, _ := http.NewRequest("POST", "/api/v1/services", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code, "Expected status 201 Created for successful service creation")

		var respBody struct {
			Data model.Service `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		require.NoError(t, err)

		assert.Equal(t, serviceName, respBody.Data.Name)
		assert.Equal(t, createReq.Description, respBody.Data.Description)
		assert.Equal(t, createReq.Version, respBody.Data.Version)
		assert.Equal(t, createReq.Status, respBody.Data.Status)
		assert.Equal(t, createReq.ExternalLink, respBody.Data.ExternalLink)
		assert.Equal(t, createdServiceTypeID, respBody.Data.ServiceTypeID)
		assert.NotZero(t, respBody.Data.ID)

		testServiceID = respBody.Data.ID // Store for subsequent tests
		// No t.Cleanup here for testServiceID, as it's needed by other service tests.
		// It will be cleaned up by the main TestServiceAPI teardown if necessary or by a specific delete test.
	})

	t.Run("CreateService_Fail_InvalidPayload", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		cases := []struct {
			name    string
			payload model.CreateServiceRequest
			expectedStatus int
		}{
			{
				name: "Missing Name",
				payload: model.CreateServiceRequest{
					ServiceTypeID: createdServiceTypeID,
					Status:        model.ServiceStatusActive,
				},
				expectedStatus: http.StatusBadRequest,
			},
			{
				name: "Missing ServiceTypeID",
				payload: model.CreateServiceRequest{
					Name:   "Service Without STID",
					Status: model.ServiceStatusActive,
				},
				expectedStatus: http.StatusBadRequest,
			},
			{
				name: "Invalid Status",
				payload: model.CreateServiceRequest{
					Name:          "Service With Invalid Status",
					ServiceTypeID: createdServiceTypeID,
					Status:        "invalid-status-value",
				},
				expectedStatus: http.StatusBadRequest,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				payloadBytes, _ := json.Marshal(tc.payload)
				req, _ := http.NewRequest("POST", "/api/v1/services", bytes.NewBuffer(payloadBytes))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+adminToken)
				w := httptest.NewRecorder()
				rtr.ServeHTTP(w, req)
				assert.Equal(t, tc.expectedStatus, w.Code, "Expected status %d for %s", tc.expectedStatus, tc.name)
			})
		}
	})

	t.Run("CreateService_Fail_ServiceTypeNotFound", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		nonExistentSTID := uint(999998)
		createReq := model.CreateServiceRequest{
			Name:          "Service With NonExistent STID",
			ServiceTypeID: nonExistentSTID,
			Status:        model.ServiceStatusActive,
		}
		payloadBytes, _ := json.Marshal(createReq)
		req, _ := http.NewRequest("POST", "/api/v1/services", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status 400 Bad Request when service type ID does not exist")
	})

	t.Run("CreateService_Fail_DuplicateName", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		// Attempt to create a service with the same name as the one created in CreateService_Success (its ID is in testServiceID)
		// To get its name, we would ideally fetch it, but for this test, let's assume we know its name or re-use the name from CreateService_Success.
		// For robustness, it's better to fetch the service created by testServiceID and use its name.
		// However, testServiceID might not be set if CreateService_Success failed or was skipped.
		// So, we'll try to create a service, then try to create it again with the same name.

		firstServiceName := "Unique Service Name For Duplicate Test " + uuid.NewString()[:8]
		firstCreateReq := model.CreateServiceRequest{
			Name:          firstServiceName,
			ServiceTypeID: createdServiceTypeID,
			Status:        model.ServiceStatusActive,
		}
		payload1Bytes, _ := json.Marshal(firstCreateReq)
		req1, _ := http.NewRequest("POST", "/api/v1/services", bytes.NewBuffer(payload1Bytes))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("Authorization", "Bearer "+adminToken)
		w1 := httptest.NewRecorder()
		rtr.ServeHTTP(w1, req1)
		require.Equal(t, http.StatusCreated, w1.Code, "Failed to create first service for duplicate name test")
		var resp1Body struct { Data model.Service `json:"data"` }
		err := json.Unmarshal(w1.Body.Bytes(), &resp1Body)
		require.NoError(t, err)
		firstServiceID := resp1Body.Data.ID
		t.Cleanup(func() { // Cleanup the first service
			if firstServiceID != 0 {
				deleteURL := fmt.Sprintf("/api/v1/services/%d", firstServiceID)
				cleanupReq, _ := http.NewRequest("DELETE", deleteURL, nil)
				cleanupReq.Header.Set("Authorization", "Bearer "+adminToken)
				cleanupW := httptest.NewRecorder()
				rtr.ServeHTTP(cleanupW, cleanupReq)
			}
		})

		// Attempt to create another service with the same name
		duplicateCreateReq := model.CreateServiceRequest{
			Name:          firstServiceName, // Same name
			ServiceTypeID: createdServiceTypeID,
			Status:        model.ServiceStatusDevelopment,
		}
		payload2Bytes, _ := json.Marshal(duplicateCreateReq)
		req2, _ := http.NewRequest("POST", "/api/v1/services", bytes.NewBuffer(payload2Bytes))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+adminToken)
		w2 := httptest.NewRecorder()
		rtr.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusConflict, w2.Code, "Expected status 409 Conflict for duplicate service name")
	})

	t.Run("GetServiceByID_Success", func(t *testing.T) {
		// This test relies on testServiceID being set by CreateService_Success
		require.NotZero(t, testServiceID, "testServiceID must be set by CreateService_Success for this test to run")

		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		getURL := fmt.Sprintf("/api/v1/services/%d", testServiceID)
		req, _ := http.NewRequest("GET", getURL, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK for getting service by ID")

		var respBody struct {
			Data model.Service `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		require.NoError(t, err)

		assert.Equal(t, testServiceID, respBody.Data.ID)
		// Add more assertions if specific fields from CreateService_Success are known and stable
		// For example, if serviceName from CreateService_Success was stored package-level or re-fetched.
		// For now, checking ID and ServiceTypeID is a good start.
		assert.Equal(t, createdServiceTypeID, respBody.Data.ServiceTypeID, "ServiceTypeID should match the one used during creation")
	})

	t.Run("GetServiceByID_Fail_NotFound", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		nonExistentID := uint(999999) // A clearly non-existent ID
		getURL := fmt.Sprintf("/api/v1/services/%d", nonExistentID)
		req, _ := http.NewRequest("GET", getURL, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status 404 Not Found for non-existent service ID")
	})

	t.Run("ListServices_Success_WithDataAndPagination", func(t *testing.T) {
		// This test relies on testServiceID (and thus createdServiceTypeID) being set.
		require.NotZero(t, testServiceID, "testServiceID must be set for this test to run")
		require.NotZero(t, createdServiceTypeID, "createdServiceTypeID must be set for this test to run")

		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		// Create a couple more auxiliary services for pagination and filtering testing
		auxServiceNames := []string{
			"Paging Service A " + uuid.NewString()[:7],
			"Paging Service B " + uuid.NewString()[:7],
		}
		createdAuxServiceIDs := []uint{}

		for _, name := range auxServiceNames {
			createReq := model.CreateServiceRequest{
				Name:          name,
				ServiceTypeID: createdServiceTypeID, // Link to the same service type
				Status:        model.ServiceStatusActive,
			}
			payload, _ := json.Marshal(createReq)
			req, _ := http.NewRequest("POST", "/api/v1/services", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+adminToken)
			w := httptest.NewRecorder()
			rtr.ServeHTTP(w, req)
			require.Equal(t, http.StatusCreated, w.Code, "Failed to create aux service for list test: "+name)
			var resp struct{ Data model.Service `json:"data"` }
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			createdAuxServiceIDs = append(createdAuxServiceIDs, resp.Data.ID)
		}

		t.Cleanup(func() {
			for _, id := range createdAuxServiceIDs {
				deleteURL := fmt.Sprintf("/api/v1/services/%d", id)
				cleanupReq, _ := http.NewRequest("DELETE", deleteURL, nil)
				cleanupReq.Header.Set("Authorization", "Bearer "+adminToken)
				cleanupW := httptest.NewRecorder()
				rtr.ServeHTTP(cleanupW, cleanupReq)
				if cleanupW.Code != http.StatusNoContent && cleanupW.Code != http.StatusOK {
					t.Logf("Failed to cleanup aux service %d for list test. Status: %d", id, cleanupW.Code)
				}
			}
		})

		// Assuming at least 3 services now (original testServiceID + 2 aux) for createdServiceTypeID
		// Test fetching page 1 with page size 2, filtered by serviceTypeId, ordered by name
		listURL := fmt.Sprintf("/api/v1/services?page=1&pageSize=2&serviceTypeId=%d&orderBy=name&sortOrder=asc", createdServiceTypeID)
		reqList, _ := http.NewRequest("GET", listURL, nil)
		reqList.Header.Set("Authorization", "Bearer "+adminToken)
		wList := httptest.NewRecorder()
		rtr.ServeHTTP(wList, reqList)

		require.Equal(t, http.StatusOK, wList.Code, "Expected status 200 OK for listing services")

		var listResp struct {
			Data struct {
				Items    []model.Service `json:"items"`
				Total    int64            `json:"total"`
				Page     int              `json:"page"`
				PageSize int              `json:"pageSize"`
			} `json:"data"`
		}
		err := json.Unmarshal(wList.Body.Bytes(), &listResp)
		require.NoError(t, err, "Failed to unmarshal list services response")

		assert.Len(t, listResp.Data.Items, 2, "Expected 2 services on page 1 with pageSize 2 for the given serviceTypeId")
		assert.GreaterOrEqual(t, listResp.Data.Total, int64(3), "Expected total count to be at least 3 for the given serviceTypeId")
		assert.Equal(t, 1, listResp.Data.Page, "Expected current page to be 1")
		assert.Equal(t, 2, listResp.Data.PageSize, "Expected page size to be 2")

		for _, item := range listResp.Data.Items {
			assert.Equal(t, createdServiceTypeID, item.ServiceTypeID, "All listed services should belong to the filtered serviceTypeId")
		}
		// Check if items are sorted by name ascending
		if len(listResp.Data.Items) == 2 {
			assert.LessOrEqual(t, listResp.Data.Items[0].Name, listResp.Data.Items[1].Name, "Services should be sorted by name ascending")
		}
	})

	t.Run("UpdateService_Success", func(t *testing.T) {
		require.NotZero(t, testServiceID, "testServiceID must be set for this test to run")
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		// Create another service type to test updating ServiceTypeID
		otherSTName := "Other ST For Service Update " + uuid.NewString()[:8]
		createOtherSTReq := model.CreateServiceTypeRequest{Name: otherSTName}
		stPayload, _ := json.Marshal(createOtherSTReq)
		stReqCreate, _ := http.NewRequest("POST", "/api/v1/service-types", bytes.NewBuffer(stPayload))
		stReqCreate.Header.Set("Content-Type", "application/json")
		stReqCreate.Header.Set("Authorization", "Bearer "+adminToken)
		stWCreate := httptest.NewRecorder()
		rtr.ServeHTTP(stWCreate, stReqCreate)
		require.Equal(t, http.StatusCreated, stWCreate.Code)
		var stRespCreate struct{ Data model.ServiceType `json:"data"` }
		err := json.Unmarshal(stWCreate.Body.Bytes(), &stRespCreate)
		require.NoError(t, err)
		otherServiceTypeID := stRespCreate.Data.ID
		require.NotZero(t, otherServiceTypeID)

		t.Cleanup(func() {
			deleteURL := fmt.Sprintf("/api/v1/service-types/%d", otherServiceTypeID)
			cleanupReq, _ := http.NewRequest("DELETE", deleteURL, nil)
			cleanupReq.Header.Set("Authorization", "Bearer "+adminToken)
			cleanupW := httptest.NewRecorder()
			rtr.ServeHTTP(cleanupW, cleanupReq)
		})

		updatedDesc := "Updated Service Description " + uuid.NewString()[:8]
		updatedVersion := "1.0.1-updated"
		updatedStatus := model.ServiceStatusInactive
		updatedLink := "http://example.com/updated-service-link"

		updateReq := model.UpdateServiceRequest{
			// Name is not updated here to avoid conflict with duplicate name tests; tested separately.
			Description:   model.StringPtr(updatedDesc),
			Version:       model.StringPtr(updatedVersion),
			Status:        &updatedStatus,
			ExternalLink:  model.StringPtr(updatedLink),
			ServiceTypeID: &otherServiceTypeID, // Update to the new service type
		}
		payloadBytes, _ := json.Marshal(updateReq)
		updateURL := fmt.Sprintf("/api/v1/services/%d", testServiceID)
		req, _ := http.NewRequest("PUT", updateURL, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK for successful service update")

		var respBody struct {
			Data model.Service `json:"data"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &respBody)
		require.NoError(t, err)

		assert.Equal(t, updatedDesc, respBody.Data.Description)
		assert.Equal(t, updatedVersion, respBody.Data.Version)
		assert.Equal(t, updatedStatus, respBody.Data.Status)
		assert.Equal(t, updatedLink, respBody.Data.ExternalLink)
		assert.Equal(t, otherServiceTypeID, respBody.Data.ServiceTypeID, "ServiceTypeID should be updated")
		assert.Equal(t, testServiceID, respBody.Data.ID) // ID should not change
	})

	t.Run("UpdateService_Fail_NotFound", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		nonExistentID := uint(999888)
		updateReq := model.UpdateServiceRequest{Description: model.StringPtr("Attempt to update non-existent")}
		payloadBytes, _ := json.Marshal(updateReq)
		updateURL := fmt.Sprintf("/api/v1/services/%d", nonExistentID)
		req, _ := http.NewRequest("PUT", updateURL, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status 404 Not Found for updating non-existent service")
	})

	t.Run("UpdateService_Fail_InvalidPayload", func(t *testing.T) {
		require.NotZero(t, testServiceID, "testServiceID must be set for this test to run")
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		invalidStatusValue := model.ServiceStatus("this-is-not-a-valid-status") // Cast to model.ServiceStatus
		updateReq := model.UpdateServiceRequest{Status: &invalidStatusValue}
		payloadBytes, _ := json.Marshal(updateReq)
		updateURL := fmt.Sprintf("/api/v1/services/%d", testServiceID)
		req, _ := http.NewRequest("PUT", updateURL, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status 400 Bad Request for invalid status value in update")
	})

	t.Run("UpdateService_Fail_DuplicateName", func(t *testing.T) {
		require.NotZero(t, testServiceID, "testServiceID must be set for this test to run")
		require.NotZero(t, createdServiceTypeID, "createdServiceTypeID must be set for this test to run")
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		// 1. Create an auxiliary service
		auxServiceName := "Aux Service For Duplicate Update Test " + uuid.NewString()[:8]
		createAuxReq := model.CreateServiceRequest{
			Name:          auxServiceName,
			ServiceTypeID: createdServiceTypeID,
			Status:        model.ServiceStatusActive,
		}
		auxPayloadBytes, _ := json.Marshal(createAuxReq)
		auxReq, _ := http.NewRequest("POST", "/api/v1/services", bytes.NewBuffer(auxPayloadBytes))
		auxReq.Header.Set("Content-Type", "application/json")
		auxReq.Header.Set("Authorization", "Bearer "+adminToken)
		auxW := httptest.NewRecorder()
		rtr.ServeHTTP(auxW, auxReq)
		require.Equal(t, http.StatusCreated, auxW.Code, "Failed to create auxiliary service for duplicate name update test")
		var auxRespBody struct{ Data model.Service `json:"data"` }
		err := json.Unmarshal(auxW.Body.Bytes(), &auxRespBody)
		require.NoError(t, err)
		auxServiceID := auxRespBody.Data.ID
		require.NotZero(t, auxServiceID)

		t.Cleanup(func() { // Cleanup the auxiliary service
			deleteURL := fmt.Sprintf("/api/v1/services/%d", auxServiceID)
			cleanupReq, _ := http.NewRequest("DELETE", deleteURL, nil)
			cleanupReq.Header.Set("Authorization", "Bearer "+adminToken)
			cleanupW := httptest.NewRecorder()
			rtr.ServeHTTP(cleanupW, cleanupReq)
		})

		// 2. Attempt to update testServiceID's name to auxServiceName
		updateReq := model.UpdateServiceRequest{Name: model.StringPtr(auxServiceName)}
		updatePayloadBytes, _ := json.Marshal(updateReq)
		updateURL := fmt.Sprintf("/api/v1/services/%d", testServiceID)
		updateAttemptReq, _ := http.NewRequest("PUT", updateURL, bytes.NewBuffer(updatePayloadBytes))
		updateAttemptReq.Header.Set("Content-Type", "application/json")
		updateAttemptReq.Header.Set("Authorization", "Bearer "+adminToken)
		updateAttemptW := httptest.NewRecorder()
		rtr.ServeHTTP(updateAttemptW, updateAttemptReq)

		assert.Equal(t, http.StatusConflict, updateAttemptW.Code, "Expected status 409 Conflict when updating service name to an existing one")
	})

	t.Run("UpdateService_Fail_ServiceTypeNotFound", func(t *testing.T) {
		require.NotZero(t, testServiceID, "testServiceID must be set for this test to run")
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		nonExistentSTID := uint(999777)
		updateReq := model.UpdateServiceRequest{ServiceTypeID: &nonExistentSTID}
		payloadBytes, _ := json.Marshal(updateReq)
		updateURL := fmt.Sprintf("/api/v1/services/%d", testServiceID)
		req, _ := http.NewRequest("PUT", updateURL, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status 400 Bad Request when updating service with non-existent ServiceTypeID")
	})

	t.Run("DeleteService_Success", func(t *testing.T) {
		require.NotZero(t, testServiceID, "testServiceID must be set for this test to run")
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		deleteURL := fmt.Sprintf("/api/v1/services/%d", testServiceID)
		reqDelete, _ := http.NewRequest("DELETE", deleteURL, nil)
		reqDelete.Header.Set("Authorization", "Bearer "+adminToken)
		wDelete := httptest.NewRecorder()
		rtr.ServeHTTP(wDelete, reqDelete)

		assert.Equal(t, http.StatusNoContent, wDelete.Code, "Expected status 204 No Content for successful service deletion")

		// Verify it's actually deleted by trying to GET it
		reqGet, _ := http.NewRequest("GET", deleteURL, nil) // deleteURL is same as getURL for this ID
		reqGet.Header.Set("Authorization", "Bearer "+adminToken)
		wGet := httptest.NewRecorder()
		rtr.ServeHTTP(wGet, reqGet)
		assert.Equal(t, http.StatusNotFound, wGet.Code, "Expected status 404 Not Found when trying to GET a deleted service")

		// Mark testServiceID as deleted for clarity, though subsequent tests shouldn't rely on it anyway
		// testServiceID = 0 // No longer strictly necessary as teardown will handle ST which might cascade or be checked
	})

	t.Run("DeleteService_Fail_NotFound", func(t *testing.T) {
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		nonExistentID := uint(999666) // A clearly non-existent ID
		deleteURL := fmt.Sprintf("/api/v1/services/%d", nonExistentID)
		req, _ := http.NewRequest("DELETE", deleteURL, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status 404 Not Found for deleting non-existent service")
	})

	// --- Teardown: Delete the test service type ---
	t.Run("Teardown_DeleteTestServiceType", func(t *testing.T) {
		if createdServiceTypeID == 0 {
			t.Log("Skipping test service type deletion as ID is zero.")
			return
		}
		components := router.SetupTestApp(t)
		adminToken := router.GetAdminToken(t, components)
		rtr := components.Router

		deleteSTURL := fmt.Sprintf("/api/v1/service-types/%d", createdServiceTypeID)
		req, _ := http.NewRequest("DELETE", deleteSTURL, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)
		
		// Deletion should return 204 No Content or 200 OK.
		// If services are associated, it might return 409 Conflict.
		// For a clean teardown, ensure no services are linked or delete them first.
		// For now, assume a clean delete is possible.
		if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
			// Check if it's a 409 due to services still existing (if we created any and didn't clean up)
			// This part of teardown might need to be smarter or services deleted first.
			t.Logf("Attempted to delete service type %d. Status: %d, Body: %s", createdServiceTypeID, w.Code, w.Body.String())
			// For a simple teardown, we might just log and not fail hard if other tests created dependencies.
			// However, for a clean state, this should ideally pass with 200/204.
		} else {
			t.Logf("Test service type %d deleted successfully or was already gone.", createdServiceTypeID)
		}
	})
}
