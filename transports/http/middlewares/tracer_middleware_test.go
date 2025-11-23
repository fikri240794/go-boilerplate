package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewTracerMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		validate func(t *testing.T, middleware *TracerMiddleware)
	}{
		{
			name: "should_create_tracer_middleware_successfully",
			validate: func(t *testing.T, middleware *TracerMiddleware) {
				assert.NotNil(t, middleware)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewTracerMiddleware()

			if tt.validate != nil {
				tt.validate(t, middleware)
			}
		})
	}
}

func TestTracerMiddleware_Start(t *testing.T) {
	tests := []struct {
		name           string
		setupRequest   func(t *testing.T) *http.Request
		setupHandler   func(t *testing.T) fiber.Handler
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response)
	}{
		{
			name: "should_extract_context_and_continue_to_next_handler",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {

					assert.NotNil(t, c.UserContext())
					return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "success"})
				}
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "should_extract_context_with_trace_headers",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
				req.Header.Set("tracestate", "congo=t61rcWkgMzE")
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {

					assert.NotNil(t, c.UserContext())
					return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "success"})
				}
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "should_handle_request_without_trace_headers",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {

					assert.NotNil(t, c.UserContext())
					return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "success"})
				}
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "should_extract_context_with_custom_headers",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/test", nil)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {
					assert.NotNil(t, c.UserContext())
					return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "created"})
				}
			},
			expectedStatus: fiber.StatusCreated,
			validate: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
			},
		},
		{
			name: "should_continue_to_next_middleware_after_extracting_context",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {

					ctx := c.UserContext()
					assert.NotNil(t, ctx)
					return c.Status(fiber.StatusOK).SendString("OK")
				}
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewTracerMiddleware()
			app := fiber.New()

			app.Use(middleware.Start)
			app.Get("/test", tt.setupHandler(t))
			app.Post("/test", tt.setupHandler(t))

			req := tt.setupRequest(t)
			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}
