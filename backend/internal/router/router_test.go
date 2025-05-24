package router_test // Use _test package suffix for integration-style tests

import (
	// Go Standard Library
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	// External Dependencies

	"github.com/stretchr/testify/assert"

	// "go.uber.org/zap" // REMOVED: Not used

	// Import gorm logger with alias
	// Internal Packages
	"EffiPlat/backend/internal/handler"
	"EffiPlat/backend/internal/model" // May need config for defaults if used by handlers/services
	"EffiPlat/backend/internal/router" // Added import for router.SetupTestApp
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

// Helper to create a test user directly in the DB - MOVED to router_test_helper.go as router.CreateTestUser

// Utility to get admin token (top-level) - MOVED to router_test_helper.go as router.GetAdminToken

// --- Test Cases ---

func TestHealthRoute(t *testing.T) {
	components := router.SetupTestApp(t)
	rtr := components.Router
	db := components.DB
	sqlDB, _ := db.DB()
	defer sqlDB.Close() // Ensure DB connection is closed after test

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	rtr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status": "UP"}`, w.Body.String())
}

func TestLoginRoute(t *testing.T) {
	components := router.SetupTestApp(t)
	rtr := components.Router
	db := components.DB
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 1. Setup: Create a user in the test DB
	_, err := router.CreateTestUser(db, testUserEmail, testUserPassword)
	assert.NoError(t, err)

	// 2. Test Successful Login
	loginPayload := model.LoginRequest{Email: testUserEmail, Password: testUserPassword}
	payloadBytes, _ := json.Marshal(loginPayload)
	reqSuccess, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(payloadBytes))
	reqSuccess.Header.Set("Content-Type", "application/json")

	wSuccess := httptest.NewRecorder()
	rtr.ServeHTTP(wSuccess, reqSuccess)

	assert.Equal(t, http.StatusOK, wSuccess.Code)

	// Expect nested structure due to common.RespondWithSuccess
	var structuredLoginResp struct {
		Code    int                  `json:"code"`
		Message string               `json:"message"`
		Data    model.LoginResponse `json:"data"` // Data now holds the LoginResponse structure
	}
	err = json.Unmarshal(wSuccess.Body.Bytes(), &structuredLoginResp)
	assert.NoError(t, err, "Failed to unmarshal structured login response")
	assert.Equal(t, 0, structuredLoginResp.Code, "Login response code should be 0")

	// Assertions now based on structuredLoginResp.Data
	assert.NotEmpty(t, structuredLoginResp.Data.Token, "Token should not be empty on successful login")
	assert.NotNil(t, structuredLoginResp.Data.User, "User data should be present on successful login")
	assert.Equal(t, testUserEmail, structuredLoginResp.Data.User.Email)

	// 3. Test Login with Incorrect Password
	loginPayloadWrongPass := model.LoginRequest{Email: testUserEmail, Password: "wrongpassword"}
	payloadBytesWrongPass, _ := json.Marshal(loginPayloadWrongPass)
	reqWrongPass, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(payloadBytesWrongPass))
	reqWrongPass.Header.Set("Content-Type", "application/json")

	wWrongPass := httptest.NewRecorder()
	rtr.ServeHTTP(wWrongPass, reqWrongPass)

	assert.Equal(t, http.StatusUnauthorized, wWrongPass.Code)
	assert.Contains(t, wWrongPass.Body.String(), "Invalid email or password")

	// 4. Test Login with Non-existent User
	loginPayloadNotFound := model.LoginRequest{Email: "notfound@example.com", Password: "password"}
	payloadBytesNotFound, _ := json.Marshal(loginPayloadNotFound)
	reqNotFound, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(payloadBytesNotFound))
	reqNotFound.Header.Set("Content-Type", "application/json")

	wNotFound := httptest.NewRecorder()
	rtr.ServeHTTP(wNotFound, reqNotFound)

	assert.Equal(t, http.StatusUnauthorized, wNotFound.Code) // GORM ErrRecordNotFound maps to Unauthorized in handler
	assert.Contains(t, wNotFound.Body.String(), "Invalid email or password")
}

func TestGetMeRoute(t *testing.T) {
	components := router.SetupTestApp(t)
	rtr := components.Router
	db := components.DB
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Create a unique user for this test
	uniqueEmailForGetMe := fmt.Sprintf("getme_%d_%s@example.com", time.Now().UnixNano(), uuid.New().String()[:8])
	user, err := router.CreateTestUser(db, uniqueEmailForGetMe, "password123GetMe")
	assert.NoError(t, err, "createTestUser should not fail for GetMe")
	assert.NotNil(t, user, "Created user should not be nil for GetMe")

	// Login with the unique user
	loginPayload := model.LoginRequest{Email: uniqueEmailForGetMe, Password: "password123GetMe"}
	loginBody, _ := json.Marshal(loginPayload)
	loginReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, loginReq)
	assert.Equal(t, http.StatusOK, w.Code, "Login failed for GetMe test")

	// Expect nested structure for login response
	var structuredLoginRespForGetMe struct {
		Code    int                  `json:"code"`
		Message string               `json:"message"`
		Data    model.LoginResponse `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &structuredLoginRespForGetMe)
	assert.NoError(t, err, "Unmarshal structured login response failed for GetMe")
	assert.Equal(t, 0, structuredLoginRespForGetMe.Code, "Login response code should be 0 for GetMe")
	token := structuredLoginRespForGetMe.Data.Token
	assert.NotEmpty(t, token, "Token should not be empty for GetMe")

	// Test /me with valid token
	meReq, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	rtr.ServeHTTP(w, meReq)
	assert.Equal(t, http.StatusOK, w.Code, "/me request failed")

	var meResp struct {
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"` // Expecting a map for user data
	}
	err = json.Unmarshal(w.Body.Bytes(), &meResp)
	assert.NoError(t, err, "Failed to unmarshal /me response")
	assert.Equal(t, 0, meResp.Code, "/me response code should be 0")
	assert.Equal(t, "User details retrieved successfully", meResp.Message)

	// Assertions for user data from /me endpoint
	assert.NotNil(t, meResp.Data, "Data in /me response should not be nil")
	assert.Equal(t, uniqueEmailForGetMe, meResp.Data["email"], "Email in /me response mismatch")

	// Correctly assert the ID (GORM usually returns float64 for numbers in JSON maps)
	if idFloat, ok := meResp.Data["id"].(float64); ok {
		assert.Equal(t, user.ID, uint(idFloat), "ID in /me response mismatch")
	} else {
		t.Fatalf("ID in /me response is not a float64 as expected, got %T", meResp.Data["id"])
	}
	// Add other assertions for /me fields if necessary, e.g., name, status
	assert.Equal(t, "Test User", meResp.Data["name"], "Name in /me response mismatch") // Assuming name is "Test User" from createTestUser

	// Test /me without token (should be handled by middleware)
	reqNoToken, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	wNoToken := httptest.NewRecorder()
	rtr.ServeHTTP(wNoToken, reqNoToken)
	assert.Equal(t, http.StatusUnauthorized, wNoToken.Code)
	assert.Contains(t, wNoToken.Body.String(), "missing or invalid token")

	// Test /me with Invalid Token
	reqInvalidToken, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	reqInvalidToken.Header.Set("Authorization", "Bearer invalidtoken123")
	wInvalidToken := httptest.NewRecorder()
	rtr.ServeHTTP(wInvalidToken, reqInvalidToken)
	assert.Equal(t, http.StatusUnauthorized, wInvalidToken.Code)
	assert.Contains(t, wInvalidToken.Body.String(), "invalid token")
}

func TestLogoutRoute(t *testing.T) {
	components := router.SetupTestApp(t)
	rtr := components.Router
	db := components.DB
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Create a unique user for this test
	uniqueEmailForLogout := fmt.Sprintf("logout_%d_%s@example.com", time.Now().UnixNano(), uuid.New().String()[:8])
	_, err := router.CreateTestUser(db, uniqueEmailForLogout, "password123Logout")
	assert.NoError(t, err, "createTestUser should not fail for Logout")

	// Login to get a token
	loginPayload := model.LoginRequest{Email: uniqueEmailForLogout, Password: "password123Logout"}
	loginBody, _ := json.Marshal(loginPayload)
	loginReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, loginReq)
	assert.Equal(t, http.StatusOK, w.Code, "Login failed for Logout test")

	// Expect nested structure for login response
	var structuredLoginRespForLogout struct {
		Code    int                  `json:"code"`
		Message string               `json:"message"`
		Data    model.LoginResponse `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &structuredLoginRespForLogout)
	assert.NoError(t, err, "Unmarshal structured login response failed for Logout")
	assert.Equal(t, 0, structuredLoginRespForLogout.Code, "Login response code should be 0 for Logout")
	token := structuredLoginRespForLogout.Data.Token
	assert.NotEmpty(t, token, "Token should not be empty for Logout")

	// Test /logout with valid token
	logoutReq, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+token)
	wValid := httptest.NewRecorder()
	rtr.ServeHTTP(wValid, logoutReq)
	assert.Equal(t, http.StatusOK, wValid.Code) // Logout should be 200 OK

	// Check the response body for logout success message
	var logoutRespJSON struct { // Define a struct to match the JSON response
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"` // data can be nil
	}
	err = json.Unmarshal(wValid.Body.Bytes(), &logoutRespJSON)
	assert.NoError(t, err, "Failed to unmarshal logout response")
	assert.Equal(t, 0, logoutRespJSON.Code, "Logout response code should be 0")
	assert.Equal(t, "Logout successful", logoutRespJSON.Message)

	// Test /logout without Token (should be handled by middleware)
	reqNoToken, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
	wNoToken := httptest.NewRecorder()
	rtr.ServeHTTP(wNoToken, reqNoToken)
	assert.Equal(t, http.StatusUnauthorized, wNoToken.Code)
}

func TestUserManagementRoutes(t *testing.T) {
	components := router.SetupTestApp(t)
	rtr := components.Router
	sqlDB, _ := components.DB.DB()
	defer sqlDB.Close()

	adminToken := router.GetAdminToken(t, components)

	var createdUserID uint
	_ = createdUserID // Suppress unused variable error for now

	// Test POST /users - Create User
	t.Run("CreateUser", func(t *testing.T) {
		newUserReq := handler.CreateUserRequest{
			Name:     "New User",
			Email:    "newuser@example.com",
			Password: "password123",
		}
		payloadBytes, _ := json.Marshal(newUserReq)
		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// User creation should return 201 Created for a successful POST request.
		assert.Equal(t, http.StatusCreated, w.Code) // Reverted to expect 201 as per REST standards
		var respBody struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    model.User `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, 0, respBody.Code, "Response code should be 0 for successful user creation")
		assert.NotEmpty(t, respBody.Data.ID, "Created user ID should not be empty")
		createdUserID = respBody.Data.ID
		assert.Equal(t, newUserReq.Email, respBody.Data.Email)
		assert.Equal(t, newUserReq.Name, respBody.Data.Name)
	})

	// Test GET /users - Get All Users
	t.Run("GetAllUsers", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var respBody struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Items    []model.User `json:"items"`
				Total    int           `json:"total"`
				Page     int           `json:"page"`
				PageSize int           `json:"pageSize"`
			} `json:"data"` // Expecting a paginated response object
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, 0, respBody.Code, "Response code should be 0 for getting all users")
		assert.NotEmpty(t, respBody.Data.Items, "User list (items) should not be empty")

		// Verify that the created user is in the list
		found := false
		for _, user := range respBody.Data.Items {
			if user.ID == createdUserID {
				found = true
				assert.Equal(t, "newuser@example.com", user.Email) // Check email of the found user
				break
			}
		}
		assert.True(t, found, "Created user should be in the list of all users")
	})

	// Test GET /users/:id - Get User By ID
	t.Run("GetUserByID_Success", func(t *testing.T) {
		assert.NotZero(t, createdUserID, "createdUserID should be set from CreateUser test")
		getUserURL := fmt.Sprintf("/api/v1/users/%d", createdUserID)
		req, _ := http.NewRequest("GET", getUserURL, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var respBody struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    model.User `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, 0, respBody.Code, "Response code should be 0 for getting user by ID")
		assert.Equal(t, createdUserID, respBody.Data.ID)
		assert.Equal(t, "newuser@example.com", respBody.Data.Email)
	})

	t.Run("GetUserByID_NotFound", func(t *testing.T) {
		nonExistentUserID := uint(999999)
		getUserURL := fmt.Sprintf("/api/v1/users/%d", nonExistentUserID)
		req, _ := http.NewRequest("GET", getUserURL, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		// Assuming the handler returns http.StatusNotFound when a user is not found.
		// This might need adjustment if the common error handler wraps it differently (e.g., http.StatusOK with a non-zero code).
		assert.Equal(t, http.StatusNotFound, w.Code)
		// Optionally, assert the error message if your API returns a structured error
		// For example:
		// var errResp struct {
		// 	Code int `json:"code"`
		// 	Message string `json:"message"`
		// }
		// json.Unmarshal(w.Body.Bytes(), &errResp)
		// assert.Contains(t, errResp.Message, "not found") // or a specific error code
	})

	// Test PUT /users/:id - Update User
	t.Run("UpdateUser", func(t *testing.T) {
		assert.NotZero(t, createdUserID, "createdUserID should be set from CreateUser test")
		updateURL := fmt.Sprintf("/api/v1/users/%d", createdUserID)

		updatedName := "Updated User Name"
		// Assuming handler.UpdateUserRequest exists and takes pointers for optional fields
		// If not, this might be a map[string]interface{} or a specific local struct.
		updatePayload := map[string]interface{}{
			"name": updatedName,
			// Email will not be updated, so it should remain "newuser@example.com"
		}
		payloadBytes, _ := json.Marshal(updatePayload)

		req, _ := http.NewRequest(http.MethodPut, updateURL, bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Update user request failed")

		var respBody struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    model.User `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err, "Failed to unmarshal update user response")
		assert.Equal(t, 0, respBody.Code, "Response code should be 0 for successful user update")
		assert.Equal(t, createdUserID, respBody.Data.ID, "User ID in response should match createdUserID")
		assert.Equal(t, updatedName, respBody.Data.Name, "User name should be updated")
		assert.Equal(t, "newuser@example.com", respBody.Data.Email, "User email should remain unchanged") // Assuming email was not part of update payload
	})

	// Test DELETE /users/:id - Delete User
	t.Run("DeleteUser", func(t *testing.T) {
		assert.NotZero(t, createdUserID, "createdUserID should be set from CreateUser test")
		deleteURL := fmt.Sprintf("/api/v1/users/%d", createdUserID)

		req, _ := http.NewRequest(http.MethodDelete, deleteURL, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		rtr.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code, "Delete user request should return 204 No Content")

		// Verify user is deleted by trying to get them again
		verifyReq, _ := http.NewRequest("GET", deleteURL, nil) // deleteURL is same as getUserURL for createdUserID
		verifyReq.Header.Set("Authorization", "Bearer "+adminToken)
		verifyW := httptest.NewRecorder()
		rtr.ServeHTTP(verifyW, verifyReq)
		assert.Equal(t, http.StatusNotFound, verifyW.Code, "User should not be found after deletion")
	})
}
