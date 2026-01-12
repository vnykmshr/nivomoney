// Package logger provides structured logging for Nivo services using zerolog.
package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// RequestIDKey is the context key for request IDs.
	RequestIDKey contextKey = "request_id"
	// UserIDKey is the context key for user IDs.
	UserIDKey contextKey = "user_id"
	// CorrelationIDKey is the context key for correlation IDs.
	CorrelationIDKey contextKey = "correlation_id"
)

// Logger wraps zerolog.Logger with additional functionality.
type Logger struct {
	logger zerolog.Logger
}

// Config holds logger configuration.
type Config struct {
	Level       string // debug, info, warn, error
	Format      string // console, json
	ServiceName string
	Output      io.Writer
}

// New creates a new Logger instance with the given configuration.
func New(cfg Config) *Logger {
	// Set output writer
	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	// Configure zerolog based on format
	var zlog zerolog.Logger
	if cfg.Format == "console" {
		// Human-readable console output for development
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	zlog = zerolog.New(output).With().
		Timestamp().
		Str("service", cfg.ServiceName).
		Logger()

	// Set log level
	level := parseLevel(cfg.Level)
	zlog = zlog.Level(level)

	return &Logger{
		logger: zlog,
	}
}

// NewDefault creates a logger with default configuration.
func NewDefault(serviceName string) *Logger {
	return New(Config{
		Level:       "info",
		Format:      "console",
		ServiceName: serviceName,
	})
}

// NewFromEnv creates a logger based on environment variables.
// Uses LOG_LEVEL (default: info) and LOG_FORMAT (default: json in production, console otherwise).
// Environment is determined by ENV or ENVIRONMENT variable.
func NewFromEnv(serviceName string) *Logger {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	format := os.Getenv("LOG_FORMAT")
	if format == "" {
		// Default to JSON in production, console in development
		env := os.Getenv("ENV")
		if env == "" {
			env = os.Getenv("ENVIRONMENT")
		}
		if env == "production" || env == "prod" {
			format = "json"
		} else {
			format = "console"
		}
	}

	return New(Config{
		Level:       level,
		Format:      format,
		ServiceName: serviceName,
	})
}

// parseLevel converts string level to zerolog.Level.
func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// WithContext returns a new logger with context values added.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.logger

	// Add request ID if present
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		logger = logger.With().Str("request_id", requestID.(string)).Logger()
	}

	// Add user ID if present
	if userID := ctx.Value(UserIDKey); userID != nil {
		logger = logger.With().Str("user_id", userID.(string)).Logger()
	}

	// Add correlation ID if present
	if correlationID := ctx.Value(CorrelationIDKey); correlationID != nil {
		logger = logger.With().Str("correlation_id", correlationID.(string)).Logger()
	}

	return &Logger{logger: logger}
}

// With returns a new logger with additional fields.
func (l *Logger) With(fields map[string]interface{}) *Logger {
	logger := l.logger.With()
	for k, v := range fields {
		logger = logger.Interface(k, v)
	}
	return &Logger{logger: logger.Logger()}
}

// WithField returns a new logger with a single field added.
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

// WithError returns a new logger with error field added.
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}
	return &Logger{
		logger: l.logger.With().Err(err).Logger(),
	}
}

// Debug logs a debug level message.
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug level message.
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// Info logs an info level message.
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted info level message.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// Warn logs a warning level message.
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning level message.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// Error logs an error level message.
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs a formatted error level message.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

// Fatal logs a fatal level message and exits.
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal level message and exits.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// GetZerologLogger returns the underlying zerolog.Logger for advanced usage.
func (l *Logger) GetZerologLogger() zerolog.Logger {
	return l.logger
}

// Global logger instance (optional, for convenience).
var defaultLogger *Logger

// InitGlobal initializes the global logger instance.
func InitGlobal(cfg Config) {
	defaultLogger = New(cfg)
}

// Global returns the global logger instance.
func Global() *Logger {
	if defaultLogger == nil {
		defaultLogger = NewDefault("nivo")
	}
	return defaultLogger
}

// Debug logs a debug message using the global logger.
func Debug(msg string) {
	Global().Debug(msg)
}

// Info logs an info message using the global logger.
func Info(msg string) {
	Global().Info(msg)
}

// Warn logs a warning message using the global logger.
func Warn(msg string) {
	Global().Warn(msg)
}

// Error logs an error message using the global logger.
func Error(msg string) {
	Global().Error(msg)
}

// Fatal logs a fatal message using the global logger and exits.
func Fatal(msg string) {
	Global().Fatal(msg)
}
