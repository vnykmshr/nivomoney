// Package database provides PostgreSQL connection management and utilities for Nivo services.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Config holds database configuration.
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	ConnectTimeout  time.Duration
}

// DB wraps sql.DB with additional functionality.
type DB struct {
	*sql.DB
	config Config
}

// Connect establishes a connection to PostgreSQL with the given configuration.
func Connect(cfg Config) (*DB, error) {
	// Build connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
		int(cfg.ConnectTimeout.Seconds()),
	)

	// Open database connection
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	db := &DB{
		DB:     sqlDB,
		config: cfg,
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// NewFromURL creates a database connection from a connection URL.
func NewFromURL(url string) (*DB, error) {
	sqlDB, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{
		DB: sqlDB,
	}

	// Verify connection with default timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// HealthCheck performs a health check on the database connection.
func (db *DB) HealthCheck(ctx context.Context) error {
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check if we can execute a simple query
	var result int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	return nil
}

// Stats returns connection pool statistics.
func (db *DB) Stats() sql.DBStats {
	return db.DB.Stats()
}

// Transaction executes a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
func (db *DB) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Execute function
	if err := fn(tx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TransactionWithOptions executes a function within a database transaction with options.
func (db *DB) TransactionWithOptions(ctx context.Context, opts *sql.TxOptions, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}

// Helper functions for common operations

// QueryRowContext is a convenience wrapper around sql.DB.QueryRowContext.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(ctx, query, args...)
}

// QueryContext is a convenience wrapper around sql.DB.QueryContext.
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, query, args...)
}

// ExecContext is a convenience wrapper around sql.DB.ExecContext.
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}

// DefaultConfig returns a default database configuration suitable for development.
func DefaultConfig() Config {
	return Config{
		Host:            "localhost",
		Port:            5432,
		User:            "nivo",
		Password:        "nivo_dev_password",
		Database:        "nivo",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
		ConnectTimeout:  5 * time.Second,
	}
}

// ConfigFromEnv creates a Config from a config package Config.
// This bridges the shared/config and shared/database packages.
func ConfigFromEnv(host string, port int, user, password, database, sslMode string) Config {
	return Config{
		Host:            host,
		Port:            port,
		User:            user,
		Password:        password,
		Database:        database,
		SSLMode:         sslMode,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
		ConnectTimeout:  5 * time.Second,
	}
}

// MustConnect is like Connect but panics on error. Useful for initialization.
func MustConnect(cfg Config) *DB {
	db, err := Connect(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}
	return db
}

// Error helpers for PostgreSQL

// IsUniqueViolation checks if an error is a unique constraint violation (23505).
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL unique violation error code is 23505
	return err.Error() == "pq: duplicate key value violates unique constraint" ||
		contains(err.Error(), "duplicate key")
}

// IsForeignKeyViolation checks if an error is a foreign key constraint violation (23503).
func IsForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL foreign key violation error code is 23503
	return contains(err.Error(), "foreign key constraint")
}

// IsCheckViolation checks if an error is a check constraint violation (23514).
func IsCheckViolation(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL check constraint violation error code is 23514
	return contains(err.Error(), "check constraint")
}

// contains checks if a string contains a substring (case-insensitive helper).
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				anySubstring(s, substr)))
}

func anySubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
