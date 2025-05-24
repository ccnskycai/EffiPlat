package router_test

import (
	"EffiPlat/backend/internal/factories"
	"EffiPlat/backend/internal/model"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"EffiPlat/backend/internal/router"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Helper function to log in a user and get a token for sub-tests
// This promotes isolation by allowing each sub-test to log in independently.
func getAuthTokenForSubTest(t *testing.T, rtr http.Handler, db *gorm.DB, email, password string) string {
	loginReqBody := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, email, password)
	req, err := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(loginReqBody))
	require.NoError(t, err, "Sub-test login: failed to create request")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "Sub-test login: request failed")

	var loginResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &loginResp)
	require.NoError(t, err, "Sub-test login: failed to unmarshal response")
	require.Equal(t, 0, loginResp.Code, "Sub-test login: response code not 0")
	require.NotEmpty(t, loginResp.Data.Token, "Sub-test login: token is empty")
	return loginResp.Data.Token
}

// TestUserRoleManagementRoutes covers assigning and removing roles from users.
func TestUserRoleManagementRoutes(t *testing.T) {
	// Global setup (like adminToken) is removed from here to be handled by each sub-test for isolation.

	// --- Test Case: Successfully Assign Roles to User ---
	t.Run("Assign_Roles_To_User_Success", func(t *testing.T) {
		components := router.SetupTestApp(t)
		db := components.DB
		rtr := components.Router

		// Create admin user for this sub-test
		adminEmail := "admin_assign_success@example.com"
		adminPassword := "password123"
		_, err := factories.CreateUser(db, &model.User{Name: "Admin Assign Success", Email: adminEmail, Password: adminPassword, Status: "active"})
		require.NoError(t, err)
		adminToken := getAuthTokenForSubTest(t, rtr, db, adminEmail, adminPassword)

		// 1. Create a target user
		targetUser, err := factories.CreateUser(db, &model.User{
			Name:  "Target User Assign",
			Email: "target_assign_ur@example.com",
		})
		require.NoError(t, err)
		require.NotNil(t, targetUser)

		// 2. Create some roles
		role1, err := factories.CreateRole(db, &model.Role{Name: "Role UR A"})
		role2, err := factories.CreateRole(db, &model.Role{Name: "Role UR B"})

		// 3. Prepare request to assign roles
		assignReqPayload := model.AssignRemoveRolesRequest{
			RoleIDs: []uint{role1.ID, role2.ID},
		}
		payloadBytes, _ := json.Marshal(assignReqPayload)

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// 4. Assertions
		assert.Equal(t, http.StatusOK, w.Code) // Or http.StatusNoContent if API returns that

		var respBody struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data"` // Data might be null or contain some info
		}
		err = json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, 0, respBody.Code)
		assert.Equal(t, "Roles assigned successfully to user", respBody.Message)

		// 5. (Optional) Verify roles were actually assigned by fetching the user
		var updatedUser model.User
		err = db.Preload("Roles").First(&updatedUser, targetUser.ID).Error
		assert.NoError(t, err)
		assert.Len(t, updatedUser.Roles, 2)
		roleIDsInUser := make(map[uint]bool)
		for _, r := range updatedUser.Roles {
			roleIDsInUser[r.ID] = true
		}
		assert.True(t, roleIDsInUser[role1.ID])
		assert.True(t, roleIDsInUser[role2.ID])
	})

	// --- Test Case: Assign Roles to User - User Not Found ---
	t.Run("Assign_Roles_To_User_User_Not_Found", func(t *testing.T) {
		// ISOLATED SETUP FOR THIS SUB-TEST
		components := router.SetupTestApp(t)
		db := components.DB
		rtr := components.Router

		// Create admin user and get token for this sub-test
		adminEmail := "admin_assign_notfound@example.com"
		adminPassword := "password123"
		_, err := factories.CreateUser(db, &model.User{Name: "Admin Assign NotFound", Email: adminEmail, Password: adminPassword, Status: "active"})
		require.NoError(t, err)
		adminToken := getAuthTokenForSubTest(t, rtr, db, adminEmail, adminPassword)

		// 1. Create a role (doesn't matter which, just need some valid role IDs)
		roleToAssign, err := factories.CreateRole(db, &model.Role{Name: "Role UR C AssignNotFound"})
		require.NoError(t, err)
		require.NotNil(t, roleToAssign)

		// 2. Prepare request with a non-existent user ID
		nonExistentUserID := uint(99999)
		assignReqPayload := model.AssignRemoveRolesRequest{
			RoleIDs: []uint{roleToAssign.ID},
		}
		payloadBytes, _ := json.Marshal(assignReqPayload)

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%d/roles", nonExistentUserID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// 3. Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)

		var respBody struct {
			// Assuming ErrorResponse structure from handler/common.go
			Code    int    `json:"code"` // HTTP status code, might be different from BizCode
			Message string `json:"message"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		// The 'Code' in ErrorResponse is the HTTP status code itself.
		assert.Equal(t, http.StatusNotFound, respBody.Code)
		assert.Contains(t, respBody.Message, fmt.Sprintf("user with id %d", nonExistentUserID))
		assert.Contains(t, respBody.Message, "not found")
	})

	// --- Test Case: Assign Roles to User - Invalid Role IDs ---
	t.Run("Assign_Roles_To_User_Invalid_Role_IDs", func(t *testing.T) {
		// ISOLATED SETUP FOR THIS SUB-TEST
		components := router.SetupTestApp(t)
		db := components.DB
		rtr := components.Router
		var err error // Declare err for reuse in this sub-test

		// Create admin user and get token for this sub-test
		adminEmail := fmt.Sprintf("admin_assign_invalid_ids@example.com")
		adminPassword := "password123"
		_, err = factories.CreateUser(db, &model.User{Name: "Admin Assign InvalidIDs", Email: adminEmail, Password: adminPassword, Status: "active"})
		require.NoError(t, err, "Failed to create admin user for assign invalid IDs test")
		adminToken := getAuthTokenForSubTest(t, rtr, db, adminEmail, adminPassword)

		// 1. Create a target user
		targetUser, err := factories.CreateUser(db, &model.User{
			Name:     "Target User Invalid Roles",
			Email:    "target_invalid_roles_ur_isolated@example.com", // Unique email
			Password: "password123",
			Status:   "active",
		})
		require.NoError(t, err, "Failed to create target user for assign invalid IDs test")
		require.NotNil(t, targetUser, "Target user must not be nil")

		// 2. Create one valid role
		validRole, err := factories.CreateRole(db, &model.Role{Name: "Role UR Valid ForInvalidTest Isolated"}) // Unique name
		require.NoError(t, err, "Failed to create valid role for assign invalid IDs test")
		require.NotNil(t, validRole, "Valid role must not be nil")
		nonExistentRoleID := uint(99998)

		// 3. Prepare request to assign roles with one valid and one invalid ID
		assignReqPayload := model.AssignRemoveRolesRequest{
			RoleIDs: []uint{validRole.ID, nonExistentRoleID},
		}
		payloadBytes, _ := json.Marshal(assignReqPayload)

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// 4. Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var respBody model.ErrorResponse // Use the standard error response model
		err = json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, respBody.Code)
		assert.Contains(t, respBody.Message, "one or more specified role IDs do not exist", "Error message should indicate role ID not found for assign")

		// 5. Verify that no roles were assigned if an invalid ID was provided
		var updatedUser model.User
		err = db.Preload("Roles").First(&updatedUser, targetUser.ID).Error
		assert.NoError(t, err)
		assert.Empty(t, updatedUser.Roles, "No roles should be assigned if transaction is rolled back due to invalid role ID")
	})

	// --- Test Case: Assign Roles to User - Empty Role IDs ---
	t.Run("Assign_Roles_To_User_Empty_Role_IDs", func(t *testing.T) {
		// ISOLATED SETUP FOR THIS SUB-TEST
		components := router.SetupTestApp(t)
		db := components.DB
		rtr := components.Router
		var err error // Declare err for reuse

		// Create admin user and get token for this sub-test
		sanitizedTestName := strings.ReplaceAll(strings.ToLower(t.Name()), "/", "_") // Sanitize t.Name()
		adminEmail := "admin_" + sanitizedTestName + "@example.com"
		adminPassword := "password123"
		_, err = factories.CreateUser(db, &model.User{Name: "Admin for " + t.Name(), Email: adminEmail, Password: adminPassword, Status: "active"})
		require.NoError(t, err, "Failed to create admin user for "+t.Name())
		adminToken := getAuthTokenForSubTest(t, rtr, db, adminEmail, adminPassword)

		// 1. Create a target user
		targetUserEmail := fmt.Sprintf("target_empty_roles_%s@example.com", uuid.New().String()[:8])
		targetUser, err := factories.CreateUser(db, &model.User{
			Name:  "Target User Empty Roles",
			Email: targetUserEmail,
		})
		require.NoError(t, err, "Failed to create target user for assign empty roles test")
		require.NotNil(t, targetUser, "Target user must not be nil for assign empty roles test")

		// 2. Prepare request with empty RoleIDs
		assignReqPayload := model.AssignRemoveRolesRequest{
			RoleIDs: []uint{},
		}
		payloadBytes, _ := json.Marshal(assignReqPayload)

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// 3. Assertions - Assuming API treats this as a no-op success
		assert.Equal(t, http.StatusOK, w.Code)

		var respBody struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, 0, respBody.Code) // Success
		// Message might vary, e.g., "No roles specified for assignment" or "Roles assigned successfully" (even if none were)
		// Let's assume a generic success or a specific one if known.
		// For now, we can check it's not empty or contains "successfully"
		assert.Contains(t, respBody.Message, "successfully") // Or a more specific message

		// 4. Verify that no roles were assigned (user had no roles initially)
		var updatedUser model.User
		err = db.Preload("Roles").First(&updatedUser, targetUser.ID).Error
		assert.NoError(t, err, "User record should still be found")
		assert.Empty(t, updatedUser.Roles, "No roles should be assigned when an empty list is provided")
	})

	// --- Test Case: Assign Roles to User - Unauthorized ---
	t.Run("Assign_Roles_To_User_Unauthorized", func(t *testing.T) {
		// ISOLATED SETUP FOR THIS SUB-TEST
		components := router.SetupTestApp(t)
		db := components.DB
		rtr := components.Router
		var err error // Declare err for use with factories.CreateUser
		// No adminToken needed for this unauthorized test scenario's main request

		// 1. Create a target user (though not strictly necessary as auth should fail first)
		targetUser, err := factories.CreateUser(db, &model.User{
			Name:  "Target User Unauthorized Assign",
			Email: "target_unauth_assign_ur@example.com",
		})

		// 2. Prepare request (valid payload, but no token)
		assignReqPayload := model.AssignRemoveRolesRequest{
			RoleIDs: []uint{1}, // A dummy role ID, content doesn't matter for auth check
		}
		payloadBytes, _ := json.Marshal(assignReqPayload)

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header is set

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// 3. Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		// Instead of checking for code and message, check for the JWT middleware's error structure
		var errResp struct {
			Error string `json:"error"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &errResp) // Use '=' as err is declared in the sub-test's setup
		assert.NoError(t, err, "Failed to unmarshal unauthorized error response for assign roles")

		assert.Equal(t, http.StatusUnauthorized, w.Code) // HTTP status code is the primary check
		assert.Contains(t, errResp.Error, "missing or invalid token", "Error message should indicate a token issue for assign roles")
	})

	// --- Test Case: Assign Roles to User - Bad Request (Malformed JSON) ---
	t.Run("Assign_Roles_To_User_Bad_Request", func(t *testing.T) {
		// ISOLATED SETUP FOR THIS SUB-TEST
		components := router.SetupTestApp(t)
		db := components.DB
		rtr := components.Router
		var err error

		// Create admin user and get token for this sub-test
		adminEmail := "admin_bad_req_assign@example.com"
		adminPassword := "password123"
		_, err = factories.CreateUser(db, &model.User{Name: "Admin Bad Req Assign", Email: adminEmail, Password: adminPassword, Status: "active"})
		require.NoError(t, err, "Failed to create admin user for bad request assign test")
		adminToken := getAuthTokenForSubTest(t, rtr, db, adminEmail, adminPassword)

		// 1. Create a target user (ID is needed for the URL, user existence doesn't matter as parsing should fail first)
		targetUser, err := factories.CreateUser(db, &model.User{
			Name:  "Target User Bad Req Assign",
			Email: "target_badreq_assign_ur@example.com",
		})

		// 2. Prepare a malformed JSON payload
		malformedPayload := "{\"role_ids\": [1, 2]" // Missing closing brace

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBufferString(malformedPayload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken) // Auth token is present

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// 3. Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var respBody struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &respBody) // Use '=' as err is declared in the sub-test's setup
		assert.NoError(t, err)                          // Gin usually returns a parseable JSON error
		assert.Equal(t, http.StatusBadRequest, respBody.Code)
		// The exact message might depend on Gin's JSON parsing or your custom error handling
		// Common messages include "invalid character", "unexpected end of JSON input", etc.
		// For robustness, we can check if the message is non-empty or contains a known substring for bind errors.
		assert.NotEmpty(t, respBody.Message, "Error message should indicate a binding/parsing issue")
	})

	// --- Test Case: Successfully Remove Roles from User ---
	t.Run("Remove_Roles_From_User_Success", func(t *testing.T) {
		// ISOLATED SETUP FOR THIS SUB-TEST
		components := router.SetupTestApp(t)
		db := components.DB
		rtr := components.Router
		var err error

		// Create admin user and get token for this sub-test
		adminEmail := "admin_remove_success@example.com"
		adminPassword := "password123"
		_, err = factories.CreateUser(db, &model.User{Name: "Admin Remove Success", Email: adminEmail, Password: adminPassword, Status: "active"})
		require.NoError(t, err, "Failed to create admin user for remove success test")
		adminToken := getAuthTokenForSubTest(t, rtr, db, adminEmail, adminPassword)

		// 1. Create a target user
		targetUser, err := factories.CreateUser(db, &model.User{
			Name:  "Target User Remove Roles",
			Email: "target_remove_ur@example.com",
		})

		// 2. Create some roles
		roleToKeep, err := factories.CreateRole(db, &model.Role{Name: "Role UR Keep"})
		roleToRemove1, err := factories.CreateRole(db, &model.Role{Name: "Role UR Remove1"})
		roleToRemove2, err := factories.CreateRole(db, &model.Role{Name: "Role UR Remove2"})

		// 3. Assign roles directly to user in DB for setup
		db.Model(&targetUser).Association("Roles").Append([]*model.Role{roleToKeep, roleToRemove1, roleToRemove2})

		// Verify initial state
		var initialUser model.User
		db.Preload("Roles").First(&initialUser, targetUser.ID)
		assert.Len(t, initialUser.Roles, 3, "User should have 3 roles initially")

		// 4. Prepare request to remove two roles
		removeReqPayload := model.AssignRemoveRolesRequest{ // Using the same DTO
			RoleIDs: []uint{roleToRemove1.ID, roleToRemove2.ID},
		}
		payloadBytes, _ := json.Marshal(removeReqPayload)

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// 5. Assertions
		assert.Equal(t, http.StatusOK, w.Code) // Or http.StatusNoContent, depending on API design

		var respBody struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &respBody) // Use '=' as err is declared in the sub-test's setup
		assert.NoError(t, err)
		assert.Equal(t, 0, respBody.Code)
		assert.Equal(t, "Roles removed successfully from user", respBody.Message)

		// 6. Verify roles were actually removed and one was kept
		var updatedUser model.User
		err = db.Preload("Roles").First(&updatedUser, targetUser.ID).Error
		assert.NoError(t, err)
		assert.Len(t, updatedUser.Roles, 1)
		assert.Equal(t, roleToKeep.ID, updatedUser.Roles[0].ID)
		assert.Equal(t, roleToKeep.Name, updatedUser.Roles[0].Name)
	})

	// --- Test Case: Remove Roles from User - User Not Found ---
	t.Run("Remove_Roles_From_User_User_Not_Found", func(t *testing.T) {
		// ISOLATED SETUP FOR THIS SUB-TEST
		components := router.SetupTestApp(t)
		db := components.DB
		rtr := components.Router
		var err error // Declare err for reuse in this sub-test

		// Create admin user for this sub-test
		// Sanitize t.Name() by replacing / with _ for use in email
		sanitizedTestName := strings.ReplaceAll(strings.ToLower(t.Name()), "/", "_")
		adminEmail := "admin_" + sanitizedTestName + "@example.com"
		adminPassword := "password123"
		_, err = factories.CreateUser(db, &model.User{Name: "Admin for " + t.Name(), Email: adminEmail, Password: adminPassword, Status: "active"})
		require.NoError(t, err, "Failed to create admin user for "+t.Name())
		adminToken := getAuthTokenForSubTest(t, rtr, db, adminEmail, adminPassword)

		// 1. Create a role (doesn't matter which, just need some valid role IDs for the payload)
		roleToRemove, err := factories.CreateRole(db, &model.Role{Name: "Role UR ToRemove D"})
		require.NoError(t, err, "Failed to create role for "+t.Name())

		// 2. Prepare request with a non-existent user ID
		nonExistentUserID := uint(99997) // Ensure this ID is different from other tests
		removeReqPayload := model.AssignRemoveRolesRequest{
			RoleIDs: []uint{roleToRemove.ID},
		}
		payloadBytes, _ := json.Marshal(removeReqPayload)

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%d/roles", nonExistentUserID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// 3. Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)

		var respBody struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &respBody) // Use '=' as err is declared in the sub-test's setup
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, respBody.Code)
		assert.Contains(t, respBody.Message, fmt.Sprintf("user with id %d", nonExistentUserID))
		assert.Contains(t, respBody.Message, "not found")
	})

	// --- Test Case: Remove Roles from User - Role Not Assigned to User (should be no-op or specific error if designed that way) ---
	t.Run("Remove_Roles_From_User_Role_Not_Assigned", func(t *testing.T) {
		// ISOLATED SETUP FOR THIS SUB-TEST
		components := router.SetupTestApp(t)
		db := components.DB
		rtr := components.Router
		var err error // Declare err for reuse in this sub-test

		// Create admin user for this sub-test
		sanitizedTestName := strings.ReplaceAll(strings.ToLower(t.Name()), "/", "_")
		adminEmail := "admin_" + sanitizedTestName + "@example.com"
		adminPassword := "password123"
		_, err = factories.CreateUser(db, &model.User{Name: "Admin for " + t.Name(), Email: adminEmail, Password: adminPassword, Status: "active"})
		require.NoError(t, err, "Failed to create admin user for "+t.Name())
		adminToken := getAuthTokenForSubTest(t, rtr, db, adminEmail, adminPassword)

		// 1. Create a target user
		targetUser, err := factories.CreateUser(db, &model.User{
			Name:  "Target User Role Not Assigned",
			Email: "target_rnas_ur@example.com",
		})
		require.NoError(t, err, "Failed to create target user for "+t.Name())

		// 2. Create a role that will be assigned, and one that won't (but we'll try to remove it)
		assignedRole, err := factories.CreateRole(db, &model.Role{Name: "Role UR Assigned"})
		require.NoError(t, err, "Failed to create assignedRole for "+t.Name())
		unassignedRole, err := factories.CreateRole(db, &model.Role{Name: "Role UR Unassigned ButTryRemove"})
		require.NoError(t, err, "Failed to create unassignedRole for "+t.Name())

		// 3. Assign one role directly to user in DB for setup
		db.Model(&targetUser).Association("Roles").Append(assignedRole)

		// Verify initial state
		var initialUser model.User
		db.Preload("Roles").First(&initialUser, targetUser.ID)
		assert.Len(t, initialUser.Roles, 1, "User should have 1 role initially")
		assert.Equal(t, assignedRole.ID, initialUser.Roles[0].ID)

		// 4. Prepare request to remove the unassignedRole and the assignedRole
		removeReqPayload := model.AssignRemoveRolesRequest{
			RoleIDs: []uint{unassignedRole.ID, assignedRole.ID},
		}
		payloadBytes, _ := json.Marshal(removeReqPayload)

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// 5. Assertions - Should be success, as the system ensures the roles are not associated
		assert.Equal(t, http.StatusOK, w.Code)

		var respBody struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &respBody) // Use '=' as err is declared in the sub-test's setup
		assert.NoError(t, err)
		assert.Equal(t, 0, respBody.Code)
		assert.Equal(t, "Roles removed successfully from user", respBody.Message)

		// 6. Verify the assignedRole was removed and user has no roles now
		var updatedUser model.User
		err = db.Preload("Roles").First(&updatedUser, targetUser.ID).Error
		assert.NoError(t, err)
		assert.Empty(t, updatedUser.Roles, "User should have no roles after removal")
	})

	// --- Test Case: Remove Roles from User - Invalid Role IDs ---
	t.Run("Remove_Roles_From_User_Invalid_Role_IDs", func(t *testing.T) {
		// ISOLATED SETUP FOR THIS SUB-TEST
		components := router.SetupTestApp(t)
		db := components.DB
		rtr := components.Router
		var err error // Declare err for reuse in this sub-test

		// Create admin user for this sub-test
		sanitizedTestName := strings.ReplaceAll(strings.ToLower(t.Name()), "/", "_")
		adminEmail := "admin_" + sanitizedTestName + "@example.com"
		adminPassword := "password123"
		_, err = factories.CreateUser(db, &model.User{Name: "Admin for " + t.Name(), Email: adminEmail, Password: adminPassword, Status: "active"})
		require.NoError(t, err, "Failed to create admin user for "+t.Name())
		adminToken := getAuthTokenForSubTest(t, rtr, db, adminEmail, adminPassword)

		// 1. Create a target user
		targetUser, err := factories.CreateUser(db, &model.User{
			Name:  "Target User Remove Invalid Roles",
			Email: "target_remove_invalid_ur@example.com",
		})
		require.NoError(t, err, "Failed to create target user for "+t.Name())

		// 2. Ccreate a role that will be assigned (this is the role that remains on the user)
		assignedRole, err := factories.CreateRole(db, &model.Role{Name: "Role UR Assigned ForInvalidRemove"})
		require.NoError(t, err, "Failed to create assignedRole for "+t.Name())
		db.Model(&targetUser).Association("Roles").Append(assignedRole)

		// Verify initial state
		var initialUser model.User
		db.Preload("Roles").First(&initialUser, targetUser.ID)
		assert.Len(t, initialUser.Roles, 1, "User should have 1 role initially")

		// 3. Prepare request: We will attempt to remove a non-existent Role ID.
		// The 'validRoleNotAssigned' previously created here was unused in the actual test payload.
		nonExistentRoleID := uint(99996) // Ensure this ID is different
		// 确保这个ID真的不存在
		var roleCheck model.Role
		result := db.First(&roleCheck, nonExistentRoleID)
		require.Error(t, result.Error, "应该找不到ID为99996的角色")
		require.True(t, errors.Is(result.Error, gorm.ErrRecordNotFound), "错误应该是记录未找到")

		removeReqPayload := model.AssignRemoveRolesRequest{
			RoleIDs: []uint{nonExistentRoleID}, // Only the non-existent one to isolate
		}
		payloadBytes, _ := json.Marshal(removeReqPayload)

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// 4. Assertions - Expecting a 400 Bad Request due to non-existent role ID
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var respBody model.ErrorResponse               // Use the standard error response model
		err = json.Unmarshal(w.Body.Bytes(), &respBody) // Use '=' as err is declared in the sub-test's setup
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, respBody.Code)
		assert.Contains(t, respBody.Message, "one or more specified role IDs do not exist", "Error message should indicate role ID not found for remove")

		// 5. Verify that the user's original roles are untouched
		var updatedUser model.User
		err = db.Preload("Roles").First(&updatedUser, targetUser.ID).Error
		assert.NoError(t, err)
		assert.Len(t, updatedUser.Roles, 1, "User's roles should be unchanged")
		assert.Equal(t, assignedRole.ID, updatedUser.Roles[0].ID)
	})

	// --- Test Case: Remove Roles from User - Empty Role IDs ---
	t.Run("Remove_Roles_From_User_Empty_Role_IDs", func(t *testing.T) {
		// ISOLATED SETUP FOR THIS SUB-TEST
		components := router.SetupTestApp(t)
		db := components.DB
		rtr := components.Router
		var err error // Declare err for reuse in this sub-test

		// Create admin user for this sub-test
		sanitizedTestName := strings.ReplaceAll(strings.ToLower(t.Name()), "/", "_")
		adminEmail := "admin_" + sanitizedTestName + "@example.com"
		adminPassword := "password123"
		_, err = factories.CreateUser(db, &model.User{Name: "Admin for " + t.Name(), Email: adminEmail, Password: adminPassword, Status: "active"})
		require.NoError(t, err, "Failed to create admin user for "+t.Name())
		adminToken := getAuthTokenForSubTest(t, rtr, db, adminEmail, adminPassword)

		// 1. Create a user and assign a role
		testUserMail := "emptyrolesuser_" + sanitizedTestName + "@example.com" // Make email unique
		createdUser, err := factories.CreateUser(db, &model.User{
			Name:     "User for Empty Roles Test",
			Email:    testUserMail,
			Password: "password", // factories.CreateUser will hash this
			Status:   "active",
		})
		require.NoError(t, err, "Failed to create user for "+t.Name())

		viewerRole := model.Role{Name: "Viewer Role for Empty Test", Description: "Viewer desc"}
		db.Create(&viewerRole)
		db.Model(&createdUser).Association("Roles").Append(&viewerRole)

		// 2. Attempt to remove an empty list of roles
		payload := model.AssignRemoveRolesRequest{RoleIDs: []uint{}}
		payloadBytes, _ := json.Marshal(payload)
		req, _ := http.NewRequest("DELETE", "/api/v1/users/"+strconv.FormatUint(uint64(createdUser.ID), 10)+"/roles", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken) // Assuming adminToken is available from test setup

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req) // rtr should be the router instance

		require.NotNil(t, w, "ResponseRecorder should not be nil after ServeHTTP")
		require.NotNil(t, w.Body, "ResponseRecorder body should not be nil after ServeHTTP")

		t.Logf("Response body for Remove_Roles_From_User_No_Roles_To_Remove: %s", w.Body.String()) // Log the body

		// 3. Assertions - Should be success (idempotent, nothing to remove)
		assert.Equal(t, http.StatusOK, w.Code, "Removing roles from a user with no roles should be a successful operation (idempotent)")

		var respBody model.SuccessResponse                   // Expecting a standard success response
		parseErr := json.Unmarshal(w.Body.Bytes(), &respBody) // Assign to new var to avoid shadow
		assert.NoError(t, parseErr)
		assert.Equal(t, 0, respBody.Code) // MODIFIED BizCode to Code
		assert.Contains(t, respBody.Message, "Roles removed successfully from user", "Message should indicate success for empty list")

		// Verify user's roles are unchanged (should still be the initial role)
		var unchangedUser model.User
		db.Preload("Roles").First(&unchangedUser, createdUser.ID)
		assert.Len(t, unchangedUser.Roles, 1, "User should still have their initial role after attempting to remove an empty list")
		if len(unchangedUser.Roles) > 0 {
			assert.Equal(t, viewerRole.Name, unchangedUser.Roles[0].Name)
		}
	})

	// --- Test Case: Remove Roles from User - Unauthorized ---
	t.Run("Remove_Roles_From_User_Unauthorized", func(t *testing.T) {
		// ISOLATED SETUP FOR THIS SUB-TEST
		components := router.SetupTestApp(t)
		db := components.DB
		rtr := components.Router
		var err error // Declare err for reuse in this sub-test

		// Create a regular (non-admin) user for this sub-test to get an unauthorized token
		sanitizedTestName := strings.ReplaceAll(strings.ToLower(t.Name()), "/", "_")
		regularUserEmail := "regular_" + sanitizedTestName + "@example.com"
		regularUserPassword := "password123"
		regularUser, err := factories.CreateUser(db, &model.User{Name: "Regular User for " + t.Name(), Email: regularUserEmail, Password: regularUserPassword, Status: "active"})
		require.NoError(t, err, "Failed to create regular user for "+t.Name())
		_ = regularUser                                                                                // Prevent declared and not used if only token is used later
		unauthorizedToken := getAuthTokenForSubTest(t, rtr, db, regularUserEmail, regularUserPassword) // This token will be from a non-admin
		_ = unauthorizedToken                                                                          // Mark as used to prevent lint error, as we are intentionally not sending it.

		// 1. Create a target user (ID needed for URL, actual user state doesn't matter for auth check)
		// We still need an admin to create the target user and roles, if they were to be actually manipulated.
		// For this auth test, targetUser just needs to exist for the URL.
		adminEmailForSetup := "admin_setup_for_unauth_" + sanitizedTestName + "@example.com"
		// Create an admin user to facilitate test data creation.
		// The actual admin status is conferred by roles, not a direct field.
		adminForSetup, err := factories.CreateUser(db, &model.User{Name: "Admin For Setup " + t.Name(), Email: adminEmailForSetup, Password: "password123", Status: "active"})
		require.NoError(t, err, "Failed to create admin user for setup in "+t.Name())
		// Assign an admin role to this setup user if needed for it to create other entities.
		// For now, assuming factories.CreateUser is permissible without specific admin role for this basic setup.
		_ = adminForSetup // Prevents declared and not used if adminTokenForSetup is not used
		// adminTokenForSetup := getAuthTokenForSubTest(t, rtr, db, adminEmailForSetup, "password123") // Not strictly needed if not performing actions as admin

		targetUserForURL, err := factories.CreateUser(db, &model.User{Name: "Target User For Unauth URL", Email: "target_unauth_url_" + sanitizedTestName + "@example.com"})
		require.NoError(t, err, "Failed to create targetUserForURL for "+t.Name())

		// Create a role (doesn't matter which, just need some valid role IDs for the payload if the request got that far)
		roleForPayload, err := factories.CreateRole(db, &model.Role{Name: "Role For Unauth Payload"})
		require.NoError(t, err, "Failed to create roleForPayload for "+t.Name())

		// 2. Prepare request with a token from a non-admin user
		removeReqPayload := model.AssignRemoveRolesRequest{
			RoleIDs: []uint{roleForPayload.ID}, // Use the role created in setup
		}
		payloadBytes, err := json.Marshal(removeReqPayload)
		require.NoError(t, err) // Add error check for marshaling

		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%d/roles", targetUserForURL.ID), bytes.NewBuffer(payloadBytes)) // Use targetUserForURL
		require.NoError(t, err)                                                                                                                   // Add error check for new request
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header is set, to test the "missing or invalid token" path

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// 3. Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		// Instead of checking for code and message, check for the JWT middleware's error structure
		var errResp struct {
			Error string `json:"error"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &errResp) // Use '=' as err is declared in the sub-test's setup
		assert.NoError(t, err, "Failed to unmarshal unauthorized error response for remove roles")

		assert.Equal(t, http.StatusUnauthorized, w.Code) // HTTP status code is the primary check
		assert.Contains(t, errResp.Error, "missing or invalid token", "Error message should indicate a token issue for remove roles")
	})

	// --- Test Case: Remove Roles from User - Bad Request (Malformed JSON) ---
	t.Run("Remove_Roles_From_User_Bad_Request", func(t *testing.T) {
		// ISOLATED SETUP FOR THIS SUB-TEST
		components := router.SetupTestApp(t)
		db := components.DB
		rtr := components.Router
		var err error // Declare err for reuse in this sub-test

		// Create an admin user and get their token for this sub-test
		sanitizedTestName := strings.ReplaceAll(strings.ToLower(t.Name()), "/", "_") // Sanitize t.Name()
		adminEmail := "admin_" + sanitizedTestName + "@example.com"
		adminPassword := "password123"
		_, err = factories.CreateUser(db, &model.User{Name: "Admin for " + t.Name(), Email: adminEmail, Password: adminPassword, Status: "active"})
		require.NoError(t, err, "Failed to create admin user for "+t.Name())
		adminToken := getAuthTokenForSubTest(t, rtr, db, adminEmail, adminPassword)
		require.NotEmpty(t, adminToken, "Admin token should not be empty for "+t.Name())

		// 1. Create a target user (ID is needed for the URL)
		targetUser, err := factories.CreateUser(db, &model.User{
			Name:  "Target User Bad Req Remove",
			Email: "target_badreq_remove_ur@example.com",
		})
		require.NoError(t, err, "Failed to create target user for bad request test")

		// 2. Prepare a malformed JSON payload
		malformedPayload := "{\"role_ids\": [1" // Missing closing bracket and brace

		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBufferString(malformedPayload))
		require.NoError(t, err, "Failed to create HTTP request for bad request test")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken) // Valid token

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// 3. Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var respBody struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &respBody) // Use '=' as err is declared in the sub-test's setup
		assert.NoError(t, err)                          // Expecting a parseable JSON error from Gin
		assert.Equal(t, http.StatusBadRequest, respBody.Code)
		assert.NotEmpty(t, respBody.Message, "Error message should indicate a binding/parsing issue")
	})

	// --- Test Case: Remove Roles From User - No Roles To Remove (User has no roles) ---
	t.Run("Remove_Roles_From_User_No_Roles_To_Remove", func(t *testing.T) {
		// ISOLATED SETUP FOR THIS SUB-TEST
		components := router.SetupTestApp(t)
		db := components.DB
		rtr := components.Router

		// Create admin user and get token for this sub-test
		adminEmailForNoRolesTest := "admin_no_roles_remove@example.com"
		adminPasswordForNoRolesTest := "password123"
		_, err := factories.CreateUser(db, &model.User{Name: "Admin No Roles Remove", Email: adminEmailForNoRolesTest, Password: adminPasswordForNoRolesTest, Status: "active"})
		require.NoError(t, err, "Failed to create admin user for no roles remove test")
		adminToken := getAuthTokenForSubTest(t, rtr, db, adminEmailForNoRolesTest, adminPasswordForNoRolesTest)

		// 1. Create a user that has no roles initially
		userNoRoles, err := factories.CreateUser(db, &model.User{
			Name:     "User No Roles UR",
			Email:    "user_no_roles_ur@example.com", // Ensure unique email if tests run in parallel or share external resources
			Password: "password123",
			Status:   "active",
		})
		require.NoError(t, err, "Failed to create user for Remove_Roles_From_User_No_Roles_To_Remove test")
		require.NotNil(t, userNoRoles, "User (userNoRoles) must not be nil for Remove_Roles_From_User_No_Roles_To_Remove test")

		// 2. Prepare request to remove a role (any valid role ID, e.g., one created for another test or a new dummy one)
		// It doesn't matter if the role exists in DB, as long as the user doesn't have it.
		dummyRoleForNoRolesTest, err := factories.CreateRole(db, &model.Role{Name: "Dummy Role For No Roles Test UR"})
		require.NoError(t, err, "Failed to create dummy role for no roles test")
		require.NotNil(t, dummyRoleForNoRolesTest, "Dummy role must not be nil")

		removeReqPayload := model.AssignRemoveRolesRequest{
			RoleIDs: []uint{dummyRoleForNoRolesTest.ID},
		}
		payloadBytes, _ := json.Marshal(removeReqPayload)

		requestURL := fmt.Sprintf("/api/v1/users/%d/roles", userNoRoles.ID)

		req, err := http.NewRequest(http.MethodDelete, requestURL, bytes.NewBuffer(payloadBytes))
		require.NoError(t, err, "Failed to create HTTP request for Remove_Roles_From_User_No_Roles_To_Remove test")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req) // Use rtr from sub-test setup

		require.NotNil(t, w, "ResponseRecorder should not be nil after ServeHTTP")
		require.NotNil(t, w.Body, "ResponseRecorder body should not be nil after ServeHTTP")

		t.Logf("Response body for Remove_Roles_From_User_No_Roles_To_Remove: %s", w.Body.String())

		// 3. Assertions - Should be success (idempotent, nothing to remove)
		assert.Equal(t, http.StatusOK, w.Code, "Removing roles from a user with no roles should be a successful operation (idempotent)")

		var respBody struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data,omitempty"` // Added omitempty for data
		}
		err = json.Unmarshal(w.Body.Bytes(), &respBody)
		require.NoError(t, err, "Failed to unmarshal response body for Remove_Roles_From_User_No_Roles_To_Remove test")
		assert.Equal(t, 0, respBody.Code)
		assert.Equal(t, "Roles removed successfully from user", respBody.Message, "Response message for removing roles from user with no roles")

		// 4. Verify user still has no roles
		var updatedUser model.User
		err = db.Preload("Roles").First(&updatedUser, userNoRoles.ID).Error
		assert.NoError(t, err, "Failed to fetch user after attempting to remove roles from a user with no roles")
		assert.Empty(t, updatedUser.Roles, "User should still have no roles after an attempt to remove roles they don't have")
	})
}
