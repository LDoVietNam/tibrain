package bus

import (
	"context"
	"testing"
	"time"
)

func TestNewRouter(t *testing.T) {
	router := NewRouter()
	if router == nil {
		t.Fatal("Expected non-nil router")
	}
	if len(router.rules) != 0 {
		t.Errorf("Expected empty rules, got %d", len(router.rules))
	}
}

func TestRouterAddRule(t *testing.T) {
	router := NewRouter()

	matcher := func(msg Message) bool {
		return msg.Topic == "test.topic"
	}

	handler := func(ctx context.Context, msg Message) error {
		return nil
	}

	router.AddRule(matcher, handler)

	if len(router.rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(router.rules))
	}
}

func TestRouterRoute(t *testing.T) {
	router := NewRouter()

	called := false
	matcher := func(msg Message) bool {
		return msg.Topic == "test.topic"
	}

	handler := func(ctx context.Context, msg Message) error {
		called = true
		return nil
	}

	router.AddRule(matcher, handler)

	msg := Message{
		Topic:   "test.topic",
		Payload: "test",
	}

	err := router.Route(context.Background(), msg)
	if err != nil {
		t.Fatalf("Failed to route: %v", err)
	}

	if !called {
		t.Error("Expected handler to be called")
	}
}

func TestNewBroker(t *testing.T) {
	bus := NewEventBus(100)
	defer bus.Close()

	broker := NewBroker(bus)
	if broker == nil {
		t.Fatal("Expected non-nil broker")
	}
	if broker.bus != bus {
		t.Error("Expected broker to have the provided bus")
	}
}

func TestBrokerSetTransformer(t *testing.T) {
	bus := NewEventBus(100)
	defer bus.Close()

	broker := NewBroker(bus)

	transformer := func(msg Message) Message {
		msg.Metadata = map[string]string{"transformed": "true"}
		return msg
	}

	broker.SetTransformer(transformer)

	received := make(chan Message, 1)
	handler := func(ctx context.Context, msg Message) error {
		received <- msg
		return nil
	}

	broker.Subscribe("test.topic", handler)

	err := broker.Publish(context.Background(), "test.topic", "test payload", nil)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	select {
	case msg := <-received:
		if msg.Metadata["transformed"] != "true" {
			t.Error("Expected transformer to be applied")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for message")
	}
}

func TestBrokerAddRoutingRule(t *testing.T) {
	bus := NewEventBus(100)
	defer bus.Close()

	broker := NewBroker(bus)

	matcher := func(msg Message) bool {
		return msg.Topic == "test.topic"
	}

	handler := func(ctx context.Context, msg Message) error {
		return nil
	}

	initialRuleCount := len(broker.router.rules)
	broker.AddRoutingRule(matcher, handler)

	if len(broker.router.rules) != initialRuleCount+1 {
		t.Errorf("Expected rule count to increase by 1, got %d", len(broker.router.rules))
	}
}

func TestBrokerSubscribe(t *testing.T) {
	bus := NewEventBus(100)
	defer bus.Close()

	broker := NewBroker(bus)

	received := make(chan Message, 1)
	handler := func(ctx context.Context, msg Message) error {
		received <- msg
		return nil
	}

	subID, err := broker.Subscribe("test.topic", handler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	if subID == "" {
		t.Error("Expected non-empty subscription ID")
	}

	err = broker.Publish(context.Background(), "test.topic", "test payload", nil)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	select {
	case msg := <-received:
		if msg.Topic != "test.topic" {
			t.Errorf("Expected topic 'test.topic', got '%s'", msg.Topic)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for message")
	}
}

func TestBrokerUnsubscribe(t *testing.T) {
	bus := NewEventBus(100)
	defer bus.Close()

	broker := NewBroker(bus)

	handler := func(ctx context.Context, msg Message) error {
		return nil
	}

	subID, err := broker.Subscribe("test.topic", handler)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	err = broker.Unsubscribe(subID)
	if err != nil {
		t.Fatalf("Failed to unsubscribe: %v", err)
	}
}

func TestBrokerClose(t *testing.T) {
	bus := NewEventBus(100)
	broker := NewBroker(bus)

	err := broker.Close()
	if err != nil {
		t.Fatalf("Failed to close broker: %v", err)
	}
}
