package bus

import (
	"context"
	"strings"
	"sync"
)

// PubSub defines the interface for publish-subscribe operations
type PubSub interface {
	// Publish publishes a message to a topic
	Publish(ctx context.Context, topic string, payload interface{}, metadata map[string]string) error

	// Subscribe subscribes to a topic pattern
	Subscribe(pattern string, handler Handler) (string, error)

	// SubscribeWithFilter subscribes with a custom filter
	SubscribeWithFilter(pattern string, handler Handler, filter func(Message) bool) (string, error)

	// Unsubscribe unsubscribes from a topic
	Unsubscribe(subscriptionID string) error

	// GetSubscribers returns the number of subscribers for a topic
	GetSubscribers(topic string) int
}

// TopicMatcher defines the interface for topic pattern matching
type TopicMatcher interface {
	Match(topic, pattern string) bool
}

// SimpleMatcher implements simple wildcard matching
type SimpleMatcher struct{}

// Match checks if a topic matches a pattern
// Pattern can include * as a wildcard
func (m *SimpleMatcher) Match(topic, pattern string) bool {
	if pattern == "*" {
		return true
	}

	if pattern == topic {
		return true
	}

	// Handle wildcard at the end
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(topic, prefix)
	}

	// Handle wildcard at the start
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(topic, suffix)
	}

	return false
}

// PubSubBus implements the PubSub interface
type PubSubBus struct {
	bus        Bus
	matcher    TopicMatcher
	mu         sync.RWMutex
	patternMap map[string]string // pattern -> subscription ID
}

// NewPubSubBus creates a new pub/sub bus
func NewPubSubBus(bus Bus) *PubSubBus {
	return &PubSubBus{
		bus:        bus,
		matcher:    &SimpleMatcher{},
		patternMap: make(map[string]string),
	}
}

// Publish publishes a message to a topic
func (p *PubSubBus) Publish(ctx context.Context, topic string, payload interface{}, metadata map[string]string) error {
	return p.bus.Publish(ctx, topic, payload, metadata)
}

// Subscribe subscribes to a topic pattern
func (p *PubSubBus) Subscribe(pattern string, handler Handler) (string, error) {
	return p.SubscribeWithFilter(pattern, handler, nil)
}

// SubscribeWithFilter subscribes with a custom filter
func (p *PubSubBus) SubscribeWithFilter(pattern string, handler Handler, filter func(Message) bool) (string, error) {
	subID, err := p.bus.Subscribe("*", func(ctx context.Context, msg Message) error {
		if !p.matcher.Match(msg.Topic, pattern) {
			return nil
		}
		if filter != nil && !filter(msg) {
			return nil
		}
		return handler(ctx, msg)
	})

	if err != nil {
		return "", err
	}

	p.mu.Lock()
	p.patternMap[subID] = pattern
	p.mu.Unlock()

	return subID, nil
}

// Unsubscribe unsubscribes from a topic
func (p *PubSubBus) Unsubscribe(subscriptionID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.patternMap[subscriptionID]; !ok {
		return ErrSubscriptionNotFound
	}

	delete(p.patternMap, subscriptionID)
	return p.bus.Unsubscribe(subscriptionID)
}

// GetSubscribers returns the number of subscribers for a topic
func (p *PubSubBus) GetSubscribers(topic string) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	count := 0
	for _, pattern := range p.patternMap {
		if p.matcher.Match(topic, pattern) {
			count++
		}
	}

	return count
}
