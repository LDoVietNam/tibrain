// Package config implements Ti's layered configuration system.
//
// Precedence (lowest to highest):
// 1. Defaults (built-in)
// 2. Global (~/.config/ti/config.json)
// 3. Project (./ti.json in CWD)
// 4. Portable (./.ti/config.json in CWD)
// 5. Custom (TI_CONFIG env → single file)
// 6. Environment variables
// 7. Inline flags (applied by callers at runtime)
package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DefaultModel     = "qwenvsclaude-3.6" // Qwen 3.6 - Free
	DefaultProvider  = "openrouter"
	DefaultHost      = "0.0.0.0"
	DefaultPort      = 1810
	DefaultRateLimit = 20
	DefaultTimeout   = 120 * time.Second
	DefaultDataDir   = "~/.ti/data"
	DefaultAuthDir   = "~/.ti/auth"
	DefaultLogLevel  = "info"
	DefaultMaxTokens = 8192
)

type ValidationIssue struct {
	Level   string `json:"level"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

type SourceSnapshot struct {
	Name    string `json:"name"`
	Path    string `json:"path,omitempty"`
	Found   bool   `json:"found"`
	Applied bool   `json:"applied"`
}

type LoadReport struct {
	Sources []SourceSnapshot  `json:"sources"`
	Hash    string            `json:"hash"`
	Issues  []ValidationIssue `json:"issues,omitempty"`
}

type CompatibleProviderConfig struct {
	Name            string            `json:"name"`
	Type            string            `json:"type,omitempty"`
	BaseURL         string            `json:"base_url"`
	ChatPath        string            `json:"chat_path,omitempty"`
	ModelsPath      string            `json:"models_path,omitempty"`
	ResponsesPath   string            `json:"responses_path,omitempty"`
	APIKey          string            `json:"api_key,omitempty"`
	APIKeyEnv       string            `json:"api_key_env,omitempty"`
	AuthHeader      string            `json:"auth_header,omitempty"`
	AuthTokenPrefix string            `json:"auth_token_prefix,omitempty"`
	CustomHeaders   map[string]string `json:"custom_headers,omitempty"`
	Models          []string          `json:"models,omitempty"`
	DefaultModel    string            `json:"default_model,omitempty"`
	CookieFile      string            `json:"cookie_file,omitempty"`
	CookieProfile   string            `json:"cookie_profile,omitempty"`
	CookieDomain    string            `json:"cookie_domain,omitempty"`
	ForceHTTPS      bool              `json:"force_https,omitempty"`
	TimeoutSec      int               `json:"timeout_sec,omitempty"`
}

type Config struct {
	DataDir    string `json:"data_dir,omitempty"`
	AuthDir    string `json:"auth_dir,omitempty"`
	LogLevel   string `json:"log_level,omitempty"`
	LogFile    string `json:"log_file,omitempty"`
	Port       int    `json:"port,omitempty"`
	Host       string `json:"host,omitempty"`
	RateLimit  int    `json:"rate_limit,omitempty"`
	MaxTokens  int    `json:"max_tokens,omitempty"`
	TimeoutSec int    `json:"timeout_sec,omitempty"`

	Provider  string            `json:"provider,omitempty"`
	Model     string            `json:"model,omitempty"`
	Models    []string          `json:"models,omitempty"`
	APIKeys   map[string]string `json:"api_keys,omitempty"`
	AuthFiles []string          `json:"auth_files,omitempty"`

	CookieFile        string `json:"cookie_file,omitempty"`
	CookieProfile     string `json:"cookie_profile,omitempty"`
	CookieDomain      string `json:"cookie_domain,omitempty"`
	CookieAPIPrimary  string `json:"cookie_api_primary,omitempty"`
	CookieAPIFallback string `json:"cookie_api_fallback,omitempty"`

	CompatibleProviders []CompatibleProviderConfig `json:"compatible_providers,omitempty"`

	DefaultPhase  string  `json:"default_phase,omitempty"`
	FallbackModel string  `json:"fallback_model,omitempty"`
	MaxBudgetUSD  float64 `json:"max_budget_usd,omitempty"`

	MemoryPath string `json:"memory_path,omitempty"`
	BEADSPath  string `json:"beads_path,omitempty"`
	QAPath     string `json:"qa_path,omitempty"`

	RouterAgentURL string `json:"router_agent_url,omitempty"`

	DefaultAgent string `json:"default_agent,omitempty"`
	SkillsDir    string `json:"skills_dir,omitempty"`

	mu      sync.RWMutex
	dataPtr atomic.Pointer[Config]
}

func Default() *Config {
	home, _ := os.UserHomeDir()
	defaultAuthFiles := []string{}

	if home != "" {
		defaultAuthFiles = []string{
			filepath.Join(home, ".ti", "auth", ".env"),
			filepath.Join(home, ".ti", "auth.env"),
		}
		if _, err := os.Stat("Z:\\00_SECRET\\.routerenv"); err == nil {
			defaultAuthFiles = append([]string{"Z:\\00_SECRET\\.routerenv"}, defaultAuthFiles...)
		}
	}

	cfg := &Config{
		DataDir:           DefaultDataDir,
		AuthDir:           DefaultAuthDir,
		LogLevel:          DefaultLogLevel,
		Port:              DefaultPort,
		Host:              DefaultHost,
		RateLimit:         DefaultRateLimit,
		MaxTokens:         DefaultMaxTokens,
		TimeoutSec:        int(DefaultTimeout.Seconds()),
		Provider:          DefaultProvider,
		Model:             DefaultModel,
		Models:            []string{DefaultModel},
		APIKeys:           map[string]string{},
		AuthFiles:         defaultAuthFiles,
		CookieFile:        "~/.ti/auth/cookies.json",
		CookieProfile:     "sharedchat",
		CookieDomain:      "chat.sharedchat.fun",
		CookieAPIPrimary:  "https://chat.sharedchat.cc",
		CookieAPIFallback: "https://chat.sharedchat.fun",
		DefaultPhase:      "implement",
		MemoryPath:        "~/.ti/data/memory.db",
		BEADSPath:         "~/.ti/data/beads.jsonl",
		QAPath:            "~/.ti/data/qa-pairs.db",
	}
	cfg.ensureDefaults()
	return cfg
}

func Load(path string) (*Config, error) {
	cfg := Default()
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("read config: %w", err)
			}
		} else if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}
	cfg.applyEnvOverrides()
	cfg.normalize()
	return cfg, nil
}

func LoadAuto() (*Config, error) {
	cfg, _, err := LoadAutoWithReport()
	return cfg, err
}

func LoadAutoWithReport() (*Config, *LoadReport, error) {
	cfg := Default()
	report := &LoadReport{Sources: []SourceSnapshot{{Name: "defaults", Found: true, Applied: true}}}
	home, _ := os.UserHomeDir()
	candidates := []SourceSnapshot{
		{Name: "global", Path: filepath.Join(home, ".config", "ti", "config.json")},
		{Name: "project", Path: "ti.json"},
		{Name: "portable", Path: filepath.Join(".ti", "config.json")},
		{Name: "custom", Path: os.Getenv("TI_CONFIG")},
	}
	for _, src := range candidates {
		snap := src
		if snap.Path == "" || !fileExists(snap.Path) {
			report.Sources = append(report.Sources, snap)
			continue
		}
		snap.Found = true
		if err := mergeFileInto(cfg, snap.Path); err != nil {
			return nil, report, fmt.Errorf("load %s: %w", snap.Name, err)
		}
		snap.Applied = true
		report.Sources = append(report.Sources, snap)
	}
	cfg.applyEnvOverrides()
	cfg.normalize()
	report.Hash = cfg.Hash()
	report.Issues = cfg.Validate()
	return cfg, report, nil
}

func mergeFileInto(dst *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var partial Config
	if err := json.Unmarshal(data, &partial); err != nil {
		return err
	}
	dst.MergeFrom(&partial)
	return nil
}

func (c *Config) MergeFrom(src *Config) {
	if src == nil {
		return
	}
	if src.DataDir != "" {
		c.DataDir = src.DataDir
	}
	if src.AuthDir != "" {
		c.AuthDir = src.AuthDir
	}
	if src.LogLevel != "" {
		c.LogLevel = src.LogLevel
	}
	if src.LogFile != "" {
		c.LogFile = src.LogFile
	}
	if src.Port != 0 {
		c.Port = src.Port
	}
	if src.Host != "" {
		c.Host = src.Host
	}
	if src.RateLimit != 0 {
		c.RateLimit = src.RateLimit
	}
	if src.MaxTokens != 0 {
		c.MaxTokens = src.MaxTokens
	}
	if src.TimeoutSec != 0 {
		c.TimeoutSec = src.TimeoutSec
	}
	if src.Provider != "" {
		c.Provider = src.Provider
	}
	if src.Model != "" {
		c.Model = src.Model
	}
	if len(src.Models) > 0 {
		c.Models = append([]string(nil), src.Models...)
	}
	if len(src.APIKeys) > 0 {
		if c.APIKeys == nil {
			c.APIKeys = map[string]string{}
		}
		for k, v := range src.APIKeys {
			if v != "" {
				c.APIKeys[k] = v
			}
		}
	}
	if src.CookieFile != "" {
		c.CookieFile = src.CookieFile
	}
	if src.CookieProfile != "" {
		c.CookieProfile = src.CookieProfile
	}
	if src.CookieDomain != "" {
		c.CookieDomain = src.CookieDomain
	}
	if src.CookieAPIPrimary != "" {
		c.CookieAPIPrimary = src.CookieAPIPrimary
	}
	if src.CookieAPIFallback != "" {
		c.CookieAPIFallback = src.CookieAPIFallback
	}
	if len(src.CompatibleProviders) > 0 {
		c.CompatibleProviders = append([]CompatibleProviderConfig(nil), src.CompatibleProviders...)
	}
	if src.DefaultPhase != "" {
		c.DefaultPhase = src.DefaultPhase
	}
	if src.FallbackModel != "" {
		c.FallbackModel = src.FallbackModel
	}
	if src.MaxBudgetUSD != 0 {
		c.MaxBudgetUSD = src.MaxBudgetUSD
	}
	if src.MemoryPath != "" {
		c.MemoryPath = src.MemoryPath
	}
	if src.BEADSPath != "" {
		c.BEADSPath = src.BEADSPath
	}
	if src.QAPath != "" {
		c.QAPath = src.QAPath
	}
	if src.DefaultAgent != "" {
		c.DefaultAgent = src.DefaultAgent
	}
	if src.SkillsDir != "" {
		c.SkillsDir = src.SkillsDir
	}
}

func (c *Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	data, err := json.MarshalIndent(c.Public(), "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

func (c *Config) Public() *Config {
	apiKeys := make(map[string]string, len(c.APIKeys))
	for k, v := range c.APIKeys {
		apiKeys[k] = v
	}
	return &Config{
		DataDir:             c.DataDir,
		AuthDir:             c.AuthDir,
		LogLevel:            c.LogLevel,
		LogFile:             c.LogFile,
		Port:                c.Port,
		Host:                c.Host,
		RateLimit:           c.RateLimit,
		MaxTokens:           c.MaxTokens,
		TimeoutSec:          c.TimeoutSec,
		Provider:            c.Provider,
		Model:               c.Model,
		Models:              append([]string(nil), c.Models...),
		APIKeys:             apiKeys,
		AuthFiles:           append([]string(nil), c.AuthFiles...),
		CookieFile:          c.CookieFile,
		CookieProfile:       c.CookieProfile,
		CookieDomain:        c.CookieDomain,
		CookieAPIPrimary:    c.CookieAPIPrimary,
		CookieAPIFallback:   c.CookieAPIFallback,
		CompatibleProviders: cloneCompatibleProviders(c.CompatibleProviders),
		DefaultPhase:        c.DefaultPhase,
		FallbackModel:       c.FallbackModel,
		MaxBudgetUSD:        c.MaxBudgetUSD,
		MemoryPath:          c.MemoryPath,
		BEADSPath:           c.BEADSPath,
		QAPath:              c.QAPath,
		RouterAgentURL:      c.RouterAgentURL,
		DefaultAgent:        c.DefaultAgent,
		SkillsDir:           c.SkillsDir,
	}
}

func (c *Config) Redacted() *Config {
	copyCfg := c.Public()
	for k, v := range copyCfg.APIKeys {
		copyCfg.APIKeys[k] = redactSecret(v)
	}
	for i := range copyCfg.CompatibleProviders {
		copyCfg.CompatibleProviders[i].APIKey = redactSecret(copyCfg.CompatibleProviders[i].APIKey)
	}
	return copyCfg
}

func (c *Config) GetSnapshot() *Config {
	ptr := c.dataPtr.Load()
	if ptr != nil {
		return ptr.Public()
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Public()
}

func (c *Config) ReplaceFrom(src *Config) {
	newCfg := src.Public()
	newCfg.normalize()
	c.dataPtr.Store(newCfg)
}

func (c *Config) Hash() string {
	data, _ := json.Marshal(c.Public())
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func (c *Config) Timeout() time.Duration { return time.Duration(c.TimeoutSec) * time.Second }

func (c *Config) Validate() []ValidationIssue {
	var issues []ValidationIssue
	add := func(level, field, message string) {
		issues = append(issues, ValidationIssue{Level: level, Field: field, Message: message})
	}
	if c.Port < 0 || c.Port > 65535 {
		add("error", "port", "must be between 0 and 65535")
	}
	if c.TimeoutSec <= 0 {
		add("warning", "timeout_sec", "should be greater than zero")
	}
	if c.MaxTokens <= 0 {
		add("warning", "max_tokens", "should be greater than zero")
	}
	if c.RateLimit < 0 {
		add("warning", "rate_limit", "should not be negative")
	}
	if c.Provider == "sharedchat" && c.CookieFile == "" {
		add("warning", "cookie_file", "sharedchat provider works best with a browser cookie file or saved cookie state")
	}
	for i, p := range c.CompatibleProviders {
		field := fmt.Sprintf("compatible_providers[%d]", i)
		if strings.TrimSpace(p.Name) == "" {
			add("error", field+".name", "provider name is required")
		}
		if strings.TrimSpace(p.BaseURL) == "" {
			add("error", field+".base_url", "base_url is required")
		}
		if strings.EqualFold(p.Type, "cookie_http") && strings.TrimSpace(p.CookieFile) == "" {
			add("warning", field+".cookie_file", "cookie_http providers need a cookie_file")
		}
	}
	return issues
}

func (c *Config) ApplyEnvSubstitution() { c.applyEnvSubstitution() }

func ExpandHome(path string) string {
	if path == "" || path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if len(path) > 1 && path[1] == '/' {
		return home + path[1:]
	}
	return home
}

func ContractHome(path string) string {
	if path == "" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

func (c *Config) applyEnvOverrides() {
	envStr := func(key string, dst *string) {
		if v := os.Getenv(key); v != "" {
			*dst = v
		}
	}
	envStr("TI_DATA_DIR", &c.DataDir)
	envStr("TI_AUTH_DIR", &c.AuthDir)
	envStr("TI_LOG_LEVEL", &c.LogLevel)
	envStr("TI_LOG_FILE", &c.LogFile)
	envStr("TI_HOST", &c.Host)
	envStr("TI_PROVIDER", &c.Provider)
	envStr("TI_MODEL", &c.Model)
	envStr("TI_COOKIE_FILE", &c.CookieFile)
	envStr("TI_COOKIE_PROFILE", &c.CookieProfile)
	envStr("TI_COOKIE_DOMAIN", &c.CookieDomain)
	envStr("TI_COOKIE_API_PRIMARY", &c.CookieAPIPrimary)
	envStr("TI_COOKIE_API_FALLBACK", &c.CookieAPIFallback)
	envStr("TI_DEFAULT_PHASE", &c.DefaultPhase)
	envStr("TI_DEFAULT_AGENT", &c.DefaultAgent)
	envStr("TI_MEMORY_PATH", &c.MemoryPath)
	envStr("TI_BEADS_PATH", &c.BEADSPath)
	envStr("TI_QA_PATH", &c.QAPath)
	envStr("TI_FALLBACK_MODEL", &c.FallbackModel)
	envStr("TI_ROUTER_AGENT_URL", &c.RouterAgentURL)
	if v := os.Getenv("TI_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &c.Port)
	}
	if v := os.Getenv("TI_RATE_LIMIT"); v != "" {
		fmt.Sscanf(v, "%d", &c.RateLimit)
	}
	if v := os.Getenv("TI_MAX_TOKENS"); v != "" {
		fmt.Sscanf(v, "%d", &c.MaxTokens)
	}
	if v := os.Getenv("TI_TIMEOUT"); v != "" {
		fmt.Sscanf(v, "%d", &c.TimeoutSec)
	}
	if v := os.Getenv("TI_MAX_BUDGET"); v != "" {
		fmt.Sscanf(v, "%f", &c.MaxBudgetUSD)
	}
	if c.APIKeys == nil {
		c.APIKeys = map[string]string{}
	}
	if v := os.Getenv("TI_API_KEY"); v != "" {
		provider := c.Provider
		if provider == "" {
			provider = DefaultProvider
		}
		c.APIKeys[provider] = v
	}
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		c.APIKeys["anthropic"] = v
	}
	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		c.APIKeys["openai"] = v
	}
	if v := os.Getenv("GOOGLE_API_KEY"); v != "" {
		c.APIKeys["google"] = v
	}
	if v := os.Getenv("GROQ_API_KEY"); v != "" {
		c.APIKeys["groq"] = v
	}
	if v := os.Getenv("OPENROUTER_API_KEY"); v != "" {
		c.APIKeys["openrouter"] = v
	}
	if v := os.Getenv("CLOUDFLARE_API_KEY"); v != "" {
		c.APIKeys["cloudflare"] = v
	}
}

func (c *Config) ensureDefaults() {
	if c.Port == 0 {
		c.Port = DefaultPort
	}
	if c.Host == "" {
		c.Host = DefaultHost
	}
	if c.LogLevel == "" {
		c.LogLevel = DefaultLogLevel
	}
	if c.Provider == "" {
		c.Provider = DefaultProvider
	}
	if c.Model == "" {
		c.Model = DefaultModel
	}
	if c.RateLimit == 0 {
		c.RateLimit = DefaultRateLimit
	}
	if c.MaxTokens == 0 {
		c.MaxTokens = DefaultMaxTokens
	}
	if c.TimeoutSec == 0 {
		c.TimeoutSec = int(DefaultTimeout.Seconds())
	}
	if c.DataDir == "" {
		c.DataDir = DefaultDataDir
	}
	if c.AuthDir == "" {
		c.AuthDir = DefaultAuthDir
	}
	if c.MemoryPath == "" {
		c.MemoryPath = "~/.ti/data/memory.db"
	}
	if c.BEADSPath == "" {
		c.BEADSPath = "~/.ti/data/beads.jsonl"
	}
	if c.QAPath == "" {
		c.QAPath = "~/.ti/data/qa-pairs.db"
	}
	if c.RouterAgentURL == "" {
		c.RouterAgentURL = "http://localhost:1808"
	}
	if c.DefaultPhase == "" {
		c.DefaultPhase = "implement"
	}
	if c.APIKeys == nil {
		c.APIKeys = map[string]string{}
	}
	if c.CookieFile == "" {
		c.CookieFile = "~/.ti/auth/cookies.json"
	}
	if c.CookieProfile == "" {
		c.CookieProfile = "sharedchat"
	}
	if c.CookieDomain == "" {
		c.CookieDomain = "chat.sharedchat.fun"
	}
	if c.CookieAPIPrimary == "" {
		c.CookieAPIPrimary = "https://chat.sharedchat.cc"
	}
	if c.CookieAPIFallback == "" {
		c.CookieAPIFallback = "https://chat.sharedchat.fun"
	}
	if len(c.Models) == 0 && c.Model != "" {
		c.Models = []string{c.Model}
	}
}

func (c *Config) normalize() {
	c.applyEnvSubstitution()
	c.applyPaths()
	c.ensureDefaults()
	seen := map[string]bool{}
	var models []string
	for _, model := range append([]string{c.Model}, c.Models...) {
		model = strings.TrimSpace(model)
		if model == "" || seen[model] {
			continue
		}
		seen[model] = true
		models = append(models, model)
	}
	if len(models) == 0 {
		models = []string{DefaultModel}
	}
	c.Model = models[0]
	if len(models) > 1 {
		sort.Strings(models[1:])
	}
	c.Models = models
}

func (c *Config) applyPaths() {
	c.DataDir = ExpandHome(c.DataDir)
	c.AuthDir = ExpandHome(c.AuthDir)
	c.LogFile = ExpandHome(c.LogFile)
	c.CookieFile = ExpandHome(c.CookieFile)
	for i := range c.CompatibleProviders {
		c.CompatibleProviders[i].CookieFile = ExpandHome(c.CompatibleProviders[i].CookieFile)
	}
	c.MemoryPath = ExpandHome(c.MemoryPath)
	c.BEADSPath = ExpandHome(c.BEADSPath)
	c.QAPath = ExpandHome(c.QAPath)
}

func (c *Config) applyEnvSubstitution() {
	replace := func(s string) string {
		for {
			start := strings.Index(s, "{env:")
			if start == -1 {
				return s
			}
			end := strings.Index(s[start:], "}")
			if end == -1 {
				return s
			}
			end += start
			varName := s[start+5 : end]
			s = s[:start] + os.Getenv(varName) + s[end+1:]
		}
	}
	c.DataDir = replace(c.DataDir)
	c.AuthDir = replace(c.AuthDir)
	c.LogFile = replace(c.LogFile)
	c.CookieFile = replace(c.CookieFile)
	c.CookieProfile = replace(c.CookieProfile)
	c.CookieDomain = replace(c.CookieDomain)
	c.CookieAPIPrimary = replace(c.CookieAPIPrimary)
	c.CookieAPIFallback = replace(c.CookieAPIFallback)
	for i := range c.CompatibleProviders {
		cp := &c.CompatibleProviders[i]
		cp.BaseURL = replace(cp.BaseURL)
		cp.ChatPath = replace(cp.ChatPath)
		cp.ModelsPath = replace(cp.ModelsPath)
		cp.ResponsesPath = replace(cp.ResponsesPath)
		cp.APIKey = replace(cp.APIKey)
		cp.APIKeyEnv = replace(cp.APIKeyEnv)
		cp.CookieFile = replace(cp.CookieFile)
		cp.CookieProfile = replace(cp.CookieProfile)
		cp.CookieDomain = replace(cp.CookieDomain)
		cp.DefaultModel = replace(cp.DefaultModel)
		for k, v := range cp.CustomHeaders {
			cp.CustomHeaders[k] = replace(v)
		}
	}
	c.MemoryPath = replace(c.MemoryPath)
	c.BEADSPath = replace(c.BEADSPath)
	c.QAPath = replace(c.QAPath)
	c.FallbackModel = replace(c.FallbackModel)
	for k, v := range c.APIKeys {
		c.APIKeys[k] = replace(v)
	}
}

func redactSecret(v string) string {
	if v == "" {
		return ""
	}
	if len(v) <= 8 {
		return strings.Repeat("*", len(v))
	}
	return v[:4] + strings.Repeat("*", len(v)-8) + v[len(v)-4:]
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

var (
	envVarRegex  = regexp.MustCompile(`\{env:([A-Za-z_][A-Za-z0-9_]*)\}`)
	fileVarRegex = regexp.MustCompile(`\{file:([^\}]+)\}`)
	envPrefix    = "{env:"
	filePrefix   = "{file:"
)

func ResolveString(s string) string {
	if !strings.Contains(s, envPrefix) && !strings.Contains(s, filePrefix) {
		return s
	}

	s = envVarRegex.ReplaceAllStringFunc(s, func(match string) string {
		key := match[5 : len(match)-1]
		return os.Getenv(key)
	})

	s = fileVarRegex.ReplaceAllStringFunc(s, func(match string) string {
		path := match[6 : len(match)-1]
		data, err := os.ReadFile(path)
		if err != nil {
			return match
		}
		content := strings.TrimSpace(string(data))
		if strings.HasPrefix(content, "{") {
			if v := parseJSONValue(content); v != "" {
				return v
			}
		}
		if strings.Contains(content, "=") {
			parsed := ParseEnvFile(content)
			for _, v := range parsed {
				if v != "" {
					return v
				}
			}
		}
		return content
	})

	return s
}

func parseJSONValue(content string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return ""
	}
	for _, v := range data {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

func ResolveMap(m map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = ResolveString(v)
	}
	return result
}

func ParseEnvFile(content string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			result[key] = value
		}
	}
	return result
}

func cloneCompatibleProviders(in []CompatibleProviderConfig) []CompatibleProviderConfig {
	out := make([]CompatibleProviderConfig, len(in))
	for i, p := range in {
		out[i] = p
		out[i].Models = append([]string(nil), p.Models...)
		if p.CustomHeaders != nil {
			out[i].CustomHeaders = make(map[string]string, len(p.CustomHeaders))
			for k, v := range p.CustomHeaders {
				out[i].CustomHeaders[k] = v
			}
		}
	}
	return out
}
