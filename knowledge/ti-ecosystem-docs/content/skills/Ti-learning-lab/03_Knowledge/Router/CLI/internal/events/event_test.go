package events

import (
	"context"
	"testing"
	"time"
)

func TestNewDispatcher(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	if dispatcher == nil {
		t.Fatal("Expected non-nil dispatcher")
	}
	if dispatcher.workers != 5 {
		t.Errorf("Expected 5 workers, got %d", dispatcher.workers)
	}
}

func TestNewDispatcherDefaults(t *testing.T) {
	dispatcher := NewDispatcher(0, 0)
	if dispatcher.workers != 5 {
		t.Errorf("Expected default 5 workers, got %d", dispatcher.workers)
	}
}

func TestSubscribe(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	defer dispatcher.Close()

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	handlerID, err := dispatcher.Subscribe("test.event", handler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	if handlerID == "" {
		t.Error("Expected non-empty handler ID")
	}
}

func TestSubscribeWithFilter(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	defer dispatcher.Close()

	filter := func(event Event) bool {
		if metadata, ok := event.Metadata["key"]; ok && metadata == "value" {
			return true
		}
		return false
	}

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	handlerID, err := dispatcher.SubscribeWithFilter("test.event", handler, filter, 0)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	if handlerID == "" {
		t.Error("Expected non-empty handler ID")
	}
}

func TestUnsubscribe(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	defer dispatcher.Close()

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	handlerID, err := dispatcher.Subscribe("test.event", handler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	err = dispatcher.Unsubscribe(handlerID)
	if err != nil {
		t.Fatalf("Failed to unsubscribe: %v", err)
	}

	// Try to unsubscribe again
	err = dispatcher.Unsubscribe(handlerID)
	if err != ErrHandlerNotFound {
		t.Errorf("Expected ErrHandlerNotFound, got %v", err)
	}
}

func TestPublish(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	defer dispatcher.Close()

	received := make(chan Event, 1)
	handler := func(ctx context.Context, event Event) error {
		received <- event
		return nil
	}

	_, err := dispatcher.Subscribe("test.event", handler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	event := Event{
		Type:    "test.event",
		Payload: "test payload",
	}

	err = dispatcher.Publish(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	select {
	case receivedEvent := <-received:
		if receivedEvent.Type != "test.event" {
			t.Errorf("Expected type 'test.event', got '%s'", receivedEvent.Type)
		}
		if receivedEvent.Payload != "test payload" {
			t.Errorf("Expected payload 'test payload', got '%v'", receivedEvent.Payload)
		}
		if receivedEvent.ID == "" {
			t.Error("Expected non-empty event ID")
		}
		if receivedEvent.Timestamp.IsZero() {
			t.Error("Expected non-zero timestamp")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for event")
	}
}

func TestPublishWithFilter(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	defer dispatcher.Close()

	received := make(chan Event, 1)
	filter := func(event Event) bool {
		if metadata, ok := event.Metadata["key"]; ok && metadata == "value" {
			return true
		}
		return false
	}

	handler := func(ctx context.Context, event Event) error {
		received <- event
		return nil
	}

	_, err := dispatcher.SubscribeWithFilter("test.event", handler, filter, 0)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	metadata := map[string]string{"key": "value"}
	event := Event{
		Type:     "test.event",
		Payload:  "test payload",
		Metadata: metadata,
	}

	err = dispatcher.Publish(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	select {
	case <-received:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for event")
	}
}

func TestDispatcherClosed(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	dispatcher.Close()

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	_, err := dispatcher.Subscribe("test.event", handler)
	if err != ErrDispatcherClosed {
		t.Errorf("Expected ErrDispatcherClosed, got %v", err)
	}

	event := Event{
		Type:    "test.event",
		Payload: "test payload",
	}

	err = dispatcher.Publish(context.Background(), event)
	if err != ErrDispatcherClosed {
		t.Errorf("Expected ErrDispatcherClosed, got %v", err)
	}
}

func TestMultipleHandlers(t *testing.T) {
	dispatcher := NewDispatcher(100, 5)
	defer dispatcher.Close()

	count := 0
	received := make(chan struct{}, 3)
	handler := func(ctx context.Context, event Event) error {
		count++
		received <- struct{}{}
		return nil
	}

	// Subscribe 3 handlers
	for i := 0; i < 3; i++ {
		_, err := dispatcher.Subscribe("test.event", handler)
		if err != nil {
			t.Fatalf("Failed to subscribe: %v", err)
		}
	}

	event := Event{
		Type:    "test.event",
		Payload: "test payload",
	}

	err := dispatcher.Publish(context.Background(), event)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	// Wait for all 3 handlers to receive the event
	for i := 0; i < 3; i++ {
		select {
		case <-received:
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for event")
		}
	}

	if count != 3 {
		t.Errorf("Expected 3 handlers to be called, got %d", count)
	}
}
