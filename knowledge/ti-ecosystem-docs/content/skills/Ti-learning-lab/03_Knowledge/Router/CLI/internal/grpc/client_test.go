package grpc

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

func TestNewClient(t *testing.T) {
	config := &ClientConfig{
		Address: "localhost:50051",
	}

	client := NewClient(config)
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	if client.config.Address != "localhost:50051" {
		t.Errorf("Expected address 'localhost:50051', got '%s'", client.config.Address)
	}
}

func TestNewClientDefaultConfig(t *testing.T) {
	client := NewClient(nil)
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	if client.config.Address != "localhost:50051" {
		t.Errorf("Expected default address 'localhost:50051', got '%s'", client.config.Address)
	}
	if client.config.MaxRetries != 3 {
		t.Errorf("Expected default max retries 3, got %d", client.config.MaxRetries)
	}
}

func TestClientIsConnected(t *testing.T) {
	client := NewClient(nil)

	if client.IsConnected() {
		t.Error("Expected client to not be connected initially")
	}
}

func TestClientDisconnectWithoutConnect(t *testing.T) {
	client := NewClient(nil)

	err := client.Disconnect()
	if err != nil {
		t.Fatalf("Failed to disconnect: %v", err)
	}
}

func TestConnectionPool(t *testing.T) {
	config := &ClientConfig{
		Address: "localhost:50051",
	}

	pool := NewConnectionPool(config, 2)
	if pool == nil {
		t.Fatal("Expected non-nil pool")
	}
	if pool.maxSize != 2 {
		t.Errorf("Expected max size 2, got %d", pool.maxSize)
	}
}

func TestConnectionPoolDefaultSize(t *testing.T) {
	config := &ClientConfig{
		Address: "localhost:50051",
	}

	pool := NewConnectionPool(config, 0)
	if pool.maxSize != 10 {
		t.Errorf("Expected default max size 10, got %d", pool.maxSize)
	}
}

func TestConnectionPoolClose(t *testing.T) {
	config := &ClientConfig{
		Address: "localhost:50051",
	}

	pool := NewConnectionPool(config, 2)

	err := pool.Close()
	if err != nil {
		t.Fatalf("Failed to close pool: %v", err)
	}

	if pool.current != 0 {
		t.Errorf("Expected current connections to be 0, got %d", pool.current)
	}
}

func TestUnaryServerChain(t *testing.T) {
	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return nil, nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test/Method",
	}

	// Test with no interceptors
	chain := UnaryServerChain()
	resp, err := chain(context.Background(), nil, info, handler)
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}
	if resp != nil {
		t.Error("Expected nil response")
	}
	if !called {
		t.Error("Expected handler to be called")
	}
}

func TestUnaryClientChain(t *testing.T) {
	called := false
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		called = true
		return nil
	}

	// Test with no interceptors
	chain := UnaryClientChain()
	err := chain(context.Background(), "/test/Method", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("Invoker failed: %v", err)
	}
	if !called {
		t.Error("Expected invoker to be called")
	}
}
