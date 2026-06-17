// Package brain provides the Brain Gateway API that combines
// the learning engine, memory palace, and context pack into a unified
// interface for AI agent context management.
package brain

import (
	"fmt"
	"strings"

	"github.com/ti/cli/internal/memory"
	"github.com/ti/cli/internal/repoindex"
)

// Gateway is the unified brain gateway that combines learning engine,
// memory palace, and context pack into a single API.
type Gateway struct {
	engine *Engine
	palace *memory.Palace
}

// NewGateway creates a new brain gateway.
func NewGateway(engine *Engine, palace *memory.Palace) *Gateway {
	return &Gateway{
		engine: engine,
		palace: palace,
	}
}

// ContextBundle represents the complete context bundle for an AI agent.
type ContextBundle struct {
	TaskType    string                `json:"task_type"`
	Project     string                `json:"project"`
	Goal        string                `json:"goal"`
	Memory      *memory.Stack         `json:"memory"`
	Patterns    map[string]string     `json:"patterns"`               // task_type → preferred provider
	Skills      []memory.Drawer       `json:"skills"`                 // Relevant skills
	Files       []string              `json:"files"`                  // Candidate files
	RepoHints   []string              `json:"repo_hints"`             // Repository hints
	HandoffData *memory.HandoffBundle `json:"handoff_data,omitempty"` // If continuing workflow
}

// GetContext retrieves the complete context bundle for a given task and project.
func (g *Gateway) GetContext(project, task string, options ...ContextOption) (*ContextBundle, error) {
	opts := defaultContextOptions()
	for _, opt := range options {
		opt(opts)
	}

	// 1. Classify task type (simple heuristic: use first word)
	taskType := classifyTaskType(task)

	// 2. Recall memory (L0 + L1 + L2)
	memStack := g.palace.Recall(project, opts.memoryLimit)

	// 3. Get patterns from brain (preferred providers)
	patterns := g.engine.GetPreferredProviders()

	// 4. Load skills from memory (skill drawers)
	skills := g.palace.ListSkills("") // Get all skills
	if opts.skillCategory != "" {
		skills = g.palace.ListSkills(opts.skillCategory)
	}

	// 5. Build context bundle
	bundle := &ContextBundle{
		TaskType:  taskType,
		Project:   project,
		Goal:      task,
		Memory:    memStack,
		Patterns:  patterns,
		Skills:    skills,
		RepoHints: []string{},
	}

	// 6. Add repo hints if summary provided
	if opts.summary != nil {
		bundle.RepoHints = buildRepoHints(opts.summary)
	}

	// 7. Add candidate files if provided
	if len(opts.candidateFiles) > 0 {
		bundle.Files = opts.candidateFiles
	}

	return bundle, nil
}

// GetContextWithHandoff retrieves context including previous handoff data.
func (g *Gateway) GetContextWithHandoff(project, task, fromAgent, toAgent string, options ...ContextOption) (*ContextBundle, error) {
	// Get base context
	bundle, err := g.GetContext(project, task, options...)
	if err != nil {
		return nil, err
	}

	// Load handoff data
	handoff, err := g.palace.RecallHandoff(fromAgent, toAgent)
	if err != nil {
		// Handoff not found - that's OK, continue without it
		return bundle, nil
	}

	bundle.HandoffData = handoff
	return bundle, nil
}

// RecallMemory retrieves memory from the palace.
func (g *Gateway) RecallMemory(project string, limit int) (*memory.Stack, error) {
	if limit <= 0 {
		limit = memory.MaxL1Drawers
	}
	return g.palace.Recall(project, limit), nil
}

// SearchMemory performs deep search across the palace.
func (g *Gateway) SearchMemory(query string, wings []string, limit int) (*memory.Stack, error) {
	if limit <= 0 {
		limit = memory.DefaultL3Limit
	}
	return g.palace.Search(query, wings, limit), nil
}

// ObservePattern records a pattern observation to the brain engine.
func (g *Gateway) ObservePattern(event string) error {
	// This would typically come from log ingestion
	// For now, we'll implement a simple version
	return nil
}

// GetSkill retrieves a skill by name.
func (g *Gateway) GetSkill(name string) (*memory.Drawer, error) {
	return g.palace.GetSkill(name)
}

// ListSkills lists all skills, optionally filtered by category.
func (g *Gateway) ListSkills(category string) []memory.Drawer {
	return g.palace.ListSkills(category)
}

// StoreSkill stores a skill in the palace.
func (g *Gateway) StoreSkill(name, content, category string) error {
	return g.palace.StoreSkill(name, content, category)
}

// CreateHandoff creates a handoff between agents.
func (g *Gateway) CreateHandoff(fromAgent, toAgent, context, output string, metadata map[string]string) (string, error) {
	return g.palace.CreateHandoff(fromAgent, toAgent, context, output, metadata)
}

// RecallHandoff retrieves a handoff between agents.
func (g *Gateway) RecallHandoff(fromAgent, toAgent string) (*memory.HandoffBundle, error) {
	return g.palace.RecallHandoff(fromAgent, toAgent)
}

// ListHandoffs lists all handoffs involving a specific agent.
func (g *Gateway) ListHandoffs(agent string) []memory.HandoffBundle {
	return g.palace.ListHandoffs(agent)
}

// RenderContext converts a context bundle to a string for AI injection.
func (g *Gateway) RenderContext(bundle *ContextBundle) string {
	var sb strings.Builder

	// Project and goal
	if bundle.Project != "" {
		sb.WriteString(fmt.Sprintf("Project: %s\n", bundle.Project))
	}
	if bundle.Goal != "" {
		sb.WriteString(fmt.Sprintf("Task: %s\n", bundle.Goal))
	}
	if bundle.TaskType != "" {
		sb.WriteString(fmt.Sprintf("Task Type: %s\n", bundle.TaskType))
	}
	sb.WriteString("\n")

	// Handoff data (if present)
	if bundle.HandoffData != nil {
		sb.WriteString("## Previous Handoff\n")
		sb.WriteString(fmt.Sprintf("From: %s\n", bundle.HandoffData.FromAgent))
		sb.WriteString(fmt.Sprintf("To: %s\n", bundle.HandoffData.ToAgent))
		sb.WriteString(fmt.Sprintf("Output: %s\n", bundle.HandoffData.Output))
		sb.WriteString("\n")
	}

	// Memory
	if bundle.Memory != nil {
		sb.WriteString(memory.StackToContext(bundle.Memory))
	}

	// Skills
	if len(bundle.Skills) > 0 {
		sb.WriteString("## Relevant Skills\n")
		for _, skill := range bundle.Skills {
			snippet := strings.ReplaceAll(strings.TrimSpace(skill.Text), "\n", " ")
			if len(snippet) > 200 {
				snippet = snippet[:200] + "..."
			}
			sb.WriteString(fmt.Sprintf("- [%s] %s\n", skill.Room, snippet))
		}
		sb.WriteString("\n")
	}

	// Patterns
	if len(bundle.Patterns) > 0 {
		sb.WriteString("## Learned Patterns\n")
		for taskType, provider := range bundle.Patterns {
			sb.WriteString(fmt.Sprintf("- %s → %s\n", taskType, provider))
		}
		sb.WriteString("\n")
	}

	// Repo hints
	if len(bundle.RepoHints) > 0 {
		sb.WriteString("## Repository Hints\n")
		for _, hint := range bundle.RepoHints {
			sb.WriteString(fmt.Sprintf("- %s\n", hint))
		}
		sb.WriteString("\n")
	}

	// Candidate files
	if len(bundle.Files) > 0 {
		sb.WriteString("## Candidate Files\n")
		for _, file := range bundle.Files {
			sb.WriteString(fmt.Sprintf("- %s\n", file))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// ─────────────────────────────────────────────────────────────
// Context Options
// ─────────────────────────────────────────────────────────────

// ContextOption is a function that configures context retrieval.
type ContextOption func(*contextOptions)

type contextOptions struct {
	memoryLimit    int
	skillCategory  string
	candidateFiles []string
	summary        *repoindex.Summary
}

func defaultContextOptions() *contextOptions {
	return &contextOptions{
		memoryLimit:    memory.MaxL1Drawers,
		skillCategory:  "",
		candidateFiles: nil,
		summary:        nil,
	}
}

// WithMemoryLimit sets the memory retrieval limit.
func WithMemoryLimit(limit int) ContextOption {
	return func(opts *contextOptions) {
		opts.memoryLimit = limit
	}
}

// WithSkillCategory filters skills by category.
func WithSkillCategory(category string) ContextOption {
	return func(opts *contextOptions) {
		opts.skillCategory = category
	}
}

// WithCandidateFiles sets the candidate files for context.
func WithCandidateFiles(files []string) ContextOption {
	return func(opts *contextOptions) {
		opts.candidateFiles = files
	}
}

// WithSummary sets the repository summary for repo hints.
func WithSummary(summary *repoindex.Summary) ContextOption {
	return func(opts *contextOptions) {
		opts.summary = summary
	}
}

// ─────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────

func classifyTaskType(task string) string {
	task = strings.ToLower(strings.TrimSpace(task))
	switch {
	case strings.Contains(task, "code") || strings.Contains(task, "implement"):
		return "code"
	case strings.Contains(task, "debug") || strings.Contains(task, "fix") || strings.Contains(task, "error"):
		return "debugging"
	case strings.Contains(task, "test"):
		return "testing"
	case strings.Contains(task, "doc") || strings.Contains(task, "write"):
		return "writing"
	case strings.Contains(task, "explain") || strings.Contains(task, "review"):
		return "explaining"
	case strings.Contains(task, "plan") || strings.Contains(task, "design"):
		return "plan"
	default:
		return "general"
	}
}

func buildRepoHints(summary *repoindex.Summary) []string {
	if summary == nil {
		return nil
	}

	hints := []string{}
	if summary.Module != "" {
		hints = append(hints, fmt.Sprintf("Go module `%s`", summary.Module))
	}
	if len(summary.EntryPoints) > 0 {
		hints = append(hints, fmt.Sprintf("Likely entry points: %s", strings.Join(summary.EntryPoints[:min(4, len(summary.EntryPoints))], ", ")))
	}
	if len(summary.TopDirectories) > 0 {
		hints = append(hints, fmt.Sprintf("Top directories: %s", strings.Join(summary.TopDirectories[:min(5, len(summary.TopDirectories))], ", ")))
	}
	if summary.TestFileCount > 0 {
		hints = append(hints, fmt.Sprintf("Detected %d test files", summary.TestFileCount))
	}

	return hints
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
