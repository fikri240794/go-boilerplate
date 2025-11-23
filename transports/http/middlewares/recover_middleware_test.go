package middlewares

import (
	"context"
	"go-boilerplate/pkg/constants"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewRecoverMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		validate func(t *testing.T, middleware *RecoverMiddleware)
	}{
		{
			name: "should_create_recover_middleware_successfully",
			validate: func(t *testing.T, middleware *RecoverMiddleware) {
				assert.NotNil(t, middleware)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRecoverMiddleware()

			if tt.validate != nil {
				tt.validate(t, middleware)
			}
		})
	}
}

func TestRecoverMiddleware_Recover(t *testing.T) {
	tests := []struct {
		name           string
		setupRequest   func(t *testing.T) *http.Request
		setupHandler   func(t *testing.T) fiber.Handler
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response)
	}{
		{
			name: "should_continue_to_next_handler_when_no_panic",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				ctx := req.Context()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {
					return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "success"})
				}
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "should_recover_from_panic_and_return_500",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				ctx := req.Context()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {
					panic("test panic")
				}
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
			},
		},
		{
			name: "should_recover_from_panic_without_context_values",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {
					panic("test panic without context")
				}
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
			},
		},
		{
			name: "should_recover_from_panic_with_request_body",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/test", httptest.NewRequest(http.MethodPost, "/", nil).Body)
				req.Header.Set("Content-Type", "application/json")
				ctx := req.Context()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {
					panic("test panic with body")
				}
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
			},
		},
		{
			name: "should_recover_from_nil_panic",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				ctx := req.Context()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {
					panic("nil pointer dereference")
				}
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRecoverMiddleware()
			app := fiber.New()

			app.Use(middleware.Recover)
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
