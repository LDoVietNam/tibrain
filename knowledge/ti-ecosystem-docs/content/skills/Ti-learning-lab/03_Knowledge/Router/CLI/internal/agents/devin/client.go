package devin

import (
	"context"
	"fmt"
	"time"

	"github.com/ti/cli/internal/grpc"
)

// Client wraps gRPC client for Devin agent communication
type Client struct {
	grpcClient *grpc.Client
	config     *Config
}

// ClientConfig holds Devin client configuration
type ClientConfig struct {
	Address        string
	Port           int
	Timeout        time.Duration
	MaxRetries     int
	EnableTLS      bool
	ConnectTimeout time.Duration
}

// NewClient creates a new Devin gRPC client
func NewClient(config *ClientConfig) (*Client, error) {
	if config == nil {
		config = &ClientConfig{
			Address:        "localhost",
			Port:           50052,
			Timeout:        30 * time.Second,
			MaxRetries:     3,
			EnableTLS:      false,
			ConnectTimeout: 10 * time.Second,
		}
	}

	grpcConfig := &grpc.ClientConfig{
		Address:    fmt.Sprintf("%s:%d", config.Address, config.Port),
		EnableTLS:  config.EnableTLS,
		MaxRetries: config.MaxRetries,
		RetryDelay: 100 * time.Millisecond,
	}

	grpcClient := grpc.NewClient(grpcConfig)

	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	if err := grpcClient.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Devin gRPC server: %w", err)
	}

	return &Client{
		grpcClient: grpcClient,
		config:     &Config{Port: config.Port, Address: config.Address},
	}, nil
}

// Execute executes a task with streaming response
func (c *Client) Execute(ctx context.Context, taskID, prompt string, options map[string]interface{}) (<-chan ExecuteResponse, error) {
	// This would use the actual gRPC generated code
	// For now, return a mock channel

	responseChan := make(chan ExecuteResponse, 10)

	go func() {
		defer close(responseChan)

		// Send status update
		responseChan <- ExecuteResponse{
			Type:   ResponseTypeStatus,
			Status: "starting",
		}

		// Simulate processing
		time.Sleep(100 * time.Millisecond)

		// Send output
		responseChan <- ExecuteResponse{
			Type:   ResponseTypeOutput,
			Output: fmt.Sprintf("Processing: %s", prompt[:min(50, len(prompt))]),
			Format: "text",
		}

		// Send complete
		responseChan <- ExecuteResponse{
			Type:    ResponseTypeComplete,
			Success: true,
			Result:  fmt.Sprintf("Task %s completed", taskID),
		}
	}()

	return responseChan, nil
}

// GetStatus gets the current status of the agent
func (c *Client) GetStatus(ctx context.Context) (*StatusResponse, error) {
	// Mock implementation
	return &StatusResponse{
		Status:   "idle",
		Metadata: map[string]string{},
	}, nil
}

// HealthCheck performs a health check on the agent
func (c *Client) HealthCheck(ctx context.Context) (*HealthCheckResponse, error) {
	if !c.grpcClient.IsConnected() {
		return &HealthCheckResponse{
			Healthy: false,
			Message: "Not connected to gRPC server",
			Details: map[string]string{},
		}, nil
	}

	return &HealthCheckResponse{
		Healthy: true,
		Message: "OK",
		Details: map[string]string{
			"address": fmt.Sprintf("%s:%d", c.config.Address, c.config.Port),
		},
	}, nil
}

// Shutdown shuts down the agent
func (c *Client) Shutdown(ctx context.Context) error {
	return c.grpcClient.Disconnect()
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.grpcClient.Disconnect()
}

// ExecuteResponse represents a response from Execute
type ExecuteResponse struct {
	Type    ResponseType
	Status  string
	Output  string
	Format  string
	Success bool
	Result  string
	Error   *ExecuteError
}

// ResponseType represents the type of response
type ResponseType int

const (
	ResponseTypeStatus ResponseType = iota
	ResponseTypeOutput
	ResponseTypeError
	ResponseTypeComplete
)

// ExecuteError represents an execution error
type ExecuteError struct {
	Code    string
	Message string
	Details string
}

// StatusResponse represents the agent status
type StatusResponse struct {
	Status   string
	Metadata map[string]string
}

// HealthCheckResponse represents a health check result
type HealthCheckResponse struct {
	Healthy bool
	Message string
	Details map[string]string
}

// ShutdownResponse represents a shutdown result
type ShutdownResponse struct {
	Success bool
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
