// Package providers implements AI provider backends for Ti.
// SharedChatProvider handles cookie-based auth with dual-domain failover,
// token caching, retry with exponential backoff, and health tracking.
package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// SharedChatProvider implements the Provider interface for SharedChat platform.
// Features:
//   - Dual-domain failover: primary (.cc) -> fallback (.fun)
//   - Token exchange: cookie -> /api/auth/session -> Bearer token with cache + expiry
//   - SSE streaming with conversation context tracking
//   - Cookie auto-refresh: token expired -> re-fetch from /api/auth/session
//   - Retry with exponential backoff (max 3 retries, 2s base, 30s cap)
//   - Health check: sessionID presence + last successful request time
//   - Thread-safe: sync.Mutex for session state
//   - Conversation continuity: track conversationId across turns
//   - Context cancellation support
type SharedChatProvider struct {
	// Domains
	apiPrimary   string // "https://chat.sharedchat.cc"
	apiFallback  string // "https://chat.sharedchat.fun"
	cookieDomain string // "chat.sharedchat.fun"

	// Session state (thread-safe)
	mu             sync.Mutex
	sessionID      string
	accessToken    string
	tokenExpireAt  time.Time
	conversationID string
	healthy        bool
	lastSuccessAt  time.Time

	// HTTP client with cookie jar
	httpClient *http.Client
	cookies    []*http.Cookie

	// Identity
	models       []string
	defaultModel string
	userID       string

	// Retry config
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration

	// Request timeout
	requestTimeout time.Duration

	// Cost optimization
	enableCaching   bool            // Enable prompt caching
	reasoningEffort ReasoningEffort // Default reasoning effort

	// Atomic counter for consecutive failures (thread-safe without mutex)
	consecutiveFailures atomic.Int64

	// Simple LRU cache for prompt caching (in-memory)
	promptCache     map[string]*cacheEntry
	promptCacheMu   sync.RWMutex
	promptCacheSize int
}

type cacheEntry struct {
	response  *ChatResponse
	createdAt time.Time
	ttl       time.Duration
}

// SharedChatConfig holds SharedChat provider configuration from JSON.
type SharedChatConfig struct {
	SessionID    string         `json:"session_id"`
	APIPrimary   string         `json:"api_primary,omitempty"`
	APIFallback  string         `json:"api_fallback,omitempty"`
	CookieDomain string         `json:"cookie_domain,omitempty"`
	Models       []string       `json:"models,omitempty"`
	DefaultModel string         `json:"default_model,omitempty"`
	Cookies      []CookieConfig `json:"cookies,omitempty"`
}

// CookieConfig represents a single cookie (reused from cookie_provider.go pattern).
type CookieConfig struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Secure   bool   `json:"secure,omitempty"`
	HTTPOnly bool   `json:"httpOnly,omitempty"`
	SameSite string `json:"sameSite,omitempty"`
}

// NewSharedChatProvider creates a new provider with dual-domain support.
func NewSharedChatProvider(cfg *SharedChatConfig) (*SharedChatProvider, error) {
	if cfg.SessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	primary := cfg.APIPrimary
	if primary == "" {
		primary = "https://chat.sharedchat.cc"
	}
	fallback := cfg.APIFallback
	if fallback == "" {
		fallback = "https://chat.sharedchat.fun"
	}
	cookieDomain := cfg.CookieDomain
	if cookieDomain == "" {
		cookieDomain = "chat.sharedchat.fun"
	}

	models := cfg.Models
	if len(models) == 0 {
		models = []string{"gpt-4o", "claude-3-sonnet", "gemini-pro"}
	}
	defaultModel := cfg.DefaultModel
	if defaultModel == "" {
		defaultModel = "gpt-4o"
	}

	// Build cookie jar if cookies provided
	var jar http.CookieJar
	var cookies []*http.Cookie
	if len(cfg.Cookies) > 0 {
		var err error
		jar, err = cookiejar.New(nil)
		if err != nil {
			return nil, fmt.Errorf("create cookie jar: %w", err)
		}

		for _, c := range cfg.Cookies {
			cookie := &http.Cookie{
				Name:     c.Name,
				Value:    c.Value,
				Domain:   c.Domain,
				Path:     c.Path,
				Secure:   c.Secure,
				HttpOnly: c.HTTPOnly,
			}

			switch strings.ToLower(c.SameSite) {
			case "strict":
				cookie.SameSite = http.SameSiteStrictMode
			case "none":
				cookie.SameSite = http.SameSiteNoneMode
			default:
				cookie.SameSite = http.SameSiteLaxMode
			}

			cookies = append(cookies, cookie)
		}

		// Set cookies on jar
		u, err := url.Parse("https://" + cookieDomain)
		if err == nil {
			jar.SetCookies(u, cookies)
		}
	}

	// Seed rand once per provider
	userID := fmt.Sprintf("%d", rand.Int63())

	return &SharedChatProvider{
		apiPrimary:   primary,
		apiFallback:  fallback,
		cookieDomain: cookieDomain,
		sessionID:    cfg.SessionID,
		models:       models,
		defaultModel: defaultModel,
		userID:       userID,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
			Jar:     jar,
		},
		cookies:         cookies,
		healthy:         true,
		maxRetries:      3,
		baseDelay:       2 * time.Second,
		maxDelay:        30 * time.Second,
		requestTimeout:  120 * time.Second,
		reasoningEffort: ReasoningMedium,
		promptCache:     make(map[string]*cacheEntry),
		promptCacheSize: 100,
	}, nil
}

// Name implements Provider interface.
func (p *SharedChatProvider) Name() string {
	return "sharedchat"
}

// DefaultModel implements Provider interface.
func (p *SharedChatProvider) DefaultModel() string {
	return p.defaultModel
}

// Models implements Provider interface.
func (p *SharedChatProvider) Models() []string {
	return p.models
}

// IsHealthy implements Provider interface.
func (p *SharedChatProvider) IsHealthy() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.sessionID != "" && p.healthy
}

// HealthStatus returns detailed health information.
func (p *SharedChatProvider) HealthStatus() map[string]any {
	p.mu.Lock()
	defer p.mu.Unlock()

	return map[string]any{
		"session_id_present": p.sessionID != "",
		"healthy":            p.healthy,
		"last_success_at":    p.lastSuccessAt,
		"time_since_success": time.Since(p.lastSuccessAt).String(),
		"consecutive_errors": p.consecutiveFailures.Load(),
		"api_primary":        p.apiPrimary,
		"api_fallback":       p.apiFallback,
		"has_token":          p.accessToken != "",
		"token_expires_at":   p.tokenExpireAt,
	}
}

// Chat sends a message and waits for the full response.
// It auto-refreshes the access token and handles dual-domain failover with retry.
func (p *SharedChatProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	var result strings.Builder
	cb := func(chunk StreamChunk) {
		if chunk.Content != "" {
			result.WriteString(chunk.Content)
		}
	}
	return p.ChatStream(ctx, req, cb)
}

// ChatStream sends a message and streams chunks via callback.
// Supports dual-domain failover and retry with exponential backoff.
func (p *SharedChatProvider) ChatStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("no messages in request")
	}

	// Retry loop with exponential backoff
	var lastErr error
	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			// Check context before retry
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("request cancelled after %d retries: %w", attempt-1, ctx.Err())
			default:
			}

			// Exponential backoff with jitter
			backoff := p.calculateBackoff(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				// Continue to retry
			}
		}

		resp, err := p.executeStreamWithFailover(ctx, req, onChunk)
		if err == nil {
			// Success - reset failure counter and record timestamp
			p.consecutiveFailures.Store(0)
			p.mu.Lock()
			p.healthy = true
			p.lastSuccessAt = time.Now()
			p.mu.Unlock()
			return resp, nil
		}

		lastErr = err
		p.consecutiveFailures.Add(1)

		// If context cancelled, don't retry
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}

	// All retries exhausted
	p.mu.Lock()
	p.healthy = false
	p.mu.Unlock()

	return nil, fmt.Errorf("sharedchat request failed after %d retries: %w", p.maxRetries, lastErr)
}

// executeStreamWithFailover tries primary domain, falls back to secondary.
func (p *SharedChatProvider) executeStreamWithFailover(ctx context.Context, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error) {
	// Try primary domain
	resp, err := p.streamFromDomain(ctx, p.apiPrimary, req, onChunk)
	if err == nil {
		return resp, nil
	}

	// Primary failed -> try fallback
	resp, err2 := p.streamFromDomain(ctx, p.apiFallback, req, onChunk)
	if err2 != nil {
		return nil, fmt.Errorf("primary (%s): %w, fallback (%s): %v", p.apiPrimary, err, p.apiFallback, err2)
	}

	return resp, nil
}

// streamFromDomain sends a streaming request to a specific base URL.
func (p *SharedChatProvider) streamFromDomain(ctx context.Context, baseURL string, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error) {
	// 1. Get access token (auto-refresh, scoped to domain)
	token, err := p.getAccessToken(ctx, baseURL)
	if err != nil {
		return nil, fmt.Errorf("token error (%s): %w", baseURL, err)
	}

	// 2. Build payload with conversation context
	payload := map[string]any{
		"messages": []map[string]string{
			{"role": "user", "content": req.Messages[len(req.Messages)-1].Content},
		},
		"userId": p.userID,
	}

	p.mu.Lock()
	if p.conversationID != "" {
		payload["conversationId"] = p.conversationID
	}
	p.mu.Unlock()

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	// 3. Create request with timeout context
	reqCtx, cancel := context.WithTimeout(ctx, p.requestTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, "POST",
		baseURL+"/backend-api/conversation",
		bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("User-Agent", "Ti-CLI/1.0")

	// Attach cookies if available
	for _, ck := range p.cookies {
		httpReq.AddCookie(ck)
	}

	// 4. Send request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request (%s): %w", baseURL, err)
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		bodyStr := string(body)

		// Detect session expiry for cookie auto-refresh
		if strings.Contains(bodyStr, "会话已失效") || strings.Contains(bodyStr, "expired") || resp.StatusCode == http.StatusUnauthorized {
			// Force token refresh on next call
			p.mu.Lock()
			p.accessToken = ""
			p.tokenExpireAt = time.Time{}
			p.mu.Unlock()
			return nil, fmt.Errorf("session expired (401): please update gfsessionid cookie")
		}

		return nil, fmt.Errorf("API error %d (%s): %s", resp.StatusCode, baseURL, bodyStr)
	}

	// 5. Parse SSE stream
	result, err := p.parseSSE(ctx, resp.Body, onChunk)
	if err != nil {
		return result, fmt.Errorf("parse stream (%s): %w", baseURL, err)
	}

	return result, nil
}

// getAccessToken fetches and caches the Bearer token from /api/auth/session.
// Token is scoped to the base URL (different domains may have different tokens).
// If cached token is still valid, returns it immediately.
func (p *SharedChatProvider) getAccessToken(ctx context.Context, baseURL string) (string, error) {
	p.mu.Lock()
	// Return cached token if still valid (with 30s safety margin)
	if p.accessToken != "" && time.Now().Add(30*time.Second).Before(p.tokenExpireAt) {
		token := p.accessToken
		p.mu.Unlock()
		return token, nil
	}
	p.mu.Unlock()

	// Fetch fresh token
	reqURL := baseURL + "/api/auth/session"
	httpReq, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("create auth request: %w", err)
	}

	// Set cookie header for authentication
	httpReq.Header.Set("Cookie", fmt.Sprintf("gfsessionid=%s", p.sessionID))
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "Ti-CLI/1.0")

	// Attach cookies if available
	for _, ck := range p.cookies {
		httpReq.AddCookie(ck)
	}

	authClient := &http.Client{Timeout: 15 * time.Second}
	resp, err := authClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("auth failed (%d): %s", resp.StatusCode, string(body))
	}

	var data struct {
		AccessToken string `json:"accessToken"`
		ExpiresIn   int    `json:"expiresIn"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode auth response: %w", err)
	}

	if data.AccessToken == "" {
		return "", fmt.Errorf("empty access token received")
	}

	// Cache token with expiry
	p.mu.Lock()
	p.accessToken = data.AccessToken
	expires := time.Duration(data.ExpiresIn) * time.Second
	if expires == 0 {
		expires = 10 * time.Minute // Default 10 minutes
	}
	p.tokenExpireAt = time.Now().Add(expires)
	p.mu.Unlock()

	return data.AccessToken, nil
}

// parseSSE reads Server-Sent Events from the response body.
// Handles both "data: {json}" and raw JSON formats.
// Tracks conversationId for continuity across turns.
func (p *SharedChatProvider) parseSSE(ctx context.Context, body io.ReadCloser, onChunk func(StreamChunk)) (*ChatResponse, error) {
	result := &ChatResponse{FinishReason: "stop"}
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024) // 10MB max line

	var fullContent strings.Builder

	for scanner.Scan() {
		// Check for context cancellation (Ctrl+C support)
		select {
		case <-ctx.Done():
			result.Content = fullContent.String()
			result.FinishReason = "cancelled"
			if onChunk != nil {
				onChunk(StreamChunk{Done: true})
			}
			return result, ctx.Err()
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// Handle SSE format "data: { ... }"
		if strings.HasPrefix(line, "data:") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}

		// Skip [DONE] marker
		if line == "[DONE]" {
			break
		}

		// Skip non-JSON lines
		if !strings.HasPrefix(line, "{") {
			// Treat as raw text content
			fullContent.WriteString(line)
			if onChunk != nil {
				onChunk(StreamChunk{Content: line})
			}
			continue
		}

		// Parse JSON chunk
		var chunk struct {
			Role           string `json:"role"`
			Content        string `json:"content"`
			MessageChunk   string `json:"messageChunk"`
			ConversationID string `json:"conversationId,omitempty"`
			Done           bool   `json:"done"`
		}
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			// If not valid JSON, treat as raw text
			fullContent.WriteString(line)
			if onChunk != nil {
				onChunk(StreamChunk{Content: line})
			}
			continue
		}

		// Update conversation ID for continuity
		if chunk.ConversationID != "" {
			p.mu.Lock()
			p.conversationID = chunk.ConversationID
			p.mu.Unlock()
		}

		// Extract content from various possible fields
		content := chunk.Content
		if content == "" {
			content = chunk.MessageChunk
		}

		if content != "" {
			fullContent.WriteString(content)
			if onChunk != nil {
				onChunk(StreamChunk{Content: content})
			}
		}

		// Check for done signal
		if chunk.Done {
			result.FinishReason = "stop"
			break
		}
	}

	result.Content = fullContent.String()

	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("SSE scanner error: %w", err)
	}

	// Signal completion
	if onChunk != nil {
		onChunk(StreamChunk{Done: true})
	}

	return result, nil
}

// calculateBackoff computes exponential backoff with jitter.
// Formula: min(baseDelay * 2^attempt + randomJitter, maxDelay)
func (p *SharedChatProvider) calculateBackoff(attempt int) time.Duration {
	// Exponential: baseDelay * 2^(attempt-1)
	backoff := p.baseDelay * time.Duration(1<<uint(attempt-1))

	// Cap at maxDelay
	if backoff > p.maxDelay {
		backoff = p.maxDelay
	}

	// Add jitter: +/- 20% of backoff
	jitterRange := backoff / 5
	jitter := time.Duration(rand.Int63n(int64(jitterRange)*2)) - jitterRange
	backoff += jitter

	// Ensure positive duration
	if backoff < 0 {
		backoff = p.baseDelay / 2
	}

	return backoff
}

// SessionID returns the current gfsessionid.
func (p *SharedChatProvider) SessionID() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.sessionID
}

// ConversationID returns the current conversation ID for continuity.
func (p *SharedChatProvider) ConversationID() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.conversationID
}

// SetConversationID sets the conversation ID to continue an existing conversation.
func (p *SharedChatProvider) SetConversationID(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conversationID = id
}

// ResetConversation clears the conversation context.
func (p *SharedChatProvider) ResetConversation() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conversationID = ""
}

// RefreshToken forces a fresh token fetch from the auth endpoint.
// Useful when the current token is known to be expired.
func (p *SharedChatProvider) RefreshToken(ctx context.Context) error {
	p.mu.Lock()
	p.accessToken = ""
	p.tokenExpireAt = time.Time{}
	p.mu.Unlock()

	_, err := p.getAccessToken(ctx, p.apiPrimary)
	if err != nil {
		// Try fallback domain
		_, err = p.getAccessToken(ctx, p.apiFallback)
	}
	return err
}

// UpdateCookies reloads cookies from the provided list.
// Resets token cache since session may have changed.
func (p *SharedChatProvider) UpdateCookies(cookies []*http.Cookie) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.cookies = cookies
	// Reset token cache
	p.accessToken = ""
	p.tokenExpireAt = time.Time{}

	// Update sessionID from new cookies
	for _, ck := range cookies {
		if ck.Name == "gfsessionid" {
			p.sessionID = ck.Value
			break
		}
	}
}

// Cookies returns current cookies (for inspection/debug).
func (p *SharedChatProvider) Cookies() []*http.Cookie {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cookies
}

// APIDomains returns the current primary and fallback domains.
func (p *SharedChatProvider) APIDomains() (primary, fallback string) {
	return p.apiPrimary, p.apiFallback
}

// SetAPIDomains updates the primary and fallback domains.
func (p *SharedChatProvider) SetAPIDomains(primary, fallback string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.apiPrimary = primary
	p.apiFallback = fallback
}

// ConsecutiveFailures returns the number of consecutive failed requests.
func (p *SharedChatProvider) ConsecutiveFailures() int64 {
	return p.consecutiveFailures.Load()
}

// LastSuccessAt returns the timestamp of the last successful request.
func (p *SharedChatProvider) LastSuccessAt() time.Time {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastSuccessAt
}
