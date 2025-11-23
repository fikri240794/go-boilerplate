package middlewares

import (
	"context"
	"errors"
	"go-boilerplate/configs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewTimeoutMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		setupCfg func(t *testing.T) *configs.Config
		validate func(t *testing.T, mw *TimeoutMiddleware)
	}{
		{
			name: "should_create_new_timeout_middleware_with_config",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.GRPC.RequestTimeout = 5 * time.Second
				return cfg
			},
			validate: func(t *testing.T, mw *TimeoutMiddleware) {
				assert.NotNil(t, mw)
				assert.NotNil(t, mw.cfg)
				assert.Equal(t, 5*time.Second, mw.cfg.Server.GRPC.RequestTimeout)
			},
		},
		{
			name: "should_create_new_timeout_middleware_with_nil_config",
			setupCfg: func(t *testing.T) *configs.Config {
				return nil
			},
			validate: func(t *testing.T, mw *TimeoutMiddleware) {
				assert.NotNil(t, mw)
				assert.Nil(t, mw.cfg)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupCfg(t)
			mw := NewTimeoutMiddleware(cfg)

			if tt.validate != nil {
				tt.validate(t, mw)
			}
		})
	}
}

func TestTimeoutMiddleware_Timeout(t *testing.T) {
	tests := []struct {
		name          string
		setupCfg      func(t *testing.T) *configs.Config
		setupHandler  func(t *testing.T) grpc.UnaryHandler
		setupContext  func(t *testing.T) context.Context
		setupRequest  func(t *testing.T) interface{}
		setupInfo     func(t *testing.T) *grpc.UnaryServerInfo
		validateError func(t *testing.T, err error)
		validate      func(t *testing.T, res interface{}, err error)
	}{
		{
			name: "should_handle_request_successfully_within_timeout",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.GRPC.RequestTimeout = 1 * time.Second
				return cfg
			},
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
			name: "should_return_deadline_exceeded_when_handler_takes_too_long",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.GRPC.RequestTimeout = 100 * time.Millisecond
				return cfg
			},
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {

					time.Sleep(200 * time.Millisecond)
					return "delayed response", nil
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			setupRequest: func(t *testing.T) interface{} {
				return "slow request"
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/SlowMethod",
				}
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, codes.DeadlineExceeded, st.Code())
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.Error(t, err)
				assert.Nil(t, res)
			},
		},
		{
			name: "should_propagate_handler_error_within_timeout",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.GRPC.RequestTimeout = 1 * time.Second
				return cfg
			},
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
					FullMethod: "/test.Service/ErrorMethod",
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
			name: "should_handle_fast_response_with_short_timeout",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.GRPC.RequestTimeout = 50 * time.Millisecond
				return cfg
			},
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {

					return "fast response", nil
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			setupRequest: func(t *testing.T) interface{} {
				return "fast request"
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/FastMethod",
				}
			},
			validateError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "fast response", res)
			},
		},
		{
			name: "should_handle_grpc_status_error_within_timeout",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.GRPC.RequestTimeout = 1 * time.Second
				return cfg
			},
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return nil, status.Error(codes.InvalidArgument, "invalid argument")
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			setupRequest: func(t *testing.T) interface{} {
				return map[string]string{"field": "invalid"}
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/ValidationMethod",
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
			name: "should_handle_request_with_long_timeout",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.GRPC.RequestTimeout = 5 * time.Second
				return cfg
			},
			setupHandler: func(t *testing.T) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {

					time.Sleep(50 * time.Millisecond)
					return "moderate response", nil
				}
			},
			setupContext: func(t *testing.T) context.Context {
				return context.Background()
			},
			setupRequest: func(t *testing.T) interface{} {
				return "moderate request"
			},
			setupInfo: func(t *testing.T) *grpc.UnaryServerInfo {
				return &grpc.UnaryServerInfo{
					FullMethod: "/test.Service/ModerateMethod",
				}
			},
			validateError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, res interface{}, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "moderate response", res)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupCfg(t)
			mw := NewTimeoutMiddleware(cfg)

			ctx := tt.setupContext(t)
			req := tt.setupRequest(t)
			info := tt.setupInfo(t)
			handler := tt.setupHandler(t)

			res, err := mw.Timeout(ctx, req, info, handler)

			if tt.validateError != nil {
				tt.validateError(t, err)
			}
			if tt.validate != nil {
				tt.validate(t, res, err)
			}
		})
	}
}
