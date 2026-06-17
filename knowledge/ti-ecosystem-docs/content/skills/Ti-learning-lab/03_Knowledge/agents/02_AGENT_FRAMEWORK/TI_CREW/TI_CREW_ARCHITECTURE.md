# Ti Crew - Architecture Design

> **Version**: 1.0.0
> **Last Updated**: 2026-05-05
> **Purpose**: Ti Crew - Agent orchestration framework tương tự OpenClaw/Hermers để quản lý 104+ agents trong Ti ecosystem

---

## 🎯 Overview

Ti project hiện có **104 agents** trong `content/agents/agents/`. Ti Crew được design để:

- **Orchestrate** 104+ agents efficiently
- **Route tasks** đến đúng agent phù hợp
- **Track performance** của từng agent
- **Enable learning** và knowledge sharing giữa agents
- **Provide marketplace** để discover và reuse agents
- **Auto-scale** agent instances dựa trên load

---

## 📊 Current Agent Inventory

### Agent Categories

| Category | Count | Examples |
|----------|-------|----------|
| **Technical Specialists** | 30+ | golang-pro, python-pro, rust-pro, cpp-reviewer |
| **Domain Experts** | 20+ | backend-architect, frontend-design, database-admin |
| **Quality Assurance** | 15+ | code-auditor, security-reviewer, test-generator |
| **Business/Analysis** | 10+ | business-analyst, data-analysis-agent |
| **Operations** | 10+ | devops-troubleshooter, performance-optimizer |
| **Content/Design** | 10+ | content-marketer, ui-ux-designer |
| **System Orchestrators** | 5+ | jarvis, planner, reviewer |
| **Ti-Specific** | 4+ | tibrain-specialist, mcp-expert, handoff-coordinator |

**Total**: 104 agents

---

## 🏗️ Framework Architecture

### Core Components

```
┌─────────────────────────────────────────────────────────────────┐
│                         Ti Crew                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ Agent        │  │ Task         │  │ Knowledge     │  │
│  │ Registry     │  │ Router       │  │ Graph        │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ Agent        │  │ Performance  │  │ Agent        │  │
│  │ Marketplace │  │ Monitor     │  │ Learning     │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ Agent        │  │ Agent        │  │ Agent        │  │
│  │ Orchestrator │  │ Communicator  │  │ Lifecycle     │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Component Details

### 1. Agent Registry

**Purpose**: Quản lý metadata và lifecycle của tất cả agents

**File**: `apps/ticrew/internal/registry/registry.go`

```go
package registry

import (
	"context"
	"encoding/json"
	"sync"
)

// AgentMetadata holds metadata about an agent
type AgentMetadata struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`        // technical, domain, qa, business, ops
	Description string            `json:"description"`
	Version     string            `json:"version"`
	Author      string            `json:"author"`
	Capabilities []string          `json:"capabilities"`
	Tools       []string          `json:"tools"`
	MCPs        []string          `json:"mcps"`
	Model       string            `json:"model"`        // gemini-2.5-pro, claude-sonnet, etc
	Priority    string            `json:"priority"`     // P0, P1, P2
	Color       string            `json:"color"`
	Tags        []string          `json:"tags"`
	FilePath    string            `json:"file_path"`
	Status      AgentStatus       `json:"status"`
	Stats       AgentStats        `json:"stats"`
	LastUpdated int64             `json:"last_updated"`
}

type AgentStatus string

const (
	StatusActive   AgentStatus = "active"
	StatusInactive AgentStatus = "inactive"
	StatusDeprecated AgentStatus = "deprecated"
	StatusTesting  AgentStatus = "testing"
)

type AgentStats struct {
	TaskCount      int64   `json:"task_count"`
	SuccessCount   int64   `json:"success_count"`
	FailureCount   int64   `json:"failure_count"`
	AvgDuration    float64 `json:"avg_duration_ms"`
	LastUsed       int64   `json:"last_used"`
	QualityScore   float64 `json:"quality_score"`
}

// AgentRegistry manages agent metadata
type AgentRegistry struct {
	agents map[string]*AgentMetadata
	mu     sync.RWMutex
}

// NewAgentRegistry creates a new agent registry
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]*AgentMetadata),
	}
}

// Register registers an agent
func (r *AgentRegistry) Register(ctx context.Context, metadata AgentMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	metadata.Status = StatusActive
	metadata.LastUpdated = time.Now().Unix()

	r.agents[metadata.ID] = &metadata
	return nil
}

// Get gets an agent by ID
func (r *AgentRegistry) Get(ctx context.Context, id string) (*AgentMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, ok := r.agents[id]
	if !ok {
		return nil, fmt.Errorf("agent not found: %s", id)
	}

	return agent, nil
}

// List lists all agents
func (r *AgentRegistry) List(ctx context.Context) ([]*AgentMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*AgentMetadata, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}

	return agents, nil
}

// Search searches agents by criteria
func (r *AgentRegistry) Search(ctx context.Context, query SearchQuery) ([]*AgentMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := []*AgentMetadata{}

	for _, agent := range r.agents {
		if r.matchesQuery(agent, query) {
			results = append(results, agent)
		}
	}

	return results, nil
}

// Discover discovers agents from filesystem
func (r *AgentRegistry) Discover(ctx context.Context, agentDir string) error {
	entries, err := os.ReadDir(agentDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		metadata, err := r.parseAgentFile(filepath.Join(agentDir, entry.Name()))
		if err != nil {
			continue
		}

		r.Register(ctx, metadata)
	}

	return nil
}

// parseAgentFile parses agent metadata from markdown file
func (r *AgentRegistry) parseAgentFile(path string) (AgentMetadata, error) {
	// Parse YAML frontmatter and markdown body
	// Extract agent metadata from frontmatter
	return AgentMetadata{}, nil
}
```

---

### 2. Task Router

**Purpose**: Route tasks đến đúng agent dựa trên criteria

**File**: `apps/ticrew/internal/router/router.go`

```go
package router

import (
	"context"
	"fmt"
)

// Task represents a task to be routed
type Task struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`        // coding, review, planning, etc.
	Priority    string                 `json:"priority"`
	Tags        []string               `json:"tags"`
	Language    string                 `json:"language"`    // go, python, rust, etc.
	Framework   string                 `json:"framework"`   // react, django, spring, etc.
	Context     map[string]any         `json:"context"`
	Requirements map[string]string      `json:"requirements"`
}

// RoutingResult represents routing result
type RoutingResult struct {
	AgentID    string   `json:"agent_id"`
	Confidence float64  `json:"confidence"`
	Reason     string   `json:"reason"`
	Alternatives []string `json:"alternatives"`
}

// TaskRouter routes tasks to appropriate agents
type TaskRouter struct {
	registry *AgentRegistry
	llm      LLMClient
}

// NewTaskRouter creates a new task router
func NewTaskRouter(registry *AgentRegistry, llm LLMClient) *TaskRouter {
	return &TaskRouter{
		registry: registry,
		llm:      llm,
	}
}

// Route routes a task to the best agent
func (tr *TaskRouter) Route(ctx context.Context, task Task) (*RoutingResult, error) {
	// Step 1: Filter agents by capabilities
	candidates := tr.filterAgents(ctx, task)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no suitable agent found for task")
	}

	// Step 2: Use LLM to rank candidates
	ranking, err := tr.rankCandidates(ctx, task, candidates)
	if err != nil {
		// Fallback to simple scoring
		ranking = tr.simpleScore(ctx, task, candidates)
	}

	// Step 3: Select top agent
	bestAgent := ranking[0]

	return &RoutingResult{
		AgentID:    bestAgent.ID,
		Confidence: bestAgent.Score,
		Reason:     bestAgent.Reason,
		Alternatives: tr.extractAlternatives(ranking),
	}, nil
}

// filterAgents filters agents by capabilities
func (tr *TaskRouter) filterAgents(ctx context.Context, task Task) []*AgentScore {
	agents, err := tr.registry.List(ctx)
	if err != nil {
		return nil
	}

	candidates := []*AgentScore{}

	for _, agent := range agents {
		if agent.Status != StatusActive {
			continue
		}

	// Check if agent has required capabilities
		if tr.hasRequiredCapabilities(agent, task) {
			candidates = append(candidates, &AgentScore{
				Agent:  agent,
				Score:   0,
				Reason:  "",
			})
		}
	}

	return candidates
}

// rankCandidates uses LLM to rank candidates
func (tr *TaskRouter) rankCandidates(ctx context.Context, task Task, candidates []*AgentScore) ([]*AgentScore, error) {
	// Build prompt for LLM
	prompt := tr.buildRankingPrompt(task, candidates)

	// Call LLM
	response, err := tr.llm.Call(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse LLM response
	return tr.parseRankingResponse(response, candidates)
}

// simpleScore provides fallback scoring without LLM
func (tr *TaskRouter) simpleScore(ctx context.Context, task Task, candidates []*AgentScore) []*AgentScore {
	for _, candidate := range candidates {
		score := 0.0

		// Score based on language match
		if task.Language != "" {
			for _, tag := range candidate.Agent.Tags {
				if tag == task.Language {
					score += 30
				}
			}
		}

		// Score based on framework match
		if task.Framework != "" {
			for _, tag := range candidate.Agent.Tags {
				if tag == task.Framework {
					score += 30
				}
			}
		}

		// Score based on priority
		switch candidate.Agent.Priority {
		case "P0":
			score += 20
		case "P1":
			score += 10
		}

		candidate.Score = score
		candidate.Reason = fmt.Sprintf("Score: %.2f (language: %s, framework: %s, priority: %s)",
			score, task.Language, task.Framework, candidate.Agent.Priority)
	}

	// Sort by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	return candidates
}

type AgentScore struct {
	Agent  *AgentMetadata
	Score  float64
	Reason string
}
```

---

### 3. Agent Orchestrator

**Purpose**: Orchestrate multi-agent workflows

**File**: `apps/ticrew/internal/orchestrator/orchestrator.go`

```go
package orchestrator

import (
	"context"
	"fmt"
)

// Workflow represents a multi-agent workflow
type Workflow struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Steps       []WorkflowStep `json:"steps"`
	Status      WorkflowStatus `json:"status"`
}

type WorkflowStep struct {
	AgentID  string `json:"agent_id"`
	Task     string `json:"task"`
	Parallel bool   `json:"parallel"`
	Depends  []string `json:"depends"`
}

type WorkflowStatus string

const (
	WorkflowPending   WorkflowStatus = "pending"
	WorkflowRunning   WorkflowStatus = "running"
	WorkflowCompleted WorkflowStatus = "completed"
	WorkflowFailed   WorkflowStatus = "failed"
)

// AgentOrchestrator orchestrates multi-agent workflows
type AgentOrchestrator struct {
	registry *AgentRegistry
	router  *TaskRouter
	llm      LLMClient
	beads    *BeadsClient
}

// NewAgentOrchestrator creates a new agent orchestrator
func NewAgentOrchestrator(registry *AgentRegistry, router *TaskRouter, llm LLMClient, beads *BeadsClient) *AgentOrchestrator {
	return &AgentOrchestrator{
		registry: registry,
		router:  router,
		llm:      llm,
		beads:    beads,
	}
}

// Execute executes a workflow
func (ao *AgentOrchestrator) Execute(ctx context.Context, workflow Workflow) error {
	workflow.Status = WorkflowRunning

	// Build dependency graph
	graph := ao.buildDependencyGraph(workflow.Steps)

	// Execute steps in topological order
	steps := ao.topologicalSort(graph)

	for _, step := range steps {
		if step.Parallel {
			// Execute in parallel
			err := ao.executeParallel(ctx, step)
		} else {
			// Execute sequentially
			err := ao.executeSequential(ctx, step)
		}

		if err != nil {
			workflow.Status = WorkflowFailed
			return err
		}
	}

	workflow.Status = WorkflowCompleted
	return nil
}

// executeSequential executes a step sequentially
func (ao *AgentOrchestrator) executeSequential(ctx context.Context, step WorkflowStep) error {
	// Route task to agent
	result, err := ao.router.Route(ctx, Task{
		Description: step.Task,
	})
	if err != nil {
		return err
	}

	// Execute task on agent
	return ao.executeOnAgent(ctx, result.AgentID, step.Task)
}

// executeParallel executes steps in parallel
func (ao *AgentOrchestrator) executeParallel(ctx context.Context, step WorkflowStep) error {
	// Implement parallel execution
	return nil
}

// executeOnAgent executes a task on a specific agent
func (ao *AgentOrchestrator executeOnAgent(ctx context.Context, agentID, task string) error {
	// Get agent
	agent, err := ao.registry.Get(ctx, agentID)
	if err != nil {
		return err
	}

	// Load agent definition
	definition, err := ao.loadAgentDefinition(agent.FilePath)
	if err != nil {
		return err
	}

	// Execute with LLM
	return ao.executeWithLLM(ctx, definition, task)
}
```

---

### 4. Agent Marketplace

**Purpose**: Marketplace để discover, rate, reuse agents

**File**: `apps/ticrew/internal/marketplace/marketplace.go`

```go
package marketplace

import (
	"context"
	"fmt"
)

// Marketplace manages agent marketplace
type Marketplace struct {
	registry    *AgentRegistry
	ratings     map[string]*AgentRating
	reviews     map[string][]AgentReview
	downloads   map[string]int64
	popularity  map[string]float64
}

type AgentRating struct {
	AgentID    string  `json:"agent_id"`
	AvgRating  float64 `json:"avg_rating"`
	Count      int     `json:"count"`
	Categories map[string]float64 `json:"categories"`
}

type AgentReview struct {
	AgentID    string  `json:"agent_id"`
	UserID     string  `json:"user_id"`
	Rating     int     `json:"rating"`
	Comment    string  `json:"comment"`
	Timestamp  int64   `json:"timestamp"`
}

// NewMarketplace creates a new marketplace
func NewMarketplace(registry *AgentRegistry) *Marketplace {
	return &Marketplace{
		registry:   registry,
		ratings:    make(map[string]*AgentRating),
		reviews:    make(map[string][]AgentReview),
		downloads:  make(map[string]int64),
		popularity: make(map[string]float64),
	}
}

// Discover discovers agents by criteria
func (m *Marketplace) Discover(ctx context.Context, query MarketplaceQuery) ([]*AgentMetadata, error) {
	agents, err := m.registry.Search(ctx, SearchQuery{
		Type:    query.Type,
		Tags:    query.Tags,
		Language: query.Language,
	})
	if err != nil {
		return nil, err
	}

	// Sort by popularity
	sort.Slice(agents, func(i, j int) bool {
		return m.popularity[agents[i].ID] > m.popularity[agents[j].ID]
	})

	return agents, nil
}

// Rate rates an agent
func (m *Marketplace) Rate(ctx context.Context, agentID, userID string, rating int, comment string) error {
	// Add review
	review := AgentReview{
		AgentID:   agentID,
		UserID:    userID,
		Rating:    rating,
		Comment:   comment,
		Timestamp: time.Now().Unix(),
	}

	m.reviews[agentID] = append(m.reviews[agentID], review)

	// Recalculate rating
	m.recalculateRating(agentID)

	return nil
}

// Download increments download count
func (m *Marketplace) Download(ctx context.Context, agentID string) error {
	m.downloads[agentID]++

	// Update popularity score
	m.updatePopularity(agentID)

	return nil
}

// GetTrending returns trending agents
func (m *Marketplace) GetTrending(ctx context.Context, limit int) ([]*AgentMetadata, error) {
	agents, err := m.registry.List(ctx)
	if err != nil {
		return nil, err
	}

	// Sort by popularity (recent downloads + ratings)
	sorted := make([]*AgentMetadata, len(agents))
	copy(sorted, agents)

	sort.Slice(sorted, func(i, j int) bool {
		return m.popularity[sorted[i].ID] > m.popularity[sorted[j].ID]
	})

	if limit > len(sorted) {
		limit = len(sorted)
	}

	return sorted[:limit], nil
}
```

---

### 5. Performance Monitor

**Purpose**: Track performance của từng agent

**File**: `apps/ticrew/internal/monitor/monitor.go`

```go
package monitor

import (
	"context"
	"time"
)

// PerformanceMonitor tracks agent performance
type PerformanceMonitor struct {
	registry *AgentRegistry
	metrics  map[string]*AgentMetrics
	mu       sync.RWMutex
}

type AgentMetrics struct {
	AgentID         string            `json:"agent_id"`
	TaskCount       int64             `json:"task_count"`
	SuccessCount    int64             `json:"success_count"`
	FailureCount    int64             `json:"failure_count"`
	AvgDuration     float64           `json:"avg_duration_ms"`
	P95Duration     float64           `json:"p95_duration_ms"`
	P99Duration     float64           `json:"p99_duration_ms"`
	ErrorRate       float64           `json:"error_rate"`
	QualityScore    float64           `json:"quality_score"`
	LastUsed        int64             `json:"last_used"`
	UsagePattern    []UsagePoint      `json:"usage_pattern"`
	ResourceUsage   ResourceUsage     `json:"resource_usage"`
}

type UsagePoint struct {
	Timestamp int64  `json:"timestamp"`
	TaskType  string  `json:"task_type"`
	Duration  float64 `json:"duration_ms"`
	Success   bool    `json:"success"`
}

type ResourceUsage struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	TokenCost   float64 `json:"token_cost"`
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(registry *AgentRegistry) *PerformanceMonitor {
	return &PerformanceMonitor{
		registry: registry,
		metrics:  make(map[string]*AgentMetrics),
	}
}

// RecordTask records a task execution
func (pm *PerformanceMonitor) RecordTask(ctx context.Context, agentID string, taskType string, duration float64, success bool) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	metrics, ok := pm.metrics[agentID]
	if !ok {
		metrics = &AgentMetrics{
			AgentID:      agentID,
			UsagePattern: []UsagePoint{},
		}
		pm.metrics[agentID] = metrics
	}

	// Update metrics
	metrics.TaskCount++
	metrics.LastUsed = time.Now().Unix()

	if success {
		metrics.SuccessCount++
	} else {
		metrics.FailureCount++
	}

	// Update average duration
	metrics.AvgDuration = (metrics.AvgDuration*float64(metrics.TaskCount-1) + duration) / float64(metrics.TaskCount)

	// Update error rate
	metrics.ErrorRate = float64(metrics.FailureCount) / float64(metrics.TaskCount)

	// Add usage point
	metrics.UsagePattern = append(metrics.UsagePattern, UsagePoint{
		Timestamp: time.Now().Unix(),
		TaskType:  taskType,
		Duration:  duration,
		Success:   success,
	})

	// Update agent stats in registry
	pm.updateAgentStats(agentID, metrics)

	return nil
}

// GetMetrics gets metrics for an agent
func (pm *PerformanceMonitor) GetMetrics(ctx context.Context, agentID string) (*AgentMetrics, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	metrics, ok := pm.metrics[agentID]
	if !ok {
		return nil, fmt.Errorf("metrics not found for agent: %s", agentID)
	}

	return metrics, nil
}

// GetLeaderboard returns agent leaderboard
func (pm *PerformanceMonitor) GetLeaderboard(ctx context.Context, sortBy string) ([]*AgentMetrics, error) {
	pm.mu.RLock()
	defer pm.RUnlock()

	metricsList := make([]*AgentMetrics, 0, len(pm.metrics))
	for _, metrics := range pm.metrics {
		metricsList = append(metricsList, metrics)
	}

	// Sort by specified metric
	switch sortBy {
	case "quality":
		sort.Slice(metricsList, func(i, j int) bool {
			return metricsList[i].QualityScore > metricsList[j].QualityScore
		})
	case "success_rate":
		sort.Slice(metricsList, func(i, j int) bool {
			return (metricsList[i].SuccessCount / metricsList[i].TaskCount) >
				(metricsList[j].SuccessCount / metricsList[j].TaskCount)
		})
	default:
		sort.Slice(metricsList, func(i, j int) bool {
			return metricsList[i].TaskCount > metricsList[j].TaskCount
		})
	}

	return metricsList, nil
}
```

---

### 6. Agent Learning

**Purpose**: Enable agents học từ experience và share knowledge

**File**: `apps/ticrew/internal/learning/learning.go`

```go
package learning

import (
	"context"
	"encoding/json"
)

// AgentLearning manages agent learning
type AgentLearning struct {
	knowledgeGraph *KnowledgeGraphClient
	patternDB      *PatternDatabase
	llm            LLMClient
}

// Pattern represents a learned pattern
type Pattern struct {
	ID          string            `json:"id"`
	AgentID     string            `json:"agent_id"`
	Type        string            `json:"type"`        // success-pattern, failure-pattern, optimization
	Context     string            `json:"context"`
	Description string            `json:"description"`
	Confidence  float64           `json:"confidence"`
	UsageCount  int64             `json:"usage_count"`
	LastUsed    int64             `json:"last_used"`
}

// NewAgentLearning creates a new agent learning system
func NewAgentLearning(kg *KnowledgeGraphClient, patternDB *PatternDatabase, llm LLMClient) *AgentLearning {
	return &AgentLearning{
		knowledgeGraph: kg,
		patternDB:      patternDB,
		llm:            llm,
	}
}

// LearnFromSuccess learns from successful task execution
func (al *AgentLearning) LearnFromSuccess(ctx context.Context, agentID, task, output string) error {
	// Extract pattern using LLM
	pattern, err := al.extractPattern(ctx, agentID, task, output)
	if err != nil {
		return err
	}

	pattern.Type = "success-pattern"
	pattern.Confidence = 0.8

	// Store in pattern database
	return al.patternDB.Store(ctx, pattern)
}

// LearnFromFailure learns from failed task execution
func (al *AgentLearning) LearnFromFailure(ctx context.Context, agentID, task, error string) error {
	// Extract pattern using LLM
	pattern, err := al.extractPattern(ctx, agentID, task, error)
	if err != nil {
		return err
	}

	pattern.Type = "failure-pattern"
	pattern.Confidence = 0.7

	// Store in pattern database
	return al.patternDB.Store(ctx, pattern)
}

// GetPatterns gets patterns for an agent
func (al *AgentLearning) GetPatterns(ctx context.Context, agentID string) ([]Pattern, error) {
	return al.patternDB.GetByAgent(ctx, agentID)
}

// SharePattern shares a pattern with other agents
func (al *AgentLearning) SharePattern(ctx context.Context, patternID string, targetAgentIDs []string) error {
	// Get pattern
	pattern, err := al.patternDB.Get(ctx, patternID)
	if err != nil {
		return err
	}

	// Store pattern for target agents
	for _, targetID := range targetAgentIDs {
		patternCopy := pattern
		patternCopy.AgentID = targetID
		if err := al.patternDB.Store(ctx, patternCopy); err != nil {
			return err
		}
	}

	return nil
}
```

---

### 7. Agent Communicator

**Purpose**: Enable communication giữa agents

**File**: `apps/ticrew/internal/communicator/communicator.go`

```go
package communicator

import (
	"context"
)

// AgentCommunicator manages agent communication
type AgentCommunicator struct {
	registry *AgentRegistry
	llm      LLMClient
}

// Message represents a message between agents
type Message struct {
	From    string      `json:"from"`
	To      string      `json:"to"`
	Type    MessageType `json:"type"`
	Content string      `json:"content"`
	Context interface{} `json:"context"`
	Timestamp int64       `json:"timestamp"`
}

type MessageType string

const (
	MessageRequest     MessageType = "request"
	MessageResponse    MessageType = "response"
	MessageNotification MessageType = "notification"
	MessageDelegate   MessageType = "delegate"
)

// NewAgentCommunicator creates a new agent communicator
func NewAgentCommunicator(registry *AgentRegistry, llm LLMClient) *AgentCommunicator {
	return &AgentCommunicator{
		registry: registry,
		llm:      llm,
	}
}

// Send sends a message to an agent
func (ac *AgentCommunicator) Send(ctx context.Context, from, to string, message Message) error {
	// Validate agents exist
	if _, err := ac.registry.Get(ctx, from); err != nil {
		return err
	}
	if _, err := ac.registry.Get(ctx, to); err != nil {
		return err
	}

	message.From = from
	message.To = to
	message.Timestamp = time.Now().Unix()

	// Deliver message
	return ac.deliver(ctx, message)
}

// Broadcast broadcasts a message to multiple agents
func (ac *AgentCommunicator) Broadcast(ctx context.Context, from string, to []string, message Message) error {
	for _, recipient := range to {
		if err := ac.Send(ctx, from, recipient, message); err != nil {
			// Log error but continue
		}
	}
	return nil
}

// deliver delivers a message to an agent
func (ac *AgentCommunicator) deliver(ctx context.Context, message Message) error {
	// Get agent
	agent, err := ac.registry.Get(ctx, message.To)
	if err != nil {
		return err
	}

	// Load agent definition
	definition, err := ac.loadAgentDefinition(agent.FilePath)
	if err != nil {
		return err
	}

	// Process message with agent
	return ac.processWithAgent(ctx, definition, message)
}
```

---

### 8. Agent Lifecycle

**Purpose**: Manage agent lifecycle (start, stop, scale)

**File**: `apps/ticrew/internal/lifecycle/lifecycle.go`

```go
package lifecycle

import (
	"context"
)

// AgentLifecycle manages agent lifecycle
type AgentLifecycle struct {
	registry *AgentRegistry
	instances map[string]*AgentInstance
	mu       sync.RWMutex
}

type AgentInstance struct {
	InstanceID string
	AgentID    string
	Status     InstanceStatus
	StartedAt  int64
	LastActive int64
}

type InstanceStatus string

const (
	InstanceRunning   InstanceStatus = "running"
	InstanceIdle      InstanceStatus = "idle"
	InstanceStopped  InstanceStatus = "stopped"
	InstanceError    InstanceStatus = "error"
)

// NewAgentLifecycle creates a new agent lifecycle manager
func NewAgentLifecycle(registry *AgentRegistry) *AgentLifecycle {
	return &AgentLifecycle{
		registry:  registry,
		instances: make(map[string]*AgentInstance),
	}
}

// Start starts an agent instance
func (al *AgentLifecycle) Start(ctx context.Context, agentID string) (*AgentInstance, error) {
	al.mu.Lock()
	defer al.mu.Unlock()

	// Check if agent exists
	_, err := al.registry.Get(ctx, agentID)
	if err != nil {
		return nil, err
	}

	// Create instance
	instance := &AgentInstance{
		InstanceID: generateInstanceID(),
		AgentID:    agentID,
		Status:     InstanceRunning,
		StartedAt:  time.Now().Unix(),
		LastActive: time.Now().Unix(),
	}

	al.instances[instance.InstanceID] = instance

	return instance, nil
}

// Stop stops an agent instance
func (al *AgentLifecycle) Stop(ctx context.Context, instanceID string) error {
	al.mu.Lock()
	defer al.mu.Unlock()

	instance, ok := al.instances[instanceID]
	if !ok {
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	instance.Status = InstanceStopped
	delete(al.instances, instanceID)

	return nil
}

// Scale scales agent instances based on load
func (al *AgentLifecycle) Scale(ctx context.Context, agentID string, targetCount int) error {
	al.mu.Lock()
	defer al.mu.Unlock()

	currentCount := 0
	for _, instance := range al.instances {
		if instance.AgentID == agentID && instance.Status == InstanceRunning {
			currentCount++
		}
	}

	// Scale up
	if targetCount > currentCount {
		for i := 0; i < targetCount-currentCount; i++ {
			al.Start(ctx, agentID)
		}
	}

	// Scale down
	if targetCount < currentCount {
		stopped := 0
		for _, instance := range al.instances {
			if instance.AgentID == agentID && instance.Status == InstanceRunning {
				if stopped < currentCount-targetCount {
					al.Stop(ctx, instance.InstanceID)
					stopped++
				}
			}
		}
	}

	return nil
}

// GetInstances gets instances for an agent
func (al *AgentLifecycle) GetInstances(ctx context.Context, agentID string) ([]*AgentInstance, error) {
	al.mu.RLock()
	defer al.mu.RUnlock()

	instances := []*AgentInstance{}
	for _, instance := range al.instances {
		if instance.AgentID == agentID {
			instances = append(instances, instance)
		}
	}

	return instances, nil
}
```

---

## 🚀 Implementation Plan

### Phase 1: Foundation (Week 1-2)
1. Implement Agent Registry
2. Implement Task Router (simple scoring)
3. Implement Agent Lifecycle
4. Add unit tests

### Phase 2: Core Features (Week 3-4)
1. Implement Agent Orchestrator
2. Implement Performance Monitor
3. Integrate with existing Beads system
4. Integrate with Knowledge Graph Memory

### Phase 3: Advanced Features (Week 5-6)
1. Implement Agent Marketplace
2. Implement Agent Learning
3. Implement Agent Communicator
4. Add LLM integration for routing/learning

### Phase 4: UI & Tools (Week 7-8)
1. Build CLI tool for framework management
2. Build web UI for agent marketplace
3. Build monitoring dashboard
4. Add analytics and reporting

---

## 📊 Feasibility Assessment

### ✅ Can Build With Current Resources

**Strengths**:
- 104 agents already defined (YAML + markdown)
- Existing infrastructure (Beads, MCP, Knowledge Graph)
- Go-based (Ti project is Go-heavy)
- Team has experience with agent systems

**Challenges**:
- LLM integration (need API keys/credits)
- Complex orchestration logic
- Performance at scale (104 agents)

### 📈 Estimated Effort

| Phase | Duration | Complexity |
|-------|----------|------------|
| Foundation | 2 weeks | Medium |
| Core Features | 2 weeks | High |
| Advanced Features | 2 weeks | Very High |
| UI & Tools | 2 weeks | Medium |
| **Total** | **8 weeks** | **High** |

---

## 🎯 Recommendation

**YES - Có thể xây dựng Ti Crew tương tự OpenClaw/Hermers với 104 agents!**

**Approach**:
1. **Sử dụng infrastructure đang có** (Beads, MCP, Knowledge Graph)
2. **Build incremental** (Foundation → Core → Advanced → UI)
3. **Leverage existing agent definitions** (104 agents đã có)
4. **Add LLM integration** (cho routing và learning)
8 tuần để build MVP với core features.

Bạn muốn tôi bắt đầu implement Phase 1 (Foundation) không?
