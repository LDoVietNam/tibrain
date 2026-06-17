package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdaptiveRetrievalDistillsPromotableCandidate(t *testing.T) {
	sourceDir := t.TempDir()
	docPath := filepath.Join(sourceDir, "routing-playbook.md")
	content := strings.Join([]string{
		"# Router Retrieval Playbook",
		"",
		"Provider routing in TiBrain should go through the retrieval router.",
		"Validated sources must be scored before they become reusable runtime logic.",
		"Promotion into runtime only happens after verification and distillation.",
	}, "\n")
	if err := os.WriteFile(docPath, []byte(content), 0644); err != nil {
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
			Category:    "playbooks",
			Description: "routing playbooks",
		}},
	})
	if err != nil {
		t.Fatalf("Index() error = %v", err)
	}

	rag := NewRAGSystemManager(hub)
	result, err := rag.ProcessQueryWithContext(t.Context(), RAGQuery{
		Query: "How should provider routing be verified before promotion into runtime logic?",
	})
	if err != nil {
		t.Fatalf("ProcessQueryWithContext() error = %v", err)
	}
	if result.Runtime == nil {
		t.Fatalf("expected runtime trace, got %#v", result)
	}
	if result.Runtime.QualityScore <= 0 {
		t.Fatalf("expected positive quality score, got %#v", result.Runtime)
	}
	if !result.Runtime.Verified {
		t.Fatalf("expected runtime trace to be verified, got %#v", result.Runtime)
	}

	hub.asyncWriter.flush()

	var traceCount int
	if err := hub.db.QueryRow("SELECT COUNT(*) FROM rag_runtime_traces").Scan(&traceCount); err != nil {
		t.Fatalf("query rag_runtime_traces: %v", err)
	}
	if traceCount != 1 {
		t.Fatalf("expected one runtime trace, got %d", traceCount)
	}

	var candidateCount int
	if err := hub.db.QueryRow("SELECT COUNT(*) FROM brain_pattern_candidates WHERE status = 'candidate'").Scan(&candidateCount); err != nil {
		t.Fatalf("query brain_pattern_candidates: %v", err)
	}
	if candidateCount != 1 {
		t.Fatalf("expected one pattern candidate, got %d", candidateCount)
	}
}

func TestAPIV2RetrieveReportsCacheHits(t *testing.T) {
	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "memory.md"), []byte("Task memory must be verified before reuse in runtime orchestration."), 0644); err != nil {
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
			Category:    "memory",
			Description: "memory docs",
		}},
	})
	if err != nil {
		t.Fatalf("Index() error = %v", err)
	}

	api := NewAPIServer(hub, NewIntegrationManager(hub))
	body := []byte(`{"query":"verified task memory reuse","limit":3}`)

	req1 := httptest.NewRequest(http.MethodPost, "/api/v2/retrieve", bytes.NewReader(body))
	rec1 := httptest.NewRecorder()
	api.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected first request status 200, got %d: %s", rec1.Code, rec1.Body.String())
	}

	var first struct {
		Runtime struct {
			CacheHit bool `json:"cache_hit"`
		} `json:"runtime"`
	}
	if err := json.Unmarshal(rec1.Body.Bytes(), &first); err != nil {
		t.Fatalf("decode first response: %v", err)
	}
	if first.Runtime.CacheHit {
		t.Fatalf("expected first request to miss cache")
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v2/retrieve", bytes.NewReader(body))
	rec2 := httptest.NewRecorder()
	api.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected second request status 200, got %d: %s", rec2.Code, rec2.Body.String())
	}

	var second struct {
		Runtime struct {
			CacheHit bool `json:"cache_hit"`
		} `json:"runtime"`
	}
	if err := json.Unmarshal(rec2.Body.Bytes(), &second); err != nil {
		t.Fatalf("decode second response: %v", err)
	}
	if !second.Runtime.CacheHit {
		t.Fatalf("expected second request to hit cache")
	}
}

func TestPromotedPatternsAreSeparatedFromCandidates(t *testing.T) {
	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "promotion.md"), []byte("Verified routing rules can be promoted into the runtime registry after distillation."), 0644); err != nil {
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
			Category:    "rules",
			Description: "rule docs",
		}},
	})
	if err != nil {
		t.Fatalf("Index() error = %v", err)
	}

	rag := NewRAGSystemManager(hub)
	_, err = rag.ProcessQueryWithContext(t.Context(), RAGQuery{
		Query: "When can a verified routing rule be promoted into runtime?",
	})
	if err != nil {
		t.Fatalf("ProcessQueryWithContext() error = %v", err)
	}
	hub.asyncWriter.flush()

	var candidateID string
	if err := hub.db.QueryRow("SELECT id FROM brain_pattern_candidates LIMIT 1").Scan(&candidateID); err != nil {
		t.Fatalf("load candidate id: %v", err)
	}

	api := NewAPIServer(hub, NewIntegrationManager(hub))

	promoteBody := []byte(`{"candidate_id":"` + candidateID + `"}`)
	promoteReq := httptest.NewRequest(http.MethodPost, "/api/v2/brain/patterns/promote", bytes.NewReader(promoteBody))
	promoteRec := httptest.NewRecorder()
	api.ServeHTTP(promoteRec, promoteReq)
	if promoteRec.Code != http.StatusOK {
		t.Fatalf("expected promote status 200, got %d: %s", promoteRec.Code, promoteRec.Body.String())
	}

	registryReq := httptest.NewRequest(http.MethodGet, "/api/v2/runtime/registry", nil)
	registryRec := httptest.NewRecorder()
	api.ServeHTTP(registryRec, registryReq)
	if registryRec.Code != http.StatusOK {
		t.Fatalf("expected registry status 200, got %d: %s", registryRec.Code, registryRec.Body.String())
	}

	var registry struct {
		Promoted []map[string]interface{} `json:"promoted"`
	}
	if err := json.Unmarshal(registryRec.Body.Bytes(), &registry); err != nil {
		t.Fatalf("decode registry response: %v", err)
	}
	if len(registry.Promoted) != 1 {
		t.Fatalf("expected one promoted pattern, got %#v", registry)
	}

	var remainingCandidates int
	if err := hub.db.QueryRow("SELECT COUNT(*) FROM brain_pattern_candidates WHERE status = 'candidate'").Scan(&remainingCandidates); err != nil {
		t.Fatalf("count remaining candidates: %v", err)
	}
	if remainingCandidates != 0 {
		t.Fatalf("expected no remaining candidate after promotion, got %d", remainingCandidates)
	}
}
