package events

import (
	"context"
	"testing"
	"time"
)

func TestNewEmitter(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	defer dispatcher.Close()

	emitter := NewEmitter(dispatcher, "test-source")
	if emitter == nil {
		t.Fatal("Expected non-nil emitter")
	}
	if emitter.source != "test-source" {
		t.Errorf("Expected source 'test-source', got '%s'", emitter.source)
	}
}

func TestEmitterEmit(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	defer dispatcher.Close()

	received := make(chan Event, 1)
	handler := func(ctx context.Context, event Event) error {
		received <- event
		return nil
	}

	dispatcher.Subscribe("test.event", handler)

	emitter := NewEmitter(dispatcher, "test-source")

	err := emitter.Emit(context.Background(), "test.event", "payload", nil)
	if err != nil {
		t.Fatalf("Failed to emit: %v", err)
	}

	select {
	case event := <-received:
		if event.Source != "test-source" {
			t.Errorf("Expected source 'test-source', got '%s'", event.Source)
		}
		if event.Metadata[MetadataSource] != "test-source" {
			t.Error("Expected source in metadata")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive event")
	}
}

func TestEmitterSetSource(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	defer dispatcher.Close()

	emitter := NewEmitter(dispatcher, "original-source")

	emitter.SetSource("new-source")

	if emitter.GetSource() != "new-source" {
		t.Errorf("Expected source 'new-source', got '%s'", emitter.GetSource())
	}
}

func TestEmitterEmitPluginLoaded(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	defer dispatcher.Close()

	received := make(chan Event, 1)
	handler := func(ctx context.Context, event Event) error {
		received <- event
		return nil
	}

	dispatcher.Subscribe(TypePluginLoaded, handler)

	emitter := NewEmitter(dispatcher, "test-source")

	err := emitter.EmitPluginLoaded(context.Background(), "test-plugin", "1.0.0")
	if err != nil {
		t.Fatalf("Failed to emit plugin loaded: %v", err)
	}

	select {
	case event := <-received:
		if event.Type != TypePluginLoaded {
			t.Errorf("Expected type '%s', got '%s'", TypePluginLoaded, event.Type)
		}
		payload, ok := event.Payload.(PluginEventPayload)
		if !ok {
			t.Error("Expected PluginEventPayload")
		}
		if payload.PluginName != "test-plugin" {
			t.Errorf("Expected plugin name 'test-plugin', got '%s'", payload.PluginName)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive event")
	}
}

func TestEmitterEmitAgentStarted(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	defer dispatcher.Close()

	received := make(chan Event, 1)
	handler := func(ctx context.Context, event Event) error {
		received <- event
		return nil
	}

	dispatcher.Subscribe(TypeAgentStarted, handler)

	emitter := NewEmitter(dispatcher, "test-source")

	err := emitter.EmitAgentStarted(context.Background(), "agent-123", "test-agent")
	if err != nil {
		t.Fatalf("Failed to emit agent started: %v", err)
	}

	select {
	case event := <-received:
		if event.Type != TypeAgentStarted {
			t.Errorf("Expected type '%s', got '%s'", TypeAgentStarted, event.Type)
		}
		payload, ok := event.Payload.(AgentEventPayload)
		if !ok {
			t.Error("Expected AgentEventPayload")
		}
		if payload.AgentID != "agent-123" {
			t.Errorf("Expected agent ID 'agent-123', got '%s'", payload.AgentID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive event")
	}
}

func TestEmitterEmitHealthStatus(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	defer dispatcher.Close()

	received := make(chan Event, 1)
	handler := func(ctx context.Context, event Event) error {
		received <- event
		return nil
	}

	dispatcher.Subscribe(TypeHealthStatus, handler)

	emitter := NewEmitter(dispatcher, "test-source")

	details := map[string]interface{}{"cpu": 50.0}
	err := emitter.EmitHealthStatus(context.Background(), "test-component", true, details)
	if err != nil {
		t.Fatalf("Failed to emit health status: %v", err)
	}

	select {
	case event := <-received:
		if event.Type != TypeHealthStatus {
			t.Errorf("Expected type '%s', got '%s'", TypeHealthStatus, event.Type)
		}
		payload, ok := event.Payload.(HealthEventPayload)
		if !ok {
			t.Error("Expected HealthEventPayload")
		}
		if payload.Component != "test-component" {
			t.Errorf("Expected component 'test-component', got '%s'", payload.Component)
		}
		if !payload.Healthy {
			t.Error("Expected healthy to be true")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive event")
	}
}

func TestEventBuilder(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	defer dispatcher.Close()

	received := make(chan Event, 1)
	handler := func(ctx context.Context, event Event) error {
		received <- event
		return nil
	}

	dispatcher.Subscribe("test.event", handler)

	builder := NewEventBuilder(dispatcher).
		Type("test.event").
		Source("test-source").
		Payload("test payload").
		Metadata("key", "value")

	err := builder.Publish(context.Background())
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	select {
	case event := <-received:
		if event.Type != "test.event" {
			t.Errorf("Expected type 'test.event', got '%s'", event.Type)
		}
		if event.Source != "test-source" {
			t.Errorf("Expected source 'test-source', got '%s'", event.Source)
		}
		if event.Payload != "test payload" {
			t.Errorf("Expected payload 'test payload', got '%v'", event.Payload)
		}
		if event.Metadata["key"] != "value" {
			t.Errorf("Expected metadata key='value', got '%s'", event.Metadata["key"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive event")
	}
}

func TestEventBuilderBuild(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	defer dispatcher.Close()

	builder := NewEventBuilder(dispatcher).
		Type("test.event").
		Source("test-source")

	event := builder.Build()

	if event.Type != "test.event" {
		t.Errorf("Expected type 'test.event', got '%s'", event.Type)
	}
	if event.Source != "test-source" {
		t.Errorf("Expected source 'test-source', got '%s'", event.Source)
	}
	if event.ID == "" {
		t.Error("Expected non-empty event ID")
	}
}
