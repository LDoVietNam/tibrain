package management

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/ti/cli/internal/api"
	"github.com/ti/cli/internal/router"
)

func NewHandler(svc *Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/management/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "service": "ti-management", "version": version})
	})
	mux.HandleFunc("/api/management/bootstrap", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, svc.Bootstrap())
	})
	mux.HandleFunc("/api/management/dashboard", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, svc.Dashboard())
	})
	mux.HandleFunc("/api/management/providers", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"items": svc.Providers()})
	})
	mux.HandleFunc("/api/management/api-keys", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, map[string]any{"items": svc.APIKeys()})
		case http.MethodPost:
			var req AddAPIKeyRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, WriteResult{OK: false, Msg: "invalid JSON: " + err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, svc.AddAPIKey(req))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/management/auth-files", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, map[string]any{"items": svc.AuthFiles()})
		case http.MethodPost:
			var req AddAuthFileRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, WriteResult{OK: false, Msg: "invalid JSON: " + err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, svc.AddAuthFile(req))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/management/cookies", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, map[string]any{"items": svc.Cookies(), "starters": svc.CookieStarters()})
		case http.MethodPost:
			var req AddCookieRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, WriteResult{OK: false, Msg: "invalid JSON: " + err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, svc.AddCookie(req))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/management/oauth", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, map[string]any{"items": svc.OAuthConnections(), "starters": svc.OAuthStarters()})
		case http.MethodPost:
			var req AddOAuthRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, WriteResult{OK: false, Msg: "invalid JSON: " + err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, svc.AddOAuth(req))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/management/config", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, svc.ConfigView())
	})
	mux.HandleFunc("/api/management/system", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, svc.System())
	})
	mux.HandleFunc("/api/management/logs", func(w http.ResponseWriter, r *http.Request) {
		tail := 200
		if raw := strings.TrimSpace(r.URL.Query().Get("tail")); raw != "" {
			if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 5000 {
				tail = n
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"tail": tail, "lines": tailFile(svc.cfg.LogFile, tail)})
	})
	mux.HandleFunc("/v1/autocombo/models", func(w http.ResponseWriter, r *http.Request) {
		taskType := strings.TrimSpace(r.URL.Query().Get("task"))
		writeJSON(w, http.StatusOK, map[string]any{
			"pool":      router.AutoComboPool(),
			"task_type": taskType,
			"selected":  router.AutoComboSelect(taskType),
		})
	})
	mux.HandleFunc("/v1/models", svc.handleModels)
	mux.HandleFunc("/api/management/antigravity-auth-url", api.AntigravityAuthURL)
	mux.HandleFunc("/api/management/antigravity/callback", api.AntigravityCallback)
	mux.HandleFunc("/api/management/antigravity/health", api.AntigravityHealth)
	mux.HandleFunc("/v1/embeddings", svc.handleEmbeddings)
	mux.HandleFunc("/v1/images/generations", svc.handleImageGenerations)
	mux.HandleFunc("/v1/audio/speech", svc.handleAudioSpeech)
	mux.HandleFunc("/v1/moderations", svc.handleModerations)
	mux.HandleFunc("/v1/rerank", svc.handleRerank)
	mux.HandleFunc("/v1/responses", svc.handleResponses)
	mux.HandleFunc("/v1", svc.handleV1Root)
	mux.HandleFunc("/v1/messages", svc.handleMessages)
	mux.HandleFunc("/v1/messages/count_tokens", svc.handleMessagesCountTokens)
	mux.HandleFunc("/v1/api/chat", svc.handleV1APIChat)
	mux.HandleFunc("/v1/providers/", svc.handleProviderScopedV1)
	mux.HandleFunc("/v1/audio/transcriptions", svc.handleAudioTranscriptions)
	mux.HandleFunc("/v1/videos/generations", svc.handleVideoGenerations)
	mux.HandleFunc("/v1/music/generations", svc.handleMusicGenerations)
	mux.HandleFunc("/v1/chat/completions", svc.handleChatCompletions)
	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, X-Request-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func tailFile(path string, maxLines int) []string {
	if strings.TrimSpace(path) == "" {
		return []string{}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{}
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	clean := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		clean = append(clean, line)
	}
	return clean
}
