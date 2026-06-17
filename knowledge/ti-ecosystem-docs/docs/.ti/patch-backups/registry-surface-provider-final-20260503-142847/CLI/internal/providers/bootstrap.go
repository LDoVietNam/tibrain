// Package providers - registry bootstrap from config.
package providers

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ti/cli/internal/config"
	cookiestore "github.com/ti/cli/internal/cookie"
)

// BootstrapIssue describes a non-fatal problem encountered while building
// the registry from config (e.g., missing API key for a provider).
type BootstrapIssue struct {
	Provider string `json:"provider"`
	Message  string `json:"message"`
}

// BuildRegistryFromConfig constructs a provider registry from the supplied
// config. Returns the registry plus any bootstrap issues that did not
// prevent registration of at least one provider.
func BuildRegistryFromConfig(cfg *config.Config) (*Registry, []BootstrapIssue) {
	reg := NewRegistry()
	var issues []BootstrapIssue

	// 1. Try Ti Router first (preferred - has all providers + routing)
	if router := detectRouter(); router != nil {
		if err := reg.Register(router); err != nil {
			issues = append(issues, BootstrapIssue{
				Provider: "router",
				Message:  fmt.Sprintf("register router: %v", err),
			})
		}
	}

	// 2. Wire SharedChatProvider if session/cookie available.
	if chatCfg := buildSharedChatConfig(cfg); chatCfg != nil {
		p, err := NewSharedChatProvider(chatCfg)
		if err != nil {
			issues = append(issues, BootstrapIssue{
				Provider: "sharedchat",
				Message:  fmt.Sprintf("create provider: %v", err),
			})
		} else if err := reg.Register(p); err != nil {
			issues = append(issues, BootstrapIssue{
				Provider: "sharedchat",
				Message:  fmt.Sprintf("register provider: %v", err),
			})
		}
	}

	// 3. Wire generic cookie providers from env specs.
	if cookieProviders, cookieIssues := DetectCookieProvidersFromEnv(); len(cookieProviders) > 0 || len(cookieIssues) > 0 {
		issues = append(issues, cookieIssues...)
		for _, p := range cookieProviders {
			if err := reg.Register(p); err != nil {
				issues = append(issues, BootstrapIssue{
					Provider: p.Name(),
					Message:  fmt.Sprintf("register provider: %v", err),
				})
			}
		}
	}

	// 4. Wire free-claude-code style OpenAI-compatible providers from env.
	if compatProviders, compatIssues := DetectCompatibleProvidersFromEnv(); len(compatProviders) > 0 || len(compatIssues) > 0 {
		issues = append(issues, compatIssues...)
		for _, p := range compatProviders {
			if err := reg.Register(p); err != nil {
				issues = append(issues, BootstrapIssue{
					Provider: p.Name(),
					Message:  fmt.Sprintf("register provider: %v", err),
				})
			}
		}
	}

	// 4b. Wire saved custom OpenAI-compatible providers from ~/.config/ti/providers.json.
	if customProviders, customIssues := LoadCustomCompatibleProviders(); len(customProviders) > 0 || len(customIssues) > 0 {
		issues = append(issues, customIssues...)
		for _, p := range customProviders {
			if err := reg.Register(p); err != nil {
				issues = append(issues, BootstrapIssue{
					Provider: p.Name(),
					Message:  fmt.Sprintf("register custom provider: %v", err),
				})
			}
		}
	}

	// 5. Wire API-key providers from merged config. Environment variables are
	// already folded into cfg.APIKeys by config.applyEnvOverrides(), so this
	// makes ti.json/config files work as well as shell env vars.
	if cfg != nil && cfg.APIKeys != nil {
		if key := strings.TrimSpace(cfg.APIKeys["openai"]); key != "" {
			if err := reg.Register(NewOpenAIProvider(key)); err != nil {
				issues = append(issues, BootstrapIssue{Provider: "openai", Message: fmt.Sprintf("register provider: %v", err)})
			}
		}
		if key := strings.TrimSpace(cfg.APIKeys["anthropic"]); key != "" {
			if err := reg.Register(NewAnthropicProvider(key)); err != nil {
				issues = append(issues, BootstrapIssue{Provider: "anthropic", Message: fmt.Sprintf("register provider: %v", err)})
			}
		}
		if key := strings.TrimSpace(cfg.APIKeys["openrouter"]); key != "" {
			if err := reg.Register(NewOpenRouterProvider(key)); err != nil {
				issues = append(issues, BootstrapIssue{Provider: "openrouter", Message: fmt.Sprintf("register provider: %v", err)})
			}
		}
	}

	// 6. Auto-detect API key providers from environment for compatibility.
	providers := []func() Provider{
		detectOpenAI,
		detectAnthropic,
		detectOpenRouter,
	}
	for _, detect := range providers {
		if p := detect(); p != nil {
			if err := reg.Register(p); err != nil {
				issues = append(issues, BootstrapIssue{
					Provider: p.Name(),
					Message:  fmt.Sprintf("register provider: %v", err),
				})
			}
		}
	}

	return reg, issues
}

// buildSharedChatConfig attempts to extract SharedChat credentials from config.
// Returns nil if no valid credentials found.
func buildSharedChatConfig(cfg *config.Config) *SharedChatConfig {
	if cfg == nil {
		return nil
	}

	// Try browser-cookie export first. This gives the full Cookie header and
	// lets SharedChat keep all companion cookies, not only gfsessionid.
	var cookieConfigs []CookieConfig
	sessionID := ""
	if cfg.CookieFile != "" {
		m := cookiestore.NewManager()
		if err := m.LoadFile(cfg.CookieFile, cfg.CookieProfile); err == nil {
			if sid, ok := m.GetSessionID(cfg.CookieProfile); ok {
				sessionID = sid
			}
			if httpCookies, ok := m.GetHTTPCookies(cfg.CookieProfile); ok {
				cookieConfigs = httpCookiesToConfig(httpCookies)
			}
		}
	}

	// Raw cookie header may be supplied through env; the cookie value appears at
	// the end of provider specs and also works here for quick testing.
	if rawCookie := firstNonEmpty(os.Getenv("SHAREDCHAT_COOKIE"), os.Getenv("TI_COOKIE")); rawCookie != "" {
		cookieConfigs = append(cookieConfigs, parseCookieHeader(rawCookie, cfg.CookieDomain)...)
		if sid := sessionIDFromCookieHeader(rawCookie); sid != "" {
			sessionID = sid
		}
	}

	// Fallback: read session ID from auth files.
	if sessionID == "" {
		for _, authFile := range cfg.AuthFiles {
			if sid := readSessionIDFromFile(authFile); sid != "" {
				sessionID = sid
				break
			}
		}
	}

	// If no session ID found, can't create native SharedChat provider.
	if sessionID == "" {
		return nil
	}

	chatCfg := &SharedChatConfig{
		SessionID:    sessionID,
		APIPrimary:   cfg.CookieAPIPrimary,
		APIFallback:  cfg.CookieAPIFallback,
		CookieDomain: cfg.CookieDomain,
		DefaultModel: cfg.Model,
		Cookies:      cookieConfigs,
	}

	// Add models from config if available.
	if len(cfg.Models) > 0 {
		chatCfg.Models = append([]string(nil), cfg.Models...)
	}

	return chatCfg
}

// readSessionIDFromFile attempts to read a session ID from a file.
// Supports formats: KEY=VALUE, JSON {"session_id":"..."}, or raw string.
func readSessionIDFromFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	content := string(data)

	// Try KEY=VALUE format (.env style)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if strings.HasPrefix(strings.ToUpper(line), "SESSIONID=") ||
			strings.HasPrefix(strings.ToUpper(line), "SESSION_ID=") ||
			strings.HasPrefix(strings.ToUpper(line), "SHAREDCHAT_SESSION=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			}
		}
	}

	// Try raw content (just the session ID itself)
	raw := strings.TrimSpace(content)
	raw = strings.Trim(raw, `"'`)
	if len(raw) > 10 && !strings.Contains(raw, "=") && !strings.Contains(raw, "{") {
		return raw
	}

	return ""
}

func httpCookiesToConfig(cookies []*http.Cookie) []CookieConfig {
	out := make([]CookieConfig, 0, len(cookies))
	for _, ck := range cookies {
		if ck == nil || ck.Name == "" {
			continue
		}
		out = append(out, CookieConfig{
			Name:     ck.Name,
			Value:    ck.Value,
			Domain:   ck.Domain,
			Path:     firstNonEmpty(ck.Path, "/"),
			Secure:   ck.Secure,
			HTTPOnly: ck.HttpOnly,
		})
	}
	return out
}

func parseCookieHeader(header, domain string) []CookieConfig {
	parts := strings.Split(header, ";")
	out := make([]CookieConfig, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		idx := strings.Index(part, "=")
		if idx <= 0 {
			continue
		}
		out = append(out, CookieConfig{
			Name:   strings.TrimSpace(part[:idx]),
			Value:  strings.TrimSpace(part[idx+1:]),
			Domain: domain,
			Path:   "/",
			Secure: true,
		})
	}
	return out
}

func sessionIDFromCookieHeader(header string) string {
	for _, ck := range parseCookieHeader(header, "") {
		if strings.EqualFold(ck.Name, "gfsessionid") || strings.EqualFold(ck.Name, "sessionid") || strings.EqualFold(ck.Name, "session_id") {
			return ck.Value
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
