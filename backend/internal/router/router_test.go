package router_test // Use _test package suffix for integration-style tests

import (
	// Go Standard Library
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	// External Dependencies
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	// "go.uber.org/zap" // REMOVED: Not used
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger" // Import gorm logger with alias

	// Internal Packages
	"EffiPlat/backend/internal/handler"
	"EffiPlat/backend/internal/models" // May need config for defaults if used by handlers/services

	// Import for DB logger integration if needed, though using simpler logger here
	// "EffiPlat/backend/internal/pkg/logger" // REMOVED: Not used
	"EffiPlat/backend/internal/repository"
	"EffiPlat/backend/internal/router" // Package being tested
	"EffiPlat/backend/internal/service"
)

// Constants for testing
const (
	testUserEmail    = "testuser1@example.com"
	testUserPassword = "password"
	testJWTSecret    = "test_secret_key_for_router_tests" // Use a fixed secret for tests
)

// Helper to setup router with real dependencies connected to an in-memory DB
func setupTestRouter() (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode) // Set Gin to test mode

	// Setup in-memory SQLite for testing
	// Using "file::memory:?cache=shared" allows multiple connections in the same test process if needed,
	// but requires careful handling or ensuring single connection use per test run.
	// Simpler "file:test.db?mode=memory&cache=shared" might also work.
	// Using a unique name per test run avoids conflicts if tests run concurrently (less likely in Go by default).
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		// Use a simpler logger for tests to avoid noise, or configure properly
		Logger: gormlogger.Default.LogMode(gormlogger.Silent), // Use gorm's logger
	})
	if err != nil {
		log.Fatalf("Failed to connect test database: %v", err)
	}

	// Run Migrations for the test database
	// This assumes migrate CLI is available or uses GORM's AutoMigrate
	// Using AutoMigrate for simplicity here. Ensure all required models are included.
	err = db.AutoMigrate(
		&models.User{},
		// Add other models needed by the tested routes if any
		// &models.Role{},
		// &models.Permission{},
		// ...
	)
	if err != nil {
		log.Fatalf("Failed to migrate test database: %v", err)
	}

	// --- Initialize Dependencies ---
	// Normally use a test logger, using Nop for simplicity now
	// testLogger := zap.NewNop() // REMOVED: Unused variable

	// JWT Key
	jwtKey := []byte(testJWTSecret)

	// Create real instances pointing to the test DB
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, jwtKey)
	authHandler := handler.NewAuthHandler(authService)
	// Initialize other handlers if needed for tested routes

	// Setup router with test dependencies
	r := router.SetupRouter(authHandler, jwtKey /*, other handlers... */)
	return r, db // Return DB for test data setup/teardown
}

// Helper to create a test user directly in the DB
func createTestUser(db *gorm.DB, email, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &models.User{
		Name:         "Test User",
		Email:        email,
		PasswordHash: string(hashedPassword),
		Status:       "active",
	}
	result := db.Create(user)
	return user, result.Error
}

// --- Test Cases ---

func TestHealthRoute(t *testing.T) {
	router, db := setupTestRouter()
	sqlDB, _ := db.DB()
	defer sqlDB.Close() // Ensure DB connection is closed after test

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status": "UP"}`, w.Body.String())
}

func TestLoginRoute(t *testing.T) {
	router, db := setupTestRouter()
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
	router, db := setupTestRouter()
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
	router, db := setupTestRouter()
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

// TODO: Add tests for other routes as they are implemented
