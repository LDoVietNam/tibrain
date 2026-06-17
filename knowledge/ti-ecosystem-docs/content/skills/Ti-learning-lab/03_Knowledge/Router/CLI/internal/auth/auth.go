package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ti/cli/internal/config"
)

const (
	StatusValid   = "valid"
	StatusExpired = "expired"
	StatusMissing = "missing"
)

var (
	sourcePriority = map[string]int{"env": 100, "config": 90, "auth_file": 85, "token_file": 80, "cookie_file": 70, "oauth_file": 60, "oauth": 60, "external": 50}
	previewCache   sync.Map // Cache for previewSecret to avoid repeated allocations
)

var ProviderMap = map[string]string{
	"OPENROUTER_API_KEY": "openrouter",
	"GROQ_API_KEY":       "groq",
	"CEREBRAS_API_KEY":   "cerebras",
	"DEEPSEEK_API_KEY":   "deepseek",
	"ANTHROPIC_API_KEY":  "anthropic",
	"GOOGLE_API_KEY":     "google",
	"GEMINI_API_KEY":     "google",
	"MODAL_API_KEY":      "modal",
	"EXA_API_KEY":        "exa",
	"OPENAI_API_KEY":     "openai",
}

type Entry struct {
	ProviderID string    `json:"provider_id"`
	AuthType   string    `json:"auth_type"`
	Source     string    `json:"source"`
	Value      string    `json:"value"`
	Expiry     time.Time `json:"expiry,omitempty"`
	Status     string    `json:"status"`
	LoadedAt   time.Time `json:"loaded_at"`
}

type Summary struct {
	ProviderID string    `json:"provider_id"`
	AuthType   string    `json:"auth_type"`
	Source     string    `json:"source"`
	Status     string    `json:"status"`
	Expiry     time.Time `json:"expiry,omitempty"`
	Preview    string    `json:"preview,omitempty"`
}

type Loader struct {
	mu        sync.RWMutex
	entries   sync.Map // Changed from map to sync.Map for better concurrency
	authDir   string
	authFiles []string
}

func NewLoader(authDir string) *Loader {
	if authDir == "" {
		home, _ := os.UserHomeDir()
		authDir = filepath.Join(home, ".ti", "auth")
	}
	return &Loader{authDir: authDir, authFiles: []string{}}
}

func (l *Loader) SetAuthFiles(files []string) {
	l.authFiles = files
}

func (l *Loader) LoadAll() error {
	if err := l.LoadFromEnv(); err != nil {
		return fmt.Errorf("env: %w", err)
	}
	if err := l.LoadFromConfig(filepath.Join(l.authDir, "..", "config.json")); err != nil {
	}
	if err := l.LoadFromAuthFiles(); err != nil {
		return fmt.Errorf("auth files: %w", err)
	}
	if err := l.LoadFromTokenFiles(); err != nil {
		return fmt.Errorf("token files: %w", err)
	}
	if err := l.LoadFromCookieFiles(); err != nil {
		return fmt.Errorf("cookie files: %w", err)
	}
	if err := l.LoadFromOAuthFiles(); err != nil {
		return fmt.Errorf("oauth files: %w", err)
	}
	return nil
}

func (l *Loader) LoadFromEnv() error {
	envKeys := map[string]string{
		"ANTHROPIC_API_KEY":  "anthropic",
		"OPENAI_API_KEY":     "openai",
		"GOOGLE_API_KEY":     "google",
		"GROQ_API_KEY":       "groq",
		"OPENROUTER_API_KEY": "openrouter",
		"CLOUDFLARE_API_KEY": "cloudflare",
	}

	now := time.Now()
	for envVar, provider := range envKeys {
		if v := strings.TrimSpace(os.Getenv(envVar)); v != "" {
			l.upsertEntry(&Entry{ProviderID: provider, AuthType: "api_key", Source: "env", Value: v, Status: StatusValid, LoadedAt: now})
		}
	}
	return nil
}

func (l *Loader) LoadFromConfig(configPath string) error {
	if configPath == "" {
		return nil
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var cfg struct {
		APIKeys map[string]string `json:"api_keys"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	for provider, key := range cfg.APIKeys {
		key = resolveConfigEnv(key)
		if key == "" {
			continue
		}
		l.upsertEntry(&Entry{ProviderID: provider, AuthType: "api_key", Source: "config", Value: key, Status: StatusValid, LoadedAt: time.Now()})
	}
	return nil
}

func (l *Loader) LoadFromAuthFiles() error {
	if len(l.authFiles) == 0 {
		return nil
	}

	for _, filePath := range l.authFiles {
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		envVars := config.ParseEnvFile(string(data))
		for envKey, provider := range ProviderMap {
			if value, ok := envVars[envKey]; ok && value != "" {
				value = config.ResolveString(value)
				l.upsertEntry(&Entry{ProviderID: provider, AuthType: "api_key", Source: "auth_file", Value: value, Status: StatusValid, LoadedAt: time.Now()})
			}
		}
	}
	return nil
}

func (l *Loader) LoadFromTokenFiles() error  { return l.loadFiles("*.token", "token") }
func (l *Loader) LoadFromCookieFiles() error { return l.loadFiles("*.cookie", "cookie") }
func (l *Loader) LoadFromOAuthFiles() error  { return l.loadFiles("*.oauth", "oauth") }

func (l *Loader) loadFiles(pattern, authType string) error {
	if err := os.MkdirAll(l.authDir, 0755); err != nil {
		return err
	}
	matches, err := filepath.Glob(filepath.Join(l.authDir, pattern))
	if err != nil {
		return err
	}
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		provider := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if entry := parseFileEntry(provider, authType, data); entry != nil {
			l.upsertEntry(entry)
		}
	}
	return nil
}

func parseFileEntry(provider, authType string, data []byte) *Entry {
	now := time.Now()
	entry := &Entry{ProviderID: provider, AuthType: authType, Source: authType + "_file", Status: StatusValid, LoadedAt: now}

	// Trim once instead of multiple times
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil
	}

	if strings.HasPrefix(trimmed, "{") {
		var raw struct{ Value, Expiry string }
		if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
			return nil
		}
		entry.Value = strings.TrimSpace(raw.Value)
		if raw.Expiry != "" {
			entry.Expiry, _ = time.Parse(time.RFC3339, raw.Expiry)
			if !entry.Expiry.IsZero() && now.After(entry.Expiry) {
				entry.Status = StatusExpired
			}
		}
		return entry
	}

	entry.Value = trimmed
	return entry
}

func (l *Loader) Get(providerID, authType string) (*Entry, bool) {
	key := keyFor(providerID, authType)
	entryRaw, ok := l.entries.Load(key)
	if !ok {
		return nil, false
	}
	entry := entryRaw.(*Entry)
	if entry.Status == StatusExpired || entry.Status == StatusMissing {
		return nil, false
	}
	copyEntry := *entry
	return &copyEntry, true
}

func (l *Loader) GetAPIKey(providerID string) (string, bool) {
	if entry, ok := l.Get(providerID, "api_key"); ok {
		return entry.Value, true
	}
	if entry, ok := l.Get(providerID, "token"); ok {
		return entry.Value, true
	}
	return "", false
}

func (l *Loader) List() []Entry {
	var keys []string
	l.entries.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	sort.Strings(keys)
	result := make([]Entry, 0, len(keys))
	for _, k := range keys {
		entryRaw, _ := l.entries.Load(k)
		entry := entryRaw.(*Entry)
		result = append(result, *entry)
	}
	return result
}

func (l *Loader) Summaries() []Summary {
	entries := l.List()
	out := make([]Summary, 0, len(entries))
	for _, e := range entries {
		out = append(out, Summary{ProviderID: e.ProviderID, AuthType: e.AuthType, Source: e.Source, Status: e.Status, Expiry: e.Expiry, Preview: previewSecret(e.Value)})
	}
	return out
}

func (l *Loader) ValidCount() int {
	n := 0
	l.entries.Range(func(key, value interface{}) bool {
		entry := value.(*Entry)
		if entry.Status == StatusValid {
			n++
		}
		return true
	})
	return n
}

func (l *Loader) SaveToken(provider, value string, expiry time.Time) error {
	if err := os.MkdirAll(l.authDir, 0755); err != nil {
		return err
	}
	entry := map[string]string{"value": value}
	if !expiry.IsZero() {
		entry["expiry"] = expiry.Format(time.RFC3339)
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(l.authDir, provider+".token"), data, 0600)
}

func (l *Loader) Delete(providerID, authType string) {
	key := keyFor(providerID, authType)
	l.entries.Delete(key)
}
func (l *Loader) Dir() string { return l.authDir }

func (l *Loader) upsertEntry(entry *Entry) {
	if entry == nil || entry.ProviderID == "" || entry.AuthType == "" {
		return
	}
	key := keyFor(entry.ProviderID, entry.AuthType)

	// Load existing entry
	existingRaw, ok := l.entries.Load(key)
	var existing *Entry
	if ok {
		existing = existingRaw.(*Entry)
	}

	// Check priority
	if existing != nil && comparePriority(existing, entry) >= 0 {
		return
	}

	// Store new entry
	copyEntry := *entry
	l.entries.Store(key, &copyEntry)
}

func comparePriority(existing, incoming *Entry) int {
	ep, ip := sourcePriority[existing.Source], sourcePriority[incoming.Source]
	if ep != ip {
		return ep - ip
	}
	if existing.Status != incoming.Status {
		if existing.Status == StatusValid {
			return 1
		}
		if incoming.Status == StatusValid {
			return -1
		}
	}
	if existing.LoadedAt.After(incoming.LoadedAt) {
		return 1
	}
	if existing.LoadedAt.Before(incoming.LoadedAt) {
		return -1
	}
	return 0
}

func resolveConfigEnv(v string) string {
	if strings.HasPrefix(v, "{env:") && strings.HasSuffix(v, "}") {
		return strings.TrimSpace(os.Getenv(v[5 : len(v)-1]))
	}
	if strings.HasPrefix(v, "{file:") && strings.HasSuffix(v, "}") {
		return config.ResolveString(v)
	}
	return strings.TrimSpace(v)
}

func keyFor(providerID, authType string) string { return providerID + ":" + authType }

func previewSecret(v string) string {
	if v == "" {
		return ""
	}

	// Check cache first
	if cached, ok := previewCache.Load(v); ok {
		return cached.(string)
	}

	var result string
	if len(v) <= 8 {
		result = strings.Repeat("*", len(v))
	} else {
		var sb strings.Builder
		sb.Grow(len(v))
		sb.WriteString(v[:3])
		for i := 0; i < len(v)-6; i++ {
			sb.WriteByte('*')
		}
		sb.WriteString(v[len(v)-3:])
		result = sb.String()
	}

	// Cache the result
	previewCache.Store(v, result)
	return result
}
