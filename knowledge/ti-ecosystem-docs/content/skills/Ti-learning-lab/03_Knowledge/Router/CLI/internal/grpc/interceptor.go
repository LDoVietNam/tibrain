package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InterceptorConfig holds configuration for interceptors
type InterceptorConfig struct {
	EnableLogging   bool
	EnableRecovery  bool
	EnableMetrics   bool
	EnableRateLimit bool
}

// LoggingInterceptor logs gRPC requests and responses
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	duration := time.Since(start)

	// Log in production
	_ = duration
	_ = info.FullMethod

	return resp, err
}

// RecoveryInterceptor recovers from panics in gRPC handlers
func RecoveryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Log panic in production
			err = status.Errorf(codes.Internal, "internal server error: %v", r)
		}
	}()

	return handler(ctx, req)
}

// ClientLoggingInterceptor logs client-side gRPC calls
func ClientLoggingInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()

	err := invoker(ctx, method, req, reply, cc, opts...)

	duration := time.Since(start)

	// Log in production
	_ = duration
	_ = method

	return err
}

// ClientRetryInterceptor implements retry logic on the client side
func ClientRetryInterceptor(maxRetries int, retryDelay time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var lastErr error

		for attempt := 0; attempt <= maxRetries; attempt++ {
			if attempt > 0 {
				select {
				case <-time.After(retryDelay):
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			err := invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}

			lastErr = err

			// Check if error is retryable
			st, ok := status.FromError(err)
			if !ok {
				// Unknown error, don't retry
				return err
			}

			// Don't retry on certain status codes
			switch st.Code() {
			case codes.InvalidArgument,
				codes.AlreadyExists,
				codes.PermissionDenied,
				codes.Unauthenticated,
				codes.NotFound,
				codes.FailedPrecondition:
				return err
			}
		}

		return lastErr
	}
}

// TimeoutInterceptor adds a timeout to gRPC calls
func TimeoutInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// UnaryServerChain chains multiple unary server interceptors
func UnaryServerChain(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	n := len(interceptors)

	if n == 0 {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	if n == 1 {
		return interceptors[0]
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		buildChain := func(current grpc.UnaryServerInterceptor, next grpc.UnaryHandler) grpc.UnaryHandler {
			return func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return current(currentCtx, currentReq, info, next)
			}
		}

		chainHandler := handler
		for i := n - 1; i >= 0; i-- {
			chainHandler = buildChain(interceptors[i], chainHandler)
		}

		return chainHandler(ctx, req)
	}
}

// UnaryClientChain chains multiple unary client interceptors
func UnaryClientChain(interceptors ...grpc.UnaryClientInterceptor) grpc.UnaryClientInterceptor {
	n := len(interceptors)

	if n == 0 {
		return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
	}

	if n == 1 {
		return interceptors[0]
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		buildChain := func(current grpc.UnaryClientInterceptor, next grpc.UnaryInvoker) grpc.UnaryInvoker {
			return func(currentCtx context.Context, currentMethod string, currentReq, currentReply interface{}, currentConn *grpc.ClientConn, currentOpts ...grpc.CallOption) error {
				return current(currentCtx, currentMethod, currentReq, currentReply, currentConn, next, currentOpts...)
			}
		}

		chainInvoker := invoker
		for i := n - 1; i >= 0; i-- {
			chainInvoker = buildChain(interceptors[i], chainInvoker)
		}

		return chainInvoker(ctx, method, req, reply, cc, opts...)
	}
}
