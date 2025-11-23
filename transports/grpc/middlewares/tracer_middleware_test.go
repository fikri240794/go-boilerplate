package middlewares

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestNewTracerMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		validate func(t *testing.T, mw *TracerMiddleware)
	}{
		{
			name: "should_create_new_tracer_middleware_successfully",
			validate: func(t *testing.T, mw *TracerMiddleware) {
				assert.NotNil(t, mw)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := NewTracerMiddleware()

			if tt.validate != nil {
				tt.validate(t, mw)
			}
		})
	}
}

func TestTracerMiddleware_Start(t *testing.T) {
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
			name: "should_handle_request_successfully_without_metadata",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return "success response", nil
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
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
			name: "should_handle_request_with_metadata",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return "success with metadata", nil
				}
			},
			setupContext: func(t *testing.T) context.Context {
				md := metadata.New(map[string]string{
					"traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			setupRequest: func(t *testing.T) interface{} {
				return "test request with trace"
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/TracedMethod",
				}
			},
			validateError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "success with metadata", res)
			},
		},
		{
			name: "should_handle_request_with_empty_metadata",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return "response", nil
				}
			},
			setupContext: func(t *testing.T) context.Context {
				md := metadata.New(map[string]string{})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			setupRequest: func(t *testing.T) interface{} {
				return "request"
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
				assert.Equal(t, "response", res)
			},
		},
		{
			name: "should_propagate_handler_error",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return nil, errors.New("handler error")
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
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
				assert.Error(t, err)
				assert.Equal(t, "handler error", err.Error())
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.Error(t, err)
				assert.Nil(t, res)
			},
		},
		{
			name: "should_handle_request_with_multiple_metadata_values",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return "multi metadata response", nil
				}
			},
			setupContext: func(t *testing.T) context.Context {
				md := metadata.New(map[string]string{
					"traceparent":   "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
					"tracestate":    "congo=t61rcWkgMzE",
					"custom-header": "custom-value",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			setupRequest: func(t *testing.T) interface{} {
				return map[string]string{"key": "value"}
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/ComplexMethod",
				}
			},
			validateError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "multi metadata response", res)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := NewTracerMiddleware()

			ctx := tt.setupContext(t)
			req := tt.setupRequest(t)
			info := tt.setupInfo(t)
			handler := tt.setupHandler(t)

			res, err := mw.Start(ctx, req, info, handler)

			if tt.validateError != nil {
				tt.validateError(t, err)
			}
			if tt.validate != nil {
				tt.validate(t, res, err)
			}
		})
	}
}
