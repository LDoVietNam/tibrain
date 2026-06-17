// Package chatproxy provides an OpenAI-compatible /v1/chat/completions proxy
// that intercepts requests from Antigravity IDE, injects the Ti system prompt,
// selects the best provider via routeagent/autocombo, and streams the response.
package chatproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ti/cli/internal/config"
	"github.com/ti/cli/internal/router"
)

// providerBaseURLs maps provider name → OpenAI-compatible base URL.
var providerBaseURLs = map[string]string{
	"openrouter": "https://openrouter.ai/api/v1",
	"groq":       "https://api.groq.com/openai/v1",
	"cerebras":   "https://api.cerebras.ai/v1",
	"deepseek":   "https://api.deepseek.com/v1",
	"google":     "https://generativelanguage.googleapis.com/v1beta/openai",
	"openai":     "https://api.openai.com/v1",
	"anthropic":  "https://api.anthropic.com/v1", // via compat layer
}

// providerAPIKeyEnv maps provider name → env var name (for error messages).
var providerAPIKeyEnv = map[string]string{
	"openrouter": "OPENROUTER_API_KEY",
	"groq":       "GROQ_API_KEY",
	"cerebras":   "CEREBRAS_API_KEY",
	"deepseek":   "DEEPSEEK_API_KEY",
	"google":     "GOOGLE_API_KEY",
	"openai":     "OPENAI_API_KEY",
	"anthropic":  "ANTHROPIC_API_KEY",
}

// tiSystemPrompt is injected as the first system message when no system message exists.
const tiSystemPrompt = `You are Ti, an expert developer assistant built into the Ti IDE.
You have deep knowledge of software engineering, architecture, debugging, and code review.
You are precise, concise, and always provide actionable answers.
When writing code, follow the project's existing style and conventions.`

// Handler returns an http.Handler that serves:
//
//	POST /v1/chat/completions  — proxied + Ti-augmented chat
//	GET  /v1/models            — list of available Ti models
func Handler(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models", modelsHandler(cfg))
	mux.HandleFunc("/v1/chat/completions", chatHandler(cfg))
	return mux
}

// ── /v1/models ────────────────────────────────────────────────────────────────

func modelsHandler(_ *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		pool := router.AutoComboPool()
		type modelObj struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			OwnedBy string `json:"owned_by"`
		}
		models := make([]modelObj, 0, len(pool))
		seen := map[string]bool{}
		for _, c := range pool {
			id := c.Provider + "/" + c.Model
			if seen[id] {
				continue
			}
			seen[id] = true
			models = append(models, modelObj{
				ID:      id,
				Object:  "model",
				OwnedBy: c.Provider,
			})
		}
		// Always include "auto"
		models = append([]modelObj{{ID: "auto", Object: "model", OwnedBy: "ti"}}, models...)
		writeJSON(w, map[string]any{"object": "list", "data": models})
	}
}

// ── /v1/chat/completions ──────────────────────────────────────────────────────

type chatReq struct {
	Model       string           `json:"model"`
	Messages    []router.Message `json:"messages"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
}

func chatHandler(cfg *config.Config) http.HandlerFunc {
	httpClient := &http.Client{Timeout: 300 * time.Second}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 1. Parse incoming request
		var req chatReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		// 2. Inject Ti system prompt if no system message present
		req.Messages = injectSystemPrompt(req.Messages)

		// 3. Resolve provider + model via autocombo
		providerName, model := resolveProviderModel(req.Model, req.Messages)

		// 4. Resolve API key
		apiKey := resolveAPIKey(cfg, providerName)
		if apiKey == "" {
			envVar := providerAPIKeyEnv[providerName]
			http.Error(w, fmt.Sprintf(
				"no API key for provider %q — set %s or add to config api_keys",
				providerName, envVar,
			), http.StatusUnauthorized)
			return
		}

		// 5. Build upstream URL
		baseURL, ok := providerBaseURLs[providerName]
		if !ok {
			http.Error(w, "unknown provider: "+providerName, http.StatusBadRequest)
			return
		}
		upstreamURL := strings.TrimSuffix(baseURL, "/") + "/chat/completions"

		// 6. Build upstream request body
		upstreamReq := chatReq{
			Model:       model,
			Messages:    req.Messages,
			MaxTokens:   req.MaxTokens,
			Temperature: req.Temperature,
			Stream:      req.Stream,
		}
		body, err := json.Marshal(upstreamReq)
		if err != nil {
			http.Error(w, "marshal error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 7. Forward to upstream
		upReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstreamURL, bytes.NewReader(body))
		if err != nil {
			http.Error(w, "upstream request error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		upReq.Header.Set("Content-Type", "application/json")
		upReq.Header.Set("Authorization", "Bearer "+apiKey)
		upReq.Header.Set("User-Agent", "TiProxy/1.0")
		// OpenRouter requires HTTP-Referer
		if providerName == "openrouter" {
			upReq.Header.Set("HTTP-Referer", "https://ti.dev")
			upReq.Header.Set("X-Title", "Ti IDE")
		}

		resp, err := httpClient.Do(upReq)
		if err != nil {
			http.Error(w, "upstream error: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// 8. Stream response back to IDE
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.Header().Set("X-Ti-Provider", providerName)
		w.Header().Set("X-Ti-Model", model)
		w.WriteHeader(resp.StatusCode)

		if req.Stream {
			// Flush-aware streaming
			flusher, canFlush := w.(http.Flusher)
			buf := make([]byte, 4096)
			for {
				n, readErr := resp.Body.Read(buf)
				if n > 0 {
					_, _ = w.Write(buf[:n])
					if canFlush {
						flusher.Flush()
					}
				}
				if readErr == io.EOF {
					break
				}
				if readErr != nil {
					break
				}
			}
		} else {
			_, _ = io.Copy(w, resp.Body)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// injectSystemPrompt prepends the Ti system prompt if no system message exists.
func injectSystemPrompt(msgs []router.Message) []router.Message {
	for _, m := range msgs {
		if m.Role == "system" {
			return msgs
		}
	}
	sys := router.Message{Role: "system", Content: tiSystemPrompt}
	return append([]router.Message{sys}, msgs...)
}

// resolveProviderModel picks provider + model from the request model string.
// Supports "auto", "provider/model", or bare model names.
func resolveProviderModel(model string, msgs []router.Message) (provider, resolvedModel string) {
	if model == "" || model == "auto" {
		taskType := inferTaskType(msgs)
		return router.SelectAutoModel(taskType, nil)
	}
	// "provider/model" format
	if idx := strings.Index(model, "/"); idx > 0 {
		p := model[:idx]
		m := model[idx+1:]
		if _, known := providerBaseURLs[p]; known {
			return p, m
		}
	}
	// Bare model — use configured provider
	return "openrouter", model
}

// resolveAPIKey looks up the API key for a provider from config.
func resolveAPIKey(cfg *config.Config, provider string) string {
	if cfg.APIKeys != nil {
		if k := cfg.APIKeys[provider]; k != "" {
			return k
		}
	}
	return ""
}

// inferTaskType is a lightweight heuristic matching router.inferTaskType logic.
func inferTaskType(msgs []router.Message) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		content := strings.ToLower(msgs[i].Content)
		switch {
		case strings.Contains(content, "fix") || strings.Contains(content, "bug") ||
			strings.Contains(content, "error") || strings.Contains(content, "debug"):
			return "debugging"
		case strings.Contains(content, "review") || strings.Contains(content, "check"):
			return "review"
		case strings.Contains(content, "plan") || strings.Contains(content, "design") ||
			strings.Contains(content, "architect"):
			return "planning"
		case strings.Contains(content, "write") || strings.Contains(content, "implement") ||
			strings.Contains(content, "create") || strings.Contains(content, "code"):
			return "coding"
		}
	}
	return "default"
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
