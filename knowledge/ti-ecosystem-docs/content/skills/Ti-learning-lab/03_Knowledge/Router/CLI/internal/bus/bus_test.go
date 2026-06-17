package bus

import (
	"context"
	"testing"
	"time"
)

func TestNewEventBus(t *testing.T) {
	bus := NewEventBus(100)
	if bus == nil {
		t.Fatal("Expected non-nil bus")
	}
	if bus.closed {
		t.Error("Expected bus to be open")
	}
}

func TestPublishSubscribe(t *testing.T) {
	bus := NewEventBus(100)
	defer bus.Close()

	received := make(chan Message, 1)
	handler := func(ctx context.Context, msg Message) error {
		received <- msg
		return nil
	}

	subID, err := bus.Subscribe("test.topic", handler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	if subID == "" {
		t.Error("Expected non-empty subscription ID")
	}

	err = bus.Publish(context.Background(), "test.topic", "test payload", nil)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	select {
	case msg := <-received:
		if msg.Topic != "test.topic" {
			t.Errorf("Expected topic 'test.topic', got '%s'", msg.Topic)
		}
		if msg.Payload != "test payload" {
			t.Errorf("Expected payload 'test payload', got '%v'", msg.Payload)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestSubscribeWithFilter(t *testing.T) {
	bus := NewEventBus(100)
	defer bus.Close()

	received := make(chan Message, 1)
	filter := func(msg Message) bool {
		if metadata, ok := msg.Metadata["key"]; ok && metadata == "value" {
			return true
		}
		return false
	}

	handler := func(ctx context.Context, msg Message) error {
		received <- msg
		return nil
	}

	_, err := bus.SubscribeWithFilter("test.topic", handler, filter)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish message that matches filter
	metadata := map[string]string{"key": "value"}
	err = bus.Publish(context.Background(), "test.topic", "test payload", metadata)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	select {
	case <-received:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := NewEventBus(100)
	defer bus.Close()

	handler := func(ctx context.Context, msg Message) error {
		return nil
	}

	subID, err := bus.Subscribe("test.topic", handler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	err = bus.Unsubscribe(subID)
	if err != nil {
		t.Fatalf("Failed to unsubscribe: %v", err)
	}

	// Try to unsubscribe again
	err = bus.Unsubscribe(subID)
	if err != ErrSubscriptionNotFound {
		t.Errorf("Expected ErrSubscriptionNotFound, got %v", err)
	}
}

func TestBusClosed(t *testing.T) {
	bus := NewEventBus(100)
	bus.Close()

	handler := func(ctx context.Context, msg Message) error {
		return nil
	}

	_, err := bus.Subscribe("test.topic", handler)
	if err != ErrBusClosed {
		t.Errorf("Expected ErrBusClosed, got %v", err)
	}

	err = bus.Publish(context.Background(), "test.topic", "test payload", nil)
	if err != ErrBusClosed {
		t.Errorf("Expected ErrBusClosed, got %v", err)
	}
}

func TestMultipleSubscribers(t *testing.T) {
	bus := NewEventBus(100)
	defer bus.Close()

	count := 0
	received := make(chan struct{}, 3)
	handler := func(ctx context.Context, msg Message) error {
		count++
		received <- struct{}{}
		return nil
	}

	// Subscribe 3 handlers
	for i := 0; i < 3; i++ {
		_, err := bus.Subscribe("test.topic", handler)
		if err != nil {
			t.Fatalf("Failed to subscribe: %v", err)
		}
	}

	err := bus.Publish(context.Background(), "test.topic", "test payload", nil)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	// Wait for all 3 handlers to receive the message
	for i := 0; i < 3; i++ {
		select {
		case <-received:
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for message")
		}
	}

	if count != 3 {
		t.Errorf("Expected 3 messages received, got %d", count)
	}
}

func TestMessageMetadata(t *testing.T) {
	bus := NewEventBus(100)
	defer bus.Close()

	received := make(chan Message, 1)
	handler := func(ctx context.Context, msg Message) error {
		received <- msg
		return nil
	}

	_, err := bus.Subscribe("test.topic", handler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	metadata := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	err = bus.Publish(context.Background(), "test.topic", "test payload", metadata)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	select {
	case msg := <-received:
		if msg.Metadata["key1"] != "value1" {
			t.Errorf("Expected metadata key1='value1', got '%s'", msg.Metadata["key1"])
		}
		if msg.Metadata["key2"] != "value2" {
			t.Errorf("Expected metadata key2='value2', got '%s'", msg.Metadata["key2"])
		}
		if msg.ID == "" {
			t.Error("Expected non-empty message ID")
		}
		if msg.Timestamp.IsZero() {
			t.Error("Expected non-zero timestamp")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message")
	}
}
