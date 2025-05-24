package router

import (
	"EffiPlat/backend/internal/handler"
	"EffiPlat/backend/internal/model"
	pkgmodel "EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/pkg/config"
	pkgdb "EffiPlat/backend/internal/pkg/database"
	"EffiPlat/backend/internal/pkg/logger"
	"EffiPlat/backend/internal/repository"
	"EffiPlat/backend/internal/service"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TestAppComponents holds all initialized components for testing.
// This makes it easier to pass around and reduces the number of return values from setup.
// Responsibility and ResponsibilityGroup components are added.
// Renamed some existing handlers for clarity (e.g., UserHdlr -> UserHandler)
type TestAppComponents struct {
	Router                     *gin.Engine
	DB                         *gorm.DB
	Logger                     *zap.Logger
	AuthHandler                *handler.AuthHandler
	UserHandler                *handler.UserHandler
	RoleHandler                *handler.RoleHandler
	PermissionHandler          *handler.PermissionHandler
	ResponsibilityHandler      *handler.ResponsibilityHandler
	ResponsibilityGroupHandler *handler.ResponsibilityGroupHandler
	EnvironmentHandler         *handler.EnvironmentHandler
	AssetHandler               *handler.AssetHandler
	ServiceHandler             *handler.ServiceHandler
	ServiceInstanceHandler     *handler.ServiceInstanceHandler
	BusinessHandler            *handler.BusinessHandler
	BugHandler                 *handler.BugHandler
	AuditLogHandler            *handler.AuditLogHandler   // 新增审计日志处理器
	AuditLogService            service.AuditLogService   // 新增审计日志服务
	JWTKey                     []byte
}

// setupTestApp initializes a new router with all dependencies for integration tests.
// It now initializes and returns all handlers in the TestAppComponents struct.
func SetupTestApp(t *testing.T) TestAppComponents {
	gin.SetMode(gin.TestMode)

	cfg := config.AppConfig{
		Database: config.DBConfig{DSN: "file::memory:?cache=shared", Type: "sqlite"},
		Logger:   logger.Config{Level: "error", Encoding: "console"},
		Server:   config.ServerConfig{Port: 8088}, // Port for test server if needed
	}

	appLogger, err := logger.NewLogger(cfg.Logger)
	assert.NoError(t, err)

	db, err := pkgdb.NewConnection(cfg.Database, appLogger)
	assert.NoError(t, err)
	// err = pkgdb.AutoMigrate(db, appLogger) // Ensure all tables including new ones are migrated
	// Directly migrate all necessary models for tests, including new ones
	err = db.AutoMigrate(
		&pkgmodel.User{},
		&pkgmodel.Role{},
		&pkgmodel.Permission{},
		&pkgmodel.Responsibility{},
		&pkgmodel.ResponsibilityGroup{},
		&pkgmodel.Environment{},
		&pkgmodel.Asset{},
		&pkgmodel.ServiceType{}, // Added ServiceType model for migration
		&pkgmodel.Service{},     // Added Service model for migration
		&model.ServiceInstance{}, // Changed to model.ServiceInstance
		&model.Business{},        // Changed to model.Business
		&model.AuditLog{},        // Added AuditLog model for migration
	)
	assert.NoError(t, err, "AutoMigrate should not fail")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db, appLogger)
	roleRepo := repository.NewRoleRepository(db, appLogger)
	permRepo := repository.NewPermissionRepository(db, appLogger)
	responsibilityRepo := repository.NewGormResponsibilityRepository(db, appLogger)
	responsibilityGroupRepo := repository.NewGormResponsibilityGroupRepository(db, appLogger)
	environmentRepo := repository.NewGormEnvironmentRepository(db, appLogger)
	assetRepo := repository.NewGormAssetRepository(db, appLogger)
	serviceRepo := repository.NewGormServiceRepository(db)                        // Updated ServiceRepository
	serviceTypeRepo := repository.NewGormServiceTypeRepository(db)                // Added ServiceTypeRepository
	serviceInstanceRepo := repository.NewServiceInstanceRepository(db, appLogger) // Added
	bugRepo := repository.NewBugRepository(db, appLogger) // Added BugRepository
	auditLogRepo := repository.NewAuditLogRepository(db, appLogger) // 审计日志存储库
	businessRepo := repository.NewBusinessRepository(db, appLogger)               // Added

	// Initialize services
	jwtKey := []byte(os.Getenv("JWT_SECRET_TEST"))
	if len(jwtKey) == 0 {
		jwtKey = []byte("test_secret_key_for_router_tests_effiplat")
	}
	authService := service.NewAuthService(userRepo, jwtKey, appLogger)
	userService := service.NewUserService(userRepo, roleRepo, appLogger)
	roleService := service.NewRoleService(roleRepo, appLogger)
	permissionService := service.NewPermissionService(permRepo, roleRepo, appLogger)
	responsibilityService := service.NewResponsibilityService(responsibilityRepo, appLogger)
	responsibilityGroupService := service.NewResponsibilityGroupService(responsibilityGroupRepo, responsibilityRepo, appLogger)
	environmentService := service.NewEnvironmentService(environmentRepo, appLogger)
	assetService := service.NewAssetService(assetRepo, environmentRepo, appLogger)
	serviceService := service.NewServiceService(serviceRepo, serviceTypeRepo, appLogger)                                      // Renamed serviceSvc to serviceService and added logger
	serviceInstanceService := service.NewServiceInstanceService(serviceInstanceRepo, serviceRepo, environmentRepo, appLogger) // Added
	businessService := service.NewBusinessService(businessRepo, appLogger)                                                    // Added
	bugService := service.NewBugService(bugRepo)
	auditLogService := service.NewAuditLogService(auditLogRepo, appLogger) // 审计日志服务

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService, auditLogService, appLogger)
	roleHandler := handler.NewRoleHandler(roleService, auditLogService, appLogger)
	permissionHandler := handler.NewPermissionHandler(permissionService, auditLogService, appLogger)
	responsibilityHandler := handler.NewResponsibilityHandler(responsibilityService, auditLogService, appLogger)
	responsibilityGroupHandler := handler.NewResponsibilityGroupHandler(responsibilityGroupService, auditLogService, appLogger)
	environmentHandler := handler.NewEnvironmentHandler(environmentService, auditLogService, appLogger)
	assetHandler := handler.NewAssetHandler(assetService, auditLogService, appLogger)
	serviceHandler := handler.NewServiceHandler(serviceService, auditLogService, appLogger)                     // Use handler.NewServiceHandler
	serviceInstanceHandler := handler.NewServiceInstanceHandler(serviceInstanceService, auditLogService, appLogger) // Added
	businessHandler := handler.NewBusinessHandler(businessService, appLogger)                      // Added
	bugHandler := handler.NewBugHandler(bugService) // Added BugHandler
	auditLogHandler := handler.NewAuditLogHandler(auditLogService, appLogger) // 审计日志处理器

	routerInstance := SetupRouter(
		authHandler,
		userHandler,
		roleHandler,
		permissionHandler,
		responsibilityHandler,
		responsibilityGroupHandler,
		environmentHandler,
		assetHandler,
		serviceHandler,
		serviceInstanceHandler, // Pass the new handler
		businessHandler,        // Pass the new handler
		bugHandler,             // Pass the new handler
		auditLogHandler,
		auditLogService,
		jwtKey,
	)

	return TestAppComponents{
		Router:                     routerInstance,
		DB:                         db,
		Logger:                     appLogger,
		AuthHandler:                authHandler,
		UserHandler:                userHandler,
		RoleHandler:                roleHandler,
		PermissionHandler:          permissionHandler,
		ResponsibilityHandler:      responsibilityHandler,
		ResponsibilityGroupHandler: responsibilityGroupHandler,
		EnvironmentHandler:         environmentHandler,
		AssetHandler:               assetHandler,
		ServiceHandler:             serviceHandler,
		ServiceInstanceHandler:     serviceInstanceHandler, // Added
		BusinessHandler:            businessHandler,        // Added
		BugHandler:                 bugHandler,             // Added
		AuditLogHandler:            auditLogHandler,
		AuditLogService:            auditLogService,
		JWTKey:                     jwtKey,
	}
}

// GetAuthTokenForTest creates a standard test user if not exists, logs in, and returns a JWT token.
func GetAuthTokenForTest(t *testing.T, r *gin.Engine, db *gorm.DB) string {
	// Use a unique email for each test run to avoid conflicts
	testUserEmail := fmt.Sprintf("testuser_%d_%s@example.com", time.Now().UnixNano(), uuid.New().String()[:8])
	const testUserPassword = "password123"

	// Check if user exists, if not create
	var user model.User
	err := db.Where("email = ?", testUserEmail).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testUserPassword), bcrypt.DefaultCost)
		user = model.User{
			Name:     "Test User",
			Email:    testUserEmail,
			Password: string(hashedPassword),
			Status:   "active",
		}
		err = db.Create(&user).Error
		assert.NoError(t, err, "Failed to create test user")
	} else {
		assert.NoError(t, err, "Error checking for test user")
	}

	// Login to get token
	loginPayload := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, testUserEmail, testUserPassword)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(loginPayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Test user login failed")

	var structuredLoginResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string       `json:"token"`
			User  *model.User `json:"user"`
		} `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &structuredLoginResp)
	assert.NoError(t, err, "Failed to unmarshal structured login response")
	assert.Equal(t, 0, structuredLoginResp.Code, "Login response code should be 0")
	assert.NotEmpty(t, structuredLoginResp.Data.Token, "Login token should not be empty")

	return structuredLoginResp.Data.Token
}

// Helper to create a test user directly in the DB
// Moved from router_test.go
func CreateTestUser(db *gorm.DB, email, password string) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Name:     "Test User", // Name can be generic for this helper
		Email:    email,       // Use the provided email directly
		Password: string(hashedPassword),
		Status:   "active",
	}
	result := db.Create(user)
	return user, result.Error
}

// GetAdminToken utility to get admin token.
// Moved from router_test.go and modified to use TestAppComponents.
func GetAdminToken(t *testing.T, components TestAppComponents) string {
	// Ensuring a consistent admin user for these tests
	adminEmailForToken := "admin_get_token_user@example.com" // Use a distinct email
	adminPasswordForToken := "AdminPassSecure123!"

	// Attempt to find the admin user first
	var existingAdmin model.User
	err := components.DB.Where("email = ?", adminEmailForToken).First(&existingAdmin).Error

	if errors.Is(err, gorm.ErrRecordNotFound) { // User does not exist, create it
		_, createErr := CreateTestUser(components.DB, adminEmailForToken, adminPasswordForToken)
		// We require successful creation if the user wasn't found. Using require for fatal error.
		require.NoErrorf(t, createErr, "Failed to create admin user '%s' for token generation: %v", adminEmailForToken, createErr)
	} else if err != nil { // Some other DB error occurred during find
		// If there was an error other than not found, we should fail fast.
		t.Fatalf("Failed to query for admin user '%s': %v", adminEmailForToken, err)
	}
	// If user was found or successfully created, proceed to login

	loginPayload := model.LoginRequest{Email: adminEmailForToken, Password: adminPasswordForToken}
	payloadBytes, _ := json.Marshal(loginPayload)
	loginReq, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(payloadBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	components.Router.ServeHTTP(loginW, loginReq)
	assert.Equal(t, http.StatusOK, loginW.Code, "Admin login failed for token generation in GetAdminToken")

	var adminLoginResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string       `json:"token"`
			User  *model.User `json:"user"`
		} `json:"data"`
	}
	err = json.Unmarshal(loginW.Body.Bytes(), &adminLoginResp)
	assert.NoError(t, err, "Failed to unmarshal admin login response in GetAdminToken")
	assert.Equal(t, 0, adminLoginResp.Code, "Admin login response code should be 0")
	assert.NotEmpty(t, adminLoginResp.Data.Token, "Admin token should not be empty in GetAdminToken")
	return adminLoginResp.Data.Token
}

// TODO: Define mock repositories if not using actual GORM implementations yet
// type mockResponsibilityRepository struct{}
// func (m *mockResponsibilityRepository) Create(ctx context.Context, r *model.Responsibility) (*model.Responsibility, error) { return nil, errors.New("mock: not implemented")}
// func (m *mockResponsibilityRepository) List(ctx context.Context, p model.ResponsibilityListParams) ([]model.Responsibility, int64, error) { return nil, 0, errors.New("mock: not implemented")}
// func (m *mockResponsibilityRepository) GetByID(ctx context.Context, id uint) (*model.Responsibility, error) { return nil, errors.New("mock: not implemented")}
// func (m *mockResponsibilityRepository) Update(ctx context.Context, r *model.Responsibility) (*model.Responsibility, error) { return nil, errors.New("mock: not implemented")}
// func (m *mockResponsibilityRepository) Delete(ctx context.Context, id uint) error { return errors.New("mock: not implemented")}

// type mockResponsibilityGroupRepository struct{}
// ... (similar mock implementations for group repo)

// 测试辅助函数：创建责任并带调试日志
func CreateTestResponsibility(t *testing.T, router http.Handler, token string, respModel *model.Responsibility) model.Responsibility {
	jsonData, err := json.Marshal(respModel)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/responsibilities", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Println("[DEBUG] Create responsibility response:", w.Body.String())
	assert.Equal(t, http.StatusCreated, w.Code)

	var createdResp struct {
		Data model.Responsibility `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &createdResp)
	assert.NoError(t, err)
	fmt.Println("[DEBUG] Created responsibility ID:", createdResp.Data.ID)
	assert.NotZero(t, createdResp.Data.ID)
	return createdResp.Data
}
