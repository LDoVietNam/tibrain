package main

// CodeGraphService handles code graph operations
type CodeGraphService struct {
	hub *Hub
}

// NewCodeGraphService creates a new code graph service
func NewCodeGraphService(hub *Hub) *CodeGraphService {
	return &CodeGraphService{hub: hub}
}

// BuildOrUpdateGraph builds or updates the code graph
func (c *CodeGraphService) BuildOrUpdateGraph() error {
	return nil
}