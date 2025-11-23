package middlewares

import (
	"context"
	"encoding/json"
	"go-boilerplate/configs"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewTimeoutMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		setupCfg func(t *testing.T) *configs.Config
		validate func(t *testing.T, mw *TimeoutMiddleware, cfg *configs.Config)
	}{
		{
			name: "should_create_timeout_middleware_successfully",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.HTTP.RequestTimeout = 5 * time.Second
				return cfg
			},
			validate: func(t *testing.T, mw *TimeoutMiddleware, cfg *configs.Config) {
				assert.NotNil(t, mw)
				assert.Equal(t, cfg, mw.cfg)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupCfg(t)
			mw := NewTimeoutMiddleware(cfg)
			tt.validate(t, mw, cfg)
		})
	}
}

func TestTimeoutMiddleware_Timeout(t *testing.T) {
	tests := []struct {
		name           string
		setupCfg       func(t *testing.T) *configs.Config
		setupRequest   func(t *testing.T) *http.Request
		setupHandler   func(t *testing.T) fiber.Handler
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request)
	}{
		{
			name: "should_set_timeout_context_successfully",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.HTTP.RequestTimeout = 5 * time.Second
				return cfg
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {
					ctx := c.UserContext()

					deadline, ok := ctx.Deadline()
					assert.True(t, ok, "Context should have a deadline")
					assert.True(t, time.Until(deadline) > 0, "Deadline should be in the future")

					return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "success"})
				}
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)

				var body map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, "success", body["message"])
			},
		},
		{
			name: "should_continue_to_next_handler",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.HTTP.RequestTimeout = 10 * time.Second
				return cfg
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/test", nil)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {
					return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": "123"})
				}
			},
			expectedStatus: fiber.StatusCreated,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {
				assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

				var body map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, "123", body["id"])
			},
		},
		{
			name: "should_handle_quick_response",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.HTTP.RequestTimeout = 3 * time.Second
				return cfg
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {

					return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
				}
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)

				var body map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, "ok", body["status"])
			},
		},
		{
			name: "should_set_timeout_with_short_duration",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.HTTP.RequestTimeout = 1 * time.Second
				return cfg
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {
					ctx := c.UserContext()

					deadline, ok := ctx.Deadline()
					assert.True(t, ok)

					timeUntilDeadline := time.Until(deadline)
					assert.True(t, timeUntilDeadline > 0 && timeUntilDeadline <= 1*time.Second)

					return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "done"})
				}
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "should_preserve_existing_context_values",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.HTTP.RequestTimeout = 5 * time.Second
				return cfg
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {

					type contextKey string
					testKey := contextKey("test-key")
					testValue := "test-value"
					ctx := context.WithValue(c.UserContext(), testKey, testValue)
					c.SetUserContext(ctx)

					return c.Next()
				}
			},
			expectedStatus: fiber.StatusNotFound,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {

				assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
			},
		},
		{
			name: "should_set_timeout_with_long_duration",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.HTTP.RequestTimeout = 30 * time.Second
				return cfg
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {
					ctx := c.UserContext()

					deadline, ok := ctx.Deadline()
					assert.True(t, ok)

					timeUntilDeadline := time.Until(deadline)
					assert.True(t, timeUntilDeadline > 25*time.Second && timeUntilDeadline <= 30*time.Second)

					return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "completed"})
				}
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			cfg := tt.setupCfg(t)
			middleware := NewTimeoutMiddleware(cfg)
			app.Use(middleware.Timeout)

			handler := tt.setupHandler(t)
			app.All("/test", handler)

			req := tt.setupRequest(t)
			resp, err := app.Test(req, -1)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.validate != nil {
				tt.validate(t, resp, app, req)
			}
		})
	}
}
