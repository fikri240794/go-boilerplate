package entities

import (
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/guregu/null/v5"
	"github.com/stretchr/testify/assert"
)

func TestGuestEntity_MarkAsDeleted(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		entity    *GuestEntity
		deletedBy string
		validate  func(t *testing.T, result *GuestEntity, originalEntity *GuestEntity)
	}{
		{
			name: "successfully mark guest as deleted with valid deletedBy",
			entity: &GuestEntity{
				ID:        uuid.Must(uuid.NewV4()),
				Name:      "John Doe",
				Address:   null.StringFrom("123 Main St"),
				CreatedAt: time.Now().UnixMilli(),
				CreatedBy: "admin",
				UpdatedAt: null.Int64{},
				UpdatedBy: null.String{},
				DeletedAt: null.Int64{},
				DeletedBy: null.String{},
			},
			deletedBy: "admin_user",
			validate: func(t *testing.T, result *GuestEntity, originalEntity *GuestEntity) {
				assert.True(t, result.DeletedAt.Valid, "DeletedAt should be valid")
				assert.Greater(t, result.DeletedAt.Int64, int64(0), "DeletedAt should be greater than 0")
				assert.True(t, result.DeletedBy.Valid, "DeletedBy should be valid")
				assert.Equal(t, "admin_user", result.DeletedBy.String)

				deletedTime := time.UnixMilli(result.DeletedAt.Int64)
				timeDiff := deletedTime.Sub(now)
				if timeDiff < 0 {
					timeDiff = -timeDiff
				}
				assert.Less(t, timeDiff, time.Second, "DeletedAt should be close to current time")
			},
		},
		{
			name: "mark as deleted with empty deletedBy string",
			entity: &GuestEntity{
				ID:        uuid.Must(uuid.NewV4()),
				Name:      "Jane Smith",
				Address:   null.StringFrom("456 Oak Ave"),
				CreatedAt: time.Now().UnixMilli(),
				CreatedBy: "system",
				UpdatedAt: null.Int64{},
				UpdatedBy: null.String{},
				DeletedAt: null.Int64{},
				DeletedBy: null.String{},
			},
			deletedBy: "",
			validate: func(t *testing.T, result *GuestEntity, originalEntity *GuestEntity) {
				assert.True(t, result.DeletedAt.Valid, "DeletedAt should be valid")
				assert.Greater(t, result.DeletedAt.Int64, int64(0), "DeletedAt should be greater than 0")
				assert.True(t, result.DeletedBy.Valid, "DeletedBy should be valid")
				assert.Equal(t, "", result.DeletedBy.String)
			},
		},
		{
			name: "mark as deleted when entity already has updated fields",
			entity: &GuestEntity{
				ID:        uuid.Must(uuid.NewV4()),
				Name:      "Bob Johnson",
				Address:   null.StringFrom("789 Pine St"),
				CreatedAt: time.Now().Add(-24 * time.Hour).UnixMilli(),
				CreatedBy: "admin",
				UpdatedAt: null.IntFrom(time.Now().Add(-1 * time.Hour).UnixMilli()),
				UpdatedBy: null.StringFrom("editor"),
				DeletedAt: null.Int64{},
				DeletedBy: null.String{},
			},
			deletedBy: "moderator",
			validate: func(t *testing.T, result *GuestEntity, originalEntity *GuestEntity) {
				assert.True(t, result.DeletedAt.Valid, "DeletedAt should be valid")
				assert.Greater(t, result.DeletedAt.Int64, int64(0), "DeletedAt should be greater than 0")
				assert.True(t, result.DeletedBy.Valid, "DeletedBy should be valid")
				assert.Equal(t, "moderator", result.DeletedBy.String)

				assert.Equal(t, "Bob Johnson", result.Name)
				assert.Equal(t, "789 Pine St", result.Address.String)
				assert.True(t, result.UpdatedAt.Valid, "UpdatedAt should still be valid")
				assert.Equal(t, "editor", result.UpdatedBy.String)
			},
		},
		{
			name: "mark as deleted when entity already deleted (overwrite)",
			entity: &GuestEntity{
				ID:        uuid.Must(uuid.NewV4()),
				Name:      "Alice Brown",
				Address:   null.StringFrom("321 Elm St"),
				CreatedAt: time.Now().Add(-48 * time.Hour).UnixMilli(),
				CreatedBy: "admin",
				UpdatedAt: null.Int64{},
				UpdatedBy: null.String{},
				DeletedAt: null.IntFrom(time.Now().Add(-1 * time.Hour).UnixMilli()),
				DeletedBy: null.StringFrom("previous_admin"),
			},
			deletedBy: "new_admin",
			validate: func(t *testing.T, result *GuestEntity, originalEntity *GuestEntity) {
				assert.True(t, result.DeletedAt.Valid, "DeletedAt should be valid")
				assert.Greater(t, result.DeletedAt.Int64, int64(0), "DeletedAt should be greater than 0")
				assert.True(t, result.DeletedBy.Valid, "DeletedBy should be valid")
				assert.Equal(t, "new_admin", result.DeletedBy.String)

				newDeletedTime := time.UnixMilli(result.DeletedAt.Int64)
				oldDeletedTime := time.Now().Add(-1 * time.Hour)
				assert.True(t, newDeletedTime.After(oldDeletedTime), "New DeletedAt should be more recent than the old one")
			},
		},
		{
			name: "mark as deleted with special characters in deletedBy",
			entity: &GuestEntity{
				ID:        uuid.Must(uuid.NewV4()),
				Name:      "Charlie Wilson",
				Address:   null.String{},
				CreatedAt: time.Now().UnixMilli(),
				CreatedBy: "system",
				UpdatedAt: null.Int64{},
				UpdatedBy: null.String{},
				DeletedAt: null.Int64{},
				DeletedBy: null.String{},
			},
			deletedBy: "admin@example.com",
			validate: func(t *testing.T, result *GuestEntity, originalEntity *GuestEntity) {
				assert.True(t, result.DeletedAt.Valid, "DeletedAt should be valid")
				assert.Greater(t, result.DeletedAt.Int64, int64(0), "DeletedAt should be greater than 0")
				assert.True(t, result.DeletedBy.Valid, "DeletedBy should be valid")
				assert.Equal(t, "admin@example.com", result.DeletedBy.String)
			},
		},
		{
			name: "returns pointer to same instance",
			entity: &GuestEntity{
				ID:        uuid.Must(uuid.NewV4()),
				Name:      "Test User",
				CreatedAt: time.Now().UnixMilli(),
				CreatedBy: "test",
			},
			deletedBy: "admin",
			validate: func(t *testing.T, result *GuestEntity, originalEntity *GuestEntity) {
				assert.Equal(t, originalEntity, result, "Should return pointer to the same entity instance")
			},
		},
		{
			name: "concurrent execution consistency",
			entity: &GuestEntity{
				ID:        uuid.Must(uuid.NewV4()),
				Name:      "Concurrent Test User",
				CreatedAt: time.Now().UnixMilli(),
				CreatedBy: "test",
			},
			deletedBy: "admin",
			validate: func(t *testing.T, result *GuestEntity, originalEntity *GuestEntity) {
				for i := 0; i < 10; i++ {
					iterResult := originalEntity.MarkAsDeleted("admin")

					assert.NotNil(t, iterResult, "Result should not be nil")
					assert.True(t, iterResult.DeletedAt.Valid, "DeletedAt should be valid")
					assert.True(t, iterResult.DeletedBy.Valid, "DeletedBy should be valid")
					assert.Equal(t, "admin", iterResult.DeletedBy.String)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entity.MarkAsDeleted(tt.deletedBy)

			assert.NotNil(t, result, "Result should not be nil")
			assert.Equal(t, tt.entity, result, "Should return the same entity instance")

			tt.validate(t, result, tt.entity)
		})
	}
}
