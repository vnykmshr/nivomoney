// Package config provides centralized configuration management for Nivo services.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	// Application
	Environment string
	ServiceName string
	ServicePort int
	LogLevel    string

	// Localization (India-centric defaults)
	Timezone        string // Default: Asia/Kolkata (IST - UTC+5:30)
	DefaultCurrency string // Default: INR (Indian Rupee)
	CountryCode     string // Default: IN (India)

	// Database
	DatabaseURL      string
	DatabaseHost     string
	DatabasePort     int
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	DatabaseSSLMode  string

	// Redis
	RedisURL      string
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int

	// NSQ
	NSQLookupDAddr string
	NSQDAddr       string

	// JWT
	JWTSecret     string
	JWTExpiry     time.Duration
	JWTRefreshExp time.Duration

	// Server
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	// Observability
	PrometheusPort  int
	EnableProfiling bool
}

// Load loads configuration from environment variables with defaults.
func Load() (*Config, error) {
	cfg := &Config{
		// Application defaults
		Environment: getEnv("ENVIRONMENT", "development"),
		ServiceName: getEnv("SERVICE_NAME", "nivo"),
		ServicePort: getEnvAsInt("SERVICE_PORT", 8080),
		LogLevel:    getEnv("LOG_LEVEL", "info"),

		// Localization defaults (India-centric)
		Timezone:        getEnv("TIMEZONE", "Asia/Kolkata"), // IST (UTC+5:30)
		DefaultCurrency: getEnv("DEFAULT_CURRENCY", "INR"),  // Indian Rupee
		CountryCode:     getEnv("COUNTRY_CODE", "IN"),       // India

		// Database defaults
		DatabaseHost:     getEnv("DATABASE_HOST", "localhost"),
		DatabasePort:     getEnvAsInt("DATABASE_PORT", 5432),
		DatabaseUser:     getEnv("DATABASE_USER", "nivo"),
		DatabasePassword: getEnv("DATABASE_PASSWORD", "nivo_dev_password"),
		DatabaseName:     getEnv("DATABASE_NAME", "nivo"),
		DatabaseSSLMode:  getEnv("DATABASE_SSL_MODE", "disable"),

		// Redis defaults
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnvAsInt("REDIS_PORT", 6379),
		RedisPassword: getEnv("REDIS_PASSWORD", "nivo_redis_password"),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// NSQ defaults
		NSQLookupDAddr: getEnv("NSQLOOKUPD_ADDR", "localhost:4161"),
		NSQDAddr:       getEnv("NSQD_ADDR", "localhost:4150"),

		// JWT defaults
		JWTSecret:     getEnv("JWT_SECRET", "nivo-dev-secret-change-in-production"),
		JWTExpiry:     getEnvAsDuration("JWT_EXPIRY", 24*time.Hour),
		JWTRefreshExp: getEnvAsDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),

		// Server defaults
		ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
		WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),

		// Observability defaults
		PrometheusPort:  getEnvAsInt("PROMETHEUS_PORT", 9090),
		EnableProfiling: getEnvAsBool("ENABLE_PROFILING", false),
	}

	// Construct composite URLs
	cfg.DatabaseURL = getEnv("DATABASE_URL",
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.DatabaseUser,
			cfg.DatabasePassword,
			cfg.DatabaseHost,
			cfg.DatabasePort,
			cfg.DatabaseName,
			cfg.DatabaseSSLMode,
		))

	cfg.RedisURL = getEnv("REDIS_URL",
		fmt.Sprintf("redis://:%s@%s:%d/%d",
			cfg.RedisPassword,
			cfg.RedisHost,
			cfg.RedisPort,
			cfg.RedisDB,
		))

	// Validate required fields in production
	if cfg.Environment == "production" {
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("configuration validation failed: %w", err)
		}
	}

	return cfg, nil
}

// Validate ensures critical configuration values are set properly.
func (c *Config) Validate() error {
	if c.Environment == "production" {
		if c.JWTSecret == "nivo-dev-secret-change-in-production" {
			return fmt.Errorf("JWT_SECRET must be changed in production")
		}

		if c.DatabasePassword == "nivo_dev_password" {
			return fmt.Errorf("DATABASE_PASSWORD must be changed in production")
		}

		if c.RedisPassword == "nivo_redis_password" {
			return fmt.Errorf("REDIS_PASSWORD must be changed in production")
		}
	}

	if c.ServicePort < 1 || c.ServicePort > 65535 {
		return fmt.Errorf("SERVICE_PORT must be between 1 and 65535")
	}

	if c.DatabasePort < 1 || c.DatabasePort > 65535 {
		return fmt.Errorf("DATABASE_PORT must be between 1 and 65535")
	}

	return nil
}

// IsDevelopment returns true if running in development environment.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development" || c.Environment == "dev"
}

// IsProduction returns true if running in production environment.
func (c *Config) IsProduction() bool {
	return c.Environment == "production" || c.Environment == "prod"
}

// Helper functions for reading environment variables

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
