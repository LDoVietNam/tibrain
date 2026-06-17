package management

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/ti/cli/internal/auth"
	"github.com/ti/cli/internal/config"
	"github.com/ti/cli/internal/providers"
)

const version = "ti-management-sync-0.1.0"

type Service struct {
	cfg       *config.Config
	authPaths AuthPaths
	registry  *providers.Registry
}

func NewService(cfg *config.Config) *Service {
	if cfg == nil {
		cfg = config.Default()
	}
	return &Service{cfg: cfg, authPaths: deriveAuthPaths(cfg)}
}

func deriveAuthPaths(cfg *config.Config) AuthPaths {
	root := strings.TrimSpace(cfg.AuthDir)
	if root == "" || root == config.DefaultAuthDir {
		if runtime.GOOS == "windows" {
			root = buildWindowsPath("Z:", "06_AUTH")
		} else if home, err := os.UserHomeDir(); err == nil {
			root = filepath.Join(home, ".ti", "auth-root")
		} else {
			root = ".ti/auth-root"
		}
	}
	return AuthPaths{
		Root:    root,
		Cookies: joinPath(root, "cookies"),
		Auth:    joinPath(root, "auth"),
		OAuth:   joinPath(root, "auth", "oauth"),
	}
}

func buildWindowsPath(segments ...string) string { return strings.Join(segments, `\`) }

func joinPath(base string, parts ...string) string {
	if strings.Contains(base, `\`) && !strings.Contains(base, "/") {
		items := append([]string{base}, parts...)
		return strings.Join(items, `\`)
	}
	items := append([]string{base}, parts...)
	return filepath.Join(items...)
}

func (s *Service) Dashboard() Dashboard {
	providers := s.Providers()
	authFiles := s.AuthFiles()
	cookies := s.Cookies()
	oauthRows := s.OAuthConnections()
	apiKeys := s.APIKeys()
	return Dashboard{
		Providers:      len(providers),
		AuthFiles:      len(authFiles),
		CookieJars:     len(cookies),
		OAuthLanes:     len(oauthRows),
		APIKeys:        len(apiKeys),
		AuthPathsValid: s.validateAuthPaths(),
		AuthRoot:       s.authPaths.Root,
	}
}

func (s *Service) validateAuthPaths() bool {
	ap := s.authPaths
	return ap.Cookies == joinPath(ap.Root, "cookies") &&
		ap.Auth == joinPath(ap.Root, "auth") &&
		ap.OAuth == joinPath(ap.Root, "auth", "oauth")
}

func (s *Service) SetRegistry(r *providers.Registry) {
	s.registry = r
}

func (s *Service) Providers() []ProviderRow {
	// Start with hardcoded baseline (external lanes / accounts)
	rows := []ProviderRow{
		{ID: "claude-code", Name: "Claude Code", Tier: "subscription", Auth: "oauth", Protocol: "device-flow", Group: "prod", Enabled: true},
		{ID: "gemini-cli", Name: "Gemini CLI", Tier: "free", Auth: "oauth", Protocol: "browser-flow", Group: "prod", Enabled: true},
		{ID: "github-copilot", Name: "GitHub Copilot", Tier: "subscription", Auth: "oauth", Protocol: "device-flow", Group: "prod", Enabled: true},
		{ID: "opencode-connect", Name: "OpenCode-style Connect", Tier: "gateway", Auth: "oauth", Protocol: "provider-connect", Group: "prod", Enabled: true},
		{ID: "chatgpt-account", Name: "ChatGPT Account", Tier: "subscription", Auth: "oauth", Protocol: "account-connect", Group: "lab", Enabled: true},
		{ID: "codex-account", Name: "Codex / OpenAI Account", Tier: "subscription", Auth: "oauth", Protocol: "account-connect", Group: "lab", Enabled: true},
		{ID: "cursor-account", Name: "Cursor Account", Tier: "agent", Auth: "oauth", Protocol: "adapter", Group: "lab", Enabled: true},
		{ID: "deepseek-connect", Name: "DeepSeek Connect", Tier: "cheap", Auth: "oauth", Protocol: "provider-connect", Group: "backup", Enabled: true},
		{ID: "cloudflare-connect", Name: "Cloudflare Connect", Tier: "gateway", Auth: "oauth", Protocol: "provider-connect", Group: "dev", Enabled: true},
		{ID: "openrouter-connect", Name: "OpenRouter Connect", Tier: "gateway", Auth: "oauth", Protocol: "provider-connect", Group: "backup", Enabled: true},
		{ID: "deepgram-account", Name: "Deepgram Account", Tier: "voice", Auth: "oauth", Protocol: "provider-connect", Group: "lab", Enabled: true},
		{ID: "qwen-account", Name: "Qwen Account", Tier: "free", Auth: "oauth", Protocol: "provider-connect", Group: "backup", Enabled: true},
		{ID: "kiro-account", Name: "Kiro Account", Tier: "free", Auth: "oauth", Protocol: "provider-connect", Group: "backup", Enabled: true},
		{ID: "qwen-cookie-import", Name: "Qwen Cookie Import", Tier: "free", Auth: "cookie", Protocol: "cookie-import", Group: "backup", Enabled: true},
		{ID: "kiro-cookie-import", Name: "Kiro Cookie Import", Tier: "free", Auth: "cookie", Protocol: "cookie-import", Group: "backup", Enabled: true},
		{ID: "custom-cookie-import", Name: "Custom Domain Cookie Import", Tier: "gateway", Auth: "cookie", Protocol: "cookie-import", Group: "dev", Enabled: true},
	}

	// Merge runtime-registered providers from registry
	if s.registry != nil {
		snapshots := s.registry.Snapshots()
		seen := map[string]bool{}
		for _, r := range rows {
			seen[strings.ToLower(r.ID)] = true
		}
		for _, snap := range snapshots {
			id := strings.ToLower(snap.Name)
			if seen[id] {
				continue
			}
			seen[id] = true
			rows = append(rows, ProviderRow{
				ID:       id,
				Name:     snap.Name,
				Tier:     "runtime",
				Auth:     "api_key",
				Protocol: "provider",
				Group:    "prod",
				Enabled:  snap.Healthy,
			})
		}
	}

	return rows
}

func (s *Service) APIKeys() []APIKeyRow {
	loader := auth.NewLoader(s.authPaths.Auth)
	loader.SetAuthFiles(s.cfg.AuthFiles)
	_ = loader.LoadAll()
	summaries := loader.Summaries()
	rows := make([]APIKeyRow, 0, len(summaries))
	for _, item := range summaries {
		if item.AuthType != "api_key" && item.AuthType != "token" {
			continue
		}
		rows = append(rows, APIKeyRow{Name: item.ProviderID, Prefix: item.Preview, Scope: item.AuthType, Status: item.Status})
	}
	return rows
}

func (s *Service) AuthFiles() []AuthFileRow {
	rows := make([]AuthFileRow, 0)
	rows = append(rows, s.scanAuthDir(s.authPaths.Cookies, map[string]string{".json": "cookie", ".cookie": "cookie"})...)
	rows = append(rows, s.scanAuthDir(s.authPaths.Auth, map[string]string{".oauth": "oauth", ".json": "auth", ".token": "token", ".cookie": "cookie"})...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].File < rows[j].File })
	return rows
}

func (s *Service) Cookies() []CookieRow {
	files := s.scanFiles(s.authPaths.Cookies, map[string]string{".json": "cookie", ".cookie": "cookie"})
	rows := make([]CookieRow, 0, len(files))
	for _, item := range files {
		rows = append(rows, CookieRow{File: item.File, Provider: item.Provider, Status: item.Status, Updated: item.Updated, Path: item.Path})
	}
	return rows
}

func (s *Service) OAuthConnections() []OAuthConnectionRow {
	entries := s.scanFiles(s.authPaths.OAuth, map[string]string{".oauth": "oauth", ".json": "oauth"})
	rows := make([]OAuthConnectionRow, 0, len(entries))
	for _, entry := range entries {
		state := "pending"
		account := "—"
		if payload, err := os.ReadFile(entry.Path); err == nil {
			var raw map[string]any
			if json.Unmarshal(payload, &raw) == nil {
				if v, ok := raw["state"].(string); ok && strings.TrimSpace(v) != "" {
					state = strings.TrimSpace(v)
				}
				for _, key := range []string{"account", "email", "user", "subject"} {
					if v, ok := raw[key].(string); ok && strings.TrimSpace(v) != "" {
						account = strings.TrimSpace(v)
						break
					}
				}
			} else {
				state = "connected"
			}
		}
		rows = append(rows, OAuthConnectionRow{Name: friendlyProviderName(entry.Provider), State: state, Account: account, Updated: entry.Updated, Path: entry.Path})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	return rows
}

func (s *Service) OAuthStarters() []OAuthStarter {
	return []OAuthStarter{
		{ID: "claude-code", Name: "Claude Code", Flow: "OAuth / device flow", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "Anthropic / Claude Code lane", Note: "Khởi tạo device flow rồi poll trạng thái cho Claude Code.", Status: "available"},
		{ID: "gemini-cli", Name: "Gemini CLI", Flow: "OAuth / browser flow", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "Google / Gemini CLI lane", Note: "Tạo browser flow riêng cho Gemini CLI theo model CLI/account-based.", Status: "available"},
		{ID: "github-copilot", Name: "GitHub Copilot", Flow: "OAuth / device flow", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "Copilot account lane", Note: "Phù hợp với pattern đăng nhập thiết bị/headless cho GitHub account.", Status: "available"},
		{ID: "gitlab", Name: "GitLab", Flow: "OAuth / device flow", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "GitLab account lane", Note: "GitLab OAuth token cho CI/CD và API access.", Status: "available"},
		{ID: "cloudflare", Name: "Cloudflare", Flow: "API token", AuthType: "api_key", Artifact: "token", Storage: s.authPaths.Auth, ProviderScope: "Cloudflare account lane", Note: "Cloudflare API token cho tunnel và API access.", Status: "available"},
		{ID: "antigravity", Name: "Antigravity (Google)", Flow: "OAuth / Device Flow", AuthType: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "Google Cloud / Antigravity IDE", Note: "Enables Gemini 3.1 Pro, Claude on Antigravity via OAuth", Status: "available"},
		{ID: "opencode-connect", Name: "OpenCode-style Connect", Flow: "Connect provider", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "Multi-provider connect surface", Note: "Học pattern của OpenCode: một bề mặt connect provider chung cho nhiều upstream/accounts.", Status: "available"},
		{ID: "iflow-cookie-import", Name: "iFlow Cookie Import", Flow: "Cookie import", AuthType: "cookie", Artifact: "cookie", Storage: s.authPaths.Cookies, ProviderScope: "Free / cookie-backed lane", Note: "Import cookie jar riêng, không gộp vào OAuth artifacts và không hiển thị như OAuth connection thật.", Status: "available"},
		{ID: "codex-account", Name: "Codex / OpenAI Account", Flow: "Account connect", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "OpenAI account-backed lane", Note: "Bề mặt connect cho Codex/OpenAI account lane.", Status: "planned"},
		{ID: "chatgpt-account", Name: "ChatGPT Account", Flow: "Account connect", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "Account-backed lane", Note: "Dành cho các lane dùng account connect thay vì API key thô.", Status: "planned"},
		{ID: "cursor-account", Name: "Cursor Account", Flow: "Provider connect", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "IDE/account-backed lane", Note: "Chuẩn bị chỗ cho lane Cursor kiểu account connect hoặc adapter.", Status: "planned"},
		{ID: "windsurf-account", Name: "Windsurf Account", Flow: "Provider connect", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "IDE/account-backed lane", Note: "Lane Windsurf cho account-backed connect surface.", Status: "planned"},
		{ID: "trae-account", Name: "TRAE Account", Flow: "Provider connect", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "IDE/account-backed lane", Note: "Lane TRAE theo hướng provider connect hoặc adapter.", Status: "planned"},
		{ID: "deepseek-connect", Name: "DeepSeek Connect", Flow: "Provider connect", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "DeepSeek account lane", Note: "Chuẩn bị bề mặt connect nhiều lane cho DeepSeek/account-backed providers.", Status: "planned"},
		{ID: "cloudflare-connect", Name: "Cloudflare Connect", Flow: "Provider connect", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "Gateway/account lane", Note: "Bề mặt connect cho gateway/account-style integrations.", Status: "planned"},
		{ID: "openrouter-connect", Name: "OpenRouter Connect", Flow: "Provider connect", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "Aggregator/account lane", Note: "Dành cho lane aggregator/provider-connect thay vì nhét chung với API key settings.", Status: "planned"},
		{ID: "deepgram-account", Name: "Deepgram Account", Flow: "Provider connect", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "Voice/account-backed lane", Note: "Giữ chỗ cho voice provider lane dạng account connect.", Status: "planned"},
		{ID: "assemblyai-account", Name: "AssemblyAI Account", Flow: "Provider connect", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "Voice/account-backed lane", Note: "Bề mặt connect cho AssemblyAI/account-style integrations.", Status: "planned"},
		{ID: "qwen-account", Name: "Qwen Account", Flow: "Provider connect", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "Free/account-backed lane", Note: "Tách riêng Qwen account lane khỏi cookie import.", Status: "planned"},
		{ID: "kiro-account", Name: "Kiro Account", Flow: "Provider connect", AuthType: "oauth", Artifact: "oauth", Storage: s.authPaths.OAuth, ProviderScope: "Free/account-backed lane", Note: "Tách riêng Kiro account lane khỏi cookie import.", Status: "planned"},
	}
}

func (s *Service) CookieStarters() []CookieStarter {
	return []CookieStarter{
		{ID: "iflow-cookie-import", Name: "iFlow Cookie Import", Source: "JSON cookie jar", Storage: s.authPaths.Cookies, ProviderScope: "Free / cookie-backed lane", Note: "Import cookie jar riêng cho iFlow. Không gộp vào OAuth artifacts.", Status: "available"},
		{ID: "qwen-cookie-import", Name: "Qwen Cookie Import", Source: "Browser-export cookie jar", Storage: s.authPaths.Cookies, ProviderScope: "Free / web-backed lane", Note: "Dùng cho các lane Qwen/web account cần cookie jar riêng.", Status: "planned"},
		{ID: "kiro-cookie-import", Name: "Kiro Cookie Import", Source: "Browser-export cookie jar", Storage: s.authPaths.Cookies, ProviderScope: "Free / cookie-backed lane", Note: "Cookie import riêng cho Kiro, tách khỏi flow OAuth.", Status: "planned"},
		{ID: "custom-cookie-import", Name: "Custom Domain Cookie Import", Source: "Generic cookie jar", Storage: s.authPaths.Cookies, ProviderScope: "Custom web lane", Note: "Import cookie jar theo domain cho các provider/web lane tự định nghĩa.", Status: "available"},
	}
}

func (s *Service) ConfigView() ConfigView {
	return ConfigView{Host: s.cfg.Host, Port: s.cfg.Port, RateLimit: s.cfg.RateLimit, Provider: s.cfg.Provider, Model: s.cfg.Model, AuthPaths: s.authPaths, RoutingStrategy: "policy-learning"}
}

func (s *Service) System() SystemInfo {
	now := time.Now().UTC()
	return SystemInfo{Version: version, BuildDate: now.Format(time.RFC3339), Now: now, ModelPath: "/v1/models"}
}

func (s *Service) Bootstrap() Bootstrap {
	return Bootstrap{Dashboard: s.Dashboard(), AuthPaths: s.authPaths, Providers: s.Providers(), APIKeys: s.APIKeys(), AuthFiles: s.AuthFiles(), Cookies: s.Cookies(), OAuth: s.OAuthConnections(), OAuthStarters: s.OAuthStarters(), CookieStarters: s.CookieStarters(), Config: s.ConfigView(), System: s.System()}
}

type fileRow struct {
	File     string
	Provider string
	Status   string
	Updated  string
	Kind     string
	Path     string
}

func (s *Service) scanAuthDir(root string, allowed map[string]string) []AuthFileRow {
	files := s.scanFiles(root, allowed)
	out := make([]AuthFileRow, 0, len(files))
	for _, item := range files {
		out = append(out, AuthFileRow{File: item.File, Provider: item.Provider, Status: item.Status, Updated: item.Updated, Kind: item.Kind, Path: item.Path})
	}
	return out
}

func (s *Service) scanFiles(root string, allowed map[string]string) []fileRow {
	rows := make([]fileRow, 0)
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return rows
	}
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(info.Name()))
		kind, ok := allowed[ext]
		if !ok {
			return nil
		}
		provider := strings.TrimSuffix(info.Name(), ext)
		status := "present"
		if kind == "oauth" {
			status = "connected"
		}
		rows = append(rows, fileRow{File: info.Name(), Provider: friendlyProviderName(provider), Status: status, Updated: info.ModTime().UTC().Format(time.RFC3339), Kind: kind, Path: path})
		return nil
	})
	sort.Slice(rows, func(i, j int) bool { return rows[i].File < rows[j].File })
	return rows
}

// --- Write methods ---

// AddAPIKey lưu API key vào token file trong auth directory.
func (s *Service) AddAPIKey(req AddAPIKeyRequest) WriteResult {
	if strings.TrimSpace(req.Provider) == "" || strings.TrimSpace(req.Key) == "" {
		return WriteResult{OK: false, Msg: "provider và key không được để trống"}
	}
	dir := s.authPaths.Auth
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return WriteResult{OK: false, Msg: fmt.Sprintf("không tạo được thư mục auth: %v", err)}
	}
	filename := strings.ToLower(strings.TrimSpace(req.Provider)) + ".token"
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(strings.TrimSpace(req.Key)), 0o600); err != nil {
		return WriteResult{OK: false, Msg: fmt.Sprintf("không ghi được file: %v", err)}
	}
	return WriteResult{OK: true, Path: path, Msg: "đã lưu API key"}
}

// AddAuthFile lưu auth file (token/oauth/cookie) vào auth directory.
func (s *Service) AddAuthFile(req AddAuthFileRequest) WriteResult {
	if strings.TrimSpace(req.Provider) == "" || strings.TrimSpace(req.Content) == "" {
		return WriteResult{OK: false, Msg: "provider và content không được để trống"}
	}
	kind := strings.ToLower(strings.TrimSpace(req.Kind))
	if kind == "" {
		kind = "token"
	}
	extMap := map[string]string{"token": ".token", "oauth": ".oauth", "cookie": ".cookie", "auth": ".json"}
	ext, ok := extMap[kind]
	if !ok {
		ext = ".token"
	}
	var dir string
	if kind == "oauth" {
		dir = s.authPaths.OAuth
	} else {
		dir = s.authPaths.Auth
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return WriteResult{OK: false, Msg: fmt.Sprintf("không tạo được thư mục: %v", err)}
	}
	filename := strings.ToLower(strings.TrimSpace(req.Provider)) + ext
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(req.Content), 0o600); err != nil {
		return WriteResult{OK: false, Msg: fmt.Sprintf("không ghi được file: %v", err)}
	}
	return WriteResult{OK: true, Path: path, Msg: "đã lưu auth file"}
}

// AddCookie lưu cookie jar JSON vào cookies directory.
func (s *Service) AddCookie(req AddCookieRequest) WriteResult {
	if strings.TrimSpace(req.Provider) == "" || strings.TrimSpace(req.Content) == "" {
		return WriteResult{OK: false, Msg: "provider và content không được để trống"}
	}
	dir := s.authPaths.Cookies
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return WriteResult{OK: false, Msg: fmt.Sprintf("không tạo được thư mục cookies: %v", err)}
	}
	filename := strings.ToLower(strings.TrimSpace(req.Provider)) + ".json"
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(req.Content), 0o600); err != nil {
		return WriteResult{OK: false, Msg: fmt.Sprintf("không ghi được file: %v", err)}
	}
	return WriteResult{OK: true, Path: path, Msg: "đã lưu cookie"}
}

// AddOAuth lưu OAuth connection JSON vào oauth directory.
func (s *Service) AddOAuth(req AddOAuthRequest) WriteResult {
	if strings.TrimSpace(req.Provider) == "" || len(req.Data) == 0 {
		return WriteResult{OK: false, Msg: "provider và data không được để trống"}
	}
	dir := s.authPaths.OAuth
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return WriteResult{OK: false, Msg: fmt.Sprintf("không tạo được thư mục oauth: %v", err)}
	}
	filename := strings.ToLower(strings.TrimSpace(req.Provider)) + ".oauth"
	path := filepath.Join(dir, filename)
	b, err := json.MarshalIndent(req.Data, "", "  ")
	if err != nil {
		return WriteResult{OK: false, Msg: fmt.Sprintf("không marshal được data: %v", err)}
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return WriteResult{OK: false, Msg: fmt.Sprintf("không ghi được file: %v", err)}
	}
	return WriteResult{OK: true, Path: path, Msg: "đã lưu OAuth connection"}
}

func friendlyProviderName(v string) string {
	v = strings.TrimSpace(strings.NewReplacer("-", " ", "_", " ").Replace(v))
	if v == "" {
		return "Unknown"
	}
	parts := strings.Fields(v)
	for i, p := range parts {
		if len(p) > 1 {
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		} else {
			parts[i] = strings.ToUpper(p)
		}
	}
	return strings.Join(parts, " ")
}
