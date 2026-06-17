// Package learn provides the BEADS v2 learning system entry point.
// This file initializes the global engine instance.

package learn

import (
	"sync"
)

var (
	globalEngine *Engine
	engineOnce   sync.Once
)

// GetEngine returns the global learning engine singleton.
func GetEngine() *Engine {
	engineOnce.Do(func() {
		// Load config from ti.json or environment
		cfg := DefaultConfig()

		// Override from config file if exists
		// TODO: load from config package

		storage, err := NewStorage("")
		if err != nil {
			panic("failed to create learn storage: " + err.Error())
		}

		engine, err := NewEngine(cfg, storage)
		if err != nil {
			panic("failed to create learn engine: " + err.Error())
		}

		globalEngine = engine
	})

	return globalEngine
}

// Init initializes the learning system (call from main).
func Init() {
	_ = GetEngine() // force init
}
