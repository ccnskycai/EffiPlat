package router_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"EffiPlat/backend/internal/models" // Assuming models are correctly imported

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	// Need to import service and repository packages, assuming the linter issues are resolved
	// "EffiPlat/backend/internal/service"
	// "EffiPlat/backend/internal/repository"
)

// TestPermissionManagementRoutes tests the permission management API endpoints.
func TestPermissionManagementRoutes(t *testing.T) {
	// Setup the test router with necessary dependencies.
	// Using the more complete setup function from user_roles_router_test.go
	routerInstance, db, _, _, _, _, _ := setupAppTestRouter(t) // Using the new setup function
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 1. Create a test user and get a token (needed for authenticated routes)
	// Assuming createTestUser is in router_test.go
	_, err := createTestUser(db, "perm_test_user@example.com", "password")
	assert.NoError(t, err)

	loginPayload := models.LoginRequest{Email: "perm_test_user@example.com", Password: "password"}
	loginPayloadBytes, _ := json.Marshal(loginPayload)
	loginReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginPayloadBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	routerInstance.ServeHTTP(loginW, loginReq)
	assert.Equal(t, http.StatusOK, loginW.Code)
	var loginResp models.LoginResponse
	err = json.Unmarshal(loginW.Body.Bytes(), &loginResp)
	assert.NoError(t, err)
	validToken := loginResp.Token

	var createdPermissionID uint
	var createdRoleID uint // Need a role to test association

	// Setup: Create a role to associate permissions with
	// Assuming role creation logic works based on previous fixes
	t.Run("Setup_Create_Role_For_Permission_Tests", func(t *testing.T) {
		rolePayload := gin.H{
			"name":        "Permission Test Role",
			"description": "Role for testing permission associations",
		}
		payloadBytes, _ := json.Marshal(rolePayload)
		req, _ := http.NewRequest("POST", "/api/v1/roles", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+validToken)

		w := httptest.NewRecorder()
		routerInstance.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    models.Role `json:"data"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		createdRoleID = response.Data.ID
		assert.NotZero(t, createdRoleID)
	})

	// 2. Test Create Permission
	t.Run("Create_Permission", func(t *testing.T) {
		permissionPayload := gin.H{
			"name":        "test:action",
			"description": "A permission for testing",
			"resource":    "test",
			"action":      "action",
		}
		payloadBytes, _ := json.Marshal(permissionPayload)
		req, _ := http.NewRequest("POST", "/api/v1/permissions", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+validToken)

		w := httptest.NewRecorder()
		routerInstance.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response struct {
			BizCode int               `json:"bizCode"`
			Message string            `json:"message"`
			Data    models.Permission `json:"data"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.BizCode)
		assert.Equal(t, "test:action", response.Data.Name)
		assert.Equal(t, "test", response.Data.Resource)
		assert.Equal(t, "action", response.Data.Action)
		createdPermissionID = response.Data.ID
		assert.NotZero(t, createdPermissionID)
	})

	// 3. Test Get All Permissions
	t.Run("Get_All_Permissions", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/permissions", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)

		w := httptest.NewRecorder()
		routerInstance.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response struct {
			BizCode int    `json:"bizCode"`
			Message string `json:"message"`
			Data    struct {
				Items    []models.Permission `json:"items"`
				Total    int64               `json:"total"`
				Page     int                 `json:"page"`
				pageSize int                 `json:"pageSize"` // Note: JSON tag should be "pageSize"
			} `json:"data"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.BizCode)
		assert.NotEmpty(t, response.Data.Items, "Permissions list should not be empty")
		// Check if the created permission is in the list
		found := false
		for _, perm := range response.Data.Items {
			if perm.ID == createdPermissionID {
				found = true
				break
			}
		}
		assert.True(t, found, "Created permission not found in list")
	})

	// 4. Test Get Permission By ID
	t.Run("Get_Permission_By_ID", func(t *testing.T) {
		assert.NotZero(t, createdPermissionID, "createdPermissionID should be set from Create Permission test")
		req, _ := http.NewRequest("GET", "/api/v1/permissions/"+strconv.FormatUint(uint64(createdPermissionID), 10), nil)
		req.Header.Set("Authorization", "Bearer "+validToken)

		w := httptest.NewRecorder()
		routerInstance.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response struct {
			BizCode int               `json:"bizCode"`
			Message string            `json:"message"`
			Data    models.Permission `json:"data"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.BizCode)
		assert.Equal(t, createdPermissionID, response.Data.ID)
		assert.Equal(t, "test:action", response.Data.Name)
	})

	// 5. Test Update Permission
	t.Run("Update_Permission", func(t *testing.T) {
		assert.NotZero(t, createdPermissionID, "createdPermissionID should be set from Create Permission test")
		updatePayload := gin.H{
			"description": "Updated testing description",
			"action":      "read", // Update action
		}
		payloadBytes, _ := json.Marshal(updatePayload)
		req, _ := http.NewRequest("PUT", "/api/v1/permissions/"+strconv.FormatUint(uint64(createdPermissionID), 10), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+validToken)

		w := httptest.NewRecorder()
		routerInstance.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response struct {
			BizCode int               `json:"bizCode"`
			Message string            `json:"message"`
			Data    models.Permission `json:"data"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.BizCode)
		assert.Equal(t, createdPermissionID, response.Data.ID)
		assert.Equal(t, "Updated testing description", response.Data.Description)
		assert.Equal(t, "read", response.Data.Action)
	})

	// 6. Test Add Permissions to Role
	t.Run("Add_Permissions_To_Role", func(t *testing.T) {
		assert.NotZero(t, createdRoleID, "createdRoleID should be set from Setup_Create_Role_For_Permission_Tests")
		assert.NotZero(t, createdPermissionID, "createdPermissionID should be set from Create_Permission test")

		permissionIDsToAdd := []uint{createdPermissionID}
		payloadBytes, _ := json.Marshal(permissionIDsToAdd)
		req, _ := http.NewRequest("POST", "/api/v1/permissions/roles/"+strconv.FormatUint(uint64(createdRoleID), 10), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+validToken)

		w := httptest.NewRecorder()
		routerInstance.ServeHTTP(w, req)

		// Assuming handler returns 200 with success response for now
		assert.Equal(t, http.StatusOK, w.Code)

		t.Logf("Raw JSON response body for AddPermissionsToRole: %s", w.Body.String()) // PRINT RAW JSON

		var response struct {
			BizCode int         `json:"bizCode"`
			Message string      `json:"message"`
			Data    interface{} `json:"data"` // Data is null on success
		}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.BizCode)
		assert.Equal(t, "Permissions added to role successfully", response.Message)
	})

	// 7. Test Get Permissions By Role ID
	t.Run("Get_Permissions_By_Role_ID", func(t *testing.T) {
		assert.NotZero(t, createdRoleID, "createdRoleID should be set")
		assert.NotZero(t, createdPermissionID, "createdPermissionID should be set")

		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/roles/%d/permissions", createdRoleID), nil)
		req.Header.Set("Authorization", "Bearer "+validToken)

		w := httptest.NewRecorder()
		routerInstance.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response struct {
			BizCode int                 `json:"bizCode"`
			Message string              `json:"message"`
			Data    []models.Permission `json:"data"` // Expecting a list of permissions
		}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.BizCode)
		assert.Equal(t, "Permissions for role retrieved successfully", response.Message)
		assert.NotEmpty(t, response.Data, "Should get associated permissions")
		// Verify the previously added permission is in the list
		found := false
		for _, perm := range response.Data {
			if perm.ID == createdPermissionID {
				found = true
				break
			}
		}
		assert.True(t, found, "Associated permission not found in list")
	})

	// 8. Test Remove Permissions from Role
	t.Run("Remove_Permissions_From_Role", func(t *testing.T) {
		assert.NotZero(t, createdRoleID, "createdRoleID should be set")
		assert.NotZero(t, createdPermissionID, "createdPermissionID should be set")

		permissionIDsToRemove := []uint{createdPermissionID}
		payloadBytes, _ := json.Marshal(permissionIDsToRemove)
		req, _ := http.NewRequest("DELETE", "/api/v1/permissions/roles/"+strconv.FormatUint(uint64(createdRoleID), 10), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+validToken)

		w := httptest.NewRecorder()
		routerInstance.ServeHTTP(w, req)

		// Assuming handler returns 200 with success response for now
		assert.Equal(t, http.StatusOK, w.Code)
		var response struct {
			BizCode int         `json:"bizCode"`
			Message string      `json:"message"`
			Data    interface{} `json:"data"` // Data is null on success
		}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.BizCode)
		assert.Equal(t, "Permissions removed from role successfully", response.Message)

		// Verify permission is no longer associated
		t.Run("Verify_Permissions_Removed", func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/roles/%d/permissions", createdRoleID), nil)
			req.Header.Set("Authorization", "Bearer "+validToken)
			routerInstance.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "Status code should be 200 after removing permissions")
			var listRespAfterRemove models.SuccessResponse
			err = json.Unmarshal(rr.Body.Bytes(), &listRespAfterRemove)
			assert.NoError(t, err, "Should unmarshal list response after removing permissions")
			assert.Equal(t, 0, listRespAfterRemove.BizCode, "Response bizCode in body should be 0 after removing")

			// Depending on how empty list is represented. Assuming data is an empty slice or nil.
			if listRespAfterRemove.Data != nil {
				permissionsList, ok := listRespAfterRemove.Data.([]interface{}) // GJSON might parse to []interface{}
				assert.True(t, ok, "Data should be a list of permissions")
				assert.Empty(t, permissionsList, "Permissions list should be empty after removal")
			} else {
				assert.Nil(t, listRespAfterRemove.Data, "Data should be nil if no permissions")
			}
		})
	})

	// 9. Test Delete Permission
	t.Run("Delete_Permission", func(t *testing.T) {
		assert.NotZero(t, createdPermissionID, "createdPermissionID should be set from Create Permission test")
		req, _ := http.NewRequest("DELETE", "/api/v1/permissions/"+strconv.FormatUint(uint64(createdPermissionID), 10), nil)
		req.Header.Set("Authorization", "Bearer "+validToken)

		w := httptest.NewRecorder()
		routerInstance.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code) // Expecting 204
	})

	// Teardown: Delete the created role
	t.Run("Teardown_Delete_Role", func(t *testing.T) {
		assert.NotZero(t, createdRoleID, "createdRoleID should be set")
		req, _ := http.NewRequest("DELETE", "/api/v1/roles/"+strconv.FormatUint(uint64(createdRoleID), 10), nil)
		req.Header.Set("Authorization", "Bearer "+validToken)

		w := httptest.NewRecorder()
		routerInstance.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code) // Expecting 204
	})
}
