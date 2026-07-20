package prompt

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ti/router/tibrain/internal/prompts"
)

// PromptFeedback represents feedback on a prompt's effectiveness
type PromptFeedback struct {
	PromptID   string                 `json:"prompt_id"`
	Intent     string                 `json:"intent"`
	Domain     string                 `json:"domain"`
	Score      float64                `json:"score"` // 0.0 to 1.0
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// PromptChoice represents a prompt selected for a given context
type PromptChoice struct {
	Prompt     *prompts.Prompt `json:"prompt"`
	Score      float64         `json:"score"` // Confidence score for this selection
	Reasoning  string          `json:"reasoning,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// PromptRequest represents a request for prompt selection
type PromptRequest struct {
	Context     map[string]interface{} `json:"context,omitempty"`
	Intent      string                 `json:"intent"`
	Domain      string                 `json:"domain"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Constraints []string               `json:"constraints,omitempty"`
}

// PromptResponse represents the response from prompt selection
type PromptResponse struct {
	Primary     *PromptChoice   `json:"primary,omitempty"`
	Alternatives []*PromptChoice `json:"alternatives,omitempty"`
	Usage       string          `json:"usage,omitempty"` // How the prompt should be used
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PromptScore is a temporary struct for scoring prompts during selection
type PromptScore struct {
	Prompt *prompts.Prompt
	Score  float64
}

// PromptIntelligence handles prompt selection, retrieval, and feedback
type PromptIntelligence struct {
	mu           sync.RWMutex
	sources      map[string]prompts.PromptSource
	feedback     []*PromptFeedback // Changed to slice of pointers
	usageStats   map[string]int
	domainModels map[string]map[string]float64 // domain -> intent -> map[promptID] affinity score
}

// New creates a new PromptIntelligence instance
func New(sources map[string]prompts.PromptSource) *PromptIntelligence {
	if sources == nil {
		sources = make(map[string]prompts.PromptSource)
	}
	
	// Add default in-memory source if none provided
	if len(sources) == 0 {
		sources["memory"] = prompts.NewInMemorySource()
	}
	
	return &PromptIntelligence{
		sources:      sources,
		feedback:     make([]*PromptFeedback, 0),
		usageStats:   make(map[string]int),
		domainModels: make(map[string]map[string]float64),
	}
}

// AddSource adds a prompt source to the intelligence system
func (p *PromptIntelligence) AddSource(name string, source prompts.PromptSource) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sources[name] = source
}

// GetSource returns a prompt source by name
func (p *PromptIntelligence) GetSource(name string) (prompts.PromptSource, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	source, exists := p.sources[name]
	return source, exists
}

// ListSources returns all registered prompt sources
func (p *PromptIntelligence) ListSources() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	sources := make([]string, 0, len(p.sources))
	for name := range p.sources {
		sources = append(sources, name)
	}
	return sources
}

// Process processes a prompt request and returns the best matching prompt(s)
func (p *PromptIntelligence) Process(ctx context.Context, req *PromptRequest) (*PromptResponse, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	
	// Collect all available prompts from all sources
	allPrompts := []*prompts.Prompt{}
	for _, source := range p.sources {
		ps, err := source.List(ctx)
		if err != nil {
			// Continue with what we have from other sources
			continue
		}
		// Convert []prompts.Prompt to []*prompts.Prompt
		for _, p := range ps {
			allPrompts = append(allPrompts, &p)
		}
	}
	
	if len(allPrompts) == 0 {
		return &PromptResponse{
			Usage: "no_prompts_available",
			Metadata: map[string]interface{}{
				"reason": "no prompts found in any source",
			},
		}, nil
	}
	
	// Score each prompt based on relevance to the request
	scoredPrompts := []*PromptScore{}
	for _, prompt := range allPrompts {
		score := p.scorePrompt(prompt, req)
		scoredPrompts = append(scoredPrompts, &PromptScore{
			Prompt: prompt,
			Score:  score,
		})
	}
	
	// Sort by score descending
	sort.Slice(scoredPrompts, func(i, j int) bool {
		return scoredPrompts[i].Score > scoredPrompts[j].Score
	})
	
	// Take the top prompt as primary, and up to 2 as alternatives
	var primary *PromptChoice
	var alternatives []*PromptChoice
	
	if len(scoredPrompts) > 0 {
		top := scoredPrompts[0]
		primary = &PromptChoice{
			Prompt:     top.Prompt,
			Score:      top.Score,
			Reasoning:  p.generateReasoning(top.Prompt, req),
			Metadata:   map[string]interface{}{"rank": 1},
		}
		
		// Add up to 2 alternatives
		for i := 1; i < min(3, len(scoredPrompts)); i++ {
			alt := scoredPrompts[i]
			alternatives = append(alternatives, &PromptChoice{
				Prompt:     alt.Prompt,
				Score:      alt.Score,
				Reasoning:  p.generateReasoning(alt.Prompt, req),
				Metadata:   map[string]interface{}{"rank": i + 1},
			})
		}
	}
	
	// Record usage for the selected prompt
	if primary != nil {
		p.recordUsage(primary.Prompt.ID)
		p.updateDomainModel(req.Domain, req.Intent, primary.Prompt.ID, 0.1) // Small positive reinforcement
	}
	
	// Determine usage suggestion
	usage := "direct"
	if len(alternatives) > 0 {
		usage = "consider_alternatives"
	} else if primary != nil && primary.Score < 0.5 {
		usage = "low_confidence"
	}
	
	return &PromptResponse{
		Primary:     primary,
		Alternatives: alternatives,
		Usage:       usage,
		Metadata: map[string]interface{}{
			"total_candidates": len(allPrompts),
			"processed_at":     fmt.Sprintf("%d", time.Now().Unix()),
		},
	}, nil
}

// RecordFeedback records feedback about a prompt's effectiveness for learning
func (p *PromptIntelligence) RecordFeedback(ctx context.Context, feedback *PromptFeedback) error {
	if feedback == nil {
		return errors.New("nil feedback")
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Validate feedback
	if feedback.PromptID == "" {
		return errors.New("prompt_id is required")
	}
	if feedback.Score < 0.0 || feedback.Score > 1.0 {
		return errors.New("score must be between 0.0 and 1.0")
	}
	
	// Store feedback
	p.feedback = append(p.feedback, feedback)
	
	// Update domain model based on feedback
	if feedback.Score > 0.7 {
		// Positive reinforcement
		p.updateDomainModel(feedback.Domain, feedback.Intent, feedback.PromptID, 0.2)
	} else if feedback.Score < 0.3 {
		// Negative reinforcement
		p.updateDomainModel(feedback.Domain, feedback.Intent, feedback.PromptID, -0.1)
	}
	
	return nil
}

// GetFeedback returns all collected feedback
func (p *PromptIntelligence) GetFeedback() []*PromptFeedback {
	p.mu.RLock()
	defer p.mu.RUnlock()
	// Return a copy to prevent external modification
	result := make([]*PromptFeedback, len(p.feedback))
	copy(result, p.feedback)
	return result
}

// GetStats returns usage statistics
func (p *PromptIntelligence) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	stats := map[string]interface{}{
		"total_feedback": len(p.feedback),
		"usage_stats":    make(map[string]int),
		"sources":        p.ListSources(),
	}
	
	// Copy usage stats
	for k, v := range p.usageStats {
		stats["usage_stats"].(map[string]int)[k] = v
	}
	
	return stats
}

// scorePrompt calculates how well a prompt matches a request
func (p *PromptIntelligence) scorePrompt(prompt *prompts.Prompt, req *PromptRequest) float64 {
	score := 0.0
	
	// Base score from domain match
	if strings.EqualFold(prompt.Domain, req.Domain) {
		score += 0.4
	} else if prompt.Domain == "" || req.Domain == "" {
		// Wildcard match for empty domains
		score += 0.2
	}
	
	// Boost from learned domain model
	p.mu.RLock()
	if domainModel, ok := p.domainModels[req.Domain]; ok {
		if affinity, ok := domainModel[req.Intent]; ok { // Note: Changed from prompt.ID to req.Intent
			score += affinity * 0.3 // Weight for learned affinity
		}
	}
	p.mu.RUnlock()
	
	// Check for keyword matches in name/description
	content := strings.ToLower(prompt.Name + " " + prompt.Description)
	intentTerms := strings.Fields(strings.ToLower(req.Intent))
	
	matches := 0
	for _, term := range intentTerms {
		if strings.Contains(content, term) {
			matches++
		}
	}
	
	if len(intentTerms) > 0 {
		score += float64(matches) / float64(len(intentTerms)) * 0.3
	}
	
	// Apply usage popularity boost (logarithmic to prevent runaway feedback loops)
	p.mu.RLock()
	usageCount := p.usageStats[prompt.ID]
	p.mu.RUnlock()
	if usageCount > 0 {
		score += math.Min(0.1, math.Log(float64(usageCount+1))/10.0)
	}
	
	// Ensure score is in [0,1] range
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}
	
	return score
}

// updateDomainModel updates the learned affinity between a domain/intent and a prompt
func (p *PromptIntelligence) updateDomainModel(domain, intent, pid string, delta float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	// Initialize domain map if needed
	if p.domainModels[domain] == nil {
		p.domainModels[domain] = make(map[string]float64)
	}
	
	// Get current affinity for this domain/intent pair
	current := p.domainModels[domain][intent]
	
	// Calculate new value with bounds checking
	newValue := current + delta
	if newValue > 1.0 {
		newValue = 1.0
	}
	if newValue < 0.0 {
		newValue = 0.0
	}
	
	// Store the updated affinity
	p.domainModels[domain][intent] = newValue
}

// recordUsage increments the usage count for a prompt
func (p *PromptIntelligence) recordUsage(promptID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.usageStats[promptID]++
}

// generateReasoning creates a human-readable explanation for why a prompt was selected
func (p *PromptIntelligence) generateReasoning(prompt *prompts.Prompt, req *PromptRequest) string {
	var reasons []string
	
	if strings.EqualFold(prompt.Domain, req.Domain) {
		reasons = append(reasons, "domain match")
	} else if prompt.Domain == "" || req.Domain == "" {
		reasons = append(reasons, "wildcard domain")
	}
	
	// Check for keyword matches
	content := strings.ToLower(prompt.Name + " " + prompt.Description)
	intentTerms := strings.Fields(strings.ToLower(req.Intent))
	matches := []string{}
	
	for _, term := range intentTerms {
		if strings.Contains(content, term) {
			matches = append(matches, term)
		}
	}
	
	if len(matches) > 0 {
		reasons = append(reasons, fmt.Sprintf("keyword matches: %v", strings.Join(matches, ", ")))
	}
	
	if len(reasons) == 0 {
		reasons = append(reasons, "default selection")
	}
	
	return strings.Join(reasons, "; ")
}

// Helper function to get the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}