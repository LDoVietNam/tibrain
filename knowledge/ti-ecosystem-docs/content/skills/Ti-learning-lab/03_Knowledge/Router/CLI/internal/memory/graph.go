// Package memory — Knowledge Graph component: SQLite-backed temporal
// knowledge graph storing subject-predicate-object triples with
// validity windows, entity registry, and BFS path-finding.
//
// Inspired by MemPalace's knowledge_graph.py but written natively in Go.
package memory

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// ─────────────────────────────────────────────────────────────
// Core Data Types
// ─────────────────────────────────────────────────────────────

// Graph represents a temporal knowledge graph backed by SQLite.
type Graph struct {
	db       *sql.DB
	mu       sync.RWMutex
	halfLife time.Duration
}

// Triple represents a subject-predicate-object relationship.
type Triple struct {
	ID         string  `json:"id"`
	Subject    string  `json:"subject"`
	Predicate  string  `json:"predicate"`
	Object     string  `json:"object"`
	Confidence float64 `json:"confidence"`
	ValidFrom  string  `json:"valid_from"` // "2026-01-01" or empty
	ValidTo    string  `json:"valid_to"`   // "2026-12-31" or empty (ongoing)
	Source     string  `json:"source"`     // source drawer ID or file
	CreatedAt  int64   `json:"created_at"`
}

// Entity represents a named thing in the graph.
type Entity struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`       // person|project|task|event|concept|location|organization
	Aliases    []string          `json:"aliases"`    // alternative names
	Attributes map[string]string `json:"attributes"` // key-value metadata
	CreatedAt  int64             `json:"created_at"`
}

// ─────────────────────────────────────────────────────────────
// Construction & Init
// ─────────────────────────────────────────────────────────────

// NewGraph creates a Graph on an existing SQLite connection.
// If db is nil it returns nil — the caller must supply an open DB.
func NewGraph(db *sql.DB) *Graph {
	if db == nil {
		return nil
	}
	g := &Graph{
		db:       db,
		halfLife: 7 * 24 * time.Hour, // 7 days default
	}
	return g
}

// InitDB creates the entities, attributes, and triples tables if they
// do not already exist. Safe to call multiple times.
func (g *Graph) InitDB() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	schema := `
	CREATE TABLE IF NOT EXISTS entities (
		id         TEXT PRIMARY KEY,
		name       TEXT NOT NULL UNIQUE,
		type       TEXT NOT NULL DEFAULT 'concept',
		aliases    TEXT DEFAULT '[]',
		created_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS attributes (
		id        INTEGER PRIMARY KEY AUTOINCREMENT,
		entity_id TEXT NOT NULL REFERENCES entities(id),
		key       TEXT NOT NULL,
		value     TEXT NOT NULL,
		UNIQUE(entity_id, key)
	);

	CREATE TABLE IF NOT EXISTS triples (
		id         TEXT PRIMARY KEY,
		subject    TEXT NOT NULL,
		predicate  TEXT NOT NULL,
		object     TEXT NOT NULL,
		confidence REAL NOT NULL DEFAULT 1.0,
		valid_from TEXT DEFAULT '',
		valid_to   TEXT DEFAULT '',
		source     TEXT DEFAULT '',
		created_at INTEGER NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_triples_subject   ON triples(subject);
	CREATE INDEX IF NOT EXISTS idx_triples_predicate ON triples(predicate);
	CREATE INDEX IF NOT EXISTS idx_triples_object    ON triples(object);
	CREATE INDEX IF NOT EXISTS idx_triples_valid_from ON triples(valid_from);
	CREATE INDEX IF NOT EXISTS idx_triples_valid_to   ON triples(valid_to);
	`

	if _, err := g.db.Exec(schema); err != nil {
		return fmt.Errorf("create graph schema: %w", err)
	}
	return nil
}

// SetHalfLife sets the decay half-life for confidence scoring.
func (g *Graph) SetHalfLife(d time.Duration) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.halfLife = d
}

// ─────────────────────────────────────────────────────────────
// Entity CRUD
// ─────────────────────────────────────────────────────────────

// AddEntity inserts or updates an entity in the graph.
// If the entity ID is empty one is generated from the name.
func (g *Graph) AddEntity(e Entity) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if e.ID == "" {
		e.ID = generateID("ent_" + e.Name)
	}
	if e.Type == "" {
		e.Type = DetectEntity(e.Name)
	}
	if e.Aliases == nil {
		e.Aliases = []string{}
	}

	aliasesJSON, err := json.Marshal(e.Aliases)
	if err != nil {
		return fmt.Errorf("marshal aliases: %w", err)
	}

	now := time.Now().Unix()

	result, err := g.db.Exec(
		`INSERT INTO entities (id, name, type, aliases, created_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(name) DO UPDATE SET
			type = excluded.type,
			aliases = excluded.aliases`,
		e.ID, e.Name, e.Type, string(aliasesJSON), now,
	)
	if err != nil {
		return fmt.Errorf("add entity: %w", err)
	}

	// Store attributes if any
	if len(e.Attributes) > 0 {
		for k, v := range e.Attributes {
			if _, err := g.db.Exec(
				`INSERT INTO attributes (entity_id, key, value) VALUES (?, ?, ?)
				 ON CONFLICT(entity_id, key) DO UPDATE SET value = excluded.value`,
				e.ID, k, v,
			); err != nil {
				return fmt.Errorf("set attribute %s: %w", k, err)
			}
		}
	}

	_ = result // suppress unused variable warning

	return nil
}

// GetEntity looks up an entity by name (case-insensitive).
func (g *Graph) GetEntity(name string) (*Entity, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var id, entityType, aliasesJSON string
	var createdAt int64

	err := g.db.QueryRow(
		`SELECT id, name, type, aliases, created_at FROM entities WHERE LOWER(name) = LOWER(?)`,
		name,
	).Scan(&id, &name, &entityType, &aliasesJSON, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("entity %q not found", name)
		}
		return nil, fmt.Errorf("get entity: %w", err)
	}

	var aliases []string
	if err := json.Unmarshal([]byte(aliasesJSON), &aliases); err != nil {
		aliases = []string{}
	}

	e := &Entity{
		ID:         id,
		Name:       name,
		Type:       entityType,
		Aliases:    aliases,
		Attributes: make(map[string]string),
	}

	// Load attributes
	attrRows, err := g.db.Query(`SELECT key, value FROM attributes WHERE entity_id = ?`, id)
	if err == nil {
		defer attrRows.Close()
		for attrRows.Next() {
			var k, v string
			if err := attrRows.Scan(&k, &v); err == nil {
				e.Attributes[k] = v
			}
		}
	}

	return e, nil
}

// FindEntities returns all entities matching a given type.
// Pass empty string to return all entities.
func (g *Graph) FindEntities(entityType string) ([]Entity, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var query string
	var args []interface{}

	if entityType != "" {
		query = `SELECT id, name, type, aliases, created_at FROM entities WHERE type = ? ORDER BY created_at DESC`
		args = append(args, entityType)
	} else {
		query = `SELECT id, name, type, aliases, created_at FROM entities ORDER BY created_at DESC`
	}

	rows, err := g.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("find entities: %w", err)
	}
	defer rows.Close()

	var entities []Entity
	for rows.Next() {
		var e Entity
		var aliasesJSON string
		e.Attributes = make(map[string]string)

		if err := rows.Scan(&e.ID, &e.Name, &e.Type, &aliasesJSON, &e.CreatedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(aliasesJSON), &e.Aliases); err != nil {
			e.Aliases = []string{}
		}

		// Load attributes
		attrRows, aErr := g.db.Query(`SELECT key, value FROM attributes WHERE entity_id = ?`, e.ID)
		if aErr == nil {
			for attrRows.Next() {
				var k, v string
				if attrRows.Scan(&k, &v) == nil {
					e.Attributes[k] = v
				}
			}
			attrRows.Close()
		}

		entities = append(entities, e)
	}
	return entities, rows.Err()
}

// AddAlias adds an alternative name for an entity.
func (g *Graph) AddAlias(entityName, alias string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Check entity exists
	var id, aliasesJSON string
	err := g.db.QueryRow(
		`SELECT id, aliases FROM entities WHERE LOWER(name) = LOWER(?)`,
		entityName,
	).Scan(&id, &aliasesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("entity %q not found", entityName)
		}
		return fmt.Errorf("lookup entity: %w", err)
	}

	var aliases []string
	if err := json.Unmarshal([]byte(aliasesJSON), &aliases); err != nil {
		aliases = []string{}
	}

	// Check if alias already exists
	for _, a := range aliases {
		if strings.EqualFold(a, alias) {
			return nil // idempotent
		}
	}

	aliases = append(aliases, alias)
	aliasesBytes, err := json.Marshal(aliases)
	if err != nil {
		return fmt.Errorf("marshal aliases: %w", err)
	}

	if _, err := g.db.Exec(
		`UPDATE entities SET aliases = ? WHERE id = ?`,
		string(aliasesBytes), id,
	); err != nil {
		return fmt.Errorf("update aliases: %w", err)
	}
	return nil
}

// SetAttribute adds or updates a key-value attribute on an entity.
func (g *Graph) SetAttribute(entityName, key, value string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	var id string
	err := g.db.QueryRow(
		`SELECT id FROM entities WHERE LOWER(name) = LOWER(?)`,
		entityName,
	).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("entity %q not found", entityName)
		}
		return fmt.Errorf("lookup entity: %w", err)
	}

	if _, err := g.db.Exec(
		`INSERT INTO attributes (entity_id, key, value) VALUES (?, ?, ?)
		 ON CONFLICT(entity_id, key) DO UPDATE SET value = excluded.value`,
		id, key, value,
	); err != nil {
		return fmt.Errorf("set attribute: %w", err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────
// Triple CRUD
// ─────────────────────────────────────────────────────────────

// AddTriple inserts a relationship triple.
func (g *Graph) AddTriple(t Triple) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if t.ID == "" {
		t.ID = generateTripleID(t.Subject, t.Predicate, t.Object)
	}
	if t.Confidence <= 0 || t.Confidence > 1 {
		t.Confidence = 1.0
	}
	if t.CreatedAt == 0 {
		t.CreatedAt = time.Now().Unix()
	}

	_, err := g.db.Exec(
		`INSERT OR REPLACE INTO triples (id, subject, predicate, object, confidence, valid_from, valid_to, source, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.Subject, t.Predicate, t.Object, t.Confidence, t.ValidFrom, t.ValidTo, t.Source, t.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("add triple: %w", err)
	}
	return nil
}

// DeleteTriple removes a triple by ID.
func (g *Graph) DeleteTriple(id string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	result, err := g.db.Exec(`DELETE FROM triples WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete triple: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil
	}
	if affected == 0 {
		return fmt.Errorf("triple %q not found", id)
	}
	return nil
}

// Query finds triples matching the given fields.
// Empty string means wildcard (match all).
func (g *Graph) Query(subject, predicate, object string) ([]Triple, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	query := `SELECT id, subject, predicate, object, confidence, valid_from, valid_to, source, created_at FROM triples WHERE 1=1`
	var args []interface{}

	if subject != "" {
		query += ` AND subject = ?`
		args = append(args, subject)
	}
	if predicate != "" {
		query += ` AND predicate = ?`
		args = append(args, predicate)
	}
	if object != "" {
		query += ` AND object = ?`
		args = append(args, object)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := g.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query triples: %w", err)
	}
	defer rows.Close()

	return scanTriples(rows)
}

// QueryValid returns all triples that are valid at the given time.
// now should be in "YYYY-MM-DD" format.
func (g *Graph) QueryValid(now string) ([]Triple, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	query := `SELECT id, subject, predicate, object, confidence, valid_from, valid_to, source, created_at FROM triples
		WHERE (valid_from = '' OR valid_from <= ?)
		  AND (valid_to   = '' OR valid_to   >= ?)
		ORDER BY confidence DESC, created_at DESC`

	rows, err := g.db.Query(query, now, now)
	if err != nil {
		return nil, fmt.Errorf("query valid triples: %w", err)
	}
	defer rows.Close()

	return scanTriples(rows)
}

// GetRelated returns all triples where the entity appears as subject or object.
func (g *Graph) GetRelated(entityName string) ([]Triple, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	rows, err := g.db.Query(
		`SELECT id, subject, predicate, object, confidence, valid_from, valid_to, source, created_at
		 FROM triples WHERE subject = ? OR object = ?
		 ORDER BY created_at DESC`,
		entityName, entityName,
	)
	if err != nil {
		return nil, fmt.Errorf("get related triples: %w", err)
	}
	defer rows.Close()

	return scanTriples(rows)
}

// scanTriples reads all rows from a sql.Rows into []Triple.
func scanTriples(rows *sql.Rows) ([]Triple, error) {
	var triples []Triple
	for rows.Next() {
		var t Triple
		if err := rows.Scan(&t.ID, &t.Subject, &t.Predicate, &t.Object, &t.Confidence, &t.ValidFrom, &t.ValidTo, &t.Source, &t.CreatedAt); err != nil {
			return nil, err
		}
		triples = append(triples, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return triples, nil
}

// ─────────────────────────────────────────────────────────────
// BFS Path Finding
// ─────────────────────────────────────────────────────────────

// GetPath finds all shortest paths between two entities using BFS.
// Returns a list of paths, each path being a slice of entity names
// [start, ..., end]. Returns nil if no path exists within maxHops.
func (g *Graph) GetPath(start, end string, maxHops int) ([][]string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if maxHops <= 0 {
		return nil, nil
	}

	// Build adjacency map from triples
	adj, err := g.buildAdjacency()
	if err != nil {
		return nil, err
	}

	// Check start/end exist
	if _, ok := adj[start]; !ok {
		// Not in adjacency — maybe isolated entity, no paths possible
		return nil, nil
	}
	if _, ok := adj[end]; !ok {
		return nil, nil
	}

	// BFS collecting all shortest paths
	// queue entries: current node + path so far
	type queueEntry struct {
		node string
		path []string
	}

	queue := []queueEntry{{node: start, path: []string{start}}}
	visited := map[string]int{start: 0} // node -> depth at which first visited
	var allPaths [][]string
	found := false
	currentDepth := 0

	for len(queue) > 0 {
		entry := queue[0]
		queue = queue[1:]

		depth := len(entry.path) - 1
		if depth > currentDepth {
			currentDepth = depth
			if found {
				// Already found paths at previous depth, stop
				break
			}
		}
		if depth > maxHops {
			break
		}

		if entry.node == end {
			found = true
			allPaths = append(allPaths, entry.path)
			continue
		}

		for neighbor := range adj[entry.node] {
			// Allow revisiting at same depth (for multiple paths) but not deeper
			prevDepth, wasVisited := visited[neighbor]
			if wasVisited && prevDepth < depth+1 {
				continue
			}
			if depth+1 > maxHops {
				continue
			}

			visited[neighbor] = depth + 1
			newPath := make([]string, len(entry.path)+1)
			copy(newPath, entry.path)
			newPath[len(entry.path)] = neighbor
			queue = append(queue, queueEntry{node: neighbor, path: newPath})
		}
	}

	if len(allPaths) == 0 {
		return nil, nil
	}
	return allPaths, nil
}

// buildAdjacency builds an undirected adjacency map from all triples.
func (g *Graph) buildAdjacency() (map[string]map[string]bool, error) {
	rows, err := g.db.Query(`SELECT DISTINCT subject, object FROM triples`)
	if err != nil {
		return nil, fmt.Errorf("build adjacency: %w", err)
	}
	defer rows.Close()

	adj := make(map[string]map[string]bool)
	for rows.Next() {
		var subj, obj string
		if err := rows.Scan(&subj, &obj); err != nil {
			continue
		}
		if adj[subj] == nil {
			adj[subj] = make(map[string]bool)
		}
		if adj[obj] == nil {
			adj[obj] = make(map[string]bool)
		}
		adj[subj][obj] = true
		adj[obj][subj] = true
	}
	return adj, rows.Err()
}

// ─────────────────────────────────────────────────────────────
// Stats
// ─────────────────────────────────────────────────────────────

// Stats returns counts of entities by type and triples by predicate.
func (g *Graph) Stats() (map[string]int, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	stats := make(map[string]int)

	// Entity count by type
	entityRows, err := g.db.Query(`SELECT type, COUNT(*) FROM entities GROUP BY type`)
	if err != nil {
		return nil, fmt.Errorf("stats entities: %w", err)
	}
	for entityRows.Next() {
		var t string
		var c int
		if err := entityRows.Scan(&t, &c); err != nil {
			entityRows.Close()
			return nil, err
		}
		stats["entity_type:"+t] = c
	}
	entityRows.Close()

	// Total entity count
	var totalEntities int
	if err := g.db.QueryRow(`SELECT COUNT(*) FROM entities`).Scan(&totalEntities); err == nil {
		stats["total_entities"] = totalEntities
	}

	// Triple count by predicate
	tripleRows, err := g.db.Query(`SELECT predicate, COUNT(*) FROM triples GROUP BY predicate`)
	if err != nil {
		return nil, fmt.Errorf("stats triples: %w", err)
	}
	for tripleRows.Next() {
		var p string
		var c int
		if err := tripleRows.Scan(&p, &c); err != nil {
			tripleRows.Close()
			return nil, err
		}
		stats["predicate:"+p] = c
	}
	tripleRows.Close()

	// Total triple count
	var totalTriples int
	if err := g.db.QueryRow(`SELECT COUNT(*) FROM triples`).Scan(&totalTriples); err == nil {
		stats["total_triples"] = totalTriples
	}

	return stats, nil
}

// ─────────────────────────────────────────────────────────────
// Export / Import
// ─────────────────────────────────────────────────────────────

// GraphExport is the JSON-serializable representation of the graph.
type GraphExport struct {
	Entities []Entity `json:"entities"`
	Triples  []Triple `json:"triples"`
}

// Export serializes the entire graph to JSON.
func (g *Graph) Export() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	data := GraphExport{
		Entities: []Entity{},
		Triples:  []Triple{},
	}

	// Export entities with attributes
	entityRows, err := g.db.Query(`SELECT id, name, type, aliases, created_at FROM entities ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("export entities: %w", err)
	}

	var entities []Entity
	for entityRows.Next() {
		var e Entity
		var aliasesJSON string
		e.Attributes = make(map[string]string)
		if err := entityRows.Scan(&e.ID, &e.Name, &e.Type, &aliasesJSON, &e.CreatedAt); err != nil {
			entityRows.Close()
			return nil, err
		}
		if err := json.Unmarshal([]byte(aliasesJSON), &e.Aliases); err != nil {
			e.Aliases = []string{}
		}

		// Load attributes
		attrRows, aErr := g.db.Query(`SELECT key, value FROM attributes WHERE entity_id = ?`, e.ID)
		if aErr == nil {
			for attrRows.Next() {
				var k, v string
				if attrRows.Scan(&k, &v) == nil {
					e.Attributes[k] = v
				}
			}
			attrRows.Close()
		}

		entities = append(entities, e)
	}
	entityRows.Close()
	data.Entities = entities

	// Export triples
	tripleRows, err := g.db.Query(`SELECT id, subject, predicate, object, confidence, valid_from, valid_to, source, created_at FROM triples ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("export triples: %w", err)
	}
	defer tripleRows.Close()

	for tripleRows.Next() {
		var t Triple
		if err := tripleRows.Scan(&t.ID, &t.Subject, &t.Predicate, &t.Object, &t.Confidence, &t.ValidFrom, &t.ValidTo, &t.Source, &t.CreatedAt); err != nil {
			return nil, err
		}
		data.Triples = append(data.Triples, t)
	}

	if err := tripleRows.Err(); err != nil {
		return nil, err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal export: %w", err)
	}
	return jsonData, nil
}

// Import loads entities and triples from a JSON export.
func (g *Graph) Import(data []byte) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	var exp GraphExport
	if err := json.Unmarshal(data, &exp); err != nil {
		return fmt.Errorf("unmarshal import: %w", err)
	}

	tx, err := g.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Import entities
	for _, e := range exp.Entities {
		if e.ID == "" {
			e.ID = generateID("ent_" + e.Name)
		}
		if e.Type == "" {
			e.Type = DetectEntity(e.Name)
		}
		if e.Aliases == nil {
			e.Aliases = []string{}
		}

		aliasesJSON, mErr := json.Marshal(e.Aliases)
		if mErr != nil {
			return fmt.Errorf("marshal aliases: %w", mErr)
		}

		if _, err := tx.Exec(
			`INSERT OR REPLACE INTO entities (id, name, type, aliases, created_at)
			 VALUES (?, ?, ?, ?, ?)`,
			e.ID, e.Name, e.Type, string(aliasesJSON), e.CreatedAt,
		); err != nil {
			return fmt.Errorf("import entity %s: %w", e.Name, err)
		}

		// Import attributes
		for k, v := range e.Attributes {
			if _, err := tx.Exec(
				`INSERT OR REPLACE INTO attributes (entity_id, key, value) VALUES (?, ?, ?)`,
				e.ID, k, v,
			); err != nil {
				return fmt.Errorf("import attribute %s.%s: %w", e.Name, k, err)
			}
		}
	}

	// Import triples
	now := time.Now().Unix()
	for _, t := range exp.Triples {
		if t.ID == "" {
			t.ID = generateTripleID(t.Subject, t.Predicate, t.Object)
		}
		if t.Confidence <= 0 || t.Confidence > 1 {
			t.Confidence = 1.0
		}
		if t.CreatedAt == 0 {
			t.CreatedAt = now
		}

		if _, err := tx.Exec(
			`INSERT OR REPLACE INTO triples (id, subject, predicate, object, confidence, valid_from, valid_to, source, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			t.ID, t.Subject, t.Predicate, t.Object, t.Confidence, t.ValidFrom, t.ValidTo, t.Source, t.CreatedAt,
		); err != nil {
			return fmt.Errorf("import triple %s: %w", t.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit import: %w", err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────
// Entity Auto-Detection
// ─────────────────────────────────────────────────────────────

// personNamePattern matches typical person names: "John Smith", "Nguyen Van A"
var personNamePattern = regexp.MustCompile(`^[A-Z][a-z]+(\s+[A-Z][a-z]+)+$`)

// knownOrgs is a set of common organization name patterns.
var knownOrgs = []string{" Inc", " Corp", " LLC", " Ltd", " GmbH", " SARL", " Co.", " LLP", ", Inc", ", Corp", ", LLC", ", Ltd"}

// knownCities contains common city/country keywords for location detection.
var knownCities = map[string]bool{
	"new york": true, "london": true, "tokyo": true, "paris": true,
	"berlin": true, "beijing": true, "shanghai": true, "moscow": true,
	"sydney": true, "toronto": true, "singapore": true, "dubai": true,
	"mumbai": true, "seoul": true, "bangkok": true, "hanoi": true,
	"ho chi minh": true, "amsterdam": true, "zurich": true, "san francisco": true,
	"los angeles": true, "chicago": true, "boston": true, "seattle": true,
}

// knownProjectKeywords indicate a project name.
var knownProjectKeywords = []string{".git", "github.com", "gitlab.com", "repo", "project"}

// DetectEntity attempts to classify an entity name into a type.
func DetectEntity(name string) string {
	if name == "" {
		return "concept"
	}

	nameLower := strings.ToLower(name)

	// Task: contains #, TODO, FIXME (highest priority -- very specific)
	if strings.Contains(name, "#") || strings.Contains(strings.ToUpper(name), "TODO") || strings.Contains(strings.ToUpper(name), "FIXME") {
		return "task"
	}

	// Organization: Inc, Corp, LLC, Ltd, etc. (check before person name)
	for _, org := range knownOrgs {
		if strings.Contains(name, org) {
			return "organization"
		}
	}

	// Person: contains @ or looks like a proper name (Title Case Words)
	if strings.Contains(name, "@") {
		return "person"
	}
	if personNamePattern.MatchString(name) {
		wordCount := len(strings.Fields(name))
		if wordCount >= 2 && wordCount <= 5 {
			return "person"
		}
	}

	// Task: contains #, TODO, FIXME
	if strings.Contains(name, "#") || strings.Contains(strings.ToUpper(name), "TODO") || strings.Contains(strings.ToUpper(name), "FIXME") {
		return "task"
	}

	// Project: contains /, .git, or known project keywords
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return "project"
	}
	for _, kw := range knownProjectKeywords {
		if strings.Contains(nameLower, kw) {
			return "project"
		}
	}

	// Event: contains conference, meetup, summit, etc. (check before location)
	eventKeywords := []string{"conference", "meetup", "summit", "hackathon", "workshop", "seminar", "symposium", "festival"}
	for _, kw := range eventKeywords {
		if strings.Contains(nameLower, kw) {
			return "event"
		}
	}

	// Location: known city names
	if knownCities[nameLower] {
		return "location"
	}

	// Default: concept
	return "concept"
}

// ─────────────────────────────────────────────────────────────
// ID Generation Helpers
// ─────────────────────────────────────────────────────────────

// generateID creates a deterministic short ID from input strings.
func generateID(parts ...string) string {
	raw := strings.Join(parts, "|")
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:8]) // first 16 hex chars
}

// generateTripleID creates a deterministic ID for a triple.
func generateTripleID(subject, predicate, object string) string {
	return generateID(subject, predicate, object)
}
