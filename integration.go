// TiBrain Integration Module
// Handles agent orchestration (Router Agent Brain merged into Ti Brain).
package main

import (
	"fmt"
	"log"
	"time"
)

// IntegrationManager handles agent orchestration within Ti Brain.
// Router Agent Brain is fully merged — no external brain communication needed.
type IntegrationManager struct {
	hub    *Hub
	agents map[string]*AgentClient
}

func NewIntegrationManager(hub *Hub) *IntegrationManager {
	return &IntegrationManager{
		hub:    hub,
		agents: make(map[string]*AgentClient),
	}
}

// AgentClient represents a connection to a TiCrew agent.
type AgentClient struct {
	ID       string
	Name     string
	Type     string
	Endpoint string
	Status   string
	LastSeen time.Time
}

// RegisterAgent registers a new agent with the integration manager.
func (im *IntegrationManager) RegisterAgent(agent *AgentClient) error {
	im.agents[agent.ID] = agent

	query := `
		INSERT OR REPLACE INTO agent_registry
		(id, name, type, endpoint, status, last_seen, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	_, err := im.hub.db.Exec(query,
		agent.ID, agent.Name, agent.Type, agent.Endpoint,
		agent.Status, agent.LastSeen.Unix(), now, now,
	)
	if err != nil {
		return fmt.Errorf("register agent: %w", err)
	}

	log.Printf("Registered agent: %s (%s)", agent.Name, agent.Type)
	return nil
}

// OrchestrateQuery orchestrates a query within the unified Ti Brain.
// Router Agent Brain is merged — all queries use the local knowledge base.
func (im *IntegrationManager) OrchestrateQuery(query string) (*OrchestrationResponse, error) {
	response := &OrchestrationResponse{
		Query:     query,
		Timestamp: time.Now(),
		Sources:   make(map[string]interface{}),
	}

	// Unified Ti Brain handles all queries locally
	if tiResult, err := im.QueryLocalBrain(query); err == nil {
		response.Sources["ti_brain"] = tiResult
		log.Printf("Ti Brain responded with %d results", len(tiResult.Results))
	}

	return response, nil
}

// QueryLocalBrain queries the local Ti Brain knowledge base.
func (im *IntegrationManager) QueryLocalBrain(query string) (*LocalBrainResponse, error) {
	return &LocalBrainResponse{
		Query:   query,
		Results: []LocalBrainResult{},
		Status:  "active",
	}, nil
}

type OrchestrationResponse struct {
	Query     string                 `json:"query"`
	Timestamp time.Time              `json:"timestamp"`
	Sources   map[string]interface{} `json:"sources"`
}

type LocalBrainResponse struct {
	Query   string             `json:"query"`
	Results []LocalBrainResult `json:"results"`
	Status  string             `json:"status"`
}

type LocalBrainResult struct {
	Content string  `json:"content"`
	Score   float64 `json:"score"`
	Source  string  `json:"source"`
}
