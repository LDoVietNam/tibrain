package learn

import (
	"strings"
	"sync"
	"time"

	"github.com/ti/cli/internal/qa"
)

// ContextFinder retrieves relevant past sessions to inject as context.
type ContextFinder struct {
	mu       sync.RWMutex
	qaStore  *qa.Store
	sessions map[string]*Session // sessionID → Session
	maxCache int
}

// Session represents a conversation session for context.
type Session struct {
	ID        string    `json:"id"`
	TaskType  string    `json:"task_type"`
	Domain    string    `json:"domain"`
	Prompt    string    `json:"prompt"`
	Summary   string    `json:"summary"` // Generated summary (first 200 chars)
	Model     string    `json:"model"`
	Quality   float64   `json:"quality"`
	CreatedAt time.Time `json:"created_at"`
}

// NewContextFinder creates a new finder.
func NewContextFinder(qaStore *qa.Store) *ContextFinder {
	return &ContextFinder{
		qaStore:  qaStore,
		sessions: make(map[string]*Session),
		maxCache: 1000,
	}
}

// IndexSession adds a session to the context index.
func (cf *ContextFinder) IndexSession(sessionID, taskType, domain, prompt, model string, quality float64) {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	if len(cf.sessions) >= cf.maxCache {
		// Evict oldest
		var oldest string
		var oldestTime time.Time
		for id, s := range cf.sessions {
			if oldest == "" || s.CreatedAt.Before(oldestTime) {
				oldest = id
				oldestTime = s.CreatedAt
			}
		}
		if oldest != "" {
			delete(cf.sessions, oldest)
		}
	}

	summary := prompt
	if len(summary) > 200 {
		summary = summary[:200] + "..."
	}

	cf.sessions[sessionID] = &Session{
		ID:        sessionID,
		TaskType:  taskType,
		Domain:    domain,
		Prompt:    prompt,
		Summary:   summary,
		Model:     model,
		Quality:   quality,
		CreatedAt: time.Now(),
	}
}

// FindSimilar returns sessions similar to the given prompt.
// Uses keyword matching (simple) - can upgrade to embeddings later.
func (cf *ContextFinder) FindSimilar(prompt string, taskType string, limit int) []ContextItem {
	cf.mu.RLock()
	defer cf.mu.RUnlock()

	// Extract keywords from query
	queryKeywords := extractKeywords(prompt)

	var scored []ContextItem
	for _, session := range cf.sessions {
		// Skip if task type mismatch (除非通用)
		if session.TaskType != taskType && taskType != "general" {
			continue
		}

		// Score by keyword overlap + quality bonus
		sessionKeywords := extractKeywords(session.Prompt)
		overlap := keywordOverlap(queryKeywords, sessionKeywords)
		score := overlap

		// Boost by quality
		score += session.Quality / 10.0

		// Boost by recency (decay: newer = higher)
		ageHours := time.Since(session.CreatedAt).Hours()
		recencyBonus := 1.0 / (1.0 + ageHours/24.0) // Decay over days
		score += recencyBonus * 0.5

		if score > 0.3 {
			scored = append(scored, ContextItem{
				SessionID: session.ID,
				TaskType:  session.TaskType,
				Model:     session.Model,
				Content:   session.Summary,
				Score:     score,
			})
		}
	}

	// Sort by score desc
	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].Score > scored[i].Score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Return top N
	if len(scored) > limit {
		scored = scored[:limit]
	}

	return scored
}

// ContextItem is a retrieved context snippet.
type ContextItem struct {
	SessionID string  `json:"session_id"`
	TaskType  string  `json:"task_type"`
	Model     string  `json:"model"`
	Content   string  `json:"content"` // Truncated prompt/result
	Score     float64 `json:"score"`   // Relevance score 0-1
}

// QA context: Use existing QA store for exemplars.
func (cf *ContextFinder) FindQAExamples(taskType string, limit int) []qa.Entry {
	if cf.qaStore == nil {
		return nil
	}

	entries, err := cf.qaStore.FilterByTaskType(taskType)
	if err != nil {
		return nil
	}

	// Return top by quality, limited
	if len(entries) > limit {
		entries = entries[:limit]
	}
	return entries
}

// extractKeywords is a simple keyword extractor.
func extractKeywords(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	stop := map[string]bool{"the": true, "a": true, "an": true, "for": true, "to": true, "and": true}
	keywords := make([]string, 0, len(words))
	for _, w := range words {
		w = strings.Trim(w, ".,;:!?()[]{}\"'")
		if len(w) > 2 && !stop[w] {
			keywords = append(keywords, w)
		}
	}
	return keywords
}

// keywordOverlap computes Jaccard similarity.
func keywordOverlap(a, b []string) float64 {
	setA := make(map[string]bool)
	for _, w := range a {
		setA[w] = true
	}
	overlap := 0
	for _, w := range b {
		if setA[w] {
			overlap++
		}
	}
	union := len(a) + len(b) - overlap
	if union == 0 {
		return 0.0
	}
	return float64(overlap) / float64(union)
}
