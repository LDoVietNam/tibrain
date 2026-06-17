package events

import (
	"context"
	"sync"
	"time"
)

// Event represents a system event
type Event struct {
	ID        string
	Type      string
	Source    string
	Payload   interface{}
	Metadata  map[string]string
	Timestamp time.Time
}

// Handler is a function that handles an event
type Handler func(ctx context.Context, event Event) error

// EventHandler represents a registered event handler
type EventHandler struct {
	ID       string
	Type     string
	Handler  Handler
	Filter   func(Event) bool
	Priority int
}

// Dispatcher manages event dispatching
type Dispatcher struct {
	mu        sync.RWMutex
	handlers  map[string][]EventHandler
	eventChan chan Event
	workers   int
	wg        sync.WaitGroup
	closed    bool
}

// NewDispatcher creates a new event dispatcher
func NewDispatcher(bufferSize int, workers int) *Dispatcher {
	if bufferSize <= 0 {
		bufferSize = 1000
	}
	if workers <= 0 {
		workers = 5
	}

	d := &Dispatcher{
		handlers:  make(map[string][]EventHandler),
		eventChan: make(chan Event, bufferSize),
		workers:   workers,
	}

	d.start()
	return d
}

// start starts the event dispatcher workers
func (d *Dispatcher) start() {
	for i := 0; i < d.workers; i++ {
		d.wg.Add(1)
		go d.worker()
	}
}

// worker processes events from the channel
func (d *Dispatcher) worker() {
	defer d.wg.Done()

	for event := range d.eventChan {
		d.dispatch(context.Background(), event)
	}
}

// dispatch dispatches an event to all matching handlers
func (d *Dispatcher) dispatch(ctx context.Context, event Event) {
	d.mu.RLock()
	handlers, ok := d.handlers[event.Type]
	d.mu.RUnlock()

	if !ok {
		return
	}

	// Sort handlers by priority (higher priority first)
	// For now, just process in order

	for _, handler := range handlers {
		if handler.Filter != nil && !handler.Filter(event) {
			continue
		}

		// Run handler in goroutine to avoid blocking
		go func(h Handler) {
			_ = h(ctx, event)
		}(handler.Handler)
	}
}

// Publish publishes an event
func (d *Dispatcher) Publish(ctx context.Context, event Event) error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed {
		return ErrDispatcherClosed
	}

	// Set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Set ID if not set
	if event.ID == "" {
		event.ID = generateEventID()
	}

	select {
	case d.eventChan <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrBufferFull
	}
}

// Subscribe subscribes to an event type
func (d *Dispatcher) Subscribe(eventType string, handler Handler) (string, error) {
	return d.SubscribeWithFilter(eventType, handler, nil, 0)
}

// SubscribeWithFilter subscribes with a filter and priority
func (d *Dispatcher) SubscribeWithFilter(eventType string, handler Handler, filter func(Event) bool, priority int) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return "", ErrDispatcherClosed
	}

	handlerID := generateEventID()
	eventHandler := EventHandler{
		ID:       handlerID,
		Type:     eventType,
		Handler:  handler,
		Filter:   filter,
		Priority: priority,
	}

	d.handlers[eventType] = append(d.handlers[eventType], eventHandler)

	return handlerID, nil
}

// Unsubscribe unsubscribes from an event type
func (d *Dispatcher) Unsubscribe(handlerID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return ErrDispatcherClosed
	}

	for eventType, handlers := range d.handlers {
		for i, handler := range handlers {
			if handler.ID == handlerID {
				d.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				return nil
			}
		}
	}

	return ErrHandlerNotFound
}

// Close closes the dispatcher
func (d *Dispatcher) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	d.closed = true
	close(d.eventChan)
	d.wg.Wait()

	return nil
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string
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
	ErrDispatcherClosed = &EventError{Message: "dispatcher is closed"}
	ErrBufferFull       = &EventError{Message: "event buffer is full"}
	ErrHandlerNotFound  = &EventError{Message: "handler not found"}
)

// EventError represents an event error
type EventError struct {
	Message string
}

func (e *EventError) Error() string {
	return e.Message
}
