package events

import (
	"time"
)

// Event type constants
const (
	// System events
	TypeSystemStartup  = "system.startup"
	TypeSystemShutdown = "system.shutdown"
	TypeSystemReady    = "system.ready"
	TypeSystemError    = "system.error"

	// Plugin events
	TypePluginLoaded   = "plugin.loaded"
	TypePluginUnloaded = "plugin.unloaded"
	TypePluginStarted  = "plugin.started"
	TypePluginStopped  = "plugin.stopped"
	TypePluginError    = "plugin.error"
	TypePluginUpdated  = "plugin.updated"

	// Agent events
	TypeAgentStarted   = "agent.started"
	TypeAgentStopped   = "agent.stopped"
	TypeAgentMessage   = "agent.message"
	TypeAgentError     = "agent.error"
	TypeAgentCompleted = "agent.completed"

	// Automation events
	TypeAutomationStarted   = "automation.started"
	TypeAutomationCompleted = "automation.completed"
	TypeAutomationFailed    = "automation.failed"
	TypeAutomationProgress  = "automation.progress"

	// Configuration events
	TypeConfigChanged  = "config.changed"
	TypeConfigReloaded = "config.reloaded"
	TypeConfigError    = "config.error"

	// Health events
	TypeHealthCheck    = "health.check"
	TypeHealthStatus   = "health.status"
	TypeHealthDegraded = "health.degraded"

	// Network events
	TypeNetworkConnected    = "network.connected"
	TypeNetworkDisconnected = "network.disconnected"
	TypeNetworkError        = "network.error"

	// Resource events
	TypeResourceWarning  = "resource.warning"
	TypeResourceCritical = "resource.critical"
)

// Event source constants
const (
	SourceSystem     = "system"
	SourcePlugin     = "plugin"
	SourceAgent      = "agent"
	SourceAutomation = "automation"
	SourceConfig     = "config"
	SourceHealth     = "health"
	SourceNetwork    = "network"
	SourceUser       = "user"
)

// Event metadata keys
const (
	MetadataPluginName    = "plugin.name"
	MetadataPluginVersion = "plugin.version"
	MetadataAgentID       = "agent.id"
	MetadataTaskID        = "task.id"
	MetadataErrorType     = "error.type"
	MetadataErrorCode     = "error.code"
	MetadataSeverity      = "severity"
	MetadataComponent     = "component"
	MetadataResourceType  = "resource.type"
	MetadataResourceValue = "resource.value"
	MetadataSource        = "source"
)

// Severity levels
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityError    = "error"
	SeverityCritical = "critical"
)

// PluginEventPayload represents a plugin event payload
type PluginEventPayload struct {
	PluginName    string
	PluginVersion string
	PluginPath    string
	State         string
	Error         error
}

// AgentEventPayload represents an agent event payload
type AgentEventPayload struct {
	AgentID   string
	AgentType string
	Message   string
	State     string
	Tasks     []string
	Error     error
}

// AutomationEventPayload represents an automation event payload
type AutomationEventPayload struct {
	AutomationID string
	TaskID       string
	TaskName     string
	Status       string
	Progress     float64
	Result       interface{}
	Error        error
}

// HealthEventPayload represents a health event payload
type HealthEventPayload struct {
	Component string
	Status    string
	Healthy   bool
	Details   map[string]interface{}
	CheckTime time.Time
}

// ConfigEventPayload represents a configuration event payload
type ConfigEventPayload struct {
	ConfigPath string
	ChangeType string
	OldValue   interface{}
	NewValue   interface{}
	ChangedBy  string
}

// ResourceEventPayload represents a resource event payload
type ResourceEventPayload struct {
	ResourceType string
	ResourceName string
	CurrentValue float64
	Threshold    float64
	Unit         string
}
