package main

import (
	"EffiPlat/backend/internal"
	"EffiPlat/backend/internal/pkg/config"
	"EffiPlat/backend/internal/pkg/database"
	"EffiPlat/backend/internal/pkg/logger"
	"EffiPlat/backend/internal/router" // <-- 添加 router 包导入
	"fmt"                              // <-- 添加 fmt 包导入 (如果之前没有)
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
	dbConn, err := database.NewConnection(cfg.Database, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	appLogger.Info("Database connection established successfully")

	// 4. Get JWT Key (从配置或环境变量)
	jwtKey := []byte(os.Getenv("JWT_SECRET")) // 或者 cfg.Server.JWTSecret
	if len(jwtKey) == 0 {
		appLogger.Warn("JWT_SECRET not configured. Authentication middleware might fail if enabled.")
		// 如果 JWT 是必须的，这里应该 Fatal
		// appLogger.Fatal("JWT_SECRET not configured")
	}

	// 5. Initialize Dependencies using Wire
	// 确保 InitAuthHandler 返回的是实例而不是接口，或者调整 Wire 配置
	authHandler := internal.InitAuthHandler(dbConn, jwtKey, appLogger) // 传递 Logger

	// --- 移除旧的 Gin 引擎创建和路由注册 ---
	// r := gin.Default()
	// r.POST("/api/v1/auth/login", authHandler.Login)
	// ---------------------------------------

	// 6. Setup Router
	// 将需要的 Handler 传递给 SetupRouter
	// 如果 SetupRouter 需要 jwtKey，也需要传递进去
	r := router.SetupRouter(authHandler, jwtKey) // <-- 传递 jwtKey

	// 7. Start Server
	portStr := fmt.Sprintf(":%d", cfg.Server.Port)
	appLogger.Info("Starting backend server", zap.String("address", "http://localhost"+portStr))
	if err := r.Run(portStr); err != nil {
		appLogger.Fatal("Failed to start server", zap.Error(err))
	}
}
