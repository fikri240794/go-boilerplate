package middlewares

import (
	"context"
	"errors"
	"go-boilerplate/pkg/constants"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewLogMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		validate func(t *testing.T, mw *LogMiddleware)
	}{
		{
			name: "should_create_new_log_middleware_successfully",
			validate: func(t *testing.T, mw *LogMiddleware) {
				assert.NotNil(t, mw)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := NewLogMiddleware()

			if tt.validate != nil {
				tt.validate(t, mw)
			}
		})
	}
}

func TestLogMiddleware_Log(t *testing.T) {
	tests := []struct {
		name          string
		setupHandler  func(t *testing.T) grpc.UnaryHandler
		setupContext  func(t *testing.T) context.Context
		setupRequest  func(t *testing.T) interface{}
		setupInfo     func(t *testing.T) *grpc.UnaryServerInfo
		validateError func(t *testing.T, err error)
		validate      func(t *testing.T, res interface{}, err error)
	}{
		{
			name: "should_log_successful_request_with_info_level",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return "success response", nil
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id-123")
			},
			setupRequest: func(t *testing.T) interface{} {
				return "test request"
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/Method",
				}
			},
			validateError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "success response", res)
			},
		},
		{
			name: "should_log_request_with_4xx_error_as_info_level",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return nil, status.Error(codes.InvalidArgument, "validation error")
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id-456")
			},
			setupRequest: func(t *testing.T) interface{} {
				return map[string]string{"field": "invalid"}
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/ValidateMethod",
				}
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, codes.InvalidArgument, st.Code())
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.Error(t, err)
				assert.Nil(t, res)
			},
		},
		{
			name: "should_log_request_with_5xx_error_as_error_level",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return nil, status.Error(codes.Internal, "internal server error")
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id-789")
			},
			setupRequest: func(t *testing.T) interface{} {
				return "test request"
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/FailingMethod",
				}
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, codes.Internal, st.Code())
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.Error(t, err)
				assert.Nil(t, res)
			},
		},
		{
			name: "should_log_request_with_unavailable_error_as_error_level",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return nil, status.Error(codes.Unavailable, "service unavailable")
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id-unavail")
			},
			setupRequest: func(t *testing.T) interface{} {
				return "unavailable request"
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/UnavailableMethod",
				}
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.Error(t, err)
				assert.Nil(t, res)
			},
		},
		{
			name: "should_log_request_with_not_found_error_as_info_level",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return nil, status.Error(codes.NotFound, "resource not found")
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id-notfound")
			},
			setupRequest: func(t *testing.T) interface{} {
				return map[string]string{"id": "non-existent"}
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/FindMethod",
				}
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.Error(t, err)
				assert.Nil(t, res)
			},
		},
		{
			name: "should_log_request_with_non_grpc_error_as_info_level",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return nil, errors.New("standard error")
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id-std")
			},
			setupRequest: func(t *testing.T) interface{} {
				return "test request"
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/StandardErrorMethod",
				}
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Equal(t, "standard error", err.Error())
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.Error(t, err)
				assert.Nil(t, res)
			},
		},
		{
			name: "should_log_request_without_request_id_in_context",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return "response without request id", nil
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			setupRequest: func(t *testing.T) interface{} {
				return "request without id"
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/NoRequestIDMethod",
				}
			},
			validateError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "response without request id", res)
			},
		},
		{
			name: "should_measure_latency_correctly",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {

					return "latency test response", nil
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-latency-id")
			},
			setupRequest: func(t *testing.T) interface{} {
				return "latency test request"
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/LatencyMethod",
				}
			},
			validateError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "latency test response", res)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := NewLogMiddleware()

			ctx := tt.setupContext(t)
			req := tt.setupRequest(t)
			info := tt.setupInfo(t)
			handler := tt.setupHandler(t)

			res, err := mw.Log(ctx, req, info, handler)

			if tt.validateError != nil {
				tt.validateError(t, err)
			}
			if tt.validate != nil {
				tt.validate(t, res, err)
			}
		})
	}
}
