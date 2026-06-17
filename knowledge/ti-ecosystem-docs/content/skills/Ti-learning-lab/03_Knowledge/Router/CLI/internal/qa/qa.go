// Package qa implements a Q&A pair store for Brain learning.
// Stores high-quality Q&A pairs from Claude responses with search,
// usage tracking, and quality filtering.
package qa

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// Entry is a single Q&A pair.
type Entry struct {
	ID        string  `json:"id"`
	Question  string  `json:"question"`
	Answer    string  `json:"answer"`
	Model     string  `json:"model"`
	Quality   float64 `json:"quality"`
	TaskType  string  `json:"task_type"`
	AgentID   string  `json:"agent_id"`
	CreatedAt int64   `json:"created_at"`
	UsedCount int64   `json:"used_count"`
}

// Stats holds aggregate statistics.
type Stats struct {
	Total      int64            `json:"total"`
	AvgQuality float64          `json:"avg_quality"`
	ByTaskType map[string]int64 `json:"by_task_type"`
	ByModel    map[string]int64 `json:"by_model"`
	TopUsed    []TopEntry       `json:"top_used"`
}

// TopEntry is a frequently used pair.
type TopEntry struct {
	ID        string `json:"id"`
	Question  string `json:"question"`
	UsedCount int64  `json:"used_count"`
}

// Config for QA store.
type Config struct {
	Path        string  // SQLite file path (default: ~/.ti/data/qa-pairs.db)
	MinQuality  float64 // Minimum quality to store (default: 0.7)
	SearchLimit int     // Default search limit (default: 20)
}

// Store manages Q&A pairs with search and usage tracking.
type Store struct {
	db     *sql.DB
	config Config
	mu     sync.RWMutex
}

// NewStore creates or opens a QA pair store.
func NewStore(cfg Config) (*Store, error) {
	if cfg.Path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		cfg.Path = filepath.Join(home, ".ti", "data", "qa-pairs.db")
	}
	if cfg.MinQuality == 0 {
		cfg.MinQuality = 0.7
	}
	if cfg.SearchLimit == 0 {
		cfg.SearchLimit = 20
	}

	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", cfg.Path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}

	s := &Store{db: db, config: cfg}
	if err := s.initDB(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// Close the store.
func (s *Store) Close() error {
	return s.db.Close()
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return hex.EncodeToString(b)
}

func (s *Store) initDB() error {
	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS qa_pairs (
		id         TEXT PRIMARY KEY,
		question   TEXT NOT NULL,
		answer     TEXT NOT NULL,
		model      TEXT NOT NULL DEFAULT '',
		quality    REAL NOT NULL DEFAULT 0.7,
		task_type  TEXT NOT NULL DEFAULT 'other',
		agent_id   TEXT NOT NULL DEFAULT 'default',
		created_at INTEGER NOT NULL,
		used_count INTEGER NOT NULL DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_qa_question ON qa_pairs(question);
	CREATE INDEX IF NOT EXISTS idx_qa_agent ON qa_pairs(agent_id);
	CREATE INDEX IF NOT EXISTS idx_qa_task_type ON qa_pairs(task_type);
	CREATE INDEX IF NOT EXISTS idx_qa_quality ON qa_pairs(quality);
	`)
	return err
}

// Store saves a new Q&A pair. Returns error if quality < MinQuality.
func (s *Store) Store(entry Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry.Quality < s.config.MinQuality {
		return fmt.Errorf("qa: quality %.3f below minimum %.3f", entry.Quality, s.config.MinQuality)
	}
	if entry.ID == "" {
		entry.ID = generateID()
	}
	if entry.CreatedAt == 0 {
		entry.CreatedAt = time.Now().Unix()
	}
	if entry.Model == "" {
		entry.Model = "unknown"
	}
	if entry.TaskType == "" {
		entry.TaskType = "other"
	}
	if entry.AgentID == "" {
		entry.AgentID = "default"
	}

	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO qa_pairs (id, question, answer, model, quality, task_type, agent_id, created_at, used_count)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.Question, entry.Answer, entry.Model, entry.Quality,
		entry.TaskType, entry.AgentID, entry.CreatedAt, entry.UsedCount,
	)
	return err
}

// Search queries Q&A pairs by text matching on question + answer.
// Returns results sorted by quality DESC, then used_count DESC.
func (s *Store) Search(query string, limit int) ([]Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = s.config.SearchLimit
	}

	pattern := "%" + query + "%"
	rows, err := s.db.Query(
		`SELECT id, question, answer, model, quality, task_type, agent_id, created_at, used_count
		 FROM qa_pairs
		 WHERE question LIKE ? OR answer LIKE ?
		 ORDER BY quality DESC, used_count DESC, created_at DESC
		 LIMIT ?`,
		pattern, pattern, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Question, &e.Answer, &e.Model, &e.Quality, &e.TaskType, &e.AgentID, &e.CreatedAt, &e.UsedCount); err != nil {
			return nil, err
		}
		results = append(results, e)
	}
	return results, nil
}

// IncrementUsed atomically increments the usage counter for a pair.
func (s *Store) IncrementUsed(id string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`UPDATE qa_pairs SET used_count = used_count + 1 WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	var count int64
	if err := s.db.QueryRow(`SELECT used_count FROM qa_pairs WHERE id = ?`, id).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Delete removes a Q&A pair by ID.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`DELETE FROM qa_pairs WHERE id = ?`, id)
	return err
}

// Count returns total Q&A pairs.
func (s *Store) Count() (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var n int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM qa_pairs`).Scan(&n)
	return n, err
}

// Stats returns aggregate statistics about stored pairs.
func (s *Store) Stats() (*Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &Stats{
		ByTaskType: make(map[string]int64),
		ByModel:    make(map[string]int64),
	}

	if err := s.db.QueryRow(`SELECT COUNT(*) FROM qa_pairs`).Scan(&stats.Total); err != nil {
		return nil, err
	}
	if stats.Total == 0 {
		return stats, nil
	}

	if err := s.db.QueryRow(`SELECT AVG(quality) FROM qa_pairs`).Scan(&stats.AvgQuality); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`SELECT task_type, COUNT(*) FROM qa_pairs GROUP BY task_type`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var tt string
		var c int64
		if err := rows.Scan(&tt, &c); err != nil {
			rows.Close()
			return nil, err
		}
		stats.ByTaskType[tt] = c
	}
	rows.Close()

	rows, err = s.db.Query(`SELECT model, COUNT(*) FROM qa_pairs GROUP BY model`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var m string
		var c int64
		if err := rows.Scan(&m, &c); err != nil {
			rows.Close()
			return nil, err
		}
		stats.ByModel[m] = c
	}
	rows.Close()

	rows, err = s.db.Query(`SELECT id, question, used_count FROM qa_pairs ORDER BY used_count DESC LIMIT 10`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var te TopEntry
		if err := rows.Scan(&te.ID, &te.Question, &te.UsedCount); err != nil {
			return nil, err
		}
		stats.TopUsed = append(stats.TopUsed, te)
	}

	return stats, nil
}

// Get retrieves a single Q&A pair by ID.
func (s *Store) Get(id string) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var e Entry
	err := s.db.QueryRow(
		`SELECT id, question, answer, model, quality, task_type, agent_id, created_at, used_count
		 FROM qa_pairs WHERE id = ?`,
		id,
	).Scan(&e.ID, &e.Question, &e.Answer, &e.Model, &e.Quality, &e.TaskType, &e.AgentID, &e.CreatedAt, &e.UsedCount)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// FilterByQuality returns pairs with quality in [min, max].
func (s *Store) FilterByQuality(min, max float64) ([]Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		`SELECT id, question, answer, model, quality, task_type, agent_id, created_at, used_count
		 FROM qa_pairs WHERE quality >= ? AND quality <= ? ORDER BY quality DESC`,
		min, max,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Question, &e.Answer, &e.Model, &e.Quality, &e.TaskType, &e.AgentID, &e.CreatedAt, &e.UsedCount); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// FilterByAgent returns pairs for a specific agent.
func (s *Store) FilterByAgent(agentID string) ([]Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		`SELECT id, question, answer, model, quality, task_type, agent_id, created_at, used_count
		 FROM qa_pairs WHERE agent_id = ? ORDER BY used_count DESC, quality DESC`,
		agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Question, &e.Answer, &e.Model, &e.Quality, &e.TaskType, &e.AgentID, &e.CreatedAt, &e.UsedCount); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// FilterByTaskType returns pairs for a specific task type.
func (s *Store) FilterByTaskType(taskType string) ([]Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		`SELECT id, question, answer, model, quality, task_type, agent_id, created_at, used_count
		 FROM qa_pairs WHERE task_type = ? ORDER BY quality DESC, used_count DESC`,
		taskType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Question, &e.Answer, &e.Model, &e.Quality, &e.TaskType, &e.AgentID, &e.CreatedAt, &e.UsedCount); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// TopUsed returns the N most-used pairs.
func (s *Store) TopUsed(n int) ([]TopEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(
		`SELECT id, question, used_count FROM qa_pairs ORDER BY used_count DESC LIMIT ?`,
		n,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []TopEntry
	for rows.Next() {
		var te TopEntry
		if err := rows.Scan(&te.ID, &te.Question, &te.UsedCount); err != nil {
			return nil, err
		}
		entries = append(entries, te)
	}
	return entries, nil
}

// ResetUsedCount resets all usage counters to zero.
func (s *Store) ResetUsedCount() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE qa_pairs SET used_count = 0`)
	return err
}

// DeleteByAgent removes all pairs for a specific agent.
func (s *Store) DeleteByAgent(agentID string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	res, err := s.db.Exec(`DELETE FROM qa_pairs WHERE agent_id = ?`, agentID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// SearchWithPrefix returns pairs whose question starts with the prefix.
func (s *Store) SearchWithPrefix(prefix string, limit int) ([]Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = s.config.SearchLimit
	}

	rows, err := s.db.Query(
		`SELECT id, question, answer, model, quality, task_type, agent_id, created_at, used_count
		 FROM qa_pairs WHERE question LIKE ? ORDER BY quality DESC, used_count DESC LIMIT ?`,
		prefix+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Question, &e.Answer, &e.Model, &e.Quality, &e.TaskType, &e.AgentID, &e.CreatedAt, &e.UsedCount); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// ContainsSubstring checks if any pair contains the given substring.
func (s *Store) ContainsSubstring(sub string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pattern := "%" + sub + "%"
	var exists bool
	err := s.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM qa_pairs WHERE question LIKE ? OR answer LIKE ? LIMIT 1)`,
		pattern, pattern,
	).Scan(&exists)
	return exists, err
}

// HasID checks if a pair with the given ID exists.
func (s *Store) HasID(id string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM qa_pairs WHERE id = ? LIMIT 1)`, id).Scan(&exists)
	return exists, err
}

// UpdateQuality changes the quality score of a pair.
func (s *Store) UpdateQuality(id string, newQuality float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE qa_pairs SET quality = ? WHERE id = ?`, newQuality, id)
	return err
}

// UpdateTaskType changes the task type of a pair.
func (s *Store) UpdateTaskType(id string, newTaskType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE qa_pairs SET task_type = ? WHERE id = ?`, newTaskType, id)
	return err
}

// BatchStore saves multiple entries atomically.
// Validates all entries before inserting any.
func (s *Store) BatchStore(entries []Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	for i := range entries {
		if entries[i].Quality < s.config.MinQuality {
			return fmt.Errorf("qa: entry %d quality %.3f below minimum", i, entries[i].Quality)
		}
	}

	for i := range entries {
		e := &entries[i]
		if e.ID == "" {
			e.ID = generateID()
		}
		if e.CreatedAt == 0 {
			e.CreatedAt = now
		}
		if e.Model == "" {
			e.Model = "unknown"
		}
		if e.TaskType == "" {
			e.TaskType = "other"
		}
		if e.AgentID == "" {
			e.AgentID = "default"
		}

		_, err := s.db.Exec(
			`INSERT OR REPLACE INTO qa_pairs (id, question, answer, model, quality, task_type, agent_id, created_at, used_count)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			e.ID, e.Question, e.Answer, e.Model, e.Quality,
			e.TaskType, e.AgentID, e.CreatedAt, e.UsedCount,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
