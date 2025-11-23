package handlers

import (
	service_mocks "go-boilerplate/internal/services/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewImplementedBoilerplateServer(t *testing.T) {
	tests := []struct {
		name     string
		validate func(t *testing.T, server *ImplementedBoilerplateServer)
	}{
		{
			name: "should_create_new_implemented_boilerplate_server_successfully",
			validate: func(t *testing.T, server *ImplementedBoilerplateServer) {
				assert.NotNil(t, server)
				assert.NotNil(t, server.guestService)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := service_mocks.NewGuestServiceMock(t)

			server := NewImplementedBoilerplateServer(mockService)

			if tt.validate != nil {
				tt.validate(t, server)
			}
		})
	}
}
