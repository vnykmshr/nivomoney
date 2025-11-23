package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Migrator handles database migrations.
type Migrator struct {
	db              *sql.DB
	migrationsDir   string
	migrationsTable string
}

// NewMigrator creates a new migrator.
func NewMigrator(db *sql.DB, migrationsDir string) *Migrator {
	return &Migrator{
		db:              db,
		migrationsDir:   migrationsDir,
		migrationsTable: "schema_migrations",
	}
}

// Up runs all pending migrations.
func (m *Migrator) Up() error {
	// Create migrations table if not exists
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	files, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, file := range files {
		if _, ok := applied[file]; ok {
			continue // Already applied
		}

		if err := m.runMigration(file); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file, err)
		}
	}

	return nil
}

// createMigrationsTable creates the schema_migrations table.
func (m *Migrator) createMigrationsTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`, m.migrationsTable)

	_, err := m.db.Exec(query)
	return err
}

// getMigrationFiles returns a sorted list of migration files.
func (m *Migrator) getMigrationFiles() ([]string, error) {
	entries, err := os.ReadDir(m.migrationsDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Only include .up.sql files
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		files = append(files, name)
	}

	// Sort files to ensure consistent ordering
	sort.Strings(files)

	return files, nil
}

// getAppliedMigrations returns a map of applied migration versions.
func (m *Migrator) getAppliedMigrations() (map[string]bool, error) {
	query := fmt.Sprintf("SELECT version FROM %s", m.migrationsTable)
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// runMigration runs a single migration file.
func (m *Migrator) runMigration(filename string) error {
	// Read migration file
	path := filepath.Join(m.migrationsDir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Begin transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration
	query := fmt.Sprintf("INSERT INTO %s (version) VALUES ($1)", m.migrationsTable)
	if _, err := tx.Exec(query, filename); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// Down rolls back the last migration (not implemented for safety).
func (m *Migrator) Down() error {
	return fmt.Errorf("down migrations not implemented - use manual rollback for safety")
}
