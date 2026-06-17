package providers

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type HealthSnapshot struct {
	Name         string    `json:"name"`
	DefaultModel string    `json:"default_model"`
	Healthy      bool      `json:"healthy"`
	Models       []string  `json:"models"`
	CheckedAt    time.Time `json:"checked_at"`
}

type SelectionCriteria struct {
	Model          string
	Preferred      []string
	RequireHealthy bool
}

type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

func NewRegistry() *Registry { return &Registry{providers: map[string]Provider{}} }

func (r *Registry) Register(p Provider) error {
	if p == nil {
		return fmt.Errorf("provider is nil")
	}
	name := strings.TrimSpace(p.Name())
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = p
	return nil
}

func (r *Registry) Get(name string) (Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *Registry) Snapshots() []HealthSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]HealthSnapshot, 0, len(r.providers))
	now := time.Now()
	for _, p := range r.providers {
		out = append(out, HealthSnapshot{Name: p.Name(), DefaultModel: p.DefaultModel(), Healthy: p.IsHealthy(), Models: append([]string(nil), p.Models()...), CheckedAt: now})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (r *Registry) Select(criteria SelectionCriteria) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.providers) == 0 {
		return nil, fmt.Errorf("no providers registered")
	}
	candidates := make([]Provider, 0, len(r.providers))
	seen := map[string]bool{}
	for _, name := range criteria.Preferred {
		if p, ok := r.providers[name]; ok && !seen[name] {
			candidates = append(candidates, p)
			seen[name] = true
		}
	}
	if len(candidates) == 0 {
		names := make([]string, 0, len(r.providers))
		for name := range r.providers {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			candidates = append(candidates, r.providers[name])
		}
	}
	var fallback Provider
	for _, p := range candidates {
		if criteria.RequireHealthy && !p.IsHealthy() {
			continue
		}
		if criteria.Model == "" || providerSupportsModel(p, criteria.Model) {
			return p, nil
		}
		if fallback == nil && (!criteria.RequireHealthy || p.IsHealthy()) {
			fallback = p
		}
	}
	if fallback != nil {
		return fallback, nil
	}
	return nil, fmt.Errorf("no provider matched criteria")
}

func (r *Registry) ChatWithFallback(ctx context.Context, req ChatRequest, preferred ...string) (*ChatResponse, string, error) {
	providers := r.resolveFallbackOrder(preferred)
	if len(providers) == 0 {
		return nil, "", fmt.Errorf("no providers registered")
	}
	var lastErr error
	for _, p := range providers {
		if !p.IsHealthy() {
			continue
		}
		resp, err := p.Chat(ctx, req)
		if err == nil {
			return resp, p.Name(), nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("all candidate providers were unhealthy")
	}
	return nil, "", lastErr
}

func (r *Registry) resolveFallbackOrder(preferred []string) []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	seen := map[string]bool{}
	ordered := make([]Provider, 0, len(r.providers))
	for _, name := range preferred {
		if p, ok := r.providers[name]; ok && !seen[name] {
			ordered = append(ordered, p)
			seen[name] = true
		}
	}
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		if !seen[name] {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		ordered = append(ordered, r.providers[name])
	}
	return ordered
}

func providerSupportsModel(p Provider, model string) bool {
	if model == "" || p.DefaultModel() == model {
		return true
	}
	for _, candidate := range p.Models() {
		if candidate == model {
			return true
		}
	}
	return false
}
