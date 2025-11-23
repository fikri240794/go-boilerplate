package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"go-boilerplate/pkg/constants"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewLogMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		validate func(t *testing.T, mw *LogMiddleware)
	}{
		{
			name: "should_create_log_middleware_successfully",
			validate: func(t *testing.T, mw *LogMiddleware) {
				assert.NotNil(t, mw)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := NewLogMiddleware()
			tt.validate(t, mw)
		})
	}
}

func TestLogMiddleware_Log(t *testing.T) {
	tests := []struct {
		name           string
		setupRequest   func(t *testing.T) *http.Request
		setupHandler   func(t *testing.T) fiber.Handler
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request)
	}{
		{
			name: "should_log_successful_request_with_info_level",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("User-Agent", "test-agent")
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {

					ctx := context.WithValue(c.UserContext(), constants.ContextKeyRequestID, "test-request-id")
					c.SetUserContext(ctx)
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
			name: "should_log_with_warn_level_for_4xx_errors",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(`{"invalid": "data"}`))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
				}
			},
			expectedStatus: fiber.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {
				assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

				var body map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, "bad request", body["error"])
			},
		},
		{
			name: "should_log_with_error_level_for_5xx_errors",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
				}
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {
				assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

				var body map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, "internal server error", body["error"])
			},
		},
		{
			name: "should_log_with_error_level_when_handler_returns_error",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {
					return fiber.NewError(fiber.StatusInternalServerError, "handler error")
				}
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {
				assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
			},
		},
		{
			name: "should_log_request_and_response_details_with_query_params",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test?page=1&limit=10", nil)
				req.Header.Set("Authorization", "Bearer token")
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {
					ctx := context.WithValue(c.UserContext(), constants.ContextKeyRequestID, "query-test-id")
					c.SetUserContext(ctx)

					page := c.Query("page")
					limit := c.Query("limit")

					return c.Status(fiber.StatusOK).JSON(fiber.Map{
						"page":  page,
						"limit": limit,
					})
				}
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)

				var body map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, "1", body["page"])
				assert.Equal(t, "10", body["limit"])
			},
		},
		{
			name: "should_log_request_with_body",
			setupRequest: func(t *testing.T) *http.Request {
				reqBody := map[string]interface{}{
					"name":  "John Doe",
					"email": "john@example.com",
				}
				bodyBytes, _ := json.Marshal(reqBody)
				req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {
					var body map[string]interface{}
					if err := c.BodyParser(&body); err != nil {
						return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
					}

					return c.Status(fiber.StatusCreated).JSON(fiber.Map{
						"message": "created",
						"data":    body,
					})
				}
			},
			expectedStatus: fiber.StatusCreated,
			validate: func(t *testing.T, resp *http.Response, app *fiber.App, req *http.Request) {
				assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

				var body map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, "created", body["message"])
			},
		},
		{
			name: "should_log_without_request_id_in_context",
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
			},
		},
		{
			name: "should_calculate_latency_correctly",
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req
			},
			setupHandler: func(t *testing.T) fiber.Handler {
				return func(c *fiber.Ctx) error {

					return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "done"})
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
			app := fiber.New(fiber.Config{
				ErrorHandler: func(c *fiber.Ctx, err error) error {
					code := fiber.StatusInternalServerError
					if e, ok := err.(*fiber.Error); ok {
						code = e.Code
					}
					return c.Status(code).JSON(fiber.Map{"error": err.Error()})
				},
			})

			middleware := NewLogMiddleware()
			app.Use(middleware.Log)

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
