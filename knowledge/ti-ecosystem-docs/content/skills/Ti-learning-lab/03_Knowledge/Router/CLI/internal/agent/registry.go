// Package agent provides the Agent Registry for the Ti platform.
// Agents represent logical AI agent instances with their capabilities,
// models, configuration, and runtime status.
package agent

import (
	"errors"
	"sort"
	"sync"
	"time"
)

const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusError    = "error"
)

const healthCheckWindow = 5 * time.Minute

// ErrInvalidID is returned when an agent ID is empty.
var ErrInvalidID = errors.New("agent: id cannot be empty")

// ErrInvalidName is returned when an agent name is empty.
var ErrInvalidName = errors.New("agent: name cannot be empty")

// ErrNoExpertise is returned when an agent has no expertise tags.
var ErrNoExpertise = errors.New("agent: at least 1 expertise required")

// ErrNoModels is returned when an agent has no models.
var ErrNoModels = errors.New("agent: at least 1 model required")

// ErrNotFound is returned when an agent ID does not exist.
var ErrNotFound = errors.New("agent: not found")

// Agent represents a logical AI agent with its capabilities and configuration.
type Agent struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Expertise    []string          `json:"expertise"`
	Models       []string          `json:"models"`
	DefaultModel string            `json:"default_model"`
	Config       map[string]string `json:"config"`
	Status       string            `json:"status"`
	LastSeen     int64             `json:"last_seen"`
	TaskCount    int               `json:"task_count"`
	AvgQuality   float64           `json:"avg_quality"`
}

// Registry manages a collection of agents with thread-safe access.
type Registry struct {
	mu     sync.RWMutex
	agents map[string]*Agent
}

// NewRegistry creates an empty agent registry.
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]*Agent),
	}
}

// Register adds a new agent to the registry.
// It validates the agent fields and sets status to "active" with last_seen to the current time.
// Returns ErrInvalidID, ErrInvalidName, ErrNoExpertise, ErrNoModels for validation failures,
// or an error if the ID already exists.
func (r *Registry) Register(a *Agent) error {
	if a.ID == "" {
		return ErrInvalidID
	}
	if a.Name == "" {
		return ErrInvalidName
	}
	if len(a.Expertise) == 0 {
		return ErrNoExpertise
	}
	if len(a.Models) == 0 {
		return ErrNoModels
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[a.ID]; exists {
		return errors.New("agent: id already registered")
	}

	// DeepCopy config to prevent external mutation.
	configCopy := make(map[string]string)
	for k, v := range a.Config {
		configCopy[k] = v
	}

	// DeepCopy expertise to prevent external mutation.
	expertiseCopy := make([]string, len(a.Expertise))
	copy(expertiseCopy, a.Expertise)

	// DeepCopy models to prevent external mutation.
	modelsCopy := make([]string, len(a.Models))
	copy(modelsCopy, a.Models)

	now := time.Now().Unix()
	r.agents[a.ID] = &Agent{
		ID:           a.ID,
		Name:         a.Name,
		Description:  a.Description,
		Expertise:    expertiseCopy,
		Models:       modelsCopy,
		DefaultModel: a.DefaultModel,
		Config:       configCopy,
		Status:       StatusActive,
		LastSeen:     now,
		TaskCount:    0,
		AvgQuality:   0,
	}

	return nil
}

// Deregister removes an agent from the registry.
// Returns ErrNotFound if the ID does not exist.
func (r *Registry) Deregister(id string) error {
	if id == "" {
		return ErrInvalidID
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[id]; !exists {
		return ErrNotFound
	}
	delete(r.agents, id)
	return nil
}

// Get retrieves an agent by ID.
// Returns the agent and true if found, or nil and false if not found.
// The returned agent is a snapshot (deep copy) and is safe to use outside the lock.
func (r *Registry) Get(id string) (*Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, exists := r.agents[id]
	if !exists {
		return nil, false
	}
	return copyAgent(a), true
}

// List returns all agents sorted by ID ascending.
// Returns deep copies safe for concurrent use.
func (r *Registry) List() []*Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Agent, 0, len(r.agents))
	for _, a := range r.agents {
		result = append(result, copyAgent(a))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// FindByTaskType returns agents whose expertise overlaps with the given taskType.
// Matching is case-insensitive. Returns deep copies sorted by ID.
func (r *Registry) FindByTaskType(taskType string) []*Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	target := normalize(taskType)
	var result []*Agent

	for _, a := range r.agents {
		for _, exp := range a.Expertise {
			if normalize(exp) == target {
				result = append(result, copyAgent(a))
				break
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// UpdateStatus changes the status of an agent.
// Returns ErrNotFound if the ID does not exist.
func (r *Registry) UpdateStatus(id, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	a, exists := r.agents[id]
	if !exists {
		return ErrNotFound
	}
	a.Status = status
	a.LastSeen = time.Now().Unix()
	return nil
}

// RecordTask increments the task_count for an agent and updates the avg_quality
// using an incremental moving average.
// If the agent does not exist, this is a no-op.
func (r *Registry) RecordTask(id string, quality float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	a, exists := r.agents[id]
	if !exists {
		return
	}
	a.TaskCount++
	// Incremental moving average: avg = avg + (new - avg) / n
	a.AvgQuality += (quality - a.AvgQuality) / float64(a.TaskCount)
}

// HealthCheck returns true if the agent's last_seen is within 5 minutes of now.
// Returns false if the agent does not exist.
func (r *Registry) HealthCheck(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, exists := r.agents[id]
	if !exists {
		return false
	}
	age := time.Since(time.Unix(a.LastSeen, 0))
	return age <= healthCheckWindow
}

// Count returns the number of registered agents.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}

// copyAgent returns a deep copy of an Agent.
func copyAgent(a *Agent) *Agent {
	configCopy := make(map[string]string)
	for k, v := range a.Config {
		configCopy[k] = v
	}
	expertiseCopy := make([]string, len(a.Expertise))
	copy(expertiseCopy, a.Expertise)
	modelsCopy := make([]string, len(a.Models))
	copy(modelsCopy, a.Models)

	return &Agent{
		ID:           a.ID,
		Name:         a.Name,
		Description:  a.Description,
		Expertise:    expertiseCopy,
		Models:       modelsCopy,
		DefaultModel: a.DefaultModel,
		Config:       configCopy,
		Status:       a.Status,
		LastSeen:     a.LastSeen,
		TaskCount:    a.TaskCount,
		AvgQuality:   a.AvgQuality,
	}
}

// normalize lowercases a string for case-insensitive comparison.
func normalize(s string) string {
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		out[i] = c
	}
	return string(out)
}
