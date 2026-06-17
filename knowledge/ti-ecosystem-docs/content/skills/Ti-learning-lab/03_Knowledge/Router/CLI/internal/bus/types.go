package bus

// Topic constants
const (
	// Plugin lifecycle topics
	TopicPluginLoaded   = "plugin.loaded"
	TopicPluginUnloaded = "plugin.unloaded"
	TopicPluginStarted  = "plugin.started"
	TopicPluginStopped  = "plugin.stopped"
	TopicPluginError    = "plugin.error"

	// Agent topics
	TopicAgentStarted = "agent.started"
	TopicAgentStopped = "agent.stopped"
	TopicAgentMessage = "agent.message"
	TopicAgentError   = "agent.error"

	// Automation topics
	TopicAutomationStarted   = "automation.started"
	TopicAutomationCompleted = "automation.completed"
	TopicAutomationFailed    = "automation.failed"

	// System topics
	TopicSystemShutdown = "system.shutdown"
	TopicSystemError    = "system.error"
	TopicSystemReady    = "system.ready"

	// Health topics
	TopicHealthCheck  = "health.check"
	TopicHealthStatus = "health.status"

	// Config topics
	TopicConfigChanged  = "config.changed"
	TopicConfigReloaded = "config.reloaded"
)

// Metadata keys
const (
	MetadataPluginName    = "plugin.name"
	MetadataPluginVersion = "plugin.version"
	MetadataAgentID       = "agent.id"
	MetadataTaskID        = "task.id"
	MetadataErrorType     = "error.type"
	MetadataSource        = "source"
	MetadataTimestamp     = "timestamp"
)

// PluginEvent represents a plugin lifecycle event
type PluginEvent struct {
	PluginName    string
	PluginVersion string
	State         string
	Error         error
}

// AgentEvent represents an agent event
type AgentEvent struct {
	AgentID string
	Message string
	State   string
	Error   error
}

// AutomationEvent represents an automation event
type AutomationEvent struct {
	AutomationID string
	TaskID       string
	Status       string
	Result       interface{}
	Error        error
}

// HealthEvent represents a health check event
type HealthEvent struct {
	Component string
	Status    string
	Details   map[string]interface{}
}

// ConfigEvent represents a configuration change event
type ConfigEvent struct {
	ConfigPath string
	ChangeType string
	OldValue   interface{}
	NewValue   interface{}
}
