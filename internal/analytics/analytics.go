package analytics

import (
	"fmt"
	"sync"
	"time"
)

// AnalyticsEngine provides analytics and metrics
type AnalyticsEngine struct {
	mu       sync.RWMutex
	metrics  map[string]*Metric
	events   []Event
}

// Metric represents a performance metric
type Metric struct {
	Name      string
	Value     float64
	Timestamp time.Time
	Labels    map[string]string
}

// Event represents an event log
type Event struct {
	ID        string
	Type      string
	Timestamp time.Time
	Data      map[string]interface{}
}

// NewAnalyticsEngine creates a new analytics engine
func NewAnalyticsEngine() *AnalyticsEngine {
	return &AnalyticsEngine{
		metrics: make(map[string]*Metric),
		events:  make([]Event, 0),
	}
}

// RecordMetric records a metric
func (a *AnalyticsEngine) RecordMetric(name string, value float64, labels map[string]string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.metrics[name] = &Metric{
		Name:      name,
		Value:     value,
		Timestamp: time.Now(),
		Labels:    labels,
	}
}

// RecordEvent records an event
func (a *AnalyticsEngine) RecordEvent(eventType string, data map[string]interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()

	event := Event{
		ID:        generateID(),
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
	a.events = append(a.events, event)
}

// GetMetric retrieves a metric by name
func (a *AnalyticsEngine) GetMetric(name string) *Metric {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.metrics[name]
}

// ListMetrics lists all metrics
func (a *AnalyticsEngine) ListMetrics() []*Metric {
	a.mu.RLock()
	defer a.mu.RUnlock()

	metrics := make([]*Metric, 0, len(a.metrics))
	for _, m := range a.metrics {
		metrics = append(metrics, m)
	}
	return metrics
}

// ListEvents lists all events
func (a *AnalyticsEngine) ListEvents() []Event {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return append([]Event{}, a.events...)
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}