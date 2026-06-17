package grpc

import (
	"testing"

	"google.golang.org/grpc"
)

func TestNewServer(t *testing.T) {
	config := &ServerConfig{
		Address: "localhost",
		Port:    50051,
	}

	server := NewServer(config)
	if server == nil {
		t.Fatal("Expected non-nil server")
	}
	if server.config.Address != "localhost" {
		t.Errorf("Expected address 'localhost', got '%s'", server.config.Address)
	}
	if server.config.Port != 50051 {
		t.Errorf("Expected port 50051, got %d", server.config.Port)
	}
}

func TestNewServerDefaultConfig(t *testing.T) {
	server := NewServer(nil)
	if server == nil {
		t.Fatal("Expected non-nil server")
	}
	if server.config.Address != "0.0.0.0" {
		t.Errorf("Expected default address '0.0.0.0', got '%s'", server.config.Address)
	}
	if server.config.Port != 50051 {
		t.Errorf("Expected default port 50051, got %d", server.config.Port)
	}
}

func TestServerStartStop(t *testing.T) {
	config := &ServerConfig{
		Address: "localhost",
		Port:    0, // Use random port
	}

	server := NewServer(config)

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	if !server.IsStarted() {
		t.Error("Expected server to be started")
	}

	// Starting again should fail
	err = server.Start()
	if err == nil {
		t.Error("Expected error when starting already started server")
	}

	err = server.Stop()
	if err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}

	if server.IsStarted() {
		t.Error("Expected server to be stopped")
	}
}

func TestServerRegisterService(t *testing.T) {
	config := &ServerConfig{
		Address: "localhost",
		Port:    0,
	}

	server := NewServer(config)

	// Register a mock service
	mockService := &mockGrpcService{}
	err := server.RegisterService("mock", mockService)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Register non-registrar service should fail
	err = server.RegisterService("invalid", "not a service")
	if err == nil {
		t.Error("Expected error when registering non-registrar service")
	}
}

func TestServerRegisterAfterStart(t *testing.T) {
	config := &ServerConfig{
		Address: "localhost",
		Port:    0,
	}

	server := NewServer(config)

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Registering after start should fail
	mockService := &mockGrpcService{}
	err = server.RegisterService("mock", mockService)
	if err == nil {
		t.Error("Expected error when registering service after start")
	}
}

// mockGrpcService implements grpcServiceRegistrar for testing
type mockGrpcService struct{}

func (m *mockGrpcService) Register(server *grpc.Server) {
	// Mock implementation - do nothing
}
