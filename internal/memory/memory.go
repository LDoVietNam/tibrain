package memory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
)

// DB is the minimal database surface required by the cognitive memory
// manager. *sql.DB (and root *Hub via delegation) satisfies this interface.
type DB interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// MemoryType identifies the category of a stored memory entry.
type MemoryType string

// MemorySemantic is the semantic memory classification.
const MemorySemantic MemoryType = "semantic"

// MemoryEntry is a single cognitive memory record. Field tags make the slice
// returned by QueryMemory directly JSON-marshalable by the root package.
type MemoryEntry struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
	Metadata  string `json:"metadata,omitempty"`
}

// Options configures a CognitiveMemoryManager.
type Options struct {
	MaxEntries int
}

// CognitiveMemoryManager stores and queries cognitive memory entries.
type CognitiveMemoryManager struct {
	db   DB
	opts *Options
}

// NewCognitiveMemoryManager builds a manager over the given DB. opts may be nil,
// in which case a default configuration is applied.
func NewCognitiveMemoryManager(db DB, opts *Options) (*CognitiveMemoryManager, error) {
	if db == nil {
		return nil, errors.New("memory: db must not be nil")
	}
	if opts == nil {
		opts = &Options{MaxEntries: 1000}
	}
	return &CognitiveMemoryManager{db: db, opts: opts}, nil
}

// QueryMemory returns memory entries of the given type, up to limit rows.
// If the backing table does not yet exist, it returns an empty slice without
// error. There is no panic path.
func (m *CognitiveMemoryManager) QueryMemory(ctx context.Context, query string, memType MemoryType, limit int) ([]MemoryEntry, error) {
	if limit <= 0 {
		limit = m.opts.MaxEntries
	}

	rows, err := m.db.QueryContext(ctx,
		`SELECT id, type, content, created_at, metadata
		 FROM cognitive_memory
		 WHERE type = ? AND content LIKE ?
		 ORDER BY created_at DESC
		 LIMIT ?`,
		string(memType), "%"+query+"%", limit)
	if err != nil {
		if strings.Contains(err.Error(), "no such table") {
			log.Printf("memory: cognitive_memory table not present yet, returning empty: %v", err)
			return []MemoryEntry{}, nil
		}
		return nil, fmt.Errorf("query cognitive memory: %w", err)
	}
	defer rows.Close()

	var entries []MemoryEntry
	for rows.Next() {
		var e MemoryEntry
		var metadata sql.NullString
		if err := rows.Scan(&e.ID, &e.Type, &e.Content, &e.CreatedAt, &metadata); err != nil {
			return nil, fmt.Errorf("scan cognitive memory row: %w", err)
		}
		e.Metadata = metadata.String
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cognitive memory rows: %w", err)
	}

	return entries, nil
}
