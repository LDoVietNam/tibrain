package devin

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// LifecycleState represents the lifecycle state
type LifecycleState string

const (
	StateStopped  LifecycleState = "stopped"
	StateStarting LifecycleState = "starting"
	StateRunning  LifecycleState = "running"
	StateStopping LifecycleState = "stopping"
	StateError    LifecycleState = "error"
)

// LifecycleManager manages the lifecycle of the Devin agent process
type LifecycleManager struct {
	mu                  sync.RWMutex
	state               LifecycleState
	process             *exec.Cmd
	client              *Client
	config              *Config
	startTime           time.Time
	healthCheckInterval time.Duration
	healthCheckTicker   *time.Ticker
	shutdownHooks       []func()
	cancelHealthCheck   context.CancelFunc
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(config *Config) *LifecycleManager {
	return &LifecycleManager{
		state:               StateStopped,
		config:              config,
		healthCheckInterval: config.HealthCheckInterval,
		shutdownHooks:       make([]func(), 0),
	}
}

// Start starts the Devin agent process
func (lm *LifecycleManager) Start(ctx context.Context) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.state != StateStopped {
		return fmt.Errorf("cannot start, current state: %s", lm.state)
	}

	lm.state = StateStarting

	// Start Python gRPC server
	if err := lm.startPythonServer(ctx); err != nil {
		lm.state = StateError
		return fmt.Errorf("failed to start Python server: %w", err)
	}

	// Connect gRPC client
	clientConfig := &ClientConfig{
		Address:        lm.config.Address,
		Port:           lm.config.Port,
		Timeout:        lm.config.Timeout,
		ConnectTimeout: lm.config.ConnectTimeout,
		MaxRetries:     lm.config.MaxRetries,
		EnableTLS:      lm.config.EnableTLS,
	}
	client, err := NewClient(clientConfig)
	if err != nil {
		lm.state = StateError
		return fmt.Errorf("failed to create gRPC client: %w", err)
	}

	lm.client = client
	lm.startTime = time.Now()
	lm.state = StateRunning

	// Start health check
	lm.startHealthCheck(ctx)

	return nil
}

// startPythonServer starts the Python gRPC server
func (lm *LifecycleManager) startPythonServer(ctx context.Context) error {
	// Path to Python gRPC server
	serverPath := "internal/bridge/grpc_server.py"

	// Build command
	cmd := exec.CommandContext(ctx, "python", serverPath, "--port", fmt.Sprintf("%d", lm.config.Port))

	// Set output pipes
	cmd.Stdout = nil // In production, redirect to log
	cmd.Stderr = nil // In production, redirect to log

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Python server: %w", err)
	}

	lm.process = cmd

	// Wait a bit for server to start
	time.Sleep(1 * time.Second)

	return nil
}

// Stop stops the Devin agent process gracefully
func (lm *LifecycleManager) Stop(ctx context.Context) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.state == StateStopped {
		return nil
	}

	lm.state = StateStopping

	// Stop health check
	if lm.cancelHealthCheck != nil {
		lm.cancelHealthCheck()
	}

	if lm.healthCheckTicker != nil {
		lm.healthCheckTicker.Stop()
	}

	// Call shutdown hooks
	for _, hook := range lm.shutdownHooks {
		hook()
	}

	// Close gRPC client
	if lm.client != nil {
		if err := lm.client.Close(); err != nil {
			return fmt.Errorf("failed to close client: %w", err)
		}
	}

	// Stop Python process
	if lm.process != nil && lm.process.Process != nil {
		if err := lm.process.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}

	lm.state = StateStopped
	lm.client = nil
	lm.process = nil

	return nil
}

// Restart restarts the Devin agent process
func (lm *LifecycleManager) Restart(ctx context.Context) error {
	if err := lm.Stop(ctx); err != nil {
		return err
	}

	// Wait a bit before starting
	time.Sleep(1 * time.Second)

	return lm.Start(ctx)
}

// GetState returns the current state
func (lm *LifecycleManager) GetState() LifecycleState {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.state
}

// GetClient returns the gRPC client
func (lm *LifecycleManager) GetClient() *Client {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.client
}

// GetUptime returns the uptime of the agent
func (lm *LifecycleManager) GetUptime() time.Duration {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if lm.startTime.IsZero() {
		return 0
	}

	return time.Since(lm.startTime)
}

// startHealthCheck starts the health check loop
func (lm *LifecycleManager) startHealthCheck(ctx context.Context) {
	healthCtx, cancel := context.WithCancel(ctx)
	lm.cancelHealthCheck = cancel

	lm.healthCheckTicker = time.NewTicker(lm.healthCheckInterval)

	go func() {
		defer lm.healthCheckTicker.Stop()

		for {
			select {
			case <-healthCtx.Done():
				return
			case <-lm.healthCheckTicker.C:
				lm.performHealthCheck(healthCtx)
			}
		}
	}()
}

// performHealthCheck performs a health check
func (lm *LifecycleManager) performHealthCheck(ctx context.Context) {
	if lm.client == nil {
		return
	}

	response, err := lm.client.HealthCheck(ctx)
	if err != nil {
		// Log error, could trigger auto-restart
		return
	}

	if !response.Healthy {
		// Log unhealthy status, could trigger auto-restart
	}
}

// AddShutdownHook adds a function to be called on shutdown
func (lm *LifecycleManager) AddShutdownHook(hook func()) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.shutdownHooks = append(lm.shutdownHooks, hook)
}

// HealthStatus represents the health status
type HealthStatus struct {
	Healthy   bool
	State     LifecycleState
	Message   string
	Details   map[string]string
	Timestamp time.Time
}

// GetHealthStatus returns the current health status
func (lm *LifecycleManager) GetHealthStatus(ctx context.Context) *HealthStatus {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	status := &HealthStatus{
		State:     lm.state,
		Message:   "",
		Details:   make(map[string]string),
		Timestamp: time.Now(),
	}

	if lm.state != StateRunning {
		status.Healthy = false
		status.Message = fmt.Sprintf("Agent not running, state: %s", lm.state)
		return status
	}

	if lm.client == nil {
		status.Healthy = false
		status.Message = "Client not initialized"
		return status
	}

	response, err := lm.client.HealthCheck(ctx)
	if err != nil {
		status.Healthy = false
		status.Message = fmt.Sprintf("Health check failed: %v", err)
		return status
	}

	status.Healthy = response.Healthy
	status.Message = response.Message
	status.Details = response.Details

	return status
}
