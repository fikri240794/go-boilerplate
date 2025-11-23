package handlers

import (
	"go-boilerplate/internal/services/mocks"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestHandlers_SetupRoutes(t *testing.T) {
	tests := []struct {
		name         string
		setupHandler func(t *testing.T) *Handlers
		validate     func(t *testing.T, app *fiber.App)
	}{
		{
			name: "should_setup_routes_successfully",
			setupHandler: func(t *testing.T) *Handlers {
				mockService := mocks.NewGuestServiceMock(t)
				guestHandler := NewGuestHandler(mockService)
				return &Handlers{
					Guest: guestHandler,
				}
			},
			validate: func(t *testing.T, app *fiber.App) {

				routes := app.GetRoutes()
				assert.NotEmpty(t, routes)

				var foundDeleteGuest bool
				var foundGetGuestByID bool
				var foundPutGuest bool

				for _, route := range routes {
					if route.Method == "DELETE" && route.Path == "/guests/:id" {
						foundDeleteGuest = true
					}
					if route.Method == "GET" && route.Path == "/guests/:id" {
						foundGetGuestByID = true
					}
					if route.Method == "PUT" && route.Path == "/guests/:id" {
						foundPutGuest = true
					}
				}

				assert.True(t, foundDeleteGuest, "DELETE /guests/:id route should be registered")
				assert.True(t, foundGetGuestByID, "GET /guests/:id route should be registered")
				assert.True(t, foundPutGuest, "PUT /guests/:id route should be registered")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers := tt.setupHandler(t)
			app := fiber.New()

			handlers.SetupRoutes(app)

			if tt.validate != nil {
				tt.validate(t, app)
			}
		})
	}
}
