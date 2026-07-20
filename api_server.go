package main

import (
	"database/sql"
	"net/http"
)

// APIServer handles extended REST API endpoints
type APIServer struct {
	hub         *Hub
	integration *IntegrationManager
}

// NewAPIServer creates a new API server
func NewAPIServer(hub *Hub, integration *IntegrationManager) *APIServer {
	return &APIServer{
		hub:         hub,
		integration: integration,
	}
}

// ServeHTTP implements http.Handler
func (a *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

// DB method placeholder to satisfy interface
func (a *APIServer) DB() *sql.DB {
	if a.hub != nil {
		return a.hub.db
	}
	return nil
}

// healthHandler is a simple health check
func (a *APIServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}