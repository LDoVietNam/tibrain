package db

import (
	"database/sql"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

type InternalHub struct {
	db *sql.DB
}

func NewHub(dataDir string) (*InternalHub, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", filepath.Join(dataDir, "tibrain.db"))
	if err != nil {
		return nil, err
	}

	// Configure SQLite for single-connection mode (per AGENTS.md)
	db.SetMaxOpenConns(1)

	return &InternalHub{db: db}, nil
}

func (h *InternalHub) DB() *sql.DB {
	return h.db
}

func (h *InternalHub) Close() error {
	return h.db.Close()
}

// ApplyMigrations runs all migration files in the migrations directory
func ApplyMigrations(db *sql.DB) error {
	migrationsDir := "internal/db/migrations"

	// Check if directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// No migrations directory - nothing to apply
		return nil
	}

	// Get all .sql files in migrations directory
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	// Apply each migration file
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			return err
		}

		// Execute migration
		if _, err := db.Exec(string(content)); err != nil {
			return err
		}
	}

	return nil
}