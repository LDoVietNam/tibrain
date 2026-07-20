package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ti/router/tibrain/internal/execution"
	"github.com/ti/router/tibrain/internal/memory"
	"github.com/ti/router/tibrain/internal/prediction"
	"github.com/ti/router/tibrain/internal/ports"
	"github.com/ti/router/tibrain/internal/verification"
)

// AgentRuntime manages agent execution
type AgentRuntime struct {
	predictionEngine *prediction.PredictionEngine
	executor         *execution.Engine
	verifier         *verification.SchemaValidator
	memoryStore      *memory.MemoryStore
	tracer           *execution.InMemoryTracer
}

// Agent represents an AI agent
type Agent struct {
	ID       string
	Name     string
	Type     string
	Capabilities []string
	Status   string
	LastSeen time.Time
}

// NewAgentRuntime creates a new agent runtime
func NewAgentRuntime() *AgentRuntime {
	tracer := execution.NewInMemoryTracer()
	verifier := verification.NewSchemaValidator()

	return &AgentRuntime{
		predictionEngine: prediction.NewPredictionEngine(),
		executor:         execution.NewEngine(&schemaVerifierAdapter{inner: verifier}, tracer),
		verifier:         verifier,
		memoryStore:      memory.NewMemoryStore(),
		tracer:           tracer,
	}
}

// RunTask executes a task through the agent runtime
func (r *AgentRuntime) RunTask(ctx context.Context, task string) (*execution.ExecutionContext, error) {
	// Create execution context
	execCtx := &execution.ExecutionContext{
		ExecutionID:   fmt.Sprintf("exec-%d", time.Now().UnixNano()),
		TaskID:        fmt.Sprintf("task-%d", time.Now().UnixNano()),
		UserGoal:      task,
		State:         execution.StateReceived,
		Attempt:       1,
		MaxAttempts:   3,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Execute through state machine
	if err := r.executor.Execute(ctx, execCtx); err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	return execCtx, nil
}

// AddModel adds a model to the prediction engine
func (r *AgentRuntime) AddModel(model prediction.Model) {
	r.predictionEngine.AddModel(model)
}

// AddTool adds a tool to the prediction engine
func (r *AgentRuntime) AddTool(tool prediction.Tool) {
	r.predictionEngine.AddTool(tool)
}

// AddSkill adds a skill to the prediction engine
func (r *AgentRuntime) AddSkill(skill prediction.Skill) {
	r.predictionEngine.AddSkill(skill)
}

// GetMemoryStore returns the memory store
func (r *AgentRuntime) GetMemoryStore() *memory.MemoryStore {
	return r.memoryStore
}

// GetTracer returns the tracer
func (r *AgentRuntime) GetTracer() *execution.InMemoryTracer {
	return r.tracer
}

// AgentManager manages multiple agents
type AgentManager struct {
	mu    sync.RWMutex
	agents map[string]*Agent
}

// NewAgentManager creates a new agent manager
func NewAgentManager() *AgentManager {
	return &AgentManager{
		agents: make(map[string]*Agent),
	}
}

// RegisterAgent registers a new agent
func (m *AgentManager) RegisterAgent(agent *Agent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents[agent.ID] = agent
}

// GetAgent retrieves an agent by ID
func (m *AgentManager) GetAgent(id string) (*Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, ok := m.agents[id]
	if !ok {
		return nil, fmt.Errorf("agent not found: %s", id)
	}
	return agent, nil
}

// ListAgents lists all registered agents
func (m *AgentManager) ListAgents() []*Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]*Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	return agents
}

// UnregisterAgent removes an agent
func (m *AgentManager) UnregisterAgent(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.agents[id]; !ok {
		return fmt.Errorf("agent not found: %s", id)
	}
	delete(m.agents, id)
	return nil
}

// ToolExecutor implements ports.ToolPort
type ToolExecutor struct {
	runtime *AgentRuntime
}

// NewToolExecutor creates a tool executor
func NewToolExecutor(runtime *AgentRuntime) *ToolExecutor {
	return &ToolExecutor{runtime: runtime}
}

// Execute executes a tool request
func (e *ToolExecutor) Execute(ctx context.Context, req ports.ToolRequest) (ports.ToolResult, error) {
	// Store execution in memory
	entry := map[string]interface{}{
		"tool_name": req.ToolName,
		"params":    req.Params,
		"executed":  time.Now().Unix(),
	}
	e.runtime.GetMemoryStore().Store(ctx, req.ToolName, entry, nil)

	return ports.ToolResult{
		Success: true,
		Data:    map[string]interface{}{"result": "tool executed"},
	}, nil
}

// PromptExecutor implements ports.PromptPort
type PromptExecutor struct {
	runtime *AgentRuntime
}

// NewPromptExecutor creates a prompt executor
func NewPromptExecutor(runtime *AgentRuntime) *PromptExecutor {
	return &PromptExecutor{runtime: runtime}
}

// Preflight returns a prompt envelope
func (e *PromptExecutor) Preflight(ctx context.Context, intent, domain string) (*ports.PromptEnvelope, error) {
	result, err := e.runtime.predictionEngine.Predict(ctx, fmt.Sprintf("%s: %s", intent, domain))
	if err != nil {
		return nil, err
	}

	return &ports.PromptEnvelope{
		ID:          result.Skill.ID,
		Name:        result.Skill.Name,
		Description: fmt.Sprintf("Prompt for %s", intent),
		Domain:      domain,
		Intent:      []string{intent},
		Constraints: []string{},
	}, nil
}

// RecordFeedback records feedback
func (e *PromptExecutor) RecordFeedback(ctx context.Context, promptID string, rating int, note string) error {
	// Store feedback in memory
	e.runtime.GetMemoryStore().Store(ctx, fmt.Sprintf("feedback:%s", promptID), map[string]interface{}{
		"rating": rating,
		"note":   note,
	}, nil)
	return nil
}

// schemaVerifierAdapter adapts verification.SchemaValidator (Verify(ctx,
// interface{})) to the execution.Verifier interface (Verify(ctx,
// *execution.ExecutionContext)) without mutating the original signature.
type schemaVerifierAdapter struct {
	inner *verification.SchemaValidator
}

func (a *schemaVerifierAdapter) Verify(ctx context.Context, exec *execution.ExecutionContext) (*execution.VerificationResult, error) {
	res, err := a.inner.Verify(ctx, exec)
	if err != nil {
		return nil, err
	}
	return &execution.VerificationResult{
		Passed:   res.Passed,
		Feedback: res.Feedback,
		Errors:   res.Errors,
	}, nil
}