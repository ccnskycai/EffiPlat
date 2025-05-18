package router_test

import (
	"EffiPlat/backend/internal/factories"
	"EffiPlat/backend/internal/handler"
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/pkg/config"
	pkgdb "EffiPlat/backend/internal/pkg/database"
	"EffiPlat/backend/internal/pkg/logger"
	"EffiPlat/backend/internal/repository"
	"EffiPlat/backend/internal/router"
	"EffiPlat/backend/internal/service"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// setupAppTestRouter initializes a new router with all dependencies for user-role tests.
// It returns the router, database connection, logger, and handlers.
func setupAppTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, *zap.Logger, *handler.AuthHandler, *handler.UserHandler, *handler.RoleHandler, *handler.PermissionHandler) {
	gin.SetMode(gin.TestMode)

	// Use the actual config types from the config package
	cfg := config.AppConfig{ // Changed from config.Config to config.AppConfig
		Database: config.DBConfig{DSN: "file::memory:?cache=shared", Type: "sqlite"}, // Changed to config.DBConfig, assuming Type field exists or corresponds to Driver
		Logger:   logger.Config{Level: "error", Encoding: "console"},                 // Changed Format to Encoding
		Server:   config.ServerConfig{Port: 8088},
	}

	appLogger, err := logger.NewLogger(cfg.Logger) // Pass cfg.Logger (which is logger.Config)
	assert.NoError(t, err)

	db, err := pkgdb.NewConnection(cfg.Database, appLogger)
	assert.NoError(t, err)
	err = pkgdb.AutoMigrate(db, appLogger) // Uncommented and handle error
	assert.NoError(t, err, "AutoMigrate should not fail")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db, appLogger)
	roleRepo := repository.NewRoleRepository(db, appLogger)
	permRepo := repository.NewPermissionRepository(db, appLogger) // Needed for permission handler

	// Initialize services
	jwtKey := []byte(os.Getenv("JWT_SECRET_TEST"))
	if len(jwtKey) == 0 {
		jwtKey = []byte("test_secret_key_for_router_tests")
	}
	authService := service.NewAuthService(userRepo, jwtKey, appLogger)
	userService := service.NewUserService(userRepo)
	roleService := service.NewRoleService(roleRepo, appLogger)
	permissionService := service.NewPermissionService(permRepo, roleRepo, appLogger)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	roleHandler := handler.NewRoleHandler(roleService, appLogger)
	permissionHandler := handler.NewPermissionHandler(permissionService, appLogger)

	r := router.SetupRouter(authHandler, userHandler, roleHandler, permissionHandler, jwtKey)
	return r, db, appLogger, authHandler, userHandler, roleHandler, permissionHandler
}

// TestUserRoleManagementRoutes covers assigning and removing roles from users.
func TestUserRoleManagementRoutes(t *testing.T) {
	r, db, _, _, _, _, _ := setupAppTestRouter(t)

	// Helper to create a user and get a token (simulating admin login for now)
	createAdminAndLogin := func() (adminUser *models.User, token string) {
		adminEmail := "admin_ur@example.com"
		adminPassword := "SecurePassword123"
		adminUser, _ = factories.CreateUser(db, &models.User{
			Name:     "Admin User UR",
			Email:    adminEmail,
			Password: adminPassword, // Will be hashed by factory/service
			Status:   "active",
		})
		// For simplicity in test, we assume CreateUser also assigns an admin role or tests don't strictly check role
		// or we'd need to create an admin role and assign it first.
		// Here, we just log in as this user.

		loginReqBody := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, adminEmail, adminPassword)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(loginReqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var loginResp struct {
			Token string `json:"token"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &loginResp)
		assert.NoError(t, err)
		assert.NotEmpty(t, loginResp.Token)
		return adminUser, loginResp.Token
	}

	_, adminToken := createAdminAndLogin()

	// --- Test Case: Successfully Assign Roles to User ---
	t.Run("Assign_Roles_To_User_Success", func(t *testing.T) {
		// 1. Create a target user
		targetUser, _ := factories.CreateUser(db, &models.User{
			Name:  "Target User Assign",
			Email: "target_assign_ur@example.com",
		})

		// 2. Create some roles
		role1, _ := factories.CreateRole(db, &models.Role{Name: "Role UR A"})
		role2, _ := factories.CreateRole(db, &models.Role{Name: "Role UR B"})

		// 3. Prepare request to assign roles
		assignReqPayload := models.AssignRemoveRolesRequest{
			RoleIDs: []uint{role1.ID, role2.ID},
		}
		payloadBytes, _ := json.Marshal(assignReqPayload)

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// 4. Assertions
		assert.Equal(t, http.StatusOK, w.Code) // Or http.StatusNoContent if API returns that

		var respBody struct {
			BizCode int         `json:"bizCode"`
			Message string      `json:"message"`
			Data    interface{} `json:"data"` // Data might be null or contain some info
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, 0, respBody.BizCode)
		assert.Equal(t, "Roles assigned successfully to user", respBody.Message)

		// 5. (Optional) Verify roles were actually assigned by fetching the user
		var updatedUser models.User
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
		// 1. Create a role (doesn't matter which, just need some valid role IDs)
		roleToAssign, _ := factories.CreateRole(db, &models.Role{Name: "Role UR C"})

		// 2. Prepare request with a non-existent user ID
		nonExistentUserID := uint(99999)
		assignReqPayload := models.AssignRemoveRolesRequest{
			RoleIDs: []uint{roleToAssign.ID},
		}
		payloadBytes, _ := json.Marshal(assignReqPayload)

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%d/roles", nonExistentUserID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// 3. Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)

		var respBody struct {
			// Assuming ErrorResponse structure from handler/common.go
			Code    int    `json:"code"` // HTTP status code, might be different from BizCode
			Message string `json:"message"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		// The 'Code' in ErrorResponse is the HTTP status code itself.
		assert.Equal(t, http.StatusNotFound, respBody.Code)
		assert.Contains(t, respBody.Message, "user not found") // Or the exact error message from service.ErrUserNotFound
	})

	// --- Test Case: Assign Roles to User - Invalid Role IDs ---
	t.Run("Assign_Roles_To_User_Invalid_Role_IDs", func(t *testing.T) {
		// 1. Create a target user
		targetUser, _ := factories.CreateUser(db, &models.User{
			Name:  "Target User Invalid Roles",
			Email: "target_invalid_roles_ur@example.com",
		})

		// 2. Create one valid role
		validRole, _ := factories.CreateRole(db, &models.Role{Name: "Role UR Valid ForInvalidTest"}) // Renamed for clarity
		nonExistentRoleID := uint(99998)

		// 3. Prepare request to assign roles with one valid and one invalid ID
		assignReqPayload := models.AssignRemoveRolesRequest{
			RoleIDs: []uint{validRole.ID, nonExistentRoleID},
		}
		payloadBytes, _ := json.Marshal(assignReqPayload)

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// 4. Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var respBody models.ErrorResponse // Use the standard error response model
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, respBody.Code)
		assert.Contains(t, respBody.Message, "one or more roles not found", "Error message should indicate role not found for assign")

		// 5. Verify that no roles were assigned if an invalid ID was provided
		var updatedUser models.User
		err = db.Preload("Roles").First(&updatedUser, targetUser.ID).Error
		assert.NoError(t, err, "User record should still be found")
		assert.Empty(t, updatedUser.Roles, "No roles should be assigned if transaction is rolled back due to invalid role ID")
	})

	// --- Test Case: Assign Roles to User - Empty Role IDs ---
	t.Run("Assign_Roles_To_User_Empty_Role_IDs", func(t *testing.T) {
		// 1. Create a target user
		targetUser, _ := factories.CreateUser(db, &models.User{
			Name:  "Target User Empty Roles",
			Email: "target_empty_roles_ur@example.com",
		})

		// 2. Prepare request with empty RoleIDs
		assignReqPayload := models.AssignRemoveRolesRequest{
			RoleIDs: []uint{},
		}
		payloadBytes, _ := json.Marshal(assignReqPayload)

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// 3. Assertions - Assuming API treats this as a no-op success
		assert.Equal(t, http.StatusOK, w.Code)

		var respBody struct {
			BizCode int    `json:"bizCode"`
			Message string `json:"message"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, 0, respBody.BizCode) // Success
		// Message might vary, e.g., "No roles specified for assignment" or "Roles assigned successfully" (even if none were)
		// Let's assume a generic success or a specific one if known.
		// For now, we can check it's not empty or contains "successfully"
		assert.Contains(t, respBody.Message, "successfully") // Or a more specific message

		// 4. Verify that no roles were assigned (user had no roles initially)
		var updatedUser models.User
		err = db.Preload("Roles").First(&updatedUser, targetUser.ID).Error
		assert.NoError(t, err, "User record should still be found")
		assert.Empty(t, updatedUser.Roles, "No roles should be assigned when an empty list is provided")
	})

	// --- Test Case: Assign Roles to User - Unauthorized ---
	t.Run("Assign_Roles_To_User_Unauthorized", func(t *testing.T) {
		// 1. Create a target user (though not strictly necessary as auth should fail first)
		targetUser, _ := factories.CreateUser(db, &models.User{
			Name:  "Target User Unauthorized Assign",
			Email: "target_unauth_assign_ur@example.com",
		})

		// 2. Prepare request (valid payload, but no token)
		assignReqPayload := models.AssignRemoveRolesRequest{
			RoleIDs: []uint{1}, // A dummy role ID, content doesn't matter for auth check
		}
		payloadBytes, _ := json.Marshal(assignReqPayload)

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header is set

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// 3. Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		// Instead of checking for code and message, check for the JWT middleware's error structure
		var errResp struct {
			Error string `json:"error"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err, "Failed to unmarshal unauthorized error response for assign roles")

		assert.Equal(t, http.StatusUnauthorized, w.Code) // HTTP status code is the primary check
		assert.Contains(t, errResp.Error, "missing or invalid token", "Error message should indicate a token issue for assign roles")
	})

	// --- Test Case: Assign Roles to User - Bad Request (Malformed JSON) ---
	t.Run("Assign_Roles_To_User_Bad_Request", func(t *testing.T) {
		// 1. Create a target user (ID is needed for the URL, user existence doesn't matter as parsing should fail first)
		targetUser, _ := factories.CreateUser(db, &models.User{
			Name:  "Target User Bad Req Assign",
			Email: "target_badreq_assign_ur@example.com",
		})

		// 2. Prepare a malformed JSON payload
		malformedPayload := "{\"role_ids\": [1, 2]" // Missing closing brace

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBufferString(malformedPayload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken) // Auth token is present

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// 3. Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var respBody struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err) // Gin usually returns a parseable JSON error
		assert.Equal(t, http.StatusBadRequest, respBody.Code)
		// The exact message might depend on Gin's JSON parsing or your custom error handling
		// Common messages include "invalid character", "unexpected end of JSON input", etc.
		// For robustness, we can check if the message is non-empty or contains a known substring for bind errors.
		assert.NotEmpty(t, respBody.Message, "Error message should indicate a binding/parsing issue")
	})

	// --- Test Case: Successfully Remove Roles from User ---
	t.Run("Remove_Roles_From_User_Success", func(t *testing.T) {
		// 1. Create a target user
		targetUser, _ := factories.CreateUser(db, &models.User{
			Name:  "Target User Remove Roles",
			Email: "target_remove_ur@example.com",
		})

		// 2. Create some roles
		roleToKeep, _ := factories.CreateRole(db, &models.Role{Name: "Role UR Keep"})
		roleToRemove1, _ := factories.CreateRole(db, &models.Role{Name: "Role UR Remove1"})
		roleToRemove2, _ := factories.CreateRole(db, &models.Role{Name: "Role UR Remove2"})

		// 3. Assign roles directly to user in DB for setup
		db.Model(&targetUser).Association("Roles").Append([]*models.Role{roleToKeep, roleToRemove1, roleToRemove2})

		// Verify initial state
		var initialUser models.User
		db.Preload("Roles").First(&initialUser, targetUser.ID)
		assert.Len(t, initialUser.Roles, 3, "User should have 3 roles initially")

		// 4. Prepare request to remove two roles
		removeReqPayload := models.AssignRemoveRolesRequest{ // Using the same DTO
			RoleIDs: []uint{roleToRemove1.ID, roleToRemove2.ID},
		}
		payloadBytes, _ := json.Marshal(removeReqPayload)

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// 5. Assertions
		assert.Equal(t, http.StatusOK, w.Code) // Or http.StatusNoContent, depending on API design

		var respBody struct {
			BizCode int    `json:"bizCode"`
			Message string `json:"message"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, 0, respBody.BizCode)
		assert.Equal(t, "Roles removed successfully from user", respBody.Message)

		// 6. Verify roles were actually removed and one was kept
		var updatedUser models.User
		err = db.Preload("Roles").First(&updatedUser, targetUser.ID).Error
		assert.NoError(t, err)
		assert.Len(t, updatedUser.Roles, 1)
		assert.Equal(t, roleToKeep.ID, updatedUser.Roles[0].ID)
		assert.Equal(t, roleToKeep.Name, updatedUser.Roles[0].Name)
	})

	// --- Test Case: Remove Roles from User - User Not Found ---
	t.Run("Remove_Roles_From_User_User_Not_Found", func(t *testing.T) {
		// 1. Create a role (doesn't matter which, just need some valid role IDs for the payload)
		roleToRemove, _ := factories.CreateRole(db, &models.Role{Name: "Role UR ToRemove D"})

		// 2. Prepare request with a non-existent user ID
		nonExistentUserID := uint(99997) // Ensure this ID is different from other tests
		removeReqPayload := models.AssignRemoveRolesRequest{
			RoleIDs: []uint{roleToRemove.ID},
		}
		payloadBytes, _ := json.Marshal(removeReqPayload)

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%d/roles", nonExistentUserID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// 3. Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)

		var respBody struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, respBody.Code)
		assert.Contains(t, respBody.Message, "user not found")
	})

	// --- Test Case: Remove Roles from User - Role Not Assigned to User ---
	t.Run("Remove_Roles_From_User_Role_Not_Assigned", func(t *testing.T) {
		// 1. Create a target user
		targetUser, _ := factories.CreateUser(db, &models.User{
			Name:  "Target User Role Not Assigned",
			Email: "target_rnas_ur@example.com",
		})

		// 2. Create a role that will be assigned, and one that won't (but we'll try to remove it)
		assignedRole, _ := factories.CreateRole(db, &models.Role{Name: "Role UR Assigned"})
		unassignedRole, _ := factories.CreateRole(db, &models.Role{Name: "Role UR Unassigned ButTryRemove"})

		// 3. Assign one role directly to user in DB for setup
		db.Model(&targetUser).Association("Roles").Append(assignedRole)

		// Verify initial state
		var initialUser models.User
		db.Preload("Roles").First(&initialUser, targetUser.ID)
		assert.Len(t, initialUser.Roles, 1, "User should have 1 role initially")
		assert.Equal(t, assignedRole.ID, initialUser.Roles[0].ID)

		// 4. Prepare request to remove the unassignedRole and the assignedRole
		removeReqPayload := models.AssignRemoveRolesRequest{
			RoleIDs: []uint{unassignedRole.ID, assignedRole.ID},
		}
		payloadBytes, _ := json.Marshal(removeReqPayload)

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// 5. Assertions - Should be success, as the system ensures the roles are not associated
		assert.Equal(t, http.StatusOK, w.Code)

		var respBody struct {
			BizCode int    `json:"bizCode"`
			Message string `json:"message"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, 0, respBody.BizCode)
		assert.Equal(t, "Roles removed successfully from user", respBody.Message)

		// 6. Verify the assignedRole was removed and user has no roles now
		var updatedUser models.User
		err = db.Preload("Roles").First(&updatedUser, targetUser.ID).Error
		assert.NoError(t, err)
		assert.Empty(t, updatedUser.Roles, "User should have no roles after removal")
	})

	// --- Test Case: Remove Roles from User - Invalid Role IDs ---
	t.Run("Remove_Roles_From_User_Invalid_Role_IDs", func(t *testing.T) {
		// 1. Create a target user
		targetUser, _ := factories.CreateUser(db, &models.User{
			Name:  "Target User Remove Invalid Roles",
			Email: "target_remove_invalid_ur@example.com",
		})

		// 2. Create a role that will be assigned
		assignedRole, _ := factories.CreateRole(db, &models.Role{Name: "Role UR Assigned ForInvalidRemove"})
		db.Model(&targetUser).Association("Roles").Append(assignedRole)

		// Verify initial state
		var initialUser models.User
		db.Preload("Roles").First(&initialUser, targetUser.ID)
		assert.Len(t, initialUser.Roles, 1, "User should have 1 role initially")

		// 3. Prepare request with one valid (but perhaps not assigned to this user) and one non-existent Role ID
		validRoleNotAssigned, _ := factories.CreateRole(db, &models.Role{Name: "Role UR ValidButNotAssigned"})
		nonExistentRoleID := uint(99996) // Ensure this ID is different

		removeReqPayload := models.AssignRemoveRolesRequest{
			RoleIDs: []uint{validRoleNotAssigned.ID, nonExistentRoleID},
		}
		payloadBytes, _ := json.Marshal(removeReqPayload)

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// 4. Assertions - Expecting a 400 Bad Request due to non-existent role ID
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var respBody models.ErrorResponse // Use the standard error response model
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, respBody.Code)
		assert.Contains(t, respBody.Message, "one or more roles not found", "Error message should indicate role not found for remove")

		// 5. Verify that the user's original roles are untouched
		var updatedUser models.User
		err = db.Preload("Roles").First(&updatedUser, targetUser.ID).Error
		assert.NoError(t, err)
		assert.Len(t, updatedUser.Roles, 1, "User's roles should be unchanged")
		assert.Equal(t, assignedRole.ID, updatedUser.Roles[0].ID)
	})

	// --- Test Case: Remove Roles from User - Empty Role IDs ---
	t.Run("Remove_Roles_From_User_Empty_Role_IDs", func(t *testing.T) {
		// 1. Create a user and assign a role
		testUserMail := "emptyrolesuser@example.com"
		createdUser, err := createTestUser(db, testUserMail, "password")
		assert.NoError(t, err)

		viewerRole := models.Role{Name: "Viewer Role for Empty Test", Description: "Viewer desc"}
		db.Create(&viewerRole)
		db.Model(&createdUser).Association("Roles").Append(&viewerRole)

		// 2. Attempt to remove an empty list of roles
		payload := models.AssignRemoveRolesRequest{RoleIDs: []uint{}}
		payloadBytes, _ := json.Marshal(payload)
		req, _ := http.NewRequest("DELETE", "/api/v1/users/"+strconv.FormatUint(uint64(createdUser.ID), 10)+"/roles", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken) // Assuming adminToken is available from test setup

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req) // rtr should be the router instance

		assert.Equal(t, http.StatusOK, w.Code) // Expect 200 OK for successful no-op

		var respBody models.SuccessResponse                   // Expecting a standard success response
		parseErr := json.Unmarshal(w.Body.Bytes(), &respBody) // Assign to new var to avoid shadow
		assert.NoError(t, parseErr)
		assert.Equal(t, 0, respBody.BizCode) // BizCode 0 for success
		assert.Contains(t, respBody.Message, "Roles removed successfully from user", "Message should indicate success for empty list")

		// Verify user's roles are unchanged (should still be the initial role)
		var unchangedUser models.User
		db.Preload("Roles").First(&unchangedUser, createdUser.ID)
		assert.Len(t, unchangedUser.Roles, 1, "User should still have their initial role after attempting to remove an empty list")
		if len(unchangedUser.Roles) > 0 {
			assert.Equal(t, viewerRole.Name, unchangedUser.Roles[0].Name)
		}
	})

	// --- Test Case: Remove Roles from User - Unauthorized ---
	t.Run("Remove_Roles_From_User_Unauthorized", func(t *testing.T) {
		// 1. Create a target user (ID needed for URL, actual user state doesn't matter for auth check)
		targetUser, _ := factories.CreateUser(db, &models.User{
			Name:  "Target User Unauth Remove",
			Email: "target_unauth_remove_ur@example.com",
		})

		// 2. Prepare request (valid payload, but no token)
		removeReqPayload := models.AssignRemoveRolesRequest{
			RoleIDs: []uint{1}, // Dummy role ID
		}
		payloadBytes, _ := json.Marshal(removeReqPayload)

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// 3. Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		// Instead of checking for code and message, check for the JWT middleware's error structure
		var errResp struct {
			Error string `json:"error"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err, "Failed to unmarshal unauthorized error response for remove roles")

		assert.Equal(t, http.StatusUnauthorized, w.Code) // HTTP status code is the primary check
		assert.Contains(t, errResp.Error, "missing or invalid token", "Error message should indicate a token issue for remove roles")
	})

	// --- Test Case: Remove Roles from User - Bad Request (Malformed JSON) ---
	t.Run("Remove_Roles_From_User_Bad_Request", func(t *testing.T) {
		// 1. Create a target user (ID is needed for the URL)
		targetUser, _ := factories.CreateUser(db, &models.User{
			Name:  "Target User Bad Req Remove",
			Email: "target_badreq_remove_ur@example.com",
		})

		// 2. Prepare a malformed JSON payload
		malformedPayload := "{\"role_ids\": [1" // Missing closing bracket and brace

		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%d/roles", targetUser.ID), bytes.NewBufferString(malformedPayload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken) // Valid token

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// 3. Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var respBody struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err) // Expecting a parseable JSON error from Gin
		assert.Equal(t, http.StatusBadRequest, respBody.Code)
		assert.NotEmpty(t, respBody.Message, "Error message should indicate a binding/parsing issue")
	})
}
