// TiBrain CLI Context System
// Provides local context management for CLI operations
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// ─────────────────────────────────────────────────────────────
// CLI Context Data Structures
// ─────────────────────────────────────────────────────────────

// CLIContext represents a context entry
type CLIContext struct {
	ID          string    `json:"id"`
	CLIID       string    `json:"cli_id"`
	ContextType string    `json:"context_type"` // config, help, cache, preference
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Metadata    string    `json:"metadata"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	AccessCount int       `json:"access_count"`
	LastAccessed *time.Time `json:"last_accessed,omitempty"`
}

// CLICacheEntry represents a cached query result
type CLICacheEntry struct {
	ID         string     `json:"id"`
	QueryHash  string     `json:"query_hash"`
	Query      string     `json:"query"`
	Response   string     `json:"response"`
	Confidence float64    `json:"confidence"`
	SourceType string     `json:"source_type"` // local, api, hybrid
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	HitCount   int        `json:"hit_count"`
	LastHit    *time.Time `json:"last_hit,omitempty"`
}

// CLIHelpEntry represents help documentation
type CLIHelpEntry struct {
	ID        string    `json:"id"`
	Command   string    `json:"command"`
	Topic     string    `json:"topic"`
	Content   string    `json:"content"`
	Keywords  string    `json:"keywords"`
	Category  string    `json:"category"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CLIPreference represents user preferences
type CLIPreference struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ─────────────────────────────────────────────────────────────
// CLI Context Manager
// ─────────────────────────────────────────────────────────────

type CLIContextManager struct {
	hub    *Hub
	cache  *sync.Map // L1 cache for hot data
	mutex  sync.RWMutex
}

func NewCLIContextManager(hub *Hub) *CLIContextManager {
	return &CLIContextManager{
		hub:   hub,
		cache: &sync.Map{},
	}
}

// ─────────────────────────────────────────────────────────────
// Context Operations
// ─────────────────────────────────────────────────────────────

// SetContext stores a context entry
func (cm *CLIContextManager) SetContext(cliID, contextType, key, value string, metadata map[string]string, ttl time.Duration) error {
	timestamp := time.Now().Unix()
	var expiresAt *int64
	
	if ttl > 0 {
		exp := time.Now().Add(ttl).Unix()
		expiresAt = &exp
	}
	
	metadataJSON, _ := json.Marshal(metadata)
	
	_, err := cm.hub.db.Exec(`
		INSERT OR REPLACE INTO cli_context 
		(id, cli_id, context_type, key, value, metadata, created_at, updated_at, expires_at, access_count, last_accessed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?)
	`,
		generateID(), cliID, contextType, key, value, string(metadataJSON),
		timestamp, timestamp, expiresAt, timestamp)
	
	if err != nil {
		return fmt.Errorf("set context: %w", err)
	}
	
	// Update L1 cache
	cacheKey := fmt.Sprintf("%s:%s:%s", cliID, contextType, key)
	cm.cache.Store(cacheKey, value)
	
	logger.Info("CLI context set: %s/%s/%s", cliID, contextType, key)
	return nil
}

// GetContext retrieves a context entry
func (cm *CLIContextManager) GetContext(cliID, contextType, key string) (*CLIContext, error) {
	// Try L1 cache first
	cacheKey := fmt.Sprintf("%s:%s:%s", cliID, contextType, key)
	if cached, ok := cm.cache.Load(cacheKey); ok {
		// Update access count asynchronously
		go cm.updateAccessCount(cliID, contextType, key)
		return &CLIContext{
			CLIID:       cliID,
			ContextType: contextType,
			Key:         key,
			Value:       cached.(string),
		}, nil
	}
	
	var ctx CLIContext
	var createdAt, updatedAt, lastAccessed int64
	var expiresAt sql.NullInt64
	
	err := cm.hub.db.QueryRow(`
		SELECT id, cli_id, context_type, key, value, metadata, 
		       created_at, updated_at, expires_at, access_count, last_accessed
		FROM cli_context 
		WHERE cli_id = ? AND context_type = ? AND key = ?
		  AND (expires_at IS NULL OR expires_at > ?)
	`, cliID, contextType, key, time.Now().Unix()).Scan(
		&ctx.ID, &ctx.CLIID, &ctx.ContextType, &ctx.Key, &ctx.Value, &ctx.Metadata,
		&createdAt, &updatedAt, &expiresAt, &ctx.AccessCount, &lastAccessed)
	
	if err != nil {
		return nil, fmt.Errorf("get context: %w", err)
	}
	
	ctx.CreatedAt = time.Unix(createdAt, 0)
	ctx.UpdatedAt = time.Unix(updatedAt, 0)
	if lastAccessed > 0 {
		t := time.Unix(lastAccessed, 0)
		ctx.LastAccessed = &t
	}
	if expiresAt.Valid {
		t := time.Unix(expiresAt.Int64, 0)
		ctx.ExpiresAt = &t
	}
	
	// Update cache and access count
	cm.cache.Store(cacheKey, ctx.Value)
	go cm.updateAccessCount(cliID, contextType, key)
	
	return &ctx, nil
}

// DeleteContext removes a context entry
func (cm *CLIContextManager) DeleteContext(cliID, contextType, key string) error {
	_, err := cm.hub.db.Exec(`
		DELETE FROM cli_context WHERE cli_id = ? AND context_type = ? AND key = ?
	`, cliID, contextType, key)
	
	if err != nil {
		return fmt.Errorf("delete context: %w", err)
	}
	
	// Remove from cache
	cacheKey := fmt.Sprintf("%s:%s:%s", cliID, contextType, key)
	cm.cache.Delete(cacheKey)
	
	logger.Info("CLI context deleted: %s/%s/%s", cliID, contextType, key)
	return nil
}

// ListContexts lists all contexts for a CLI
func (cm *CLIContextManager) ListContexts(cliID, contextType string) ([]CLIContext, error) {
	var args []interface{}
	query := `
		SELECT id, cli_id, context_type, key, value, metadata, 
		       created_at, updated_at, expires_at, access_count, last_accessed
		FROM cli_context 
		WHERE cli_id = ?
	`
	args = append(args, cliID)
	
	if contextType != "" {
		query += " AND context_type = ?"
		args = append(args, contextType)
	}
	
	query += " ORDER BY updated_at DESC"
	
	rows, err := cm.hub.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list contexts: %w", err)
	}
	defer rows.Close()
	
	var contexts []CLIContext
	for rows.Next() {
		var ctx CLIContext
		var createdAt, updatedAt, lastAccessed int64
		var expiresAt sql.NullInt64
		
		err := rows.Scan(
			&ctx.ID, &ctx.CLIID, &ctx.ContextType, &ctx.Key, &ctx.Value, &ctx.Metadata,
			&createdAt, &updatedAt, &expiresAt, &ctx.AccessCount, &lastAccessed)
		if err != nil {
			return nil, fmt.Errorf("scan context: %w", err)
		}
		
		ctx.CreatedAt = time.Unix(createdAt, 0)
		ctx.UpdatedAt = time.Unix(updatedAt, 0)
		if lastAccessed > 0 {
			t := time.Unix(lastAccessed, 0)
			ctx.LastAccessed = &t
		}
		if expiresAt.Valid {
			t := time.Unix(expiresAt.Int64, 0)
			ctx.ExpiresAt = &t
		}
		
		contexts = append(contexts, ctx)
	}
	
	return contexts, nil
}

// ─────────────────────────────────────────────────────────────
// Cache Operations
// ─────────────────────────────────────────────────────────────

// SetCache stores a query result in cache
func (cm *CLIContextManager) SetCache(query, response string, confidence float64, sourceType string, ttl time.Duration) error {
	queryHash := generateQueryHash(query)
	timestamp := time.Now().Unix()
	var expiresAt *int64
	
	if ttl > 0 {
		exp := time.Now().Add(ttl).Unix()
		expiresAt = &exp
	}
	
	_, err := cm.hub.db.Exec(`
		INSERT OR REPLACE INTO cli_query_cache 
		(id, query_hash, query, response, confidence, source_type, created_at, expires_at, hit_count, last_hit)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?)
	`,
		generateID(), queryHash, query, response, confidence, sourceType,
		timestamp, expiresAt, timestamp)
	
	if err != nil {
		return fmt.Errorf("set cache: %w", err)
	}
	
	// Update L1 cache
	cm.cache.Store("cache:"+queryHash, response)
	
	logger.Info("CLI cache set: %s (source: %s)", queryHash[:8], sourceType)
	return nil
}

// GetCache retrieves a cached query result
func (cm *CLIContextManager) GetCache(query string) (*CLICacheEntry, error) {
	queryHash := generateQueryHash(query)
	
	// Try L1 cache first
	if cached, ok := cm.cache.Load("cache:" + queryHash); ok {
		// Update hit count asynchronously
		go cm.updateCacheHitCount(queryHash)
		return &CLICacheEntry{
			QueryHash: queryHash,
			Query:     query,
			Response:  cached.(string),
		}, nil
	}
	
	var entry CLICacheEntry
	var createdAt, lastHit int64
	var expiresAt sql.NullInt64
	
	err := cm.hub.db.QueryRow(`
		SELECT id, query_hash, query, response, confidence, source_type,
		       created_at, expires_at, hit_count, last_hit
		FROM cli_query_cache 
		WHERE query_hash = ? AND (expires_at IS NULL OR expires_at > ?)
	`, queryHash, time.Now().Unix()).Scan(
		&entry.ID, &entry.QueryHash, &entry.Query, &entry.Response, &entry.Confidence, &entry.SourceType,
		&createdAt, &expiresAt, &entry.HitCount, &lastHit)
	
	if err != nil {
		return nil, fmt.Errorf("get cache: %w", err)
	}
	
	entry.CreatedAt = time.Unix(createdAt, 0)
	if lastHit > 0 {
		t := time.Unix(lastHit, 0)
		entry.LastHit = &t
	}
	if expiresAt.Valid {
		t := time.Unix(expiresAt.Int64, 0)
		entry.ExpiresAt = &t
	}
	
	// Update cache and hit count
	cm.cache.Store("cache:"+queryHash, entry.Response)
	go cm.updateCacheHitCount(queryHash)
	
	return &entry, nil
}

// ClearCache removes expired cache entries
func (cm *CLIContextManager) ClearCache() error {
	result, err := cm.hub.db.Exec(`
		DELETE FROM cli_query_cache WHERE expires_at IS NOT NULL AND expires_at <= ?
	`, time.Now().Unix())
	
	if err != nil {
		return fmt.Errorf("clear cache: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	logger.Info("CLI cache cleared: %d expired entries removed", rowsAffected)
	
	return nil
}

// ─────────────────────────────────────────────────────────────
// Help System Operations
// ─────────────────────────────────────────────────────────────

// AddHelp adds help documentation
func (cm *CLIContextManager) AddHelp(command, topic, content, keywords, category string, priority int) error {
	timestamp := time.Now().Unix()
	
	_, err := cm.hub.db.Exec(`
		INSERT OR REPLACE INTO cli_help_index 
		(id, command, topic, content, keywords, category, priority, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		generateID(), command, topic, content, keywords, category, priority, timestamp, timestamp)
	
	if err != nil {
		return fmt.Errorf("add help: %w", err)
	}
	
	logger.Info("CLI help added: %s/%s", command, topic)
	return nil
}

// SearchHelp searches help documentation
func (cm *CLIContextManager) SearchHelp(query string, limit int) ([]CLIHelpEntry, error) {
	// Simple text search for now
	searchPattern := "%" + query + "%"
	
	rows, err := cm.hub.db.Query(`
		SELECT id, command, topic, content, keywords, category, priority, created_at, updated_at
		FROM cli_help_index 
		WHERE content LIKE ? OR command LIKE ? OR topic LIKE ? OR keywords LIKE ?
		ORDER BY priority DESC, updated_at DESC
		LIMIT ?
	`, searchPattern, searchPattern, searchPattern, searchPattern, limit)
	
	if err != nil {
		return nil, fmt.Errorf("search help: %w", err)
	}
	defer rows.Close()
	
	var entries []CLIHelpEntry
	for rows.Next() {
		var entry CLIHelpEntry
		var createdAt, updatedAt int64
		
		err := rows.Scan(
			&entry.ID, &entry.Command, &entry.Topic, &entry.Content, &entry.Keywords,
			&entry.Category, &entry.Priority, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan help entry: %w", err)
		}
		
		entry.CreatedAt = time.Unix(createdAt, 0)
		entry.UpdatedAt = time.Unix(updatedAt, 0)
		entries = append(entries, entry)
	}
	
	return entries, nil
}

// ─────────────────────────────────────────────────────────────
// Preference Operations
// ─────────────────────────────────────────────────────────────

// SetPreference sets a user preference
func (cm *CLIContextManager) SetPreference(userID, key, value string) error {
	timestamp := time.Now().Unix()
	
	_, err := cm.hub.db.Exec(`
		INSERT OR REPLACE INTO cli_preferences 
		(id, user_id, key, value, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		generateID(), userID, key, value, timestamp, timestamp)
	
	if err != nil {
		return fmt.Errorf("set preference: %w", err)
	}
	
	// Update cache
	cacheKey := fmt.Sprintf("pref:%s:%s", userID, key)
	cm.cache.Store(cacheKey, value)
	
	logger.Info("CLI preference set: %s/%s", userID, key)
	return nil
}

// GetPreference gets a user preference
func (cm *CLIContextManager) GetPreference(userID, key string) (*CLIPreference, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("pref:%s:%s", userID, key)
	if cached, ok := cm.cache.Load(cacheKey); ok {
		return &CLIPreference{
			UserID: userID,
			Key:    key,
			Value:  cached.(string),
		}, nil
	}
	
	var pref CLIPreference
	var createdAt, updatedAt int64
	
	err := cm.hub.db.QueryRow(`
		SELECT id, user_id, key, value, created_at, updated_at
		FROM cli_preferences 
		WHERE user_id = ? AND key = ?
	`, userID, key).Scan(
		&pref.ID, &pref.UserID, &pref.Key, &pref.Value, &createdAt, &updatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("get preference: %w", err)
	}
	
	pref.CreatedAt = time.Unix(createdAt, 0)
	pref.UpdatedAt = time.Unix(updatedAt, 0)
	
	// Update cache
	cm.cache.Store(cacheKey, pref.Value)
	
	return &pref, nil
}

// ─────────────────────────────────────────────────────────────
// Utility Functions
// ─────────────────────────────────────────────────────────────

func (cm *CLIContextManager) updateAccessCount(cliID, contextType, key string) {
	_, err := cm.hub.db.Exec(`
		UPDATE cli_context 
		SET access_count = access_count + 1, last_accessed = ?
		WHERE cli_id = ? AND context_type = ? AND key = ?
	`, time.Now().Unix(), cliID, contextType, key)
	
	if err != nil {
		logger.Warn("Failed to update access count: %v", err)
	}
}

func (cm *CLIContextManager) updateCacheHitCount(queryHash string) {
	_, err := cm.hub.db.Exec(`
		UPDATE cli_query_cache 
		SET hit_count = hit_count + 1, last_hit = ?
		WHERE query_hash = ?
	`, time.Now().Unix(), queryHash)
	
	if err != nil {
		logger.Warn("Failed to update cache hit count: %v", err)
	}
}

func generateQueryHash(query string) string {
	return fmt.Sprintf("%x", len(query)+int(time.Now().Unix()))
}

// ─────────────────────────────────────────────────────────────
// HTTP Handlers for CLI Context System
// ─────────────────────────────────────────────────────────────

func (h *Hub) setupCLIContextHandlers() {
	contextManager := NewCLIContextManager(h)

	// Context management endpoints
	http.HandleFunc("/cli/context", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			var req struct {
				CLIID       string            `json:"cli_id"`
				ContextType string            `json:"context_type"`
				Key         string            `json:"key"`
				Value       string            `json:"value"`
				Metadata    map[string]string `json:"metadata,omitempty"`
				TTL         int               `json:"ttl,omitempty"` // seconds
			}
			
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			
			ttl := time.Duration(req.TTL) * time.Second
			if err := contextManager.SetContext(req.CLIID, req.ContextType, req.Key, req.Value, req.Metadata, ttl); err != nil {
				http.Error(w, fmt.Sprintf("Set context error: %v", err), http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
			
		case "GET":
			cliID := r.URL.Query().Get("cli_id")
			contextType := r.URL.Query().Get("type")
			key := r.URL.Query().Get("key")
			
			if cliID == "" || key == "" {
				http.Error(w, "cli_id and key required", http.StatusBadRequest)
				return
			}
			
			ctx, err := contextManager.GetContext(cliID, contextType, key)
			if err != nil {
				http.Error(w, fmt.Sprintf("Get context error: %v", err), http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ctx)
		}
	})

	// Cache management endpoints
	http.HandleFunc("/cli/cache", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			var req struct {
				Query      string  `json:"query"`
				Response   string  `json:"response"`
				Confidence float64 `json:"confidence"`
				SourceType string  `json:"source_type"`
				TTL        int     `json:"ttl,omitempty"`
			}
			
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			
			ttl := time.Duration(req.TTL) * time.Second
			if err := contextManager.SetCache(req.Query, req.Response, req.Confidence, req.SourceType, ttl); err != nil {
				http.Error(w, fmt.Sprintf("Set cache error: %v", err), http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
			
		case "GET":
			query := r.URL.Query().Get("query")
			if query == "" {
				http.Error(w, "query required", http.StatusBadRequest)
				return
			}
			
			entry, err := contextManager.GetCache(query)
			if err != nil {
				http.Error(w, fmt.Sprintf("Get cache error: %v", err), http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(entry)
		}
	})

	// Help system endpoints
	http.HandleFunc("/cli/help", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		limit := 5
		
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
				limit = parsedLimit
			}
		}
		
		if query == "" {
			http.Error(w, "query required", http.StatusBadRequest)
			return
		}
		
		entries, err := contextManager.SearchHelp(query, limit)
		if err != nil {
			http.Error(w, fmt.Sprintf("Search help error: %v", err), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	})

	// Preference management endpoints
	http.HandleFunc("/cli/preferences", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			var req struct {
				UserID string `json:"user_id"`
				Key    string `json:"key"`
				Value  string `json:"value"`
			}
			
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			
			if err := contextManager.SetPreference(req.UserID, req.Key, req.Value); err != nil {
				http.Error(w, fmt.Sprintf("Set preference error: %v", err), http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
			
		case "GET":
			userID := r.URL.Query().Get("user_id")
			key := r.URL.Query().Get("key")
			
			if userID == "" || key == "" {
				http.Error(w, "user_id and key required", http.StatusBadRequest)
				return
			}
			
			pref, err := contextManager.GetPreference(userID, key)
			if err != nil {
				http.Error(w, fmt.Sprintf("Get preference error: %v", err), http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(pref)
		}
	})

	logger.Info("CLI Context system HTTP handlers registered")
}