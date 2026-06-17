package grpc

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// ServerConfig holds configuration for the gRPC server
type ServerConfig struct {
	Address         string
	Port            int
	MaxConnections  int
	EnableTLS       bool
	CertFile        string
	KeyFile         string
	KeepaliveParams keepalive.ServerParameters
}

// Server wraps a gRPC server with additional functionality
type Server struct {
	config   *ServerConfig
	server   *grpc.Server
	listener net.Listener
	mu       sync.RWMutex
	started  bool
	services map[string]interface{}
}

// NewServer creates a new gRPC server
func NewServer(config *ServerConfig) *Server {
	if config == nil {
		config = &ServerConfig{
			Address: "0.0.0.0",
			Port:    50051,
		}
	}

	if config.KeepaliveParams.MaxConnectionIdle == 0 {
		config.KeepaliveParams = keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:              10 * time.Second,
			Timeout:           1 * time.Second,
		}
	}

	return &Server{
		config:   config,
		services: make(map[string]interface{}),
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("server already started")
	}

	addr := fmt.Sprintf("%s:%d", s.config.Address, s.config.Port)

	var opts []grpc.ServerOption

	// Keepalive
	opts = append(opts, grpc.KeepaliveParams(s.config.KeepaliveParams))

	// TLS
	if s.config.EnableTLS {
		if s.config.CertFile == "" || s.config.KeyFile == "" {
			return fmt.Errorf("TLS enabled but cert/key files not provided")
		}

		cert, err := tls.LoadX509KeyPair(s.config.CertFile, s.config.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificates: %w", err)
		}

		creds := credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.NoClientCert,
		})
		opts = append(opts, grpc.Creds(creds))
	}

	// Max connections
	if s.config.MaxConnections > 0 {
		opts = append(opts, grpc.MaxConcurrentStreams(uint32(s.config.MaxConnections)))
	}

	s.server = grpc.NewServer(opts...)

	// Register services
	for name, service := range s.services {
		switch svc := service.(type) {
		case grpcServiceRegistrar:
			svc.Register(s.server)
		default:
			return fmt.Errorf("service %s does not implement grpcServiceRegistrar", name)
		}
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener
	s.started = true

	go func() {
		if err := s.server.Serve(listener); err != nil {
			// Log error in production
		}
	}()

	return nil
}

// Stop stops the gRPC server gracefully
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return nil
	}

	s.server.GracefulStop()
	s.started = false

	return nil
}

// RegisterService registers a gRPC service
func (s *Server) RegisterService(name string, service interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("cannot register service after server has started")
	}

	if _, ok := service.(grpcServiceRegistrar); !ok {
		return fmt.Errorf("service does not implement grpcServiceRegistrar")
	}

	s.services[name] = service
	return nil
}

// GetServer returns the underlying gRPC server
func (s *Server) GetServer() *grpc.Server {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.server
}

// IsStarted returns whether the server is started
func (s *Server) IsStarted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.started
}

// grpcServiceRegistrar is an interface for services that can register themselves
type grpcServiceRegistrar interface {
	Register(server *grpc.Server)
}
