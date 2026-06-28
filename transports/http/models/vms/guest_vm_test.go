package vms

import (
	"go-boilerplate/internal/models/dtos"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CreateGuestRequestVM_ToDTO(t *testing.T) {
	type fields struct {
		Name    string
		Address string
	}
	type args struct {
		createdBy string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *dtos.CreateGuestRequestDTO
	}{
		{
			name: "success - convert VM to DTO with all fields",
			fields: fields{
				Name:    "John Snow",
				Address: "123 Main Street, Apt. 4B, New York, NY 10001, USA",
			},
			args: args{
				createdBy: "Daenerys",
			},
			want: &dtos.CreateGuestRequestDTO{
				Name:      "John Snow",
				Address:   "123 Main Street, Apt. 4B, New York, NY 10001, USA",
				CreatedBy: "Daenerys",
			},
		},
		{
			name: "success - convert VM to DTO with empty address",
			fields: fields{
				Name:    "John Snow",
				Address: "",
			},
			args: args{
				createdBy: "Daenerys",
			},
			want: &dtos.CreateGuestRequestDTO{
				Name:      "John Snow",
				Address:   "",
				CreatedBy: "Daenerys",
			},
		},
		{
			name: "success - convert VM to DTO with empty name",
			fields: fields{
				Name:    "",
				Address: "123 Main Street, Apt. 4B, New York, NY 10001, USA",
			},
			args: args{
				createdBy: "Daenerys",
			},
			want: &dtos.CreateGuestRequestDTO{
				Name:      "",
				Address:   "123 Main Street, Apt. 4B, New York, NY 10001, USA",
				CreatedBy: "Daenerys",
			},
		},
		{
			name: "success - convert VM to DTO with all empty fields",
			fields: fields{
				Name:    "",
				Address: "",
			},
			args: args{
				createdBy: "",
			},
			want: &dtos.CreateGuestRequestDTO{
				Name:      "",
				Address:   "",
				CreatedBy: "",
			},
		},
		{
			name: "success - convert VM to DTO with special characters",
			fields: fields{
				Name:    "O'Brien & Sons <Test>",
				Address: "123 \"Main\" Street, #4B",
			},
			args: args{
				createdBy: "Admin@System",
			},
			want: &dtos.CreateGuestRequestDTO{
				Name:      "O'Brien & Sons <Test>",
				Address:   "123 \"Main\" Street, #4B",
				CreatedBy: "Admin@System",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &CreateGuestRequestVM{
				Name:    tt.fields.Name,
				Address: tt.fields.Address,
			}
			got := vm.ToDTO(tt.args.createdBy)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_DeleteGuestByIDRequestVM_ToDTO(t *testing.T) {
	type fields struct {
		ID string
	}
	type args struct {
		deletedBy string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *dtos.DeleteGuestByIDRequestDTO
	}{
		{
			name: "success - convert VM to DTO with valid ID and deletedBy",
			fields: fields{
				ID: "01932293-d710-7f55-a9f6-66e6248ae72f",
			},
			args: args{
				deletedBy: "Daenerys",
			},
			want: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "01932293-d710-7f55-a9f6-66e6248ae72f",
				DeletedBy: "Daenerys",
			},
		},
		{
			name: "success - convert VM to DTO with empty ID",
			fields: fields{
				ID: "",
			},
			args: args{
				deletedBy: "Daenerys",
			},
			want: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "",
				DeletedBy: "Daenerys",
			},
		},
		{
			name: "success - convert VM to DTO with empty deletedBy",
			fields: fields{
				ID: "01932293-d710-7f55-a9f6-66e6248ae72f",
			},
			args: args{
				deletedBy: "",
			},
			want: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "01932293-d710-7f55-a9f6-66e6248ae72f",
				DeletedBy: "",
			},
		},
		{
			name: "success - convert VM to DTO with all empty fields",
			fields: fields{
				ID: "",
			},
			args: args{
				deletedBy: "",
			},
			want: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "",
				DeletedBy: "",
			},
		},
		{
			name: "success - convert VM to DTO with special characters in deletedBy",
			fields: fields{
				ID: "test-id-123",
			},
			args: args{
				deletedBy: "Admin@System.User",
			},
			want: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "test-id-123",
				DeletedBy: "Admin@System.User",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &DeleteGuestByIDRequestVM{
				ID: tt.fields.ID,
			}
			got := vm.ToDTO(tt.args.deletedBy)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_FindAllGuestRequestVM_ToDTO(t *testing.T) {
	type fields struct {
		Keyword string
		Sorts   string
		Take    uint64
		Skip    uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   *dtos.FindAllGuestRequestDTO
	}{
		{
			name: "success - convert VM to DTO with all fields",
			fields: fields{
				Keyword: "John",
				Sorts:   "name:asc,created_at:desc",
				Take:    20,
				Skip:    10,
			},
			want: &dtos.FindAllGuestRequestDTO{
				Keyword: "John",
				Sorts:   "name:asc,created_at:desc",
				Take:    20,
				Skip:    10,
			},
		},
		{
			name: "success - convert VM to DTO with default Take when Take is 0",
			fields: fields{
				Keyword: "Snow",
				Sorts:   "name:asc",
				Take:    0,
				Skip:    5,
			},
			want: &dtos.FindAllGuestRequestDTO{
				Keyword: "Snow",
				Sorts:   "name:asc",
				Take:    10,
				Skip:    5,
			},
		},
		{
			name: "success - convert VM to DTO with no Skip when Skip is 0",
			fields: fields{
				Keyword: "Guest",
				Sorts:   "created_at:desc",
				Take:    15,
				Skip:    0,
			},
			want: &dtos.FindAllGuestRequestDTO{
				Keyword: "Guest",
				Sorts:   "created_at:desc",
				Take:    15,
				Skip:    0,
			},
		},
		{
			name: "success - convert VM to DTO with empty Sorts",
			fields: fields{
				Keyword: "Test",
				Sorts:   "",
				Take:    25,
				Skip:    20,
			},
			want: &dtos.FindAllGuestRequestDTO{
				Keyword: "Test",
				Sorts:   "",
				Take:    25,
				Skip:    20,
			},
		},
		{
			name: "success - convert VM to DTO with only Keyword",
			fields: fields{
				Keyword: "Search",
				Sorts:   "",
				Take:    0,
				Skip:    0,
			},
			want: &dtos.FindAllGuestRequestDTO{
				Keyword: "Search",
				Sorts:   "",
				Take:    10,
				Skip:    0,
			},
		},
		{
			name: "success - convert VM to DTO with all empty/zero fields",
			fields: fields{
				Keyword: "",
				Sorts:   "",
				Take:    0,
				Skip:    0,
			},
			want: &dtos.FindAllGuestRequestDTO{
				Keyword: "",
				Sorts:   "",
				Take:    10,
				Skip:    0,
			},
		},
		{
			name: "success - convert VM to DTO with special characters in Keyword",
			fields: fields{
				Keyword: "O'Brien & Sons <Test>",
				Sorts:   "name:asc",
				Take:    30,
				Skip:    15,
			},
			want: &dtos.FindAllGuestRequestDTO{
				Keyword: "O'Brien & Sons <Test>",
				Sorts:   "name:asc",
				Take:    30,
				Skip:    15,
			},
		},
		{
			name: "success - convert VM to DTO with large values",
			fields: fields{
				Keyword: "Test",
				Sorts:   "id:desc",
				Take:    1000,
				Skip:    5000,
			},
			want: &dtos.FindAllGuestRequestDTO{
				Keyword: "Test",
				Sorts:   "id:desc",
				Take:    1000,
				Skip:    5000,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &FindAllGuestRequestVM{
				Keyword: tt.fields.Keyword,
				Sorts:   tt.fields.Sorts,
				Take:    tt.fields.Take,
				Skip:    tt.fields.Skip,
			}
			got := vm.ToDTO()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewFindAllGuestResponseVM(t *testing.T) {
	type args struct {
		dto *dtos.FindAllGuestResponseDTO
	}
	tests := []struct {
		name string
		args args
		want *FindAllGuestResponseVM
	}{
		{
			name: "success - convert DTO to VM with multiple items",
			args: args{
				dto: &dtos.FindAllGuestResponseDTO{
					List: []dtos.GuestResponseDTO{
						{
							ID:        "01932293-d710-7f55-a9f6-66e6248ae72f",
							Name:      "John Snow",
							Address:   "123 Main Street, Apt. 4B, New York, NY 10001, USA",
							CreatedAt: 1731452061534,
							CreatedBy: "Daenerys",
							UpdatedAt: 1731452061534,
							UpdatedBy: "Daenerys",
						},
						{
							ID:        "01932293-d710-7f55-a9f6-66e6248ae72e",
							Name:      "Arya Stark",
							Address:   "Winterfell",
							CreatedAt: 1731452061535,
							CreatedBy: "Ned",
							UpdatedAt: 1731452061536,
							UpdatedBy: "Ned",
						},
					},
					Count: 2,
				},
			},
			want: &FindAllGuestResponseVM{
				List: []GuestResponseVM{
					{
						ID:        "01932293-d710-7f55-a9f6-66e6248ae72f",
						Name:      "John Snow",
						Address:   "123 Main Street, Apt. 4B, New York, NY 10001, USA",
						CreatedAt: 1731452061534,
						CreatedBy: "Daenerys",
						UpdatedAt: 1731452061534,
						UpdatedBy: "Daenerys",
					},
					{
						ID:        "01932293-d710-7f55-a9f6-66e6248ae72e",
						Name:      "Arya Stark",
						Address:   "Winterfell",
						CreatedAt: 1731452061535,
						CreatedBy: "Ned",
						UpdatedAt: 1731452061536,
						UpdatedBy: "Ned",
					},
				},
				Count: 2,
			},
		},
		{
			name: "success - convert DTO to VM with single item",
			args: args{
				dto: &dtos.FindAllGuestResponseDTO{
					List: []dtos.GuestResponseDTO{
						{
							ID:        "test-id-001",
							Name:      "Test User",
							Address:   "Test Address",
							CreatedAt: 1234567890,
							CreatedBy: "Admin",
							UpdatedAt: 1234567891,
							UpdatedBy: "Admin",
						},
					},
					Count: 1,
				},
			},
			want: &FindAllGuestResponseVM{
				List: []GuestResponseVM{
					{
						ID:        "test-id-001",
						Name:      "Test User",
						Address:   "Test Address",
						CreatedAt: 1234567890,
						CreatedBy: "Admin",
						UpdatedAt: 1234567891,
						UpdatedBy: "Admin",
					},
				},
				Count: 1,
			},
		},
		{
			name: "success - convert DTO to VM with empty list",
			args: args{
				dto: &dtos.FindAllGuestResponseDTO{
					List:  []dtos.GuestResponseDTO{},
					Count: 0,
				},
			},
			want: &FindAllGuestResponseVM{
				List:  nil,
				Count: 0,
			},
		},
		{
			name: "success - convert DTO to VM with nil list",
			args: args{
				dto: &dtos.FindAllGuestResponseDTO{
					List:  nil,
					Count: 0,
				},
			},
			want: &FindAllGuestResponseVM{
				List:  nil,
				Count: 0,
			},
		},
		{
			name: "success - convert DTO to VM with large count",
			args: args{
				dto: &dtos.FindAllGuestResponseDTO{
					List: []dtos.GuestResponseDTO{
						{
							ID:        "id-1",
							Name:      "Guest 1",
							Address:   "Address 1",
							CreatedAt: 1000000001,
							CreatedBy: "System",
							UpdatedAt: 1000000002,
							UpdatedBy: "System",
						},
						{
							ID:        "id-2",
							Name:      "Guest 2",
							Address:   "Address 2",
							CreatedAt: 1000000003,
							CreatedBy: "System",
							UpdatedAt: 1000000004,
							UpdatedBy: "System",
						},
						{
							ID:        "id-3",
							Name:      "Guest 3",
							Address:   "Address 3",
							CreatedAt: 1000000005,
							CreatedBy: "System",
							UpdatedAt: 1000000006,
							UpdatedBy: "System",
						},
					},
					Count: 1000,
				},
			},
			want: &FindAllGuestResponseVM{
				List: []GuestResponseVM{
					{
						ID:        "id-1",
						Name:      "Guest 1",
						Address:   "Address 1",
						CreatedAt: 1000000001,
						CreatedBy: "System",
						UpdatedAt: 1000000002,
						UpdatedBy: "System",
					},
					{
						ID:        "id-2",
						Name:      "Guest 2",
						Address:   "Address 2",
						CreatedAt: 1000000003,
						CreatedBy: "System",
						UpdatedAt: 1000000004,
						UpdatedBy: "System",
					},
					{
						ID:        "id-3",
						Name:      "Guest 3",
						Address:   "Address 3",
						CreatedAt: 1000000005,
						CreatedBy: "System",
						UpdatedAt: 1000000006,
						UpdatedBy: "System",
					},
				},
				Count: 1000,
			},
		},
		{
			name: "success - convert DTO to VM with special characters",
			args: args{
				dto: &dtos.FindAllGuestResponseDTO{
					List: []dtos.GuestResponseDTO{
						{
							ID:        "special-id-<>",
							Name:      "O'Brien & Sons",
							Address:   "123 \"Main\" Street #4B",
							CreatedAt: 9999999999,
							CreatedBy: "Admin@System",
							UpdatedAt: 9999999999,
							UpdatedBy: "Admin@System",
						},
					},
					Count: 1,
				},
			},
			want: &FindAllGuestResponseVM{
				List: []GuestResponseVM{
					{
						ID:        "special-id-<>",
						Name:      "O'Brien & Sons",
						Address:   "123 \"Main\" Street #4B",
						CreatedAt: 9999999999,
						CreatedBy: "Admin@System",
						UpdatedAt: 9999999999,
						UpdatedBy: "Admin@System",
					},
				},
				Count: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFindAllGuestResponseVM(tt.args.dto)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_FindGuestByIDRequestVM_ToDTO(t *testing.T) {
	type fields struct {
		ID string
	}
	tests := []struct {
		name   string
		fields fields
		want   *dtos.FindGuestByIDRequestDTO
	}{
		{
			name: "success - convert VM to DTO with valid UUID",
			fields: fields{
				ID: "01932293-d710-7f55-a9f6-66e6248ae72f",
			},
			want: &dtos.FindGuestByIDRequestDTO{
				ID: "01932293-d710-7f55-a9f6-66e6248ae72f",
			},
		},
		{
			name: "success - convert VM to DTO with empty ID",
			fields: fields{
				ID: "",
			},
			want: &dtos.FindGuestByIDRequestDTO{
				ID: "",
			},
		},
		{
			name: "success - convert VM to DTO with numeric ID",
			fields: fields{
				ID: "12345",
			},
			want: &dtos.FindGuestByIDRequestDTO{
				ID: "12345",
			},
		},
		{
			name: "success - convert VM to DTO with alphanumeric ID",
			fields: fields{
				ID: "guest-abc-123",
			},
			want: &dtos.FindGuestByIDRequestDTO{
				ID: "guest-abc-123",
			},
		},
		{
			name: "success - convert VM to DTO with special characters in ID",
			fields: fields{
				ID: "id_with-special.chars@123",
			},
			want: &dtos.FindGuestByIDRequestDTO{
				ID: "id_with-special.chars@123",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &FindGuestByIDRequestVM{
				ID: tt.fields.ID,
			}
			got := vm.ToDTO()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewGuestResponseVM(t *testing.T) {
	type args struct {
		dto *dtos.GuestResponseDTO
	}
	tests := []struct {
		name string
		args args
		want *GuestResponseVM
	}{
		{
			name: "success - convert DTO to VM with all fields",
			args: args{
				dto: &dtos.GuestResponseDTO{
					ID:        "01932293-d710-7f55-a9f6-66e6248ae72f",
					Name:      "John Snow",
					Address:   "123 Main Street, Apt. 4B, New York, NY 10001, USA",
					CreatedAt: 1731452061534,
					CreatedBy: "Daenerys",
					UpdatedAt: 1731452061534,
					UpdatedBy: "Daenerys",
				},
			},
			want: &GuestResponseVM{
				ID:        "01932293-d710-7f55-a9f6-66e6248ae72f",
				Name:      "John Snow",
				Address:   "123 Main Street, Apt. 4B, New York, NY 10001, USA",
				CreatedAt: 1731452061534,
				CreatedBy: "Daenerys",
				UpdatedAt: 1731452061534,
				UpdatedBy: "Daenerys",
			},
		},
		{
			name: "success - convert DTO to VM with empty optional fields",
			args: args{
				dto: &dtos.GuestResponseDTO{
					ID:        "test-id-001",
					Name:      "Test User",
					Address:   "",
					CreatedAt: 1234567890,
					CreatedBy: "Admin",
					UpdatedAt: 0,
					UpdatedBy: "",
				},
			},
			want: &GuestResponseVM{
				ID:        "test-id-001",
				Name:      "Test User",
				Address:   "",
				CreatedAt: 1234567890,
				CreatedBy: "Admin",
				UpdatedAt: 0,
				UpdatedBy: "",
			},
		},
		{
			name: "success - convert DTO to VM with minimal fields",
			args: args{
				dto: &dtos.GuestResponseDTO{
					ID:        "minimal-id",
					Name:      "Minimal User",
					Address:   "",
					CreatedAt: 1000000000,
					CreatedBy: "System",
					UpdatedAt: 0,
					UpdatedBy: "",
				},
			},
			want: &GuestResponseVM{
				ID:        "minimal-id",
				Name:      "Minimal User",
				Address:   "",
				CreatedAt: 1000000000,
				CreatedBy: "System",
				UpdatedAt: 0,
				UpdatedBy: "",
			},
		},
		{
			name: "success - convert DTO to VM with special characters",
			args: args{
				dto: &dtos.GuestResponseDTO{
					ID:        "special-id-<>",
					Name:      "O'Brien & Sons <Test>",
					Address:   "123 \"Main\" Street #4B",
					CreatedAt: 9999999999,
					CreatedBy: "Admin@System",
					UpdatedAt: 9999999999,
					UpdatedBy: "Admin@System.Updated",
				},
			},
			want: &GuestResponseVM{
				ID:        "special-id-<>",
				Name:      "O'Brien & Sons <Test>",
				Address:   "123 \"Main\" Street #4B",
				CreatedAt: 9999999999,
				CreatedBy: "Admin@System",
				UpdatedAt: 9999999999,
				UpdatedBy: "Admin@System.Updated",
			},
		},
		{
			name: "success - convert DTO to VM with long text fields",
			args: args{
				dto: &dtos.GuestResponseDTO{
					ID:        "long-text-id-123456789",
					Name:      "Very Long Name With Multiple Words And Spaces",
					Address:   "Building 1, Floor 2, Unit 3, Street Name Avenue, District Area, City Name, State Province, Country Name, Postal Code 12345-6789",
					CreatedAt: 1700000000000,
					CreatedBy: "Administrator User",
					UpdatedAt: 1700000001000,
					UpdatedBy: "System Administrator",
				},
			},
			want: &GuestResponseVM{
				ID:        "long-text-id-123456789",
				Name:      "Very Long Name With Multiple Words And Spaces",
				Address:   "Building 1, Floor 2, Unit 3, Street Name Avenue, District Area, City Name, State Province, Country Name, Postal Code 12345-6789",
				CreatedAt: 1700000000000,
				CreatedBy: "Administrator User",
				UpdatedAt: 1700000001000,
				UpdatedBy: "System Administrator",
			},
		},
		{
			name: "success - convert DTO to VM with zero timestamps",
			args: args{
				dto: &dtos.GuestResponseDTO{
					ID:        "zero-timestamp-id",
					Name:      "Zero Timestamp User",
					Address:   "Zero Address",
					CreatedAt: 0,
					CreatedBy: "Creator",
					UpdatedAt: 0,
					UpdatedBy: "",
				},
			},
			want: &GuestResponseVM{
				ID:        "zero-timestamp-id",
				Name:      "Zero Timestamp User",
				Address:   "Zero Address",
				CreatedAt: 0,
				CreatedBy: "Creator",
				UpdatedAt: 0,
				UpdatedBy: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGuestResponseVM(tt.args.dto)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_UpdateGuestByIDRequestVM_ToDTO(t *testing.T) {
	type fields struct {
		ID      string
		Name    string
		Address string
	}
	type args struct {
		updatedBy string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *dtos.UpdateGuestByIDRequestDTO
	}{
		{
			name: "success - convert VM to DTO with all fields",
			fields: fields{
				ID:      "01932293-d710-7f55-a9f6-66e6248ae72f",
				Name:    "John Snow Updated",
				Address: "456 New Street, Apt. 5C, New York, NY 10002, USA",
			},
			args: args{
				updatedBy: "Daenerys",
			},
			want: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "01932293-d710-7f55-a9f6-66e6248ae72f",
				Name:      "John Snow Updated",
				Address:   "456 New Street, Apt. 5C, New York, NY 10002, USA",
				UpdatedBy: "Daenerys",
			},
		},
		{
			name: "success - convert VM to DTO with empty address",
			fields: fields{
				ID:      "test-id-001",
				Name:    "Test User",
				Address: "",
			},
			args: args{
				updatedBy: "Admin",
			},
			want: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "test-id-001",
				Name:      "Test User",
				Address:   "",
				UpdatedBy: "Admin",
			},
		},
		{
			name: "success - convert VM to DTO with empty name",
			fields: fields{
				ID:      "guest-123",
				Name:    "",
				Address: "Some Address",
			},
			args: args{
				updatedBy: "System",
			},
			want: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "guest-123",
				Name:      "",
				Address:   "Some Address",
				UpdatedBy: "System",
			},
		},
		{
			name: "success - convert VM to DTO with empty updatedBy",
			fields: fields{
				ID:      "id-456",
				Name:    "Guest Name",
				Address: "Guest Address",
			},
			args: args{
				updatedBy: "",
			},
			want: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "id-456",
				Name:      "Guest Name",
				Address:   "Guest Address",
				UpdatedBy: "",
			},
		},
		{
			name: "success - convert VM to DTO with all empty fields except ID",
			fields: fields{
				ID:      "only-id-789",
				Name:    "",
				Address: "",
			},
			args: args{
				updatedBy: "",
			},
			want: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "only-id-789",
				Name:      "",
				Address:   "",
				UpdatedBy: "",
			},
		},
		{
			name: "success - convert VM to DTO with special characters",
			fields: fields{
				ID:      "special-id-<>",
				Name:    "O'Brien & Sons <Updated>",
				Address: "789 \"New\" Street, #10A",
			},
			args: args{
				updatedBy: "Admin@System.User",
			},
			want: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "special-id-<>",
				Name:      "O'Brien & Sons <Updated>",
				Address:   "789 \"New\" Street, #10A",
				UpdatedBy: "Admin@System.User",
			},
		},
		{
			name: "success - convert VM to DTO with long text fields",
			fields: fields{
				ID:      "long-text-id-123456789",
				Name:    "Very Long Updated Name With Multiple Words And Spaces And More Text",
				Address: "Building 10, Floor 20, Unit 30, Long Street Name Avenue Boulevard, District Area Region, City Name Metropolitan, State Province Territory, Country Name Republic, Postal Code 98765-4321",
			},
			args: args{
				updatedBy: "System Administrator Manager",
			},
			want: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "long-text-id-123456789",
				Name:      "Very Long Updated Name With Multiple Words And Spaces And More Text",
				Address:   "Building 10, Floor 20, Unit 30, Long Street Name Avenue Boulevard, District Area Region, City Name Metropolitan, State Province Territory, Country Name Republic, Postal Code 98765-4321",
				UpdatedBy: "System Administrator Manager",
			},
		},
		{
			name: "success - convert VM to DTO with numeric-like strings",
			fields: fields{
				ID:      "12345",
				Name:    "Guest 12345",
				Address: "Address 67890",
			},
			args: args{
				updatedBy: "User123",
			},
			want: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "12345",
				Name:      "Guest 12345",
				Address:   "Address 67890",
				UpdatedBy: "User123",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &UpdateGuestByIDRequestVM{
				ID:      tt.fields.ID,
				Name:    tt.fields.Name,
				Address: tt.fields.Address,
			}
			got := vm.ToDTO(tt.args.updatedBy)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Bulk HTTP VM tests

func validUUIDStr() string {
	return "01932293-d710-7f55-a9f6-66e6248ae72f"
}

func Test_BulkCreateGuestsRequestVM_ToDTO(t *testing.T) {
	type fields struct {
		Items []CreateGuestRequestVM
	}
	type args struct {
		createdBy string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *dtos.BulkCreateGuestsRequestDTO
	}{
		{
			name: "success - convert single item VM to DTO",
			fields: fields{
				Items: []CreateGuestRequestVM{
					{Name: "John Snow", Address: "123 Main St"},
				},
			},
			args: args{createdBy: "admin"},
			want: &dtos.BulkCreateGuestsRequestDTO{
				Items: []dtos.CreateGuestRequestDTO{
					{Name: "John Snow", Address: "123 Main St", CreatedBy: "admin"},
				},
			},
		},
		{
			name: "success - convert multiple items VM to DTO",
			fields: fields{
				Items: []CreateGuestRequestVM{
					{Name: "John Snow", Address: "123 Main St"},
					{Name: "Jane Smith"},
				},
			},
			args: args{createdBy: "system"},
			want: &dtos.BulkCreateGuestsRequestDTO{
				Items: []dtos.CreateGuestRequestDTO{
					{Name: "John Snow", Address: "123 Main St", CreatedBy: "system"},
					{Name: "Jane Smith", CreatedBy: "system"},
				},
			},
		},
		{
			name: "success - convert empty items VM to DTO",
			fields: fields{
				Items: []CreateGuestRequestVM{},
			},
			args: args{createdBy: "admin"},
			want: &dtos.BulkCreateGuestsRequestDTO{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &BulkCreateGuestsRequestVM{
				Items: tt.fields.Items,
			}
			got := vm.ToDTO(tt.args.createdBy)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_NewBulkCreateGuestsResponseVM(t *testing.T) {
	type args struct {
		dto *dtos.BulkCreateGuestsResponseDTO
	}
	tests := []struct {
		name string
		args args
		want *[]GuestResponseVM
	}{
		{
			name: "success - create response VM with guests",
			args: args{
				dto: &dtos.BulkCreateGuestsResponseDTO{
					Guests: []dtos.GuestResponseDTO{
						{ID: "id-1", Name: "John", CreatedBy: "admin"},
						{ID: "id-2", Name: "Jane", CreatedBy: "system"},
					},
				},
			},
			want: &[]GuestResponseVM{
				{ID: "id-1", Name: "John", CreatedBy: "admin"},
				{ID: "id-2", Name: "Jane", CreatedBy: "system"},
			},
		},
		{
			name: "success - create response VM with empty guests",
			args: args{
				dto: &dtos.BulkCreateGuestsResponseDTO{
					Guests: []dtos.GuestResponseDTO{},
				},
			},
			want: &[]GuestResponseVM{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewBulkCreateGuestsResponseVM(tt.args.dto)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_BulkUpdateGuestItemVM_ToDTO(t *testing.T) {
	type fields struct {
		ID      string
		Name    string
		Address string
	}
	type args struct {
		updatedBy string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *dtos.UpdateGuestByIDRequestDTO
	}{
		{
			name: "success - convert VM to DTO with all fields",
			fields: fields{
				ID:      "id-1",
				Name:    "Updated Name",
				Address: "123 Main St",
			},
			args: args{updatedBy: "admin"},
			want: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "id-1",
				Name:      "Updated Name",
				Address:   "123 Main St",
				UpdatedBy: "admin",
			},
		},
		{
			name: "success - convert VM to DTO without address",
			fields: fields{
				ID:   "id-2",
				Name: "Name Only",
			},
			args: args{updatedBy: "system"},
			want: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "id-2",
				Name:      "Name Only",
				UpdatedBy: "system",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &BulkUpdateGuestItemVM{
				ID:      tt.fields.ID,
				Name:    tt.fields.Name,
				Address: tt.fields.Address,
			}
			got := vm.ToDTO(tt.args.updatedBy)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_BulkUpdateGuestsRequestVM_ToDTO(t *testing.T) {
	type fields struct {
		Items []BulkUpdateGuestItemVM
	}
	type args struct {
		updatedBy string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *dtos.BulkUpdateGuestsRequestDTO
	}{
		{
			name: "success - convert single item VM to DTO",
			fields: fields{
				Items: []BulkUpdateGuestItemVM{
					{ID: "id-1", Name: "Updated Name", Address: "123 Main St"},
				},
			},
			args: args{updatedBy: "admin"},
			want: &dtos.BulkUpdateGuestsRequestDTO{
				Items: []dtos.UpdateGuestByIDRequestDTO{
					{ID: "id-1", Name: "Updated Name", Address: "123 Main St", UpdatedBy: "admin"},
				},
			},
		},
		{
			name: "success - convert multiple items VM to DTO",
			fields: fields{
				Items: []BulkUpdateGuestItemVM{
					{ID: "id-1", Name: "Name 1"},
					{ID: "id-2", Name: "Name 2"},
				},
			},
			args: args{updatedBy: "system"},
			want: &dtos.BulkUpdateGuestsRequestDTO{
				Items: []dtos.UpdateGuestByIDRequestDTO{
					{ID: "id-1", Name: "Name 1", UpdatedBy: "system"},
					{ID: "id-2", Name: "Name 2", UpdatedBy: "system"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &BulkUpdateGuestsRequestVM{
				Items: tt.fields.Items,
			}
			got := vm.ToDTO(tt.args.updatedBy)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_NewBulkUpdateGuestsResponseVM(t *testing.T) {
	type args struct {
		dto *dtos.BulkUpdateGuestsResponseDTO
	}
	tests := []struct {
		name string
		args args
		want *[]GuestResponseVM
	}{
		{
			name: "success - create response VM with guests",
			args: args{
				dto: &dtos.BulkUpdateGuestsResponseDTO{
					Guests: []dtos.GuestResponseDTO{
						{ID: "id-1", Name: "Updated John", CreatedBy: "admin"},
					},
				},
			},
			want: &[]GuestResponseVM{
				{ID: "id-1", Name: "Updated John", CreatedBy: "admin"},
			},
		},
		{
			name: "success - create response VM with empty guests",
			args: args{
				dto: &dtos.BulkUpdateGuestsResponseDTO{
					Guests: []dtos.GuestResponseDTO{},
				},
			},
			want: &[]GuestResponseVM{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewBulkUpdateGuestsResponseVM(tt.args.dto)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_BulkDeleteGuestsRequestVM_ToDTO(t *testing.T) {
	type fields struct {
		IDs []string
	}
	tests := []struct {
		name      string
		fields    fields
		deletedBy string
		want      *dtos.BulkDeleteGuestsRequestDTO
	}{
		{
			name: "success - convert VM to DTO with multiple IDs",
			fields: fields{
				IDs: []string{"id-1", "id-2"},
			},
			deletedBy: "admin",
			want: &dtos.BulkDeleteGuestsRequestDTO{
				IDs:       []string{"id-1", "id-2"},
				DeletedBy: "admin",
			},
		},
		{
			name: "success - convert VM to DTO with single ID",
			fields: fields{
				IDs: []string{"id-1"},
			},
			deletedBy: "system",
			want: &dtos.BulkDeleteGuestsRequestDTO{
				IDs:       []string{"id-1"},
				DeletedBy: "system",
			},
		},
		{
			name: "success - convert VM to DTO with empty IDs",
			fields: fields{
				IDs: []string{},
			},
			deletedBy: "admin",
			want: &dtos.BulkDeleteGuestsRequestDTO{
				IDs:       []string{},
				DeletedBy: "admin",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := &BulkDeleteGuestsRequestVM{
				IDs: tt.fields.IDs,
			}
			got := vm.ToDTO(tt.deletedBy)
			assert.Equal(t, tt.want, got)
		})
	}
}
