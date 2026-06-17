package bus

import (
	"context"
	"strings"
	"sync"
	"time"
)

// Message represents a message sent through the bus
type Message struct {
	ID        string
	Topic     string
	Payload   interface{}
	Metadata  map[string]string
	Timestamp time.Time
}

// Handler is a function that handles a message
type Handler func(ctx context.Context, msg Message) error

// Subscription represents a subscription to a topic
type Subscription struct {
	ID      string
	Topic   string
	Handler Handler
	Filter  func(Message) bool
}

// Bus defines the interface for the message bus
type Bus interface {
	// Publish publishes a message to a topic
	Publish(ctx context.Context, topic string, payload interface{}, metadata map[string]string) error

	// Subscribe subscribes to a topic with a handler
	Subscribe(topic string, handler Handler) (string, error)

	// SubscribeWithFilter subscribes to a topic with a handler and filter
	SubscribeWithFilter(topic string, handler Handler, filter func(Message) bool) (string, error)

	// Unsubscribe unsubscribes from a topic
	Unsubscribe(subscriptionID string) error

	// Close closes the bus and cleans up resources
	Close() error
}

// EventBus is an in-memory implementation of the message bus
type EventBus struct {
	mu            sync.RWMutex
	subscriptions map[string][]Subscription
	messageChan   chan Message
	closed        bool
	wg            sync.WaitGroup
}

// NewEventBus creates a new event bus
func NewEventBus(bufferSize int) *EventBus {
	bus := &EventBus{
		subscriptions: make(map[string][]Subscription),
		messageChan:   make(chan Message, bufferSize),
	}

	bus.start()
	return bus
}

// start starts the event loop
func (b *EventBus) start() {
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		for msg := range b.messageChan {
			b.handleMessage(context.Background(), msg)
		}
	}()
}

// handleMessage handles a message by routing it to subscribers
func (b *EventBus) handleMessage(ctx context.Context, msg Message) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Get subscribers for the exact topic
	subs, ok := b.subscriptions[msg.Topic]
	if !ok {
		subs = []Subscription{}
	}

	// Also get subscribers for wildcard patterns
	if wildcardSubs, ok := b.subscriptions["*"]; ok {
		subs = append(subs, wildcardSubs...)
	}

	// Also check for pattern-based subscriptions
	for topic, topicSubs := range b.subscriptions {
		if topic != msg.Topic && topic != "*" {
			// Simple wildcard matching
			if strings.HasSuffix(topic, "*") {
				prefix := strings.TrimSuffix(topic, "*")
				if strings.HasPrefix(msg.Topic, prefix) {
					subs = append(subs, topicSubs...)
				}
			}
		}
	}

	for _, sub := range subs {
		if sub.Filter != nil && !sub.Filter(msg) {
			continue
		}

		b.wg.Add(1)
		go func(handler Handler) {
			defer b.wg.Done()
			_ = handler(ctx, msg)
		}(sub.Handler)
	}
}

// Publish publishes a message to a topic
func (b *EventBus) Publish(ctx context.Context, topic string, payload interface{}, metadata map[string]string) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return ErrBusClosed
	}

	msg := Message{
		ID:        generateID(),
		Topic:     topic,
		Payload:   payload,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	select {
	case b.messageChan <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrBufferFull
	}
}

// Subscribe subscribes to a topic with a handler
func (b *EventBus) Subscribe(topic string, handler Handler) (string, error) {
	return b.SubscribeWithFilter(topic, handler, nil)
}

// SubscribeWithFilter subscribes to a topic with a handler and filter
func (b *EventBus) SubscribeWithFilter(topic string, handler Handler, filter func(Message) bool) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return "", ErrBusClosed
	}

	subID := generateID()
	sub := Subscription{
		ID:      subID,
		Topic:   topic,
		Handler: handler,
		Filter:  filter,
	}

	b.subscriptions[topic] = append(b.subscriptions[topic], sub)

	return subID, nil
}

// Unsubscribe unsubscribes from a topic
func (b *EventBus) Unsubscribe(subscriptionID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return ErrBusClosed
	}

	for topic, subs := range b.subscriptions {
		for i, sub := range subs {
			if sub.ID == subscriptionID {
				b.subscriptions[topic] = append(subs[:i], subs[i+1:]...)
				return nil
			}
		}
	}

	return ErrSubscriptionNotFound
}

// Close closes the bus and cleans up resources
func (b *EventBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true
	close(b.messageChan)
	b.wg.Wait()

	return nil
}

// generateID generates a unique ID
func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// Errors
var (
	ErrBusClosed            = &BusError{Message: "bus is closed"}
	ErrBufferFull           = &BusError{Message: "message buffer is full"}
	ErrSubscriptionNotFound = &BusError{Message: "subscription not found"}
)

// BusError represents a bus error
type BusError struct {
	Message string
}

func (e *BusError) Error() string {
	return e.Message
}
