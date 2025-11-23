package middlewares

import (
	"context"
	"errors"
	"go-boilerplate/pkg/constants"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestNewRequestIDMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		validate func(t *testing.T, mw *RequestIDMiddleware)
	}{
		{
			name: "should_create_new_requestid_middleware_successfully",
			validate: func(t *testing.T, mw *RequestIDMiddleware) {
				assert.NotNil(t, mw)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := NewRequestIDMiddleware()

			if tt.validate != nil {
				tt.validate(t, mw)
			}
		})
	}
}

func TestRequestIDMiddleware_Generate(t *testing.T) {
	tests := []struct {
		name          string
		setupHandler  func(t *testing.T) grpc.UnaryHandler
		setupContext  func(t *testing.T) context.Context
		setupRequest  func(t *testing.T) interface{}
		setupInfo     func(t *testing.T) *grpc.UnaryServerInfo
		validateError func(t *testing.T, err error)
		validate      func(t *testing.T, ctx context.Context, res interface{}, err error)
	}{
		{
			name: "should_generate_new_request_id_when_not_present_in_metadata",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {

					requestID := ctx.Value(constants.ContextKeyRequestID)
					assert.NotNil(t, requestID)
					assert.NotEmpty(t, requestID)
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
			validate: func(t *testing.T, ctx context.Context, res interface{}, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "success response", res)
			},
		},
		{
			name: "should_use_existing_request_id_from_metadata",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return "response with existing id", nil
				}
			},
			setupContext: func(t *testing.T) context.Context {
				md := metadata.New(map[string]string{
					constants.HeaderKeyRequestID: "existing-request-id-123",
				})
				return metadata.NewIncomingContext(context.Background(), md)
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
			validate: func(t *testing.T, ctx context.Context, res interface{}, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "response with existing id", res)
			},
		},
		{
			name: "should_generate_request_id_when_metadata_present_but_empty",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					requestID := ctx.Value(constants.ContextKeyRequestID)
					assert.NotNil(t, requestID)
					assert.NotEmpty(t, requestID)
					return "response", nil
				}
			},
			setupContext: func(t *testing.T) context.Context {
				md := metadata.New(map[string]string{
					"other-header": "value",
				})
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
			validate: func(t *testing.T, ctx context.Context, res interface{}, err error) {
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
			validate: func(t *testing.T, ctx context.Context, res interface{}, err error) {
				assert.Error(t, err)
				assert.Nil(t, res)
			},
		},
		{
			name: "should_handle_empty_request_id_in_metadata",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					requestID := ctx.Value(constants.ContextKeyRequestID)
					assert.NotNil(t, requestID)
					assert.NotEmpty(t, requestID)
					return "new id generated", nil
				}
			},
			setupContext: func(t *testing.T) context.Context {
				md := metadata.New(map[string]string{
					constants.HeaderKeyRequestID: "",
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
			validate: func(t *testing.T, ctx context.Context, res interface{}, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "new id generated", res)
			},
		},
		{
			name: "should_handle_context_without_incoming_metadata",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					requestID := ctx.Value(constants.ContextKeyRequestID)
					assert.NotNil(t, requestID)
					assert.NotEmpty(t, requestID)
					return "success", nil
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			setupRequest: func(t *testing.T) interface{} {
				return "simple request"
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/SimpleMethod",
				}
			},
			validateError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, ctx context.Context, res interface{}, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "success", res)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := NewRequestIDMiddleware()

			ctx := tt.setupContext(t)
			req := tt.setupRequest(t)
			info := tt.setupInfo(t)
			handler := tt.setupHandler(t)

			res, err := mw.Generate(ctx, req, info, handler)

			if tt.validateError != nil {
				tt.validateError(t, err)
			}
			if tt.validate != nil {
				tt.validate(t, ctx, res, err)
			}
		})
	}
}
