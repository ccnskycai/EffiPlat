package database

import (
	"context"
	"fmt"
	"time"

	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/pkg/config" // Use correct module path

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// NewConnection establishes a new database connection based on the configuration.
func NewConnection(cfg config.DBConfig, appLogger *zap.Logger) (*gorm.DB, error) {
	if cfg.Type != "sqlite" && cfg.Type != "sqlite3" {
		return nil, fmt.Errorf("unsupported database type: %s, currently only sqlite is supported for V1.0", cfg.Type)
	}

	// Configure GORM logger
	// We can integrate GORM logging with our main zap logger
	newGormLogger := logger.New(
		NewZapGormLogger(appLogger),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // Slow SQL threshold
			LogLevel:                  logger.Warn,            // Log level (Warn, Error, Info)
			IgnoreRecordNotFoundError: true,                   // Don't log ErrRecordNotFound errors
			Colorful:                  false,                  // Disable color (usually for JSON logs)
		},
	)

	// Assemble DSN - For SQLite, DSN is the file path.
	// We can append params like _foreign_keys=on if needed.
	dsn := cfg.DSN
	// if cfg.Params != "" {
	// 	dsn = fmt.Sprintf("%s?%s", cfg.DSN, cfg.Params)
	// }

	appLogger.Info("Attempting to connect to database", zap.String("type", cfg.Type), zap.String("dsn", dsn))

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: newGormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false, // Use plural table names (e.g., users)
		},
		// Add other GORM configs if needed
		// DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		appLogger.Error("Failed to connect database", zap.Error(err))
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	appLogger.Info("Database connection established successfully")

	// Optional: Configure connection pool (less critical for SQLite)
	// sqlDB, err := db.DB()
	// if err != nil {
	// 	 return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	// }
	// sqlDB.SetMaxIdleConns(cfg.Pool.MaxIdleConns)
	// sqlDB.SetMaxOpenConns(cfg.Pool.MaxOpenConns)
	// sqlDB.SetConnMaxLifetime(cfg.Pool.ConnMaxLifetime)

	// Optional: Auto-migrate schema (useful for development, be careful in production)
	// err = db.AutoMigrate(&models.User{}, &models.Role{}, ...) // Add your models here
	// if err != nil {
	// 	 return nil, fmt.Errorf("failed to auto migrate database schema: %w", err)
	// }
	// appLogger.Info("Database schema auto-migration completed (if enabled)")

	return db, nil
}

// AutoMigrate runs GORM auto-migration for all necessary models.
func AutoMigrate(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("Starting database auto-migration...")

	// List all models that need to be migrated
	err := db.AutoMigrate(
		&models.User{},           // From models/user.go
		&models.Role{},           // From models/user.go
		&models.UserRole{},       // From models/user.go
		&models.Permission{},     // From models/permission_models.go
		&models.RolePermission{}, // From models/permission_models.go
		// Add other models here if any
	)

	if err != nil {
		logger.Error("Database auto-migration failed", zap.Error(err))
		return fmt.Errorf("database auto-migration failed: %w", err)
	}

	logger.Info("Database auto-migration completed successfully.")
	return nil
}

// ---- GORM Logger Integration with Zap ----

// ZapGormLogger implements gorm logger.Interface using zap
type ZapGormLogger struct {
	zapLogger *zap.Logger
}

// NewZapGormLogger creates a new ZapGormLogger
func NewZapGormLogger(zapLogger *zap.Logger) *ZapGormLogger {
	return &ZapGormLogger{zapLogger: zapLogger.Named("gorm")}
}

func (l *ZapGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	// GORM's logger level doesn't directly map 1:1 with zap levels for dynamic change here
	// We control level via zap core. Returning `l` is standard.
	return l
}

func (l *ZapGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.zapLogger.Info(fmt.Sprintf(msg, data...))
}

func (l *ZapGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.zapLogger.Warn(fmt.Sprintf(msg, data...))
}

func (l *ZapGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.zapLogger.Error(fmt.Sprintf(msg, data...))
}

func (l *ZapGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []zap.Field{
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
		zap.String("sql", sql),
	}

	if err != nil && err != gorm.ErrRecordNotFound { // Don't log RecordNotFound as Error level by default
		l.zapLogger.Error("GORM Trace", append(fields, zap.Error(err))...)
	} else if elapsed > 200*time.Millisecond { // Log slow queries as Warn
		l.zapLogger.Warn("GORM Trace - Slow Query", fields...)
	} else {
		l.zapLogger.Debug("GORM Trace", fields...) // Log normal queries as Debug
	}
}

// Printf implements gorm.io/gorm/logger.Writer interface
func (l *ZapGormLogger) Printf(format string, v ...interface{}) {
	// Log messages from GORM's writer interface at Info level
	l.zapLogger.Info(fmt.Sprintf(format, v...))
}
