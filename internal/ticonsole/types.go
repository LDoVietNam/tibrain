package ticonsole

import "time"

// ServiceState represents the runtime state observed from a live health probe.
// A configured endpoint is never treated as active until a probe succeeds.
type ServiceState string

const (
	StateHealthy  ServiceState = "healthy"
	StateDegraded ServiceState = "degraded"
	StateDown     ServiceState = "down"
	StateUnknown  ServiceState = "unknown"
)

// ServiceConfig describes one Ti platform service to probe.
type ServiceConfig struct {
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	BaseURL     string   `json:"baseUrl"`
	HealthPaths []string `json:"healthPaths"`
}

// ServiceSnapshot is the read-only runtime view consumed by the TUI.
type ServiceSnapshot struct {
	Name       string        `json:"name"`
	Role       string        `json:"role"`
	BaseURL    string        `json:"baseUrl"`
	Endpoint   string        `json:"endpoint,omitempty"`
	State      ServiceState  `json:"state"`
	StatusCode int           `json:"statusCode,omitempty"`
	Latency    time.Duration `json:"latency"`
	Detail     string        `json:"detail,omitempty"`
	CheckedAt  time.Time     `json:"checkedAt"`
}

// Summary aggregates the current service states.
type Summary struct {
	Healthy  int `json:"healthy"`
	Degraded int `json:"degraded"`
	Down     int `json:"down"`
	Unknown  int `json:"unknown"`
}

// Snapshot is one immutable dashboard refresh.
type Snapshot struct {
	GeneratedAt time.Time         `json:"generatedAt"`
	Services    []ServiceSnapshot `json:"services"`
	Summary     Summary           `json:"summary"`
	Notices     []string          `json:"notices,omitempty"`
}

func summarize(services []ServiceSnapshot) Summary {
	var result Summary
	for _, service := range services {
		switch service.State {
		case StateHealthy:
			result.Healthy++
		case StateDegraded:
			result.Degraded++
		case StateDown:
			result.Down++
		default:
			result.Unknown++
		}
	}
	return result
}
