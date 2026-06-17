package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ti/router/tibrain/providers"
)

// Role types for the Failover Model Pool
const (
	RolePlanner  = "planner"
	RoleWorker   = "worker"
	RoleReviewer = "reviewer"
)

// Model failure cooldown duration
const failureCooldown = 5 * time.Minute

// FailoverModelPool manages client-side model failovers on rate-limits/quotas.
type FailoverModelPool struct {
	mu           sync.Mutex
	failedModels map[string]time.Time // Tracks model -> failure time
	primaryModel string               // User-selected preferred model
}

// NewFailoverModelPool creates a new FailoverModelPool instance.
func NewFailoverModelPool() *FailoverModelPool {
	return &FailoverModelPool{
		failedModels: make(map[string]time.Time),
	}
}

// SetPrimaryModel sets the user-selected preferred model to be tried first.
func (p *FailoverModelPool) SetPrimaryModel(model string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.primaryModel = model
}

// GetPoolForRole returns the prioritized list of models for a given role.
func (p *FailoverModelPool) GetPoolForRole(role string) []string {
	p.mu.Lock()
	primary := p.primaryModel
	p.mu.Unlock()

	var basePool []string
	switch role {
	case RolePlanner:
		basePool = []string{
			"claude-3.5-sonnet",
			"gpt-4o",
			"o1-preview",
			"gemini-3.5-flash",
			"gemini-1.5-pro",
		}
	case RoleWorker:
		basePool = []string{
			"claude-3.5-sonnet",
			"gpt-4o",
			"gemini-3.5-flash",
			"gemini-1.5-pro",
			"o1-mini",
		}
	case RoleReviewer:
		basePool = []string{
			"gemini-3.5-flash",
			"gemini-1.5-pro",
			"claude-3.5-sonnet",
			"gpt-4o",
		}
	default:
		basePool = []string{"claude-3.5-sonnet", "gpt-4o"}
	}

	if primary != "" {
		res := []string{primary}
		for _, m := range basePool {
			if m != primary {
				res = append(res, m)
			}
		}
		return res
	}
	return basePool
}

// MarkFailed registers that a model has failed recently.
func (p *FailoverModelPool) MarkFailed(model string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.failedModels[model] = time.Now()
}

// IsFailed checks if a model is currently in the failure cooldown period.
func (p *FailoverModelPool) IsFailed(model string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	failTime, exists := p.failedModels[model]
	if !exists {
		return false
	}
	if time.Since(failTime) > failureCooldown {
		delete(p.failedModels, model)
		return false
	}
	return true
}

// IsQuotaError checks if the error message indicates a rate limit or quota issue.
func IsQuotaError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "quota") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "429") ||
		strings.Contains(msg, "402") ||
		strings.Contains(msg, "payment") ||
		strings.Contains(msg, "limit exceeded") ||
		strings.Contains(msg, "too many requests")
}

// ChatWithFallback executes a chat request on the best available model for a role.
// If the selected model fails with a quota or rate-limit issue, it automatically
// falls back to the next model in the pool.
// Returns: (response, modelUsed, error)
func (p *FailoverModelPool) ChatWithFallback(ctx context.Context, cb *providers.CodebuffProvider, role string, messages []providers.Message, onFallback func(failedModel, nextModel string, reason error)) (string, string, error) {
	pool := p.GetPoolForRole(role)
	var lastErr error

	for _, model := range pool {
		// Skip if it failed recently
		if p.IsFailed(model) {
			continue
		}

		// Prepare request
		req := providers.ChatRequest{
			Model:    model,
			Messages: messages,
		}

		// Try call
		resp, err := cb.Chat(ctx, req)
		if err == nil && resp != nil {
			return resp.Content, model, nil
		}

		// Check if it's a quota/rate-limit error
		if IsQuotaError(err) {
			p.MarkFailed(model)
			if onFallback != nil {
				// Find the next non-failed model for logging
				nextModel := "none"
				for _, m := range pool {
					if m != model && !p.IsFailed(m) {
						nextModel = m
						break
					}
				}
				onFallback(model, nextModel, err)
			}
			lastErr = err
			continue // Try the next model
		}

		// For other severe errors (e.g. authentication), we might want to stop or retry
		lastErr = err
		// We still try fallback for robustness unless context is cancelled
		if ctx.Err() != nil {
			return "", "", ctx.Err()
		}
	}

	return "", "", fmt.Errorf("all models in the pool for role '%s' failed. Last error: %v", role, lastErr)
}
