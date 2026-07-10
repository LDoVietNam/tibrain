package db

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Hub wraps a SQLite database connection used as the central storage hub.
type Hub struct {
	db *sql.DB
}

// NewHub opens (or creates) the TiBrain SQLite database inside dataDir and
// applies the connection PRAGMAs required for safe concurrent access.
func NewHub(dataDir string) (*Hub, error) {
	path := filepath.Join(dataDir, "tibrain.db")

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite at %s: %w", path, err)
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA busy_timeout=5000;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA foreign_keys=ON;",
	}
	for _, p := range pragmas {
		if _, err := conn.Exec(p); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("apply pragma %q: %w", p, err)
		}
	}

	log.Printf("db: opened hub at %s", path)
	return &Hub{db: conn}, nil
}

// DB returns the underlying *sql.DB handle.
func (h *Hub) DB() *sql.DB {
	return h.db
}

// Close closes the underlying database connection.
func (h *Hub) Close() error {
	return h.db.Close()
}
