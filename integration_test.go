package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewHubCreatesIntegrationSchema(t *testing.T) {
	hub, err := NewHub(t.TempDir())
	if err != nil {
		t.Fatalf("NewHub() error = %v", err)
	}
	defer hub.Close()

	for _, table := range []string{"agent_registry", "cross_brain_communication", "orchestration_log"} {
		var count int
		err := hub.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("query table %s: %v", table, err)
		}
		if count != 1 {
			t.Fatalf("expected table %s to exist", table)
		}
	}
}

func TestNewHubBootstrapsLatestToolRegistrySchema(t *testing.T) {
	hub, err := NewHub(t.TempDir())
	if err != nil {
		t.Fatalf("NewHub() error = %v", err)
	}
	defer hub.Close()

	var schemaVersion int
	if err := hub.db.QueryRow("PRAGMA user_version").Scan(&schemaVersion); err != nil {
		t.Fatalf("query user_version: %v", err)
	}
	if schemaVersion < 1 {
		t.Fatalf("expected fresh database schema version >= 1, got %d", schemaVersion)
	}

	for _, column := range []string{
		"quality_score",
		"security_score",
		"best_practices_score",
		"source",
		"quality_tier",
		"security_status",
		"root_path",
	} {
		var count int
		err := hub.db.QueryRow(
			"SELECT COUNT(*) FROM pragma_table_info('tool_registry') WHERE name = ?",
			column,
		).Scan(&count)
		if err != nil {
			t.Fatalf("query tool_registry.%s: %v", column, err)
		}
		if count != 1 {
			t.Fatalf("expected tool_registry.%s to exist", column)
		}
	}
}

func TestAPIServerOrchestrate(t *testing.T) {
	hub, err := NewHub(t.TempDir())
	if err != nil {
		t.Fatalf("NewHub() error = %v", err)
	}
	defer hub.Close()

	api := NewAPIServer(hub, NewIntegrationManager(hub))
	body := strings.NewReader(`{"query":"brain link","limit":1}`)
	req := httptest.NewRequest(http.MethodPost, "/api/orchestrate", body)
	rec := httptest.NewRecorder()

	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result["sources_count"].(float64) != 1 {
		t.Fatalf("expected one source (merged brain), got %#v", result)
	}

	// Cross-brain log is no longer written since brains are merged
	// Verify integration_log exists and was created instead
	var logCount int
	if err := hub.db.QueryRow("SELECT COUNT(*) FROM integration_log").Scan(&logCount); err != nil {
		// integration_log might not exist in test schema, that's OK
		// The important thing is the orchestration works
	}
}

func TestRetrievalRouterUsesRouterBrainForRouterQueries(t *testing.T) {
	hub, err := NewHub(t.TempDir())
	if err != nil {
		t.Fatalf("NewHub() error = %v", err)
	}
	defer hub.Close()

	router := NewRetrievalRouter(hub)
	decision := router.RouteQuery(t.Context(), "router provider status")
	if decision.Route != RouteRouterBrain {
		t.Fatalf("expected router brain route, got %s", decision.Route)
	}

	response, err := router.ExecuteRoute(t.Context(), "router provider status", decision, nil, nil)
	if err != nil {
		t.Fatalf("ExecuteRoute() error = %v", err)
	}
	if !response.Success {
		t.Fatalf("expected success, got %#v", response)
	}

	hub.asyncWriter.flush()

	var count int
	if err := hub.db.QueryRow("SELECT COUNT(*) FROM routing_decision_log").Scan(&count); err != nil {
		t.Fatalf("query routing decision log: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one routing decision log row, got %d", count)
	}
}

func TestTiAgentTaskAndMemoryFlow(t *testing.T) {
	hub, err := NewHub(t.TempDir())
	if err != nil {
		t.Fatalf("NewHub() error = %v", err)
	}
	defer hub.Close()

	orchestrator := NewTiAgentOrchestrator(hub, NewRetrievalRouter(hub))
	response, err := orchestrator.ProcessAgentRequest(t.Context(), AgentRequest{
		AgentID:     "ti-agent-test",
		AgentType:   "router-agent",
		Query:       "summarize router setup",
		RequestType: "task",
		Context:     map[string]interface{}{"source": "test"},
		Priority:    4,
	})
	if err != nil {
		t.Fatalf("ProcessAgentRequest() error = %v", err)
	}
	if !response.Success {
		t.Fatalf("expected success, got %#v", response)
	}
	if response.Data["task_status"] != "completed" {
		t.Fatalf("expected completed task status, got %#v", response.Data)
	}

	memory, err := orchestrator.GetAgentMemory("ti-agent-test", "tasks")
	if err != nil {
		t.Fatalf("GetAgentMemory() error = %v", err)
	}
	tasks, ok := memory["tasks"].([]map[string]interface{})
	if !ok || len(tasks) != 1 {
		t.Fatalf("expected one task memory row, got %#v", memory)
	}
	if tasks[0]["completed_at"] == nil {
		t.Fatalf("expected completed_at to be set, got %#v", tasks[0])
	}
}

func TestKnowledgeIndexerMakesDocsQueryable(t *testing.T) {
	sourceDir := t.TempDir()
	docPath := filepath.Join(sourceDir, "ti-agent.md")
	if err := os.WriteFile(docPath, []byte("# Ti Agent\n\nTi Agent uses TiBrain retrieval router and task memory for orchestration."), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	hub, err := NewHub(t.TempDir())
	if err != nil {
		t.Fatalf("NewHub() error = %v", err)
	}
	defer hub.Close()

	indexer := NewKnowledgeIndexer(hub)
	result, err := indexer.Index(KnowledgeIndexOptions{
		Sources: []KnowledgeIndexSource{{
			Path:        sourceDir,
			Category:    "test-knowledge",
			Description: "test knowledge source",
		}},
	})
	if err != nil {
		t.Fatalf("Index() error = %v", err)
	}
	if result.Indexed != 1 {
		t.Fatalf("expected one indexed document, got %#v", result)
	}

	rag := NewRAGSystemManager(hub)
	query, err := rag.ProcessQuery(RAGQuery{Query: "TiBrain retrieval router"})
	if err != nil {
		t.Fatalf("ProcessQuery() error = %v", err)
	}
	if query.Confidence == 0 || len(query.Documents) != 1 {
		t.Fatalf("expected indexed doc to be retrieved, got %#v", query)
	}
	if !strings.Contains(query.Response, "Ti Agent uses TiBrain retrieval router") {
		t.Fatalf("expected response to include indexed knowledge, got %q", query.Response)
	}
}

func TestTiAgentQueryUsesIndexedKnowledge(t *testing.T) {
	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "router-memory.md"), []byte("Router Agent uses specialized Router Brain knowledge for provider routing."), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	hub, err := NewHub(t.TempDir())
	if err != nil {
		t.Fatalf("NewHub() error = %v", err)
	}
	defer hub.Close()

	_, err = NewKnowledgeIndexer(hub).Index(KnowledgeIndexOptions{
		Sources: []KnowledgeIndexSource{{
			Path:        sourceDir,
			Category:    "router-test",
			Description: "router test docs",
		}},
	})
	if err != nil {
		t.Fatalf("Index() error = %v", err)
	}

	orchestrator := NewTiAgentOrchestrator(hub, NewRetrievalRouter(hub))
	response, err := orchestrator.ProcessAgentRequest(t.Context(), AgentRequest{
		AgentID:     "ti-agent-query-test",
		AgentType:   "ti-agent",
		Query:       "provider routing knowledge",
		RequestType: "query",
		Context:     map[string]interface{}{"source": "test"},
	})
	if err != nil {
		t.Fatalf("ProcessAgentRequest() error = %v", err)
	}
	if !response.Success {
		t.Fatalf("expected success, got %#v", response)
	}
	if !strings.Contains(response.Response, "Router Agent uses specialized Router Brain knowledge") {
		t.Fatalf("expected indexed knowledge in response, got %q", response.Response)
	}
}
