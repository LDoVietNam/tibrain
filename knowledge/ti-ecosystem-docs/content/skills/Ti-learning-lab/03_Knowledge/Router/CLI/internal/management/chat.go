package management

// Chat proxy — OpenAI-compatible /v1/chat/completions and /v1/models endpoints.
// Integrated directly into the management server so all traffic flows through
// a single port (default 1808/1911).
//
// Request flow:
//   IDE → POST /v1/chat/completions
//       → autocombo.Select() picks best provider/model
//       → inject Ti system prompt if missing
//       → proxy to real upstream with SSE streaming
//       → response back to IDE

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ti/cli/internal/autocombo"
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
	"anthropic":  "https://api.anthropic.com/v1",
}

// tiSystemPrompt is injected as the first system message when none exists.
const tiSystemPrompt = `You are Ti, an expert developer assistant built into the Ti IDE.
You have deep knowledge of software engineering, architecture, debugging, and code review.
You are precise, concise, and always provide actionable answers.
When writing code, follow the project's existing style and conventions.`

// chatHTTPClient is shared across requests.
var chatHTTPClient = &http.Client{Timeout: 300 * time.Second}

// chatMessage mirrors the OpenAI message format.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest is the incoming request body from the IDE.
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// registerChatRoutes adds /v1/chat/completions and /v1/models to mux.
func registerChatRoutes(mux *http.ServeMux, svc *Service) {
	mux.HandleFunc("/v1/models", handleModels())
	mux.HandleFunc("/v1/chat/completions", handleChat(svc))
}

// ── /v1/models ────────────────────────────────────────────────────────────────

func handleModels() http.HandlerFunc {
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
		models := []modelObj{{ID: "auto", Object: "model", OwnedBy: "ti"}}
		seen := map[string]bool{"auto": true}
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
		writeJSON(w, http.StatusOK, map[string]any{"object": "list", "data": models})
	}
}

// ── /v1/chat/completions ──────────────────────────────────────────────────────

func handleChat(svc *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 1. Parse request
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		// 2. Inject Ti system prompt if missing
		req.Messages = chatInjectSystemPrompt(req.Messages)

		// 3. Route via autocombo — picks best provider/model for this task
		taskType := chatInferTaskType(req.Messages)
		selection := autocombo.Select(autocombo.AutoComboConfig{
			ID:              "chat-proxy",
			Name:            "Ti Chat Proxy",
			Weights:         autocombo.DefaultWeights,
			ExplorationRate: 0.05,
		}, router.AutoComboPool(), taskType)

		providerName := selection.Provider
		model := selection.Model

		// Override if caller specified a concrete "provider/model"
		if req.Model != "" && req.Model != "auto" {
			if idx := strings.Index(req.Model, "/"); idx > 0 {
				p := req.Model[:idx]
				if _, known := providerBaseURLs[p]; known {
					providerName = p
					model = req.Model[idx+1:]
				}
			} else {
				// bare model name — keep autocombo provider
				model = req.Model
			}
		}

		// 4. Resolve API key from config
		apiKey := ""
		if svc.cfg.APIKeys != nil {
			apiKey = svc.cfg.APIKeys[providerName]
		}
		if apiKey == "" {
			http.Error(w, fmt.Sprintf(
				"no API key for provider %q — add to config api_keys or set env var",
				providerName,
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

		// 6. Build upstream body — convert chatMessage → router.Message
		upMsgs := make([]router.Message, len(req.Messages))
		for i, m := range req.Messages {
			upMsgs[i] = router.Message{Role: m.Role, Content: m.Content}
		}
		upBody, err := json.Marshal(map[string]any{
			"model":       model,
			"messages":    upMsgs,
			"max_tokens":  req.MaxTokens,
			"temperature": req.Temperature,
			"stream":      req.Stream,
		})
		if err != nil {
			http.Error(w, "marshal error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 7. Forward to upstream
		upReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstreamURL, bytes.NewReader(upBody))
		if err != nil {
			http.Error(w, "upstream request error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		upReq.Header.Set("Content-Type", "application/json")
		upReq.Header.Set("Authorization", "Bearer "+apiKey)
		upReq.Header.Set("User-Agent", "TiProxy/1.0")
		if providerName == "openrouter" {
			upReq.Header.Set("HTTP-Referer", "https://ti.dev")
			upReq.Header.Set("X-Title", "Ti IDE")
		}

		resp, err := chatHTTPClient.Do(upReq)
		if err != nil {
			http.Error(w, "upstream error: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// 8. Stream response back
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.Header().Set("X-Ti-Provider", providerName)
		w.Header().Set("X-Ti-Model", model)
		w.Header().Set("X-Ti-Task", taskType)
		w.Header().Set("X-Ti-Score", fmt.Sprintf("%.3f", selection.Score))
		w.WriteHeader(resp.StatusCode)

		if req.Stream {
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
				if readErr == io.EOF || readErr != nil {
					break
				}
			}
		} else {
			_, _ = io.Copy(w, resp.Body)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func chatInjectSystemPrompt(msgs []chatMessage) []chatMessage {
	for _, m := range msgs {
		if m.Role == "system" {
			return msgs
		}
	}
	return append([]chatMessage{{Role: "system", Content: tiSystemPrompt}}, msgs...)
}

func chatInferTaskType(msgs []chatMessage) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		c := strings.ToLower(msgs[i].Content)
		switch {
		case strings.Contains(c, "fix") || strings.Contains(c, "bug") ||
			strings.Contains(c, "error") || strings.Contains(c, "debug"):
			return "debugging"
		case strings.Contains(c, "review") || strings.Contains(c, "check"):
			return "review"
		case strings.Contains(c, "plan") || strings.Contains(c, "design") ||
			strings.Contains(c, "architect"):
			return "planning"
		case strings.Contains(c, "write") || strings.Contains(c, "implement") ||
			strings.Contains(c, "create") || strings.Contains(c, "code"):
			return "coding"
		}
	}
	return "default"
}
