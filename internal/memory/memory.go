// Package memory implements the TiBrain Cognitive Memory System.
//
// Inspired by human cognition:
//   - episodic:   personal experiences ("I learned X yesterday")
//   - semantic:   facts and knowledge ("X is Y")
//   - procedural: skills and procedures ("to do X, do Y then Z")
//   - working:    current context (in-memory only, never persisted)
//   - long_term:  consolidated knowledge (after sleep-like process)
//
// Brain-like properties implemented here:
//   - Importance-weighted retrieval (not just recency)
//   - Access-based strengthening (every read reinforces the memory)
//   - Sleep-like consolidation (episodic -> long_term after enough access)
//   - Forgetting via expires_at (TTL)
//   - Atomic consolidation via transaction
//   - Context-aware DB ops (cancellable queries)
//
// Dependencies are abstracted via small interfaces (DB, GraphStore, IDGen) so
// this package has no compile-time dependency on the root `main` package.
package memory

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// MemoryType defines the type of memory.
type MemoryType string

const (
	MemoryEpisodic   MemoryType = "episodic"   // Personal experiences
	MemorySemantic   MemoryType = "semantic"   // Facts and knowledge
	MemoryProcedural MemoryType = "procedural" // Skills and procedures
	MemoryWorking    MemoryType = "working"    // Current context (in-memory only)
	MemoryLongTerm   MemoryType = "long_term"  // Consolidated knowledge
)

// DB is the minimal database interface CognitiveMemoryManager needs.
// Satisfied by *sql.DB and any wrapper that exposes ExecContext / QueryContext
// / PrepareContext / QueryRowContext.
type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// GraphNode is the minimal graph-node shape this package needs.
// Callers should pass concrete GraphNode types from their graph package.
type GraphNode struct {
	ID         string
	Labels     []string
	Properties map[string]interface{}
}

// GraphStore is the minimal graph-store interface for optional Neo4j integration.
// The node argument is typed as `any` so callers can pass their own GraphNode type
// without requiring a hard dependency on this package's GraphNode.
type GraphStore interface {
	AddNode(ctx context.Context, node any) (string, error)
}

// IDGen generates unique memory IDs. Implementations should guarantee
// uniqueness even under rapid repeated calls (see defaultIDGen below).
type IDGen func() string

// defaultIDGen produces unique IDs by combining nanosecond timestamp,
// an atomic counter, and a 4-byte random suffix. Guarantees uniqueness
// even when called multiple times per nanosecond — which the original
// implementation did not, and which caused UNIQUE constraint failures.
var defaultCounter uint64

func defaultIDGen() string {
	counter := atomic.AddUint64(&defaultCounter, 1)
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		ts := time.Now().UnixNano()
		b[0] = byte(ts)
		b[1] = byte(ts >> 8)
		b[2] = byte(ts >> 16)
		b[3] = byte(ts >> 24)
	}
	return fmt.Sprintf("mem_%d_%d_%s", time.Now().UnixNano(), counter, fmt.Sprintf("%x", b))
}

// CognitiveMemoryManager manages the cognitive memory system.
//
// Concurrency model:
//   - workingMu protects the in-memory working memory map
//   - DB ops use SQLite's own connection-level locking (no app-level locks needed)
//   - Graph writes are fire-and-forget goroutines (don't block the SQL write)
type CognitiveMemoryManager struct {
	db         DB
	graph      GraphStore
	generateID IDGen

	workingMu     sync.RWMutex
	workingMemory map[string]interface{}

	// Prepared statements (lazy, one-time init).
	stmtOnce sync.Once
	stmtErr  error
	insert   *sql.Stmt
	strength *sql.Stmt
}

// MemoryEntry represents a single memory entry.
type MemoryEntry struct {
	ID          string                 `json:"id"`
	Type        MemoryType             `json:"type"`
	Content     string                 `json:"content"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Importance  float64                `json:"importance"`
	AccessCount int64                  `json:"access_count"`
	LastAccess  time.Time              `json:"last_access"`
	CreatedAt   time.Time              `json:"created_at"`
	Tags        []string               `json:"tags,omitempty"`
	Source      string                 `json:"source,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// NewCognitiveMemoryManager creates a new cognitive memory manager.
// graph may be nil. If generateID is nil, uses a process-unique default.
func NewCognitiveMemoryManager(db DB, graph GraphStore) *CognitiveMemoryManager {
	return NewCognitiveMemoryManagerWithIDGen(db, graph, nil)
}

// NewCognitiveMemoryManagerWithIDGen is like NewCognitiveMemoryManager but
// accepts a custom ID generator (useful for tests).
func NewCognitiveMemoryManagerWithIDGen(db DB, graph GraphStore, idGen IDGen) *CognitiveMemoryManager {
	if idGen == nil {
		idGen = defaultIDGen
	}
	return &CognitiveMemoryManager{
		db:            db,
		graph:         graph,
		generateID:    idGen,
		workingMemory: make(map[string]interface{}),
	}
}

// prepareStatements lazily initializes prepared statements. Safe to call from
// multiple goroutines — only the first call does work.
func (cm *CognitiveMemoryManager) prepareStatements(ctx context.Context) error {
	cm.stmtOnce.Do(func() {
		cm.insert, cm.stmtErr = cm.db.PrepareContext(ctx, `
			INSERT INTO memories
				(id, api_key_id, type, key, content, metadata, importance, access_count, last_access, created_at, updated_at, expires_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
		if cm.stmtErr != nil {
			return
		}
		cm.strength, cm.stmtErr = cm.db.PrepareContext(ctx, `
			UPDATE memories
			SET access_count = access_count + 1,
			    last_access  = ?,
			    updated_at   = ?
			WHERE id = ?`)
	})
	return cm.stmtErr
}

// StoreEpisodicMemory stores a personal experience.
func (cm *CognitiveMemoryManager) StoreEpisodicMemory(ctx context.Context, content string, contextData map[string]interface{}) (string, error) {
	return cm.store(ctx, MemoryEpisodic, "", content, contextData, 0.7, nil, "agent_interaction", 0)
}

// StoreSemanticMemory stores factual knowledge.
func (cm *CognitiveMemoryManager) StoreSemanticMemory(ctx context.Context, content string, tags []string, source string) (string, error) {
	return cm.store(ctx, MemorySemantic, "", content, nil, 0.8, tags, source, 0)
}

// StoreProceduralMemory stores skills and procedures.
func (cm *CognitiveMemoryManager) StoreProceduralMemory(ctx context.Context, procedure string, contextData map[string]interface{}) (string, error) {
	return cm.store(ctx, MemoryProcedural, "", procedure, contextData, 0.9, nil, "procedure_learning", 0)
}

// StoreMemoryWithTTL stores an arbitrary memory type with a time-to-live.
// After ttl elapses, ForgetExpired will remove it.
func (cm *CognitiveMemoryManager) StoreMemoryWithTTL(ctx context.Context, memType MemoryType, content string, ttl time.Duration) (string, error) {
	return cm.store(ctx, memType, "", content, nil, 0.5, nil, "ttl_store", ttl)
}

// store is the single internal write path used by all Store* methods.
// This eliminates the 3× duplicated INSERT logic in the original code.
func (cm *CognitiveMemoryManager) store(
	ctx context.Context,
	memType MemoryType,
	key, content string,
	contextData map[string]interface{},
	importance float64,
	tags []string,
	source string,
	ttl time.Duration,
) (string, error) {
	if memType == MemoryWorking {
		// Working memory is in-memory only — caller should use UpdateWorkingMemory.
		return "", errors.New("working memory must be set via UpdateWorkingMemory")
	}

	if err := cm.prepareStatements(ctx); err != nil {
		return "", fmt.Errorf("prepare statements: %w", err)
	}

	id := cm.generateID()
	now := time.Now()

	entry := MemoryEntry{
		ID:         id,
		Type:       memType,
		Content:    content,
		Context:    contextData,
		Importance: importance,
		Tags:       tags,
		Source:     source,
		CreatedAt:  now,
		LastAccess: now,
	}

	// Compose metadata: user-provided context + tags + source (if any).
	meta := map[string]interface{}{}
	for k, v := range contextData {
		meta[k] = v
	}
	if len(tags) > 0 {
		meta["tags"] = tags
	}
	if source != "" {
		meta["source"] = source
	}
	meta["memory_type"] = string(memType)
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return "", fmt.Errorf("marshal metadata: %w", err)
	}

	var expiresAt *string
	if ttl > 0 {
		exp := now.Add(ttl).UTC().Format(time.RFC3339Nano)
		expiresAt = &exp
	}

	var expiresArg interface{}
	if expiresAt != nil {
		expiresArg = *expiresAt
	}

	_, err = cm.insert.ExecContext(ctx,
		entry.ID,
		"local", // api_key_id — single-tenant local brain
		entry.Type,
		key,
		entry.Content,
		string(metaJSON),
		entry.Importance,
		1, // initial access_count
		now.UTC().Format(time.RFC3339Nano),
		now.UTC().Format(time.RFC3339Nano),
		now.UTC().Format(time.RFC3339Nano),
		expiresArg,
	)
	if err != nil {
		return "", fmt.Errorf("insert %s memory: %w", memType, err)
	}

	// Async graph write — never block the SQL insert on Neo4j latency.
	if cm.graph != nil {
		go cm.addToGraph(memType, entry)
	}

	return id, nil
}

// addToGraph pushes a memory into Neo4j (runs in goroutine; logs on failure).
func (cm *CognitiveMemoryManager) addToGraph(memType MemoryType, entry MemoryEntry) {
	bg, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	labels := []string{string(memType) + "Memory", "Memory"}
	props := map[string]interface{}{
		"content":    entry.Content,
		"created_at": entry.CreatedAt,
		"type":       string(memType),
	}
	if len(entry.Tags) > 0 {
		props["tags"] = entry.Tags
	}
	if _, err := cm.graph.AddNode(bg, GraphNode{
		ID:         entry.ID,
		Labels:     labels,
		Properties: props,
	}); err != nil {
		log.Printf("[cognitive_memory] graph add failed for %s %s: %v", memType, entry.ID, err)
	}
}

// QueryMemory retrieves memories by type, optionally filtered by a free-text
// query on content. If query is empty, returns the top-N by importance + recency.
// Expired memories (expires_at <= now) are excluded.
func (cm *CognitiveMemoryManager) QueryMemory(ctx context.Context, query string, memoryType MemoryType, limit int) ([]MemoryEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)

	var (
		rows *sql.Rows
		err  error
	)

	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		// No text filter — order by importance + recency.
		rows, err = cm.db.QueryContext(ctx, `
			SELECT id, type, content, metadata, importance, access_count, last_access, created_at, expires_at
			FROM memories
			WHERE type = ?
			  AND (expires_at IS NULL OR expires_at > ?)
			ORDER BY importance DESC, access_count DESC, last_access DESC
			LIMIT ?`,
			memoryType, now, limit)
	} else {
		// Text filter via LIKE — simple substring match. For richer search,
		// use the FTS5 virtual table from migration 022.
		like := "%" + trimmedQuery + "%"
		rows, err = cm.db.QueryContext(ctx, `
			SELECT id, type, content, metadata, importance, access_count, last_access, created_at, expires_at
			FROM memories
			WHERE type = ?
			  AND content LIKE ?
			  AND (expires_at IS NULL OR expires_at > ?)
			ORDER BY importance DESC, access_count DESC
			LIMIT ?`,
			memoryType, like, now, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("query %s memory: %w", memoryType, err)
	}
	defer rows.Close()

	var results []MemoryEntry
	for rows.Next() {
		entry, scanErr := scanMemoryRow(rows)
		if scanErr != nil {
			log.Printf("[cognitive_memory] scan row: %v", scanErr)
			continue
		}
		results = append(results, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate memory rows: %w", err)
	}
	return results, nil
}

// scanMemoryRow decodes one row from the memories table into a MemoryEntry.
func scanMemoryRow(rows *sql.Rows) (MemoryEntry, error) {
	var (
		entry                            MemoryEntry
		metadataJSON                     sql.NullString
		lastAccess, createdAt, expiresAt sql.NullString
	)
	if err := rows.Scan(
		&entry.ID, &entry.Type, &entry.Content, &metadataJSON,
		&entry.Importance, &entry.AccessCount,
		&lastAccess, &createdAt, &expiresAt,
	); err != nil {
		return entry, err
	}
	if lastAccess.Valid {
		if t, err := time.Parse(time.RFC3339Nano, lastAccess.String); err == nil {
			entry.LastAccess = t
		}
	}
	if createdAt.Valid {
		if t, err := time.Parse(time.RFC3339Nano, createdAt.String); err == nil {
			entry.CreatedAt = t
		}
	}
	if expiresAt.Valid && expiresAt.String != "" {
		if t, err := time.Parse(time.RFC3339Nano, expiresAt.String); err == nil {
			entry.ExpiresAt = &t
		}
	}
	if metadataJSON.Valid && metadataJSON.String != "" {
		var meta map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON.String), &meta); err == nil {
			entry.Metadata = meta
			// Promote tags from metadata if present.
			if t, ok := meta["tags"].([]interface{}); ok {
				for _, v := range t {
					if s, ok := v.(string); ok {
						entry.Tags = append(entry.Tags, s)
					}
				}
			}
			// Promote source from metadata if present.
			if s, ok := meta["source"].(string); ok {
				entry.Source = s
			}
		}
	}
	return entry, nil
}

// UpdateWorkingMemory updates the current working memory context (in-memory only).
func (cm *CognitiveMemoryManager) UpdateWorkingMemory(key string, value interface{}) {
	cm.workingMu.Lock()
	defer cm.workingMu.Unlock()
	cm.workingMemory[key] = value
}

// GetWorkingMemory retrieves a value from working memory.
func (cm *CognitiveMemoryManager) GetWorkingMemory(key string) (interface{}, bool) {
	cm.workingMu.RLock()
	defer cm.workingMu.RUnlock()
	val, ok := cm.workingMemory[key]
	return val, ok
}

// ClearWorkingMemory removes all entries from working memory.
func (cm *CognitiveMemoryManager) ClearWorkingMemory() {
	cm.workingMu.Lock()
	defer cm.workingMu.Unlock()
	cm.workingMemory = make(map[string]interface{})
}

// StrengthenMemory increments access_count and updates last_access.
// Returns ErrMemoryNotFound if no memory has that ID.
func (cm *CognitiveMemoryManager) StrengthenMemory(ctx context.Context, memoryID string) error {
	if err := cm.prepareStatements(ctx); err != nil {
		return fmt.Errorf("prepare statements: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := cm.strength.ExecContext(ctx, now, now, memoryID)
	if err != nil {
		return fmt.Errorf("strengthen memory: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("strengthen memory rows: %w", err)
	}
	if n == 0 {
		return ErrMemoryNotFound
	}
	return nil
}

// ErrMemoryNotFound is returned when a memory ID does not exist.
var ErrMemoryNotFound = errors.New("memory not found")

// ConsolidateMemory promotes frequently-accessed episodic memories to long_term.
// "Sleep-like" process: anything episodic accessed >= minAccess times gets transferred.
//
// The whole promotion runs in a single transaction so callers see either all or none.
// Returns the number of memories promoted.
func (cm *CognitiveMemoryManager) ConsolidateMemory(ctx context.Context, minAccess int64) (int64, error) {
	if minAccess <= 0 {
		minAccess = 5
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := cm.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin consolidation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		UPDATE memories
		SET type        = ?,
		    last_access = ?,
		    updated_at  = ?
		WHERE type = ?
		  AND access_count >= ?
		  AND (expires_at IS NULL OR expires_at > ?)`,
		MemoryLongTerm, now, now, MemoryEpisodic, minAccess, now)
	if err != nil {
		return 0, fmt.Errorf("consolidate memories: %w", err)
	}
	n, _ := res.RowsAffected()
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit consolidation: %w", err)
	}
	log.Printf("[cognitive_memory] consolidated %d episodic memories to long_term (threshold=%d)", n, minAccess)
	return n, nil
}

// ForgetExpired removes all memories whose expires_at <= now.
// Returns the number of rows deleted.
func (cm *CognitiveMemoryManager) ForgetExpired(ctx context.Context) (int64, error) {
	res, err := cm.db.ExecContext(ctx, `
		DELETE FROM memories
		WHERE expires_at IS NOT NULL AND expires_at <= ?`,
		time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return 0, fmt.Errorf("forget expired: %w", err)
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		log.Printf("[cognitive_memory] forgot %d expired memories", n)
	}
	return n, nil
}

// GetMemoryStats returns aggregate statistics.
// Single GROUP BY query instead of N+1 count queries.
func (cm *CognitiveMemoryManager) GetMemoryStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	now := time.Now().UTC().Format(time.RFC3339Nano)

	// Counts by type, in a single query.
	rows, err := cm.db.QueryContext(ctx, `
		SELECT type, COUNT(*)
		FROM memories
		WHERE expires_at IS NULL OR expires_at > ?
		GROUP BY type`, now)
	if err != nil {
		return nil, fmt.Errorf("count memories by type: %w", err)
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		var memType string
		var count int
		if err := rows.Scan(&memType, &count); err != nil {
			log.Printf("[cognitive_memory] stats scan: %v", err)
			continue
		}
		stats[memType] = count
		total += count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stats: %w", err)
	}
	stats["total"] = total

	// Top 5 most-accessed.
	topRows, err := cm.db.QueryContext(ctx, `
		SELECT id, content, access_count
		FROM memories
		WHERE expires_at IS NULL OR expires_at > ?
		ORDER BY access_count DESC, last_access DESC
		LIMIT 5`, now)
	if err == nil {
		defer topRows.Close()
		top := make([]map[string]interface{}, 0)
		for topRows.Next() {
			var id, content string
			var accesses int64
			if err := topRows.Scan(&id, &content, &accesses); err == nil {
				top = append(top, map[string]interface{}{
					"id":       id,
					"content":  content,
					"accesses": accesses,
				})
			}
		}
		stats["top_memories"] = top
	}

	cm.workingMu.RLock()
	stats["working_memory_keys"] = len(cm.workingMemory)
	cm.workingMu.RUnlock()

	return stats, nil
}

// GetRecentExperience retrieves the most recent episodic memories.
func (cm *CognitiveMemoryManager) GetRecentExperience(ctx context.Context, limit int) ([]MemoryEntry, error) {
	return cm.QueryMemory(ctx, "", MemoryEpisodic, limit)
}

// GetRelevantKnowledge retrieves semantic memories matching a query.
func (cm *CognitiveMemoryManager) GetRelevantKnowledge(ctx context.Context, query string, limit int) ([]MemoryEntry, error) {
	return cm.QueryMemory(ctx, query, MemorySemantic, limit)
}
