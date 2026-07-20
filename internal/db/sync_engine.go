package db

import (
	"fmt"
	"sync"
)

// InternalSyncEvent represents an event when a knowledge state changes
type InternalSyncEvent struct {
	Type       string // e.g., "ConfidenceDecayed", "KnowledgeGapIdentified"
	EntityID   string
	Properties map[string]interface{}
}

// SyncEngine coordinates state changes across SQLite and Qdrant
type SyncEngine struct {
	listeners []chan InternalSyncEvent
	mu        sync.RWMutex
}

// GlobalSyncEngine is the singleton instance
var GlobalSyncEngine = &SyncEngine{}

// Subscribe adds a listener to the sync engine
func (s *SyncEngine) Subscribe(listener chan InternalSyncEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listeners = append(s.listeners, listener)
}

// Publish broadcasts an event to all listeners
func (s *SyncEngine) Publish(event InternalSyncEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, listener := range s.listeners {
		select {
		case listener <- event:
		default:
			// If listener is blocked, we drop the event to avoid blocking ALM
			fmt.Println("⚠️ SyncEngine: Dropped event due to blocked listener")
		}
	}
}
