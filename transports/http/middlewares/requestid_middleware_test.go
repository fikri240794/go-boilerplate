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

func TestNewRequestIDMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		validate func(t *testing.T, middleware *RequestIDMiddleware)
	}{
		{
			name: "should_create_requestid_middleware_successfully",
			validate: func(t *testing.T, middleware *RequestIDMiddleware) {
				assert.NotNil(t, middleware)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRequestIDMiddleware()

			if tt.validate != nil {
				tt.validate(t, middleware)
			}
		})
	}
}

func TestRequestIDMiddleware_Generate(t *testing.T) {
	tests := []struct {
		name           string
		setupRequest   func(t *testing.T) *http.Request
		setupHandler   func(t *testing.T) fiber.Handler
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request)
	}{
		{
			name: "should_generate_new_request_id_when_not_provided",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {

					requestID := c.UserContext().Value(constants.ContextKeyRequestID)
					assert.NotNil(t, requestID)
					assert.NotEmpty(t, requestID)

					responseRequestID := c.Response().Header.Peek(constants.HeaderKeyRequestID)
					assert.NotEmpty(t, string(responseRequestID))

					assert.Equal(t, requestID, string(responseRequestID))

					return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "success"})
				}
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)

				requestID := resp.Header.Get(constants.HeaderKeyRequestID)
				assert.NotEmpty(t, requestID)
			},
		},
		{
			name: "should_use_existing_request_id_when_provided",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set(constants.HeaderKeyRequestID, "existing-request-id-123")
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {

					requestID := c.UserContext().Value(constants.ContextKeyRequestID)
					assert.Equal(t, "existing-request-id-123", requestID)

					headerRequestID := c.Get(constants.HeaderKeyRequestID)
					assert.Equal(t, "existing-request-id-123", headerRequestID)

					return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "success"})
				}
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)

				requestID := resp.Header.Get(constants.HeaderKeyRequestID)
				assert.Equal(t, "existing-request-id-123", requestID)
			},
		},
		{
			name: "should_set_request_id_in_context_and_header",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/test", nil)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {

					ctx := c.UserContext()
					requestID := ctx.Value(constants.ContextKeyRequestID)
					assert.NotNil(t, requestID)

					headerRequestID := c.Response().Header.Peek(constants.HeaderKeyRequestID)
					assert.NotEmpty(t, string(headerRequestID))

					return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": "123"})
				}
			},
			expectedStatus: fiber.StatusCreated,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {
				assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
				assert.NotEmpty(t, resp.Header.Get(constants.HeaderKeyRequestID))
			},
		},
		{
			name: "should_continue_to_next_handler_after_generating_request_id",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {

					requestID := c.UserContext().Value(constants.ContextKeyRequestID)
					assert.NotNil(t, requestID)
					return c.SendString("OK")
				}
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "should_handle_empty_request_id_header",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set(constants.HeaderKeyRequestID, "")
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {

					requestID := c.UserContext().Value(constants.ContextKeyRequestID)
					assert.NotNil(t, requestID)
					assert.NotEmpty(t, requestID)

					headerRequestID := c.Response().Header.Peek(constants.HeaderKeyRequestID)
					assert.NotEmpty(t, string(headerRequestID))

					return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "success"})
				}
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)

				requestID := resp.Header.Get(constants.HeaderKeyRequestID)
				assert.NotEmpty(t, requestID)
			},
		},
		{
			name: "should_preserve_user_context_and_add_request_id",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				type contextKey string
				ctx := context.WithValue(req.Context(), contextKey("custom-key"), "custom-value")
				return req.WithContext(ctx)
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {

					requestID := c.UserContext().Value(constants.ContextKeyRequestID)
					assert.NotNil(t, requestID)

					return c.Status(fiber.StatusOK).SendString("OK")
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
			middleware := NewRequestIDMiddleware()
			app := fiber.New()

			app.Use(middleware.Generate)
			app.Get("/test", tt.setupHandler(t))
			app.Post("/test", tt.setupHandler(t))

			req := tt.setupRequest(t)
			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.validate != nil {
				tt.validate(t, resp, app, req)
			}
		})
	}
}
