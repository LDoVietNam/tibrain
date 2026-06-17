package modelregistry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Model struct {
	Name       string   `json:"name"`
	Canonical  string   `json:"canonical"`
	Aliases    []string `json:"aliases,omitempty"`
	Roles      []string `json:"roles,omitempty"`
	RouteHint  string   `json:"route_hint,omitempty"`
	Provider   string   `json:"provider,omitempty"`
	Strength   string   `json:"strength,omitempty"`
	Deprecated bool     `json:"deprecated,omitempty"`
}

type Registry struct {
	models []Model
	index  map[string]Model
}

type State struct {
	DefaultModel string            `json:"default_model,omitempty"`
	PhaseModels  map[string]string `json:"phase_models,omitempty"`
	UpdatedAt    time.Time         `json:"updated_at,omitempty"`
}

type Resolution struct {
	Requested string `json:"requested,omitempty"`
	Phase     string `json:"phase,omitempty"`
	Model     Model  `json:"model"`
	Source    string `json:"source"`
}

func DefaultRegistry() *Registry {
	models := []Model{
		{Name: "deepseek-reasoner", Canonical: "deepseek-reasoner", Aliases: []string{"deepseek", "planner"}, Roles: []string{"plan", "spec"}, RouteHint: "cliproxyapi", Provider: "deepseek", Strength: "reasoning"},
		{Name: "claude-sonnet-4", Canonical: "claude-sonnet-4", Aliases: []string{"claude", "sonnet", "exec"}, Roles: []string{"implement", "review", "repair", "optimize"}, RouteHint: "cliproxyapi", Provider: "anthropic", Strength: "coding"},
		{Name: "claude-haiku-4", Canonical: "claude-haiku-4", Aliases: []string{"haiku", "fast"}, Roles: []string{"summarize", "fast"}, RouteHint: "cliproxyapi", Provider: "anthropic", Strength: "cheap"},
		{Name: "gpt-4o", Canonical: "gpt-4o", Aliases: []string{"4o", "openai"}, Roles: []string{"implement", "review"}, RouteHint: "cliproxyapi", Provider: "openai", Strength: "general"},
		{Name: "codex", Canonical: "codex", Aliases: []string{"codex-cli", "executor"}, Roles: []string{"implement", "repair"}, RouteHint: "cliproxyapi", Provider: "openai", Strength: "execution"},
		{Name: "sharedchat-default", Canonical: "sharedchat-default", Aliases: []string{"sharedchat"}, Roles: []string{"chat"}, RouteHint: "sharedchat", Provider: "sharedchat", Strength: "session"},
	}
	idx := map[string]Model{}
	for _, m := range models {
		idx[strings.ToLower(m.Name)] = m
		idx[strings.ToLower(m.Canonical)] = m
		for _, a := range m.Aliases {
			idx[strings.ToLower(strings.TrimSpace(a))] = m
		}
	}
	return &Registry{models: models, index: idx}
}

func (r *Registry) List() []Model {
	out := append([]Model(nil), r.models...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (r *Registry) Resolve(name string) (Model, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return Model{}, false
	}
	m, ok := r.index[name]
	return m, ok
}

func (r *Registry) Canonical(name string) string {
	if m, ok := r.Resolve(name); ok {
		return m.Canonical
	}
	return strings.TrimSpace(name)
}

func NormalizePhase(phase string) string {
	phase = strings.ToLower(strings.TrimSpace(phase))
	switch phase {
	case "exec", "execution", "code":
		return "implement"
	case "planner", "planning", "design":
		return "plan"
	case "verification":
		return "review"
	default:
		return phase
	}
}

func (s *State) ensure() {
	if s.PhaseModels == nil {
		s.PhaseModels = map[string]string{}
	}
}

func (s *State) SetDefault(model string) {
	s.ensure()
	s.DefaultModel = strings.TrimSpace(model)
	s.UpdatedAt = time.Now()
}

func (s *State) SetPhase(phase, model string) {
	s.ensure()
	phase = NormalizePhase(phase)
	if phase == "" {
		return
	}
	s.PhaseModels[phase] = strings.TrimSpace(model)
	s.UpdatedAt = time.Now()
}

func (s *State) Reset() {
	s.DefaultModel = ""
	s.PhaseModels = map[string]string{}
	s.UpdatedAt = time.Now()
}

func (r *Registry) Select(state State, phase, requested, configModel, fallback string) Resolution {
	phase = NormalizePhase(phase)
	state.ensure()
	var chosen, source string
	switch {
	case strings.TrimSpace(requested) != "":
		chosen, source = requested, "requested"
	case strings.TrimSpace(state.PhaseModels[phase]) != "":
		chosen, source = state.PhaseModels[phase], "phase-preference"
	case strings.TrimSpace(state.DefaultModel) != "":
		chosen, source = state.DefaultModel, "model-default"
	case strings.TrimSpace(configModel) != "":
		chosen, source = configModel, "config"
	default:
		chosen, source = fallback, "router-default"
	}
	if m, ok := r.Resolve(chosen); ok {
		return Resolution{Requested: requested, Phase: phase, Model: m, Source: source}
	}
	chosen = strings.TrimSpace(chosen)
	return Resolution{Requested: requested, Phase: phase, Source: source, Model: Model{Name: chosen, Canonical: chosen}}
}

type Store struct{ path string }

func NewStore(path string) *Store { return &Store{path: path} }

func (s *Store) Load() (State, error) {
	st := State{PhaseModels: map[string]string{}}
	if s == nil || s.path == "" {
		return st, nil
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return st, nil
		}
		return st, err
	}
	if err := json.Unmarshal(data, &st); err != nil {
		return State{}, err
	}
	st.ensure()
	return st, nil
}

func (s *Store) Save(st State) error {
	if s == nil || s.path == "" {
		return fmt.Errorf("model store path is empty")
	}
	st.ensure()
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}
