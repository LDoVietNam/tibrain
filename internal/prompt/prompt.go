// Package prompt implements TiBrain Prompt Intelligence: a registry of prompt
// capsules, a preflight selector (used by TiRouter to pick the best PromptEnvelope),
// an outcome-feedback logger, and a catalog-version endpoint.
//
// Schema: see internal/db/migrations/0004_prompt_intelligence.sql
package prompt

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// PromptIntelligence wraps the prompt registry DB access.
type PromptIntelligence struct {
	db *sql.DB
}

// New constructs a PromptIntelligence from the shared Hub DB.
func New(db *sql.DB) *PromptIntelligence {
	return &PromptIntelligence{db: db}
}

// Capsule is the metadata row of a prompt capsule.
type Capsule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Intent      string `json:"intent"`
	Domain      string `json:"domain"`
	Risk        string `json:"risk"`
	Status      string `json:"status"`
}

// Envelope is what TiRouter consumes: a concrete prompt version + placement.
type Envelope struct {
	CapsuleID  string `json:"capsule_id"`
	Version    string `json:"version"`
	Intent     string `json:"intent"`
	Domain     string `json:"domain"`
	Risk       string `json:"risk"`
	Placement  string `json:"placement"`
	Content    string `json:"content"`
	TokenEstimate int `json:"token_estimate"`
}

// PreflightRequest is the TiRouter preflight input.
type PreflightRequest struct {
	Intent     string `json:"intent"`
	Domain     string `json:"domain"`
	MinRisk    string `json:"min_risk,omitempty"`
}

func newID(prefix string) string {
	return fmt.Sprintf("%s_%d_%06d", prefix, time.Now().UnixNano(), rand.Intn(1000000))
}

// SeedDefaults inserts a small built-in catalog if the registry is empty.
// Safe to call on every startup (idempotent).
func (p *PromptIntelligence) SeedDefaults() error {
	var n int
	if err := p.db.QueryRow(`SELECT COUNT(*) FROM prompt_capsules`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	now := time.Now().Unix()
	seeds := []struct {
		id, name, intent, domain, risk, status, content, placement string
	}{
		{"cap_chat_general", "General Chat", "chat.general", "general", "low", "active",
			"You are TiBrain, a concise internal intelligence assistant. Answer directly.", "system"},
		{"cap_code_review", "Code Review", "code.review", "engineering", "medium", "active",
			"Review the following diff for bugs, security and style. Be specific.", "system"},
		{"cap_agent_orchestrate", "Agent Orchestration", "agent.orchestrate", "engineering", "high", "active",
			"Plan and decompose the task into verifiable steps before acting.", "system"},
	}
	for _, s := range seeds {
		if _, err := p.db.Exec(
			`INSERT OR IGNORE INTO prompt_capsules (id,name,description,intent,domain,risk,status,created_at,updated_at)
			 VALUES (?,?,?,?,?,?,?,?,?)`,
			s.id, s.name, s.name, s.intent, s.domain, s.risk, s.status, now, now); err != nil {
			return err
		}
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(s.content)))
		if _, err := p.db.Exec(
			`INSERT OR IGNORE INTO prompt_capsule_versions (id,version,content,content_hash,token_estimate,placement,created_at)
			 VALUES (?,?,?,?,?,?,?)`,
			s.id, "1.0.0", s.content, hash, len(s.content)/4, s.placement, now); err != nil {
			return err
		}
	}
	return nil
}

// Preflight selects the best active capsule for an intent/domain.
func (p *PromptIntelligence) Preflight(req PreflightRequest) (*Envelope, error) {
	if err := p.SeedDefaults(); err != nil {
		return nil, err
	}
	var id, intent, domain, risk, status string
	q := `SELECT id,intent,domain,risk,status FROM prompt_capsules
	      WHERE intent=? AND status='active' ORDER BY updated_at DESC LIMIT 1`
	err := p.db.QueryRow(q, req.Intent).Scan(&id, &intent, &domain, &risk, &status)
	if err == sql.ErrNoRows && req.Domain != "" {
		q = `SELECT id,intent,domain,risk,status FROM prompt_capsules
		     WHERE domain=? AND status='active' ORDER BY updated_at DESC LIMIT 1`
		err = p.db.QueryRow(q, req.Domain).Scan(&id, &intent, &domain, &risk, &status)
	}
	if err == sql.ErrNoRows {
		// fallback: any active capsule
		q = `SELECT id,intent,domain,risk,status FROM prompt_capsules
		     WHERE status='active' ORDER BY updated_at DESC LIMIT 1`
		err = p.db.QueryRow(q).Scan(&id, &intent, &domain, &risk, &status)
	}
	if err != nil {
		return nil, err
	}
	var version, placement, content string
	var tok int
	if err := p.db.QueryRow(
		`SELECT version,placement,content,token_estimate FROM prompt_capsule_versions
		 WHERE id=? ORDER BY created_at DESC LIMIT 1`, id,
	).Scan(&version, &placement, &content, &tok); err != nil {
		return nil, err
	}
	return &Envelope{
		CapsuleID: id, Version: version, Intent: intent, Domain: domain,
		Risk: risk, Placement: placement, Content: content, TokenEstimate: tok,
	}, nil
}

// FeedbackRequest is the outcome log from TiRouter.
type FeedbackRequest struct {
	RequestID      string `json:"request_id"`
	CapsuleID      string `json:"capsule_id"`
	CapsuleVersion string `json:"capsule_version"`
	Outcome        string `json:"outcome"`
	UserOverride   int    `json:"user_override,omitempty"`
	AddedTokens    int    `json:"added_tokens,omitempty"`
	ProviderErrorCode string `json:"provider_error_code,omitempty"`
}

// RecordFeedback persists an outcome + a route trace.
func (p *PromptIntelligence) RecordFeedback(req FeedbackRequest) error {
	now := time.Now().Unix()
	fid := newID("fb")
	if _, err := p.db.Exec(
		`INSERT INTO prompt_feedback (id,request_id,capsule_id,capsule_version,outcome,user_override,added_tokens,provider_error_code,created_at)
		 VALUES (?,?,?,?,?,?,?,?,?)`,
		fid, req.RequestID, req.CapsuleID, req.CapsuleVersion, req.Outcome,
		req.UserOverride, req.AddedTokens, req.ProviderErrorCode, now); err != nil {
		return err
	}
	tid := newID("tr")
	if _, err := p.db.Exec(
		`INSERT INTO prompt_route_traces (id,request_id,capsule_id,capsule_version,decision,confidence,reason_code,latency_ms,created_at)
		 VALUES (?,?,?,?,?,?,?,?,?)`,
		tid, req.RequestID, req.CapsuleID, req.CapsuleVersion, "feedback", 1.0, req.Outcome, 0, now); err != nil {
		return err
	}
	return nil
}

// CatalogVersion returns an ETag-like version of the current catalog.
func (p *PromptIntelligence) CatalogVersion() (string, error) {
	var maxUpdated sql.NullInt64
	if err := p.db.QueryRow(`SELECT MAX(updated_at) FROM prompt_capsules`).Scan(&maxUpdated); err != nil {
		return "", err
	}
	var count int
	if err := p.db.QueryRow(`SELECT COUNT(*) FROM prompt_capsules`).Scan(&count); err != nil {
		return "", err
	}
	raw := fmt.Sprintf("%d:%d", count, maxUpdated.Int64)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(raw))), nil
}

// RegisterRoutes wires the three Prompt Intelligence endpoints onto mux.
// Routes:
//   POST /api/v1/prompt/preflight
//   POST /api/v1/prompt/feedback
//   GET  /api/v1/prompt/catalog/version
func RegisterRoutes(mux *http.ServeMux, pi *PromptIntelligence) {
	mux.HandleFunc("/api/v1/prompt/preflight", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req PreflightRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		env, err := pi.Preflight(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if env == nil {
			http.Error(w, "no matching capsule", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, env)
	})

	mux.HandleFunc("/api/v1/prompt/feedback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req FeedbackRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.CapsuleID == "" || req.Outcome == "" {
			http.Error(w, "capsule_id and outcome required", http.StatusBadRequest)
			return
		}
		if err := pi.RecordFeedback(req); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "recorded"})
	})

	mux.HandleFunc("/api/v1/prompt/catalog/version", func(w http.ResponseWriter, r *http.Request) {
		v, err := pi.CatalogVersion()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"version": v})
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
