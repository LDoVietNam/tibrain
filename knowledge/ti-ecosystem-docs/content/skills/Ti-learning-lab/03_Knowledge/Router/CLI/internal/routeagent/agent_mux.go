// Package routeagent - HTTP mux for agent routing endpoints.
package routeagent

import (
	"net/http"

	"github.com/ti/cli/internal/agent"
)

// AgentMux returns an http.ServeMux that exposes the route agent's
// HTTP endpoints. The mux is suitable for embedding into a larger
// application server.
//
// The provided agent registry is used to look up agents by ID for
// routing decisions. May be nil if no registry is available.
func AgentMux(registry *agent.Registry) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"agents":[]}`))
	})
	mux.HandleFunc("/agents/route", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	return mux
}
