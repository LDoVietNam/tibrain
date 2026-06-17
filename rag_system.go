// TiBrain RAG System Integration
// Provides centralized RAG capabilities for the Ti ecosystem
package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"
)

// ─────────────────────────────────────────────────────────────
// RAG System Components
// ─────────────────────────────────────────────────────────────

// RAGDocument represents a document in the RAG system
type RAGDocument struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Path      string    `json:"path"`
	Category  string    `json:"category"`
	Tags      string    `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Status    string    `json:"status"`
	VectorID  string    `json:"vector_id"`
	Metadata  string    `json:"metadata"`
}

// RAGQuery represents a RAG query request
type RAGQuery struct {
	ID         string              `json:"id"`
	Query      string              `json:"query"`
	Context    string              `json:"context"`
	Documents  []string            `json:"documents"`
	Response   string              `json:"response"`
	Confidence float64             `json:"confidence"`
	Timestamp  time.Time           `json:"timestamp"`
	UserID     string              `json:"user_id"`
	SessionID  string              `json:"session_id"`
	Metadata   map[string]string   `json:"metadata"`
	Scopes     []string            `json:"scopes,omitempty"`
	ContextSrc *AgentContextSource `json:"context_source,omitempty"`
	Runtime    *RAGRuntimeTrace    `json:"runtime,omitempty"`
}

// RAGKnowledgeBase represents the knowledge base configuration
type RAGKnowledgeBase struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Path        string    `json:"path"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DocCount    int       `json:"doc_count"`
	VectorCount int       `json:"vector_count"`
}

// ─────────────────────────────────────────────────────────────
// RAG System Manager
// ─────────────────────────────────────────────────────────────

type RAGConfig struct {
	EmbeddingURL    string
	EmbeddingModel  string
	EmbeddingDim    int
	TopK            int
	Threshold       float64
	UseVectorSearch bool
	EnableGenerativeAnswers bool
	EmbeddingTimeout        time.Duration
	LLMTimeout              time.Duration
}

var defaultRAGConfig = RAGConfig{
	EmbeddingURL:             "http://localhost:1807/v1",
	EmbeddingModel:           "text-embedding-3-small",
	EmbeddingDim:             1536,
	TopK:                     5,
	Threshold:                0.7,
	UseVectorSearch:          false,
	EnableGenerativeAnswers:  false,
	EmbeddingTimeout:         2500 * time.Millisecond,
	LLMTimeout:               4 * time.Second,
}

type RAGSystemManager struct {
	hub         *Hub
	vectorStore *VectorStore
	embedder    *EmbeddingGenerator
	cache       *RAGCache
	reranker    *HybridReranker
	assembler   *ContextAssembler
	llm         *LLMClient
	config      RAGConfig
	tagEngine   *TagRuleEngine
	deferKnowledgeBaseStats bool
	pendingKnowledgeBaseStats map[string]bool
}

func NewRAGSystemManager(hub *Hub) *RAGSystemManager {
	config := loadRAGConfig()

	var llmClient *LLMClient
	if config.EnableGenerativeAnswers {
		llmClient = NewLLMClientWithTimeout(config.EmbeddingURL, "gpt-4o-mini", config.LLMTimeout)
	}
	if llmClient != nil && hub != nil && hub.db != nil {
		compressor := NewLocalRTKCompressor(hub.db)
		llmClient.SetCompressor(compressor)
	}

	var embedder *EmbeddingGenerator
	if config.UseVectorSearch {
		embedder = NewEmbeddingGeneratorWithTimeout(config.EmbeddingURL, config.EmbeddingModel, config.EmbeddingDim, config.EmbeddingTimeout)
	}

	return &RAGSystemManager{
		hub:         hub,
		vectorStore: NewVectorStore(hub.db, config.EmbeddingDim),
		embedder:    embedder,
		cache:       NewRAGCache(5 * time.Minute),
		reranker:    NewHybridReranker(),
		assembler:   NewContextAssembler(),
		llm:         llmClient,
		config:      config,
		tagEngine:   NewTagRuleEngine(),
		pendingKnowledgeBaseStats: make(map[string]bool),
	}
}

func loadRAGConfig() RAGConfig {
	config := defaultRAGConfig
	config.UseVectorSearch = getEnvBool("TIBRAIN_ENABLE_VECTOR_SEARCH", config.UseVectorSearch)
	config.EnableGenerativeAnswers = getEnvBool("TIBRAIN_ENABLE_LLM_ANSWERS", config.EnableGenerativeAnswers)
	config.EmbeddingTimeout = time.Duration(getEnvInt("TIBRAIN_EMBEDDING_TIMEOUT_MS", int(config.EmbeddingTimeout/time.Millisecond))) * time.Millisecond
	config.LLMTimeout = time.Duration(getEnvInt("TIBRAIN_LLM_TIMEOUT_MS", int(config.LLMTimeout/time.Millisecond))) * time.Millisecond
	return config
}

func getEnvBool(name string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvInt(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

// ─────────────────────────────────────────────────────────────
// Database Schema Initialization
// ─────────────────────────────────────────────────────────────

func initRAGSchema(db *sql.DB) error {
	schema := `
	-- RAG Documents Storage
	CREATE TABLE IF NOT EXISTS rag_documents (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		path TEXT NOT NULL,
		category TEXT NOT NULL,
		tags TEXT,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		vector_id TEXT,
		metadata TEXT,
		content_hash TEXT,
		file_size INTEGER DEFAULT 0,
		last_indexed INTEGER,
		indexing_status TEXT DEFAULT 'pending'
	);
	CREATE INDEX IF NOT EXISTS idx_rag_documents_category ON rag_documents(category);
	CREATE INDEX IF NOT EXISTS idx_rag_documents_status ON rag_documents(status);
	CREATE INDEX IF NOT EXISTS idx_rag_documents_tags ON rag_documents(tags);
	CREATE INDEX IF NOT EXISTS idx_rag_documents_updated ON rag_documents(updated_at);
	CREATE INDEX IF NOT EXISTS idx_rag_documents_content_hash ON rag_documents(content_hash);

	CREATE VIRTUAL TABLE IF NOT EXISTS rag_documents_fts USING fts5(
		id,
		title,
		content,
		path,
		category,
		tags,
		metadata,
		content='rag_documents',
		content_rowid='rowid'
	);

	CREATE TRIGGER IF NOT EXISTS rag_documents_ai AFTER INSERT ON rag_documents BEGIN
		INSERT INTO rag_documents_fts(rowid, id, title, content, path, category, tags, metadata)
		VALUES (new.rowid, new.id, new.title, new.content, new.path, new.category, COALESCE(new.tags, ''), COALESCE(new.metadata, ''));
	END;

	CREATE TRIGGER IF NOT EXISTS rag_documents_ad AFTER DELETE ON rag_documents BEGIN
		INSERT INTO rag_documents_fts(rag_documents_fts, rowid, id, title, content, path, category, tags, metadata)
		VALUES ('delete', old.rowid, old.id, old.title, old.content, old.path, old.category, COALESCE(old.tags, ''), COALESCE(old.metadata, ''));
	END;

	CREATE TRIGGER IF NOT EXISTS rag_documents_au AFTER UPDATE ON rag_documents BEGIN
		INSERT INTO rag_documents_fts(rag_documents_fts, rowid, id, title, content, path, category, tags, metadata)
		VALUES ('delete', old.rowid, old.id, old.title, old.content, old.path, old.category, COALESCE(old.tags, ''), COALESCE(old.metadata, ''));
		INSERT INTO rag_documents_fts(rowid, id, title, content, path, category, tags, metadata)
		VALUES (new.rowid, new.id, new.title, new.content, new.path, new.category, COALESCE(new.tags, ''), COALESCE(new.metadata, ''));
	END;

	INSERT INTO rag_documents_fts(rag_documents_fts) VALUES('rebuild');

	-- RAG Knowledge Bases
	CREATE TABLE IF NOT EXISTS rag_knowledge_bases (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		path TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		doc_count INTEGER DEFAULT 0,
		vector_count INTEGER DEFAULT 0,
		last_sync INTEGER,
		sync_status TEXT DEFAULT 'pending'
	);
	CREATE INDEX IF NOT EXISTS idx_rag_kb_status ON rag_knowledge_bases(status);
	CREATE INDEX IF NOT EXISTS idx_rag_kb_updated ON rag_knowledge_bases(updated_at);

	-- RAG Query History
	CREATE TABLE IF NOT EXISTS rag_query_history (
		id TEXT PRIMARY KEY,
		query TEXT NOT NULL,
		context TEXT,
		documents TEXT,
		response TEXT,
		confidence REAL DEFAULT 0.0,
		timestamp INTEGER NOT NULL,
		user_id TEXT,
		session_id TEXT,
		metadata TEXT,
		query_hash TEXT,
		response_time INTEGER DEFAULT 0,
		model_used TEXT,
		tokens_used INTEGER DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_rag_query_timestamp ON rag_query_history(timestamp);
	CREATE INDEX IF NOT EXISTS idx_rag_query_user ON rag_query_history(user_id);
	CREATE INDEX IF NOT EXISTS idx_rag_query_session ON rag_query_history(session_id);
	CREATE INDEX IF NOT EXISTS idx_rag_query_hash ON rag_query_history(query_hash);

	-- RAG Vector Index (for future vector database integration)
	CREATE TABLE IF NOT EXISTS rag_vector_index (
		id TEXT PRIMARY KEY,
		document_id TEXT NOT NULL,
		chunk_id TEXT NOT NULL,
		vector_data TEXT, -- JSON encoded vector
		chunk_text TEXT NOT NULL,
		chunk_order INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		FOREIGN KEY (document_id) REFERENCES rag_documents(id)
	);
	CREATE INDEX IF NOT EXISTS idx_rag_vector_document ON rag_vector_index(document_id);
	CREATE INDEX IF NOT EXISTS idx_rag_vector_chunk ON rag_vector_index(chunk_id);

	-- RAG Analytics and Metrics
	CREATE TABLE IF NOT EXISTS rag_analytics (
		id TEXT PRIMARY KEY,
		metric_type TEXT NOT NULL,
		metric_name TEXT NOT NULL,
		metric_value REAL NOT NULL,
		metadata TEXT,
		timestamp INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_rag_analytics_type ON rag_analytics(metric_type);
	CREATE INDEX IF NOT EXISTS idx_rag_analytics_timestamp ON rag_analytics(timestamp);

	-- RAG Feedback Loop
	CREATE TABLE IF NOT EXISTS rag_feedback (
		id TEXT PRIMARY KEY,
		query_id TEXT NOT NULL,
		rating INTEGER NOT NULL,
		feedback TEXT,
		improvement_suggestions TEXT,
		timestamp INTEGER NOT NULL,
		user_id TEXT,
		FOREIGN KEY (query_id) REFERENCES rag_query_history(id)
	);
	CREATE INDEX IF NOT EXISTS idx_rag_feedback_query ON rag_feedback(query_id);
	CREATE INDEX IF NOT EXISTS idx_rag_feedback_rating ON rag_feedback(rating);

	-- Learning Plane: runtime retrieval traces
	CREATE TABLE IF NOT EXISTS rag_runtime_traces (
		id TEXT PRIMARY KEY,
		query_id TEXT NOT NULL,
		query_text TEXT NOT NULL,
		mode TEXT NOT NULL,
		cache_hit INTEGER NOT NULL DEFAULT 0,
		verified INTEGER NOT NULL DEFAULT 0,
		used_slow_path INTEGER NOT NULL DEFAULT 0,
		confidence_score REAL DEFAULT 0.0,
		groundedness_score REAL DEFAULT 0.0,
		coverage_score REAL DEFAULT 0.0,
		quality_score REAL DEFAULT 0.0,
		fast_candidates INTEGER DEFAULT 0,
		slow_candidates INTEGER DEFAULT 0,
		selected_documents INTEGER DEFAULT 0,
		distilled_candidate_id TEXT,
		metadata TEXT,
		created_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_rag_runtime_traces_query ON rag_runtime_traces(query_id);
	CREATE INDEX IF NOT EXISTS idx_rag_runtime_traces_created ON rag_runtime_traces(created_at);
	CREATE INDEX IF NOT EXISTS idx_rag_runtime_traces_verified ON rag_runtime_traces(verified);

	-- Learning Plane: distilled candidates waiting for explicit promotion
	CREATE TABLE IF NOT EXISTS brain_pattern_candidates (
		id TEXT PRIMARY KEY,
		query_id TEXT NOT NULL,
		title TEXT NOT NULL,
		summary TEXT NOT NULL,
		pattern_body TEXT NOT NULL,
		evidence TEXT,
		score REAL NOT NULL DEFAULT 0.0,
		status TEXT NOT NULL DEFAULT 'candidate',
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		promoted_at INTEGER
	);
	CREATE INDEX IF NOT EXISTS idx_brain_pattern_candidates_status ON brain_pattern_candidates(status);
	CREATE INDEX IF NOT EXISTS idx_brain_pattern_candidates_score ON brain_pattern_candidates(score);
	CREATE INDEX IF NOT EXISTS idx_brain_pattern_candidates_updated ON brain_pattern_candidates(updated_at);

	-- CLI Context Storage
	CREATE TABLE IF NOT EXISTS cli_context (
		id TEXT PRIMARY KEY,
		cli_id TEXT NOT NULL,
		context_type TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		metadata TEXT,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		expires_at INTEGER,
		access_count INTEGER DEFAULT 0,
		last_accessed INTEGER
	);
	CREATE INDEX IF NOT EXISTS idx_cli_context_cli ON cli_context(cli_id);
	CREATE INDEX IF NOT EXISTS idx_cli_context_type ON cli_context(context_type);
	CREATE INDEX IF NOT EXISTS idx_cli_context_key ON cli_context(key);
	CREATE INDEX IF NOT EXISTS idx_cli_context_expires ON cli_context(expires_at);

	-- CLI Query Cache
	CREATE TABLE IF NOT EXISTS cli_query_cache (
		id TEXT PRIMARY KEY,
		query_hash TEXT NOT NULL,
		query TEXT NOT NULL,
		response TEXT NOT NULL,
		confidence REAL,
		source_type TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		expires_at INTEGER,
		hit_count INTEGER DEFAULT 0,
		last_hit INTEGER
	);
	CREATE INDEX IF NOT EXISTS idx_cli_cache_hash ON cli_query_cache(query_hash);
	CREATE INDEX IF NOT EXISTS idx_cli_cache_expires ON cli_query_cache(expires_at);
	CREATE INDEX IF NOT EXISTS idx_cli_cache_source ON cli_query_cache(source_type);

	-- CLI Help Index
	CREATE TABLE IF NOT EXISTS cli_help_index (
		id TEXT PRIMARY KEY,
		command TEXT NOT NULL,
		topic TEXT NOT NULL,
		content TEXT NOT NULL,
		keywords TEXT,
		category TEXT,
		priority INTEGER DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_cli_help_command ON cli_help_index(command);
	CREATE INDEX IF NOT EXISTS idx_cli_help_topic ON cli_help_index(topic);
	CREATE INDEX IF NOT EXISTS idx_cli_help_category ON cli_help_index(category);
	CREATE INDEX IF NOT EXISTS idx_cli_help_keywords ON cli_help_index(keywords);

	-- CLI Preferences
	CREATE TABLE IF NOT EXISTS cli_preferences (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_cli_prefs_user ON cli_preferences(user_id);
	CREATE INDEX IF NOT EXISTS idx_cli_prefs_key ON cli_preferences(key);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("create RAG schema: %w", err)
	}
	if err := ensureRAGSchemaColumns(db); err != nil {
		return err
	}

	return nil
}

func ensureRAGSchemaColumns(db *sql.DB) error {
	columns := map[string]string{
		"last_sync":   "ALTER TABLE rag_knowledge_bases ADD COLUMN last_sync INTEGER",
		"sync_status": "ALTER TABLE rag_knowledge_bases ADD COLUMN sync_status TEXT DEFAULT 'pending'",
	}
	for column, statement := range columns {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('rag_knowledge_bases') WHERE name=?", column).Scan(&count); err != nil {
			return fmt.Errorf("check rag_knowledge_bases.%s column: %w", column, err)
		}
		if count == 0 {
			if _, err := db.Exec(statement); err != nil {
				return fmt.Errorf("add rag_knowledge_bases.%s column: %w", column, err)
			}
		}
	}
	return nil
}

// ─────────────────────────────────────────────────────────────
// Document Management
// ─────────────────────────────────────────────────────────────

// AddDocument adds a new document to the RAG system
func (r *RAGSystemManager) AddDocument(doc RAGDocument) error {
	timestamp := time.Now().Unix()
	if doc.ID == "" {
		doc.ID = generateContentHash(doc.Path)
	}
	if doc.Status == "" {
		doc.Status = "active"
	}
	if doc.Category == "" {
		doc.Category = "general"
	}

	_, err := r.hub.db.Exec(`
		INSERT OR REPLACE INTO rag_documents 
		(id, title, content, path, category, tags, created_at, updated_at, status, vector_id, metadata, content_hash, file_size, last_indexed, indexing_status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		doc.ID, doc.Title, doc.Content, doc.Path, doc.Category, doc.Tags,
		timestamp, timestamp, doc.Status, doc.VectorID, doc.Metadata,
		generateContentHash(doc.Content), len(doc.Content), timestamp, "pending")

	if err != nil {
		return fmt.Errorf("add document: %w", err)
	}

	if r.deferKnowledgeBaseStats {
		r.pendingKnowledgeBaseStats[doc.Category] = true
	} else {
		r.updateKnowledgeBaseDocCount(doc.Category)
	}

	logger.Info("RAG document added: %s (%s)", doc.ID, doc.Title)
	return nil
}

// GetDocument retrieves a document by ID
func (r *RAGSystemManager) GetDocument(id string) (*RAGDocument, error) {
	var doc RAGDocument
	var createdAt, updatedAt int64

	err := r.hub.db.QueryRow(`
		SELECT id, title, content, path, category, tags, created_at, updated_at, 
		       status, vector_id, metadata
		FROM rag_documents WHERE id = ?
	`, id).Scan(
		&doc.ID, &doc.Title, &doc.Content, &doc.Path, &doc.Category, &doc.Tags,
		&createdAt, &updatedAt, &doc.Status, &doc.VectorID, &doc.Metadata)

	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}

	doc.CreatedAt = time.Unix(createdAt, 0)
	doc.UpdatedAt = time.Unix(updatedAt, 0)

	return &doc, nil
}

// SearchDocuments performs hybrid semantic + keyword search across documents
func (r *RAGSystemManager) SearchDocuments(query string, limit int, category string, sqlFilter string) ([]RAGDocument, error) {
	if limit <= 0 {
		limit = r.config.TopK
	}

	// Try vector search first only when explicitly enabled and vectors exist.
	if r.config.UseVectorSearch && r.vectorStore != nil && r.embedder != nil && r.hasIndexedVectors() {
		vecDocs, err := r.searchVector(query, limit, category, sqlFilter)
		if err == nil && len(vecDocs) > 0 {
			logger.Debug("Vector search returned %d results for query: %s", len(vecDocs), truncateQuery(query, 30))
			return vecDocs, nil
		}
		if err != nil {
			logger.Warn("Vector search failed (%v), falling back to keyword search", err)
		}
	}

	// Fallback to keyword search
	return r.searchKeyword(query, limit, category, sqlFilter)
}

// searchVector performs cosine similarity search using embeddings
func (r *RAGSystemManager) searchVector(query string, limit int, category string, sqlFilter string) ([]RAGDocument, error) {
	queryVector, err := r.embedder.GenerateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("generate query embedding: %w", err)
	}

	results, err := r.vectorStore.SearchSimilar(queryVector, limit, r.config.Threshold, sqlFilter)
	if err != nil {
		return nil, fmt.Errorf("vector similarity search: %w", err)
	}

	var documents []RAGDocument
	seen := make(map[string]bool)
	for _, result := range results {
		if seen[result.DocumentID] {
			continue
		}
		seen[result.DocumentID] = true

		doc, err := r.GetDocument(result.DocumentID)
		if err != nil {
			continue
		}
		if category != "" && doc.Category != category {
			continue
		}
		documents = append(documents, *doc)
	}

	return documents, nil
}

// searchKeyword performs traditional LIKE-based keyword search
func (r *RAGSystemManager) searchKeyword(query string, limit int, category string, sqlFilter string) ([]RAGDocument, error) {
	var args []interface{}
	sqlQuery := `
		SELECT id, title, content, path, category, tags, created_at, updated_at,
		       status, vector_id, metadata
		FROM rag_documents
		WHERE status = 'active' ` + sqlFilter + `
	`

	if category != "" {
		sqlQuery += " AND category = ?"
		args = append(args, category)
	}

	tokens := queryTokens(query)
	if len(tokens) == 0 {
		tokens = []string{query}
	}

	clauses := make([]string, 0, len(tokens))
	for _, token := range tokens {
		clauses = append(clauses, "(content LIKE ? OR title LIKE ? OR tags LIKE ?)")
		like := "%" + token + "%"
		args = append(args, like, like, like)
	}
	sqlQuery += " AND (" + strings.Join(clauses, " OR ") + ")"

	sqlQuery += " ORDER BY updated_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.hub.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("keyword search: %w", err)
	}
	defer rows.Close()

	var documents []RAGDocument
	for rows.Next() {
		var doc RAGDocument
		var createdAt, updatedAt int64

		err := rows.Scan(
			&doc.ID, &doc.Title, &doc.Content, &doc.Path, &doc.Category, &doc.Tags,
			&createdAt, &updatedAt, &doc.Status, &doc.VectorID, &doc.Metadata)
		if err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}

		doc.CreatedAt = time.Unix(createdAt, 0)
		doc.UpdatedAt = time.Unix(updatedAt, 0)
		documents = append(documents, doc)
	}

	return documents, nil
}

// Query Processing
// ─────────────────────────────────────────────────────────────

// ProcessQuery processes a RAG query through the adaptive retrieval runtime.
func (r *RAGSystemManager) ProcessQuery(query RAGQuery) (*RAGQuery, error) {
	return r.ProcessQueryWithContext(context.Background(), query)
}

// GetQueryHistory retrieves query history for a user
func (r *RAGSystemManager) GetQueryHistory(userID string, limit int) ([]RAGQuery, error) {
	rows, err := r.hub.db.Query(`
		SELECT id, query, context, documents, response, confidence, timestamp, 
		       user_id, session_id, metadata
		FROM rag_query_history 
		WHERE user_id = ? 
		ORDER BY timestamp DESC 
		LIMIT ?
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("get query history: %w", err)
	}
	defer rows.Close()

	var queries []RAGQuery
	for rows.Next() {
		var query RAGQuery
		var timestamp int64
		var metadataJSON string

		err := rows.Scan(
			&query.ID, &query.Query, &query.Context, &query.Documents,
			&query.Response, &query.Confidence, &timestamp,
			&query.UserID, &query.SessionID, &metadataJSON)
		if err != nil {
			return nil, fmt.Errorf("scan query: %w", err)
		}

		query.Timestamp = time.Unix(timestamp, 0)
		if metadataJSON != "" {
			json.Unmarshal([]byte(metadataJSON), &query.Metadata)
		}
		queries = append(queries, query)
	}

	return queries, nil
}

// ─────────────────────────────────────────────────────────────
// Knowledge Base Management
// ─────────────────────────────────────────────────────────────

// CreateKnowledgeBase creates a new knowledge base
func (r *RAGSystemManager) CreateKnowledgeBase(kb RAGKnowledgeBase) error {
	timestamp := time.Now().Unix()

	_, err := r.hub.db.Exec(`
		INSERT OR REPLACE INTO rag_knowledge_bases 
		(id, name, description, path, status, created_at, updated_at, doc_count, vector_count, last_sync, sync_status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		kb.ID, kb.Name, kb.Description, kb.Path, kb.Status,
		timestamp, timestamp, kb.DocCount, kb.VectorCount, timestamp, "pending")

	if err != nil {
		return fmt.Errorf("create knowledge base: %w", err)
	}

	logger.Info("RAG knowledge base created: %s (%s)", kb.ID, kb.Name)
	return nil
}

// GetKnowledgeBases retrieves all knowledge bases
func (r *RAGSystemManager) GetKnowledgeBases() ([]RAGKnowledgeBase, error) {
	rows, err := r.hub.db.Query(`
		SELECT id, name, description, path, status, created_at, updated_at, 
		       doc_count, vector_count, last_sync, sync_status
		FROM rag_knowledge_bases 
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("get knowledge bases: %w", err)
	}
	defer rows.Close()

	var kbs []RAGKnowledgeBase
	for rows.Next() {
		var kb RAGKnowledgeBase
		var createdAt, updatedAt, lastSync int64

		err := rows.Scan(
			&kb.ID, &kb.Name, &kb.Description, &kb.Path, &kb.Status,
			&createdAt, &updatedAt, &kb.DocCount, &kb.VectorCount, &lastSync)
		if err != nil {
			return nil, fmt.Errorf("scan knowledge base: %w", err)
		}

		kb.CreatedAt = time.Unix(createdAt, 0)
		kb.UpdatedAt = time.Unix(updatedAt, 0)
		kbs = append(kbs, kb)
	}

	return kbs, nil
}

// ─────────────────────────────────────────────────────────────
// Analytics and Metrics
// ─────────────────────────────────────────────────────────────

// GetAnalytics retrieves RAG system analytics
func (r *RAGSystemManager) GetAnalytics(metricType string, timeRange string) (map[string]interface{}, error) {
	var timeFilter string
	switch timeRange {
	case "1h":
		timeFilter = fmt.Sprintf("timestamp > %d", time.Now().Add(-time.Hour).Unix())
	case "24h":
		timeFilter = fmt.Sprintf("timestamp > %d", time.Now().Add(-24*time.Hour).Unix())
	case "7d":
		timeFilter = fmt.Sprintf("timestamp > %d", time.Now().Add(-7*24*time.Hour).Unix())
	default:
		timeFilter = "1=1"
	}

	analytics := make(map[string]interface{})

	// Query metrics
	if metricType == "all" || metricType == "queries" {
		var totalQueries, avgConfidence float64
		r.hub.db.QueryRow(`
			SELECT COUNT(*), AVG(confidence) 
			FROM rag_query_history 
			WHERE `+timeFilter).Scan(&totalQueries, &avgConfidence)

		analytics["total_queries"] = totalQueries
		analytics["avg_confidence"] = avgConfidence
	}

	// Document metrics
	if metricType == "all" || metricType == "documents" {
		var totalDocs, indexedDocs float64
		r.hub.db.QueryRow(`
			SELECT COUNT(*), COUNT(CASE WHEN indexing_status = 'completed' THEN 1 END) 
			FROM rag_documents`).Scan(&totalDocs, &indexedDocs)

		analytics["total_documents"] = totalDocs
		analytics["indexed_documents"] = indexedDocs
		analytics["indexing_rate"] = indexedDocs / totalDocs * 100
	}

	// Knowledge base metrics
	if metricType == "all" || metricType == "knowledge_bases" {
		var totalKBs, activeKBs float64
		r.hub.db.QueryRow(`
			SELECT COUNT(*), COUNT(CASE WHEN status = 'active' THEN 1 END) 
			FROM rag_knowledge_bases`).Scan(&totalKBs, &activeKBs)

		analytics["total_knowledge_bases"] = totalKBs
		analytics["active_knowledge_bases"] = activeKBs
	}

	return analytics, nil
}

// ─────────────────────────────────────────────────────────────
// Utility Functions
// ─────────────────────────────────────────────────────────────

var generateIDCounter uint64

func generateID() string {
	counter := atomic.AddUint64(&generateIDCounter, 1)
	// Combine nanosecond timestamp + atomic counter + a small random suffix
	// to guarantee uniqueness even when called multiple times per nanosecond.
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fall back to time-derived bytes if /dev/urandom is unavailable.
		b[0] = byte(time.Now().UnixNano())
		b[1] = byte(time.Now().UnixNano() >> 8)
		b[2] = byte(time.Now().UnixNano() >> 16)
		b[3] = byte(time.Now().UnixNano() >> 24)
	}
	return fmt.Sprintf("rag_%d_%d_%s", time.Now().UnixNano(), counter, fmt.Sprintf("%x", b))
}

func generateContentHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", sum)
}

// generateQueryHash moved to cli_context.go to avoid duplication

func marshalMetadata(metadata map[string]string) string {
	if metadata == nil {
		return ""
	}
	data, _ := json.Marshal(metadata)
	return string(data)
}

func generateSimpleResponse(query string, documents []RAGDocument) string {
	if len(documents) == 0 {
		return fmt.Sprintf("I couldn't find relevant information for your query: %s", query)
	}

	response := fmt.Sprintf("Based on the available documentation, here's what I found about '%s':\n\n", query)

	for i, doc := range documents {
		if i >= 3 { // Limit to top 3 documents
			break
		}
		response += fmt.Sprintf("From '%s':\n%s\n\n", doc.Title,
			truncateContent(doc.Content, 200))
	}

	return response
}

func calculateConfidence(query string, documents []RAGDocument) float64 {
	if len(documents) == 0 {
		return 0.0
	}

	// Simple confidence calculation based on document count and content match
	baseConfidence := float64(len(documents)) / 5.0 // Normalize to 0-1 range
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}

	// Boost confidence if query terms appear in titles
	titleBoost := 0.0
	for _, doc := range documents {
		if strings.Contains(strings.ToLower(doc.Title), strings.ToLower(query)) {
			titleBoost += 0.2
		}
	}

	confidence := baseConfidence + titleBoost
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}

func (r *RAGSystemManager) updateKnowledgeBaseDocCount(category string) {
	_, err := r.hub.db.Exec(`
		UPDATE rag_knowledge_bases 
		SET doc_count = (
			SELECT COUNT(*) FROM rag_documents 
			WHERE category = ? AND status = 'active'
		),
		updated_at = ?
		WHERE id = ?
	`, category, time.Now().Unix(), category)

	if err != nil {
		logger.Warn("Failed to update knowledge base doc count: %v", err)
	}
}

func (r *RAGSystemManager) flushKnowledgeBaseStats() {
	if !r.deferKnowledgeBaseStats {
		return
	}
	for category := range r.pendingKnowledgeBaseStats {
		r.updateKnowledgeBaseDocCount(category)
	}
	clear(r.pendingKnowledgeBaseStats)
}

func (r *RAGSystemManager) logAnalytics(metricType, metricName string, metricValue float64, metadata map[string]string) {
	r.hub.asyncWriter.Enqueue(`
		INSERT INTO rag_analytics 
		(id, metric_type, metric_name, metric_value, metadata, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`, generateID(), metricType, metricName, metricValue, marshalMetadata(metadata), time.Now().Unix())
}

// ─────────────────────────────────────────────────────────────
// HTTP Handlers for RAG System
// ─────────────────────────────────────────────────────────────

func (h *Hub) setupRAGHandlers() {
	ragManager := NewRAGSystemManager(h)

	// RAG Query endpoint
	http.HandleFunc("/rag/query", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var query RAGQuery
		if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		result, err := ragManager.ProcessQuery(query)
		if err != nil {
			http.Error(w, fmt.Sprintf("Process query error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// RAG Search endpoint
	http.HandleFunc("/rag/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		category := r.URL.Query().Get("category")
		limit := 5

		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
				limit = parsedLimit
			}
		}

		scopes := []string{}
		if scopeStr := r.URL.Query().Get("scopes"); scopeStr != "" {
			scopes = strings.Split(scopeStr, ",")
		}
		sqlFilter := BuildScopeQuery(scopes)

		documents, err := ragManager.SearchDocuments(query, limit, category, sqlFilter)
		if err != nil {
			http.Error(w, fmt.Sprintf("Search error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(documents)
	})

	// RAG Analytics endpoint
	http.HandleFunc("/rag/analytics", func(w http.ResponseWriter, r *http.Request) {
		metricType := r.URL.Query().Get("type")
		timeRange := r.URL.Query().Get("range")

		analytics, err := ragManager.GetAnalytics(metricType, timeRange)
		if err != nil {
			http.Error(w, fmt.Sprintf("Analytics error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(analytics)
	})

	// RAG Knowledge Bases endpoint
	http.HandleFunc("/rag/knowledge-bases", func(w http.ResponseWriter, r *http.Request) {
		kbs, err := ragManager.GetKnowledgeBases()
		if err != nil {
			http.Error(w, fmt.Sprintf("Get knowledge bases error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(kbs)
	})

	// RAG Document Upload endpoint
	http.HandleFunc("/rag/documents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var requestPayload struct {
				RAGDocument
				ContextSrc *AgentContextSource `json:"context_source,omitempty"`
			}
			if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			doc := requestPayload.RAGDocument
			if doc.ID == "" {
				doc.ID = generateID()
			}
			if doc.Status == "" {
				doc.Status = "active"
			}

			if requestPayload.ContextSrc != nil {
				tags, meta := ragManager.tagEngine.Process(*requestPayload.ContextSrc)
				doc.Tags = strings.Join(tags, ",")

				// Merge or set metadata
				if doc.Metadata == "" {
					metaBytes, _ := json.Marshal(meta)
					doc.Metadata = string(metaBytes)
				}
			}

			err := ragManager.AddDocument(doc)
			if err != nil {
				http.Error(w, fmt.Sprintf("Add document error: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"id": doc.ID})
		} else if r.Method == "GET" {
			id := r.URL.Query().Get("id")
			if id == "" {
				http.Error(w, "Document ID required", http.StatusBadRequest)
				return
			}

			doc, err := ragManager.GetDocument(id)
			if err != nil {
				http.Error(w, fmt.Sprintf("Get document error: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(doc)
		}
	})

	logger.Info("RAG system HTTP handlers registered")
}

func queryTokens(query string) []string {
	seen := make(map[string]bool)
	tokens := []string{}
	for _, raw := range strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}) {
		token := strings.TrimSpace(raw)
		if len(token) < 3 || seen[token] {
			continue
		}
		seen[token] = true
		tokens = append(tokens, token)
		if len(tokens) >= 8 {
			break
		}
	}
	return tokens
}
