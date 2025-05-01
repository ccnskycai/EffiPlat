package config

import (
	"fmt"
	"os"
	"strings"

	"EffiPlat/backend/internal/pkg/logger"

	"github.com/spf13/viper"
)

// AppConfig holds the application's configuration
type AppConfig struct {
	Server   ServerConfig  `mapstructure:"server"`
	Database DBConfig      `mapstructure:"database"`
	Logger   logger.Config `mapstructure:"logger"`
	// Add other configuration sections as needed
}

// ServerConfig holds server specific configuration
type ServerConfig struct {
	Port int `mapstructure:"port"`
	// ... other server settings
}

// DBConfig holds database specific configuration
type DBConfig struct {
	Type string `mapstructure:"type"`
	DSN  string `mapstructure:"dsn"`
	// ... other database settings
}

// LoadConfig reads configuration from file and environment variables
func LoadConfig(configPath string) (*AppConfig, error) {
	v := viper.New()

	// --- Configuration File Settings ---
	// Set default config file name (can be overridden by env var later)
	defaultConfigName := "config.dev" // Default to development
	configName := os.Getenv("APP_CONFIG_NAME")
	if configName == "" {
		configName = defaultConfigName
	}
	v.SetConfigName(configName) // e.g., config.dev, config.prod
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)      // Path to look for the config file in
	v.AddConfigPath("./configs")     // Look in ./configs/
	v.AddConfigPath("../configs")    // Look in ../configs/ (relative to backend/)
	v.AddConfigPath("/etc/appname/") // Look in /etc/appname/
	v.AddConfigPath(".")             // Look in the current directory

	// --- Environment Variable Settings ---
	v.SetEnvPrefix("APP") // Environment variables must be prefixed with APP_
	v.AutomaticEnv()      // Read in environment variables that match
	// Example: APP_LOGGER_LEVEL=debug will override logger.level from file
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Replace dots with underscores for env var names

	// --- Set Defaults ---
	v.SetDefault("server.port", 8080)
	// Set defaults for logger (including lumberjack) before reading config
	logger.AddLumberjackToViper(v)
	// Add other defaults here

	// --- Read Configuration ---
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired and rely on defaults/env vars
			fmt.Fprintf(os.Stderr, "Config file not found (%s.yaml), using defaults/env vars: %v\n", configName, err)
		} else {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// --- Unmarshal Config ---
	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	fmt.Printf("Configuration loaded successfully from %s.yaml\n", configName)
	// Optionally print loaded config for debugging (use logger once initialized)
	// fmt.Printf("Loaded config: %+v\n", cfg)

	return &cfg, nil
}
