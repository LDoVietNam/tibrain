// Package prompts defines the data models for prompt sources and prompts used
// by the prompt-intelligence layer (internal/prompt).
package prompts

import "context"

// Prompt represents a single reusable prompt with metadata used for selection
// and scoring (domain/intent matching, keyword affinity, etc.).
type Prompt struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Domain      string                 `json:"domain"`
	Intent      string                 `json:"intent,omitempty"`
	Content     string                 `json:"content,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PromptSource is a provider of prompts (in-memory, file, database, remote...).
type PromptSource interface {
	// List returns all prompts available from this source.
	List(ctx context.Context) ([]Prompt, error)
	// Get returns a single prompt by ID, if present.
	Get(ctx context.Context, id string) (Prompt, bool, error)
	// Add stores a prompt in the source.
	Add(ctx context.Context, p Prompt) error
}

// InMemorySource is a simple PromptSource backed by a map. It is the default
// source used when none is provided to the prompt-intelligence layer.
type InMemorySource struct {
	prompts map[string]Prompt
}

// NewInMemorySource creates an empty in-memory prompt source.
func NewInMemorySource() *InMemorySource {
	return &InMemorySource{prompts: make(map[string]Prompt)}
}

// List returns all stored prompts.
func (s *InMemorySource) List(ctx context.Context) ([]Prompt, error) {
	out := make([]Prompt, 0, len(s.prompts))
	for _, p := range s.prompts {
		out = append(out, p)
	}
	return out, nil
}

// Get returns a prompt by ID.
func (s *InMemorySource) Get(ctx context.Context, id string) (Prompt, bool, error) {
	p, ok := s.prompts[id]
	return p, ok, nil
}

// Add stores a prompt.
func (s *InMemorySource) Add(ctx context.Context, p Prompt) error {
	s.prompts[p.ID] = p
	return nil
}
