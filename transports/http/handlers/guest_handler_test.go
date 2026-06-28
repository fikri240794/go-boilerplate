package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go-boilerplate/internal/models/dtos"
	"go-boilerplate/internal/services/mocks"
	"go-boilerplate/pkg/constants"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fikri240794/gocerr"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewGuestHandler(t *testing.T) {
	tests := []struct {
		name         string
		setupService func(t *testing.T) *mocks.GuestServiceMock
		validate     func(t *testing.T, handler *GuestHandler, mockService *mocks.GuestServiceMock)
	}{
		{
			name: "should create guest handler with guest service",
			setupService: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, handler *GuestHandler, mockService *mocks.GuestServiceMock) {
				assert.NotNil(t, handler, "Expected handler to be non-nil")
				assert.Equal(t, mockService, handler.guestService, "Expected guestService to be set correctly")
			},
		},
		{
			name: "should create guest handler with nil service",
			setupService: func(t *testing.T) *mocks.GuestServiceMock {
				return nil
			},
			validate: func(t *testing.T, handler *GuestHandler, mockService *mocks.GuestServiceMock) {
				assert.NotNil(t, handler, "Expected handler to be non-nil")
				assert.Nil(t, handler.guestService, "Expected guestService to be nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := tt.setupService(t)

			handler := NewGuestHandler(mockService)

			if tt.validate != nil {
				tt.validate(t, handler, mockService)
			}
		})
	}
}

func TestGuestHandler_SetupRoutes(t *testing.T) {
	tests := []struct {
		name         string
		setupHandler func(t *testing.T) *GuestHandler
		validate     func(t *testing.T, app *fiber.App)
	}{
		{
			name: "should setup all guest routes correctly",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				return NewGuestHandler(mockService)
			},
			validate: func(t *testing.T, app *fiber.App) {
				routes := app.GetRoutes()

				routeMap := make(map[string]bool)
				for _, route := range routes {
					routeMap[route.Method+" "+route.Path] = true
				}

				hasPostGuests := false
				hasDeleteGuestsWithID := false
				hasGetGuests := false
				hasGetGuestsWithID := false
				hasPutGuestsWithID := false

				for key := range routeMap {
					if key == "POST /guests" || key == "POST /guests/" {
						hasPostGuests = true
					}
					if key == "DELETE /guests/:id" || key == "DELETE /guests/:id/" {
						hasDeleteGuestsWithID = true
					}
					if key == "GET /guests" || key == "GET /guests/" {
						hasGetGuests = true
					}
					if key == "GET /guests/:id" || key == "GET /guests/:id/" {
						hasGetGuestsWithID = true
					}
					if key == "PUT /guests/:id" || key == "PUT /guests/:id/" {
						hasPutGuestsWithID = true
					}
				}

				assert.True(t, hasPostGuests, "Expected POST /guests route to be registered")
				assert.True(t, hasDeleteGuestsWithID, "Expected DELETE /guests/:id route to be registered")
				assert.True(t, hasGetGuests, "Expected GET /guests route to be registered")
				assert.True(t, hasGetGuestsWithID, "Expected GET /guests/:id route to be registered")
				assert.True(t, hasPutGuestsWithID, "Expected PUT /guests/:id route to be registered")

				hasPostGuestsBulk := false
				hasPutGuestsBulk := false
				hasDeleteGuestsBulk := false

				for key := range routeMap {
					if key == "POST /guests/bulk" || key == "POST /guests/bulk/" {
						hasPostGuestsBulk = true
					}
					if key == "PUT /guests/bulk" || key == "PUT /guests/bulk/" {
						hasPutGuestsBulk = true
					}
					if key == "DELETE /guests/bulk" || key == "DELETE /guests/bulk/" {
						hasDeleteGuestsBulk = true
					}
				}

				assert.True(t, hasPostGuestsBulk, "Expected POST /guests/bulk route to be registered")
				assert.True(t, hasPutGuestsBulk, "Expected PUT /guests/bulk route to be registered")
				assert.True(t, hasDeleteGuestsBulk, "Expected DELETE /guests/bulk route to be registered")
			},
		},
		{
			name: "should setup routes with nil service handler",
			setupHandler: func(t *testing.T) *GuestHandler {
				return NewGuestHandler(nil)
			},
			validate: func(t *testing.T, app *fiber.App) {
				routes := app.GetRoutes()

				var guestRoutes []fiber.Route
				for _, route := range routes {
					if len(route.Path) >= 7 && route.Path[:7] == "/guests" {
						guestRoutes = append(guestRoutes, route)
					}
				}

				assert.GreaterOrEqual(t, len(guestRoutes), 8, "Expected at least 8 guest routes to be registered")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.setupHandler(t)
			app := fiber.New()

			handler.SetupRoutes(app)

			if tt.validate != nil {
				tt.validate(t, app)
			}
		})
	}
}

func TestGuestHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		setupHandler   func(t *testing.T) *GuestHandler
		setupRequest   func(t *testing.T) *http.Request
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock)
	}{
		{
			name: "should create guest successfully",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("Create", mock.Anything, mock.AnythingOfType("*dtos.CreateGuestRequestDTO")).
					Return(&dtos.GuestResponseDTO{
						ID:        "01932293-d710-7f55-a9f6-66e6248ae72f",
						Name:      "John Snow",
						Address:   "123 Main Street",
						CreatedAt: 1731452061534,
						CreatedBy: "00000000-0000-0000-0000-000000000000",
					}, nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := map[string]interface{}{
					"name":    "John Snow",
					"address": "123 Main Street",
				}
				bodyBytes, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodPost, "/guests", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")

				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusCreated,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusCreated), response["code"])
				assert.NotNil(t, response["data"])

				data := response["data"].(map[string]interface{})
				assert.Equal(t, "01932293-d710-7f55-a9f6-66e6248ae72f", data["id"])
				assert.Equal(t, "John Snow", data["name"])
				assert.Equal(t, "123 Main Street", data["address"])
			},
		},
		{
			name: "should return error when body parser fails",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {

				req := httptest.NewRequest(http.MethodPost, "/guests", bytes.NewReader([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")

				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusBadRequest), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should return error when service returns bad request error",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)

				mockService.On("Create", mock.Anything, mock.AnythingOfType("*dtos.CreateGuestRequestDTO")).
					Return((*dtos.GuestResponseDTO)(nil), errors.New("validation error"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := map[string]interface{}{
					"name":    "John Snow",
					"address": "123 Main Street",
				}
				bodyBytes, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodPost, "/guests", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")

				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should return internal server error when service fails with error code >= 500",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)

				mockService.On("Create", mock.Anything, mock.AnythingOfType("*dtos.CreateGuestRequestDTO")).
					Return((*dtos.GuestResponseDTO)(nil), gocerr.New(fiber.StatusInternalServerError, "internal server error"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := map[string]interface{}{
					"name":    "John Snow",
					"address": "123 Main Street",
				}
				bodyBytes, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodPost, "/guests", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")

				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusInternalServerError), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should return error when service fails with non-500 error",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("Create", mock.Anything, mock.AnythingOfType("*dtos.CreateGuestRequestDTO")).
					Return((*dtos.GuestResponseDTO)(nil), errors.New("generic error"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := map[string]interface{}{
					"name":    "John Snow",
					"address": "123 Main Street",
				}
				bodyBytes, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodPost, "/guests", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")

				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should handle empty request body",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("Create", mock.Anything, mock.AnythingOfType("*dtos.CreateGuestRequestDTO")).
					Return(&dtos.GuestResponseDTO{
						ID:        "01932293-d710-7f55-a9f6-66e6248ae72f",
						Name:      "",
						Address:   "",
						CreatedAt: 1731452061534,
						CreatedBy: "00000000-0000-0000-0000-000000000000",
					}, nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := map[string]interface{}{}
				bodyBytes, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodPost, "/guests", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")

				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusCreated,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusCreated), response["code"])
				assert.NotNil(t, response["data"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.setupHandler(t)
			app := fiber.New()

			app.Post("/guests", handler.Create)

			req := tt.setupRequest(t)
			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.validate != nil {
				var mockService *mocks.GuestServiceMock
				if handler.guestService != nil {
					mockService = handler.guestService.(*mocks.GuestServiceMock)
				}
				tt.validate(t, resp, mockService)
			}
		})
	}
}

func TestGuestHandler_DeleteByID(t *testing.T) {
	tests := []struct {
		name           string
		setupHandler   func(t *testing.T) *GuestHandler
		setupRequest   func(t *testing.T) *http.Request
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock)
	}{
		{
			name: "should_delete_guest_successfully",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("DeleteByID", mock.Anything, mock.AnythingOfType("*dtos.DeleteGuestByIDRequestDTO")).
					Return(nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusOK), response["code"])
				assert.Equal(t, true, response["data"])

				mockService.AssertCalled(t, "DeleteByID", mock.Anything, mock.AnythingOfType("*dtos.DeleteGuestByIDRequestDTO"))
			},
		},
		{
			name: "should_handle_request_without_context_values",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("DeleteByID", mock.Anything, mock.AnythingOfType("*dtos.DeleteGuestByIDRequestDTO")).
					Return(nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", nil)
				return req
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusOK), response["code"])
				assert.Equal(t, true, response["data"])
			},
		},
		{
			name: "should_return_error_when_service_fails_with_error_code_<_500",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("DeleteByID", mock.Anything, mock.AnythingOfType("*dtos.DeleteGuestByIDRequestDTO")).
					Return(gocerr.New(fiber.StatusNotFound, "guest not found"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusNotFound,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusNotFound), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should_return_error_when_service_fails_with_error_code_>=_500",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("DeleteByID", mock.Anything, mock.AnythingOfType("*dtos.DeleteGuestByIDRequestDTO")).
					Return(gocerr.New(fiber.StatusInternalServerError, "internal server error"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusInternalServerError), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should_return_error_when_service_fails_with_generic_error",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("DeleteByID", mock.Anything, mock.AnythingOfType("*dtos.DeleteGuestByIDRequestDTO")).
					Return(errors.New("database connection failed"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should_delete_guest_with_valid_UUID",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("DeleteByID", mock.Anything, mock.AnythingOfType("*dtos.DeleteGuestByIDRequestDTO")).
					Return(nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)

				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusOK), response["code"])
				assert.Equal(t, true, response["data"])

				mockService.AssertCalled(t, "DeleteByID", mock.Anything, mock.AnythingOfType("*dtos.DeleteGuestByIDRequestDTO"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.setupHandler(t)
			app := fiber.New()

			app.Delete("/guests/:id", handler.DeleteByID)

			req := tt.setupRequest(t)
			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.validate != nil {
				var mockService *mocks.GuestServiceMock
				if handler.guestService != nil {
					mockService = handler.guestService.(*mocks.GuestServiceMock)
				}
				tt.validate(t, resp, mockService)
			}
		})
	}
}

func TestGuestHandler_FindAll(t *testing.T) {
	tests := []struct {
		name           string
		setupHandler   func(t *testing.T) *GuestHandler
		setupRequest   func(t *testing.T) *http.Request
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock)
	}{
		{
			name: "should_return_guests_successfully",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				responseDTO := &dtos.FindAllGuestResponseDTO{
					List: []dtos.GuestResponseDTO{
						{
							ID:      "01932293-d710-7f55-a9f6-66e6248ae72f",
							Name:    "John Snow",
							Address: "123 Main Street",
						},
					},
					Count: 1,
				}
				mockService.On("FindAll", mock.Anything, mock.AnythingOfType("*dtos.FindAllGuestRequestDTO")).
					Return(responseDTO, nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/guests?keyword=John&take=10&skip=0", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusOK), response["code"])
				assert.NotNil(t, response["data"])

				mockService.AssertCalled(t, "FindAll", mock.Anything, mock.AnythingOfType("*dtos.FindAllGuestRequestDTO"))
			},
		},
		{
			name: "should_return_empty_list_when_no_guests_found",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				responseDTO := &dtos.FindAllGuestResponseDTO{
					List:  []dtos.GuestResponseDTO{},
					Count: 0,
				}
				mockService.On("FindAll", mock.Anything, mock.AnythingOfType("*dtos.FindAllGuestRequestDTO")).
					Return(responseDTO, nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/guests?take=10", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusOK), response["code"])
			},
		},
		{
			name: "should_handle_request_without_context_values",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				responseDTO := &dtos.FindAllGuestResponseDTO{
					List:  []dtos.GuestResponseDTO{},
					Count: 0,
				}
				mockService.On("FindAll", mock.Anything, mock.AnythingOfType("*dtos.FindAllGuestRequestDTO")).
					Return(responseDTO, nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/guests?take=10", nil)
				return req
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "should_return_error_when_query_parser_fails",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {

				req := httptest.NewRequest(http.MethodGet, "/guests?take=invalid", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusBadRequest), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should_return_error_when_service_fails_with_error_code_<_500",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("FindAll", mock.Anything, mock.AnythingOfType("*dtos.FindAllGuestRequestDTO")).
					Return(nil, gocerr.New(fiber.StatusBadRequest, "validation error"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/guests?take=10", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusBadRequest), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should_return_error_when_service_fails_with_error_code_>=_500",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("FindAll", mock.Anything, mock.AnythingOfType("*dtos.FindAllGuestRequestDTO")).
					Return(nil, gocerr.New(fiber.StatusInternalServerError, "internal server error"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/guests?take=10", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusInternalServerError), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should_return_error_when_service_fails_with_generic_error",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("FindAll", mock.Anything, mock.AnythingOfType("*dtos.FindAllGuestRequestDTO")).
					Return(nil, errors.New("database connection failed"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/guests?take=10", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should_handle_pagination_parameters",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				responseDTO := &dtos.FindAllGuestResponseDTO{
					List: []dtos.GuestResponseDTO{
						{
							ID:      "01932293-d710-7f55-a9f6-66e6248ae72f",
							Name:    "John Snow",
							Address: "123 Main Street",
						},
					},
					Count: 100,
				}
				mockService.On("FindAll", mock.Anything, mock.AnythingOfType("*dtos.FindAllGuestRequestDTO")).
					Return(responseDTO, nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/guests?take=10&skip=20&sorts=name.asc", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusOK), response["code"])
				assert.NotNil(t, response["data"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.setupHandler(t)
			app := fiber.New()

			app.Get("/guests", handler.FindAll)

			req := tt.setupRequest(t)
			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.validate != nil {
				var mockService *mocks.GuestServiceMock
				if handler.guestService != nil {
					mockService = handler.guestService.(*mocks.GuestServiceMock)
				}
				tt.validate(t, resp, mockService)
			}
		})
	}
}

func TestGuestHandler_FindByID(t *testing.T) {
	tests := []struct {
		name           string
		setupHandler   func(t *testing.T) *GuestHandler
		setupRequest   func(t *testing.T) *http.Request
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock)
	}{
		{
			name: "should_return_guest_successfully",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				responseDTO := &dtos.GuestResponseDTO{
					ID:      "01932293-d710-7f55-a9f6-66e6248ae72f",
					Name:    "John Snow",
					Address: "123 Main Street",
				}
				mockService.On("FindByID", mock.Anything, mock.AnythingOfType("*dtos.FindGuestByIDRequestDTO")).
					Return(responseDTO, nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusOK), response["code"])
				assert.NotNil(t, response["data"])

				mockService.AssertCalled(t, "FindByID", mock.Anything, mock.AnythingOfType("*dtos.FindGuestByIDRequestDTO"))
			},
		},
		{
			name: "should_handle_request_without_context_values",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				responseDTO := &dtos.GuestResponseDTO{
					ID:      "01932293-d710-7f55-a9f6-66e6248ae72f",
					Name:    "John Snow",
					Address: "123 Main Street",
				}
				mockService.On("FindByID", mock.Anything, mock.AnythingOfType("*dtos.FindGuestByIDRequestDTO")).
					Return(responseDTO, nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", nil)
				return req
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "should_return_error_when_service_fails_with_error_code_<_500",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("FindByID", mock.Anything, mock.AnythingOfType("*dtos.FindGuestByIDRequestDTO")).
					Return(nil, gocerr.New(fiber.StatusNotFound, "guest not found"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusNotFound,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusNotFound), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should_return_error_when_service_fails_with_error_code_>=_500",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("FindByID", mock.Anything, mock.AnythingOfType("*dtos.FindGuestByIDRequestDTO")).
					Return(nil, gocerr.New(fiber.StatusInternalServerError, "internal server error"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusInternalServerError), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should_return_error_when_service_fails_with_generic_error",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("FindByID", mock.Anything, mock.AnythingOfType("*dtos.FindGuestByIDRequestDTO")).
					Return(nil, errors.New("database connection failed"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should_return_error_when_invalid_id_format",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("FindByID", mock.Anything, mock.AnythingOfType("*dtos.FindGuestByIDRequestDTO")).
					Return(nil, gocerr.New(fiber.StatusBadRequest, "invalid id format"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/guests/invalid-uuid", nil)
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusBadRequest), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.setupHandler(t)
			app := fiber.New()

			app.Get("/guests/:id", handler.FindByID)

			req := tt.setupRequest(t)
			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.validate != nil {
				var mockService *mocks.GuestServiceMock
				if handler.guestService != nil {
					mockService = handler.guestService.(*mocks.GuestServiceMock)
				}
				tt.validate(t, resp, mockService)
			}
		})
	}
}

func TestGuestHandler_UpdateByID(t *testing.T) {
	tests := []struct {
		name           string
		setupHandler   func(t *testing.T) *GuestHandler
		setupRequest   func(t *testing.T) *http.Request
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock)
	}{
		{
			name: "should_update_guest_successfully",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				responseDTO := &dtos.GuestResponseDTO{
					ID:      "01932293-d710-7f55-a9f6-66e6248ae72f",
					Name:    "John Snow Updated",
					Address: "456 New Street",
				}
				mockService.On("UpdateByID", mock.Anything, mock.AnythingOfType("*dtos.UpdateGuestByIDRequestDTO")).
					Return(responseDTO, nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := strings.NewReader(`{"name":"John Snow Updated","address":"456 New Street"}`)
				req := httptest.NewRequest(http.MethodPut, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", body)
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusOK), response["code"])
				assert.NotNil(t, response["data"])

				mockService.AssertCalled(t, "UpdateByID", mock.Anything, mock.AnythingOfType("*dtos.UpdateGuestByIDRequestDTO"))
			},
		},
		{
			name: "should_handle_request_without_context_values",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				responseDTO := &dtos.GuestResponseDTO{
					ID:      "01932293-d710-7f55-a9f6-66e6248ae72f",
					Name:    "John Snow",
					Address: "123 Main Street",
				}
				mockService.On("UpdateByID", mock.Anything, mock.AnythingOfType("*dtos.UpdateGuestByIDRequestDTO")).
					Return(responseDTO, nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := strings.NewReader(`{"name":"John Snow","address":"123 Main Street"}`)
				req := httptest.NewRequest(http.MethodPut, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", body)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "should_return_error_when_body_parser_fails",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := strings.NewReader(`invalid json`)
				req := httptest.NewRequest(http.MethodPut, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", body)
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusBadRequest), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should_return_error_when_service_fails_with_error_code_<_500",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("UpdateByID", mock.Anything, mock.AnythingOfType("*dtos.UpdateGuestByIDRequestDTO")).
					Return(nil, gocerr.New(fiber.StatusNotFound, "guest not found"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := strings.NewReader(`{"name":"John Snow","address":"123 Main Street"}`)
				req := httptest.NewRequest(http.MethodPut, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", body)
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusNotFound,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusNotFound), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should_return_error_when_service_fails_with_error_code_>=_500",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("UpdateByID", mock.Anything, mock.AnythingOfType("*dtos.UpdateGuestByIDRequestDTO")).
					Return(nil, gocerr.New(fiber.StatusInternalServerError, "internal server error"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := strings.NewReader(`{"name":"John Snow","address":"123 Main Street"}`)
				req := httptest.NewRequest(http.MethodPut, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", body)
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusInternalServerError), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should_return_error_when_service_fails_with_generic_error",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("UpdateByID", mock.Anything, mock.AnythingOfType("*dtos.UpdateGuestByIDRequestDTO")).
					Return(nil, errors.New("database connection failed"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := strings.NewReader(`{"name":"John Snow","address":"123 Main Street"}`)
				req := httptest.NewRequest(http.MethodPut, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", body)
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should_return_error_when_validation_fails",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("UpdateByID", mock.Anything, mock.AnythingOfType("*dtos.UpdateGuestByIDRequestDTO")).
					Return(nil, gocerr.New(fiber.StatusBadRequest, "validation error"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := strings.NewReader(`{"name":"","address":"123 Main Street"}`)
				req := httptest.NewRequest(http.MethodPut, "/guests/01932293-d710-7f55-a9f6-66e6248ae72f", body)
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				var response map[string]interface{}
				err = json.Unmarshal(bodyBytes, &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(fiber.StatusBadRequest), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.setupHandler(t)
			app := fiber.New()

			app.Put("/guests/:id", handler.UpdateByID)

			req := tt.setupRequest(t)
			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.validate != nil {
				var mockService *mocks.GuestServiceMock
				if handler.guestService != nil {
					mockService = handler.guestService.(*mocks.GuestServiceMock)
				}
				tt.validate(t, resp, mockService)
			}
		})
	}
}

func TestGuestHandler_BulkCreate(t *testing.T) {
	tests := []struct {
		name           string
		setupHandler   func(t *testing.T) *GuestHandler
		setupRequest   func(t *testing.T) *http.Request
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock)
	}{
		{
			name: "should bulk create guests successfully",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("BulkCreate", mock.Anything, mock.AnythingOfType("*dtos.BulkCreateGuestsRequestDTO")).
					Return(&dtos.BulkCreateGuestsResponseDTO{
						Guests: []dtos.GuestResponseDTO{
							{ID: "01932293-d710-7f55-a9f6-66e6248ae72f", Name: "John Snow", Address: "123 Main St", CreatedAt: 1731452061534, CreatedBy: "00000000-0000-0000-0000-000000000000"},
						},
					}, nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := map[string]interface{}{
					"items": []map[string]interface{}{
						{"name": "John Snow", "address": "123 Main St"},
					},
				}
				bodyBytes, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodPost, "/guests/bulk", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusCreated,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
				bodyBytes, _ := io.ReadAll(resp.Body)
				var response map[string]interface{}
				json.Unmarshal(bodyBytes, &response)
				assert.Equal(t, float64(fiber.StatusCreated), response["code"])
				assert.NotNil(t, response["data"])
			},
		},
		{
			name: "should return error when body parser fails",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/guests/bulk", bytes.NewReader([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
				bodyBytes, _ := io.ReadAll(resp.Body)
				var response map[string]interface{}
				json.Unmarshal(bodyBytes, &response)
				assert.Equal(t, float64(fiber.StatusBadRequest), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should return error when service returns bad request error",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("BulkCreate", mock.Anything, mock.AnythingOfType("*dtos.BulkCreateGuestsRequestDTO")).
					Return((*dtos.BulkCreateGuestsResponseDTO)(nil), errors.New("validation error"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := map[string]interface{}{
					"items": []map[string]interface{}{
						{"name": "John Snow"},
					},
				}
				bodyBytes, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodPost, "/guests/bulk", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, _ := io.ReadAll(resp.Body)
				var response map[string]interface{}
				json.Unmarshal(bodyBytes, &response)
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should return internal server error when service fails with error code >= 500",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("BulkCreate", mock.Anything, mock.AnythingOfType("*dtos.BulkCreateGuestsRequestDTO")).
					Return((*dtos.BulkCreateGuestsResponseDTO)(nil), gocerr.New(fiber.StatusInternalServerError, "internal error"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := map[string]interface{}{
					"items": []map[string]interface{}{
						{"name": "John Snow"},
					},
				}
				bodyBytes, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodPost, "/guests/bulk", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
				bodyBytes, _ := io.ReadAll(resp.Body)
				var response map[string]interface{}
				json.Unmarshal(bodyBytes, &response)
				assert.Equal(t, float64(fiber.StatusInternalServerError), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.setupHandler(t)
			app := fiber.New()

			app.Post("/guests/bulk", handler.BulkCreate)

			req := tt.setupRequest(t)
			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.validate != nil {
				var mockService *mocks.GuestServiceMock
				if handler.guestService != nil {
					mockService = handler.guestService.(*mocks.GuestServiceMock)
				}
				tt.validate(t, resp, mockService)
			}
		})
	}
}

func TestGuestHandler_BulkUpdate(t *testing.T) {
	tests := []struct {
		name           string
		setupHandler   func(t *testing.T) *GuestHandler
		setupRequest   func(t *testing.T) *http.Request
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock)
	}{
		{
			name: "should bulk update guests successfully",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("BulkUpdate", mock.Anything, mock.AnythingOfType("*dtos.BulkUpdateGuestsRequestDTO")).
					Return(&dtos.BulkUpdateGuestsResponseDTO{
						Guests: []dtos.GuestResponseDTO{
							{ID: "01932293-d710-7f55-a9f6-66e6248ae72f", Name: "Updated Name", Address: "456 Oak Ave", CreatedAt: 1731452061534, CreatedBy: "admin", UpdatedAt: 1731452061535, UpdatedBy: "admin"},
						},
					}, nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := map[string]interface{}{
					"items": []map[string]interface{}{
						{"id": "01932293-d710-7f55-a9f6-66e6248ae72f", "name": "Updated Name", "address": "456 Oak Ave", "updated_by": "admin"},
					},
				}
				bodyBytes, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodPut, "/guests/bulk", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
				bodyBytes, _ := io.ReadAll(resp.Body)
				var response map[string]interface{}
				json.Unmarshal(bodyBytes, &response)
				assert.Equal(t, float64(fiber.StatusOK), response["code"])
				assert.NotNil(t, response["data"])
			},
		},
		{
			name: "should return error when body parser fails",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodPut, "/guests/bulk", bytes.NewReader([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
				bodyBytes, _ := io.ReadAll(resp.Body)
				var response map[string]interface{}
				json.Unmarshal(bodyBytes, &response)
				assert.Equal(t, float64(fiber.StatusBadRequest), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should return error when service returns error",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("BulkUpdate", mock.Anything, mock.AnythingOfType("*dtos.BulkUpdateGuestsRequestDTO")).
					Return((*dtos.BulkUpdateGuestsResponseDTO)(nil), errors.New("bulk update error"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := map[string]interface{}{
					"items": []map[string]interface{}{
						{"id": "01932293-d710-7f55-a9f6-66e6248ae72f", "name": "Updated", "updated_by": "admin"},
					},
				}
				bodyBytes, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodPut, "/guests/bulk", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, _ := io.ReadAll(resp.Body)
				var response map[string]interface{}
				json.Unmarshal(bodyBytes, &response)
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should return error when service fails with error code >= 500",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("BulkUpdate", mock.Anything, mock.AnythingOfType("*dtos.BulkUpdateGuestsRequestDTO")).
					Return((*dtos.BulkUpdateGuestsResponseDTO)(nil), gocerr.New(fiber.StatusInternalServerError, "internal error"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := map[string]interface{}{
					"items": []map[string]interface{}{
						{"id": "01932293-d710-7f55-a9f6-66e6248ae72f", "name": "Updated", "updated_by": "admin"},
					},
				}
				bodyBytes, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodPut, "/guests/bulk", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
				bodyBytes, _ := io.ReadAll(resp.Body)
				var response map[string]interface{}
				json.Unmarshal(bodyBytes, &response)
				assert.Equal(t, float64(fiber.StatusInternalServerError), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.setupHandler(t)
			app := fiber.New()

			app.Put("/guests/bulk", handler.BulkUpdate)

			req := tt.setupRequest(t)
			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.validate != nil {
				var mockService *mocks.GuestServiceMock
				if handler.guestService != nil {
					mockService = handler.guestService.(*mocks.GuestServiceMock)
				}
				tt.validate(t, resp, mockService)
			}
		})
	}
}

func TestGuestHandler_BulkDelete(t *testing.T) {
	tests := []struct {
		name           string
		setupHandler   func(t *testing.T) *GuestHandler
		setupRequest   func(t *testing.T) *http.Request
		expectedStatus int
		validate       func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock)
	}{
		{
			name: "should bulk delete guests successfully",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("BulkDelete", mock.Anything, mock.AnythingOfType("*dtos.BulkDeleteGuestsRequestDTO")).
					Return(nil)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := map[string]interface{}{
					"ids":        []string{"01932293-d710-7f55-a9f6-66e6248ae72f"},
					"deleted_by": "admin",
				}
				bodyBytes, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodDelete, "/guests/bulk", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusOK,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
				bodyBytes, _ := io.ReadAll(resp.Body)
				var response map[string]interface{}
				json.Unmarshal(bodyBytes, &response)
				assert.Equal(t, float64(fiber.StatusOK), response["code"])
				assert.True(t, response["data"].(bool))
			},
		},
		{
			name: "should return error when body parser fails",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/guests/bulk", bytes.NewReader([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusBadRequest,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
				bodyBytes, _ := io.ReadAll(resp.Body)
				var response map[string]interface{}
				json.Unmarshal(bodyBytes, &response)
				assert.Equal(t, float64(fiber.StatusBadRequest), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should return error when service returns error",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("BulkDelete", mock.Anything, mock.AnythingOfType("*dtos.BulkDeleteGuestsRequestDTO")).
					Return(errors.New("bulk delete error"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := map[string]interface{}{
					"ids":        []string{"01932293-d710-7f55-a9f6-66e6248ae72f"},
					"deleted_by": "admin",
				}
				bodyBytes, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodDelete, "/guests/bulk", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				bodyBytes, _ := io.ReadAll(resp.Body)
				var response map[string]interface{}
				json.Unmarshal(bodyBytes, &response)
				assert.NotNil(t, response["error"])
			},
		},
		{
			name: "should return error when service fails with gocerr",
			setupHandler: func(t *testing.T) *GuestHandler {
				mockService := mocks.NewGuestServiceMock(t)
				mockService.On("BulkDelete", mock.Anything, mock.AnythingOfType("*dtos.BulkDeleteGuestsRequestDTO")).
					Return(gocerr.New(fiber.StatusInternalServerError, "internal error"))
				return NewGuestHandler(mockService)
			},
			setupRequest: func(t *testing.T) *http.Request {
				body := map[string]interface{}{
					"ids":        []string{"01932293-d710-7f55-a9f6-66e6248ae72f"},
					"deleted_by": "admin",
				}
				bodyBytes, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodDelete, "/guests/bulk", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				ctx := context.WithValue(req.Context(), constants.ContextKeyRequestID, "test-request-id")
				return req.WithContext(ctx)
			},
			expectedStatus: fiber.StatusInternalServerError,
			validate: func(t *testing.T, resp *http.Response, mockService *mocks.GuestServiceMock) {
				assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
				bodyBytes, _ := io.ReadAll(resp.Body)
				var response map[string]interface{}
				json.Unmarshal(bodyBytes, &response)
				assert.Equal(t, float64(fiber.StatusInternalServerError), response["code"])
				assert.NotNil(t, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.setupHandler(t)
			app := fiber.New()

			app.Delete("/guests/bulk", handler.BulkDelete)

			req := tt.setupRequest(t)
			resp, err := app.Test(req, -1)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.validate != nil {
				var mockService *mocks.GuestServiceMock
				if handler.guestService != nil {
					mockService = handler.guestService.(*mocks.GuestServiceMock)
				}
				tt.validate(t, resp, mockService)
			}
		})
	}
}
