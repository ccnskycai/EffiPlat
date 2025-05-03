package main

import (
	configLoader "EffiPlat/backend/internal/pkg/config" // Renamed import
	db "EffiPlat/backend/internal/pkg/database"         // Renamed import
	logger "EffiPlat/backend/internal/pkg/logger"
	"EffiPlat/backend/internal/seed"
	"log"

	"go.uber.org/zap"
)

func main() {
	// 1. Load Configuration
	// Assuming config file is in ../../configs relative to cmd/seeder/main.go
	// Adjust the path ("." or "../../configs") based on where you run the built binary
	// or use environment variables.
	cfg, err := configLoader.LoadConfig("../../configs")
	if err != nil {
		log.Fatalf("CRITICAL: Failed to load configuration for seeder: %v", err)
	}

	// 2. Initialize Logger
	appLogger, err := logger.NewLogger(cfg.Logger)
	if err != nil {
		log.Fatalf("CRITICAL: Failed to initialize logger for seeder: %v", err)
	}
	defer func() {
		_ = appLogger.Sync()
	}()

	appLogger.Info("Seeder configuration and logger initialized successfully")

	// 3. Initialize Database Connection
	// Using the same database package
	dbConn, err := db.NewConnection(cfg.Database, appLogger)
	if err != nil {
		appLogger.Fatal("Seeder failed to connect to database", zap.Error(err))
	}
	appLogger.Info("Seeder database connection established successfully")

	// 4. Run the Seeder
	if err := seed.SeedAll(dbConn); err != nil {
		appLogger.Fatal("Database seeding failed", zap.Error(err))
	}

	appLogger.Info("Seeder finished successfully.")
}
