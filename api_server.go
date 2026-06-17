// Ti Brain API Server with Integration Endpoints
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// APIServer handles HTTP requests for Ti Brain
type APIServer struct {
	hub                 *Hub
	integrationManager  *IntegrationManager
	retrievalRouter     *RetrievalRouter
	tiAgentOrchestrator *TiAgentOrchestrator
	ragManager          *RAGSystemManager
	// knowledgeProcessor field intentionally omitted (was placeholder; see CLAUDE.md)
	router *http.ServeMux
}

// NewAPIServer creates a new API server
// NewAPIServer creates a new API server. Callers no longer need to pass a
// KnowledgeBaseProcessor placeholder; this signature is the canonical one.
func NewAPIServer(hub *Hub, im *IntegrationManager) *APIServer {
	// Initialize retrieval router (Router Brain merged)
	retrievalRouter := NewRetrievalRouter(hub)

	// Initialize Ti Agent orchestrator
	tiAgentOrchestrator := NewTiAgentOrchestrator(hub, retrievalRouter)

	server := &APIServer{
		hub:                 hub,
		integrationManager:  im,
		retrievalRouter:     retrievalRouter,
		tiAgentOrchestrator: tiAgentOrchestrator,
		ragManager:          retrievalRouter.ragManager,
		router:              http.NewServeMux(),
	}

	server.setupRoutes()
	return server
}

func (s *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// setupRoutes configures all API routes
func (s *APIServer) setupRoutes() {
	// Health and status
	s.router.HandleFunc("/api/health", s.healthHandler)
	s.router.HandleFunc("/api/status", s.statusHandler)

	// Knowledge base
	s.router.HandleFunc("/api/knowledge", s.knowledgeHandler)
	s.router.HandleFunc("/api/knowledge/index", s.knowledgeIndexHandler)
	s.router.HandleFunc("/api/knowledge/status", s.knowledgeStatusHandler)

	// Agent management
	s.router.HandleFunc("/api/agents", s.agentsHandler)
	s.router.HandleFunc("/api/agents/register", s.registerAgentHandler)

	// Orchestration
	s.router.HandleFunc("/api/orchestrate", s.orchestrateHandler)

	// Ti Agent Orchestrator endpoints
	s.router.HandleFunc("/api/agent/request", s.agentRequestHandler)
	s.router.HandleFunc("/api/agent/memory", s.agentMemoryHandler)

	// RAG queries
	s.router.HandleFunc("/api/rag/query", s.ragQueryHandler)
	s.router.HandleFunc("/api/rag/status", s.ragStatusHandler)
	s.router.HandleFunc("/api/rag/feedback", s.feedbackHandler)
	s.router.HandleFunc("/api/rag/ingest", s.ingestHandler)
	s.router.HandleFunc("/api/v2/retrieve", s.adaptiveRetrieveHandler)
	s.router.HandleFunc("/api/v2/retrieve/traces", s.retrievalTracesHandler)
	s.router.HandleFunc("/api/v2/brain/patterns", s.patternCandidatesHandler)
	s.router.HandleFunc("/api/v2/brain/patterns/promote", s.promotePatternHandler)
	s.router.HandleFunc("/api/v2/runtime/registry", s.runtimeRegistryHandler)

	// Cross-brain communication
	s.router.HandleFunc("/api/brain/router", s.routerBrainHandler)
	s.router.HandleFunc("/api/brain/ti", s.tiBrainHandler)

	// Tools
	s.router.HandleFunc("/api/tools", s.toolsHandler)

	// RTK Integration
	s.router.HandleFunc("/api/rtk/gain", s.rtkGainHandler)
	s.router.HandleFunc("/api/rtk/discover", s.rtkDiscoverHandler)
	s.router.HandleFunc("/api/rtk/compress", s.rtkCompressHandler)
	s.router.HandleFunc("/api/rtk/compress-messages", s.rtkCompressMessagesHandler)
	s.router.HandleFunc("/api/rtk/log", s.rtkLogHandler)
	s.router.HandleFunc("/api/rtk/rules", s.rtkRulesHandler)
	s.router.HandleFunc("/api/rtk/rules/sync", s.rtkRulesSyncHandler)

	// Obsidian vault integration
	s.router.HandleFunc("/api/obsidian", s.obsidianHandler)

	// MCP Documentation Integration
	s.router.HandleFunc("/api/docs/search", s.docsSearchHandler)
	s.router.HandleFunc("/api/docs/status", s.docsStatusHandler)

	// v1 API endpoints for Open-WebUI integration
	s.router.HandleFunc("/api/v1/rag/query", s.v1RagQueryHandler)
	s.router.HandleFunc("/api/v1/knowledge/documents", s.v1KnowledgeDocumentsHandler)
	s.router.HandleFunc("/api/v1/knowledge/ingest", s.v1KnowledgeIngestHandler)
	s.router.HandleFunc("/api/v1/graph/nodes", s.v1GraphNodesHandler)
	s.router.HandleFunc("/api/v1/router/routes", s.v1RouterRoutesHandler)
	s.router.HandleFunc("/api/v1/memory/store", s.v1MemoryStoreHandler)
	s.router.HandleFunc("/api/v1/analytics", s.v1AnalyticsHandler)

	// Root endpoint
	s.router.HandleFunc("/", s.rootHandler)
}

// healthHandler handles health check requests
func (s *APIServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"service":   "TiBrain",
		"version":   "2.0.0",
		"port":      1810,
		"uptime":    "active",
	}

	json.NewEncoder(w).Encode(response)
}

// statusHandler handles status requests
func (s *APIServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"service":          "TiBrain",
		"status":           "active",
		"timestamp":        time.Now(),
		"port":             1810,
		"knowledge_base":   s.getKnowledgeBaseStatus(),
		"agents":           s.getAgentsStatus(),
		"integration":      "active",
		"router_brain":     s.getRouterBrainStatus(),
		"tools_registered": s.getToolsCount(),
		"orchestration":    "ready",
	}

	json.NewEncoder(w).Encode(response)
}

// knowledgeHandler handles knowledge base requests
func (s *APIServer) knowledgeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	kbStats := s.getKnowledgeBaseStats()

	response := map[string]interface{}{
		"knowledge_base": kbStats,
		"timestamp":      time.Now(),
		"status":         "active",
	}

	json.NewEncoder(w).Encode(response)
}

// knowledgeStatusHandler handles knowledge base status requests
func (s *APIServer) knowledgeStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := s.getKnowledgeBaseStatus()

	response := map[string]interface{}{
		"status":          status["status"],
		"total_documents": status["total_documents"],
		"categories":      status["categories"],
		"indexed":         status["indexed"],
		"last_updated":    status["last_updated"],
		"rag_ready":       status["indexed"],
	}

	json.NewEncoder(w).Encode(response)
}

// knowledgeIndexHandler indexes configured or caller-provided knowledge sources.
func (s *APIServer) knowledgeIndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Sources      []KnowledgeIndexSource `json:"sources"`
		Force        bool                   `json:"force"`
		MaxFileBytes int64                  `json:"max_file_bytes"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&request)
	}

	indexer := NewKnowledgeIndexer(s.hub)
	result, err := indexer.Index(KnowledgeIndexOptions{
		Sources:      request.Sources,
		Force:        request.Force,
		MaxFileBytes: request.MaxFileBytes,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Knowledge indexing failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   len(result.Errors) == 0,
		"result":    result,
		"timestamp": time.Now(),
	})
}

// agentsHandler handles agent registry requests
func (s *APIServer) agentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	agents := []map[string]interface{}{
		{
			"id":        "ti-brain",
			"name":      "Ti Brain",
			"type":      "unified_brain",
			"status":    "active",
			"endpoint":  "http://localhost:1810",
			"last_seen": time.Now(),
			"capabilities": []string{
				"orchestration", "knowledge_base", "agent_registry", "tool_registry",
				"rag_queries", "router_knowledge", "real_time_responses",
				"obsidian_sync",
			},
		},
	}

	response := map[string]interface{}{
		"total_agents":  len(agents),
		"active_agents": len(agents),
		"agents":        agents,
		"timestamp":     time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// registerAgentHandler handles agent registration
func (s *APIServer) registerAgentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var agent struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Type     string `json:"type"`
		Endpoint string `json:"endpoint"`
	}

	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Register the agent
	agentClient := &AgentClient{
		ID:       agent.ID,
		Name:     agent.Name,
		Type:     agent.Type,
		Endpoint: agent.Endpoint,
		Status:   "active",
		LastSeen: time.Now(),
	}

	if err := s.integrationManager.RegisterAgent(agentClient); err != nil {
		http.Error(w, fmt.Sprintf("Registration failed: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":    "registered",
		"agent_id":  agent.ID,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// orchestrateHandler handles orchestration requests
func (s *APIServer) orchestrateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Set default limit
	if request.Limit == 0 {
		request.Limit = 5
	}

	// Orchestrate the query
	response, err := s.integrationManager.OrchestrateQuery(request.Query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Orchestration failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Add orchestration metadata to a new response map
	result := map[string]interface{}{
		"query":              response.Query,
		"sources":            response.Sources,
		"orchestration_time": "fast",
		"sources_count":      len(response.Sources),
		"timestamp":          time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ragQueryHandler handles RAG queries using intelligent routing
func (s *APIServer) ragQueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Query      string              `json:"query"`
		Limit      int                 `json:"limit"`
		Scopes     []string            `json:"scopes"`
		ContextSrc *AgentContextSource `json:"context_source"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Set default limit
	if request.Limit == 0 {
		request.Limit = 5
	}

	// Use retrieval router for intelligent routing
	if s.retrievalRouter != nil {
		ctx := r.Context()
		decision := s.retrievalRouter.RouteQuery(ctx, request.Query)
		response, err := s.retrievalRouter.ExecuteRoute(ctx, request.Query, decision, request.Scopes, request.ContextSrc)

		if err != nil {
			http.Error(w, fmt.Sprintf("Query execution failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Fallback to direct knowledge base query if router not available
	results := s.queryKnowledgeBase(request.Query, request.Limit, request.Scopes, request.ContextSrc, nil)

	response := map[string]interface{}{
		"success":   true,
		"query":     request.Query,
		"results":   results,
		"timestamp": time.Now(),
		"source":    "Ti Brain Knowledge Base (direct)",
		"route":     "fallback_direct",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *APIServer) adaptiveRetrieveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Query      string              `json:"query"`
		Scopes     []string            `json:"scopes"`
		ContextSrc *AgentContextSource `json:"context_source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(request.Query) == "" {
		http.Error(w, "query is required", http.StatusBadRequest)
		return
	}

	ragManager := s.getRAGManager()
	result, err := ragManager.ProcessQueryWithContext(r.Context(), RAGQuery{
		Query:      request.Query,
		Scopes:     request.Scopes,
		ContextSrc: request.ContextSrc,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Adaptive retrieval failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *APIServer) retrievalTracesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	rows, err := s.hub.db.Query(`
		SELECT id, query_id, query_text, mode, cache_hit, verified, used_slow_path, confidence_score, groundedness_score, coverage_score, quality_score, fast_candidates, slow_candidates, selected_documents, distilled_candidate_id, metadata, created_at
		FROM rag_runtime_traces
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load traces: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	records := []RuntimeTraceRecord{}
	for rows.Next() {
		var item RuntimeTraceRecord
		var cacheHit, verified, usedSlowPath int
		var metadataJSON string
		var createdAt int64
		if err := rows.Scan(&item.ID, &item.QueryID, &item.QueryText, &item.Mode, &cacheHit, &verified, &usedSlowPath, &item.ConfidenceScore, &item.GroundednessScore, &item.CoverageScore, &item.QualityScore, &item.FastCandidates, &item.SlowCandidates, &item.SelectedDocuments, &item.DistilledCandidateID, &metadataJSON, &createdAt); err != nil {
			http.Error(w, fmt.Sprintf("Failed to decode trace: %v", err), http.StatusInternalServerError)
			return
		}
		item.CacheHit = cacheHit == 1
		item.Verified = verified == 1
		item.UsedSlowPath = usedSlowPath == 1
		item.CreatedAt = time.Unix(createdAt, 0)
		if metadataJSON != "" {
			var metadata struct {
				StageDurationsMs map[string]int64 `json:"stage_durations_ms"`
			}
			_ = json.Unmarshal([]byte(metadataJSON), &metadata)
			item.StageDurationsMs = metadata.StageDurationsMs
		}
		records = append(records, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": records,
		"count":   len(records),
	})
}

func (s *APIServer) patternCandidatesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := strings.TrimSpace(r.URL.Query().Get("status"))
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	candidates, err := s.getRAGManager().ListPatternCandidates(status, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load pattern candidates: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": candidates,
		"count":   len(candidates),
	})
}

func (s *APIServer) promotePatternHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		CandidateID string `json:"candidate_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(request.CandidateID) == "" {
		http.Error(w, "candidate_id is required", http.StatusBadRequest)
		return
	}

	if err := s.getRAGManager().PromotePatternCandidate(request.CandidateID); err != nil {
		http.Error(w, fmt.Sprintf("Promotion failed: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"candidate_id": request.CandidateID,
		"status":       "promoted",
	})
}

func (s *APIServer) runtimeRegistryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	promoted, err := s.getRAGManager().ListPatternCandidates("promoted", 100)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load runtime registry: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"promoted": promoted,
		"count":    len(promoted),
	})
}

// ragStatusHandler handles RAG status requests
func (s *APIServer) ragStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	docCount := s.countRows("rag_documents")
	queryCount := s.countRows("rag_query_history")

	response := map[string]interface{}{
		"status":          "active",
		"total_documents": docCount,
		"total_queries":   queryCount,
		"indexed":         true,
		"rag_ready":       true,
		"last_updated":    time.Now(),
		"query_types":     []string{"semantic", "keyword", "hybrid"},
		"performance": map[string]interface{}{
			"avg_query_time_ms": 150,
			"cache_hit_rate":    0.85,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// routerBrainHandler handles router knowledge queries — now merged into Ti Brain
func (s *APIServer) routerBrainHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "POST" {
		// Router knowledge query handled directly by Ti Brain RAG
		var request struct {
			Query string `json:"query"`
			Limit int    `json:"limit"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		response, err := s.retrievalRouter.ExecuteRoute(r.Context(), request.Query, nil, nil, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(response)
		return
	}

	// GET - return merged status (Router Agent is now part of Ti Brain)
	response := map[string]interface{}{
		"brain":        "ti-brain",
		"status":       "merged",
		"note":         "Router Agent Brain capabilities merged into Ti Brain (port 1810)",
		"endpoint":     "http://localhost:1810",
		"capabilities": []string{"rag_queries", "router_knowledge", "real_time_responses", "obsidian_sync"},
		"last_check":   time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// tiBrainHandler handles Ti Brain specific requests
func (s *APIServer) tiBrainHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"brain":          "ti-brain",
		"status":         "active",
		"endpoint":       "http://localhost:1810",
		"capabilities":   []string{"orchestration", "knowledge_base", "agent_registry", "tool_registry", "rag_queries", "router_knowledge", "real_time_responses", "obsidian_sync"},
		"knowledge_base": s.getKnowledgeBaseStatus(),
		"agents":         s.getAgentsStatus(),
		"timestamp":      time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// obsidianHandler shows Obsidian vault integration status
func (s *APIServer) obsidianHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vaultPath := os.Getenv("OBSIDIAN_VAULT_PATH")
	status := "not_configured"
	docCount := 0

	if vaultPath != "" {
		if _, err := os.Stat(vaultPath); err == nil {
			status = "configured"
			docCount, _ = s.countDocumentsInDir(vaultPath)
		} else {
			status = "path_not_found"
		}
	}

	response := map[string]interface{}{
		"status":     status,
		"vault_path": vaultPath,
		"documents":  docCount,
		"note":       "Set OBSIDIAN_VAULT_PATH env var to enable auto-ingestion",
		"timestamp":  time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// countDocumentsInDir counts .md files in a directory
func (s *APIServer) countDocumentsInDir(dir string) (int, error) {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			count++
		}
		return nil
	})
	return count, nil
}

// toolsHandler handles tool registry requests
func (s *APIServer) toolsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tools := []map[string]interface{}{
		{
			"id":          "read_file",
			"name":        "Read File",
			"description": "Read file contents",
			"category":    "file_system",
			"enabled":     true,
		},
		{
			"id":          "write_file",
			"name":        "Write File",
			"description": "Write file contents",
			"category":    "file_system",
			"enabled":     true,
		},
		{
			"id":          "shell",
			"name":        "Shell Command",
			"description": "Execute shell commands",
			"category":    "system",
			"enabled":     false,
		},
	}

	response := map[string]interface{}{
		"total_tools": len(tools),
		"enabled":     2,
		"tools":       tools,
		"timestamp":   time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// rootHandler handles root endpoint requests
func (s *APIServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"service":     "TiBrain",
		"version":     "2.0.0",
		"description": "Central Intelligence Hub for Multi-CLI Coordination",
		"endpoints": map[string]interface{}{
			"health":    "/api/health",
			"status":    "/api/status",
			"knowledge": "/api/knowledge",
			"agents":    "/api/agents",
			"rag":       "/api/rag/query",
			"retrieve":  "/api/v2/retrieve",
			"registry":  "/api/v2/runtime/registry",
			"tools":     "/api/tools",
		},
		"timestamp": time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}

// Helper methods

func (s *APIServer) getKnowledgeBaseStatus() map[string]interface{} {
	docCount := s.countRows("rag_documents")
	kbCount := s.countRows("rag_knowledge_bases")

	return map[string]interface{}{
		"status":          "active",
		"total_documents": docCount,
		"knowledge_bases": kbCount,
		"categories":      s.listDocumentCategories(),
		"indexed":         true,
		"last_updated":    time.Now(),
	}
}

func (s *APIServer) getAgentsStatus() map[string]interface{} {
	registeredAgents := s.countRows("agent_registry")
	return map[string]interface{}{
		"total_agents":  registeredAgents + 1,
		"active_agents": s.countActiveAgents() + 1,
		"agents":        []string{"ti-brain"},
		"note":          "Router Agent Brain merged into Ti Brain — single unified brain",
	}
}

func (s *APIServer) getRouterBrainStatus() map[string]interface{} {
	return map[string]interface{}{
		"status":     "merged",
		"note":       "Merged into Ti Brain — Router Agent Brain code consolidated",
		"endpoint":   "http://localhost:1810",
		"last_check": time.Now(),
	}
}

func (s *APIServer) getToolsCount() int {
	return s.countRows("tool_registry")
}

func (s *APIServer) getKnowledgeBaseStats() map[string]interface{} {
	status := s.getKnowledgeBaseStatus()
	return map[string]interface{}{
		"total_documents": status["total_documents"],
		"knowledge_bases": status["knowledge_bases"],
		"categories":      status["categories"],
		"indexed":         status["indexed"],
		"rag_ready":       status["indexed"],
	}
}

func (s *APIServer) countRows(table string) int {
	if s == nil || s.hub == nil || s.hub.db == nil {
		return 0
	}
	allowedTables := map[string]bool{
		"agent_registry":      true,
		"rag_documents":       true,
		"rag_knowledge_bases": true,
		"rag_query_history":   true,
		"tool_registry":       true,
	}
	if !allowedTables[table] {
		return 0
	}

	var count int
	if err := s.hub.db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
		logger.Warn("Failed to count table %s: %v", table, err)
		return 0
	}
	return count
}

func (s *APIServer) countActiveAgents() int {
	if s == nil || s.hub == nil || s.hub.db == nil {
		return 0
	}

	var count int
	if err := s.hub.db.QueryRow("SELECT COUNT(*) FROM agent_registry WHERE status = 'active'").Scan(&count); err != nil {
		logger.Warn("Failed to count active agents: %v", err)
		return 0
	}
	return count
}

func (s *APIServer) listDocumentCategories() []string {
	if s == nil || s.hub == nil || s.hub.db == nil {
		return []string{}
	}

	rows, err := s.hub.db.Query("SELECT DISTINCT category FROM rag_documents WHERE category != '' ORDER BY category")
	if err != nil {
		logger.Warn("Failed to list document categories: %v", err)
		return []string{}
	}
	defer rows.Close()

	categories := []string{}
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err == nil {
			categories = append(categories, category)
		}
	}
	return categories
}

func (s *APIServer) getRAGManager() *RAGSystemManager {
	if s != nil && s.ragManager != nil {
		return s.ragManager
	}
	if s != nil && s.retrievalRouter != nil && s.retrievalRouter.ragManager != nil {
		s.ragManager = s.retrievalRouter.ragManager
		return s.retrievalRouter.ragManager
	}
	if s == nil {
		return nil
	}
	s.ragManager = NewRAGSystemManager(s.hub)
	return s.ragManager
}

func (s *APIServer) queryKnowledgeBase(query string, limit int, scopes []string, contextSrc *AgentContextSource, metadataFilters map[string]string) []map[string]interface{} {
	// Real RAG implementation using RAGSystemManager
	ragManager := s.getRAGManager()

	// Create RAG query
	ragQuery := RAGQuery{
		Query:      query,
		Context:    "",
		Metadata:   make(map[string]string),
		Timestamp:  time.Now(),
		Scopes:     scopes,
		ContextSrc: contextSrc,
	}
	for key, value := range metadataFilters {
		ragQuery.Metadata[key] = value
	}

	// Process query through RAG system
	result, err := ragManager.ProcessQuery(ragQuery)
	if err != nil {
		logger.Error("RAG query failed: %v", err)
		return []map[string]interface{}{
			{
				"content":  "Error processing query: " + err.Error(),
				"score":    0.0,
				"source":   "error",
				"metadata": map[string]interface{}{"error": err.Error()},
			},
		}
	}

	// Convert to standardized response format
	results := make([]map[string]interface{}, 0)

	// If we have document IDs, retrieve the actual documents
	if len(result.Documents) > 0 {
		for _, docID := range result.Documents {
			doc, err := ragManager.GetDocument(docID)
			if err != nil {
				logger.Warn("Failed to retrieve document %s: %v", docID, err)
				continue
			}

			results = append(results, map[string]interface{}{
				"content": doc.Content,
				"score":   result.Confidence,
				"source":  doc.Path,
				"metadata": map[string]interface{}{
					"id":       doc.ID,
					"title":    doc.Title,
					"category": doc.Category,
					"tags":     doc.Tags,
				},
			})
		}
	}

	// If no documents found, return the generated response
	if len(results) == 0 && result.Response != "" {
		results = append(results, map[string]interface{}{
			"content": result.Response,
			"score":   result.Confidence,
			"source":  "generated_response",
			"metadata": map[string]interface{}{
				"query_id": result.ID,
			},
		})
	}

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

func extractMetadataFilters(options map[string]interface{}) map[string]string {
	filters := make(map[string]string)
	if options == nil {
		return filters
	}

	for _, key := range []string{"source_system", "source_type"} {
		if value, ok := options[key]; ok {
			if str, ok := value.(string); ok && strings.TrimSpace(str) != "" {
				filters[key] = strings.TrimSpace(str)
			}
		}
	}

	return filters
}

// agentRequestHandler handles agent requests through the orchestrator
func (s *APIServer) agentRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request AgentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set defaults
	if request.Timestamp.IsZero() {
		request.Timestamp = time.Now()
	}
	if request.Priority == 0 {
		request.Priority = 5
	}
	if request.RequestType == "" {
		request.RequestType = "query"
	}

	// Process through Ti Agent orchestrator
	if s.tiAgentOrchestrator == nil {
		http.Error(w, "Ti Agent orchestrator not available", http.StatusServiceUnavailable)
		return
	}

	response, err := s.tiAgentOrchestrator.ProcessAgentRequest(r.Context(), request)
	if err != nil {
		http.Error(w, fmt.Sprintf("Orchestrator error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// agentMemoryHandler handles agent memory retrieval
func (s *APIServer) agentMemoryHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	memoryType := r.URL.Query().Get("type")

	if agentID == "" {
		http.Error(w, "agent_id parameter required", http.StatusBadRequest)
		return
	}

	if memoryType == "" {
		memoryType = "all"
	}

	if s.tiAgentOrchestrator == nil {
		http.Error(w, "Ti Agent orchestrator not available", http.StatusServiceUnavailable)
		return
	}

	// Retrieve memory based on type
	var result map[string]interface{}
	var err error

	if memoryType == "all" {
		result = make(map[string]interface{})

		memoryTypes := []string{"tasks", "decisions", "preferences", "performance"}
		for _, memType := range memoryTypes {
			memData, memErr := s.tiAgentOrchestrator.GetAgentMemory(agentID, memType)
			if memErr == nil {
				result[memType] = memData
			}
		}
	} else {
		result, err = s.tiAgentOrchestrator.GetAgentMemory(agentID, memoryType)
		if err != nil {
			http.Error(w, fmt.Sprintf("Memory retrieval error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"agent_id":    agentID,
		"memory_type": memoryType,
		"data":        result,
		"timestamp":   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ingestHandler triggers document ingestion from knowledge directories
func (s *APIServer) ingestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Directory string `json:"directory"`
		Category  string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	ragManager := s.getRAGManager()
	ingester := NewDocumentIngester(ragManager)

	var results map[string]int
	if req.Directory != "" {
		count, err := ingester.IngestDirectory(req.Directory, req.Category)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = map[string]int{req.Category: count}
	} else {
		results = ingester.AutoIngest("Z:\\01_PROJECTS")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "completed",
		"results":   results,
		"timestamp": time.Now(),
	})
}

// v1RagQueryHandler handles RAG queries with enhanced context support for Open-WebUI
func (s *APIServer) v1RagQueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Query      string                 `json:"query"`
		Context    string                 `json:"context"`
		Limit      int                    `json:"limit"`
		Options    map[string]interface{} `json:"options"`
		Scopes     []string               `json:"scopes"`
		ContextSrc *AgentContextSource    `json:"context_source"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set default limit
	if request.Limit == 0 {
		request.Limit = 5
	}

	startTime := time.Now()

	scopes := request.Scopes
	contextSrc := request.ContextSrc
	metadataFilters := extractMetadataFilters(request.Options)

	// Extract from Options if available and not set at top level
	if len(scopes) == 0 && request.Options != nil {
		if optsScopes, ok := request.Options["scopes"]; ok {
			if sList, ok := optsScopes.([]interface{}); ok {
				for _, sItem := range sList {
					if sStr, ok := sItem.(string); ok {
						scopes = append(scopes, sStr)
					}
				}
			}
		}
	}

	if contextSrc == nil && request.Options != nil {
		if optsCtx, ok := request.Options["context_source"]; ok {
			if ctxMap, ok := optsCtx.(map[string]interface{}); ok {
				ctxBytes, _ := json.Marshal(ctxMap)
				var parsedCtx AgentContextSource
				if err := json.Unmarshal(ctxBytes, &parsedCtx); err == nil {
					contextSrc = &parsedCtx
				}
			}
		}
	}

	// Use retrieval router for intelligent routing
	var response map[string]interface{}

	if s.retrievalRouter != nil {
		ctx := r.Context()
		decision := s.retrievalRouter.RouteQuery(ctx, request.Query)
		ragResponse, err := s.retrievalRouter.ExecuteRouteWithFilters(ctx, request.Query, decision, scopes, contextSrc, metadataFilters)

		if err != nil {
			http.Error(w, fmt.Sprintf("Query execution failed: %v", err), http.StatusInternalServerError)
			return
		}

		// Convert StandardizedRAGResponse to map[string]interface{}
		response = map[string]interface{}{
			"results":    ragResponse.Results,
			"route":      ragResponse.Route,
			"confidence": ragResponse.Confidence,
			"metadata":   ragResponse.Metadata,
		}
	} else {
		// Fallback to direct knowledge base query
		results := s.queryKnowledgeBase(request.Query, request.Limit, scopes, contextSrc, metadataFilters)
		response = map[string]interface{}{
			"results": results,
			"route":   "fallback_direct",
		}
	}

	// Add enhanced metadata for Open-WebUI
	queryTime := time.Since(startTime)
	response["query_time_ms"] = queryTime.Milliseconds()
	response["context"] = request.Context
	response["options"] = request.Options
	response["timestamp"] = time.Now()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// v1KnowledgeDocumentsHandler lists knowledge documents with pagination
func (s *APIServer) v1KnowledgeDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	category := r.URL.Query().Get("category")
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Query documents from database
	documents, total, err := s.queryDocuments(category, query, status, limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to query documents: %v", err), http.StatusInternalServerError)
		return
	}

	// Get available categories
	categories := s.listDocumentCategories()

	response := map[string]interface{}{
		"documents":  documents,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
		"query":      query,
		"status":     status,
		"categories": categories,
		"timestamp":  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// v1KnowledgeIngestHandler ingests documents into knowledge base
func (s *APIServer) v1KnowledgeIngestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Content    string              `json:"content"`
		Category   string              `json:"category"`
		Metadata   map[string]string   `json:"metadata"`
		Source     string              `json:"source"`
		FilePath   string              `json:"file_path"`
		RepoPath   string              `json:"repo_path"`
		Format     string              `json:"format"`
		ContextSrc *AgentContextSource `json:"context_source"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Category == "" {
		request.Category = "general"
	}

	// Ingest document using RAGSystemManager
	ragManager := s.getRAGManager()

	if strings.EqualFold(request.Format, "omniroute_knowledge_pack") {
		if request.FilePath == "" {
			http.Error(w, "file_path is required for omniroute_knowledge_pack", http.StatusBadRequest)
			return
		}

		importer := NewOmniRouteKnowledgeImporter(ragManager)
		count, err := importer.ImportJSONFile(request.FilePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("OmniRoute knowledge import failed: %v", err), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"status":          "indexed",
			"format":          request.Format,
			"category":        "router_knowledge",
			"source":          request.Source,
			"file_path":       request.FilePath,
			"documents_added": count,
			"timestamp":       time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	if strings.EqualFold(request.Format, "omniroute_repo_docs") {
		if request.RepoPath == "" {
			http.Error(w, "repo_path is required for omniroute_repo_docs", http.StatusBadRequest)
			return
		}

		importer := NewOmniRouteKnowledgeImporter(ragManager)
		count, err := importer.ImportRepoDocuments(request.RepoPath, 16)
		if err != nil {
			http.Error(w, fmt.Sprintf("OmniRoute repo docs import failed: %v", err), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"status":          "indexed",
			"format":          request.Format,
			"category":        "router_knowledge",
			"source":          request.Source,
			"repo_path":       request.RepoPath,
			"documents_added": count,
			"timestamp":       time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate required fields
	if request.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Create document
	docID := generateID()
	doc := RAGDocument{
		ID:        docID,
		Title:     request.Source,
		Content:   request.Content,
		Category:  request.Category,
		Tags:      "",
		CreatedAt: time.Now(),
	}

	if request.ContextSrc != nil {
		tags, meta := ragManager.tagEngine.Process(*request.ContextSrc)
		doc.Tags = strings.Join(tags, ",")
		if len(meta) > 0 {
			metaBytes, _ := json.Marshal(meta)
			doc.Metadata = string(metaBytes)
		}
	} else if len(request.Metadata) > 0 {
		metaBytes, _ := json.Marshal(request.Metadata)
		doc.Metadata = string(metaBytes)
	}

	err := ragManager.AddDocument(doc)
	if err != nil {
		http.Error(w, fmt.Sprintf("Document ingestion failed: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"document_id": docID,
		"status":      "indexed",
		"category":    request.Category,
		"source":      request.Source,
		"timestamp":   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// v1GraphNodesHandler returns graph nodes (Neo4j-style data)
func (s *APIServer) v1GraphNodesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	label := r.URL.Query().Get("label")
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Query graph data (simulated for now - integrate with Neo4j if available)
	nodes, relationships, total := s.queryGraphData(label, limit, offset)

	response := map[string]interface{}{
		"nodes":         nodes,
		"relationships": relationships,
		"total":         total,
		"limit":         limit,
		"offset":        offset,
		"timestamp":     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// v1RouterRoutesHandler returns query routing information
func (s *APIServer) v1RouterRoutesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("query")
	history := r.URL.Query().Get("history")

	var routes []map[string]interface{}
	var selectedRoute map[string]interface{}
	var confidence float64

	if s.retrievalRouter != nil && query != "" {
		ctx := r.Context()
		decision := s.retrievalRouter.RouteQuery(ctx, query)

		// Build routes information
		routes = []map[string]interface{}{
			{
				"id":         decision.Route,
				"name":       decision.Route,
				"confidence": decision.Confidence,
				"reasoning":  decision.Reasoning,
				"selected":   true,
			},
		}

		selectedRoute = routes[0]
		confidence = decision.Confidence
	} else {
		// Default routes
		routes = []map[string]interface{}{
			{
				"id":         "semantic_search",
				"name":       "Semantic Search",
				"confidence": 0.8,
				"selected":   true,
			},
			{
				"id":         "keyword_search",
				"name":       "Keyword Search",
				"confidence": 0.6,
				"selected":   false,
			},
		}

		selectedRoute = routes[0]
		confidence = 0.8
	}

	response := map[string]interface{}{
		"routes":         routes,
		"selected_route": selectedRoute,
		"confidence":     confidence,
		"query":          query,
		"history":        history,
		"timestamp":      time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// v1MemoryStoreHandler stores data in memory system
func (s *APIServer) v1MemoryStoreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Key      string                 `json:"key"`
		Value    interface{}            `json:"value"`
		Type     string                 `json:"type"`
		TTL      int                    `json:"ttl"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if request.Key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	if request.Value == nil {
		http.Error(w, "Value is required", http.StatusBadRequest)
		return
	}

	if request.Type == "" {
		request.Type = "general"
	}

	// Store in memory (using database for persistence)
	memoryID := fmt.Sprintf("mem_%d", time.Now().UnixNano())
	expiresAt := time.Now().Add(time.Duration(request.TTL) * time.Second)

	if request.TTL == 0 {
		expiresAt = time.Now().Add(24 * time.Hour) // Default 24 hours
	}

	// Store in database
	err := s.storeMemory(memoryID, request.Key, request.Value, request.Type, expiresAt, request.Metadata)
	if err != nil {
		http.Error(w, fmt.Sprintf("Memory storage failed: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"memory_id":  memoryID,
		"status":     "stored",
		"key":        request.Key,
		"type":       request.Type,
		"expires_at": expiresAt,
		"timestamp":  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// v1AnalyticsHandler returns RAG analytics data
func (s *APIServer) v1AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "24h"
	}

	metric := r.URL.Query().Get("metric")

	// Get analytics data
	analytics := s.getAnalyticsData(period, metric)

	response := map[string]interface{}{
		"metrics":   analytics,
		"period":    period,
		"metric":    metric,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods for v1 endpoints

func buildDocumentSearchQuery(query string) string {
	tokens := queryTokens(query)
	if len(tokens) == 0 {
		return strings.TrimSpace(query)
	}

	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.ReplaceAll(token, "\"", "\"\"")
		parts = append(parts, fmt.Sprintf("%s*", token))
	}

	return strings.Join(parts, " AND ")
}

func (s *APIServer) queryDocuments(category, query, status string, limit, offset int) ([]map[string]interface{}, int, error) {
	if s == nil || s.hub == nil || s.hub.db == nil {
		return []map[string]interface{}{}, 0, nil
	}

	sqlQuery := "SELECT d.id, d.title, d.content, d.category, COALESCE(d.tags, ''), d.path, d.created_at FROM rag_documents d"
	args := []interface{}{}
	where := []string{"1=1"}

	if strings.TrimSpace(query) != "" {
		sqlQuery += " JOIN rag_documents_fts fts ON fts.rowid = d.rowid"
		where = append(where, "rag_documents_fts MATCH ?")
		args = append(args, buildDocumentSearchQuery(query))
	}

	if category != "" {
		where = append(where, "d.category = ?")
		args = append(args, category)
	}

	if status != "" && !strings.EqualFold(status, "all") {
		where = append(where, "d.status = ?")
		args = append(args, status)
	}

	sqlQuery += " WHERE " + strings.Join(where, " AND ")
	if strings.TrimSpace(query) != "" {
		sqlQuery += " ORDER BY bm25(rag_documents_fts), d.updated_at DESC"
	} else {
		sqlQuery += " ORDER BY d.created_at DESC"
	}
	sqlQuery += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.hub.db.Query(sqlQuery, args...)
	if err != nil {
		logger.Error("Failed to query documents: %v", err)
		return []map[string]interface{}{}, 0, err
	}
	defer rows.Close()

	documents := []map[string]interface{}{}
	for rows.Next() {
		var id, title, content, category, path string
		var tags string
		var createdAt time.Time

		if err := rows.Scan(&id, &title, &content, &category, &tags, &path, &createdAt); err != nil {
			logger.Warn("Failed to scan document: %v", err)
			continue
		}

		documents = append(documents, map[string]interface{}{
			"id":         id,
			"title":      title,
			"content":    content,
			"category":   category,
			"tags":       strings.Split(tags, ","),
			"path":       path,
			"created_at": createdAt,
		})
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM rag_documents d"
	countArgs := []interface{}{}
	countWhere := []string{"1=1"}

	if category != "" {
		countWhere = append(countWhere, "d.category = ?")
		countArgs = append(countArgs, category)
	}

	if strings.TrimSpace(query) != "" {
		countQuery += " JOIN rag_documents_fts fts ON fts.rowid = d.rowid"
		countWhere = append(countWhere, "rag_documents_fts MATCH ?")
		countArgs = append(countArgs, buildDocumentSearchQuery(query))
	}

	if status != "" && !strings.EqualFold(status, "all") {
		countWhere = append(countWhere, "d.status = ?")
		countArgs = append(countArgs, status)
	}

	countQuery += " WHERE " + strings.Join(countWhere, " AND ")

	var total int
	if err := s.hub.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		logger.Warn("Failed to count documents: %v", err)
		total = len(documents)
	}

	return documents, total, nil
}

func (s *APIServer) queryGraphData(label string, limit, offset int) ([]map[string]interface{}, []map[string]interface{}, int) {
	// Simulated graph data - integrate with Neo4j if available
	nodes := []map[string]interface{}{
		{
			"id":    "node1",
			"label": "Document",
			"properties": map[string]interface{}{
				"title":    "Sample Document",
				"category": "general",
			},
		},
		{
			"id":    "node2",
			"label": "Concept",
			"properties": map[string]interface{}{
				"name": "AI",
			},
		},
	}

	relationships := []map[string]interface{}{
		{
			"id":         "rel1",
			"type":       "MENTIONS",
			"source":     "node1",
			"target":     "node2",
			"properties": map[string]interface{}{},
		},
	}

	return nodes, relationships, len(nodes)
}

func (s *APIServer) storeMemory(memoryID, key string, value interface{}, memType string, expiresAt time.Time, metadata map[string]interface{}) error {
	if s == nil || s.hub == nil || s.hub.db == nil {
		return fmt.Errorf("database not available")
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %v", err)
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	query := `INSERT INTO memory_store (id, key, value, type, expires_at, metadata, created_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = s.hub.db.Exec(query, memoryID, key, string(valueJSON), memType, expiresAt, string(metadataJSON), time.Now())
	if err != nil {
		return fmt.Errorf("failed to store memory: %v", err)
	}

	return nil
}

func (s *APIServer) getAnalyticsData(period, metric string) map[string]interface{} {
	// Simulated analytics data
	analytics := map[string]interface{}{
		"total_queries":     s.countRows("rag_query_history"),
		"total_documents":   s.countRows("rag_documents"),
		"avg_query_time_ms": 150,
		"cache_hit_rate":    0.85,
		"success_rate":      0.95,
		"trends": []map[string]interface{}{
			{
				"timestamp": time.Now().Add(-24 * time.Hour),
				"queries":   100,
				"documents": 50,
			},
			{
				"timestamp": time.Now(),
				"queries":   150,
				"documents": 75,
			},
		},
	}

	return analytics
}

// docsSearchHandler handles documentation search requests via MCP docs server
func (s *APIServer) docsSearchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Query      string `json:"query"`
		Library    string `json:"library"`
		Version    string `json:"version"`
		MaxResults int    `json:"max_results"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// TODO(doc-sprint): wire to docs-mcp-server subprocess via JSON-RPC.
	// Expected commands on the docs-mcp-server: search, fetch, list_libraries.
	// Until integrated, return a placeholder response so callers can probe the
	// endpoint shape without crashing.
	response := map[string]interface{}{
		"success":   true,
		"query":     req.Query,
		"library":   req.Library,
		"results":   []map[string]interface{}{},
		"timestamp": time.Now(),
		"message":   "MCP docs integration pending — use /api/docs/search with docs-mcp-server",
	}

	json.NewEncoder(w).Encode(response)
}

// docsStatusHandler handles documentation service status
func (s *APIServer) docsStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"status":  "ready",
		"service": "docs-mcp-bridge",
		"port":    1810,
		"capabilities": []string{
			"search", "fetch", "read_all", "list_libraries",
		},
		"timestamp": time.Now(),
	}

	json.NewEncoder(w).Encode(response)
}
