package interceptors

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRetryInterceptor(t *testing.T) {
	tests := []struct {
		name            string
		maxRetries      int
		errorSequence   []error // sequence of errors returned by invoker
		expectedErrCode codes.Code
		expectSuccess   bool
	}{
		{
			name:            "NoRetriesNeeded",
			maxRetries:      3,
			errorSequence:   []error{nil},
			expectedErrCode: codes.OK,
			expectSuccess:   true,
		},
		{
			name:            "RetryOnUnavailable",
			maxRetries:      2,
			errorSequence:   []error{status.Error(codes.Unavailable, "transient"), nil},
			expectedErrCode: codes.OK,
			expectSuccess:   true,
		},
		{
			name:       "RetryOnDeadlineExceeded",
			maxRetries: 2,
			errorSequence: []error{
				status.Error(codes.DeadlineExceeded, "timeout"),
				status.Error(codes.DeadlineExceeded, "timeout"),
				status.Error(codes.DeadlineExceeded, "timeout"),
			},
			expectedErrCode: codes.DeadlineExceeded,
			expectSuccess:   false,
		},
		{
			name:            "NonRetryableError",
			maxRetries:      2,
			errorSequence:   []error{status.Error(codes.InvalidArgument, "bad input")},
			expectedErrCode: codes.InvalidArgument,
			expectSuccess:   false,
		},
		{
			name:            "GenericError_NotGRPC",
			maxRetries:      1,
			errorSequence:   []error{errors.New("some error")},
			expectedErrCode: codes.Unknown,
			expectSuccess:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			interceptor := RetryInterceptor(tt.maxRetries)

			invoker := func(ctx context.Context, method string, req, reply interface{},
				cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				defer func() { callCount++ }()
				if callCount < len(tt.errorSequence) {
					return tt.errorSequence[callCount]
				}
				return nil
			}

			err := interceptor(
				context.Background(),
				"/test.Service/Method",
				nil,
				nil,
				nil,
				invoker,
			)

			if tt.expectSuccess {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedErrCode, st.Code())
			}
		})
	}
}
