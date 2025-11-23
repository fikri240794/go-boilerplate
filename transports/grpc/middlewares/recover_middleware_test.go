package middlewares

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestNewRecoverMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		validate func(t *testing.T, mw *RecoverMiddleware)
	}{
		{
			name: "should_create_new_recover_middleware_successfully",
			validate: func(t *testing.T, mw *RecoverMiddleware) {
				assert.NotNil(t, mw)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := NewRecoverMiddleware()

			if tt.validate != nil {
				tt.validate(t, mw)
			}
		})
	}
}

func TestRecoverMiddleware_Recover(t *testing.T) {
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
			name: "should_handle_request_successfully_without_panic",
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
			name: "should_handle_request_with_handler_error",
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
			name: "should_recover_from_panic_and_return_grpc_error",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					panic("something went wrong")
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
				assert.Contains(t, err.Error(), "Internal")
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.Error(t, err)
				assert.Nil(t, res)
			},
		},
		{
			name: "should_recover_from_panic_with_nil_value",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					panic("nil pointer dereference")
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
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.Error(t, err)
				assert.Nil(t, res)
			},
		},
		{
			name: "should_recover_from_panic_with_custom_error",
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					panic(errors.New("custom panic error"))
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			setupRequest: func(t *testing.T) interface{} {
				return map[string]string{"key": "value"}
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/AnotherMethod",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := NewRecoverMiddleware()

			ctx := tt.setupContext(t)
			req := tt.setupRequest(t)
			info := tt.setupInfo(t)
			handler := tt.setupHandler(t)

			res, err := mw.Recover(ctx, req, info, handler)

			if tt.validateError != nil {
				tt.validateError(t, err)
			}
			if tt.validate != nil {
				tt.validate(t, res, err)
			}
		})
	}
}
