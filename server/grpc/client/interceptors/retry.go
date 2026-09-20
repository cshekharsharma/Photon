package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RetryInterceptor retries gRPC requests on transient errors with exponential backoff.
func RetryInterceptor(maxRetries int) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		var err error
		backoff := 100 * time.Millisecond

		for i := 0; i <= maxRetries; i++ {
			err = invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}

			st, ok := status.FromError(err)
			if !ok {
				// Not a gRPC status error — do not retry
				return err
			}

			switch st.Code() {
			case codes.Unavailable, codes.ResourceExhausted, codes.DeadlineExceeded:
				if i < maxRetries {
					timer := time.NewTimer(backoff)
					select {
					case <-ctx.Done():
						timer.Stop()
						return ctx.Err()
					case <-timer.C:
					}
					backoff *= 2 // Exponential backoff
					continue
				}
			default:
				return err // Non-retryable error
			}
		}

		return err
	}
}
