package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LifecycleState represents the lifecycle state of a plugin
type LifecycleState string

const (
	LifecycleStateStopped    LifecycleState = "stopped"
	LifecycleStateStarting   LifecycleState = "starting"
	LifecycleStateRunning    LifecycleState = "running"
	LifecycleStateStopping   LifecycleState = "stopping"
	LifecycleStateError      LifecycleState = "error"
	LifecycleStateRestarting LifecycleState = "restarting"
)

// LifecycleManager manages the lifecycle of plugins
type LifecycleManager struct {
	mu            sync.RWMutex
	plugins       map[string]*PluginLifecycle
	healthChecker *HealthChecker
	restartPolicy RestartPolicy
}

// PluginLifecycle contains lifecycle information for a plugin
type PluginLifecycle struct {
	Name         string
	State        LifecycleState
	StartTime    time.Time
	LastError    error
	RestartCount int
	MaxRestarts  int
	Config       map[string]string
	mu           sync.RWMutex
}

// RestartPolicy defines how plugins should be restarted
type RestartPolicy struct {
	MaxRestarts       int
	RestartDelay      time.Duration
	BackoffMultiplier float64
}

// HealthChecker performs health checks on plugins
type HealthChecker struct {
	interval time.Duration
	timeout  time.Duration
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(checkInterval, checkTimeout time.Duration, policy RestartPolicy) *LifecycleManager {
	return &LifecycleManager{
		plugins: make(map[string]*PluginLifecycle),
		healthChecker: &HealthChecker{
			interval: checkInterval,
			timeout:  checkTimeout,
		},
		restartPolicy: policy,
	}
}

// StartPlugin starts a plugin
func (lm *LifecycleManager) StartPlugin(ctx context.Context, name string, plugin Plugin, config map[string]string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Check if plugin is already running
	if lifecycle, exists := lm.plugins[name]; exists {
		if lifecycle.State == LifecycleStateRunning {
			return fmt.Errorf("plugin %s is already running", name)
		}
	}

	// Create lifecycle entry
	lifecycle := &PluginLifecycle{
		Name:        name,
		State:       LifecycleStateStarting,
		StartTime:   time.Now(),
		Config:      config,
		MaxRestarts: lm.restartPolicy.MaxRestarts,
	}
	lm.plugins[name] = lifecycle

	// Initialize plugin
	if err := plugin.Initialize(ctx, config); err != nil {
		lifecycle.State = LifecycleStateError
		lifecycle.LastError = err
		return fmt.Errorf("failed to initialize plugin %s: %w", name, err)
	}

	// Transition to running state
	lifecycle.State = LifecycleStateRunning

	return nil
}

// StopPlugin stops a plugin
func (lm *LifecycleManager) StopPlugin(ctx context.Context, name string, plugin Plugin) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lifecycle, exists := lm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Transition to stopping state
	lifecycle.State = LifecycleStateStopping

	// Shutdown plugin
	if err := plugin.Shutdown(ctx); err != nil {
		lifecycle.State = LifecycleStateError
		lifecycle.LastError = err
		return fmt.Errorf("failed to shutdown plugin %s: %w", name, err)
	}

	// Transition to stopped state
	lifecycle.State = LifecycleStateStopped

	return nil
}

// RestartPlugin restarts a plugin
func (lm *LifecycleManager) RestartPlugin(ctx context.Context, name string, plugin Plugin) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lifecycle, exists := lm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Check restart limit
	if lifecycle.RestartCount >= lifecycle.MaxRestarts {
		return fmt.Errorf("plugin %s has reached maximum restart count (%d)", name, lifecycle.MaxRestarts)
	}

	// Increment restart count
	lifecycle.RestartCount++

	// Transition to restarting state
	lifecycle.State = LifecycleStateRestarting

	// Stop plugin first
	if err := plugin.Shutdown(ctx); err != nil {
		lifecycle.State = LifecycleStateError
		lifecycle.LastError = err
		return fmt.Errorf("failed to stop plugin %s for restart: %w", name, err)
	}

	// Wait for restart delay
	delay := lm.calculateRestartDelay(lifecycle.RestartCount)
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return ctx.Err()
	}

	// Start plugin again
	if err := plugin.Initialize(ctx, lifecycle.Config); err != nil {
		lifecycle.State = LifecycleStateError
		lifecycle.LastError = err
		return fmt.Errorf("failed to restart plugin %s: %w", name, err)
	}

	// Transition to running state
	lifecycle.State = LifecycleStateRunning

	return nil
}

// GetLifecycle returns the lifecycle state of a plugin
func (lm *LifecycleManager) GetLifecycle(name string) (*PluginLifecycle, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	lifecycle, exists := lm.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return lifecycle, nil
}

// GetAllLifecycles returns all plugin lifecycles
func (lm *LifecycleManager) GetAllLifecycles() map[string]*PluginLifecycle {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]*PluginLifecycle, len(lm.plugins))
	for name, lifecycle := range lm.plugins {
		result[name] = lifecycle
	}

	return result
}

// StartHealthChecks starts the health checker for all plugins
func (lm *LifecycleManager) StartHealthChecks(ctx context.Context, platform Platform) {
	ticker := time.NewTicker(lm.healthChecker.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lm.checkAllPlugins(ctx, platform)
		}
	}
}

// checkAllPlugins performs health checks on all plugins
func (lm *LifecycleManager) checkAllPlugins(ctx context.Context, platform Platform) {
	plugins := platform.ListPlugins()

	for _, pluginInfo := range plugins {
		name := pluginInfo.Metadata.Name

		// Create context with timeout for health check
		healthCtx, cancel := context.WithTimeout(ctx, lm.healthChecker.timeout)

		// Get plugin
		plugin, err := platform.GetPlugin(name)
		if err != nil {
			cancel()
			continue
		}

		// Perform health check
		if err := plugin.HealthCheck(healthCtx); err != nil {
			cancel()

			// Log error and attempt restart
			lm.mu.Lock()
			if lifecycle, exists := lm.plugins[name]; exists {
				lifecycle.LastError = err
				lifecycle.State = LifecycleStateError
			}
			lm.mu.Unlock()

			// Attempt restart
			_ = lm.RestartPlugin(ctx, name, plugin)
		} else {
			cancel()

			// Update health status
			lm.mu.Lock()
			if lifecycle, exists := lm.plugins[name]; exists {
				lifecycle.State = LifecycleStateRunning
				lifecycle.LastError = nil
			}
			lm.mu.Unlock()
		}
	}
}

// calculateRestartDelay calculates the delay before restarting based on exponential backoff
func (lm *LifecycleManager) calculateRestartDelay(restartCount int) time.Duration {
	baseDelay := lm.restartPolicy.RestartDelay
	multiplier := lm.restartPolicy.BackoffMultiplier

	delay := baseDelay
	for i := 1; i < restartCount; i++ {
		delay = time.Duration(float64(delay) * multiplier)
	}

	// Cap at 5 minutes
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}

	return delay
}

// DefaultRestartPolicy returns the default restart policy
func DefaultRestartPolicy() RestartPolicy {
	return RestartPolicy{
		MaxRestarts:       3,
		RestartDelay:      5 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// DefaultHealthChecker returns a default health checker
func DefaultHealthChecker() *HealthChecker {
	return &HealthChecker{
		interval: 30 * time.Second,
		timeout:  10 * time.Second,
	}
}
