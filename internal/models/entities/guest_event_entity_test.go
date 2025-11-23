package entities

import (
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/guregu/null/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewGuestEventEntity(t *testing.T) {
	tests := []struct {
		name     string
		entity   *GuestEntity
		validate func(t *testing.T, result *GuestEventEntity, original *GuestEntity)
	}{
		{
			name: "create guest event entity with complete data",
			entity: &GuestEntity{
				ID:        uuid.Must(uuid.NewV4()),
				Name:      "John Doe",
				Address:   null.StringFrom("123 Main St"),
				CreatedAt: time.Now().UnixMilli(),
				CreatedBy: "admin",
				UpdatedAt: null.IntFrom(time.Now().Add(1 * time.Hour).UnixMilli()),
				UpdatedBy: null.StringFrom("editor"),
				DeletedAt: null.Int64{},
				DeletedBy: null.String{},
			},
			validate: func(t *testing.T, result *GuestEventEntity, original *GuestEntity) {
				assert.Equal(t, original.ID.String(), result.ID)
				assert.Equal(t, original.Name, result.Name)
				assert.Equal(t, original.Address.ValueOrZero(), result.Address)
				assert.Equal(t, original.CreatedAt, result.CreatedAt)
				assert.Equal(t, original.CreatedBy, result.CreatedBy)
				assert.Equal(t, original.UpdatedAt.ValueOrZero(), result.UpdatedAt)
				assert.Equal(t, original.UpdatedBy.ValueOrZero(), result.UpdatedBy)
				assert.Equal(t, original.DeletedAt.ValueOrZero(), result.DeletedAt)
				assert.Equal(t, original.DeletedBy.ValueOrZero(), result.DeletedBy)
			},
		},
		{
			name: "create guest event entity with minimal data",
			entity: &GuestEntity{
				ID:        uuid.Must(uuid.NewV4()),
				Name:      "Jane Smith",
				Address:   null.String{},
				CreatedAt: time.Now().UnixMilli(),
				CreatedBy: "system",
				UpdatedAt: null.Int64{},
				UpdatedBy: null.String{},
				DeletedAt: null.Int64{},
				DeletedBy: null.String{},
			},
			validate: func(t *testing.T, result *GuestEventEntity, original *GuestEntity) {
				assert.Equal(t, original.ID.String(), result.ID)
				assert.Equal(t, "Jane Smith", result.Name)
				assert.Equal(t, "", result.Address)
				assert.Equal(t, "system", result.CreatedBy)
				assert.Equal(t, int64(0), result.UpdatedAt)
				assert.Equal(t, "", result.UpdatedBy)
				assert.Equal(t, int64(0), result.DeletedAt)
				assert.Equal(t, "", result.DeletedBy)
			},
		},
		{
			name: "create guest event entity with deleted data",
			entity: &GuestEntity{
				ID:        uuid.Must(uuid.NewV4()),
				Name:      "Bob Johnson",
				Address:   null.StringFrom("789 Oak Ave"),
				CreatedAt: time.Now().Add(-48 * time.Hour).UnixMilli(),
				CreatedBy: "admin",
				UpdatedAt: null.IntFrom(time.Now().Add(-24 * time.Hour).UnixMilli()),
				UpdatedBy: null.StringFrom("moderator"),
				DeletedAt: null.IntFrom(time.Now().Add(-1 * time.Hour).UnixMilli()),
				DeletedBy: null.StringFrom("admin"),
			},
			validate: func(t *testing.T, result *GuestEventEntity, original *GuestEntity) {
				assert.Equal(t, "Bob Johnson", result.Name)
				assert.Equal(t, "789 Oak Ave", result.Address)
				assert.Equal(t, "moderator", result.UpdatedBy)
				assert.Equal(t, "admin", result.DeletedBy)
				assert.NotEqual(t, int64(0), result.DeletedAt, "DeletedAt should not be 0 for deleted entity")
			},
		},
		{
			name: "create guest event entity with empty name",
			entity: &GuestEntity{
				ID:        uuid.Must(uuid.NewV4()),
				Name:      "",
				Address:   null.StringFrom("No Name Street"),
				CreatedAt: time.Now().UnixMilli(),
				CreatedBy: "test",
				UpdatedAt: null.Int64{},
				UpdatedBy: null.String{},
				DeletedAt: null.Int64{},
				DeletedBy: null.String{},
			},
			validate: func(t *testing.T, result *GuestEventEntity, original *GuestEntity) {
				assert.Equal(t, "", result.Name)
				assert.Equal(t, "No Name Street", result.Address)
				assert.Equal(t, "test", result.CreatedBy)
			},
		},
		{
			name: "create guest event entity with special characters",
			entity: &GuestEntity{
				ID:        uuid.Must(uuid.NewV4()),
				Name:      "José María O'Connor",
				Address:   null.StringFrom("123 Résidences Élysées, Apartment #5"),
				CreatedAt: time.Now().UnixMilli(),
				CreatedBy: "admin@example.com",
				UpdatedAt: null.IntFrom(time.Now().UnixMilli()),
				UpdatedBy: null.StringFrom("user.name+tag@domain.co.uk"),
				DeletedAt: null.Int64{},
				DeletedBy: null.String{},
			},
			validate: func(t *testing.T, result *GuestEventEntity, original *GuestEntity) {
				assert.Equal(t, "José María O'Connor", result.Name)
				assert.Equal(t, "123 Résidences Élysées, Apartment #5", result.Address)
				assert.Equal(t, "admin@example.com", result.CreatedBy)
				assert.Equal(t, "user.name+tag@domain.co.uk", result.UpdatedBy)
			},
		},
		{
			name: "create guest event entity verifies instance independence",
			entity: &GuestEntity{
				ID:        uuid.Must(uuid.NewV4()),
				Name:      "Independence Test",
				Address:   null.StringFrom("Original Address"),
				CreatedAt: time.Now().UnixMilli(),
				CreatedBy: "creator",
				UpdatedAt: null.Int64{},
				UpdatedBy: null.String{},
				DeletedAt: null.Int64{},
				DeletedBy: null.String{},
			},
			validate: func(t *testing.T, result *GuestEventEntity, original *GuestEntity) {
				originalName := original.Name
				original.Name = "Modified Name"
				original.Address = null.StringFrom("Modified Address")

				assert.Equal(t, originalName, result.Name, "Expected Name to remain independent of original")
				assert.Equal(t, "Original Address", result.Address, "Expected Address to remain independent of original")

				original.Name = originalName
				original.Address = null.StringFrom("Original Address")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewGuestEventEntity(tt.entity)

			assert.NotNil(t, result, "Result should not be nil")

			tt.validate(t, result, tt.entity)
		})
	}
}
