package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"EffiPlat/backend/internal/pkg/config"
	"EffiPlat/backend/internal/pkg/logger"

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
	defer func() {
		_ = appLogger.Sync()
	}()

	appLogger.Info("Configuration and Logger initialized successfully",
		zap.String("log_level", cfg.Logger.Level),
		zap.Int("server_port", cfg.Server.Port),
	)

	// 3. Initialize Database Connection (Placeholder)
	// dbConn, err := database.NewConnection(cfg.Database)
	// if err != nil {
	// 	 appLogger.Fatal("Failed to connect to database", zap.Error(err))
	// }
	// appLogger.Info("Database connection established")

	// 4. Setup Dependency Injection (Placeholder)
	// repositories := repository.NewRepositories(dbConn, appLogger)
	// services := service.NewServices(repositories, appLogger, cfg)
	// handlers := handler.NewHandlers(services, appLogger)

	// 5. Setup HTTP Server (Gin or standard net/http)
	// Example using standard net/http for now
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		appLogger.Debug("Health check requested")
		fmt.Fprintf(w, "OK")
	})
	// Add other routes here, passing handlers
	// http.HandleFunc("/api/v1/login", handlers.Login)

	// 6. Start Server
	portStr := strconv.Itoa(cfg.Server.Port)
	appLogger.Info("Starting backend server", zap.String("address", "http://localhost:"+portStr))

	if err := http.ListenAndServe(":"+portStr, nil); err != nil {
		appLogger.Fatal("Failed to start server", zap.Error(err))
	}
}
