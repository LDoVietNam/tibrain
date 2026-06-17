// Package routeagent - persistent config key/value access on the Store.
package routeagent

import "sync"

// in-memory config table; mutex protects concurrent access.
var (
	cfgMu  sync.RWMutex
	cfgMap = make(map[string]string)
)

// GetConfig returns the value of a config key from the store, or
// the supplied default if the key is unset or the store is unavailable.
//
// This is a stub implementation backed by an in-memory map; the real
// implementation would persist to the SQLite database alongside route events.
func (s *Store) GetConfig(key, defaultValue string) string {
	if s == nil {
		return defaultValue
	}
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if v, ok := cfgMap[key]; ok && v != "" {
		return v
	}
	return defaultValue
}

// SetConfig stores a config key/value pair on the store (in-memory).
func (s *Store) SetConfig(key, value string) {
	if s == nil {
		return
	}
	cfgMu.Lock()
	defer cfgMu.Unlock()
	cfgMap[key] = value
}
