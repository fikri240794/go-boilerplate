package grpc_error

import (
	"errors"
	"net/http"
	"testing"

	"github.com/fikri240794/gocerr"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFromError(t *testing.T) {
	tests := []struct {
		name             string
		inputError       error
		expectedGRPCCode codes.Code
		expectedMessage  string
		expectNil        bool
	}{
		{
			name:       "nil error should return nil",
			inputError: nil,
			expectNil:  true,
		},
		{
			name:             "non-custom error should return Internal code with empty message",
			inputError:       errors.New("some error"),
			expectedGRPCCode: codes.Internal,
			expectedMessage:  "",
		},
		{
			name:             "custom error with 400 should return InvalidArgument code",
			inputError:       gocerr.New(http.StatusBadRequest, "bad request message"),
			expectedGRPCCode: codes.InvalidArgument,
			expectedMessage:  "bad request message",
		},
		{
			name:             "custom error with 401 should return Unauthenticated code",
			inputError:       gocerr.New(http.StatusUnauthorized, "unauthorized message"),
			expectedGRPCCode: codes.Unauthenticated,
			expectedMessage:  "unauthorized message",
		},
		{
			name:             "custom error with 403 should return PermissionDenied code",
			inputError:       gocerr.New(http.StatusForbidden, "forbidden message"),
			expectedGRPCCode: codes.PermissionDenied,
			expectedMessage:  "forbidden message",
		},
		{
			name:             "custom error with 404 should return NotFound code",
			inputError:       gocerr.New(http.StatusNotFound, "not found message"),
			expectedGRPCCode: codes.NotFound,
			expectedMessage:  "not found message",
		},
		{
			name:             "custom error with 409 should return AlreadyExists code",
			inputError:       gocerr.New(http.StatusConflict, "conflict error"),
			expectedGRPCCode: codes.AlreadyExists,
			expectedMessage:  "conflict error",
		},
		{
			name:             "custom error with 429 should return ResourceExhausted code",
			inputError:       gocerr.New(http.StatusTooManyRequests, "too many requests"),
			expectedGRPCCode: codes.ResourceExhausted,
			expectedMessage:  "too many requests",
		},
		{
			name:             "custom error with 500 should return Internal code",
			inputError:       gocerr.New(http.StatusInternalServerError, "internal error"),
			expectedGRPCCode: codes.Internal,
			expectedMessage:  "internal error",
		},
		{
			name:             "custom error with 501 should return Unimplemented code",
			inputError:       gocerr.New(http.StatusNotImplemented, "not implemented"),
			expectedGRPCCode: codes.Unimplemented,
			expectedMessage:  "not implemented",
		},
		{
			name:             "custom error with 503 should return Unavailable code",
			inputError:       gocerr.New(http.StatusServiceUnavailable, "service unavailable"),
			expectedGRPCCode: codes.Unavailable,
			expectedMessage:  "service unavailable",
		},
		{
			name:             "custom error with 504 should return DeadlineExceeded code",
			inputError:       gocerr.New(http.StatusGatewayTimeout, "gateway timeout"),
			expectedGRPCCode: codes.DeadlineExceeded,
			expectedMessage:  "gateway timeout",
		},
		{
			name: "custom error with ErrorFields should use first field message",
			inputError: gocerr.New(
				http.StatusBadRequest,
				"main error message",
				gocerr.NewErrorField("email", "invalid email format"),
				gocerr.NewErrorField("name", "name is required"),
			),
			expectedGRPCCode: codes.InvalidArgument,
			expectedMessage:  "invalid email format",
		},
		{
			name: "custom error with single ErrorField should use field message",
			inputError: gocerr.New(
				http.StatusBadRequest,
				"validation error",
				gocerr.NewErrorField("username", "username already exists"),
			),
			expectedGRPCCode: codes.InvalidArgument,
			expectedMessage:  "username already exists",
		},
		{
			name:             "custom error with empty ErrorFields should use main message",
			inputError:       gocerr.New(http.StatusNotFound, "resource not found"),
			expectedGRPCCode: codes.NotFound,
			expectedMessage:  "resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromError(tt.inputError)

			if tt.expectNil {
				assert.Nil(t, result, "Expected nil")
				return
			}

			assert.NotNil(t, result, "Expected error, got nil")

			st, ok := status.FromError(result)
			assert.True(t, ok, "Expected gRPC status error")

			assert.Equal(t, tt.expectedGRPCCode, st.Code(), "Expected gRPC code %v", tt.expectedGRPCCode)
			assert.Equal(t, tt.expectedMessage, st.Message(), "Expected message '%s'", tt.expectedMessage)
		})
	}
}
