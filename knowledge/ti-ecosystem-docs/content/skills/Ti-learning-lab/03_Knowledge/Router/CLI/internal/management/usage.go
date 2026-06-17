package management

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CallLog represents a single API call record for usage tracking.
// TypeScript reference: nextjs/lib/usageDb.ts (saveCallLog)
type CallLog struct {
	Timestamp    string                 `json:"timestamp"`
	Method       string                 `json:"method"`
	Path         string                 `json:"path"`
	Status       int                    `json:"status"`
	Model        string                 `json:"model,omitempty"`
	Provider     string                 `json:"provider,omitempty"`
	DurationMs   int64                  `json:"duration_ms,omitempty"`
	Tokens       map[string]interface{} `json:"tokens,omitempty"`
	RequestBody  map[string]interface{} `json:"request_body,omitempty"`
	ResponseBody map[string]interface{} `json:"response_body,omitempty"`
	Error        string                 `json:"error,omitempty"`
	APIKeyID     string                 `json:"api_key_id,omitempty"`
	APIKeyName   string                 `json:"api_key_name,omitempty"`
	ConnectionID string                 `json:"connection_id,omitempty"`
	SourceFormat string                 `json:"source_format,omitempty"`
	TargetFormat string                 `json:"target_format,omitempty"`
}

// UsageStore persists call logs to disk.
type UsageStore struct {
	mu      sync.Mutex
	dataDir string
	buffer  []CallLog
}

// NewUsageStore creates a usage store rooted at dataDir.
func NewUsageStore(dataDir string) *UsageStore {
	return &UsageStore{
		dataDir: dataDir,
		buffer:  make([]CallLog, 0),
	}
}

// Record appends a call log to the in-memory buffer and flushes asynchronously.
func (s *UsageStore) Record(log CallLog) {
	s.mu.Lock()
	if log.Timestamp == "" {
		log.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	s.buffer = append(s.buffer, log)
	// Keep last 500 entries in memory
	if len(s.buffer) > 500 {
		s.buffer = s.buffer[len(s.buffer)-500:]
	}
	s.mu.Unlock()

	// Async flush to disk
	go s.flush(log)
}

func (s *UsageStore) flush(log CallLog) {
	if s.dataDir == "" {
		return
	}
	_ = os.MkdirAll(s.dataDir, 0o700)
	path := filepath.Join(s.dataDir, "call_logs.ndjson")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()

	b, _ := json.Marshal(log)
	_, _ = f.Write(b)
	_, _ = f.WriteString("\n")
}

// Recent returns the last n call logs from memory.
func (s *UsageStore) Recent(n int) []CallLog {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n >= len(s.buffer) {
		out := make([]CallLog, len(s.buffer))
		copy(out, s.buffer)
		return out
	}
	return s.buffer[len(s.buffer)-n:]
}

// Stats returns aggregated usage statistics.
func (s *UsageStore) Stats() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	totalCalls := len(s.buffer)
	tokensPrompt := 0
	tokensCompletion := 0
	byProvider := map[string]int{}
	byModel := map[string]int{}

	for _, log := range s.buffer {
		byProvider[log.Provider]++
		byModel[log.Model]++
		if t, ok := log.Tokens["prompt_tokens"].(float64); ok {
			tokensPrompt += int(t)
		}
		if t, ok := log.Tokens["completion_tokens"].(float64); ok {
			tokensCompletion += int(t)
		}
	}

	return map[string]interface{}{
		"total_calls":       totalCalls,
		"prompt_tokens":     tokensPrompt,
		"completion_tokens": tokensCompletion,
		"by_provider":       byProvider,
		"by_model":          byModel,
	}
}
