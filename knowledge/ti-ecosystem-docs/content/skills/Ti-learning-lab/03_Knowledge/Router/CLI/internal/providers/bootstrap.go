// Package providers - registry bootstrap from config.
package providers

import (
	"bufio"
	"fmt"
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
// config. Returns the registry plus any bootstrap issues that did not prevent
// registration of at least one provider.
func BuildRegistryFromConfig(cfg *config.Config) (*Registry, []BootstrapIssue) {
	reg := NewRegistry()
	var issues []BootstrapIssue

	// 1. Try Ti Router first (preferred - has all providers + routing).
	if router := detectRouter(); router != nil {
		if err := reg.Register(router); err != nil {
			issues = append(issues, BootstrapIssue{Provider: "router", Message: fmt.Sprintf("register router: %v", err)})
		}
	}

	// 2. Wire SharedChatProvider if session/cookie is available.
	if chatCfg := buildSharedChatConfig(cfg); chatCfg != nil {
		p, err := NewSharedChatProvider(chatCfg)
		if err != nil {
			issues = append(issues, BootstrapIssue{Provider: "sharedchat", Message: fmt.Sprintf("create provider: %v", err)})
		} else if err := reg.Register(p); err != nil {
			issues = append(issues, BootstrapIssue{Provider: "sharedchat", Message: fmt.Sprintf("register provider: %v", err)})
		}
	}

	// 3. Register config-defined compatible providers. This is where the
	// chatgpt-cli-style OpenAI-compatible backends and cookie_http providers live.
	for _, providerCfg := range cfg.CompatibleProviders {
		p, err := buildCompatibleProvider(providerCfg)
		if err != nil {
			name := providerCfg.Name
			if name == "" {
				name = "compatible"
			}
			issues = append(issues, BootstrapIssue{Provider: name, Message: err.Error()})
			continue
		}
		if err := reg.Register(p); err != nil {
			issues = append(issues, BootstrapIssue{Provider: p.Name(), Message: fmt.Sprintf("register provider: %v", err)})
		}
	}

	// 4. Auto-detect API key providers from environment.
	providers := []func() Provider{detectOpenAI, detectAnthropic, detectOpenRouter}
	for _, detect := range providers {
		if p := detect(); p != nil {
			if err := reg.Register(p); err != nil {
				issues = append(issues, BootstrapIssue{Provider: p.Name(), Message: fmt.Sprintf("register provider: %v", err)})
			}
		}
	}

	return reg, issues
}

// buildSharedChatConfig attempts to extract SharedChat credentials from config.
// Returns nil if no valid credentials are found. It supports both legacy auth
// files and the new imported browser-cookie state file.
func buildSharedChatConfig(cfg *config.Config) *SharedChatConfig {
	if cfg == nil {
		return nil
	}

	sessionID := ""
	var cookies []CookieConfig

	if cfg.CookieFile != "" {
		mgr := cookiestore.NewManager()
		if err := loadSharedChatCookies(mgr, cfg.CookieFile, cfg.CookieProfile); err == nil {
			profile := cfg.CookieProfile
			if profile == "" {
				profile = "sharedchat"
			}
			if _, ok := mgr.GetHTTPCookies(profile); !ok {
				if best, found := mgr.BestProfileForHost("chat.sharedchat.fun"); found {
					profile = best
				}
			}
			if sid, ok := mgr.GetSessionID(profile); ok {
				sessionID = sid
			}
			if httpCookies, ok := mgr.GetHTTPCookies(profile); ok {
				cookies = make([]CookieConfig, 0, len(httpCookies))
				for _, ck := range httpCookies {
					cookies = append(cookies, CookieConfig{
						Name:     ck.Name,
						Value:    ck.Value,
						Domain:   ck.Domain,
						Path:     ck.Path,
						Secure:   ck.Secure,
						HTTPOnly: ck.HttpOnly,
					})
				}
			}
		}
	}

	// Fallback to legacy auth files.
	if sessionID == "" {
		for _, authFile := range cfg.AuthFiles {
			if sid := readSessionIDFromFile(config.ExpandHome(authFile)); sid != "" {
				sessionID = sid
				break
			}
		}
	}

	if sessionID == "" {
		return nil
	}

	chatCfg := &SharedChatConfig{
		SessionID:    sessionID,
		APIPrimary:   cfg.CookieAPIPrimary,
		APIFallback:  cfg.CookieAPIFallback,
		CookieDomain: cfg.CookieDomain,
		DefaultModel: cfg.Model,
		Cookies:      cookies,
	}
	if len(cfg.Models) > 0 {
		chatCfg.Models = append([]string(nil), cfg.Models...)
	}
	return chatCfg
}

func loadSharedChatCookies(mgr *cookiestore.Manager, path, profile string) error {
	path = config.ExpandHome(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if profile == "" {
		profile = "sharedchat"
	}
	trimmed := strings.TrimSpace(string(data))
	if strings.HasPrefix(trimmed, "{") && strings.Contains(trimmed, "\"profiles\"") {
		if err := mgr.LoadStateFile(path); err != nil {
			return err
		}
		if _, ok := mgr.GetHTTPCookies(profile); ok {
			return nil
		}
		if best, ok := mgr.BestProfileForHost("chat.sharedchat.fun"); ok {
			profile = best
			return nil
		}
		return fmt.Errorf("profile %q not found", profile)
	}
	return mgr.LoadChromeJSON(data, profile)
}

// readSessionIDFromFile attempts to read a session ID from a file.
// Supports formats: KEY=VALUE, JSON-like files with a session_id string, or raw string.
func readSessionIDFromFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	content := string(data)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		upper := strings.ToUpper(line)
		if strings.HasPrefix(upper, "SESSIONID=") ||
			strings.HasPrefix(upper, "SESSION_ID=") ||
			strings.HasPrefix(upper, "SHAREDCHAT_SESSION=") ||
			strings.HasPrefix(upper, "GFSESSIONID=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			}
		}
	}

	raw := strings.TrimSpace(content)
	raw = strings.Trim(raw, `"'`)
	if len(raw) > 10 && !strings.Contains(raw, "=") && !strings.Contains(raw, "{") {
		return raw
	}
	return ""
}
