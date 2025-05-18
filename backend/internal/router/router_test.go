package router_test // Use _test package suffix for integration-style tests

import (
	// Go Standard Library
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	// External Dependencies
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	// "go.uber.org/zap" // REMOVED: Not used
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	// Import gorm logger with alias
	// Internal Packages
	"EffiPlat/backend/internal/handler"
	"EffiPlat/backend/internal/models" // May need config for defaults if used by handlers/services
	// Import for DB logger integration if needed, though using simpler logger here
	// "EffiPlat/backend/internal/pkg/logger" // REMOVED: Not used
	// Package being tested
	// Ensure service package is imported
)

// Constants for testing
const (
	testUserEmail    = "testuser1@example.com"
	testUserPassword = "password"
	testJWTSecret    = "test_secret_key_for_router_tests" // Use a fixed secret for tests
)

// Helper to create a test user directly in the DB
func createTestUser(db *gorm.DB, email, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		Name:     "Test User",
		Email:    email,
		Password: string(hashedPassword),
		Status:   "active",
	}
	result := db.Create(user)
	return user, result.Error
}

// --- Test Cases ---

func TestHealthRoute(t *testing.T) {
	router, db, _, _, _, _, _ := setupAppTestRouter(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close() // Ensure DB connection is closed after test

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status": "UP"}`, w.Body.String())
}

func TestLoginRoute(t *testing.T) {
	router, db, _, _, _, _, _ := setupAppTestRouter(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 1. Setup: Create a user in the test DB
	_, err := createTestUser(db, testUserEmail, testUserPassword)
	assert.NoError(t, err)

	// 2. Test Successful Login
	loginPayload := models.LoginRequest{Email: testUserEmail, Password: testUserPassword}
	payloadBytes, _ := json.Marshal(loginPayload)
	reqSuccess, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(payloadBytes))
	reqSuccess.Header.Set("Content-Type", "application/json")

	wSuccess := httptest.NewRecorder()
	router.ServeHTTP(wSuccess, reqSuccess)

	assert.Equal(t, http.StatusOK, wSuccess.Code)
	// Check if the response body contains a token (basic check)
	var loginResp models.LoginResponse
	err = json.Unmarshal(wSuccess.Body.Bytes(), &loginResp)
	assert.NoError(t, err)
	assert.NotEmpty(t, loginResp.Token, "Token should not be empty on successful login")
	assert.NotNil(t, loginResp.User, "User data should be present on successful login")
	assert.Equal(t, testUserEmail, loginResp.User.Email)

	// 3. Test Login with Incorrect Password
	loginPayloadWrongPass := models.LoginRequest{Email: testUserEmail, Password: "wrongpassword"}
	payloadBytesWrongPass, _ := json.Marshal(loginPayloadWrongPass)
	reqWrongPass, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(payloadBytesWrongPass))
	reqWrongPass.Header.Set("Content-Type", "application/json")

	wWrongPass := httptest.NewRecorder()
	router.ServeHTTP(wWrongPass, reqWrongPass)

	assert.Equal(t, http.StatusUnauthorized, wWrongPass.Code)
	assert.Contains(t, wWrongPass.Body.String(), "invalid credentials")

	// 4. Test Login with Non-existent User
	loginPayloadNotFound := models.LoginRequest{Email: "notfound@example.com", Password: "password"}
	payloadBytesNotFound, _ := json.Marshal(loginPayloadNotFound)
	reqNotFound, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(payloadBytesNotFound))
	reqNotFound.Header.Set("Content-Type", "application/json")

	wNotFound := httptest.NewRecorder()
	router.ServeHTTP(wNotFound, reqNotFound)

	assert.Equal(t, http.StatusUnauthorized, wNotFound.Code) // GORM ErrRecordNotFound maps to Unauthorized in handler
	assert.Contains(t, wNotFound.Body.String(), "invalid credentials")
}

func TestGetMeRoute(t *testing.T) {
	router, db, _, _, _, _, _ := setupAppTestRouter(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 1. Setup: Create user and get a valid token via login
	createdUser, err := createTestUser(db, testUserEmail, testUserPassword)
	assert.NoError(t, err)

	loginPayload := models.LoginRequest{Email: testUserEmail, Password: testUserPassword}
	payloadBytes, _ := json.Marshal(loginPayload)
	loginReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(payloadBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)
	assert.Equal(t, http.StatusOK, loginW.Code)
	var loginResp models.LoginResponse
	err = json.Unmarshal(loginW.Body.Bytes(), &loginResp)
	assert.NoError(t, err)
	validToken := loginResp.Token

	// 2. Test /me with Valid Token
	reqValid, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	reqValid.Header.Set("Authorization", "Bearer "+validToken)

	wValid := httptest.NewRecorder()
	router.ServeHTTP(wValid, reqValid)

	assert.Equal(t, http.StatusOK, wValid.Code)
	// Check response body contains expected user info
	var meResp map[string]interface{}
	err = json.Unmarshal(wValid.Body.Bytes(), &meResp)
	assert.NoError(t, err)
	assert.Equal(t, float64(createdUser.ID), meResp["id"]) // JSON numbers are float64
	assert.Equal(t, createdUser.Name, meResp["name"])
	assert.Equal(t, createdUser.Email, meResp["email"])

	// 3. Test /me without Token
	reqNoToken, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	wNoToken := httptest.NewRecorder()
	router.ServeHTTP(wNoToken, reqNoToken)
	assert.Equal(t, http.StatusUnauthorized, wNoToken.Code)
	assert.Contains(t, wNoToken.Body.String(), "missing or invalid token")

	// 4. Test /me with Invalid Token
	reqInvalidToken, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	reqInvalidToken.Header.Set("Authorization", "Bearer invalidtoken123")
	wInvalidToken := httptest.NewRecorder()
	router.ServeHTTP(wInvalidToken, reqInvalidToken)
	assert.Equal(t, http.StatusUnauthorized, wInvalidToken.Code)
	assert.Contains(t, wInvalidToken.Body.String(), "invalid token")
}

func TestLogoutRoute(t *testing.T) {
	router, db, _, _, _, _, _ := setupAppTestRouter(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 1. Setup: Get a valid token
	_, err := createTestUser(db, testUserEmail, testUserPassword)
	assert.NoError(t, err)
	loginPayload := models.LoginRequest{Email: testUserEmail, Password: testUserPassword}
	payloadBytes, _ := json.Marshal(loginPayload)
	loginReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(payloadBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)
	assert.Equal(t, http.StatusOK, loginW.Code)
	var loginResp models.LoginResponse
	err = json.Unmarshal(loginW.Body.Bytes(), &loginResp)
	assert.NoError(t, err)
	validToken := loginResp.Token

	// 2. Test /logout with Valid Token
	reqValid, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
	reqValid.Header.Set("Authorization", "Bearer "+validToken)

	wValid := httptest.NewRecorder()
	router.ServeHTTP(wValid, reqValid)

	assert.Equal(t, http.StatusOK, wValid.Code)
	assert.Contains(t, wValid.Body.String(), "logout successful")

	// 3. Test /logout without Token (should fail at middleware)
	reqNoToken, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
	wNoToken := httptest.NewRecorder()
	router.ServeHTTP(wNoToken, reqNoToken)
	assert.Equal(t, http.StatusUnauthorized, wNoToken.Code)
}

func TestUserManagementRoutes(t *testing.T) {
	router, db, _, _, _, _, _ := setupAppTestRouter(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 1. 创建测试用户并获取 token
	createdUser, err := createTestUser(db, testUserEmail, testUserPassword)
	assert.NoError(t, err)

	// 登录获取 token
	loginPayload := models.LoginRequest{Email: testUserEmail, Password: testUserPassword}
	payloadBytes, _ := json.Marshal(loginPayload)
	loginReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(payloadBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)
	assert.Equal(t, http.StatusOK, loginW.Code)
	var loginResp models.LoginResponse
	err = json.Unmarshal(loginW.Body.Bytes(), &loginResp)
	assert.NoError(t, err)
	validToken := loginResp.Token

	// 2. 测试创建用户
	createUserPayload := struct {
		Name       string `json:"name"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		Department string `json:"department"`
	}{
		Name:       "New User",
		Email:      "newuser@example.com",
		Password:   "password123",
		Department: "IT",
	}
	createPayloadBytes, _ := json.Marshal(createUserPayload)
	createReq, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(createPayloadBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+validToken)
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)
	assert.Equal(t, http.StatusCreated, createW.Code)

	// 3. 测试获取用户列表
	listReq, _ := http.NewRequest("GET", "/api/v1/users", nil)
	listReq.Header.Set("Authorization", "Bearer "+validToken)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)
	assert.Equal(t, http.StatusOK, listW.Code)

	// 4. 测试获取特定用户
	getUserReq, _ := http.NewRequest("GET", "/api/v1/users/"+strconv.FormatUint(uint64(createdUser.ID), 10), nil)
	getUserReq.Header.Set("Authorization", "Bearer "+validToken)
	getUserW := httptest.NewRecorder()
	router.ServeHTTP(getUserW, getUserReq)
	assert.Equal(t, http.StatusOK, getUserW.Code)

	// 5. 测试更新用户
	updateUserPayload := struct {
		Name       *string `json:"name,omitempty"`
		Department *string `json:"department,omitempty"`
	}{
		Name:       stringPtr("Updated Name"),
		Department: stringPtr("Updated Department"),
	}
	updatePayloadBytes, _ := json.Marshal(updateUserPayload)
	updateReq, _ := http.NewRequest("PUT", "/api/v1/users/"+strconv.FormatUint(uint64(createdUser.ID), 10), bytes.NewBuffer(updatePayloadBytes))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+validToken)
	updateW := httptest.NewRecorder()
	router.ServeHTTP(updateW, updateReq)
	assert.Equal(t, http.StatusOK, updateW.Code)

	// 6. 测试删除用户
	deleteReq, _ := http.NewRequest("DELETE", "/api/v1/users/"+strconv.FormatUint(uint64(createdUser.ID), 10), nil)
	deleteReq.Header.Set("Authorization", "Bearer "+validToken)
	deleteW := httptest.NewRecorder()
	router.ServeHTTP(deleteW, deleteReq)
	assert.Equal(t, http.StatusNoContent, deleteW.Code)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// TODO: Add tests for other routes as they are implemented

// ADDED: TestRoleManagementRoutes
func TestRoleManagementRoutes(t *testing.T) {
	routerInstance, db, _, _, _, _, _ := setupAppTestRouter(t) // Renamed router to routerInstance to avoid conflict with package name
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 1. Create a test user and get a token (needed for authenticated routes)
	_, err := createTestUser(db, "role_test_user@example.com", "password")
	assert.NoError(t, err)

	loginPayload := models.LoginRequest{Email: "role_test_user@example.com", Password: "password"}
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

	var createdRoleID uint

	// 2. Test Create Role
	t.Run("Create Role", func(t *testing.T) {
		rolePayload := gin.H{
			"name":        "Test Role",
			"description": "A role for testing purposes",
		}
		payloadBytes, _ := json.Marshal(rolePayload)
		req, _ := http.NewRequest("POST", "/api/v1/roles", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+validToken)

		w := httptest.NewRecorder()
		routerInstance.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response struct {
			BizCode int         `json:"bizCode"`
			Message string      `json:"message"`
			Data    models.Role `json:"data"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Test Role", response.Data.Name)
		createdRoleID = response.Data.ID
		assert.NotZero(t, createdRoleID)
	})

	// 3. Test Get All Roles
	t.Run("Get All Roles", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/roles", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		w := httptest.NewRecorder()
		routerInstance.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response struct {
			BizCode int             `json:"bizCode"`
			Message string          `json:"message"`
			Data    json.RawMessage `json:"data"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.BizCode)

		var rolesData struct {
			Items    []models.Role `json:"items"`
			Total    int           `json:"total"`
			Page     int           `json:"page"`
			PageSize int           `json:"pageSize"`
		}
		err = json.Unmarshal(response.Data, &rolesData)
		assert.NoError(t, err)

		assert.True(t, rolesData.Total > 0, "Expected at least one role")
		assert.True(t, len(rolesData.Items) > 0, "Expected at least one role in items list")

		foundCreatedRole := false
		for _, role := range rolesData.Items {
			if role.ID == createdRoleID {
				assert.Equal(t, "Test Role", role.Name)
				foundCreatedRole = true
				break
			}
		}
		assert.True(t, foundCreatedRole, "Created role should be present in the list")
	})

	// 4. Test Get Role By ID
	t.Run("Get Role By ID", func(t *testing.T) {
		if createdRoleID == 0 {
			t.Skip("Skipping Get Role By ID test as createdRoleID is 0")
		}
		req, _ := http.NewRequest("GET", "/api/v1/roles/"+strconv.FormatUint(uint64(createdRoleID), 10), nil)
		req.Header.Set("Authorization", "Bearer "+validToken)

		w := httptest.NewRecorder()
		routerInstance.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response struct {
			BizCode int             `json:"bizCode"`
			Message string          `json:"message"`
			Data    json.RawMessage `json:"data"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		var fetchedRole models.Role
		err = json.Unmarshal(response.Data, &fetchedRole)
		assert.NoError(t, err)
		assert.Equal(t, createdRoleID, fetchedRole.ID)
		assert.Equal(t, "Test Role", fetchedRole.Name)

		// Check the response body against the expected role
		var updatedRole models.Role
		err = json.Unmarshal(response.Data, &updatedRole)
		assert.NoError(t, err)
		assert.Equal(t, "Test Role", updatedRole.Name)

		// Also verify Get Role By ID after update
		t.Run("Get Role By ID After Update", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/roles/"+strconv.FormatUint(uint64(createdRoleID), 10), nil)
			req.Header.Set("Authorization", "Bearer "+validToken)

			w := httptest.NewRecorder()
			routerInstance.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var getResp struct {
				BizCode int             `json:"bizCode"`
				Message string          `json:"message"`
				Data    json.RawMessage `json:"data"`
			}
			err = json.Unmarshal(w.Body.Bytes(), &getResp)
			assert.NoError(t, err)

			var fetchedUpdatedRole models.RoleDetails
			err = json.Unmarshal(getResp.Data, &fetchedUpdatedRole)
			assert.NoError(t, err)

			assert.Equal(t, "Test Role", fetchedUpdatedRole.Name)
			assert.Equal(t, 0, getResp.BizCode)
		})
	})

	// Test Update Role
	t.Run("Update Role", func(t *testing.T) {
		if createdRoleID == 0 {
			t.Skip("Skipping Update Role test as createdRoleID is 0")
		}

		updateData := handler.UpdateRoleRequest{Name: "Updated Test Role", Description: "Updated description"}
		updatePayloadBytes, _ := json.Marshal(updateData)

		req, _ := http.NewRequest("PUT", "/api/v1/roles/"+strconv.FormatUint(uint64(createdRoleID), 10), bytes.NewBuffer(updatePayloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+validToken) // Assuming update requires auth

		w := httptest.NewRecorder()
		routerInstance.ServeHTTP(w, req)

		// Assert the status code for the update request
		assert.Equal(t, http.StatusOK, w.Code)

		// Check the response body for the update request against the expected updated role
		var updateResp struct {
			BizCode int             `json:"bizCode"`
			Message string          `json:"message"`
			Data    json.RawMessage `json:"data"` // Use json.RawMessage for the data payload
		}
		err = json.Unmarshal(w.Body.Bytes(), &updateResp)
		assert.NoError(t, err)
		assert.Equal(t, 0, updateResp.BizCode)

		var updatedRole models.Role // Assuming update returns the updated Role object in data
		err = json.Unmarshal(updateResp.Data, &updatedRole)
		assert.NoError(t, err)
		assert.Equal(t, updateData.Name, updatedRole.Name)
		assert.Equal(t, updateData.Description, updatedRole.Description)
		// Add more assertions if needed for the updated role object itself

		// Verify Get Role By ID after update
		t.Run("Get Role By ID After Update", func(t *testing.T) {
			reqGet, _ := http.NewRequest("GET", "/api/v1/roles/"+strconv.FormatUint(uint64(createdRoleID), 10), nil)
			reqGet.Header.Set("Authorization", "Bearer "+validToken) // Assuming Get by ID requires auth

			wGet := httptest.NewRecorder()
			routerInstance.ServeHTTP(wGet, reqGet)

			assert.Equal(t, http.StatusOK, wGet.Code)

			var getResp struct {
				BizCode int             `json:"bizCode"`
				Message string          `json:"message"`
				Data    json.RawMessage `json:"data"`
			}
			err = json.Unmarshal(wGet.Body.Bytes(), &getResp)
			assert.NoError(t, err)
			assert.Equal(t, 0, getResp.BizCode)

			var fetchedUpdatedRole models.RoleDetails // Assuming Get By ID returns RoleDetails
			err = json.Unmarshal(getResp.Data, &fetchedUpdatedRole)
			assert.NoError(t, err)

			assert.Equal(t, updateData.Name, fetchedUpdatedRole.Name)
			assert.Equal(t, updateData.Description, fetchedUpdatedRole.Description)
			// Add more assertions for RoleDetails if needed
		})
	})

	// Test Delete Role
	t.Run("Delete Role", func(t *testing.T) {
		if createdRoleID == 0 {
			t.Skip("Skipping Delete Role test as createdRoleID is 0")
		}
		assert.NotZero(t, createdRoleID, "createdRoleID should be set from Create Role test")
		req, _ := http.NewRequest("DELETE", "/api/v1/roles/"+strconv.FormatUint(uint64(createdRoleID), 10), nil)
		req.Header.Set("Authorization", "Bearer "+validToken)

		w := httptest.NewRecorder()
		routerInstance.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify role is deleted
		verifyReq, _ := http.NewRequest("GET", "/api/v1/roles/"+strconv.FormatUint(uint64(createdRoleID), 10), nil)
		verifyReq.Header.Set("Authorization", "Bearer "+validToken)
		verifyW := httptest.NewRecorder()
		routerInstance.ServeHTTP(verifyW, verifyReq)
		assert.Equal(t, http.StatusNotFound, verifyW.Code)
	})
}
