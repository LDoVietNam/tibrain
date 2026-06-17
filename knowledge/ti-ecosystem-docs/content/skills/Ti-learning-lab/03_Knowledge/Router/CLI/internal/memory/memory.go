// Package memory implements Ti's 4-Layer Memory Stack — ported from MemPalace
// architecture but written natively in Go with SQLite persistence.
//
// Architecture:
//
//	Layer 0: Identity       (~50-100 tokens)  — Always loaded. "Who am I?"
//	Layer 1: Essential Story (~500-800 tokens) — Always loaded. Top weighted memories.
//	Layer 2: On-Demand      (~200-500 tokens)  — Wing/room filtered retrieval.
//	Layer 3: Deep Search    (unlimited)         — Full semantic search with scoring.
//
// Palace → Wings → Rooms → Drawers
// A Drawer is the atomic memory unit: verbatim text with metadata.
package memory

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// ─────────────────────────────────────────────────────────────
// Data Types
// ─────────────────────────────────────────────────────────────

// Palace is the top-level memory store, analogous to a knowledge base.
// It contains Wings (projects/agents), each containing Rooms (topics),
// each containing Drawers (individual memory chunks).
type Palace struct {
	Path     string           // Base directory for palace data (~/.ti/memory/)
	Identity string           // identity.txt content
	Wings    map[string]*Wing // wing name → Wing
	db       *sql.DB          // SQLite database handle
	mu       sync.RWMutex     // Thread safety
	halfLife float64          // Half-life in hours for recency decay
}

// Wing represents a project or agent-specific memory collection.
// Each wing contains multiple Rooms organized by topic.
type Wing struct {
	Name  string           // "my_project", "agent_coder"
	Rooms map[string]*Room // room name → Room
}

// Room is a topic within a wing, grouping related Drawers together.
type Room struct {
	Name    string   // "auth", "api", "deployment"
	Drawers []Drawer // Individual memory chunks
}

// Drawer is the atomic memory unit — a verbatim text chunk with metadata.
type Drawer struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	Importance float64 `json:"importance"`
	Category   string  `json:"category"`
	CreatedAt  int64   `json:"created_at"`
	SourceFile string  `json:"source_file,omitempty"`
	Wing       string  `json:"wing"`
	Room       string  `json:"room"`
}

// Stack represents the active memory retrieval for a session,
// combining all four layers of the memory stack.
type Stack struct {
	L0 string   `json:"l0_identity"`    // Identity text
	L1 []Drawer `json:"l1_essential"`   // Essential story (top weighted)
	L2 []Drawer `json:"l2_on_demand"`   // On-demand (wing/room filtered)
	L3 []Drawer `json:"l3_deep_search"` // Deep search results
}

// Message represents a chat message for conversation mining.
type Message struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"` // Message text
	Time    int64  `json:"time"`    // Unix timestamp
}

// SearchResult holds a drawer with its computed search score.
type SearchResult struct {
	Drawer Drawer  `json:"drawer"`
	Score  float64 `json:"score"`
}

// ─────────────────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────────────────

const (
	// Default half-life for recency decay (7 days in hours).
	DefaultHalfLifeHours = 7 * 24

	// Chunking constants for file mining.
	ChunkSize    = 800
	ChunkOverlap = 100
	MinChunkSize = 50

	// Layer 1 limits.
	MaxL1Drawers = 15
	MaxL1Chars   = 3200

	// L2 retrieval limit default.
	DefaultL2Limit = 10

	// L3 search limit default.
	DefaultL3Limit = 10

	// Category constants.
	CategoryFact       = "fact"
	CategoryDecision   = "decision"
	CategoryPreference = "preference"
	CategoryEntity     = "entity"
	CategoryMilestone  = "milestone"
	CategoryProblem    = "problem"
	CategoryOther      = "other"
	CategoryHandoff    = "handoff"
	CategorySkill      = "skill"
)

// Category keywords for auto-detection during file mining.
var categoryKeywords = map[string][]string{
	CategoryDecision:   {"decided", "chose", "picked", "switched", "migrated", "replaced", "trade-off", "approach", "we will", "let's use", "going with"},
	CategoryPreference: {"prefer", "like", "want", "always", "never", "should", "best practice", "style", "convention"},
	CategoryEntity:     {"person", "team", "company", "service", "library", "framework", "tool", "database", "api"},
	CategoryMilestone:  {"released", "deployed", "launched", "completed", "finished", "shipped", "v1", "v2", "beta", "production"},
	CategoryProblem:    {"bug", "issue", "broken", "failed", "crash", "error", "stuck", "workaround", "fix", "regression"},
}

// Room detection keywords for file mining.
var roomKeywords = map[string][]string{
	"auth":        {"auth", "oauth", "login", "session", "token", "password", "jwt", "sso", "permission"},
	"api":         {"api", "endpoint", "route", "handler", "rest", "graphql", "http", "request", "response"},
	"database":    {"database", "db", "sql", "query", "table", "schema", "migration", "postgres", "sqlite", "mongo"},
	"deployment":  {"deploy", "docker", "kubernetes", "k8s", "ci", "cd", "pipeline", "server", "infra", "cloud"},
	"frontend":    {"ui", "frontend", "component", "react", "vue", "html", "css", "style", "theme"},
	"backend":     {"backend", "server", "service", "middleware", "router", "controller", "logic"},
	"testing":     {"test", "spec", "assert", "mock", "coverage", "bench", "e2e", "integration"},
	"config":      {"config", "env", "setting", "yaml", "toml", "json", "flag", "option"},
	"performance": {"performance", "slow", "fast", "optimize", "cache", "latency", "throughput", "benchmark"},
	"security":    {"security", "vulnerability", "cve", "exploit", "sanitize", "validate", "encrypt"},
}

// Readable file extensions for mining.
var readableExtensions = map[string]bool{
	".txt": true, ".md": true, ".go": true, ".py": true, ".js": true,
	".ts": true, ".tsx": true, ".jsx": true, ".rs": true, ".rb": true,
	".java": true, ".sh": true, ".sql": true, ".json": true, ".yaml": true,
	".yml": true, ".toml": true, ".html": true, ".css": true, ".csv": true,
}

// Directories to skip during file scanning.
var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "__pycache__": true,
	".venv": true, "venv": true, "env": true, "dist": true,
	"build": true, ".next": true, "coverage": true, ".cache": true,
	"vendor": true, ".idea": true, ".vscode": true, "target": true,
}

// ─────────────────────────────────────────────────────────────
// Palace Construction & Lifecycle
// ─────────────────────────────────────────────────────────────

// NewPalace creates or loads a Palace from the given path.
// The path is the base directory (~/.ti/memory/). It creates the
// directory if it doesn't exist and loads identity.txt if present.
func NewPalace(path string) *Palace {
	p := &Palace{
		Path:     path,
		Wings:    make(map[string]*Wing),
		halfLife: DefaultHalfLifeHours,
	}
	// Load identity.txt if it exists
	identityPath := filepath.Join(path, "identity.txt")
	if data, err := os.ReadFile(identityPath); err == nil {
		p.Identity = strings.TrimSpace(string(data))
	}
	if p.Identity == "" {
		p.Identity = "No identity configured. Create identity.txt in the palace directory."
	}
	return p
}

// SetHalfLife sets the recency decay half-life in hours.
func (p *Palace) SetHalfLife(hours float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.halfLife = hours
}

// ─────────────────────────────────────────────────────────────
// SQLite Persistence
// ─────────────────────────────────────────────────────────────

// Load loads the palace from SQLite database. Creates the DB if it doesn't exist.
func (p *Palace) Load() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	dbPath := filepath.Join(p.Path, "memory.db")
	if err := os.MkdirAll(p.Path, 0755); err != nil {
		return fmt.Errorf("create palace dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}

	if err := p.initDB(db); err != nil {
		db.Close()
		return fmt.Errorf("init db: %w", err)
	}

	p.db = db

	// Load drawers from DB into in-memory structure
	if err := p.loadDrawers(); err != nil {
		return fmt.Errorf("load drawers: %w", err)
	}

	return nil
}

// Save persists all in-memory drawers to the SQLite database.
func (p *Palace) Save() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.db == nil {
		return fmt.Errorf("database not loaded — call Load() first")
	}

	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Clear existing data and re-insert
	if _, err := tx.Exec(`DELETE FROM drawers`); err != nil {
		return fmt.Errorf("clear drawers: %w", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO drawers (id, wing, room, text, importance, category, created_at, source_file)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer stmt.Close()

	for _, wing := range p.Wings {
		for _, room := range wing.Rooms {
			for _, d := range room.Drawers {
				if _, err := stmt.Exec(d.ID, wing.Name, room.Name, d.Text, d.Importance, d.Category, d.CreatedAt, d.SourceFile); err != nil {
					return fmt.Errorf("insert drawer %s: %w", d.ID, err)
				}
			}
		}
	}

	return tx.Commit()
}

// Close releases the database connection.
func (p *Palace) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *Palace) initDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS drawers (
		id         TEXT PRIMARY KEY,
		wing       TEXT NOT NULL,
		room       TEXT NOT NULL,
		text       TEXT NOT NULL,
		importance REAL NOT NULL DEFAULT 0.7,
		category   TEXT NOT NULL DEFAULT 'other',
		created_at INTEGER NOT NULL,
		source_file TEXT DEFAULT ''
	);
	CREATE INDEX IF NOT EXISTS idx_drawers_wing ON drawers(wing);
	CREATE INDEX IF NOT EXISTS idx_drawers_room ON drawers(room);
	CREATE INDEX IF NOT EXISTS idx_drawers_category ON drawers(category);
	CREATE INDEX IF NOT EXISTS idx_drawers_created ON drawers(created_at);
	`

	// Enable FTS5 for full-text search
	ftsSchema := `
	CREATE VIRTUAL TABLE IF NOT EXISTS drawers_fts USING fts5(
		text, wing, room, category, source_file,
		content='drawers',
		content_rowid='rowid'
	);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	// Try FTS5 creation — may fail if FTS5 not available
	if _, err := db.Exec(ftsSchema); err != nil {
		// FTS5 not available — that's OK, fall back to LIKE-based search
		// Don't fail, just note it
	}

	return nil
}

func (p *Palace) loadDrawers() error {
	rows, err := p.db.Query(`SELECT id, wing, room, text, importance, category, created_at, COALESCE(source_file, '') FROM drawers ORDER BY created_at DESC`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var d Drawer
		if err := rows.Scan(&d.ID, &d.Wing, &d.Room, &d.Text, &d.Importance, &d.Category, &d.CreatedAt, &d.SourceFile); err != nil {
			return err
		}

		// Ensure wing exists
		if _, ok := p.Wings[d.Wing]; !ok {
			p.Wings[d.Wing] = &Wing{Name: d.Wing, Rooms: make(map[string]*Room)}
		}
		// Ensure room exists
		if _, ok := p.Wings[d.Wing].Rooms[d.Room]; !ok {
			p.Wings[d.Wing].Rooms[d.Room] = &Room{Name: d.Room}
		}
		p.Wings[d.Wing].Rooms[d.Room].Drawers = append(p.Wings[d.Wing].Rooms[d.Room].Drawers, d)
	}

	return rows.Err()
}

// ─────────────────────────────────────────────────────────────
// Drawer Operations
// ─────────────────────────────────────────────────────────────

// AddDrawer adds a memory drawer to the specified wing and room.
// It updates both the in-memory structure and the database.
func (p *Palace) AddDrawer(wing, room string, d Drawer) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if d.ID == "" {
		d.ID = generateDrawerID(wing, room, d.Text, d.CreatedAt)
	}
	if d.CreatedAt == 0 {
		d.CreatedAt = time.Now().Unix()
	}
	if d.Importance == 0 {
		d.Importance = 0.7
	}
	if d.Category == "" {
		d.Category = CategoryOther
	}
	d.Wing = wing
	d.Room = room

	// Ensure wing exists
	if _, ok := p.Wings[wing]; !ok {
		p.Wings[wing] = &Wing{Name: wing, Rooms: make(map[string]*Room)}
	}
	// Ensure room exists
	if _, ok := p.Wings[wing].Rooms[room]; !ok {
		p.Wings[wing].Rooms[room] = &Room{Name: room}
	}
	p.Wings[wing].Rooms[room].Drawers = append(p.Wings[wing].Rooms[room].Drawers, d)

	// Persist to DB if loaded
	if p.db != nil {
		if _, err := p.db.Exec(
			`INSERT OR REPLACE INTO drawers (id, wing, room, text, importance, category, created_at, source_file)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			d.ID, wing, room, d.Text, d.Importance, d.Category, d.CreatedAt, d.SourceFile,
		); err != nil {
			return fmt.Errorf("persist drawer: %w", err)
		}
	}

	return nil
}

// DeleteDrawer removes a drawer by ID from both memory and database.
func (p *Palace) DeleteDrawer(wing, room, id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	w, wok := p.Wings[wing]
	if !wok {
		return nil
	}
	r, rok := w.Rooms[room]
	if !rok {
		return nil
	}

	// Remove from in-memory slice
	found := false
	for i, d := range r.Drawers {
		if d.ID == id {
			r.Drawers = append(r.Drawers[:i], r.Drawers[i+1:]...)
			found = true
			break
		}
	}

	// Remove from DB
	if p.db != nil {
		if _, err := p.db.Exec(`DELETE FROM drawers WHERE id = ?`, id); err != nil {
			return fmt.Errorf("delete from db: %w", err)
		}
	}

	if !found {
		return fmt.Errorf("drawer %s not found", id)
	}
	return nil
}

// GetDrawer retrieves a single drawer by wing, room, and ID.
func (p *Palace) GetDrawer(wing, room, id string) (*Drawer, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	w, ok := p.Wings[wing]
	if !ok {
		return nil, fmt.Errorf("wing %q not found", wing)
	}
	r, ok := w.Rooms[room]
	if !ok {
		return nil, fmt.Errorf("room %q not found in wing %q", room, wing)
	}
	for _, d := range r.Drawers {
		if d.ID == id {
			return &d, nil
		}
	}
	return nil, fmt.Errorf("drawer %q not found", id)
}

// UpdateDrawer modifies an existing drawer's text and importance.
func (p *Palace) UpdateDrawer(wing, room, id string, text string, importance float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	w, ok := p.Wings[wing]
	if !ok {
		return fmt.Errorf("wing %q not found", wing)
	}
	r, ok := w.Rooms[room]
	if !ok {
		return fmt.Errorf("room %q not found", room)
	}

	for i, d := range r.Drawers {
		if d.ID == id {
			r.Drawers[i].Text = text
			r.Drawers[i].Importance = importance

			if p.db != nil {
				if _, err := p.db.Exec(
					`UPDATE drawers SET text = ?, importance = ? WHERE id = ?`,
					text, importance, id,
				); err != nil {
					return fmt.Errorf("update drawer: %w", err)
				}
			}
			return nil
		}
	}
	return fmt.Errorf("drawer %q not found", id)
}

// CountDrawers returns the total number of drawers across all wings.
func (p *Palace) CountDrawers() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := 0
	for _, wing := range p.Wings {
		for _, room := range wing.Rooms {
			total += len(room.Drawers)
		}
	}
	return total
}

// ListWings returns all wing names sorted alphabetically.
func (p *Palace) ListWings() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	names := make([]string, 0, len(p.Wings))
	for name := range p.Wings {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ListRooms returns all room names within a wing, sorted alphabetically.
func (p *Palace) ListRooms(wing string) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	w, ok := p.Wings[wing]
	if !ok {
		return nil, fmt.Errorf("wing %q not found", wing)
	}

	names := make([]string, 0, len(w.Rooms))
	for name := range w.Rooms {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// GetTaxonomy returns the full wing → room → drawer count tree.
func (p *Palace) GetTaxonomy() map[string]map[string]int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	taxonomy := make(map[string]map[string]int)
	for wingName, wing := range p.Wings {
		taxonomy[wingName] = make(map[string]int)
		for roomName, room := range wing.Rooms {
			taxonomy[wingName][roomName] = len(room.Drawers)
		}
	}
	return taxonomy
}

// ─────────────────────────────────────────────────────────────
// 4-Layer Memory Stack
// ─────────────────────────────────────────────────────────────

// Recall returns L0 + L1 memory for a given wing — the essential
// story used for wake-up context.
func (p *Palace) Recall(wing string, limit int) *Stack {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if limit <= 0 {
		limit = MaxL1Drawers
	}

	stack := &Stack{
		L0: p.Identity,
	}

	// Collect all drawers for the wing (or all if wing is empty)
	var allDrawers []Drawer
	if wing != "" {
		if w, ok := p.Wings[wing]; ok {
			for _, room := range w.Rooms {
				allDrawers = append(allDrawers, room.Drawers...)
			}
		}
	} else {
		for _, w := range p.Wings {
			for _, room := range w.Rooms {
				allDrawers = append(allDrawers, room.Drawers...)
			}
		}
	}

	// Score each drawer: importance × recency
	scored := make([]SearchResult, 0, len(allDrawers))
	now := time.Now()
	for _, d := range allDrawers {
		recency := recencyDecay(d.CreatedAt, now.Unix(), p.halfLife)
		score := d.Importance*0.6 + recency*0.4
		scored = append(scored, SearchResult{Drawer: d, Score: score})
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	// Take top N, respecting char limit
	totalChars := 0
	for _, sr := range scored {
		if len(stack.L1) >= limit {
			break
		}
		if totalChars+utf8.RuneCountInString(sr.Drawer.Text) > MaxL1Chars {
			break
		}
		stack.L1 = append(stack.L1, sr.Drawer)
		totalChars += utf8.RuneCountInString(sr.Drawer.Text)
	}

	return stack
}

// OnDemand returns L2 filtered retrieval for a specific wing and room.
// If query is provided, it performs text-based filtering within the room.
func (p *Palace) OnDemand(wing, room, query string) *Stack {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stack := &Stack{
		L0: p.Identity,
	}

	limit := DefaultL2Limit

	if wing == "" && room == "" {
		return stack
	}

	var drawers []Drawer

	if wing != "" {
		if w, ok := p.Wings[wing]; ok {
			if room != "" {
				if r, ok := w.Rooms[room]; ok {
					drawers = r.Drawers
				}
			} else {
				// All rooms in the wing
				for _, r := range w.Rooms {
					drawers = append(drawers, r.Drawers...)
				}
			}
		}
	} else if room != "" {
		// Room across all wings
		for _, w := range p.Wings {
			if r, ok := w.Rooms[room]; ok {
				drawers = append(drawers, r.Drawers...)
			}
		}
	}

	// If query provided, filter by text match and score
	if query != "" {
		scored := scoreDrawersByText(drawers, query, p.halfLife, time.Now().Unix())
		if len(scored) > limit {
			scored = scored[:limit]
		}
		for _, sr := range scored {
			stack.L2 = append(stack.L2, sr.Drawer)
		}
	} else {
		// No query — just return first N drawers
		if len(drawers) > limit {
			drawers = drawers[:limit]
		}
		stack.L2 = drawers
	}

	return stack
}

// Search performs L3 deep semantic search across all or specified wings.
// Uses FTS5 if available, falls back to LIKE-based search with scoring.
func (p *Palace) Search(query string, wings []string, limit int) *Stack {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if limit <= 0 {
		limit = DefaultL3Limit
	}

	stack := &Stack{
		L0: p.Identity,
	}

	// Try FTS5 search first if DB is loaded
	if p.db != nil {
		ftsResults := p.searchFTS(query, wings, limit)
		if len(ftsResults) > 0 {
			stack.L3 = ftsResults
			return stack
		}
	}

	// Fall back to in-memory scoring
	var allDrawers []Drawer
	wingSet := make(map[string]bool)
	for _, w := range wings {
		wingSet[w] = true
	}

	for wingName, wing := range p.Wings {
		if len(wings) > 0 && !wingSet[wingName] {
			continue
		}
		for _, room := range wing.Rooms {
			allDrawers = append(allDrawers, room.Drawers...)
		}
	}

	scored := scoreDrawersByText(allDrawers, query, p.halfLife, time.Now().Unix())
	if len(scored) > limit {
		scored = scored[:limit]
	}
	for _, sr := range scored {
		stack.L3 = append(stack.L3, sr.Drawer)
	}

	return stack
}

// searchFTS attempts FTS5 full-text search.
func (p *Palace) searchFTS(query string, wings []string, limit int) []Drawer {
	// Check if FTS5 table exists
	var exists int
	err := p.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='drawers_fts'`).Scan(&exists)
	if err != nil || exists == 0 {
		return nil
	}

	queryStr := `SELECT d.id, d.wing, d.room, d.text, d.importance, d.category, d.created_at, COALESCE(d.source_file, '')
		FROM drawers d
		JOIN drawers_fts fts ON d.rowid = fts.rowid
		WHERE drawers_fts MATCH ?`

	args := []interface{}{query}

	if len(wings) > 0 {
		placeholders := make([]string, len(wings))
		for i, w := range wings {
			placeholders[i] = "?"
			args = append(args, w)
		}
		queryStr += ` AND d.wing IN (` + strings.Join(placeholders, ",") + `)`
	}

	queryStr += ` ORDER BY rank DESC LIMIT ?`
	args = append(args, limit)

	rows, err := p.db.Query(queryStr, args...)
	if err != nil {
		// FTS5 query failed — fall back
		return nil
	}
	defer rows.Close()

	var results []Drawer
	for rows.Next() {
		var d Drawer
		if err := rows.Scan(&d.ID, &d.Wing, &d.Room, &d.Text, &d.Importance, &d.Category, &d.CreatedAt, &d.SourceFile); err != nil {
			continue
		}
		results = append(results, d)
	}
	return results
}

// scoreDrawersByText scores drawers based on text similarity to query,
// importance, and recency decay.
func scoreDrawersByText(drawers []Drawer, query string, halfLife float64, nowUnix int64) []SearchResult {
	queryLower := strings.ToLower(query)
	queryTerms := strings.Fields(queryLower)

	scored := make([]SearchResult, 0, len(drawers))
	for _, d := range drawers {
		textLower := strings.ToLower(d.Text)

		// Text similarity: count query term occurrences
		var termScore float64
		for _, term := range queryTerms {
			count := strings.Count(textLower, term)
			termScore += float64(count)
		}
		// Normalize: 0-1 range using sigmoid-like function
		textSim := termScore / (termScore + 1)

		// Recency decay
		recency := recencyDecay(d.CreatedAt, nowUnix, halfLife)

		// Combined score: text_sim × importance × recency
		score := textSim * d.Importance * recency

		if score > 0 {
			scored = append(scored, SearchResult{Drawer: d, Score: score})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})
	return scored
}

// WakeUp returns L0 + L1 formatted as a context string for injection
// into AI system prompts.
func (p *Palace) WakeUp(wing string) string {
	stack := p.Recall(wing, MaxL1Drawers)

	var sb strings.Builder

	// L0: Identity
	sb.WriteString("## L0 — IDENTITY\n")
	sb.WriteString(stack.L0)
	sb.WriteString("\n\n")

	// L1: Essential Story
	sb.WriteString("## L1 — ESSENTIAL STORY\n")
	if len(stack.L1) == 0 {
		sb.WriteString("No memories yet.")
	} else {
		// Group by room
		byRoom := make(map[string][]Drawer)
		for _, d := range stack.L1 {
			byRoom[d.Room] = append(byRoom[d.Room], d)
		}

		// Sort room names for consistent output
		roomNames := make([]string, 0, len(byRoom))
		for name := range byRoom {
			roomNames = append(roomNames, name)
		}
		sort.Strings(roomNames)

		for _, roomName := range roomNames {
			sb.WriteString(fmt.Sprintf("\n[%s]\n", roomName))
			for _, d := range byRoom[roomName] {
				snippet := strings.ReplaceAll(strings.TrimSpace(d.Text), "\n", " ")
				if utf8.RuneCountInString(snippet) > 200 {
					snippet = snippet[:197] + "..."
				}
				sb.WriteString(fmt.Sprintf("  - %s", snippet))
				if d.SourceFile != "" {
					sb.WriteString(fmt.Sprintf("  (%s)", filepath.Base(d.SourceFile)))
				}
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

// ─────────────────────────────────────────────────────────────
// File Mining
// ─────────────────────────────────────────────────────────────

// MineFile parses a file into drawers by chunking its content.
// It auto-detects category and room from content and file path.
func (p *Palace) MineFile(filePath string) ([]Drawer, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if !readableExtensions[ext] && ext != "" {
		// Still try to process if no extension check passes
		// This allows processing files with unusual extensions
	}

	text := strings.TrimSpace(string(content))
	if len(text) < MinChunkSize {
		return nil, nil // Too small to mine
	}

	// Auto-detect room from file path and content
	room := detectRoomFromFile(filePath, text)

	// Chunk the text
	chunks := chunkText(text)

	// Create drawers
	var drawers []Drawer
	now := time.Now().Unix()
	for i, chunk := range chunks {
		category := detectCategory(chunk)
		d := Drawer{
			ID:         generateDrawerID("mined", room, chunk, now+int64(i)),
			Text:       chunk,
			Importance: 0.7,
			Category:   category,
			CreatedAt:  now + int64(i),
			SourceFile: filePath,
		}
		drawers = append(drawers, d)
	}

	return drawers, nil
}

// MineFiles processes multiple files and adds their drawers to the palace.
func (p *Palace) MineFiles(filePaths []string, wing string) error {
	for _, fp := range filePaths {
		drawers, err := p.MineFile(fp)
		if err != nil {
			return fmt.Errorf("mine file %s: %w", fp, err)
		}

		// Detect room from file path
		room := detectRoomFromFile(fp, "")

		for _, d := range drawers {
			if err := p.AddDrawer(wing, room, d); err != nil {
				return fmt.Errorf("add drawer: %w", err)
			}
		}
	}
	return nil
}

// chunkText splits text into overlapping chunks of ~ChunkSize chars.
func chunkText(text string) []string {
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return nil
	}

	var chunks []string
	start := 0

	for start < len(text) {
		end := start + ChunkSize
		if end > len(text) {
			end = len(text)
		} else if end < len(text) {
			// Try to break at paragraph boundary
			if nl := strings.LastIndex(text[start:end], "\n\n"); nl > ChunkSize/2 {
				end = start + nl
			} else if nl := strings.LastIndex(text[start:end], "\n"); nl > ChunkSize/2 {
				end = start + nl
			}
		}

		chunk := strings.TrimSpace(text[start:end])
		if utf8.RuneCountInString(chunk) >= MinChunkSize {
			chunks = append(chunks, chunk)
		}

		if end >= len(text) {
			break
		}
		start = end - ChunkOverlap
		if start < 0 {
			start = 0
		}
	}

	return chunks
}

// detectRoomFromFile determines the room name based on file path and content.
func detectRoomFromFile(filePath, content string) string {
	relPath := strings.ToLower(filePath)
	contentLower := strings.ToLower(content)
	if len(contentLower) > 2000 {
		contentLower = contentLower[:2000]
	}

	// Check path components first — exact room name match takes priority
	pathParts := strings.Split(relPath, string(filepath.Separator))
	knownRooms := []string{"auth", "api", "database", "deployment", "frontend", "backend", "testing", "config", "performance", "security"}
	for _, part := range pathParts {
		for _, room := range knownRooms {
			if part == room {
				return room
			}
		}
	}
	// Then check if part contains a room keyword
	for _, part := range pathParts {
		for _, room := range knownRooms {
			keywords := roomKeywords[room]
			for _, kw := range keywords {
				if strings.Contains(part, kw) || strings.Contains(kw, part) {
					return room
				}
			}
		}
	}

	// Score content against room keywords
	scores := make(map[string]int)
	for room, keywords := range roomKeywords {
		for _, kw := range keywords {
			if strings.Contains(contentLower, kw) {
				scores[room]++
			}
		}
	}

	bestRoom := "general"
	bestScore := 0
	for room, score := range scores {
		if score > bestScore {
			bestRoom = room
			bestScore = score
		}
	}

	return bestRoom
}

// detectCategory auto-detects the category of a text chunk.
func detectCategory(text string) string {
	textLower := strings.ToLower(text)
	scores := make(map[string]int)

	for category, keywords := range categoryKeywords {
		for _, kw := range keywords {
			if strings.Contains(textLower, kw) {
				scores[category]++
			}
		}
	}

	bestCat := CategoryOther
	bestScore := 0
	// Sort keys for deterministic tie-breaking
	categories := make([]string, 0, len(scores))
	for cat := range scores {
		categories = append(categories, cat)
	}
	sort.Strings(categories)
	for _, cat := range categories {
		score := scores[cat]
		if score > bestScore || (score == bestScore && cat < bestCat) {
			bestCat = cat
			bestScore = score
		}
	}

	return bestCat
}

// ─────────────────────────────────────────────────────────────
// Conversation Mining
// ─────────────────────────────────────────────────────────────

// MineConversation parses chat messages into drawers.
// Each user+assistant exchange pair becomes one drawer.
// Auto-detects decisions, preferences, facts, and entities.
func (p *Palace) MineConversation(messages []Message, wing string) ([]Drawer, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(messages) == 0 {
		return nil, nil
	}

	var drawers []Drawer
	var userBuffer string
	var turnStart int64

	for _, msg := range messages {
		if msg.Role == "user" {
			// Process previous exchange if exists
			if userBuffer != "" {
				// We have an orphan user message — store it
			}
			userBuffer = msg.Content
			turnStart = msg.Time
		} else if msg.Role == "assistant" && userBuffer != "" {
			// Complete exchange — create drawer
			exchange := fmt.Sprintf("> %s\n%s", userBuffer, msg.Content)
			category := detectCategoryFromConversation(userBuffer, msg.Content)
			room := detectRoomFromConversation(userBuffer, msg.Content)
			ts := turnStart
			if ts == 0 {
				ts = time.Now().Unix()
			}

			d := Drawer{
				ID:         generateDrawerID(wing, room, exchange, ts),
				Text:       exchange,
				Importance: 0.7,
				Category:   category,
				CreatedAt:  ts,
			}
			drawers = append(drawers, d)
			userBuffer = ""
		}
	}

	// Handle orphan user message
	if userBuffer != "" {
		room := detectRoomFromConversation(userBuffer, "")
		d := Drawer{
			ID:         generateDrawerID(wing, room, userBuffer, time.Now().Unix()),
			Text:       userBuffer,
			Importance: 0.5,
			Category:   CategoryOther,
			CreatedAt:  time.Now().Unix(),
		}
		drawers = append(drawers, d)
	}

	return drawers, nil
}

// detectCategoryFromConversation extracts category from a Q+A exchange.
func detectCategoryFromConversation(userMsg, assistantMsg string) string {
	combined := strings.ToLower(userMsg + " " + assistantMsg)

	// Check for decision patterns
	decisionSignals := []string{"decided", "chose", "we should", "let's use", "going with", "switched to", "agreed on"}
	for _, sig := range decisionSignals {
		if strings.Contains(combined, sig) {
			return CategoryDecision
		}
	}

	// Check for preference patterns
	prefSignals := []string{"prefer", "i like", "always use", "never use", "best way", "i want"}
	for _, sig := range prefSignals {
		if strings.Contains(combined, sig) {
			return CategoryPreference
		}
	}

	// Check for problem patterns
	probSignals := []string{"bug", "error", "failed", "broken", "fix", "issue", "problem", "debug"}
	for _, sig := range probSignals {
		if strings.Contains(combined, sig) {
			return CategoryProblem
		}
	}

	// Check for entity patterns
	entitySignals := []string{"is a", "uses", "built with", "version", "library", "framework"}
	for _, sig := range entitySignals {
		if strings.Contains(combined, sig) {
			return CategoryEntity
		}
	}

	return CategoryFact
}

// detectRoomFromConversation determines the room from conversation content.
func detectRoomFromConversation(userMsg, assistantMsg string) string {
	combined := strings.ToLower(userMsg + " " + assistantMsg)

	scores := make(map[string]int)
	for room, keywords := range roomKeywords {
		for _, kw := range keywords {
			if strings.Contains(combined, kw) {
				scores[room]++
			}
		}
	}

	bestRoom := "general"
	bestScore := 0
	for room, score := range scores {
		if score > bestScore {
			bestRoom = room
			bestScore = score
		}
	}

	return bestRoom
}

// ─────────────────────────────────────────────────────────────
// Merge & Export
// ─────────────────────────────────────────────────────────────

// Merge combines drawers from another palace into this one.
// Drawers with the same ID are updated with the other palace's data.
func (p *Palace) Merge(other *Palace) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for wingName, wing := range other.Wings {
		for roomName, room := range wing.Rooms {
			for _, d := range room.Drawers {
				// Check if drawer already exists
				exists := false
				if w, ok := p.Wings[wingName]; ok {
					if r, ok := w.Rooms[roomName]; ok {
						for i, existing := range r.Drawers {
							if existing.ID == d.ID {
								r.Drawers[i] = d // Update
								exists = true
								break
							}
						}
					}
				}

				if !exists {
					// Add new drawer
					if _, ok := p.Wings[wingName]; !ok {
						p.Wings[wingName] = &Wing{Name: wingName, Rooms: make(map[string]*Room)}
					}
					if _, ok := p.Wings[wingName].Rooms[roomName]; !ok {
						p.Wings[wingName].Rooms[roomName] = &Room{Name: roomName}
					}
					p.Wings[wingName].Rooms[roomName].Drawers = append(p.Wings[wingName].Rooms[roomName].Drawers, d)
				}

				// Persist to DB
				if p.db != nil {
					if _, err := p.db.Exec(
						`INSERT OR REPLACE INTO drawers (id, wing, room, text, importance, category, created_at, source_file)
						 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
						d.ID, wingName, roomName, d.Text, d.Importance, d.Category, d.CreatedAt, d.SourceFile,
					); err != nil {
						return fmt.Errorf("merge drawer %s: %w", d.ID, err)
					}
				}
			}
		}
	}

	return nil
}

// Export serializes the entire palace as JSON.
func (p *Palace) Export() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	type exportRoom struct {
		Name    string   `json:"name"`
		Drawers []Drawer `json:"drawers"`
	}
	type exportWing struct {
		Name  string                `json:"name"`
		Rooms map[string]exportRoom `json:"rooms"`
	}

	export := struct {
		Identity string                `json:"identity"`
		Wings    map[string]exportWing `json:"wings"`
	}{
		Identity: p.Identity,
		Wings:    make(map[string]exportWing),
	}

	for wingName, wing := range p.Wings {
		ew := exportWing{Name: wing.Name, Rooms: make(map[string]exportRoom)}
		for roomName, room := range wing.Rooms {
			ew.Rooms[roomName] = exportRoom{
				Name:    room.Name,
				Drawers: room.Drawers,
			}
		}
		export.Wings[wingName] = ew
	}

	return json.MarshalIndent(export, "", "  ")
}

// Import loads a palace from JSON export data.
func (p *Palace) Import(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	type importRoom struct {
		Name    string   `json:"name"`
		Drawers []Drawer `json:"drawers"`
	}
	type importWing struct {
		Name  string                `json:"name"`
		Rooms map[string]importRoom `json:"rooms"`
	}
	var importData struct {
		Identity string                `json:"identity"`
		Wings    map[string]importWing `json:"wings"`
	}

	if err := json.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	p.Identity = importData.Identity

	for wingName, iw := range importData.Wings {
		if _, ok := p.Wings[wingName]; !ok {
			p.Wings[wingName] = &Wing{Name: iw.Name, Rooms: make(map[string]*Room)}
		}
		for roomName, ir := range iw.Rooms {
			if _, ok := p.Wings[wingName].Rooms[roomName]; !ok {
				p.Wings[wingName].Rooms[roomName] = &Room{Name: ir.Name}
			}
			p.Wings[wingName].Rooms[roomName].Drawers = append(
				p.Wings[wingName].Rooms[roomName].Drawers,
				ir.Drawers...,
			)
		}
	}

	return nil
}

// ─────────────────────────────────────────────────────────────
// Utility Functions
// ─────────────────────────────────────────────────────────────

// generateDrawerID creates a deterministic ID from wing, room, text, and timestamp.
func generateDrawerID(wing, room, text string, createdAt int64) string {
	h := sha256.New()
	h.Write([]byte(wing))
	h.Write([]byte(room))
	h.Write([]byte(text))
	h.Write([]byte(strconv.FormatInt(createdAt, 10)))
	return "drawer_" + hex.EncodeToString(h.Sum(nil))[:32]
}

// recencyDecay computes temporal decay using exponential half-life.
// Returns a value between 0 and 1, where 1 is brand new.
func recencyDecay(createdAt, nowUnix int64, halfLifeHours float64) float64 {
	if halfLifeHours <= 0 {
		halfLifeHours = DefaultHalfLifeHours
	}
	hours := float64(nowUnix-createdAt) / 3600.0
	if hours < 0 {
		hours = 0
	}
	return math.Exp(-math.Ln2 * hours / halfLifeHours)
}

// FormatStack returns a human-readable string representation of a Stack.
func FormatStack(s *Stack) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("L0 Identity (%d chars)\n", utf8.RuneCountInString(s.L0)))
	sb.WriteString(s.L0)
	sb.WriteString("\n\n")

	sb.WriteString(fmt.Sprintf("L1 Essential Story (%d drawers)\n", len(s.L1)))
	for i, d := range s.L1 {
		sb.WriteString(fmt.Sprintf("  [%d] [%s] %s\n", i+1, d.Category, d.Text[:minInt(100, len(d.Text))]))
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("L2 On-Demand (%d drawers)\n", len(s.L2)))
	for i, d := range s.L2 {
		sb.WriteString(fmt.Sprintf("  [%d] [%s/%s] %s\n", i+1, d.Wing, d.Room, d.Text[:minInt(100, len(d.Text))]))
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("L3 Deep Search (%d results)\n", len(s.L3)))
	for i, d := range s.L3 {
		sb.WriteString(fmt.Sprintf("  [%d] [%s/%s] %s\n", i+1, d.Wing, d.Room, d.Text[:minInt(100, len(d.Text))]))
	}

	return sb.String()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ─────────────────────────────────────────────────────────────
// Exported test helpers (used by memory_test.go)
// ─────────────────────────────────────────────────────────────

// ChunkTextForTest exposes chunkText for testing.
func ChunkTextForTest(text string) []string {
	return chunkText(text)
}

// MinChunkSizeForTest exposes MinChunkSize for testing.
func MinChunkSizeForTest() int {
	return MinChunkSize
}

// DetectCategoryForTest exposes detectCategory for testing.
func DetectCategoryForTest(text string) string {
	return detectCategory(text)
}

// DetectRoomForTest exposes detectRoomFromFile for testing.
func DetectRoomForTest(filePath, content string) string {
	return detectRoomFromFile(filePath, content)
}

// DetectCategoryFromConversationForTest exposes detectCategoryFromConversation for testing.
func DetectCategoryFromConversationForTest(userMsg, assistantMsg string) string {
	return detectCategoryFromConversation(userMsg, assistantMsg)
}

// RecencyDecayForTest exposes recencyDecay for testing.
func RecencyDecayForTest(createdAt, nowUnix int64, halfLifeHours float64) float64 {
	return recencyDecay(createdAt, nowUnix, halfLifeHours)
}

// GenerateDrawerIDForTest exposes generateDrawerID for testing.
func GenerateDrawerIDForTest(wing, room, text string, createdAt int64) string {
	return generateDrawerID(wing, room, text, createdAt)
}

// StackToContext converts a Stack to a context string for AI injection.
func StackToContext(s *Stack) string {
	var sb strings.Builder

	sb.WriteString(s.L0)
	sb.WriteString("\n\n")

	if len(s.L1) > 0 {
		sb.WriteString("## Recent Memories\n")
		for _, d := range s.L1 {
			snippet := strings.ReplaceAll(strings.TrimSpace(d.Text), "\n", " ")
			sb.WriteString(fmt.Sprintf("- [%s] %s\n", d.Category, snippet))
		}
		sb.WriteString("\n")
	}

	if len(s.L2) > 0 {
		sb.WriteString("## Topic Memories\n")
		for _, d := range s.L2 {
			snippet := strings.ReplaceAll(strings.TrimSpace(d.Text), "\n", " ")
			sb.WriteString(fmt.Sprintf("- [%s/%s] %s\n", d.Wing, d.Room, snippet))
		}
		sb.WriteString("\n")
	}

	if len(s.L3) > 0 {
		sb.WriteString("## Search Results\n")
		for _, d := range s.L3 {
			snippet := strings.ReplaceAll(strings.TrimSpace(d.Text), "\n", " ")
			sb.WriteString(fmt.Sprintf("- [%s/%s] %s\n", d.Wing, d.Room, snippet))
		}
	}

	return sb.String()
}

// ─────────────────────────────────────────────────────────────
// Handoff Storage
// ─────────────────────────────────────────────────────────────

// HandoffBundle represents a handoff between agents.
type HandoffBundle struct {
	FromAgent string            `json:"from_agent"`
	ToAgent   string            `json:"to_agent"`
	Context   string            `json:"context"`
	Output    string            `json:"output"`
	Timestamp int64             `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// CreateHandoff stores a handoff bundle in the palace.
// Returns the handoff ID.
func (p *Palace) CreateHandoff(fromAgent, toAgent, context, output string, metadata map[string]string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Generate handoff ID
	timestamp := time.Now().Unix()
	id := generateDrawerID("handoffs", fromAgent+"-"+toAgent, output, timestamp)

	// Serialize metadata
	var metadataStr string
	if metadata != nil {
		if data, err := json.Marshal(metadata); err == nil {
			metadataStr = string(data)
		}
	}

	// Create handoff text (output + metadata context)
	handoffText := output
	if metadataStr != "" {
		handoffText = fmt.Sprintf("%s\n\nMetadata: %s", output, metadataStr)
	}

	// Create drawer
	d := Drawer{
		ID:         id,
		Text:       handoffText,
		Importance: 0.9, // High importance for handoffs
		Category:   CategoryHandoff,
		CreatedAt:  timestamp,
		Wing:       "handoffs",
		Room:       fromAgent + "-" + toAgent,
	}

	// Ensure wing exists
	if _, ok := p.Wings["handoffs"]; !ok {
		p.Wings["handoffs"] = &Wing{Name: "handoffs", Rooms: make(map[string]*Room)}
	}

	// Ensure room exists
	roomName := fromAgent + "-" + toAgent
	if _, ok := p.Wings["handoffs"].Rooms[roomName]; !ok {
		p.Wings["handoffs"].Rooms[roomName] = &Room{Name: roomName}
	}

	// Add drawer
	p.Wings["handoffs"].Rooms[roomName].Drawers = append(p.Wings["handoffs"].Rooms[roomName].Drawers, d)

	// Persist to DB if loaded
	if p.db != nil {
		if _, err := p.db.Exec(
			`INSERT OR REPLACE INTO drawers (id, wing, room, text, importance, category, created_at, source_file)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			id, "handoffs", roomName, handoffText, 0.9, CategoryHandoff, timestamp, "",
		); err != nil {
			return "", fmt.Errorf("persist handoff: %w", err)
		}
	}

	return id, nil
}

// RecallHandoff retrieves the most recent handoff between two agents.
func (p *Palace) RecallHandoff(fromAgent, toAgent string) (*HandoffBundle, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	roomName := fromAgent + "-" + toAgent
	w, ok := p.Wings["handoffs"]
	if !ok {
		return nil, fmt.Errorf("no handoffs wing found")
	}

	r, ok := w.Rooms[roomName]
	if !ok {
		return nil, fmt.Errorf("no handoff from %s to %s found", fromAgent, toAgent)
	}

	if len(r.Drawers) == 0 {
		return nil, fmt.Errorf("no handoffs from %s to %s", fromAgent, toAgent)
	}

	// Get most recent handoff (last in slice)
	d := r.Drawers[len(r.Drawers)-1]

	// Parse metadata if present
	var metadata map[string]string
	metadataPrefix := "\n\nMetadata: "
	if idx := strings.Index(d.Text, metadataPrefix); idx > 0 {
		metadataStr := strings.TrimSpace(d.Text[idx+len(metadataPrefix):])
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err == nil {
			metadata = metadata
		}
		// Strip metadata from output
		d.Text = d.Text[:idx]
	}

	return &HandoffBundle{
		FromAgent: fromAgent,
		ToAgent:   toAgent,
		Output:    d.Text,
		Timestamp: d.CreatedAt,
		Metadata:  metadata,
	}, nil
}

// ListHandoffs returns all handoffs involving a specific agent.
// If agent is empty, returns all handoffs.
func (p *Palace) ListHandoffs(agent string) []HandoffBundle {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var bundles []HandoffBundle

	w, ok := p.Wings["handoffs"]
	if !ok {
		return bundles
	}

	for roomName, room := range w.Rooms {
		// Filter by agent if specified
		if agent != "" {
			if !strings.Contains(roomName, agent) {
				continue
			}
		}

		for _, d := range room.Drawers {
			// Parse room name to get from/to agents
			parts := strings.Split(roomName, "-")
			if len(parts) != 2 {
				continue
			}

			fromAgent := parts[0]
			toAgent := parts[1]

			// Parse metadata
			var metadata map[string]string
			metadataPrefix := "\n\nMetadata: "
			output := d.Text
			if idx := strings.Index(d.Text, metadataPrefix); idx > 0 {
				metadataStr := strings.TrimSpace(d.Text[idx+len(metadataPrefix):])
				if err := json.Unmarshal([]byte(metadataStr), &metadata); err == nil {
					metadata = metadata
				}
				output = d.Text[:idx]
			}

			bundles = append(bundles, HandoffBundle{
				FromAgent: fromAgent,
				ToAgent:   toAgent,
				Output:    output,
				Timestamp: d.CreatedAt,
				Metadata:  metadata,
			})
		}
	}

	// Sort by timestamp descending
	sort.Slice(bundles, func(i, j int) bool {
		return bundles[i].Timestamp > bundles[j].Timestamp
	})

	return bundles
}

// ─────────────────────────────────────────────────────────────
// Skill Storage
// ─────────────────────────────────────────────────────────────

// StoreSkill stores a skill in the palace.
// Skills are stored in the "skills" wing with room based on skill category.
func (p *Palace) StoreSkill(name, content, category string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Generate skill ID
	timestamp := time.Now().Unix()
	id := generateDrawerID("skills", name, content, timestamp)

	// Determine room (use category or "general")
	roomName := "general"
	if category != "" {
		roomName = category
	}

	// Create drawer
	d := Drawer{
		ID:         id,
		Text:       content,
		Importance: 0.8, // High importance for skills
		Category:   CategorySkill,
		CreatedAt:  timestamp,
		Wing:       "skills",
		Room:       roomName,
	}

	// Ensure wing exists
	if _, ok := p.Wings["skills"]; !ok {
		p.Wings["skills"] = &Wing{Name: "skills", Rooms: make(map[string]*Room)}
	}

	// Ensure room exists
	if _, ok := p.Wings["skills"].Rooms[roomName]; !ok {
		p.Wings["skills"].Rooms[roomName] = &Room{Name: roomName}
	}

	// Add drawer
	p.Wings["skills"].Rooms[roomName].Drawers = append(p.Wings["skills"].Rooms[roomName].Drawers, d)

	// Persist to DB if loaded
	if p.db != nil {
		if _, err := p.db.Exec(
			`INSERT OR REPLACE INTO drawers (id, wing, room, text, importance, category, created_at, source_file)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			id, "skills", roomName, content, 0.8, CategorySkill, timestamp, "",
		); err != nil {
			return fmt.Errorf("persist skill: %w", err)
		}
	}

	return nil
}

// GetSkill retrieves a skill by name.
// Searches across all skill rooms.
func (p *Palace) GetSkill(name string) (*Drawer, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	w, ok := p.Wings["skills"]
	if !ok {
		return nil, fmt.Errorf("no skills wing found")
	}

	// Search all rooms for skill by name (text contains name)
	for _, room := range w.Rooms {
		for _, d := range room.Drawers {
			// Check if drawer text contains the skill name
			// Skills are typically stored with the name in the first line
			firstLine := strings.Split(d.Text, "\n")[0]
			if strings.Contains(firstLine, name) || strings.Contains(d.Text, name) {
				return &d, nil
			}
		}
	}

	return nil, fmt.Errorf("skill %q not found", name)
}

// ListSkills returns all skills, optionally filtered by category/room.
func (p *Palace) ListSkills(category string) []Drawer {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var skills []Drawer

	w, ok := p.Wings["skills"]
	if !ok {
		return skills
	}

	// Filter by category if specified
	roomName := "general"
	if category != "" {
		roomName = category
	}

	if category == "" {
		// Return all skills from all rooms
		for _, room := range w.Rooms {
			skills = append(skills, room.Drawers...)
		}
	} else {
		// Return skills from specific room
		if room, ok := w.Rooms[roomName]; ok {
			skills = append(skills, room.Drawers...)
		}
	}

	// Sort by importance descending
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Importance > skills[j].Importance
	})

	return skills
}

// DeleteSkill removes a skill by name.
func (p *Palace) DeleteSkill(name string) error {
	p.mu.Lock()
	defer p.mu.RUnlock()

	w, ok := p.Wings["skills"]
	if !ok {
		return fmt.Errorf("no skills wing found")
	}

	// Search and delete
	for roomName, room := range w.Rooms {
		for i, d := range room.Drawers {
			firstLine := strings.Split(d.Text, "\n")[0]
			if strings.Contains(firstLine, name) || strings.Contains(d.Text, name) {
				// Remove from in-memory slice
				room.Drawers = append(room.Drawers[:i], room.Drawers[i+1:]...)

				// Remove from DB
				if p.db != nil {
					if _, err := p.db.Exec(`DELETE FROM drawers WHERE id = ?`, d.ID); err != nil {
						return fmt.Errorf("delete skill from db: %w", err)
					}
				}

				// Clean up empty rooms
				if len(room.Drawers) == 0 {
					delete(w.Rooms, roomName)
				}

				return nil
			}
		}
	}

	return fmt.Errorf("skill %q not found", name)
}
