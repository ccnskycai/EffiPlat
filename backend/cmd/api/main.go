package main

import (
	"EffiPlat/backend/internal"                // Wire生成的依赖注入初始化函数
	"EffiPlat/backend/internal/pkg/config"
	pkgdb "EffiPlat/backend/internal/pkg/database"
	"EffiPlat/backend/internal/pkg/logger"
	"EffiPlat/backend/internal/router"

	"fmt"
	"log"
	"os"

	// "github.com/gin-gonic/gin" // <-- Gin 初始化移到 router 包，这里可能不再需要
	"go.uber.org/zap"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("CRITICAL: Failed to load configuration: %v", err)
	}

	// 2. Initialize Logger
	appLogger, err := logger.NewLogger(cfg.Logger)
	if err != nil {
		log.Fatalf("CRITICAL: Failed to initialize logger: %v", err)
	}
	defer func() { _ = appLogger.Sync() }()
	appLogger.Info("Configuration and Logger initialized successfully")

	// 3. Initialize Database Connection
	dbConn, err := pkgdb.NewConnection(cfg.Database, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	appLogger.Info("Database connection established successfully")

	// 4. Get JWT Key
	jwtKey := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtKey) == 0 {
		appLogger.Warn("JWT_SECRET not configured. Using default, which is insecure.")
		jwtKey = []byte("default_insecure_secret_key_for_dev_only")
	}

	// 5. Initialize Dependencies
	// Initialize Auth components using Wire
	authHandler, err := internal.InitializeAuthHandler(dbConn, jwtKey, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize auth handler", zap.Error(err))
	}

	// Initialize User components using Wire
	userHandler, err := internal.InitializeUserHandler(dbConn, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize user handler", zap.Error(err))
	}

	// Initialize Role components using Wire
	roleHandler, err := internal.InitializeRoleHandler(dbConn, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize role handler", zap.Error(err))
	}

	// Initialize Permission components (assuming a similar Wire setup)
	permissionHandler, err := internal.InitializePermissionHandler(dbConn, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize permission handler", zap.Error(err))
	}

	// Initialize Responsibility components (assuming a similar Wire setup)
	responsibilityHandler, err := internal.InitializeResponsibilityHandler(dbConn, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize responsibility handler", zap.Error(err))
	}

	// Initialize ResponsibilityGroup components (assuming a similar Wire setup)
	responsibilityGroupHandler, err := internal.InitializeResponsibilityGroupHandler(dbConn, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize responsibility group handler", zap.Error(err))
	}

	// Initialize Environment components
	environmentHandler, err := internal.InitializeEnvironmentHandler(dbConn, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize environment handler", zap.Error(err))
	}

	// 获取环境仓库以供其他组件使用
	environmentRepository, err := internal.InitializeEnvironmentRepository(dbConn, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize environment repository", zap.Error(err))
	}

	// Initialize Asset components
	assetHandler, err := internal.InitializeAssetHandler(dbConn, appLogger, environmentRepository)
	if err != nil {
		appLogger.Fatal("Failed to initialize asset handler", zap.Error(err))
	}

	// Initialize Service components
	serviceHandler, err := internal.InitializeServiceHandler(dbConn, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize service handler", zap.Error(err))
	}

	// 获取服务仓库以供其他组件使用
	serviceRepository, err := internal.InitializeServiceRepository(dbConn)
	if err != nil {
		appLogger.Fatal("Failed to initialize service repository", zap.Error(err))
	}

	// Initialize ServiceInstance components using Wire
	serviceInstanceHandler, err := internal.InitializeServiceInstanceHandler(dbConn, appLogger, serviceRepository, environmentRepository)
	if err != nil {
		appLogger.Fatal("Failed to initialize service instance handler", zap.Error(err))
	}

	// Initialize Business components
	businessHandler, err := internal.InitializeBusinessHandler(dbConn, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize business handler", zap.Error(err))
	}

	// Initialize Bug components
	bugHandler, err := internal.InitializeBugHandler(dbConn, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize bug handler", zap.Error(err))
	}
	
	// Initialize Audit Log components
	auditLogService, err := internal.InitializeAuditLogService(dbConn, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize audit log service", zap.Error(err))
	}
	
	auditLogHandler, err := internal.InitializeAuditLogHandler(dbConn, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize audit log handler", zap.Error(err))
	}

	// 6. Setup Router
	// SetupRouter expects *handler.AuthHandler and *handler.UserHandler (after UserHandler moves)
	r := router.SetupRouter(
		authHandler,
		userHandler,
		roleHandler,
		permissionHandler,
		responsibilityHandler,
		responsibilityGroupHandler,
		environmentHandler,
		assetHandler,
		serviceHandler,
		serviceInstanceHandler,
		businessHandler,
		bugHandler,
		auditLogHandler,        // 添加审计日志处理器
		auditLogService,        // 添加审计日志服务
		jwtKey,
	)

	// 7. Start Server
	portStr := fmt.Sprintf(":%d", cfg.Server.Port)
	appLogger.Info("Starting backend server", zap.String("address", "http://localhost"+portStr))
	if err := r.Run(portStr); err != nil {
		appLogger.Fatal("Failed to start server", zap.Error(err))
	}
}
