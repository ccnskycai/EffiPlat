package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper" // Assuming viper is used for config reading elsewhere
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config structure maps to the 'logger' section in YAML config
type Config struct {
	Level            string                 `mapstructure:"level"`
	Encoding         string                 `mapstructure:"encoding"`
	OutputPaths      []string               `mapstructure:"outputPaths"`
	ErrorOutputPaths []string               `mapstructure:"errorOutputPaths"`
	Development      bool                   `mapstructure:"development"`
	EncoderConfig    EncoderConfig          `mapstructure:"encoderConfig"`
	InitialFields    map[string]interface{} `mapstructure:"initialFields"`
	Lumberjack       *LumberjackConfig      `mapstructure:"lumberjack"` // Optional lumberjack config
}

// EncoderConfig maps to zapcore EncoderConfig options
type EncoderConfig struct {
	MessageKey      string `mapstructure:"messageKey"`
	LevelKey        string `mapstructure:"levelKey"`
	TimeKey         string `mapstructure:"timeKey"`
	NameKey         string `mapstructure:"nameKey"`
	CallerKey       string `mapstructure:"callerKey"`
	StacktraceKey   string `mapstructure:"stacktraceKey"`
	LineEnding      string `mapstructure:"lineEnding"`
	LevelEncoder    string `mapstructure:"levelEncoder"`
	TimeEncoder     string `mapstructure:"timeEncoder"`
	DurationEncoder string `mapstructure:"durationEncoder"`
	CallerEncoder   string `mapstructure:"callerEncoder"`
	NameEncoder     string `mapstructure:"nameEncoder"`
}

// LumberjackConfig maps to lumberjack configuration options
type LumberjackConfig struct {
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"maxsize"` // megabytes
	MaxAge     int    `mapstructure:"maxage"`  // days
	MaxBackups int    `mapstructure:"maxbackups"`
	LocalTime  bool   `mapstructure:"localtime"`
	Compress   bool   `mapstructure:"compress"`
}

// Global atomic level for dynamic level changes
var atom = zap.NewAtomicLevel()

// NewLogger initializes a zap logger based on the provided configuration
func NewLogger(cfg Config) (*zap.Logger, error) {

	// Set the global atomic level
	if err := atom.UnmarshalText([]byte(cfg.Level)); err != nil {
		atom.SetLevel(zap.InfoLevel) // Default to info if config is invalid
		fmt.Fprintf(os.Stderr, "Invalid log level '%s' in config, defaulting to 'info': %v\n", cfg.Level, err)
	}

	// Configure the encoder
	encoderCfg := buildEncoderConfig(cfg)

	var encoder zapcore.Encoder
	if cfg.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else if cfg.Encoding == "json" {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		fmt.Fprintf(os.Stderr, "Invalid logger encoding '%s', defaulting to 'json'\n", cfg.Encoding)
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	// Determine output writers
	writeSyncer, errSyncer, err := buildWriteSyncers(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build write syncers: %w", err)
	}

	// Create the core
	core := zapcore.NewCore(encoder, writeSyncer, atom)

	// Build logger options
	opts := []zap.Option{
		zap.ErrorOutput(errSyncer), // Zap internal errors go here
		zap.AddCallerSkip(1),       // Adjust caller skip if logger is wrapped
	}
	if cfg.Development {
		opts = append(opts, zap.Development())                     // Adds caller, stacktrace on Warn+
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel)) // Add stacktrace for errors in dev
	} else {
		opts = append(opts, zap.AddStacktrace(zapcore.PanicLevel)) // Add stacktrace only for panics in prod
	}
	if len(cfg.InitialFields) > 0 {
		fields := make([]zap.Field, 0, len(cfg.InitialFields))
		for k, v := range cfg.InitialFields {
			fields = append(fields, zap.Any(k, v))
		}
		opts = append(opts, zap.Fields(fields...))
	}

	// Build the logger
	logger := zap.New(core, opts...)

	logger.Info("Logger initialized", zap.String("level", cfg.Level), zap.String("encoding", cfg.Encoding))

	return logger, nil
}

// buildEncoderConfig creates zapcore.EncoderConfig from our Config
func buildEncoderConfig(cfg Config) zapcore.EncoderConfig {
	encoderCfg := zap.NewProductionEncoderConfig() // Start with prod defaults
	if cfg.Development {
		encoderCfg = zap.NewDevelopmentEncoderConfig() // Switch to dev defaults if needed
	}

	// Apply overrides from config
	if cfg.EncoderConfig.MessageKey != "" {
		encoderCfg.MessageKey = cfg.EncoderConfig.MessageKey
	}
	if cfg.EncoderConfig.LevelKey != "" {
		encoderCfg.LevelKey = cfg.EncoderConfig.LevelKey
	}
	if cfg.EncoderConfig.TimeKey != "" {
		encoderCfg.TimeKey = cfg.EncoderConfig.TimeKey
	}
	if cfg.EncoderConfig.NameKey != "" {
		encoderCfg.NameKey = cfg.EncoderConfig.NameKey
	}
	if cfg.EncoderConfig.CallerKey != "" {
		encoderCfg.CallerKey = cfg.EncoderConfig.CallerKey
	}
	if cfg.EncoderConfig.StacktraceKey != "" {
		encoderCfg.StacktraceKey = cfg.EncoderConfig.StacktraceKey
	}
	if cfg.EncoderConfig.LineEnding != "" {
		encoderCfg.LineEnding = cfg.EncoderConfig.LineEnding
	}

	// Handle specific encoder functions based on strings
	parseLevelEncoder(cfg.EncoderConfig.LevelEncoder, &encoderCfg)
	parseTimeEncoder(cfg.EncoderConfig.TimeEncoder, &encoderCfg)
	parseDurationEncoder(cfg.EncoderConfig.DurationEncoder, &encoderCfg)
	parseCallerEncoder(cfg.EncoderConfig.CallerEncoder, &encoderCfg)
	parseNameEncoder(cfg.EncoderConfig.NameEncoder, &encoderCfg)

	return encoderCfg
}

// buildWriteSyncers creates the zapcore.WriteSyncer for logs and errors
func buildWriteSyncers(cfg Config) (zapcore.WriteSyncer, zapcore.WriteSyncer, error) {
	var logSyncer, errSyncer zapcore.WriteSyncer
	var err error

	logSyncer, err = openLogSyncers(cfg.OutputPaths, cfg.Lumberjack)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open log output syncers: %w", err)
	}

	errSyncer, err = openLogSyncers(cfg.ErrorOutputPaths, cfg.Lumberjack) // Use same lumberjack config if file path matches
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open error output syncers: %w", err)
	}

	return logSyncer, errSyncer, nil
}

// openLogSyncers opens multiple log destinations (stdout, stderr, files with lumberjack)
func openLogSyncers(paths []string, ljConfig *LumberjackConfig) (zapcore.WriteSyncer, error) {
	syncers := make([]zapcore.WriteSyncer, 0, len(paths))
	for _, path := range paths {
		lowerPath := strings.ToLower(path)
		if lowerPath == "stdout" {
			syncers = append(syncers, zapcore.AddSync(os.Stdout))
		} else if lowerPath == "stderr" {
			syncers = append(syncers, zapcore.AddSync(os.Stderr))
		} else {
			// Assume it's a file path, potentially use lumberjack
			if ljConfig != nil && ljConfig.Filename == path { // Check if this path uses lumberjack
				lumberLog := &lumberjack.Logger{
					Filename:   ljConfig.Filename,
					MaxSize:    ljConfig.MaxSize,
					MaxAge:     ljConfig.MaxAge,
					MaxBackups: ljConfig.MaxBackups,
					LocalTime:  ljConfig.LocalTime,
					Compress:   ljConfig.Compress,
				}
				syncers = append(syncers, zapcore.AddSync(lumberLog))
			} else {
				// Simple file output without rotation (less common)
				file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					return nil, fmt.Errorf("failed to open log file %s: %w", path, err)
				}
				syncers = append(syncers, zapcore.AddSync(file))
			}
		}
	}

	if len(syncers) == 0 {
		// Default to stdout if no paths specified
		return zapcore.AddSync(os.Stdout), nil
	}
	return zapcore.NewMultiWriteSyncer(syncers...), nil
}

// Helper functions to parse encoder strings
func parseLevelEncoder(s string, cfg *zapcore.EncoderConfig) {
	switch strings.ToLower(s) {
	case "capital":
		cfg.EncodeLevel = zapcore.CapitalLevelEncoder
	case "capitalcolor", "capitalcolour":
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	case "lowercase", "lower":
		cfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	case "lowercasecolor", "lowercasecolour": // Note: zap doesn't have built-in lowercase color
		cfg.EncodeLevel = zapcore.LowercaseLevelEncoder // Default to lowercase
	}
}

func parseTimeEncoder(s string, cfg *zapcore.EncoderConfig) {
	switch strings.ToLower(s) {
	case "epoch":
		cfg.EncodeTime = zapcore.EpochTimeEncoder
	case "epochmillis":
		cfg.EncodeTime = zapcore.EpochMillisTimeEncoder
	case "epochnanos":
		cfg.EncodeTime = zapcore.EpochNanosTimeEncoder
	case "iso8601":
		cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	case "rfc3339":
		cfg.EncodeTime = zapcore.RFC3339TimeEncoder
	case "rfc3339nano":
		cfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	}
}

func parseDurationEncoder(s string, cfg *zapcore.EncoderConfig) {
	switch strings.ToLower(s) {
	case "seconds":
		cfg.EncodeDuration = zapcore.SecondsDurationEncoder
	case "nanos":
		cfg.EncodeDuration = zapcore.NanosDurationEncoder
	case "string":
		cfg.EncodeDuration = zapcore.StringDurationEncoder
	}
}

func parseCallerEncoder(s string, cfg *zapcore.EncoderConfig) {
	switch strings.ToLower(s) {
	case "short":
		cfg.EncodeCaller = zapcore.ShortCallerEncoder
	case "full":
		cfg.EncodeCaller = zapcore.FullCallerEncoder
	}
}

func parseNameEncoder(s string, cfg *zapcore.EncoderConfig) {
	switch strings.ToLower(s) {
	case "full":
		cfg.EncodeName = zapcore.FullNameEncoder
	}
}

// GetLevel returns the current global log level
func GetLevel() zapcore.Level {
	return atom.Level()
}

// SetLevel dynamically changes the global log level
// Returns error if the level string is invalid
func SetLevel(level string) error {
	newLevel := zap.NewAtomicLevel()
	err := newLevel.UnmarshalText([]byte(level))
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	atom.SetLevel(newLevel.Level())
	fmt.Printf("Log level changed to: %s\n", atom.Level().String()) // Log level change confirmation
	return nil
}

// AddLumberjackToViper adds default lumberjack config to viper defaults
// Call this before viper.ReadInConfig() if you want lumberjack defaults
func AddLumberjackToViper(v *viper.Viper) {
	v.SetDefault("logger.lumberjack.filename", "logs/app.log") // Default log file path
	v.SetDefault("logger.lumberjack.maxsize", 100)             // 100 MB
	v.SetDefault("logger.lumberjack.maxage", 30)               // 30 days
	v.SetDefault("logger.lumberjack.maxbackups", 5)
	v.SetDefault("logger.lumberjack.localtime", false)
	v.SetDefault("logger.lumberjack.compress", true)
}
