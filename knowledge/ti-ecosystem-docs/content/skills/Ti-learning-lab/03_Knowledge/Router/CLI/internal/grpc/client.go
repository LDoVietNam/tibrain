package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ClientConfig holds configuration for the gRPC client
type ClientConfig struct {
	Address          string
	EnableTLS        bool
	ServerName       string
	DisableTLSVerify bool
	MaxRetries       int
	RetryDelay       time.Duration
	KeepaliveParams  keepalive.ClientParameters
	DialOptions      []grpc.DialOption
}

// Client wraps a gRPC client connection with additional functionality
type Client struct {
	config *ClientConfig
	conn   *grpc.ClientConn
	mu     sync.RWMutex
	closed bool
}

// NewClient creates a new gRPC client
func NewClient(config *ClientConfig) *Client {
	if config == nil {
		config = &ClientConfig{
			Address:    "localhost:50051",
			MaxRetries: 3,
			RetryDelay: 100 * time.Millisecond,
		}
	}

	if config.KeepaliveParams.Time == 0 {
		config.KeepaliveParams = keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             time.Second,
			PermitWithoutStream: true,
		}
	}

	return &Client{
		config: config,
	}
}

// Connect establishes a connection to the gRPC server
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return nil
	}

	var opts []grpc.DialOption

	// Keepalive
	opts = append(opts, grpc.WithKeepaliveParams(c.config.KeepaliveParams))

	// TLS
	if c.config.EnableTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: c.config.DisableTLSVerify,
		}

		if c.config.ServerName != "" {
			tlsConfig.ServerName = c.config.ServerName
		}

		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Custom dial options
	opts = append(opts, c.config.DialOptions...)

	// Block until connection is established
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.DialContext(ctx, c.config.Address, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", c.config.Address, err)
	}

	c.conn = conn
	return nil
}

// Disconnect closes the connection to the gRPC server
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	c.closed = true

	return err
}

// GetConn returns the underlying gRPC client connection
func (c *Client) GetConn() *grpc.ClientConn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil
}

// Invoke invokes a gRPC method with retry logic
func (c *Client) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(c.config.RetryDelay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := c.invoke(ctx, method, args, reply, opts...)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on context cancellation or certain errors
		if ctx.Err() != nil {
			return err
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", c.config.MaxRetries+1, lastErr)
}

// invoke performs a single invocation without retry
func (c *Client) invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("client not connected")
	}

	return conn.Invoke(ctx, method, args, reply, opts...)
}

// NewStream creates a new gRPC stream
func (c *Client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (interface{}, error) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return nil, fmt.Errorf("client not connected")
	}

	return conn.NewStream(ctx, desc, method, opts...)
}

// ConnectionPool manages a pool of gRPC client connections
type ConnectionPool struct {
	config      *ClientConfig
	connections []*Client
	mu          sync.RWMutex
	maxSize     int
	current     int
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config *ClientConfig, maxSize int) *ConnectionPool {
	if maxSize <= 0 {
		maxSize = 10
	}

	return &ConnectionPool{
		config:      config,
		connections: make([]*Client, 0, maxSize),
		maxSize:     maxSize,
	}
}

// Get gets a connection from the pool
func (p *ConnectionPool) Get(ctx context.Context) (*Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Return existing connection if available
	if len(p.connections) > 0 {
		client := p.connections[0]
		p.connections = p.connections[1:]
		return client, nil
	}

	// Create new connection if under max size
	if p.current < p.maxSize {
		client := NewClient(p.config)
		if err := client.Connect(ctx); err != nil {
			return nil, err
		}
		p.current++
		return client, nil
	}

	// Pool is full, create temporary connection
	client := NewClient(p.config)
	if err := client.Connect(ctx); err != nil {
		return nil, err
	}
	return client, nil
}

// Put returns a connection to the pool
func (p *ConnectionPool) Put(client *Client) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.connections) >= p.maxSize {
		// Pool is full, close the connection
		return client.Disconnect()
	}

	p.connections = append(p.connections, client)
	return nil
}

// Close closes all connections in the pool
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var lastErr error
	for _, client := range p.connections {
		if err := client.Disconnect(); err != nil {
			lastErr = err
		}
	}

	p.connections = make([]*Client, 0)
	p.current = 0

	return lastErr
}
