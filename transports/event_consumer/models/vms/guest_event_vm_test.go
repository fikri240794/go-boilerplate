package vms

import (
	"testing"

	"go-boilerplate/internal/models/dtos"

	"github.com/stretchr/testify/assert"
)

func TestGuestEventRequestVM_ToDTO(t *testing.T) {
	tests := []struct {
		name     string
		setupVM  func() *GuestEventRequestVM
		validate func(t *testing.T, dto *dtos.GuestEventRequestDTO)
	}{
		{
			name: "should_convert_vm_to_dto_with_all_fields",
			setupVM: func() *GuestEventRequestVM {
				return &GuestEventRequestVM{
					ID:        "guest-123",
					Name:      "John Doe",
					Address:   "123 Main St",
					CreatedAt: 1700000000,
					CreatedBy: "user-1",
					UpdatedAt: 1700000100,
					UpdatedBy: "user-2",
					DeletedAt: 1700000200,
					DeletedBy: "user-3",
				}
			},
			validate: func(t *testing.T, dto *dtos.GuestEventRequestDTO) {
				assert.NotNil(t, dto)
				assert.Equal(t, "guest-123", dto.ID)
				assert.Equal(t, "John Doe", dto.Name)
				assert.Equal(t, "123 Main St", dto.Address)
				assert.Equal(t, int64(1700000000), dto.CreatedAt)
				assert.Equal(t, "user-1", dto.CreatedBy)
				assert.Equal(t, int64(1700000100), dto.UpdatedAt)
				assert.Equal(t, "user-2", dto.UpdatedBy)
				assert.Equal(t, int64(1700000200), dto.DeletedAt)
				assert.Equal(t, "user-3", dto.DeletedBy)
			},
		},
		{
			name: "should_convert_vm_to_dto_with_required_fields_only",
			setupVM: func() *GuestEventRequestVM {
				return &GuestEventRequestVM{
					ID:        "guest-456",
					Name:      "Jane Smith",
					CreatedAt: 1700001000,
					CreatedBy: "admin",
				}
			},
			validate: func(t *testing.T, dto *dtos.GuestEventRequestDTO) {
				assert.NotNil(t, dto)
				assert.Equal(t, "guest-456", dto.ID)
				assert.Equal(t, "Jane Smith", dto.Name)
				assert.Equal(t, "", dto.Address)
				assert.Equal(t, int64(1700001000), dto.CreatedAt)
				assert.Equal(t, "admin", dto.CreatedBy)
				assert.Equal(t, int64(0), dto.UpdatedAt)
				assert.Equal(t, "", dto.UpdatedBy)
				assert.Equal(t, int64(0), dto.DeletedAt)
				assert.Equal(t, "", dto.DeletedBy)
			},
		},
		{
			name: "should_convert_vm_to_dto_with_empty_values",
			setupVM: func() *GuestEventRequestVM {
				return &GuestEventRequestVM{
					ID:        "",
					Name:      "",
					Address:   "",
					CreatedAt: 0,
					CreatedBy: "",
					UpdatedAt: 0,
					UpdatedBy: "",
					DeletedAt: 0,
					DeletedBy: "",
				}
			},
			validate: func(t *testing.T, dto *dtos.GuestEventRequestDTO) {
				assert.NotNil(t, dto)
				assert.Equal(t, "", dto.ID)
				assert.Equal(t, "", dto.Name)
				assert.Equal(t, "", dto.Address)
				assert.Equal(t, int64(0), dto.CreatedAt)
				assert.Equal(t, "", dto.CreatedBy)
				assert.Equal(t, int64(0), dto.UpdatedAt)
				assert.Equal(t, "", dto.UpdatedBy)
				assert.Equal(t, int64(0), dto.DeletedAt)
				assert.Equal(t, "", dto.DeletedBy)
			},
		},
		{
			name: "should_convert_vm_with_updated_fields_only",
			setupVM: func() *GuestEventRequestVM {
				return &GuestEventRequestVM{
					ID:        "guest-789",
					Name:      "Bob Wilson",
					Address:   "456 Oak Ave",
					CreatedAt: 1700002000,
					CreatedBy: "system",
					UpdatedAt: 1700002500,
					UpdatedBy: "moderator",
				}
			},
			validate: func(t *testing.T, dto *dtos.GuestEventRequestDTO) {
				assert.NotNil(t, dto)
				assert.Equal(t, "guest-789", dto.ID)
				assert.Equal(t, "Bob Wilson", dto.Name)
				assert.Equal(t, "456 Oak Ave", dto.Address)
				assert.Equal(t, int64(1700002000), dto.CreatedAt)
				assert.Equal(t, "system", dto.CreatedBy)
				assert.Equal(t, int64(1700002500), dto.UpdatedAt)
				assert.Equal(t, "moderator", dto.UpdatedBy)
				assert.Equal(t, int64(0), dto.DeletedAt)
				assert.Equal(t, "", dto.DeletedBy)
			},
		},
		{
			name: "should_convert_vm_with_deleted_fields",
			setupVM: func() *GuestEventRequestVM {
				return &GuestEventRequestVM{
					ID:        "guest-999",
					Name:      "Charlie Brown",
					Address:   "789 Pine Rd",
					CreatedAt: 1700003000,
					CreatedBy: "admin",
					UpdatedAt: 1700003100,
					UpdatedBy: "admin",
					DeletedAt: 1700003200,
					DeletedBy: "admin",
				}
			},
			validate: func(t *testing.T, dto *dtos.GuestEventRequestDTO) {
				assert.NotNil(t, dto)
				assert.Equal(t, "guest-999", dto.ID)
				assert.Equal(t, "Charlie Brown", dto.Name)
				assert.Equal(t, "789 Pine Rd", dto.Address)
				assert.Equal(t, int64(1700003000), dto.CreatedAt)
				assert.Equal(t, "admin", dto.CreatedBy)
				assert.Equal(t, int64(1700003100), dto.UpdatedAt)
				assert.Equal(t, "admin", dto.UpdatedBy)
				assert.Equal(t, int64(1700003200), dto.DeletedAt)
				assert.Equal(t, "admin", dto.DeletedBy)
			},
		},
		{
			name: "should_handle_special_characters_in_fields",
			setupVM: func() *GuestEventRequestVM {
				return &GuestEventRequestVM{
					ID:        "guest-special-!@#$%",
					Name:      "O'Brien & Sons",
					Address:   "123 Main St, Apt #456",
					CreatedAt: 1700004000,
					CreatedBy: "user@example.com",
					UpdatedAt: 1700004100,
					UpdatedBy: "admin@example.com",
				}
			},
			validate: func(t *testing.T, dto *dtos.GuestEventRequestDTO) {
				assert.NotNil(t, dto)
				assert.Equal(t, "guest-special-!@#$%", dto.ID)
				assert.Equal(t, "O'Brien & Sons", dto.Name)
				assert.Equal(t, "123 Main St, Apt #456", dto.Address)
				assert.Equal(t, int64(1700004000), dto.CreatedAt)
				assert.Equal(t, "user@example.com", dto.CreatedBy)
				assert.Equal(t, int64(1700004100), dto.UpdatedAt)
				assert.Equal(t, "admin@example.com", dto.UpdatedBy)
			},
		},
		{
			name: "should_handle_long_strings",
			setupVM: func() *GuestEventRequestVM {
				longString := "This is a very long string that contains many characters to test the conversion of long field values in the view model to data transfer object"
				return &GuestEventRequestVM{
					ID:        "guest-long-123",
					Name:      longString,
					Address:   longString,
					CreatedAt: 1700005000,
					CreatedBy: longString,
					UpdatedAt: 1700005100,
					UpdatedBy: longString,
				}
			},
			validate: func(t *testing.T, dto *dtos.GuestEventRequestDTO) {
				longString := "This is a very long string that contains many characters to test the conversion of long field values in the view model to data transfer object"
				assert.NotNil(t, dto)
				assert.Equal(t, "guest-long-123", dto.ID)
				assert.Equal(t, longString, dto.Name)
				assert.Equal(t, longString, dto.Address)
				assert.Equal(t, int64(1700005000), dto.CreatedAt)
				assert.Equal(t, longString, dto.CreatedBy)
				assert.Equal(t, int64(1700005100), dto.UpdatedAt)
				assert.Equal(t, longString, dto.UpdatedBy)
			},
		},
		{
			name: "should_handle_negative_timestamps",
			setupVM: func() *GuestEventRequestVM {
				return &GuestEventRequestVM{
					ID:        "guest-negative",
					Name:      "Negative Test",
					Address:   "Test Address",
					CreatedAt: -1,
					CreatedBy: "tester",
					UpdatedAt: -100,
					UpdatedBy: "tester",
					DeletedAt: -200,
					DeletedBy: "tester",
				}
			},
			validate: func(t *testing.T, dto *dtos.GuestEventRequestDTO) {
				assert.NotNil(t, dto)
				assert.Equal(t, "guest-negative", dto.ID)
				assert.Equal(t, "Negative Test", dto.Name)
				assert.Equal(t, "Test Address", dto.Address)
				assert.Equal(t, int64(-1), dto.CreatedAt)
				assert.Equal(t, "tester", dto.CreatedBy)
				assert.Equal(t, int64(-100), dto.UpdatedAt)
				assert.Equal(t, "tester", dto.UpdatedBy)
				assert.Equal(t, int64(-200), dto.DeletedAt)
				assert.Equal(t, "tester", dto.DeletedBy)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := tt.setupVM()

			dto := vm.ToDTO()

			tt.validate(t, dto)
		})
	}
}
