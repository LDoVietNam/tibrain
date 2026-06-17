package routeagent

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type RouteEvent struct {
	Timestamp     time.Time     `json:"timestamp"`
	Project       string        `json:"project,omitempty"`
	TaskType      string        `json:"task_type,omitempty"`
	ChosenRoute   string        `json:"chosen_route"`
	FallbackRoute string        `json:"fallback_route,omitempty"`
	Latency       time.Duration `json:"latency"`
	Success       bool          `json:"success"`
	ErrorClass    string        `json:"error_class,omitempty"`
	PromptHash    string        `json:"prompt_hash,omitempty"`
}

type CredentialState struct {
	Provider      string    `json:"provider"`
	Profile       string    `json:"profile,omitempty"`
	Domain        string    `json:"domain,omitempty"`
	Status        string    `json:"status"`
	HasSessionID  bool      `json:"has_session_id"`
	ExpiryGuess   int64     `json:"expiry_guess,omitempty"`
	Note          string    `json:"note,omitempty"`
	LastCheckedAt time.Time `json:"last_checked_at"`
}

type RouteStats struct {
	Route          string  `json:"route"`
	Success30m     float64 `json:"success_30m"`
	Success24h     float64 `json:"success_24h"`
	LatencyScore   float64 `json:"latency_score"`
	PolicyScore    float64 `json:"policy_score"`
	RecentAttempts int     `json:"recent_attempts"`
	DailyAttempts  int     `json:"daily_attempts"`
}

type MemoryStats struct {
	RouteEvents      int            `json:"route_events"`
	CredentialStates int            `json:"credential_states"`
	Policies         int            `json:"policies"`
	PromptPolicies   int            `json:"prompt_policies"`
	TopRoutes        map[string]int `json:"top_routes"`
	RecentEvents     []RouteEvent   `json:"recent_events,omitempty"`
}

type PolicyStats struct {
	SuccessCount int     `json:"success_count"`
	FailureCount int     `json:"failure_count"`
	AvgLatencyMS float64 `json:"avg_latency_ms"`
}

type PromptPolicyStats struct {
	SuccessCount    int     `json:"success_count"`
	FailureCount    int     `json:"failure_count"`
	AvgQuality      float64 `json:"avg_quality"`
	VerifySuccesses int     `json:"verify_successes"`
}

type Store struct{ db *sql.DB }

func NewStore(path string) (*Store, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, ".ti", "data", "route-agent.db")
	}
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.initDB(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) initDB() error {
	schema := `
CREATE TABLE IF NOT EXISTS credential_state (
    provider TEXT NOT NULL,
    profile TEXT NOT NULL DEFAULT '',
    domain TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '',
    has_session_id INTEGER NOT NULL DEFAULT 0,
    expiry_guess INTEGER NOT NULL DEFAULT 0,
    note TEXT NOT NULL DEFAULT '',
    last_checked_at INTEGER NOT NULL,
    PRIMARY KEY(provider, profile, domain)
);
CREATE TABLE IF NOT EXISTS route_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ts INTEGER NOT NULL,
    project TEXT NOT NULL DEFAULT '',
    task_type TEXT NOT NULL DEFAULT '',
    chosen_route TEXT NOT NULL,
    fallback_route TEXT NOT NULL DEFAULT '',
    latency_ms INTEGER NOT NULL DEFAULT 0,
    success INTEGER NOT NULL DEFAULT 0,
    error_class TEXT NOT NULL DEFAULT '',
    prompt_hash TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_route_events_ts ON route_events(ts DESC);
CREATE INDEX IF NOT EXISTS idx_route_events_route ON route_events(chosen_route, ts DESC);
CREATE INDEX IF NOT EXISTS idx_route_events_task ON route_events(task_type, ts DESC);
CREATE TABLE IF NOT EXISTS routing_policies (
    project TEXT NOT NULL DEFAULT '',
    task_type TEXT NOT NULL DEFAULT '',
    route TEXT NOT NULL,
    success_count INTEGER NOT NULL DEFAULT 0,
    failure_count INTEGER NOT NULL DEFAULT 0,
    avg_latency_ms REAL NOT NULL DEFAULT 0,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY(project, task_type, route)
);
CREATE TABLE IF NOT EXISTS prompt_policies (
    project TEXT NOT NULL DEFAULT '',
    task_type TEXT NOT NULL DEFAULT '',
    route TEXT NOT NULL,
    prompt_shape TEXT NOT NULL DEFAULT '',
    success_count INTEGER NOT NULL DEFAULT 0,
    failure_count INTEGER NOT NULL DEFAULT 0,
    avg_quality REAL NOT NULL DEFAULT 0,
    verify_successes INTEGER NOT NULL DEFAULT 0,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY(project, task_type, route, prompt_shape)
);`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) RecordCredentialState(state CredentialState) error {
	if state.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if state.LastCheckedAt.IsZero() {
		state.LastCheckedAt = time.Now()
	}
	_, err := s.db.Exec(`INSERT INTO credential_state (provider, profile, domain, status, has_session_id, expiry_guess, note, last_checked_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(provider, profile, domain) DO UPDATE SET
status=excluded.status,
has_session_id=excluded.has_session_id,
expiry_guess=excluded.expiry_guess,
note=excluded.note,
last_checked_at=excluded.last_checked_at`, state.Provider, state.Profile, state.Domain, state.Status, boolToInt(state.HasSessionID), state.ExpiryGuess, state.Note, state.LastCheckedAt.Unix())
	return err
}

func (s *Store) RecordRouteEvent(ev RouteEvent) error {
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now()
	}
	_, err := s.db.Exec(`INSERT INTO route_events (ts, project, task_type, chosen_route, fallback_route, latency_ms, success, error_class, prompt_hash)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, ev.Timestamp.Unix(), ev.Project, ev.TaskType, ev.ChosenRoute, ev.FallbackRoute, ev.Latency.Milliseconds(), boolToInt(ev.Success), ev.ErrorClass, ev.PromptHash)
	return err
}

func (s *Store) LearnPolicy(project, taskType, route string, success bool, latency time.Duration) error {
	if route == "" {
		return nil
	}
	now := time.Now().Unix()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var succ, fail int
	var avg float64
	err = tx.QueryRow(`SELECT success_count, failure_count, avg_latency_ms FROM routing_policies WHERE project=? AND task_type=? AND route=?`, project, taskType, route).Scan(&succ, &fail, &avg)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	count := succ + fail
	if success {
		succ++
	} else {
		fail++
	}
	latencyMS := float64(latency.Milliseconds())
	if count == 0 {
		avg = latencyMS
	} else if latencyMS > 0 {
		avg = ((avg * float64(count)) + latencyMS) / float64(count+1)
	}
	_, err = tx.Exec(`INSERT INTO routing_policies (project, task_type, route, success_count, failure_count, avg_latency_ms, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project, task_type, route) DO UPDATE SET
success_count=excluded.success_count,
failure_count=excluded.failure_count,
avg_latency_ms=excluded.avg_latency_ms,
updated_at=excluded.updated_at`, project, taskType, route, succ, fail, avg, now)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) LearnFromQA(project, taskType, route, promptShape string, quality float64, preferred bool, verifyPassed bool) error {
	if route == "" || route == "local-verify" {
		return nil
	}
	now := time.Now().Unix()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var succ, fail int
	var avgLatency float64
	err = tx.QueryRow(`SELECT success_count, failure_count, avg_latency_ms FROM routing_policies WHERE project=? AND task_type=? AND route=?`, project, taskType, route).Scan(&succ, &fail, &avgLatency)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if preferred {
		succ++
	} else {
		fail++
	}
	_, err = tx.Exec(`INSERT INTO routing_policies (project, task_type, route, success_count, failure_count, avg_latency_ms, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project, task_type, route) DO UPDATE SET
success_count=excluded.success_count,
failure_count=excluded.failure_count,
avg_latency_ms=excluded.avg_latency_ms,
updated_at=excluded.updated_at`, project, taskType, route, succ, fail, avgLatency, now)
	if err != nil {
		return err
	}

	if promptShape != "" {
		var psucc, pfail, verifySucc int
		var avgQuality float64
		err = tx.QueryRow(`SELECT success_count, failure_count, avg_quality, verify_successes FROM prompt_policies WHERE project=? AND task_type=? AND route=? AND prompt_shape=?`, project, taskType, route, promptShape).Scan(&psucc, &pfail, &avgQuality, &verifySucc)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		count := psucc + pfail
		if preferred {
			psucc++
		} else {
			pfail++
		}
		if verifyPassed {
			verifySucc++
		}
		if count == 0 {
			avgQuality = quality
		} else {
			avgQuality = ((avgQuality * float64(count)) + quality) / float64(count+1)
		}
		_, err = tx.Exec(`INSERT INTO prompt_policies (project, task_type, route, prompt_shape, success_count, failure_count, avg_quality, verify_successes, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project, task_type, route, prompt_shape) DO UPDATE SET
success_count=excluded.success_count,
failure_count=excluded.failure_count,
avg_quality=excluded.avg_quality,
verify_successes=excluded.verify_successes,
updated_at=excluded.updated_at`, project, taskType, route, promptShape, psucc, pfail, avgQuality, verifySucc, now)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) RouteStats(project, taskType, route string) (RouteStats, error) {
	stats := RouteStats{Route: route, Success30m: 0.5, Success24h: 0.5, LatencyScore: 0.5, PolicyScore: 0.5}
	if route == "" {
		return stats, nil
	}
	since30m := time.Now().Add(-30 * time.Minute).Unix()
	since24h := time.Now().Add(-24 * time.Hour).Unix()
	stats.Success30m, stats.RecentAttempts, _ = s.windowSuccess(route, taskType, since30m)
	stats.Success24h, stats.DailyAttempts, stats.LatencyScore = s.windowSuccessWithLatency(route, taskType, since24h)
	policy, err := s.Policy(project, taskType, route)
	if err != nil {
		return stats, err
	}
	if policy.SuccessCount+policy.FailureCount > 0 {
		stats.PolicyScore = float64(policy.SuccessCount+1) / float64(policy.SuccessCount+policy.FailureCount+2)
		if policy.AvgLatencyMS > 0 {
			policyLatency := latencyToScore(policy.AvgLatencyMS)
			stats.LatencyScore = (stats.LatencyScore + policyLatency) / 2
		}
	}
	return stats, nil
}

func (s *Store) windowSuccess(route, taskType string, since int64) (float64, int, error) {
	var total, succ int
	err := s.db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(success),0) FROM route_events WHERE chosen_route=? AND task_type=? AND ts>=?`, route, taskType, since).Scan(&total, &succ)
	if err != nil {
		return 0.5, 0, err
	}
	if total == 0 {
		return 0.5, 0, nil
	}
	return float64(succ) / float64(total), total, nil
}

func (s *Store) windowSuccessWithLatency(route, taskType string, since int64) (float64, int, float64) {
	var total, succ int
	var avgLatency sql.NullFloat64
	err := s.db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(success),0), AVG(latency_ms) FROM route_events WHERE chosen_route=? AND task_type=? AND ts>=?`, route, taskType, since).Scan(&total, &succ, &avgLatency)
	if err != nil || total == 0 {
		return 0.5, total, 0.5
	}
	return float64(succ) / float64(total), total, latencyToScore(avgLatency.Float64)
}

func (s *Store) Policy(project, taskType, route string) (PolicyStats, error) {
	var p PolicyStats
	err := s.db.QueryRow(`SELECT success_count, failure_count, avg_latency_ms FROM routing_policies WHERE project=? AND task_type=? AND route=?`, project, taskType, route).Scan(&p.SuccessCount, &p.FailureCount, &p.AvgLatencyMS)
	if err == sql.ErrNoRows {
		return PolicyStats{}, nil
	}
	return p, err
}

func (s *Store) PromptPolicy(project, taskType, route, promptShape string) (PromptPolicyStats, error) {
	var p PromptPolicyStats
	err := s.db.QueryRow(`SELECT success_count, failure_count, avg_quality, verify_successes FROM prompt_policies WHERE project=? AND task_type=? AND route=? AND prompt_shape=?`, project, taskType, route, promptShape).Scan(&p.SuccessCount, &p.FailureCount, &p.AvgQuality, &p.VerifySuccesses)
	if err == sql.ErrNoRows {
		return PromptPolicyStats{}, nil
	}
	return p, err
}

func (s *Store) Stats() (*MemoryStats, error) {
	out := &MemoryStats{TopRoutes: map[string]int{}}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM route_events`).Scan(&out.RouteEvents); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM credential_state`).Scan(&out.CredentialStates); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM routing_policies`).Scan(&out.Policies); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM prompt_policies`).Scan(&out.PromptPolicies); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(`SELECT chosen_route, COUNT(*) FROM route_events GROUP BY chosen_route ORDER BY COUNT(*) DESC LIMIT 5`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var route string
		var count int
		if err := rows.Scan(&route, &count); err != nil {
			return nil, err
		}
		out.TopRoutes[route] = count
	}
	events, err := s.RecentEvents(5)
	if err != nil {
		return nil, err
	}
	out.RecentEvents = events
	return out, nil
}

func (s *Store) RecentEvents(limit int) ([]RouteEvent, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.Query(`SELECT ts, project, task_type, chosen_route, fallback_route, latency_ms, success, error_class, prompt_hash FROM route_events ORDER BY ts DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []RouteEvent{}
	for rows.Next() {
		var ts int64
		var latencyMS int64
		var success int
		var ev RouteEvent
		if err := rows.Scan(&ts, &ev.Project, &ev.TaskType, &ev.ChosenRoute, &ev.FallbackRoute, &latencyMS, &success, &ev.ErrorClass, &ev.PromptHash); err != nil {
			return nil, err
		}
		ev.Timestamp = time.Unix(ts, 0)
		ev.Latency = time.Duration(latencyMS) * time.Millisecond
		ev.Success = success == 1
		out = append(out, ev)
	}
	return out, nil
}

func PromptHash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func latencyToScore(ms float64) float64 {
	switch {
	case ms <= 0:
		return 0.5
	case ms <= 1500:
		return 1.0
	case ms <= 3500:
		return 0.8
	case ms <= 6000:
		return 0.65
	case ms <= 10000:
		return 0.45
	default:
		return 0.25
	}
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
