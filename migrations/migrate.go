// Package migrations provides database migration support for TiBrain
// to align with OmniRoute schema
package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed *.sql
var migrationFiles embed.FS

// Migration represents a single database migration
type Migration struct {
	Version string
	Name    string
	SQL     string
}

// Migrator handles database migrations
type Migrator struct {
	db         *sql.DB
	migrations []Migration
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB) (*Migrator, error) {
	m := &Migrator{db: db}

	// Load embedded migration files
	if err := m.loadMigrations(); err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	return m, nil
}

// loadMigrations loads SQL files from the embedded filesystem
func (m *Migrator) loadMigrations() error {
	entries, err := migrationFiles.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, err := migrationFiles.ReadFile(entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		// Parse version and name from filename: 001_initial_schema.sql
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) < 2 {
			log.Printf("Skipping migration with invalid filename: %s", entry.Name())
			continue
		}

		version := parts[0]
		name := strings.TrimSuffix(parts[1], ".sql")

		m.migrations = append(m.migrations, Migration{
			Version: version,
			Name:    name,
			SQL:     string(content),
		})
	}

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	return nil
}

// ensureMigrationTable creates the migration tracking table if it doesn't exist
func (m *Migrator) ensureMigrationTable() error {
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS _tiroute_migrations (
			version TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		)
	`)
	return err
}

// getAppliedMigrations returns a set of already applied migration versions
func (m *Migrator) getAppliedMigrations() (map[string]bool, error) {
	applied := make(map[string]bool)

	rows, err := m.db.Query("SELECT version FROM _tiroute_migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// Up applies all pending migrations
func (m *Migrator) Up() error {
	if err := m.ensureMigrationTable(); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, migration := range m.migrations {
		if applied[migration.Version] {
			log.Printf("Skipping already applied migration: %s_%s", migration.Version, migration.Name)
			continue
		}

		log.Printf("Applying migration: %s_%s", migration.Version, migration.Name)

		// Execute migration SQL
		if _, err := tx.Exec(migration.SQL); err != nil {
			return fmt.Errorf("failed to apply migration %s_%s: %w", migration.Version, migration.Name, err)
		}

		// Record migration
		if _, err := tx.Exec(
			"INSERT INTO _tiroute_migrations (version, name, applied_at) VALUES (?, ?, ?)",
			migration.Version, migration.Name, time.Now().Format(time.RFC3339),
		); err != nil {
			return fmt.Errorf("failed to record migration %s_%s: %w", migration.Version, migration.Name, err)
		}

		log.Printf("Successfully applied migration: %s_%s", migration.Version, migration.Name)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	return nil
}

// Status returns the current migration status
func (m *Migrator) Status() error {
	if err := m.ensureMigrationTable(); err != nil {
		return err
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	log.Println("Migration Status:")
	log.Println("=================")

	for _, migration := range m.migrations {
		status := "PENDING"
		if applied[migration.Version] {
			status = "APPLIED"
		}
		log.Printf("%s_%s: %s", migration.Version, migration.Name, status)
	}

	return nil
}

// RunMigrations is a convenience function to run migrations on a database
func RunMigrations(dbPath string) error {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&mode=rwc")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	migrator, err := NewMigrator(db)
	if err != nil {
		return err
	}

	if err := migrator.Up(); err != nil {
		return err
	}

	return nil
}
