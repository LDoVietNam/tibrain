package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ti/cli/internal/config"
	cookiestore "github.com/ti/cli/internal/cookie"
)

// CompatibleProvider implements an OpenAI-compatible /chat/completions backend.
// It is intentionally generic so local gateways, internal routers, and hosted
// OpenAI-compatible services can be wired without adding a bespoke provider.
type CompatibleProvider struct {
	name            string
	baseURL         string
	chatPath        string
	apiKey          string
	authHeader      string
	authTokenPrefix string
	customHeaders   map[string]string
	models          []string
	defaultModel    string
	timeout         time.Duration
}

// NewCompatibleProvider creates a provider from config. API keys can come from
// api_key or api_key_env. api_key may also use {env:NAME} or {file:path}.
func NewCompatibleProvider(src config.CompatibleProviderConfig) (*CompatibleProvider, error) {
	name := strings.TrimSpace(src.Name)
	if name == "" {
		return nil, fmt.Errorf("compatible provider name is required")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(src.BaseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("compatible provider %q requires base_url", name)
	}
	if src.ForceHTTPS {
		u, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("parse base_url: %w", err)
		}
		if u.Scheme != "https" && u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost" {
			return nil, fmt.Errorf("provider %q requires https base_url when force_https=true", name)
		}
	}
	chatPath := src.ChatPath
	if chatPath == "" {
		chatPath = "/chat/completions"
	}
	if !strings.HasPrefix(chatPath, "/") {
		chatPath = "/" + chatPath
	}

	apiKey := config.ResolveString(src.APIKey)
	if apiKey == "" && src.APIKeyEnv != "" {
		apiKey = os.Getenv(src.APIKeyEnv)
	}
	authHeader := src.AuthHeader
	if authHeader == "" {
		authHeader = "Authorization"
	}
	authPrefix := src.AuthTokenPrefix
	if authPrefix == "" {
		authPrefix = "Bearer"
	}

	models := append([]string(nil), src.Models...)
	defaultModel := src.DefaultModel
	if defaultModel == "" && len(models) > 0 {
		defaultModel = models[0]
	}
	if defaultModel == "" {
		defaultModel = "gpt-4o-mini"
		models = append(models, defaultModel)
	}

	timeout := time.Duration(src.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}

	return &CompatibleProvider{
		name:            name,
		baseURL:         baseURL,
		chatPath:        chatPath,
		apiKey:          apiKey,
		authHeader:      authHeader,
		authTokenPrefix: authPrefix,
		customHeaders:   cloneHeaders(src.CustomHeaders),
		models:          models,
		defaultModel:    defaultModel,
		timeout:         timeout,
	}, nil
}

func (p *CompatibleProvider) Name() string         { return p.name }
func (p *CompatibleProvider) DefaultModel() string { return p.defaultModel }
func (p *CompatibleProvider) Models() []string     { return append([]string(nil), p.models...) }

func (p *CompatibleProvider) IsHealthy() bool {
	// Some local gateways do not require auth, so base URL + model is enough.
	return p != nil && p.baseURL != "" && p.defaultModel != ""
}

func (p *CompatibleProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	return p.doChat(ctx, req, nil)
}

func (p *CompatibleProvider) ChatStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error) {
	resp, err := p.doChat(ctx, req, nil)
	if err == nil && onChunk != nil {
		onChunk(StreamChunk{Content: resp.Content, Done: true})
	}
	return resp, err
}

func (p *CompatibleProvider) doChat(ctx context.Context, req ChatRequest, decorate func(*http.Request) error) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = p.defaultModel
	}
	body, err := json.Marshal(buildOpenAIRequest(req))
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+p.chatPath, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range p.customHeaders {
		if strings.TrimSpace(k) != "" && v != "" {
			httpReq.Header.Set(k, v)
		}
	}
	if p.apiKey != "" {
		value := p.apiKey
		if p.authTokenPrefix != "" {
			value = strings.TrimSpace(p.authTokenPrefix + " " + p.apiKey)
		}
		httpReq.Header.Set(p.authHeader, value)
	}
	if decorate != nil {
		if err := decorate(httpReq); err != nil {
			return nil, err
		}
	}

	client := &http.Client{Timeout: p.timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, redactHTTPBody(respBody))
	}
	return parseOpenAIResponse(respBody)
}

// CookieHTTPProvider is an OpenAI-compatible provider that authenticates with a
// browser-exported cookie profile. It never prints cookie values and only sends
// them to the configured base_url host.
type CookieHTTPProvider struct {
	*CompatibleProvider
	cookieFile    string
	cookieProfile string
	cookieHeader  string
	cookieOK      bool
}

func NewCookieHTTPProvider(src config.CompatibleProviderConfig) (*CookieHTTPProvider, error) {
	base, err := NewCompatibleProvider(src)
	if err != nil {
		return nil, err
	}
	cookieFile := config.ExpandHome(src.CookieFile)
	if cookieFile == "" {
		return nil, fmt.Errorf("cookie_http provider %q requires cookie_file", src.Name)
	}
	profile := strings.TrimSpace(src.CookieProfile)
	if profile == "" {
		profile = src.Name
	}
	mgr := cookiestore.NewManager()
	if err := loadCookieStateOrExport(mgr, cookieFile, profile); err != nil {
		return nil, err
	}
	if _, ok := mgr.GetHTTPCookies(profile); !ok {
		if u, err := url.Parse(base.baseURL); err == nil {
			if best, found := mgr.BestProfileForHost(u.Hostname()); found {
				profile = best
			}
		}
	}
	if src.CookieDomain != "" {
		if err := ensureCookieDomainAllowed(mgr, profile, src.CookieDomain); err != nil {
			return nil, err
		}
	}
	return &CookieHTTPProvider{
		CompatibleProvider: base,
		cookieFile:         cookieFile,
		cookieProfile:      profile,
		cookieHeader:       mgr.GetHeader(profile),
		cookieOK:           !mgr.IsExpired(profile),
	}, nil
}

func (p *CookieHTTPProvider) IsHealthy() bool {
	return p != nil && p.CompatibleProvider.IsHealthy() && p.cookieOK && p.cookieHeader != ""
}

func (p *CookieHTTPProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	return p.doChat(ctx, req, p.attachCookieHeader)
}

func (p *CookieHTTPProvider) ChatStream(ctx context.Context, req ChatRequest, onChunk func(StreamChunk)) (*ChatResponse, error) {
	resp, err := p.Chat(ctx, req)
	if err == nil && onChunk != nil {
		onChunk(StreamChunk{Content: resp.Content, Done: true})
	}
	return resp, err
}

func (p *CookieHTTPProvider) attachCookieHeader(req *http.Request) error {
	if p.cookieHeader == "" {
		return fmt.Errorf("cookie profile %q is empty or unavailable", p.cookieProfile)
	}
	req.Header.Set("Cookie", p.cookieHeader)
	return nil
}

func buildCompatibleProvider(src config.CompatibleProviderConfig) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(src.Type)) {
	case "", "openai_compatible", "compatible", "http":
		return NewCompatibleProvider(src)
	case "cookie_http", "cookie", "browser_cookie":
		return NewCookieHTTPProvider(src)
	default:
		return nil, fmt.Errorf("unknown provider type %q", src.Type)
	}
}

func loadCookieStateOrExport(mgr *cookiestore.Manager, path, profile string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read cookie file: %w", err)
	}
	trimmed := strings.TrimSpace(string(data))
	if strings.HasPrefix(trimmed, "{") && strings.Contains(trimmed, "\"profiles\"") {
		if err := mgr.LoadStateFile(path); err != nil {
			return err
		}
		return nil
	}
	return mgr.LoadChromeJSON(data, profile)
}

func ensureCookieDomainAllowed(mgr *cookiestore.Manager, profile string, allowedDomain string) error {
	cks, ok := mgr.GetHTTPCookies(profile)
	if !ok {
		return fmt.Errorf("cookie profile %q not found", profile)
	}
	allowed := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(allowedDomain)), ".")
	for _, ck := range cks {
		domain := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ck.Domain)), ".")
		if domain == "" {
			continue
		}
		if domain != allowed && !strings.HasSuffix(domain, "."+allowed) {
			return fmt.Errorf("cookie %q domain %q is outside allowed domain %q", ck.Name, ck.Domain, allowedDomain)
		}
	}
	return nil
}

func cloneHeaders(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func redactHTTPBody(body []byte) string {
	text := string(body)
	if len(text) > 500 {
		text = text[:500] + "..."
	}
	for _, marker := range []string{"access_token", "refresh_token", "session", "cookie", "authorization"} {
		text = strings.ReplaceAll(text, marker, marker+":***")
	}
	return text
}
