package management

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ti/cli/internal/brain"
	"github.com/ti/cli/internal/config"
	"github.com/ti/cli/internal/learn"
)

// OpenAIChatRequest matches the OpenAI chat completions request format.
type OpenAIChatRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature *float64        `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIChatResponse matches the OpenAI chat completions response format.
type OpenAIChatResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage,omitempty"`
}

type OpenAIChoice struct {
	Index   int            `json:"index"`
	Message OpenAIMessage  `json:"message,omitempty"`
	Delta   *OpenAIMessage `json:"delta,omitempty"`
}

type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIError matches the OpenAI error response format.
type OpenAIError struct {
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Param   interface{} `json:"param"`
	Code    interface{} `json:"code"`
}

type OpenAIErrorResponse struct {
	Error OpenAIError `json:"error"`
}

func writeOpenAIError(w http.ResponseWriter, status int, message, errType string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(OpenAIErrorResponse{
		Error: OpenAIError{
			Message: message,
			Type:    errType,
			Param:   nil,
			Code:    nil,
		},
	})
}

func (svc *Service) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeOpenAIError(w, http.StatusMethodNotAllowed, "method not allowed", "invalid_request_error")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "read body: "+err.Error(), "invalid_request_error")
		return
	}
	defer r.Body.Close()

	var req OpenAIChatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "parse request: "+err.Error(), "invalid_request_error")
		return
	}

	if err := validateClientAuth(r, svc.cfg); err != nil {
		writeOpenAIError(w, http.StatusUnauthorized, err.Error(), "authentication_error")
		return
	}

	// Build prompt text from messages for brain analysis
	var promptText string
	for _, m := range req.Messages {
		if m.Role == "user" {
			promptText += m.Content + "\n"
		}
	}

	// Resolve model: explicit → Router Agent /decide → BEADS v2 → brain v1 → config default
	model := req.Model
	if strings.TrimSpace(model) == "" {
		// 1) Try Router Agent /decide
		if agentModel := queryRouterAgentDecide(svc.cfg, promptText); agentModel != "" {
			model = agentModel
		}
		// 2) Try BEADS v2
		if model == "" {
			if eng := learn.GetEngine(); eng.IsEnabled() && eng.GetMode() != "off" {
				analysis := eng.AnalyzePrompt(promptText)
				decision := eng.SelectModel(analysis, "")
				if decision.Model != "" && decision.Confidence > 0.5 {
					model = decision.Model
				}
			}
		}
		// 3) Fallback to brain v1
		if model == "" {
			if cfg := svc.cfg; cfg != nil {
				if eng := openBrainFromCfg(cfg); eng != nil {
					taskType := brain.GetTaskType(promptText)
					if pref := eng.GetPreferredModel(taskType); pref != "" {
						model = pref
					}
				}
			}
		}
		// 4) Final fallback from config
		if model == "" && svc.cfg != nil {
			model = svc.cfg.Model
		}
	}
	if strings.TrimSpace(model) == "" {
		model = "claude-sonnet-4"
	}

	// Determine upstream base URL and API key
	var upstreamBase, apiKey string
	if key := svc.cfg.APIKeys["openrouter"]; key != "" {
		upstreamBase = "https://openrouter.ai/api"
		apiKey = key
	} else if key := svc.cfg.APIKeys["openai"]; key != "" {
		upstreamBase = "https://api.openai.com"
		apiKey = key
	} else if key := svc.cfg.APIKeys["anthropic"]; key != "" {
		// Anthropic doesn't have a standard /v1/chat/completions endpoint
		// Try Ti-Router as fallback
		upstreamBase = "http://localhost:1807"
		apiKey = key
	} else {
		// No API key configured; try Ti-Router
		upstreamBase = "http://localhost:1807"
	}

	// Rewrite request body with resolved model
	req.Model = model
	newBody, err := json.Marshal(req)
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, "marshal request: "+err.Error(), "server_error")
		return
	}

	// Build target URL
	targetURL, err := url.Parse(upstreamBase + "/v1/chat/completions")
	if err != nil {
		writeOpenAIError(w, http.StatusInternalServerError, "parse upstream: "+err.Error(), "server_error")
		return
	}

	// Record start time for brain logging
	started := time.Now()

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.FlushInterval = 50 * time.Millisecond // flush SSE chunks immediately
	proxy.Director = func(outReq *http.Request) {
		outReq.URL = targetURL
		outReq.Host = targetURL.Host
		outReq.Header.Set("Authorization", "Bearer "+apiKey)
		outReq.Header.Set("Content-Type", "application/json")
		outReq.Body = io.NopCloser(bytes.NewReader(newBody))
		outReq.ContentLength = int64(len(newBody))
		outReq.Header.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))
		outReq.Header.Del("Authorization") // remove client auth if present
		if apiKey != "" {
			outReq.Header.Set("Authorization", "Bearer "+apiKey)
		}
	}

	// Capture response for brain logging
	proxy.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode == http.StatusOK {
			if req.Stream {
				// Streaming: forward chunks immediately; log success without body capture
				go func() {
					recordBrainLogFromAPI(svc.cfg, promptText, model, true, time.Since(started).Milliseconds())
				}()
				return nil
			}
			// Tee response body for brain logging (non-stream only)
			var buf bytes.Buffer
			tee := io.TeeReader(resp.Body, &buf)
			resp.Body = io.NopCloser(tee)
			go func() {
				all, _ := io.ReadAll(&buf)
				var openAIResp OpenAIChatResponse
				if err := json.Unmarshal(all, &openAIResp); err == nil {
					success := len(openAIResp.Choices) > 0
					recordBrainLogFromAPI(svc.cfg, promptText, model, success, time.Since(started).Milliseconds())
				}
			}()
		}
		return nil
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		writeOpenAIError(w, http.StatusBadGateway, "upstream error: "+err.Error(), "server_error")
	}

	proxy.ServeHTTP(w, r)
}

func validateClientAuth(r *http.Request, cfg *config.Config) error {
	var mgmtKey string
	if cfg != nil {
		mgmtKey = cfg.APIKeys["management"]
	}
	if mgmtKey == "" {
		mgmtKey = os.Getenv("TI_MGMT_API_KEY")
	}
	if strings.TrimSpace(mgmtKey) == "" {
		return nil // no management key configured; allow all
	}
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return fmt.Errorf("missing or invalid authorization header")
	}
	token := strings.TrimSpace(auth[len(prefix):])
	if subtle.ConstantTimeCompare([]byte(token), []byte(mgmtKey)) != 1 {
		return fmt.Errorf("invalid API key")
	}
	return nil
}

func brainDataDirFromCfg(cfg *config.Config) string {
	base := ""
	if cfg != nil && strings.TrimSpace(cfg.DataDir) != "" {
		base = filepath.Join(cfg.DataDir, "brain")
	} else {
		// Persist brain data under ~/.ti/data/brain instead of temp
		home, _ := os.UserHomeDir()
		if home != "" {
			base = filepath.Join(home, ".ti", "data", "brain")
		} else {
			base = filepath.Join(os.TempDir(), "ti-brain")
		}
	}
	return base
}

// OpenAIModelItem matches the OpenAI /v1/models response.
type OpenAIModelItem struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}
type OpenAIModelList struct {
	Object string            `json:"object"`
	Data   []OpenAIModelItem `json:"data"`
}

func (svc *Service) handleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeOpenAIError(w, http.StatusMethodNotAllowed, "method not allowed", "invalid_request_error")
		return
	}
	now := time.Now().Unix()
	list := OpenAIModelList{Object: "list", Data: []OpenAIModelItem{}}

	// Build from runtime provider registry when available
	if svc.registry != nil {
		seen := map[string]bool{}
		for _, snap := range svc.registry.Snapshots() {
			owner := snap.Name
			models := append([]string(nil), snap.Models...)
			if snap.DefaultModel != "" {
				models = append([]string{snap.DefaultModel}, models...)
			}
			for _, id := range models {
				id = strings.TrimSpace(id)
				if id == "" || seen[id] {
					continue
				}
				seen[id] = true
				list.Data = append(list.Data, OpenAIModelItem{ID: id, Object: "model", Created: now, OwnedBy: owner})
			}
		}
	}

	// Fallback baseline when registry empty
	if len(list.Data) == 0 {
		list.Data = []OpenAIModelItem{
			{ID: "claude-sonnet-4", Object: "model", Created: now, OwnedBy: "anthropic"},
			{ID: "claude-sonnet-4-20251022", Object: "model", Created: now, OwnedBy: "anthropic"},
			{ID: "gemini-2.5-pro", Object: "model", Created: now, OwnedBy: "google"},
			{ID: "gemini-2.5-flash", Object: "model", Created: now, OwnedBy: "google"},
			{ID: "openrouter/auto", Object: "model", Created: now, OwnedBy: "openrouter"},
			{ID: "gpt-4o", Object: "model", Created: now, OwnedBy: "openai"},
			{ID: "gpt-4o-mini", Object: "model", Created: now, OwnedBy: "openai"},
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(list)
}

func openBrainFromCfg(cfg *config.Config) *brain.Engine {
	eng, err := brain.NewEngine(brain.Config{
		DataDir:       brainDataDirFromCfg(cfg),
		LearningRate:  0.1,
		MinConfidence: 0.75,
	})
	if err != nil {
		return nil
	}
	return eng
}

func recordBrainLogFromAPI(cfg *config.Config, prompt, model string, success bool, latencyMs int64) {
	if cfg == nil {
		return
	}
	taskType := brain.GetTaskType(prompt)

	// Always send to Router Agent asynchronously (independent of brain engine)
	go sendRecordToRouterAgent(cfg, prompt, model, taskType, success, latencyMs)

	// Also log to local brain engine when available
	eng := openBrainFromCfg(cfg)
	if eng == nil {
		return
	}
	log := brain.RouterLogEntry{
		UserPrompt: prompt,
		Model:      model,
		TaskType:   taskType,
		Success:    success,
		LatencyMs:  latencyMs,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
	_ = eng.IngestLogs(context.TODO(), []brain.RouterLogEntry{log})
}

// sendRecordToRouterAgent sends a record to the Router Agent /record endpoint.
func sendRecordToRouterAgent(cfg *config.Config, prompt, model, taskType string, success bool, latencyMs int64) {
	agentURL := strings.TrimSpace(cfg.RouterAgentURL)
	if agentURL == "" {
		agentURL = os.Getenv("TI_ROUTER_AGENT_URL")
	}
	if agentURL == "" {
		return // no router agent configured
	}

	payload := map[string]interface{}{
		"user_prompt": prompt,
		"model":       model,
		"task_type":   taskType,
		"success":     success,
		"latency_ms":  latencyMs,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	url := strings.TrimRight(agentURL, "/") + "/record"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

// queryRouterAgentDecide queries the Router Agent /decide endpoint for a model recommendation.
func queryRouterAgentDecide(cfg *config.Config, prompt string) string {
	if cfg == nil {
		return ""
	}
	agentURL := strings.TrimSpace(cfg.RouterAgentURL)
	if agentURL == "" {
		agentURL = os.Getenv("TI_ROUTER_AGENT_URL")
	}
	if agentURL == "" {
		return "" // no router agent configured
	}

	payload := map[string]interface{}{
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"model": "",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return ""
	}

	url := strings.TrimRight(agentURL, "/") + "/decide"
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}
	var result struct {
		Decision struct {
			Model      string  `json:"model"`
			Confidence float64 `json:"confidence"`
		} `json:"decision"`
		Reason string `json:"reason,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}
	if result.Decision.Confidence >= 0.5 && result.Decision.Model != "" {
		return result.Decision.Model
	}
	return ""
}
