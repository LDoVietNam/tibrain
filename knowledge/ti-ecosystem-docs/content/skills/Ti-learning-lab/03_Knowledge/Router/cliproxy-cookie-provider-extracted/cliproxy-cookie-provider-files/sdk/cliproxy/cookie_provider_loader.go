package cliproxy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	internalexecutor "github.com/router-for-me/CLIProxyAPI/v7/internal/runtime/executor"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v7/sdk/config"
)

const cookieProviderKind = "cookie-provider"

// loadCookieProviders registers cookie-backed OpenAI-compatible providers declared
// in config. It uses WithSkipPersist because config is the source of truth and the
// cookie itself is kept in Auth.Runtime, not in persisted Auth JSON.
func (s *Service) loadCookieProviders(ctx context.Context, cfg *config.Config) error {
	if s == nil || s.coreManager == nil || cfg == nil {
		return nil
	}
	providers := cfg.CookieProviders
	if len(providers) == 0 {
		return nil
	}

	ctx = coreauth.WithSkipPersist(ctx)
	registered := map[string]bool{}
	for idx, provider := range providers {
		if provider.Disabled {
			continue
		}
		providerName := strings.TrimSpace(provider.Name)
		if providerName == "" {
			return fmt.Errorf("cookie-providers[%d].name is required", idx)
		}
		baseURL := strings.TrimSpace(provider.BaseURL)
		if baseURL == "" {
			return fmt.Errorf("cookie-providers[%d].base-url is required", idx)
		}

		providerKey := "cookie-" + sanitizeProviderKey(providerName)
		if providerKey == "cookie-" {
			providerKey = fmt.Sprintf("cookie-provider-%d", idx+1)
		}
		if !registered[providerKey] {
			s.coreManager.RegisterExecutor(internalexecutor.NewCookieProviderExecutor(providerKey))
			registered[providerKey] = true
		}

		authID := stableCookieProviderAuthID(providerKey, provider.Prefix, baseURL, idx)
		now := time.Now().UTC()
		auth := &coreauth.Auth{
			ID:       authID,
			Provider: providerKey,
			Prefix:   provider.Prefix,
			Label:    providerName,
			Disabled: provider.Disabled,
			ProxyURL: provider.ProxyURL,
			Attributes: map[string]string{
				"base_url":    baseURL,
				"path":        provider.Path,
				"cookie_env":  provider.CookieEnv,
				"cookie_file": provider.CookieFile,
			},
			Metadata: map[string]any{
				"kind":            cookieProviderKind,
				"config_index":    idx,
				"disable_cooling": provider.DisableCooling,
				"request_retry":   provider.RequestRetry,
			},
			Runtime: &internalexecutor.CookieProviderRuntime{
				Cookie:   provider.Cookie,
				Headers:  provider.Headers,
				ModelMap: cookieProviderModelMap(provider),
			},
			CreatedAt: now,
			UpdatedAt: now,
		}
		if _, err := s.coreManager.Register(ctx, auth); err != nil {
			return fmt.Errorf("register cookie provider %q: %w", providerName, err)
		}
		models := cookieProviderModelInfos(providerKey, provider)
		if len(models) > 0 {
			GlobalModelRegistry().RegisterClient(auth.ID, providerKey, models)
			s.coreManager.ReconcileRegistryModelStates(ctx, auth.ID)
			s.coreManager.RefreshSchedulerEntry(auth.ID)
		}
	}
	return nil
}

func cookieProviderModelInfos(providerKey string, provider config.CookieProvider) []*ModelInfo {
	models := make([]*ModelInfo, 0, len(provider.Models))
	now := time.Now().Unix()
	for _, entry := range provider.Models {
		upstream := strings.TrimSpace(entry.Name)
		if upstream == "" {
			continue
		}
		modelID := strings.TrimSpace(entry.Alias)
		if modelID == "" {
			modelID = upstream
		}
		thinking := entry.Thinking
		if thinking == nil {
			thinking = provider.Thinking
		}
		models = append(models, &ModelInfo{
			ID:                         modelID,
			Object:                     "model",
			Created:                    now,
			OwnedBy:                    providerKey,
			Type:                       "openai",
			DisplayName:                modelID,
			Name:                       upstream,
			Description:                fmt.Sprintf("Cookie provider model %s", upstream),
			SupportedGenerationMethods: []string{"generateContent", "streamGenerateContent"},
			SupportedParameters:        []string{"stream", "tools", "tool_choice", "temperature", "top_p", "max_tokens"},
			SupportedInputModalities:   []string{"TEXT", "IMAGE"},
			SupportedOutputModalities:  []string{"TEXT"},
			Thinking:                   thinking,
			UserDefined:                true,
		})
	}
	return models
}

func cookieProviderModelMap(provider config.CookieProvider) map[string]string {
	if len(provider.Models) == 0 {
		return nil
	}
	mapping := make(map[string]string, len(provider.Models)*2)
	for _, entry := range provider.Models {
		upstream := strings.TrimSpace(entry.Name)
		if upstream == "" {
			continue
		}
		modelID := strings.TrimSpace(entry.Alias)
		if modelID == "" {
			modelID = upstream
		}
		mapping[modelID] = upstream
		mapping[upstream] = upstream
	}
	return mapping
}

func stableCookieProviderAuthID(providerKey, prefix, baseURL string, idx int) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\n%s\n%s\n%d", providerKey, prefix, baseURL, idx)))
	return "cookie-provider-" + hex.EncodeToString(sum[:])[:20]
}

func sanitizeProviderKey(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		isWord := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isWord {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
