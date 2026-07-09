package ticonsole

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxHealthBody = 64 * 1024

// Backend is intentionally small so the renderer can later be replaced by the
// upstream AI DevKit console without moving Ti orchestration into the UI.
type Backend interface {
	Snapshot(ctx context.Context) Snapshot
}

// HTTPBackend probes the Ti control-plane services through their public HTTP APIs.
type HTTPBackend struct {
	client   *http.Client
	services []ServiceConfig
	now      func() time.Time
}

// NewHTTPBackend creates a read-only Ti backend adapter.
func NewHTTPBackend(client *http.Client, services []ServiceConfig) *HTTPBackend {
	if client == nil {
		client = &http.Client{Timeout: 1800 * time.Millisecond}
	}
	cloned := make([]ServiceConfig, len(services))
	copy(cloned, services)
	return &HTTPBackend{
		client:   client,
		services: cloned,
		now:      time.Now,
	}
}

// NewDefaultHTTPBackend builds the default local Ti topology. Router endpoints
// are candidates only; each endpoint must pass a live health probe.
func NewDefaultHTTPBackend() *HTTPBackend {
	return NewHTTPBackend(nil, DefaultServices())
}

// DefaultServices reads endpoint overrides from environment variables.
func DefaultServices() []ServiceConfig {
	services := []ServiceConfig{
		{
			Name:        "TiBrain",
			Role:        "control-plane",
			BaseURL:     envOr("TIBRAIN_URL", "http://127.0.0.1:1810"),
			HealthPaths: []string{"/api/health", "/health"},
		},
		{
			Name:        "Beads",
			Role:        "task-bus",
			BaseURL:     envOr("TI_BEADS_URL", "http://127.0.0.1:1811"),
			HealthPaths: []string{"/health", "/api/health"},
		},
	}

	rawRouters := envOr("TI_ROUTER_URLS", "http://127.0.0.1:1809,http://127.0.0.1:1807,http://127.0.0.1:1817")
	seen := map[string]struct{}{}
	for _, raw := range strings.Split(rawRouters, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if _, exists := seen[raw]; exists {
			continue
		}
		seen[raw] = struct{}{}
		services = append(services, ServiceConfig{
			Name:        routerDisplayName(raw),
			Role:        "model-router-candidate",
			BaseURL:     raw,
			HealthPaths: []string{"/health", "/api/health"},
		})
	}
	return services
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func routerDisplayName(raw string) string {
	parsed, err := url.Parse(raw)
	if err == nil {
		if port := parsed.Port(); port != "" {
			return "Router@" + port
		}
		if host := parsed.Hostname(); host != "" {
			return "Router@" + host
		}
	}
	return "Router"
}

// Snapshot probes every configured service concurrently and returns a stable,
// sorted runtime view.
func (b *HTTPBackend) Snapshot(ctx context.Context) Snapshot {
	services := make([]ServiceSnapshot, len(b.services))
	var wg sync.WaitGroup
	for index, config := range b.services {
		index, config := index, config
		wg.Add(1)
		go func() {
			defer wg.Done()
			services[index] = b.probe(ctx, config)
		}()
	}
	wg.Wait()

	sort.SliceStable(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	notices := []string{
		"Read-only MVP: Ti Console only observes runtime state.",
		"Router endpoints are candidates; configured ports are not marked active without a successful probe.",
	}
	return Snapshot{
		GeneratedAt: b.now(),
		Services:    services,
		Summary:     summarize(services),
		Notices:     notices,
	}
}

func (b *HTTPBackend) probe(ctx context.Context, config ServiceConfig) ServiceSnapshot {
	checkedAt := b.now()
	base := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	result := ServiceSnapshot{
		Name:      config.Name,
		Role:      config.Role,
		BaseURL:   base,
		State:     StateUnknown,
		CheckedAt: checkedAt,
	}

	if _, err := url.ParseRequestURI(base); err != nil {
		result.Detail = "invalid base URL: " + err.Error()
		return result
	}

	paths := config.HealthPaths
	if len(paths) == 0 {
		paths = []string{"/health", "/api/health"}
	}

	var lastErr error
	for _, healthPath := range paths {
		endpoint, err := joinEndpoint(base, healthPath)
		if err != nil {
			lastErr = err
			continue
		}

		started := time.Now()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "ti-console/0.1")

		response, err := b.client.Do(req)
		latency := time.Since(started)
		if err != nil {
			lastErr = err
			continue
		}

		body, readErr := io.ReadAll(io.LimitReader(response.Body, maxHealthBody))
		closeErr := response.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if closeErr != nil {
			lastErr = closeErr
		}

		// A missing route is not a service failure; try the next supported path.
		if response.StatusCode == http.StatusNotFound {
			lastErr = fmt.Errorf("%s returned 404", endpoint)
			continue
		}

		result.Endpoint = endpoint
		result.StatusCode = response.StatusCode
		result.Latency = latency
		result.State = classifyResponse(response.StatusCode, body)
		result.Detail = healthDetail(body, response.Status)
		return result
	}

	result.State = StateDown
	if lastErr != nil {
		result.Detail = compactText(lastErr.Error(), 180)
	} else {
		result.Detail = "no health endpoint configured"
	}
	return result
}

func joinEndpoint(base, healthPath string) (string, error) {
	parsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("base URL must include scheme and host")
	}
	if !strings.HasPrefix(healthPath, "/") {
		healthPath = "/" + healthPath
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + healthPath
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

func classifyResponse(statusCode int, body []byte) ServiceState {
	if statusCode >= 500 {
		return StateDown
	}
	if statusCode < 200 || statusCode >= 300 {
		return StateDegraded
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return StateHealthy
	}

	if value, exists := firstValue(payload, "healthy", "ok", "success"); exists {
		if boolean, ok := value.(bool); ok && !boolean {
			return StateDegraded
		}
	}
	if value, exists := firstValue(payload, "status", "state", "health"); exists {
		status := strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
		switch status {
		case "down", "failed", "failure", "error", "unhealthy", "stopped", "offline":
			return StateDown
		case "degraded", "warning", "partial", "starting":
			return StateDegraded
		}
	}
	return StateHealthy
}

func firstValue(payload map[string]any, keys ...string) (any, bool) {
	for _, key := range keys {
		if value, exists := payload[key]; exists {
			return value, true
		}
	}
	return nil, false
}

func healthDetail(body []byte, fallback string) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return fallback
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err == nil {
		for _, key := range []string{"message", "detail", "status", "state", "version"} {
			if value, exists := payload[key]; exists {
				return compactText(fmt.Sprint(value), 180)
			}
		}
	}
	return compactText(trimmed, 180)
}

func compactText(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	if limit <= 0 || len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}

// durationMillis is kept separate to make JSON/text renderers deterministic.
func durationMillis(duration time.Duration) string {
	milliseconds := duration.Milliseconds()
	if milliseconds < 0 {
		milliseconds = 0
	}
	return strconv.FormatInt(milliseconds, 10) + "ms"
}
