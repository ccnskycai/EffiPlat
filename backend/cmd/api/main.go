package main

import (
	"EffiPlat/backend/internal"
	"EffiPlat/backend/internal/handler" // Added for NewUserHandler and AuthHandler type
	"EffiPlat/backend/internal/pkg/config"
	pkgdb "EffiPlat/backend/internal/pkg/database"
	"EffiPlat/backend/internal/pkg/logger"
	"EffiPlat/backend/internal/repository"
	"EffiPlat/backend/internal/router"
	"EffiPlat/backend/internal/service" // Added for NewUserService

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
	// InitAuthHandler is expected to return *handler.AuthHandler
	authHandler := internal.InitAuthHandler(dbConn, jwtKey, appLogger)

	// Initialize User components
	userRepoImpl := repository.NewUserRepository(dbConn)
	userService := service.NewUserService(userRepoImpl) // Use service.NewUserService
	userHandler := handler.NewUserHandler(userService) // Use handler.NewUserHandler

	// 6. Setup Router
	// SetupRouter expects *handler.AuthHandler and *handler.UserHandler (after UserHandler moves)
	r := router.SetupRouter(authHandler, userHandler, jwtKey)

	// 7. Start Server
	portStr := fmt.Sprintf(":%d", cfg.Server.Port)
	appLogger.Info("Starting backend server", zap.String("address", "http://localhost"+portStr))
	if err := r.Run(portStr); err != nil {
		appLogger.Fatal("Failed to start server", zap.Error(err))
	}
}
