package bus

import (
	"context"
	"sync"
	"time"
)

// MessageTransformer transforms messages before they are delivered
type MessageTransformer func(Message) Message

// Router routes messages to different handlers based on rules
type Router struct {
	rules []RoutingRule
	mu    sync.RWMutex
}

// RoutingRule defines a routing rule
type RoutingRule struct {
	Matcher func(Message) bool
	Handler Handler
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		rules: make([]RoutingRule, 0),
	}
}

// AddRule adds a routing rule
func (r *Router) AddRule(matcher func(Message) bool, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rules = append(r.rules, RoutingRule{
		Matcher: matcher,
		Handler: handler,
	})
}

// Route routes a message through the rules
func (r *Router) Route(ctx context.Context, msg Message) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, rule := range r.rules {
		if rule.Matcher(msg) {
			if err := rule.Handler(ctx, msg); err != nil {
				return err
			}
		}
	}

	return nil
}

// Broker provides advanced message routing and transformation
type Broker struct {
	bus         Bus
	router      *Router
	transformer MessageTransformer
	mu          sync.RWMutex
}

// NewBroker creates a new broker
func NewBroker(bus Bus) *Broker {
	return &Broker{
		bus:    bus,
		router: NewRouter(),
	}
}

// SetTransformer sets the message transformer
func (b *Broker) SetTransformer(transformer MessageTransformer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.transformer = transformer
}

// AddRoutingRule adds a routing rule
func (b *Broker) AddRoutingRule(matcher func(Message) bool, handler Handler) {
	b.router.AddRule(matcher, handler)
}

// Publish publishes a message with optional transformation
func (b *Broker) Publish(ctx context.Context, topic string, payload interface{}, metadata map[string]string) error {
	msg := Message{
		ID:        generateID(),
		Topic:     topic,
		Payload:   payload,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	// Apply transformation if set
	b.mu.RLock()
	transformer := b.transformer
	b.mu.RUnlock()

	if transformer != nil {
		msg = transformer(msg)
	}

	// Route through rules
	if err := b.router.Route(ctx, msg); err != nil {
		return err
	}

	// Publish to bus with transformed data
	return b.bus.Publish(ctx, msg.Topic, msg.Payload, msg.Metadata)
}

// Subscribe subscribes to a topic
func (b *Broker) Subscribe(topic string, handler Handler) (string, error) {
	return b.bus.Subscribe(topic, handler)
}

// SubscribeWithFilter subscribes with a filter
func (b *Broker) SubscribeWithFilter(topic string, handler Handler, filter func(Message) bool) (string, error) {
	return b.bus.SubscribeWithFilter(topic, handler, filter)
}

// Unsubscribe unsubscribes from a topic
func (b *Broker) Unsubscribe(subscriptionID string) error {
	return b.bus.Unsubscribe(subscriptionID)
}

// Close closes the broker
func (b *Broker) Close() error {
	return b.bus.Close()
}
