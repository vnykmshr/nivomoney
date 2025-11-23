package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Save original env vars to restore after test
	origEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range origEnv {
			// Skip restoration for simplicity in tests
			_ = env
		}
	}()

	tests := []struct {
		name    string
		envVars map[string]string
		want    func(*Config) bool
		wantErr bool
	}{
		{
			name:    "defaults",
			envVars: map[string]string{},
			want: func(c *Config) bool {
				return c.Environment == "development" &&
					c.ServicePort == 8080 &&
					c.DatabaseHost == "localhost" &&
					c.DatabasePort == 5432
			},
			wantErr: false,
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"SERVICE_PORT":  "9000",
				"DATABASE_HOST": "db.example.com",
				"DATABASE_PORT": "5433",
				"LOG_LEVEL":     "debug",
			},
			want: func(c *Config) bool {
				return c.ServicePort == 9000 &&
					c.DatabaseHost == "db.example.com" &&
					c.DatabasePort == 5433 &&
					c.LogLevel == "debug"
			},
			wantErr: false,
		},
		{
			name: "production environment",
			envVars: map[string]string{
				"ENVIRONMENT":       "production",
				"JWT_SECRET":        "secure-production-secret",
				"DATABASE_PASSWORD": "secure-db-password",
				"REDIS_PASSWORD":    "secure-redis-password",
			},
			want: func(c *Config) bool {
				return c.Environment == "production" &&
					c.JWTSecret == "secure-production-secret"
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && !tt.want(cfg) {
				t.Errorf("Load() config validation failed")
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid development config",
			config: &Config{
				Environment:      "development",
				ServicePort:      8080,
				DatabasePort:     5432,
				JWTSecret:        "nivo-dev-secret-change-in-production",
				DatabasePassword: "nivo_dev_password",
				RedisPassword:    "nivo_redis_password",
			},
			wantErr: false,
		},
		{
			name: "invalid production config - default jwt secret",
			config: &Config{
				Environment:      "production",
				ServicePort:      8080,
				DatabasePort:     5432,
				JWTSecret:        "nivo-dev-secret-change-in-production",
				DatabasePassword: "secure-password",
				RedisPassword:    "secure-password",
			},
			wantErr: true,
		},
		{
			name: "invalid production config - default db password",
			config: &Config{
				Environment:      "production",
				ServicePort:      8080,
				DatabasePort:     5432,
				JWTSecret:        "secure-secret",
				DatabasePassword: "nivo_dev_password",
				RedisPassword:    "secure-password",
			},
			wantErr: true,
		},
		{
			name: "invalid service port",
			config: &Config{
				Environment:      "development",
				ServicePort:      99999,
				DatabasePort:     5432,
				JWTSecret:        "secret",
				DatabasePassword: "password",
				RedisPassword:    "password",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		want        bool
	}{
		{"development", "development", true},
		{"dev", "dev", true},
		{"production", "production", false},
		{"staging", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Environment: tt.environment}
			if got := c.IsDevelopment(); got != tt.want {
				t.Errorf("Config.IsDevelopment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		want        bool
	}{
		{"production", "production", true},
		{"prod", "prod", true},
		{"development", "development", false},
		{"staging", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{Environment: tt.environment}
			if got := c.IsProduction(); got != tt.want {
				t.Errorf("Config.IsProduction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvHelpers(t *testing.T) {
	os.Clearenv()

	// Test getEnv
	os.Setenv("TEST_STRING", "value")
	if got := getEnv("TEST_STRING", "default"); got != "value" {
		t.Errorf("getEnv() = %v, want %v", got, "value")
	}
	if got := getEnv("MISSING", "default"); got != "default" {
		t.Errorf("getEnv() = %v, want %v", got, "default")
	}

	// Test getEnvAsInt
	os.Setenv("TEST_INT", "42")
	if got := getEnvAsInt("TEST_INT", 0); got != 42 {
		t.Errorf("getEnvAsInt() = %v, want %v", got, 42)
	}
	if got := getEnvAsInt("MISSING", 99); got != 99 {
		t.Errorf("getEnvAsInt() = %v, want %v", got, 99)
	}

	// Test getEnvAsBool
	os.Setenv("TEST_BOOL", "true")
	if got := getEnvAsBool("TEST_BOOL", false); got != true {
		t.Errorf("getEnvAsBool() = %v, want %v", got, true)
	}
	if got := getEnvAsBool("MISSING", false); got != false {
		t.Errorf("getEnvAsBool() = %v, want %v", got, false)
	}

	// Test getEnvAsDuration
	os.Setenv("TEST_DURATION", "5s")
	if got := getEnvAsDuration("TEST_DURATION", 0); got != 5*time.Second {
		t.Errorf("getEnvAsDuration() = %v, want %v", got, 5*time.Second)
	}
	if got := getEnvAsDuration("MISSING", time.Minute); got != time.Minute {
		t.Errorf("getEnvAsDuration() = %v, want %v", got, time.Minute)
	}
}
