package dtos

import (
	"go-boilerplate/internal/models/entities"
	"testing"
	"time"

	"github.com/fikri240794/goqube"
	"github.com/gofrs/uuid/v5"
	"github.com/guregu/null/v5"
	"github.com/stretchr/testify/assert"
)

func TestCreateGuestRequestDTO_Validate(t *testing.T) {
	tests := []struct {
		name        string
		dto         *CreateGuestRequestDTO
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "valid create guest request with all fields",
			dto: &CreateGuestRequestDTO{
				Name:      "John Doe",
				Address:   "123 Main St",
				CreatedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "valid create guest request without address",
			dto: &CreateGuestRequestDTO{
				Name:      "Jane Smith",
				Address:   "",
				CreatedBy: "system",
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "invalid create guest request missing name",
			dto: &CreateGuestRequestDTO{
				Name:      "",
				Address:   "456 Oak Ave",
				CreatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected validation error for missing name")
			},
		},
		{
			name: "invalid create guest request missing created_by",
			dto: &CreateGuestRequestDTO{
				Name:      "Bob Johnson",
				Address:   "789 Pine St",
				CreatedBy: "",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected validation error for missing created_by")
			},
		},
		{
			name: "invalid create guest request missing both required fields",
			dto: &CreateGuestRequestDTO{
				Name:      "",
				Address:   "Test Address",
				CreatedBy: "",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected validation error for missing required fields")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dto.Validate()

			tt.validate(t, err)
		})
	}
}

func TestCreateGuestRequestDTO_ToEntity(t *testing.T) {
	tests := []struct {
		name     string
		dto      *CreateGuestRequestDTO
		validate func(t *testing.T, result *entities.GuestEntity, original *CreateGuestRequestDTO)
	}{
		{
			name: "convert DTO to entity with address",
			dto: &CreateGuestRequestDTO{
				Name:      "John Doe",
				Address:   "123 Main St",
				CreatedBy: "admin",
			},
			validate: func(t *testing.T, result *entities.GuestEntity, original *CreateGuestRequestDTO) {
				assert.Equal(t, original.Name, result.Name)
				assert.True(t, result.Address.Valid, "Address should be valid when provided")
				assert.Equal(t, original.Address, result.Address.String)
				assert.Equal(t, original.CreatedBy, result.CreatedBy)
				assert.Greater(t, result.CreatedAt, int64(0), "CreatedAt should be set to current timestamp")
				assert.NotEqual(t, "", result.ID.String(), "ID should be generated")
			},
		},
		{
			name: "convert DTO to entity without address",
			dto: &CreateGuestRequestDTO{
				Name:      "Jane Smith",
				Address:   "",
				CreatedBy: "system",
			},
			validate: func(t *testing.T, result *entities.GuestEntity, original *CreateGuestRequestDTO) {
				assert.Equal(t, "Jane Smith", result.Name)
				assert.False(t, result.Address.Valid, "Address should be invalid when empty")
				assert.Equal(t, "system", result.CreatedBy)
			},
		},
		{
			name: "convert DTO to entity with special characters",
			dto: &CreateGuestRequestDTO{
				Name:      "José María",
				Address:   "123 Élysées St, Apt #5",
				CreatedBy: "admin@example.com",
			},
			validate: func(t *testing.T, result *entities.GuestEntity, original *CreateGuestRequestDTO) {
				assert.Equal(t, "José María", result.Name)
				assert.Equal(t, "123 Élysées St, Apt #5", result.Address.String)
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

func TestDeleteGuestByIDRequestDTO_Validate(t *testing.T) {
	validUUID := uuid.Must(uuid.NewV4()).String()

	tests := []struct {
		name        string
		dto         *DeleteGuestByIDRequestDTO
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "valid delete guest request",
			dto: &DeleteGuestByIDRequestDTO{
				ID:        validUUID,
				DeletedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "invalid delete guest request with invalid UUID",
			dto: &DeleteGuestByIDRequestDTO{
				ID:        "invalid-uuid",
				DeletedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected validation error for invalid UUID")
			},
		},
		{
			name: "invalid delete guest request missing deleted_by",
			dto: &DeleteGuestByIDRequestDTO{
				ID:        validUUID,
				DeletedBy: "",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected validation error for missing deleted_by")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dto.Validate()

			tt.validate(t, err)
		})
	}
}

func TestNewFindAllGuestRequestDTO(t *testing.T) {
	tests := []struct {
		name     string
		validate func(t *testing.T, result *FindAllGuestRequestDTO)
	}{
		{
			name: "create new find all guest request DTO with default values",
			validate: func(t *testing.T, result *FindAllGuestRequestDTO) {
				assert.Equal(t, uint64(10), result.Take)
				assert.Equal(t, uint64(0), result.Skip)
				assert.Equal(t, "", result.Keyword)
				assert.Equal(t, "", result.Sorts)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewFindAllGuestRequestDTO()

			assert.NotNil(t, result, "Result should not be nil")

			tt.validate(t, result)
		})
	}
}

func TestFindAllGuestRequestDTO_ToFilterAndSorts(t *testing.T) {
	tests := []struct {
		name     string
		dto      *FindAllGuestRequestDTO
		validate func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error)
	}{
		{
			name: "convert DTO without keyword and sorts",
			dto: &FindAllGuestRequestDTO{
				Keyword: "",
				Sorts:   "",
				Take:    10,
				Skip:    0,
			},
			validate: func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, filter)
				assert.Equal(t, goqube.LogicAnd, filter.Logic)
				assert.Len(t, filter.Filters, 1)

				deletedAtFilter := filter.Filters[0]
				assert.Equal(t, entities.GuestEntityDatabaseFieldDeletedAt, deletedAtFilter.Field.Column)
				assert.Equal(t, goqube.OperatorIsNull, deletedAtFilter.Operator)
				assert.Empty(t, sorts)
			},
		},
		{
			name: "convert DTO with keyword",
			dto: &FindAllGuestRequestDTO{
				Keyword: "john",
				Sorts:   "",
				Take:    10,
				Skip:    0,
			},
			validate: func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error) {
				assert.NoError(t, err)

				assert.Len(t, filter.Filters, 2)

				keywordFilter := filter.Filters[1]
				assert.Equal(t, goqube.LogicOr, keywordFilter.Logic)

				assert.Len(t, keywordFilter.Filters, 2)

				nameFilter := keywordFilter.Filters[0]
				assert.Equal(t, entities.GuestEntityDatabaseFieldName, nameFilter.Field.Column)

				assert.Equal(t, goqube.OperatorLike, nameFilter.Operator)

				assert.Equal(t, "john", nameFilter.Value.Value)

				addressFilter := keywordFilter.Filters[1]
				assert.Equal(t, entities.GuestEntityDatabaseFieldAddress, addressFilter.Field.Column)
			},
		},
		{
			name: "convert DTO with single sort field (name only)",
			dto: &FindAllGuestRequestDTO{
				Keyword: "",
				Sorts:   "name",
				Take:    10,
				Skip:    0,
			},
			validate: func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error) {
				assert.NoError(t, err)

				assert.Len(t, sorts, 1)

				if len(sorts) > 0 {
					sort := sorts[0]
					assert.Equal(t, entities.GuestEntityDatabaseFieldName, sort.Field.Column)

					assert.Equal(t, goqube.SortDirectionAscending, sort.Direction)
				}
			},
		},
		{
			name: "convert DTO with sort field and direction",
			dto: &FindAllGuestRequestDTO{
				Keyword: "",
				Sorts:   "name.DESC",
				Take:    10,
				Skip:    0,
			},
			validate: func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error) {
				assert.NoError(t, err)

				assert.Len(t, sorts, 1)

				if len(sorts) > 0 {
					sort := sorts[0]
					assert.Equal(t, entities.GuestEntityDatabaseFieldName, sort.Field.Column)

					assert.Equal(t, goqube.SortDirectionDescending, sort.Direction)
				}
			},
		},
		{
			name: "convert DTO with multiple sorts",
			dto: &FindAllGuestRequestDTO{
				Keyword: "",
				Sorts:   "name.ASC,address.DESC",
				Take:    10,
				Skip:    0,
			},
			validate: func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error) {
				assert.NoError(t, err)

				assert.Len(t, sorts, 2)

				if len(sorts) >= 2 {
					sort1 := sorts[0]
					assert.Equal(t, entities.GuestEntityDatabaseFieldName, sort1.Field.Column)

					assert.Equal(t, goqube.SortDirectionAscending, sort1.Direction)

					sort2 := sorts[1]
					assert.Equal(t, entities.GuestEntityDatabaseFieldAddress, sort2.Field.Column)

					assert.Equal(t, goqube.SortDirectionDescending, sort2.Direction)
				}
			},
		},
		{
			name: "convert DTO with keyword and sorts",
			dto: &FindAllGuestRequestDTO{
				Keyword: "test search",
				Sorts:   "name.ASC",
				Take:    20,
				Skip:    10,
			},
			validate: func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error) {
				assert.NoError(t, err)

				assert.Len(t, filter.Filters, 2)

				assert.Len(t, sorts, 1)

				if len(filter.Filters) >= 2 {
					keywordFilter := filter.Filters[1]
					if len(keywordFilter.Filters) >= 1 {
						nameFilter := keywordFilter.Filters[0]
						assert.Equal(t, "test search", nameFilter.Value.Value)
					}
				}
			},
		},
		{
			name: "invalid sort field",
			dto: &FindAllGuestRequestDTO{
				Keyword: "",
				Sorts:   "invalid_field",
				Take:    10,
				Skip:    0,
			},
			validate: func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error) {
				assert.Error(t, err, "Expected error for invalid sort field")
				assert.Nil(t, filter, "Filter should be nil on error")
				assert.Nil(t, sorts, "Sorts should be nil on error")
			},
		},
		{
			name: "invalid sort direction",
			dto: &FindAllGuestRequestDTO{
				Keyword: "",
				Sorts:   "name.invalid",
				Take:    10,
				Skip:    0,
			},
			validate: func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error) {
				assert.Error(t, err, "Expected error for invalid sort direction")
				assert.Nil(t, filter, "Filter should be nil on error")
				assert.Nil(t, sorts, "Sorts should be nil on error")
			},
		},
		{
			name: "mixed valid and invalid sorts",
			dto: &FindAllGuestRequestDTO{
				Keyword: "",
				Sorts:   "name.ASC,invalid_field.DESC",
				Take:    10,
				Skip:    0,
			},
			validate: func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error) {
				assert.Error(t, err, "Expected error for invalid sort field in mixed sorts")
			},
		},
		{
			name: "address field with ascending direction",
			dto: &FindAllGuestRequestDTO{
				Keyword: "",
				Sorts:   "address.ASC",
				Take:    10,
				Skip:    0,
			},
			validate: func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error) {
				assert.NoError(t, err)

				assert.Len(t, sorts, 1)

				if len(sorts) > 0 {
					sort := sorts[0]
					assert.Equal(t, entities.GuestEntityDatabaseFieldAddress, sort.Field.Column)

					assert.Equal(t, goqube.SortDirectionAscending, sort.Direction)
				}
			},
		},
		{
			name: "empty sorts string should result in no sorts",
			dto: &FindAllGuestRequestDTO{
				Keyword: "test",
				Sorts:   "",
				Take:    10,
				Skip:    0,
			},
			validate: func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error) {
				assert.NoError(t, err)

				assert.Len(t, sorts, 0)
			},
		},
		{
			name: "sorts with only commas should cause error",
			dto: &FindAllGuestRequestDTO{
				Keyword: "",
				Sorts:   ",,,",
				Take:    10,
				Skip:    0,
			},
			validate: func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error) {
				assert.Error(t, err, "Expected error for sorts with only commas")
				assert.Nil(t, filter, "Filter should be nil on error")
				assert.Nil(t, sorts, "Sorts should be nil on error")
			},
		},
		{
			name: "sorts with empty strings between commas should cause error",
			dto: &FindAllGuestRequestDTO{
				Keyword: "",
				Sorts:   "name,,address",
				Take:    10,
				Skip:    0,
			},
			validate: func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error) {
				assert.Error(t, err, "Expected error for empty sort field between commas")
				assert.Nil(t, filter, "Filter should be nil on error")
				assert.Nil(t, sorts, "Sorts should be nil on error")
			},
		},
		{
			name: "sorts with only dots should cause invalid sorts value error",
			dto: &FindAllGuestRequestDTO{
				Keyword: "",
				Sorts:   ".",
				Take:    10,
				Skip:    0,
			},
			validate: func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error) {
				assert.Error(t, err, "Expected error for sorts with only dots")
				assert.Nil(t, filter, "Filter should be nil on error")
				assert.Nil(t, sorts, "Sorts should be nil on error")
			},
		},
		{
			name: "sorts with multiple consecutive dots should cause error",
			dto: &FindAllGuestRequestDTO{
				Keyword: "",
				Sorts:   "name...",
				Take:    10,
				Skip:    0,
			},
			validate: func(t *testing.T, filter *goqube.Filter, sorts []goqube.Sort, err error) {
				assert.Error(t, err, "Expected error for sorts with multiple consecutive dots")
				assert.Nil(t, filter, "Filter should be nil on error")
				assert.Nil(t, sorts, "Sorts should be nil on error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, sorts, err := tt.dto.ToFilterAndSorts()

			tt.validate(t, filter, sorts, err)
		})
	}
}

func TestNewFindAllGuestResponseDTO(t *testing.T) {
	tests := []struct {
		name       string
		listEntity []entities.GuestEntity
		count      uint64
		validate   func(t *testing.T, result *FindAllGuestResponseDTO, originalList []entities.GuestEntity, originalCount uint64)
	}{
		{
			name: "create response DTO with entity list",
			listEntity: []entities.GuestEntity{
				{
					ID:        uuid.Must(uuid.NewV4()),
					Name:      "John Doe",
					Address:   null.StringFrom("123 Main St"),
					CreatedAt: time.Now().UnixMilli(),
					CreatedBy: "admin",
				},
				{
					ID:        uuid.Must(uuid.NewV4()),
					Name:      "Jane Smith",
					Address:   null.String{},
					CreatedAt: time.Now().UnixMilli(),
					CreatedBy: "system",
				},
			},
			count: 2,
			validate: func(t *testing.T, result *FindAllGuestResponseDTO, originalList []entities.GuestEntity, originalCount uint64) {
				assert.Equal(t, originalCount, result.Count)

				assert.Len(t, result.List, len(originalList))

				for i, dto := range result.List {
					assert.Equal(t, originalList[i].ID.String(), dto.ID)

					assert.Equal(t, originalList[i].Name, dto.Name)
				}
			},
		},
		{
			name:       "create response DTO with empty list",
			listEntity: []entities.GuestEntity{},
			count:      0,
			validate: func(t *testing.T, result *FindAllGuestResponseDTO, originalList []entities.GuestEntity, originalCount uint64) {
				assert.Equal(t, uint64(0), result.Count)

				assert.Len(t, result.List, 0)
			},
		},
		{
			name:       "create response DTO with nil list",
			listEntity: nil,
			count:      0,
			validate: func(t *testing.T, result *FindAllGuestResponseDTO, originalList []entities.GuestEntity, originalCount uint64) {
				assert.Equal(t, uint64(0), result.Count)

				assert.Len(t, result.List, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewFindAllGuestResponseDTO(tt.listEntity, tt.count)

			assert.NotNil(t, result, "Result should not be nil")

			tt.validate(t, result, tt.listEntity, tt.count)
		})
	}
}

func TestFindGuestByIDRequestDTO_Validate(t *testing.T) {
	validUUID := uuid.Must(uuid.NewV4()).String()

	tests := []struct {
		name        string
		dto         *FindGuestByIDRequestDTO
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "valid find guest by ID request",
			dto: &FindGuestByIDRequestDTO{
				ID: validUUID,
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "invalid find guest by ID request with invalid UUID",
			dto: &FindGuestByIDRequestDTO{
				ID: "invalid-uuid",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected validation error for invalid UUID")
			},
		},
		{
			name: "invalid find guest by ID request with empty ID",
			dto: &FindGuestByIDRequestDTO{
				ID: "",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected validation error for empty ID")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dto.Validate()

			tt.validate(t, err)
		})
	}
}

func TestNewGuestResponseDTO(t *testing.T) {
	tests := []struct {
		name     string
		entity   *entities.GuestEntity
		validate func(t *testing.T, result *GuestResponseDTO, original *entities.GuestEntity)
	}{
		{
			name: "create response DTO with complete entity",
			entity: &entities.GuestEntity{
				ID:        uuid.Must(uuid.NewV4()),
				Name:      "John Doe",
				Address:   null.StringFrom("123 Main St"),
				CreatedAt: time.Now().UnixMilli(),
				CreatedBy: "admin",
				UpdatedAt: null.IntFrom(time.Now().Add(1 * time.Hour).UnixMilli()),
				UpdatedBy: null.StringFrom("editor"),
			},
			validate: func(t *testing.T, result *GuestResponseDTO, original *entities.GuestEntity) {
				assert.Equal(t, original.ID.String(), result.ID)

				assert.Equal(t, original.Name, result.Name)

				assert.Equal(t, original.Address.ValueOrZero(), result.Address)

				assert.Equal(t, original.CreatedAt, result.CreatedAt)

				assert.Equal(t, original.CreatedBy, result.CreatedBy)

				assert.Equal(t, original.UpdatedAt.ValueOrZero(), result.UpdatedAt)

				assert.Equal(t, original.UpdatedBy.ValueOrZero(), result.UpdatedBy)
			},
		},
		{
			name: "create response DTO with minimal entity",
			entity: &entities.GuestEntity{
				ID:        uuid.Must(uuid.NewV4()),
				Name:      "Jane Smith",
				Address:   null.String{},
				CreatedAt: time.Now().UnixMilli(),
				CreatedBy: "system",
				UpdatedAt: null.Int64{},
				UpdatedBy: null.String{},
			},
			validate: func(t *testing.T, result *GuestResponseDTO, original *entities.GuestEntity) {
				assert.Equal(t, "", result.Address)

				assert.Equal(t, int64(0), result.UpdatedAt)

				assert.Equal(t, "", result.UpdatedBy)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewGuestResponseDTO(tt.entity)

			assert.NotNil(t, result, "Result should not be nil")

			tt.validate(t, result, tt.entity)
		})
	}
}

func TestUpdateGuestByIDRequestDTO_Validate(t *testing.T) {
	validUUID := uuid.Must(uuid.NewV4()).String()

	tests := []struct {
		name        string
		dto         *UpdateGuestByIDRequestDTO
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "valid update guest request with all fields",
			dto: &UpdateGuestByIDRequestDTO{
				ID:        validUUID,
				Name:      "Updated Name",
				Address:   "Updated Address",
				UpdatedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "valid update guest request without address",
			dto: &UpdateGuestByIDRequestDTO{
				ID:        validUUID,
				Name:      "Updated Name",
				Address:   "",
				UpdatedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "invalid update guest request with invalid UUID",
			dto: &UpdateGuestByIDRequestDTO{
				ID:        "invalid-uuid",
				Name:      "Updated Name",
				Address:   "Updated Address",
				UpdatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected validation error for invalid UUID")
			},
		},
		{
			name: "invalid update guest request missing name",
			dto: &UpdateGuestByIDRequestDTO{
				ID:        validUUID,
				Name:      "",
				Address:   "Updated Address",
				UpdatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected validation error for missing name")
			},
		},
		{
			name: "invalid update guest request missing updated_by",
			dto: &UpdateGuestByIDRequestDTO{
				ID:        validUUID,
				Name:      "Updated Name",
				Address:   "Updated Address",
				UpdatedBy: "",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "Expected validation error for missing updated_by")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dto.Validate()

			tt.validate(t, err)
		})
	}
}

func TestUpdateGuestByIDRequestDTO_ToExistingEntity(t *testing.T) {
	validUUID := uuid.Must(uuid.NewV4())

	tests := []struct {
		name           string
		dto            *UpdateGuestByIDRequestDTO
		existingEntity *entities.GuestEntity
		validate       func(t *testing.T, result *entities.GuestEntity, dto *UpdateGuestByIDRequestDTO, original *entities.GuestEntity)
	}{
		{
			name: "update existing entity with all fields",
			dto: &UpdateGuestByIDRequestDTO{
				ID:        validUUID.String(),
				Name:      "Updated Name",
				Address:   "Updated Address",
				UpdatedBy: "admin",
			},
			existingEntity: &entities.GuestEntity{
				ID:        validUUID,
				Name:      "Original Name",
				Address:   null.StringFrom("Original Address"),
				CreatedAt: time.Now().Add(-24 * time.Hour).UnixMilli(),
				CreatedBy: "creator",
				UpdatedAt: null.Int64{},
				UpdatedBy: null.String{},
			},
			validate: func(t *testing.T, result *entities.GuestEntity, dto *UpdateGuestByIDRequestDTO, original *entities.GuestEntity) {
				assert.Equal(t, dto.Name, result.Name)

				if !result.Address.Valid {
					assert.True(t, result.Address.Valid, "Address should be valid when provided")
				}

				assert.Equal(t, dto.Address, result.Address.String)

				if !result.UpdatedAt.Valid {
					assert.True(t, result.UpdatedAt.Valid, "UpdatedAt should be valid after update")
				}

				if result.UpdatedAt.Int64 <= 0 {
					assert.Greater(t, result.UpdatedAt.Int64, int64(0), "UpdatedAt should be set to current timestamp")
				}

				if !result.UpdatedBy.Valid {
					assert.True(t, result.UpdatedBy.Valid, "UpdatedBy should be valid after update")
				}

				assert.Equal(t, dto.UpdatedBy, result.UpdatedBy.String)

				if result.ID != original.ID {
					assert.Equal(t, original.ID, result.ID, "ID should not be changed")
				}

				if result.CreatedAt != original.CreatedAt {
					assert.Equal(t, original.CreatedAt, result.CreatedAt, "CreatedAt should not be changed")
				}

				if result.CreatedBy != original.CreatedBy {
					assert.Equal(t, original.CreatedBy, result.CreatedBy, "CreatedBy should not be changed")
				}
			},
		},
		{
			name: "update existing entity removing address",
			dto: &UpdateGuestByIDRequestDTO{
				ID:        validUUID.String(),
				Name:      "Updated Name",
				Address:   "",
				UpdatedBy: "admin",
			},
			existingEntity: &entities.GuestEntity{
				ID:        validUUID,
				Name:      "Original Name",
				Address:   null.StringFrom("Original Address"),
				CreatedAt: time.Now().Add(-24 * time.Hour).UnixMilli(),
				CreatedBy: "creator",
			},
			validate: func(t *testing.T, result *entities.GuestEntity, dto *UpdateGuestByIDRequestDTO, original *entities.GuestEntity) {
				assert.False(t, result.Address.Valid, "Address should be invalid when empty string provided")

				assert.Equal(t, "Updated Name", result.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalCopy := *tt.existingEntity

			result := tt.dto.ToExistingEntity(tt.existingEntity)

			assert.NotNil(t, result, "Result should not be nil")

			assert.Equal(t, tt.existingEntity, result, "Should return the same entity instance")

			tt.validate(t, result, tt.dto, &originalCopy)
		})
	}
}

func TestBulkCreateGuestsRequestDTO_Validate(t *testing.T) {
	tests := []struct {
		name        string
		dto         *BulkCreateGuestsRequestDTO
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "valid bulk create guests request with single item",
			dto: &BulkCreateGuestsRequestDTO{
				Items: []CreateGuestRequestDTO{
					{Name: "John Doe", Address: "123 Main St", CreatedBy: "admin"},
				},
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "valid bulk create guests request with multiple items",
			dto: &BulkCreateGuestsRequestDTO{
				Items: []CreateGuestRequestDTO{
					{Name: "John Doe", Address: "123 Main St", CreatedBy: "admin"},
					{Name: "Jane Smith", CreatedBy: "system"},
				},
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "invalid bulk create guests request with empty items",
			dto: &BulkCreateGuestsRequestDTO{
				Items: []CreateGuestRequestDTO{},
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:        "invalid bulk create guests request with nil items",
			dto:         &BulkCreateGuestsRequestDTO{},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dto.Validate()

			tt.validate(t, err)
		})
	}
}

func TestBulkCreateGuestsRequestDTO_ToEntities(t *testing.T) {
	tests := []struct {
		name     string
		dto      *BulkCreateGuestsRequestDTO
		validate func(t *testing.T, result []entities.GuestEntity)
	}{
		{
			name: "convert single item to entities",
			dto: &BulkCreateGuestsRequestDTO{
				Items: []CreateGuestRequestDTO{
					{Name: "John Doe", Address: "123 Main St", CreatedBy: "admin"},
				},
			},
			validate: func(t *testing.T, result []entities.GuestEntity) {
				assert.Len(t, result, 1)
				assert.Equal(t, "John Doe", result[0].Name)
				assert.Equal(t, "admin", result[0].CreatedBy)
			},
		},
		{
			name: "convert multiple items to entities",
			dto: &BulkCreateGuestsRequestDTO{
				Items: []CreateGuestRequestDTO{
					{Name: "John Doe", CreatedBy: "admin"},
					{Name: "Jane Smith", CreatedBy: "system"},
				},
			},
			validate: func(t *testing.T, result []entities.GuestEntity) {
				assert.Len(t, result, 2)
				assert.Equal(t, "John Doe", result[0].Name)
				assert.Equal(t, "Jane Smith", result[1].Name)
			},
		},
		{
			name: "convert empty items to empty slice",
			dto: &BulkCreateGuestsRequestDTO{
				Items: []CreateGuestRequestDTO{},
			},
			validate: func(t *testing.T, result []entities.GuestEntity) {
				assert.Len(t, result, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dto.ToEntities()

			tt.validate(t, result)
		})
	}
}

func TestNewBulkCreateGuestsResponseDTO(t *testing.T) {
	now := time.Now().UnixMilli()

	tests := []struct {
		name     string
		entities []entities.GuestEntity
		validate func(t *testing.T, result *BulkCreateGuestsResponseDTO)
	}{
		{
			name: "create response with multiple entities",
			entities: []entities.GuestEntity{
				{ID: uuid.Must(uuid.NewV7()), Name: "John Doe", CreatedAt: now, CreatedBy: "admin"},
				{ID: uuid.Must(uuid.NewV7()), Name: "Jane Smith", CreatedAt: now, CreatedBy: "system"},
			},
			validate: func(t *testing.T, result *BulkCreateGuestsResponseDTO) {
				assert.NotNil(t, result)
				assert.Len(t, result.Guests, 2)
				assert.Equal(t, "John Doe", result.Guests[0].Name)
				assert.Equal(t, "Jane Smith", result.Guests[1].Name)
			},
		},
		{
			name:     "create response with empty entities",
			entities: []entities.GuestEntity{},
			validate: func(t *testing.T, result *BulkCreateGuestsResponseDTO) {
				assert.NotNil(t, result)
				assert.Len(t, result.Guests, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewBulkCreateGuestsResponseDTO(tt.entities)

			tt.validate(t, result)
		})
	}
}

func TestBulkUpdateGuestsRequestDTO_Validate(t *testing.T) {
	validUUID := uuid.Must(uuid.NewV4()).String()

	tests := []struct {
		name        string
		dto         *BulkUpdateGuestsRequestDTO
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "valid bulk update guests request with single item",
			dto: &BulkUpdateGuestsRequestDTO{
				Items: []UpdateGuestByIDRequestDTO{
					{ID: validUUID, Name: "Updated Name", Address: "123 Main St", UpdatedBy: "admin"},
				},
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "valid bulk update guests request with multiple items",
			dto: &BulkUpdateGuestsRequestDTO{
				Items: []UpdateGuestByIDRequestDTO{
					{ID: validUUID, Name: "Name 1", UpdatedBy: "admin"},
					{ID: validUUID, Name: "Name 2", UpdatedBy: "system"},
				},
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "invalid bulk update guests request with empty items",
			dto: &BulkUpdateGuestsRequestDTO{
				Items: []UpdateGuestByIDRequestDTO{},
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:        "invalid bulk update guests request with nil items",
			dto:         &BulkUpdateGuestsRequestDTO{},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dto.Validate()

			tt.validate(t, err)
		})
	}
}

func TestNewBulkUpdateGuestsResponseDTO(t *testing.T) {
	now := time.Now().UnixMilli()

	tests := []struct {
		name     string
		entities []entities.GuestEntity
		validate func(t *testing.T, result *BulkUpdateGuestsResponseDTO)
	}{
		{
			name: "create response with multiple entities",
			entities: []entities.GuestEntity{
				{ID: uuid.Must(uuid.NewV7()), Name: "John Doe", CreatedAt: now, CreatedBy: "admin"},
				{ID: uuid.Must(uuid.NewV7()), Name: "Jane Smith", CreatedAt: now, CreatedBy: "system"},
			},
			validate: func(t *testing.T, result *BulkUpdateGuestsResponseDTO) {
				assert.NotNil(t, result)
				assert.Len(t, result.Guests, 2)
				assert.Equal(t, "John Doe", result.Guests[0].Name)
				assert.Equal(t, "Jane Smith", result.Guests[1].Name)
			},
		},
		{
			name:     "create response with empty entities",
			entities: []entities.GuestEntity{},
			validate: func(t *testing.T, result *BulkUpdateGuestsResponseDTO) {
				assert.NotNil(t, result)
				assert.Len(t, result.Guests, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewBulkUpdateGuestsResponseDTO(tt.entities)

			tt.validate(t, result)
		})
	}
}

func TestBulkDeleteGuestsRequestDTO_Validate(t *testing.T) {
	validUUID := uuid.Must(uuid.NewV4()).String()

	tests := []struct {
		name        string
		dto         *BulkDeleteGuestsRequestDTO
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "valid bulk delete guests request with single id",
			dto: &BulkDeleteGuestsRequestDTO{
				IDs:       []string{validUUID},
				DeletedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "valid bulk delete guests request with multiple ids",
			dto: &BulkDeleteGuestsRequestDTO{
				IDs:       []string{validUUID, validUUID},
				DeletedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "invalid bulk delete guests request with empty ids",
			dto: &BulkDeleteGuestsRequestDTO{
				IDs:       []string{},
				DeletedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "invalid bulk delete guests request with nil ids",
			dto: &BulkDeleteGuestsRequestDTO{
				DeletedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "invalid bulk delete guests request with invalid uuid",
			dto: &BulkDeleteGuestsRequestDTO{
				IDs:       []string{"invalid-uuid"},
				DeletedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "invalid bulk delete guests request missing deleted_by",
			dto: &BulkDeleteGuestsRequestDTO{
				IDs: []string{validUUID},
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dto.Validate()

			tt.validate(t, err)
		})
	}
}

func TestBulkUpdateGuestsRequestDTO_ToIDs(t *testing.T) {
	validUUID := uuid.Must(uuid.NewV4()).String()

	tests := []struct {
		name     string
		dto      *BulkUpdateGuestsRequestDTO
		expected []string
	}{
		{
			name: "to ids with single item",
			dto: &BulkUpdateGuestsRequestDTO{
				Items: []UpdateGuestByIDRequestDTO{
					{ID: validUUID, Name: "Name 1", UpdatedBy: "admin"},
				},
			},
			expected: []string{validUUID},
		},
		{
			name: "to ids with multiple items",
			dto: &BulkUpdateGuestsRequestDTO{
				Items: []UpdateGuestByIDRequestDTO{
					{ID: "id-1", Name: "Name 1", UpdatedBy: "admin"},
					{ID: "id-2", Name: "Name 2", UpdatedBy: "system"},
					{ID: "id-3", Name: "Name 3", UpdatedBy: "user"},
				},
			},
			expected: []string{"id-1", "id-2", "id-3"},
		},
		{
			name: "to ids with empty items",
			dto: &BulkUpdateGuestsRequestDTO{
				Items: []UpdateGuestByIDRequestDTO{},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dto.ToIDs()

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBulkDeleteGuestsRequestDTO_ToIDs(t *testing.T) {
	tests := []struct {
		name     string
		dto      *BulkDeleteGuestsRequestDTO
		expected []string
	}{
		{
			name: "to ids with multiple ids",
			dto: &BulkDeleteGuestsRequestDTO{
				IDs: []string{"id-1", "id-2", "id-3"},
			},
			expected: []string{"id-1", "id-2", "id-3"},
		},
		{
			name: "to ids with single id",
			dto: &BulkDeleteGuestsRequestDTO{
				IDs: []string{"single-id"},
			},
			expected: []string{"single-id"},
		},
		{
			name: "to ids with empty ids",
			dto: &BulkDeleteGuestsRequestDTO{
				IDs: []string{},
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dto.ToIDs()

			assert.Equal(t, tt.expected, result)
		})
	}
}
