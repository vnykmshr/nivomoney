package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "console format",
			config: Config{
				Level:       "info",
				Format:      "console",
				ServiceName: "test-service",
			},
		},
		{
			name: "json format",
			config: Config{
				Level:       "debug",
				Format:      "json",
				ServiceName: "test-service",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.config.Output = &buf

			logger := New(tt.config)
			if logger == nil {
				t.Error("New() returned nil")
			}

			logger.Info("test message")
			output := buf.String()

			if output == "" {
				t.Error("Logger produced no output")
			}
		})
	}
}

func TestNewDefault(t *testing.T) {
	logger := NewDefault("test-service")
	if logger == nil {
		t.Error("NewDefault() returned nil")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"error", "error"},
		{"fatal", "fatal"},
		{"invalid", "info"}, // defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := parseLevel(tt.input)
			if level.String() != tt.expected {
				t.Errorf("parseLevel(%s) = %s, want %s", tt.input, level.String(), tt.expected)
			}
		})
	}
}

func TestLogger_WithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test",
		Output:      &buf,
	})

	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIDKey, "req-123")
	ctx = context.WithValue(ctx, UserIDKey, "user-456")
	ctx = context.WithValue(ctx, CorrelationIDKey, "corr-789")

	contextLogger := logger.WithContext(ctx)
	contextLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "req-123") {
		t.Error("Output missing request_id")
	}
	if !strings.Contains(output, "user-456") {
		t.Error("Output missing user_id")
	}
	if !strings.Contains(output, "corr-789") {
		t.Error("Output missing correlation_id")
	}
}

func TestLogger_With(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test",
		Output:      &buf,
	})

	fields := map[string]interface{}{
		"transaction_id": "tx-123",
		"amount":         100.50,
		"currency":       "USD",
	}

	fieldLogger := logger.With(fields)
	fieldLogger.Info("transaction processed")

	output := buf.String()
	if !strings.Contains(output, "tx-123") {
		t.Error("Output missing transaction_id")
	}
	if !strings.Contains(output, "100.5") {
		t.Error("Output missing amount")
	}
	if !strings.Contains(output, "USD") {
		t.Error("Output missing currency")
	}
}

func TestLogger_WithField(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test",
		Output:      &buf,
	})

	fieldLogger := logger.WithField("user_id", "user-123")
	fieldLogger.Info("user action")

	output := buf.String()
	if !strings.Contains(output, "user-123") {
		t.Error("Output missing user_id field")
	}
}

func TestLogger_WithError(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test",
		Output:      &buf,
	})

	err := errors.New("something went wrong")
	errorLogger := logger.WithError(err)
	errorLogger.Error("operation failed")

	output := buf.String()
	if !strings.Contains(output, "something went wrong") {
		t.Error("Output missing error message")
	}
}

func TestLogger_WithError_Nil(t *testing.T) {
	logger := NewDefault("test")
	nilLogger := logger.WithError(nil)

	if nilLogger != logger {
		t.Error("WithError(nil) should return same logger instance")
	}
}

func TestLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		logFunc   func(*Logger)
		shouldLog bool
	}{
		{
			name:      "debug logged at debug level",
			logLevel:  "debug",
			logFunc:   func(l *Logger) { l.Debug("debug message") },
			shouldLog: true,
		},
		{
			name:      "debug not logged at info level",
			logLevel:  "info",
			logFunc:   func(l *Logger) { l.Debug("debug message") },
			shouldLog: false,
		},
		{
			name:      "info logged at info level",
			logLevel:  "info",
			logFunc:   func(l *Logger) { l.Info("info message") },
			shouldLog: true,
		},
		{
			name:      "warn logged at warn level",
			logLevel:  "warn",
			logFunc:   func(l *Logger) { l.Warn("warn message") },
			shouldLog: true,
		},
		{
			name:      "error logged at error level",
			logLevel:  "error",
			logFunc:   func(l *Logger) { l.Error("error message") },
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(Config{
				Level:       tt.logLevel,
				Format:      "json",
				ServiceName: "test",
				Output:      &buf,
			})

			tt.logFunc(logger)

			output := buf.String()
			if tt.shouldLog && output == "" {
				t.Errorf("Expected log output but got none")
			}
			if !tt.shouldLog && output != "" {
				t.Errorf("Expected no log output but got: %s", output)
			}
		})
	}
}

func TestLogger_FormattedLogs(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test",
		Output:      &buf,
	})

	logger.Infof("user %s performed action %s", "john", "login")

	output := buf.String()
	if !strings.Contains(output, "john") || !strings.Contains(output, "login") {
		t.Error("Formatted log message not found in output")
	}
}

func TestLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test-service",
		Output:      &buf,
	})

	logger.Info("test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	if logEntry["service"] != "test-service" {
		t.Errorf("Expected service=test-service, got %v", logEntry["service"])
	}

	if logEntry["message"] != "test message" {
		t.Errorf("Expected message='test message', got %v", logEntry["message"])
	}
}

func TestGlobalLogger(t *testing.T) {
	// Initialize global logger
	var buf bytes.Buffer
	InitGlobal(Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "global-test",
		Output:      &buf,
	})

	// Use global logger functions
	Info("global info message")

	output := buf.String()
	if !strings.Contains(output, "global info message") {
		t.Error("Global logger info message not found")
	}

	// Test Global() returns the same instance
	logger1 := Global()
	logger2 := Global()
	if logger1 != logger2 {
		t.Error("Global() should return the same instance")
	}
}

func TestLogger_GetZerologLogger(t *testing.T) {
	logger := NewDefault("test")
	zlog := logger.GetZerologLogger()

	// Just verify we can get the underlying logger
	if zlog.GetLevel() < 0 {
		t.Error("GetZerologLogger() returned invalid logger")
	}
}
