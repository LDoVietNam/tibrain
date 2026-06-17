package management

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// RateLimiter is a simple in-memory rate limiter keyed by client IP or API key.
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a rate limiter with the given request limit per window.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if the given key is within the rate limit.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Evict old entries
	times := rl.requests[key]
	idx := 0
	for i, t := range times {
		if t.After(cutoff) {
			idx = i
			break
		}
	}
	if idx > 0 {
		times = times[idx:]
	}

	if len(times) >= rl.limit {
		rl.requests[key] = times
		return false
	}

	rl.requests[key] = append(times, now)
	return true
}

// Middleware wraps an http.Handler with rate limiting keyed by Authorization header or RemoteAddr.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Authorization")
		if key == "" {
			key = r.RemoteAddr
		}
		if !rl.Allow(key) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Rate limit exceeded. Please try again later.",
					"type":    "rate_limit_error",
					"code":    "rate_limit_exceeded",
				},
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
