package dtos

import (
	"go-boilerplate/internal/models/entities"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGuestEventRequestDTO_ToEntity(t *testing.T) {
	tests := []struct {
		name     string
		dto      *GuestEventRequestDTO
		validate func(t *testing.T, result *entities.GuestEventEntity, original *GuestEventRequestDTO)
	}{
		{
			name: "convert DTO to entity with all fields populated",
			dto: &GuestEventRequestDTO{
				ID:        "550e8400-e29b-41d4-a716-446655440000",
				Name:      "John Doe",
				Address:   "123 Main St",
				CreatedAt: 1234567890,
				CreatedBy: "admin",
				UpdatedAt: 1234567999,
				UpdatedBy: "editor",
				DeletedAt: 0,
				DeletedBy: "",
			},
			validate: func(t *testing.T, result *entities.GuestEventEntity, original *GuestEventRequestDTO) {
				assert.Equal(t, original.ID, result.ID)

				assert.Equal(t, original.Name, result.Name)

				assert.Equal(t, original.Address, result.Address)

				assert.Equal(t, original.CreatedAt, result.CreatedAt)

				assert.Equal(t, original.CreatedBy, result.CreatedBy)

				assert.Equal(t, original.UpdatedAt, result.UpdatedAt)

				assert.Equal(t, original.UpdatedBy, result.UpdatedBy)

				assert.Equal(t, original.DeletedAt, result.DeletedAt)

				assert.Equal(t, original.DeletedBy, result.DeletedBy)
			},
		},
		{
			name: "convert DTO to entity with empty optional fields",
			dto: &GuestEventRequestDTO{
				ID:        "550e8400-e29b-41d4-a716-446655440001",
				Name:      "Jane Smith",
				Address:   "",
				CreatedAt: 9876543210,
				CreatedBy: "system",
				UpdatedAt: 0,
				UpdatedBy: "",
				DeletedAt: 0,
				DeletedBy: "",
			},
			validate: func(t *testing.T, result *entities.GuestEventEntity, original *GuestEventRequestDTO) {
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", result.ID)

				assert.Equal(t, "Jane Smith", result.Name)

				assert.Equal(t, "", result.Address)

				assert.Equal(t, int64(0), result.UpdatedAt)

				assert.Equal(t, "", result.UpdatedBy)
			},
		},
		{
			name: "convert DTO to entity with deleted fields populated",
			dto: &GuestEventRequestDTO{
				ID:        "550e8400-e29b-41d4-a716-446655440002",
				Name:      "Bob Johnson",
				Address:   "456 Oak Ave",
				CreatedAt: 1111111111,
				CreatedBy: "admin",
				UpdatedAt: 2222222222,
				UpdatedBy: "editor",
				DeletedAt: 3333333333,
				DeletedBy: "admin",
			},
			validate: func(t *testing.T, result *entities.GuestEventEntity, original *GuestEventRequestDTO) {
				assert.Equal(t, int64(3333333333), result.DeletedAt)

				assert.Equal(t, "admin", result.DeletedBy)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dto.ToEntity()

			assert.NotNil(t, result, "Result should not be nil")

			tt.validate(t, result, tt.dto)
		})
	}
}

func TestNewGuestEventResponseDTO(t *testing.T) {
	tests := []struct {
		name     string
		entity   *entities.GuestEventEntity
		validate func(t *testing.T, result *GuestEventResponseDTO, original *entities.GuestEventEntity)
	}{
		{
			name: "create response DTO from entity with all fields",
			entity: &entities.GuestEventEntity{
				ID:        "550e8400-e29b-41d4-a716-446655440000",
				Name:      "John Doe",
				Address:   "123 Main St",
				CreatedAt: 1234567890,
				CreatedBy: "admin",
				UpdatedAt: 1234567999,
				UpdatedBy: "editor",
				DeletedAt: 0,
				DeletedBy: "",
			},
			validate: func(t *testing.T, result *GuestEventResponseDTO, original *entities.GuestEventEntity) {
				assert.Equal(t, original.ID, result.ID)

				assert.Equal(t, original.Name, result.Name)

				assert.Equal(t, original.Address, result.Address)

				assert.Equal(t, original.CreatedAt, result.CreatedAt)

				assert.Equal(t, original.CreatedBy, result.CreatedBy)

				assert.Equal(t, original.UpdatedAt, result.UpdatedAt)

				assert.Equal(t, original.UpdatedBy, result.UpdatedBy)

				assert.Equal(t, original.DeletedAt, result.DeletedAt)

				assert.Equal(t, original.DeletedBy, result.DeletedBy)
			},
		},
		{
			name: "create response DTO from entity with minimal fields",
			entity: &entities.GuestEventEntity{
				ID:        "550e8400-e29b-41d4-a716-446655440001",
				Name:      "Jane Smith",
				Address:   "",
				CreatedAt: 9876543210,
				CreatedBy: "system",
				UpdatedAt: 0,
				UpdatedBy: "",
				DeletedAt: 0,
				DeletedBy: "",
			},
			validate: func(t *testing.T, result *GuestEventResponseDTO, original *entities.GuestEventEntity) {
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", result.ID)

				assert.Equal(t, "Jane Smith", result.Name)

				assert.Equal(t, "", result.Address)

				assert.Equal(t, int64(0), result.UpdatedAt)

				assert.Equal(t, "", result.UpdatedBy)

				assert.Equal(t, int64(0), result.DeletedAt)

				assert.Equal(t, "", result.DeletedBy)
			},
		},
		{
			name: "create response DTO from deleted entity",
			entity: &entities.GuestEventEntity{
				ID:        "550e8400-e29b-41d4-a716-446655440002",
				Name:      "Bob Johnson",
				Address:   "456 Oak Ave",
				CreatedAt: 1111111111,
				CreatedBy: "admin",
				UpdatedAt: 2222222222,
				UpdatedBy: "editor",
				DeletedAt: 3333333333,
				DeletedBy: "admin",
			},
			validate: func(t *testing.T, result *GuestEventResponseDTO, original *entities.GuestEventEntity) {
				assert.Equal(t, int64(3333333333), result.DeletedAt)

				assert.Equal(t, "admin", result.DeletedBy)

				assert.Equal(t, "Bob Johnson", result.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewGuestEventResponseDTO(tt.entity)

			assert.NotNil(t, result, "Result should not be nil")

			tt.validate(t, result, tt.entity)
		})
	}
}
