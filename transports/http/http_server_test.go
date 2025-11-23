package http

import (
	"encoding/json"
	"go-boilerplate/configs"
	"go-boilerplate/datasources"
	"go-boilerplate/transports/http/handlers"
	"go-boilerplate/transports/http/middlewares"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPServer(t *testing.T) {
	tests := []struct {
		name     string
		setupCfg func(t *testing.T) *configs.Config
		setupDS  func(t *testing.T) *datasources.Datasources
		setupMW  func(t *testing.T) *middlewares.Middlewares
		setupH   func(t *testing.T) *handlers.Handlers
		validate func(t *testing.T, s *HTTPServer, cfg *configs.Config, ds *datasources.Datasources, mw *middlewares.Middlewares, h *handlers.Handlers)
	}{
		{
			name: "should_create_http_server_successfully",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-server"
				cfg.Server.HTTP.Prefork = false
				cfg.Server.HTTP.PrintRoutes = false
				cfg.Server.HTTP.RequestTimeout = 5 * time.Second
				cfg.Server.HTTP.GracefullyShutdownDuration = 10 * time.Second
				cfg.Server.HTTP.Port = 8080
				cfg.Server.LogLevel = 1
				return cfg
			},
			setupDS: func(t *testing.T) *datasources.Datasources {
				return &datasources.Datasources{}
			},
			setupMW: func(t *testing.T) *middlewares.Middlewares {
				return &middlewares.Middlewares{
					Recover:   &middlewares.RecoverMiddleware{},
					Tracer:    &middlewares.TracerMiddleware{},
					RequestID: &middlewares.RequestIDMiddleware{},
					Log:       &middlewares.LogMiddleware{},
					Timeout:   &middlewares.TimeoutMiddleware{},
				}
			},
			setupH: func(t *testing.T) *handlers.Handlers {
				return &handlers.Handlers{}
			},
			validate: func(t *testing.T, s *HTTPServer, cfg *configs.Config, ds *datasources.Datasources, mw *middlewares.Middlewares, h *handlers.Handlers) {
				assert.NotNil(t, s)
				assert.Equal(t, cfg, s.cfg)
				assert.NotNil(t, s.server)
				assert.Equal(t, ds, s.datasources)
				assert.Equal(t, mw, s.middlewares)
				assert.Equal(t, h, s.handlers)
			},
		},
		{
			name: "should_create_http_server_with_prefork_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "prefork-server"
				cfg.Server.HTTP.Prefork = true
				cfg.Server.HTTP.PrintRoutes = true
				cfg.Server.HTTP.RequestTimeout = 10 * time.Second
				cfg.Server.HTTP.GracefullyShutdownDuration = 5 * time.Second
				cfg.Server.HTTP.Port = 9090
				cfg.Server.LogLevel = 0
				return cfg
			},
			setupDS: func(t *testing.T) *datasources.Datasources {
				return &datasources.Datasources{}
			},
			setupMW: func(t *testing.T) *middlewares.Middlewares {
				return &middlewares.Middlewares{
					Recover:   &middlewares.RecoverMiddleware{},
					Tracer:    &middlewares.TracerMiddleware{},
					RequestID: &middlewares.RequestIDMiddleware{},
					Log:       &middlewares.LogMiddleware{},
					Timeout:   &middlewares.TimeoutMiddleware{},
				}
			},
			setupH: func(t *testing.T) *handlers.Handlers {
				return &handlers.Handlers{}
			},
			validate: func(t *testing.T, s *HTTPServer, cfg *configs.Config, ds *datasources.Datasources, mw *middlewares.Middlewares, h *handlers.Handlers) {
				assert.NotNil(t, s)
				assert.Equal(t, cfg, s.cfg)
				assert.NotNil(t, s.server)
				assert.Equal(t, ds, s.datasources)
				assert.Equal(t, mw, s.middlewares)
				assert.Equal(t, h, s.handlers)
			},
		},
		{
			name: "should_configure_fiber_app_with_custom_error_handler",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "custom-error-server"
				cfg.Server.HTTP.Prefork = false
				cfg.Server.HTTP.PrintRoutes = false
				cfg.Server.HTTP.RequestTimeout = 3 * time.Second
				cfg.Server.HTTP.GracefullyShutdownDuration = 7 * time.Second
				cfg.Server.HTTP.Port = 7070
				cfg.Server.LogLevel = 2
				return cfg
			},
			setupDS: func(t *testing.T) *datasources.Datasources {
				return &datasources.Datasources{}
			},
			setupMW: func(t *testing.T) *middlewares.Middlewares {
				return &middlewares.Middlewares{
					Recover:   &middlewares.RecoverMiddleware{},
					Tracer:    &middlewares.TracerMiddleware{},
					RequestID: &middlewares.RequestIDMiddleware{},
					Log:       &middlewares.LogMiddleware{},
					Timeout:   &middlewares.TimeoutMiddleware{},
				}
			},
			setupH: func(t *testing.T) *handlers.Handlers {
				return &handlers.Handlers{}
			},
			validate: func(t *testing.T, s *HTTPServer, cfg *configs.Config, ds *datasources.Datasources, mw *middlewares.Middlewares, h *handlers.Handlers) {
				assert.NotNil(t, s)
				assert.NotNil(t, s.server)

				assert.Equal(t, cfg, s.cfg)

				app := s.server
				app.Get("/test-error", func(c *fiber.Ctx) error {
					return fiber.NewError(500, "test error")
				})

				req := httptest.NewRequest("GET", "/test-error", nil)
				resp, err := app.Test(req, -1)
				assert.NoError(t, err)
				assert.Equal(t, 500, resp.StatusCode)

				var body map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&body)
				assert.NotNil(t, body)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupCfg(t)
			ds := tt.setupDS(t)
			mw := tt.setupMW(t)
			h := tt.setupH(t)

			s := NewHTTPServer(cfg, ds, mw, h)

			tt.validate(t, s, cfg, ds, mw, h)
		})
	}
}

func TestHTTPServer_setGlobalLog(t *testing.T) {
	tests := []struct {
		name     string
		setupCfg func(t *testing.T) *configs.Config
		validate func(t *testing.T, s *HTTPServer)
	}{
		{
			name: "should_set_global_log_with_info_level",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-server"
				cfg.Server.LogLevel = 1
				return cfg
			},
			validate: func(t *testing.T, s *HTTPServer) {

				assert.NotNil(t, s)
			},
		},
		{
			name: "should_set_global_log_with_debug_level",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-server"
				cfg.Server.LogLevel = 0
				return cfg
			},
			validate: func(t *testing.T, s *HTTPServer) {
				assert.NotNil(t, s)
			},
		},
		{
			name: "should_set_global_log_with_error_level",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-server"
				cfg.Server.LogLevel = 3
				return cfg
			},
			validate: func(t *testing.T, s *HTTPServer) {
				assert.NotNil(t, s)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupCfg(t)
			ds := &datasources.Datasources{}
			mw := &middlewares.Middlewares{
				Recover:   &middlewares.RecoverMiddleware{},
				Tracer:    &middlewares.TracerMiddleware{},
				RequestID: &middlewares.RequestIDMiddleware{},
				Log:       &middlewares.LogMiddleware{},
				Timeout:   &middlewares.TimeoutMiddleware{},
			}
			h := &handlers.Handlers{}

			s := NewHTTPServer(cfg, ds, mw, h)
			s.setGlobalLog()

			tt.validate(t, s)
		})
	}
}

func TestHTTPServer_setupGlobalMiddlewares(t *testing.T) {
	tests := []struct {
		name     string
		setupCfg func(t *testing.T) *configs.Config
		setupMW  func(t *testing.T) *middlewares.Middlewares
		validate func(t *testing.T, s *HTTPServer)
	}{
		{
			name: "should_setup_middlewares_without_swagger",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-server"
				cfg.Server.HTTP.CORS.AllowOrigins = "*"
				cfg.Server.HTTP.CORS.AllowMethods = "GET,POST,PUT,DELETE"
				cfg.Server.HTTP.Docs.Swagger.Enable = false
				cfg.Server.HTTP.RequestTimeout = 5 * time.Second
				return cfg
			},
			setupMW: func(t *testing.T) *middlewares.Middlewares {
				cfg := &configs.Config{}
				cfg.Server.HTTP.RequestTimeout = 5 * time.Second

				return &middlewares.Middlewares{
					Recover:   middlewares.NewRecoverMiddleware(),
					Tracer:    middlewares.NewTracerMiddleware(),
					RequestID: middlewares.NewRequestIDMiddleware(),
					Log:       middlewares.NewLogMiddleware(),
					Timeout:   middlewares.NewTimeoutMiddleware(cfg),
				}
			},
			validate: func(t *testing.T, s *HTTPServer) {
				assert.NotNil(t, s)
				assert.NotNil(t, s.server)
			},
		},
		{
			name: "should_setup_middlewares_with_swagger_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-server"
				cfg.Server.HTTP.CORS.AllowOrigins = "http://localhost:3000"
				cfg.Server.HTTP.CORS.AllowMethods = "GET,POST"
				cfg.Server.HTTP.Docs.Swagger.Enable = true
				cfg.Server.HTTP.Docs.Swagger.FilePath = "./docs/swagger/swagger.json"
				cfg.Server.HTTP.Docs.Swagger.Path = "/swagger"
				cfg.Server.HTTP.Docs.Swagger.Title = "API Docs"
				cfg.Server.HTTP.RequestTimeout = 10 * time.Second
				return cfg
			},
			setupMW: func(t *testing.T) *middlewares.Middlewares {
				cfg := &configs.Config{}
				cfg.Server.HTTP.RequestTimeout = 10 * time.Second

				return &middlewares.Middlewares{
					Recover:   middlewares.NewRecoverMiddleware(),
					Tracer:    middlewares.NewTracerMiddleware(),
					RequestID: middlewares.NewRequestIDMiddleware(),
					Log:       middlewares.NewLogMiddleware(),
					Timeout:   middlewares.NewTimeoutMiddleware(cfg),
				}
			},
			validate: func(t *testing.T, s *HTTPServer) {
				assert.NotNil(t, s)
				assert.NotNil(t, s.server)
			},
		},
		{
			name: "should_setup_middlewares_with_custom_cors",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-server"
				cfg.Server.HTTP.CORS.AllowOrigins = "http://example.com,http://another.com"
				cfg.Server.HTTP.CORS.AllowMethods = "GET,POST,PUT,PATCH,DELETE,OPTIONS"
				cfg.Server.HTTP.Docs.Swagger.Enable = false
				cfg.Server.HTTP.RequestTimeout = 3 * time.Second
				return cfg
			},
			setupMW: func(t *testing.T) *middlewares.Middlewares {
				cfg := &configs.Config{}
				cfg.Server.HTTP.RequestTimeout = 3 * time.Second

				return &middlewares.Middlewares{
					Recover:   middlewares.NewRecoverMiddleware(),
					Tracer:    middlewares.NewTracerMiddleware(),
					RequestID: middlewares.NewRequestIDMiddleware(),
					Log:       middlewares.NewLogMiddleware(),
					Timeout:   middlewares.NewTimeoutMiddleware(cfg),
				}
			},
			validate: func(t *testing.T, s *HTTPServer) {
				assert.NotNil(t, s)
				assert.NotNil(t, s.server)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupCfg(t)
			ds := &datasources.Datasources{}
			mw := tt.setupMW(t)
			h := &handlers.Handlers{}

			s := NewHTTPServer(cfg, ds, mw, h)
			s.setupGlobalMiddlewares()

			tt.validate(t, s)
		})
	}
}
