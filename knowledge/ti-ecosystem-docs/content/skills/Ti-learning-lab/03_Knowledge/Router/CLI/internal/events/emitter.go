package events

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Emitter provides a convenient interface for emitting events
type Emitter struct {
	dispatcher *Dispatcher
	source     string
	mu         sync.RWMutex
}

// NewEmitter creates a new event emitter
func NewEmitter(dispatcher *Dispatcher, source string) *Emitter {
	return &Emitter{
		dispatcher: dispatcher,
		source:     source,
	}
}

// Emit emits an event
func (e *Emitter) Emit(ctx context.Context, eventType string, payload interface{}, metadata map[string]string) error {
	if metadata == nil {
		metadata = make(map[string]string)
	}

	metadata[MetadataSource] = e.source

	event := Event{
		Type:     eventType,
		Source:   e.source,
		Payload:  payload,
		Metadata: metadata,
	}

	return e.dispatcher.Publish(ctx, event)
}

// EmitPluginLoaded emits a plugin loaded event
func (e *Emitter) EmitPluginLoaded(ctx context.Context, name, version string) error {
	payload := PluginEventPayload{
		PluginName:    name,
		PluginVersion: version,
		State:         "loaded",
	}

	metadata := map[string]string{
		MetadataPluginName:    name,
		MetadataPluginVersion: version,
	}

	return e.Emit(ctx, TypePluginLoaded, payload, metadata)
}

// EmitPluginError emits a plugin error event
func (e *Emitter) EmitPluginError(ctx context.Context, name string, err error) error {
	payload := PluginEventPayload{
		PluginName: name,
		State:      "error",
		Error:      err,
	}

	metadata := map[string]string{
		MetadataPluginName: name,
		MetadataErrorType:  "plugin_error",
		MetadataSeverity:   SeverityError,
	}

	return e.Emit(ctx, TypePluginError, payload, metadata)
}

// EmitAgentStarted emits an agent started event
func (e *Emitter) EmitAgentStarted(ctx context.Context, agentID, agentType string) error {
	payload := AgentEventPayload{
		AgentID:   agentID,
		AgentType: agentType,
		State:     "started",
	}

	metadata := map[string]string{
		MetadataAgentID: agentID,
	}

	return e.Emit(ctx, TypeAgentStarted, payload, metadata)
}

// EmitAgentError emits an agent error event
func (e *Emitter) EmitAgentError(ctx context.Context, agentID string, err error) error {
	payload := AgentEventPayload{
		AgentID: agentID,
		State:   "error",
		Error:   err,
	}

	metadata := map[string]string{
		MetadataAgentID:   agentID,
		MetadataErrorType: "agent_error",
		MetadataSeverity:  SeverityError,
	}

	return e.Emit(ctx, TypeAgentError, payload, metadata)
}

// EmitAutomationStarted emits an automation started event
func (e *Emitter) EmitAutomationStarted(ctx context.Context, automationID, taskID, taskName string) error {
	payload := AutomationEventPayload{
		AutomationID: automationID,
		TaskID:       taskID,
		TaskName:     taskName,
		Status:       "started",
		Progress:     0.0,
	}

	metadata := map[string]string{
		MetadataTaskID: taskID,
	}

	return e.Emit(ctx, TypeAutomationStarted, payload, metadata)
}

// EmitAutomationCompleted emits an automation completed event
func (e *Emitter) EmitAutomationCompleted(ctx context.Context, automationID, taskID string, result interface{}) error {
	payload := AutomationEventPayload{
		AutomationID: automationID,
		TaskID:       taskID,
		Status:       "completed",
		Progress:     100.0,
		Result:       result,
	}

	metadata := map[string]string{
		MetadataTaskID: taskID,
	}

	return e.Emit(ctx, TypeAutomationCompleted, payload, metadata)
}

// EmitAutomationFailed emits an automation failed event
func (e *Emitter) EmitAutomationFailed(ctx context.Context, automationID, taskID string, err error) error {
	payload := AutomationEventPayload{
		AutomationID: automationID,
		TaskID:       taskID,
		Status:       "failed",
		Error:        err,
	}

	metadata := map[string]string{
		MetadataTaskID:    taskID,
		MetadataErrorType: "automation_error",
		MetadataSeverity:  SeverityError,
	}

	return e.Emit(ctx, TypeAutomationFailed, payload, metadata)
}

// EmitHealthStatus emits a health status event
func (e *Emitter) EmitHealthStatus(ctx context.Context, component string, healthy bool, details map[string]interface{}) error {
	payload := HealthEventPayload{
		Component: component,
		Status:    map[bool]string{true: "healthy", false: "unhealthy"}[healthy],
		Healthy:   healthy,
		Details:   details,
		CheckTime: time.Now(),
	}

	severity := SeverityInfo
	if !healthy {
		severity = SeverityWarning
	}

	metadata := map[string]string{
		MetadataComponent: component,
		MetadataSeverity:  severity,
	}

	return e.Emit(ctx, TypeHealthStatus, payload, metadata)
}

// EmitConfigChanged emits a configuration changed event
func (e *Emitter) EmitConfigChanged(ctx context.Context, configPath, changeType string, oldValue, newValue interface{}, changedBy string) error {
	payload := ConfigEventPayload{
		ConfigPath: configPath,
		ChangeType: changeType,
		OldValue:   oldValue,
		NewValue:   newValue,
		ChangedBy:  changedBy,
	}

	metadata := map[string]string{
		MetadataSource: changedBy,
	}

	return e.Emit(ctx, TypeConfigChanged, payload, metadata)
}

// SetSource sets the source for emitted events
func (e *Emitter) SetSource(source string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.source = source
}

// GetSource returns the current source
func (e *Emitter) GetSource() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.source
}

// EventBuilder provides a fluent interface for building events
type EventBuilder struct {
	event      Event
	dispatcher *Dispatcher
}

// NewEventBuilder creates a new event builder
func NewEventBuilder(dispatcher *Dispatcher) *EventBuilder {
	return &EventBuilder{
		event: Event{
			Timestamp: time.Now(),
			Metadata:  make(map[string]string),
		},
		dispatcher: dispatcher,
	}
}

// Type sets the event type
func (b *EventBuilder) Type(eventType string) *EventBuilder {
	b.event.Type = eventType
	return b
}

// Source sets the event source
func (b *EventBuilder) Source(source string) *EventBuilder {
	b.event.Source = source
	return b
}

// Payload sets the event payload
func (b *EventBuilder) Payload(payload interface{}) *EventBuilder {
	b.event.Payload = payload
	return b
}

// Metadata adds metadata to the event
func (b *EventBuilder) Metadata(key, value string) *EventBuilder {
	b.event.Metadata[key] = value
	return b
}

// Build builds the event
func (b *EventBuilder) Build() Event {
	if b.event.ID == "" {
		b.event.ID = generateEventID()
	}
	return b.event
}

// Publish publishes the built event
func (b *EventBuilder) Publish(ctx context.Context) error {
	event := b.Build()
	return b.dispatcher.Publish(ctx, event)
}

// MustPublish publishes the event and panics on error
func (b *EventBuilder) MustPublish(ctx context.Context) {
	if err := b.Publish(ctx); err != nil {
		panic(fmt.Sprintf("failed to publish event: %v", err))
	}
}
