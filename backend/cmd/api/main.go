package main

import (
	"EffiPlat/backend/internal"                // Added for NewUserHandler and AuthHandler type
	hdlrs "EffiPlat/backend/internal/handlers" // Use alias hdlrs
	"EffiPlat/backend/internal/pkg/config"
	pkgdb "EffiPlat/backend/internal/pkg/database"
	"EffiPlat/backend/internal/pkg/logger"
	"EffiPlat/backend/internal/repository"
	"EffiPlat/backend/internal/router"
	"EffiPlat/backend/internal/service"

	// Added for NewUserService
	// user "EffiPlat/backend/internal/user" // Removed, types now in handler & service
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
	environmentRepository := repository.NewGormEnvironmentRepository(dbConn, appLogger)
	environmentService := service.NewEnvironmentService(environmentRepository, appLogger)
	environmentHandler := hdlrs.NewEnvironmentHandler(environmentService, appLogger) // Use alias
	// TODO: Consider adding InitializeEnvironmentHandler to wire.go for consistency if this becomes permanent

	// Initialize Asset components
	assetRepository := repository.NewGormAssetRepository(dbConn, appLogger)
	// AssetService needs EnvironmentRepository to validate EnvironmentID
	assetService := service.NewAssetService(assetRepository, environmentRepository, appLogger)
	assetHandler := hdlrs.NewAssetHandler(assetService, appLogger)
	// TODO: Consider adding InitializeAssetHandler to wire.go for consistency

	// Initialize Service components
	serviceRepository := repository.NewGormServiceRepository(dbConn)
	serviceTypeRepository := repository.NewGormServiceTypeRepository(dbConn) // Added ServiceTypeRepository
	serviceService := service.NewServiceService(serviceRepository, serviceTypeRepository, appLogger) 
	serviceHandler := hdlrs.NewServiceHandler(serviceService, appLogger)
	// TODO: Consider adding InitializeServiceHandler to wire.go for consistency if this becomes permanent

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
		assetHandler,   // Added assetHandler
		serviceHandler, // Added serviceHandler
		jwtKey,
	)

	// 7. Start Server
	portStr := fmt.Sprintf(":%d", cfg.Server.Port)
	appLogger.Info("Starting backend server", zap.String("address", "http://localhost"+portStr))
	if err := r.Run(portStr); err != nil {
		appLogger.Fatal("Failed to start server", zap.Error(err))
	}
}
