package main

import (
	"fmt"
	"strings"
	"sync"
)

// AgentContextSource carries metadata from upstream agents into the RAG ingest path.
type AgentContextSource struct {
	AgentID     string                 `json:"agent_id"`
	Port        int                    `json:"port"`
	SessionType string                 `json:"session_type"`
	Extra       map[string]interface{} `json:"extra"`
}

// RAGPayload is the normalized ingest payload accepted by the brain learning plane.
type RAGPayload struct {
	Content       string             `json:"content"`
	ContextSource AgentContextSource `json:"context_source"`
}

// RuleProcessor attaches tags and metadata to incoming agent context.
type RuleProcessor func(ctx AgentContextSource) (tags []string, metadata map[string]interface{})

type TagRuleEngine struct {
	mu    sync.RWMutex
	rules map[string]RuleProcessor
}

func NewTagRuleEngine() *TagRuleEngine {
	engine := &TagRuleEngine{
		rules: make(map[string]RuleProcessor),
	}
	engine.registerCoreRules()
	return engine
}

func (e *TagRuleEngine) RegisterRule(agentID string, processor RuleProcessor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules[agentID] = processor
}

func (e *TagRuleEngine) registerCoreRules() {
	e.RegisterRule("omniroute_cli", func(ctx AgentContextSource) ([]string, map[string]interface{}) {
		tags := []string{"scope:local_cli"}
		meta := map[string]interface{}{"port": ctx.Port, "session": ctx.SessionType}
		return tags, meta
	})

	e.RegisterRule("security_agent", func(ctx AgentContextSource) ([]string, map[string]interface{}) {
		tags := []string{"scope:global_security", "priority:high"}
		meta := map[string]interface{}{"trigger": "anomaly_detect"}
		return tags, meta
	})
}

func (e *TagRuleEngine) Process(ctx AgentContextSource) ([]string, map[string]interface{}) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if processor, exists := e.rules[ctx.AgentID]; exists {
		return processor(ctx)
	}
	return []string{"scope:unclassified"}, map[string]interface{}{"warn": "unknown_agent"}
}

// BuildScopeQuery constrains retrieval when an execution client is explicitly scoped.
// Empty scopes mean learning-plane discovery is allowed to search the full corpus.
func BuildScopeQuery(allowedScopes []string) string {
	if len(allowedScopes) == 0 {
		return ""
	}

	if allowedScopes[0] == "GLOBAL" {
		return ""
	}

	var filters []string
	for _, scope := range allowedScopes {
		cleanScope := strings.ReplaceAll(scope, "'", "''")
		filters = append(filters, fmt.Sprintf("tags LIKE '%%%s%%'", cleanScope))
	}

	return fmt.Sprintf("AND (%s)", strings.Join(filters, " OR "))
}
