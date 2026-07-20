package main

// IntegrationManager handles agent orchestration and cross-brain integration
type IntegrationManager struct {
	hub *Hub
}

// NewIntegrationManager creates a new integration manager
func NewIntegrationManager(hub *Hub) *IntegrationManager {
	return &IntegrationManager{hub: hub}
}