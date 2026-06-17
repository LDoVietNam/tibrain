package session

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Session struct {
	ID         string            `json:"id"`
	CreatedAt  int64             `json:"created_at"`
	UpdatedAt  int64             `json:"updated_at"`
	Provider   string            `json:"provider"`
	Model      string            `json:"model"`
	Messages   []Message         `json:"messages"`
	TokenCount int               `json:"token_count"`
	Phase      string            `json:"phase,omitempty"`
	Dir        string            `json:"dir,omitempty"`
	Title      string            `json:"title,omitempty"`
	Summary    string            `json:"summary,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	Project    string            `json:"project,omitempty"`
	Branch     string            `json:"branch,omitempty"`
	ParentID   string            `json:"parent_id,omitempty"`
	Archived   bool              `json:"archived,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type ListOptions struct {
	Limit           int
	IncludeArchived bool
	Project         string
	Phase           string
	Search          string
}

type StoreStats struct {
	Total          int `json:"total"`
	Archived       int `json:"archived"`
	Active         int `json:"active"`
	DistinctPhases int `json:"distinct_phases"`
}

type Store struct{ db *sql.DB }

func NewStore(path string) (*Store, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, ".ti", "data", "sessions.db")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite does not support concurrent writes
	s := &Store{db: db}
	if err := s.initDB(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }
func (s *Store) DB() *sql.DB  { return s.db }

func (s *Store) initDB() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		provider TEXT NOT NULL DEFAULT '',
		model TEXT NOT NULL DEFAULT '',
		messages TEXT NOT NULL DEFAULT '[]',
		token_count INTEGER NOT NULL DEFAULT 0,
		phase TEXT NOT NULL DEFAULT '',
		dir TEXT NOT NULL DEFAULT '',
		title TEXT NOT NULL DEFAULT '',
		summary TEXT NOT NULL DEFAULT '',
		tags TEXT NOT NULL DEFAULT '[]',
		project TEXT NOT NULL DEFAULT '',
		branch TEXT NOT NULL DEFAULT '',
		parent_id TEXT NOT NULL DEFAULT '',
		archived INTEGER NOT NULL DEFAULT 0,
		metadata TEXT NOT NULL DEFAULT '{}'
	);
	CREATE INDEX IF NOT EXISTS idx_sessions_updated ON sessions(updated_at DESC);
	CREATE INDEX IF NOT EXISTS idx_sessions_phase ON sessions(phase);
	CREATE INDEX IF NOT EXISTS idx_sessions_project ON sessions(project);
	CREATE INDEX IF NOT EXISTS idx_sessions_archived ON sessions(archived);
	CREATE INDEX IF NOT EXISTS idx_sessions_parent ON sessions(parent_id);`
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}
	return s.migrateColumns()
}

func (s *Store) migrateColumns() error {
	columns := map[string]string{
		"title":     "TEXT NOT NULL DEFAULT ''",
		"summary":   "TEXT NOT NULL DEFAULT ''",
		"tags":      "TEXT NOT NULL DEFAULT '[]'",
		"project":   "TEXT NOT NULL DEFAULT ''",
		"branch":    "TEXT NOT NULL DEFAULT ''",
		"parent_id": "TEXT NOT NULL DEFAULT ''",
		"archived":  "INTEGER NOT NULL DEFAULT 0",
		"metadata":  "TEXT NOT NULL DEFAULT '{}'",
	}
	existing := map[string]bool{}
	rows, err := s.db.Query(`PRAGMA table_info(sessions)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &dflt, &pk); err != nil {
			return err
		}
		existing[name] = true
	}
	for name, decl := range columns {
		if existing[name] {
			continue
		}
		if _, err := s.db.Exec(`ALTER TABLE sessions ADD COLUMN ` + name + ` ` + decl); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Create(session *Session) error {
	if session == nil {
		return fmt.Errorf("nil session")
	}
	if session.ID == "" {
		session.ID = generateID()
	}
	now := time.Now().Unix()
	if session.CreatedAt == 0 {
		session.CreatedAt = now
	}
	if session.UpdatedAt == 0 || session.UpdatedAt < session.CreatedAt {
		session.UpdatedAt = now
	}
	if session.Metadata == nil {
		session.Metadata = map[string]string{}
	}
	session.Tags = dedupeStrings(session.Tags)
	messagesJSON, _ := json.Marshal(session.Messages)
	tagsJSON, _ := json.Marshal(session.Tags)
	metadataJSON, _ := json.Marshal(session.Metadata)
	_, err := s.db.Exec(`INSERT OR REPLACE INTO sessions (id, created_at, updated_at, provider, model, messages, token_count, phase, dir, title, summary, tags, project, branch, parent_id, archived, metadata) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		session.ID, session.CreatedAt, session.UpdatedAt, session.Provider, session.Model, string(messagesJSON), session.TokenCount, session.Phase, session.Dir, session.Title, session.Summary, string(tagsJSON), session.Project, session.Branch, session.ParentID, boolToInt(session.Archived), string(metadataJSON))
	return err
}

func (s *Store) Get(id string) (*Session, error) {
	var sess Session
	var messagesJSON, tagsJSON, metadataJSON string
	var archivedInt int
	err := s.db.QueryRow(`SELECT id, created_at, updated_at, provider, model, messages, token_count, phase, dir, title, summary, tags, project, branch, parent_id, archived, metadata FROM sessions WHERE id = ?`, id).
		Scan(&sess.ID, &sess.CreatedAt, &sess.UpdatedAt, &sess.Provider, &sess.Model, &messagesJSON, &sess.TokenCount, &sess.Phase, &sess.Dir, &sess.Title, &sess.Summary, &tagsJSON, &sess.Project, &sess.Branch, &sess.ParentID, &archivedInt, &metadataJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	sess.Archived = archivedInt == 1
	if err := decodeSessionPayloads(&sess, messagesJSON, tagsJSON, metadataJSON); err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *Store) List(count int) ([]Session, error) { return s.ListAdvanced(ListOptions{Limit: count}) }

func (s *Store) ListAdvanced(opts ListOptions) ([]Session, error) {
	query := `SELECT id, created_at, updated_at, provider, model, messages, token_count, phase, dir, title, summary, tags, project, branch, parent_id, archived, metadata FROM sessions WHERE 1=1`
	args := []any{}
	if !opts.IncludeArchived {
		query += ` AND archived = 0`
	}
	if opts.Project != "" {
		query += ` AND project = ?`
		args = append(args, opts.Project)
	}
	if opts.Phase != "" {
		query += ` AND phase = ?`
		args = append(args, opts.Phase)
	}
	if opts.Search != "" {
		pattern := "%" + opts.Search + "%"
		query += ` AND (id LIKE ? OR model LIKE ? OR provider LIKE ? OR phase LIKE ? OR title LIKE ? OR summary LIKE ? OR messages LIKE ?)`
		args = append(args, pattern, pattern, pattern, pattern, pattern, pattern, pattern)
	}
	query += ` ORDER BY updated_at DESC`
	if opts.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, opts.Limit)
	}
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []Session
	for rows.Next() {
		var sess Session
		var messagesJSON, tagsJSON, metadataJSON string
		var archivedInt int
		if err := rows.Scan(&sess.ID, &sess.CreatedAt, &sess.UpdatedAt, &sess.Provider, &sess.Model, &messagesJSON, &sess.TokenCount, &sess.Phase, &sess.Dir, &sess.Title, &sess.Summary, &tagsJSON, &sess.Project, &sess.Branch, &sess.ParentID, &archivedInt, &metadataJSON); err != nil {
			return nil, err
		}
		sess.Archived = archivedInt == 1
		if err := decodeSessionPayloads(&sess, messagesJSON, tagsJSON, metadataJSON); err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

func (s *Store) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

func (s *Store) Export(id string) ([]byte, error) {
	sess, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, fmt.Errorf("session %q not found", id)
	}
	return json.MarshalIndent(sess, "", "  ")
}

func (s *Store) Import(data []byte) (*Session, error) {
	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	if err := s.Create(&sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *Store) AppendMessage(sessionID string, msg Message, tokens int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var messagesJSON string
	var tokenCount int
	err = tx.QueryRow(`SELECT messages, token_count FROM sessions WHERE id = ?`, sessionID).Scan(&messagesJSON, &tokenCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("session %q not found", sessionID)
		}
		return err
	}
	var messages []Message
	if err := json.Unmarshal([]byte(messagesJSON), &messages); err != nil {
		return err
	}
	messages = append(messages, msg)
	updatedMessages, _ := json.Marshal(messages)
	_, err = tx.Exec(`UPDATE sessions SET messages=?, token_count=?, updated_at=? WHERE id=?`, string(updatedMessages), tokenCount+tokens, time.Now().Unix(), sessionID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) UpdateMetadata(id string, update func(*Session) error) error {
	sess, err := s.Get(id)
	if err != nil {
		return err
	}
	if sess == nil {
		return fmt.Errorf("session %q not found", id)
	}
	if err := update(sess); err != nil {
		return err
	}
	sess.UpdatedAt = time.Now().Unix()
	return s.Create(sess)
}

func (s *Store) Fork(id, title string) (*Session, error) {
	sess, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, fmt.Errorf("session %q not found", id)
	}
	clone := *sess
	clone.ID = generateID()
	clone.ParentID = sess.ID
	clone.Archived = false
	clone.CreatedAt = time.Now().Unix()
	clone.UpdatedAt = clone.CreatedAt
	clone.Tags = append([]string(nil), sess.Tags...)
	clone.Metadata = copyMap(sess.Metadata)
	if strings.TrimSpace(title) != "" {
		clone.Title = title
	}
	if err := s.Create(&clone); err != nil {
		return nil, err
	}
	return &clone, nil
}

func (s *Store) Archive(id string, archived bool) error {
	_, err := s.db.Exec(`UPDATE sessions SET archived=?, updated_at=? WHERE id=?`, boolToInt(archived), time.Now().Unix(), id)
	return err
}

func (s *Store) Search(query string, limit int) ([]Session, error) {
	return s.ListAdvanced(ListOptions{Limit: limit, Search: query, IncludeArchived: true})
}

func (s *Store) Count() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&n)
	return n, err
}

func (s *Store) Stats() (*StoreStats, error) {
	stats := &StoreStats{}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&stats.Total); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE archived = 1`).Scan(&stats.Archived); err != nil {
		return nil, err
	}
	stats.Active = stats.Total - stats.Archived
	if err := s.db.QueryRow(`SELECT COUNT(DISTINCT phase) FROM sessions WHERE phase != ''`).Scan(&stats.DistinctPhases); err != nil {
		return nil, err
	}
	return stats, nil
}

func (s *Store) Cleanup(age time.Duration) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE updated_at < ?`, time.Now().Add(-age).Unix())
	return err
}

func decodeSessionPayloads(sess *Session, messagesJSON, tagsJSON, metadataJSON string) error {
	if messagesJSON == "" {
		messagesJSON = "[]"
	}
	if tagsJSON == "" {
		tagsJSON = "[]"
	}
	if metadataJSON == "" {
		metadataJSON = "{}"
	}
	if err := json.Unmarshal([]byte(messagesJSON), &sess.Messages); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(tagsJSON), &sess.Tags); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(metadataJSON), &sess.Metadata); err != nil {
		return err
	}
	if sess.Metadata == nil {
		sess.Metadata = map[string]string{}
	}
	return nil
}

func dedupeStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func copyMap(in map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	hexID := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", hexID[0:8], hexID[8:12], hexID[12:16], hexID[16:20], hexID[20:32])
}
