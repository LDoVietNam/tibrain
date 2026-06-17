package grpc

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLoggingInterceptor(t *testing.T) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test/Method",
	}

	interceptor := LoggingInterceptor
	resp, err := interceptor(context.Background(), "request", info, handler)
	if err != nil {
		t.Fatalf("Handler failed: %v", err)
	}
	if resp != "response" {
		t.Errorf("Expected 'response', got '%v'", resp)
	}
}

func TestRecoveryInterceptor(t *testing.T) {
	panicHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("test panic")
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test/Method",
	}

	interceptor := RecoveryInterceptor
	_, err := interceptor(context.Background(), "request", info, panicHandler)
	if err == nil {
		t.Error("Expected error from panic")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Error("Expected gRPC status error")
	}
	if st.Code() != codes.Internal {
		t.Errorf("Expected Internal code, got %v", st.Code())
	}
}

func TestClientRetryInterceptor(t *testing.T) {
	attempts := 0
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		attempts++
		if attempts < 3 {
			return status.Error(codes.Unavailable, "unavailable")
		}
		return nil
	}

	interceptor := ClientRetryInterceptor(5, 10*time.Millisecond)
	err := interceptor(context.Background(), "/test/Method", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("Invoker failed: %v", err)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestClientRetryInterceptorNonRetryable(t *testing.T) {
	attempts := 0
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		attempts++
		return status.Error(codes.InvalidArgument, "invalid")
	}

	interceptor := ClientRetryInterceptor(5, 10*time.Millisecond)
	err := interceptor(context.Background(), "/test/Method", nil, nil, nil, invoker)
	if err == nil {
		t.Error("Expected error")
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt for non-retryable error, got %d", attempts)
	}
}

func TestTimeoutInterceptor(t *testing.T) {
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	}

	interceptor := TimeoutInterceptor(10 * time.Millisecond)
	err := interceptor(context.Background(), "/test/Method", nil, nil, nil, invoker)
	if err == nil {
		t.Error("Expected timeout error")
	}
}
