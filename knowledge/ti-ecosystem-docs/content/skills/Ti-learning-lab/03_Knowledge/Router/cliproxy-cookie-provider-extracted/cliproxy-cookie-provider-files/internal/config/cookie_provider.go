package config

import (
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
)

// CookieProvider represents a cookie-authenticated OpenAI-compatible upstream.
// Cookies are operator-supplied only; CLIProxyAPI does not extract them from a
// browser profile or automate account login.
type CookieProvider struct {
	Name           string                    `yaml:"name" json:"name"`
	Priority       int                       `yaml:"priority,omitempty" json:"priority,omitempty"`
	Disabled       bool                      `yaml:"disabled,omitempty" json:"disabled,omitempty"`
	Prefix         string                    `yaml:"prefix,omitempty" json:"prefix,omitempty"`
	BaseURL        string                    `yaml:"base-url" json:"base-url"`
	Path           string                    `yaml:"path,omitempty" json:"path,omitempty"`
	Cookie         string                    `yaml:"cookie,omitempty" json:"-"`
	CookieEnv      string                    `yaml:"cookie-env,omitempty" json:"cookie-env,omitempty"`
	CookieFile     string                    `yaml:"cookie-file,omitempty" json:"cookie-file,omitempty"`
	Headers        map[string]string         `yaml:"headers,omitempty" json:"headers,omitempty"`
	ProxyURL       string                    `yaml:"proxy-url,omitempty" json:"proxy-url,omitempty"`
	Models         []CookieProviderModel     `yaml:"models" json:"models"`
	ExcludedModels []string                  `yaml:"excluded-models,omitempty" json:"excluded-models,omitempty"`
	DisableCooling bool                      `yaml:"disable-cooling,omitempty" json:"disable-cooling,omitempty"`
	RequestRetry   int                       `yaml:"request-retry,omitempty" json:"request-retry,omitempty"`
	Thinking       *registry.ThinkingSupport `yaml:"thinking,omitempty" json:"thinking,omitempty"`
}

// CookieProviderModel describes a model exposed by a CookieProvider.
type CookieProviderModel struct {
	Name     string                    `yaml:"name" json:"name"`
	Alias    string                    `yaml:"alias" json:"alias"`
	Thinking *registry.ThinkingSupport `yaml:"thinking,omitempty" json:"thinking,omitempty"`
}

func (m CookieProviderModel) GetName() string  { return strings.TrimSpace(m.Name) }
func (m CookieProviderModel) GetAlias() string { return strings.TrimSpace(m.Alias) }

// SanitizeCookieProviders normalizes cookie provider entries and drops entries
// without required routing fields. It does not validate or log cookie values.
func (c *Config) SanitizeCookieProviders() {
	if c == nil || len(c.CookieProviders) == 0 {
		return
	}
	sanitized := make([]CookieProvider, 0, len(c.CookieProviders))
	for _, provider := range c.CookieProviders {
		provider.Name = strings.TrimSpace(provider.Name)
		provider.Prefix = strings.TrimSpace(provider.Prefix)
		provider.BaseURL = strings.TrimRight(strings.TrimSpace(provider.BaseURL), "/")
		provider.Path = strings.TrimSpace(provider.Path)
		provider.CookieEnv = strings.TrimSpace(provider.CookieEnv)
		provider.CookieFile = strings.TrimSpace(provider.CookieFile)
		provider.ProxyURL = strings.TrimSpace(provider.ProxyURL)
		if provider.Name == "" || provider.BaseURL == "" {
			continue
		}
		models := make([]CookieProviderModel, 0, len(provider.Models))
		for _, model := range provider.Models {
			model.Name = strings.TrimSpace(model.Name)
			model.Alias = strings.TrimSpace(model.Alias)
			if model.Name == "" {
				continue
			}
			models = append(models, model)
		}
		provider.Models = models
		sanitized = append(sanitized, provider)
	}
	c.CookieProviders = sanitized
}
