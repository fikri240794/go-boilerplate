package vms

import (
	"go-boilerplate/internal/models/dtos"
	"go-boilerplate/pkg/protobuf_boilerplate"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateGuestRequestVMToDTO(t *testing.T) {
	tests := []struct {
		name      string
		setupVM   func(t *testing.T) *protobuf_boilerplate.CreateGuestRequestVM
		createdBy string
		validate  func(t *testing.T, dto *dtos.CreateGuestRequestDTO, vm *protobuf_boilerplate.CreateGuestRequestVM, createdBy string)
	}{
		{
			name: "should_convert_vm_to_dto_successfully",
			setupVM: func(t *testing.T) *protobuf_boilerplate.CreateGuestRequestVM {
				return &protobuf_boilerplate.CreateGuestRequestVM{
					Name:    "John Doe",
					Address: "123 Main St",
				}
			},
			createdBy: "user123",
			validate: func(t *testing.T, dto *dtos.CreateGuestRequestDTO, vm *protobuf_boilerplate.CreateGuestRequestVM, createdBy string) {
				assert.NotNil(t, dto)
				assert.Equal(t, vm.GetName(), dto.Name)
				assert.Equal(t, vm.GetAddress(), dto.Address)
				assert.Equal(t, createdBy, dto.CreatedBy)
			},
		},
		{
			name: "should_convert_vm_with_empty_address",
			setupVM: func(t *testing.T) *protobuf_boilerplate.CreateGuestRequestVM {
				return &protobuf_boilerplate.CreateGuestRequestVM{
					Name:    "Jane Smith",
					Address: "",
				}
			},
			createdBy: "admin",
			validate: func(t *testing.T, dto *dtos.CreateGuestRequestDTO, vm *protobuf_boilerplate.CreateGuestRequestVM, createdBy string) {
				assert.NotNil(t, dto)
				assert.Equal(t, "Jane Smith", dto.Name)
				assert.Equal(t, "", dto.Address)
				assert.Equal(t, "admin", dto.CreatedBy)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := tt.setupVM(t)
			dto := CreateGuestRequestVMToDTO(vm, tt.createdBy)
			tt.validate(t, dto, vm, tt.createdBy)
		})
	}
}

func TestDeleteGuestByIDRequestVMToDTO(t *testing.T) {
	tests := []struct {
		name      string
		setupVM   func(t *testing.T) *protobuf_boilerplate.DeleteGuestByIDRequestVM
		deletedBy string
		validate  func(t *testing.T, dto *dtos.DeleteGuestByIDRequestDTO, vm *protobuf_boilerplate.DeleteGuestByIDRequestVM, deletedBy string)
	}{
		{
			name: "should_convert_delete_vm_to_dto_successfully",
			setupVM: func(t *testing.T) *protobuf_boilerplate.DeleteGuestByIDRequestVM {
				return &protobuf_boilerplate.DeleteGuestByIDRequestVM{
					Id: "550e8400-e29b-41d4-a716-446655440000",
				}
			},
			deletedBy: "admin",
			validate: func(t *testing.T, dto *dtos.DeleteGuestByIDRequestDTO, vm *protobuf_boilerplate.DeleteGuestByIDRequestVM, deletedBy string) {
				assert.NotNil(t, dto)
				assert.Equal(t, vm.GetId(), dto.ID)
				assert.Equal(t, deletedBy, dto.DeletedBy)
			},
		},
		{
			name: "should_convert_delete_vm_with_different_user",
			setupVM: func(t *testing.T) *protobuf_boilerplate.DeleteGuestByIDRequestVM {
				return &protobuf_boilerplate.DeleteGuestByIDRequestVM{
					Id: "123e4567-e89b-12d3-a456-426614174000",
				}
			},
			deletedBy: "user456",
			validate: func(t *testing.T, dto *dtos.DeleteGuestByIDRequestDTO, vm *protobuf_boilerplate.DeleteGuestByIDRequestVM, deletedBy string) {
				assert.NotNil(t, dto)
				assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", dto.ID)
				assert.Equal(t, "user456", dto.DeletedBy)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := tt.setupVM(t)
			dto := DeleteGuestByIDRequestVMToDTO(vm, tt.deletedBy)
			tt.validate(t, dto, vm, tt.deletedBy)
		})
	}
}

func TestFindAllGuestRequestVMToDTO(t *testing.T) {
	tests := []struct {
		name     string
		setupVM  func(t *testing.T) *protobuf_boilerplate.FindAllGuestRequestVM
		validate func(t *testing.T, dto *dtos.FindAllGuestRequestDTO, vm *protobuf_boilerplate.FindAllGuestRequestVM)
	}{
		{
			name: "should_convert_findall_vm_with_all_fields",
			setupVM: func(t *testing.T) *protobuf_boilerplate.FindAllGuestRequestVM {
				return &protobuf_boilerplate.FindAllGuestRequestVM{
					Keyword: "test",
					Sorts:   "name.asc",
					Take:    20,
					Skip:    10,
				}
			},
			validate: func(t *testing.T, dto *dtos.FindAllGuestRequestDTO, vm *protobuf_boilerplate.FindAllGuestRequestVM) {
				assert.NotNil(t, dto)
				assert.Equal(t, "test", dto.Keyword)
				assert.Equal(t, "name.asc", dto.Sorts)
				assert.Equal(t, uint64(20), dto.Take)
				assert.Equal(t, uint64(10), dto.Skip)
			},
		},
		{
			name: "should_convert_findall_vm_with_defaults",
			setupVM: func(t *testing.T) *protobuf_boilerplate.FindAllGuestRequestVM {
				return &protobuf_boilerplate.FindAllGuestRequestVM{
					Keyword: "",
					Sorts:   "",
					Take:    0,
					Skip:    0,
				}
			},
			validate: func(t *testing.T, dto *dtos.FindAllGuestRequestDTO, vm *protobuf_boilerplate.FindAllGuestRequestVM) {
				assert.NotNil(t, dto)
				assert.Equal(t, "", dto.Keyword)
				assert.Equal(t, uint64(10), dto.Take)
				assert.Equal(t, uint64(0), dto.Skip)
			},
		},
		{
			name: "should_convert_findall_vm_with_keyword_only",
			setupVM: func(t *testing.T) *protobuf_boilerplate.FindAllGuestRequestVM {
				return &protobuf_boilerplate.FindAllGuestRequestVM{
					Keyword: "john",
					Sorts:   "",
					Take:    0,
					Skip:    0,
				}
			},
			validate: func(t *testing.T, dto *dtos.FindAllGuestRequestDTO, vm *protobuf_boilerplate.FindAllGuestRequestVM) {
				assert.NotNil(t, dto)
				assert.Equal(t, "john", dto.Keyword)
				assert.Equal(t, uint64(10), dto.Take)
			},
		},
		{
			name: "should_convert_findall_vm_with_pagination",
			setupVM: func(t *testing.T) *protobuf_boilerplate.FindAllGuestRequestVM {
				return &protobuf_boilerplate.FindAllGuestRequestVM{
					Keyword: "",
					Sorts:   "address.desc",
					Take:    50,
					Skip:    100,
				}
			},
			validate: func(t *testing.T, dto *dtos.FindAllGuestRequestDTO, vm *protobuf_boilerplate.FindAllGuestRequestVM) {
				assert.NotNil(t, dto)
				assert.Equal(t, "", dto.Keyword)
				assert.Equal(t, "address.desc", dto.Sorts)
				assert.Equal(t, uint64(50), dto.Take)
				assert.Equal(t, uint64(100), dto.Skip)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := tt.setupVM(t)
			dto := FindAllGuestRequestVMToDTO(vm)
			tt.validate(t, dto, vm)
		})
	}
}

func TestNewFindAllGuestResponseVM(t *testing.T) {
	tests := []struct {
		name     string
		setupDTO func(t *testing.T) *dtos.FindAllGuestResponseDTO
		validate func(t *testing.T, vm *protobuf_boilerplate.FindAllGuestResponseVM, dto *dtos.FindAllGuestResponseDTO)
	}{
		{
			name: "should_convert_dto_to_vm_with_multiple_guests",
			setupDTO: func(t *testing.T) *dtos.FindAllGuestResponseDTO {
				return &dtos.FindAllGuestResponseDTO{
					List: []dtos.GuestResponseDTO{
						{
							ID:        "550e8400-e29b-41d4-a716-446655440000",
							Name:      "John Doe",
							Address:   "123 Main St",
							CreatedAt: 1700000000000,
							CreatedBy: "user1",
							UpdatedAt: 1700000001000,
							UpdatedBy: "user2",
						},
						{
							ID:        "123e4567-e89b-12d3-a456-426614174000",
							Name:      "Jane Smith",
							Address:   "456 Oak Ave",
							CreatedAt: 1700000002000,
							CreatedBy: "user3",
							UpdatedAt: 1700000003000,
							UpdatedBy: "user4",
						},
					},
					Count: 2,
				}
			},
			validate: func(t *testing.T, vm *protobuf_boilerplate.FindAllGuestResponseVM, dto *dtos.FindAllGuestResponseDTO) {
				assert.NotNil(t, vm)
				assert.Equal(t, dto.Count, vm.Count)
				assert.Equal(t, len(dto.List), len(vm.List))
				assert.Equal(t, dto.List[0].ID, vm.List[0].Id)
				assert.Equal(t, dto.List[0].Name, vm.List[0].Name)
				assert.Equal(t, dto.List[1].ID, vm.List[1].Id)
				assert.Equal(t, dto.List[1].Name, vm.List[1].Name)
			},
		},
		{
			name: "should_convert_dto_to_vm_with_empty_list",
			setupDTO: func(t *testing.T) *dtos.FindAllGuestResponseDTO {
				return &dtos.FindAllGuestResponseDTO{
					List:  []dtos.GuestResponseDTO{},
					Count: 0,
				}
			},
			validate: func(t *testing.T, vm *protobuf_boilerplate.FindAllGuestResponseVM, dto *dtos.FindAllGuestResponseDTO) {
				assert.NotNil(t, vm)
				assert.Equal(t, uint64(0), vm.Count)
				assert.Equal(t, 0, len(vm.List))
			},
		},
		{
			name: "should_convert_dto_to_vm_with_single_guest",
			setupDTO: func(t *testing.T) *dtos.FindAllGuestResponseDTO {
				return &dtos.FindAllGuestResponseDTO{
					List: []dtos.GuestResponseDTO{
						{
							ID:        "abc12345-e89b-12d3-a456-426614174999",
							Name:      "Test User",
							Address:   "Test Address",
							CreatedAt: 1700000000000,
							CreatedBy: "admin",
							UpdatedAt: 0,
							UpdatedBy: "",
						},
					},
					Count: 1,
				}
			},
			validate: func(t *testing.T, vm *protobuf_boilerplate.FindAllGuestResponseVM, dto *dtos.FindAllGuestResponseDTO) {
				assert.NotNil(t, vm)
				assert.Equal(t, uint64(1), vm.Count)
				assert.Equal(t, 1, len(vm.List))
				assert.Equal(t, "Test User", vm.List[0].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := tt.setupDTO(t)
			vm := NewFindAllGuestResponseVM(dto)
			tt.validate(t, vm, dto)
		})
	}
}

func TestFindGuestByIDRequestVMToDTO(t *testing.T) {
	tests := []struct {
		name     string
		setupVM  func(t *testing.T) *protobuf_boilerplate.FindGuestByIDRequestVM
		validate func(t *testing.T, dto *dtos.FindGuestByIDRequestDTO, vm *protobuf_boilerplate.FindGuestByIDRequestVM)
	}{
		{
			name: "should_convert_findbyid_vm_to_dto",
			setupVM: func(t *testing.T) *protobuf_boilerplate.FindGuestByIDRequestVM {
				return &protobuf_boilerplate.FindGuestByIDRequestVM{
					Id: "550e8400-e29b-41d4-a716-446655440000",
				}
			},
			validate: func(t *testing.T, dto *dtos.FindGuestByIDRequestDTO, vm *protobuf_boilerplate.FindGuestByIDRequestVM) {
				assert.NotNil(t, dto)
				assert.Equal(t, vm.GetId(), dto.ID)
			},
		},
		{
			name: "should_convert_findbyid_vm_with_different_id",
			setupVM: func(t *testing.T) *protobuf_boilerplate.FindGuestByIDRequestVM {
				return &protobuf_boilerplate.FindGuestByIDRequestVM{
					Id: "123e4567-e89b-12d3-a456-426614174000",
				}
			},
			validate: func(t *testing.T, dto *dtos.FindGuestByIDRequestDTO, vm *protobuf_boilerplate.FindGuestByIDRequestVM) {
				assert.NotNil(t, dto)
				assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", dto.ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := tt.setupVM(t)
			dto := FindGuestByIDRequestVMToDTO(vm)
			tt.validate(t, dto, vm)
		})
	}
}

func TestNewGuestResponseVM(t *testing.T) {
	tests := []struct {
		name     string
		setupDTO func(t *testing.T) *dtos.GuestResponseDTO
		validate func(t *testing.T, vm *protobuf_boilerplate.GuestResponseVM, dto *dtos.GuestResponseDTO)
	}{
		{
			name: "should_convert_guest_dto_to_vm_with_all_fields",
			setupDTO: func(t *testing.T) *dtos.GuestResponseDTO {
				return &dtos.GuestResponseDTO{
					ID:        "550e8400-e29b-41d4-a716-446655440000",
					Name:      "John Doe",
					Address:   "123 Main St",
					CreatedAt: 1700000000000,
					CreatedBy: "user1",
					UpdatedAt: 1700000001000,
					UpdatedBy: "user2",
				}
			},
			validate: func(t *testing.T, vm *protobuf_boilerplate.GuestResponseVM, dto *dtos.GuestResponseDTO) {
				assert.NotNil(t, vm)
				assert.Equal(t, dto.ID, vm.Id)
				assert.Equal(t, dto.Name, vm.Name)
				assert.Equal(t, dto.Address, vm.Address)
				assert.Equal(t, dto.CreatedAt, vm.CreatedAt)
				assert.Equal(t, dto.CreatedBy, vm.CreatedBy)
				assert.Equal(t, dto.UpdatedAt, vm.UpdatedAt)
				assert.Equal(t, dto.UpdatedBy, vm.UpdatedBy)
			},
		},
		{
			name: "should_convert_guest_dto_with_empty_optional_fields",
			setupDTO: func(t *testing.T) *dtos.GuestResponseDTO {
				return &dtos.GuestResponseDTO{
					ID:        "abc12345-e89b-12d3-a456-426614174999",
					Name:      "Jane Smith",
					Address:   "",
					CreatedAt: 1700000000000,
					CreatedBy: "admin",
					UpdatedAt: 0,
					UpdatedBy: "",
				}
			},
			validate: func(t *testing.T, vm *protobuf_boilerplate.GuestResponseVM, dto *dtos.GuestResponseDTO) {
				assert.NotNil(t, vm)
				assert.Equal(t, "abc12345-e89b-12d3-a456-426614174999", vm.Id)
				assert.Equal(t, "Jane Smith", vm.Name)
				assert.Equal(t, "", vm.Address)
				assert.Equal(t, int64(0), vm.UpdatedAt)
				assert.Equal(t, "", vm.UpdatedBy)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := tt.setupDTO(t)
			vm := NewGuestResponseVM(dto)
			tt.validate(t, vm, dto)
		})
	}
}

func TestUpdateGuestByIDRequestVMToDTO(t *testing.T) {
	tests := []struct {
		name      string
		setupVM   func(t *testing.T) *protobuf_boilerplate.UpdateGuestByIDRequestVM
		updatedBy string
		validate  func(t *testing.T, dto *dtos.UpdateGuestByIDRequestDTO, vm *protobuf_boilerplate.UpdateGuestByIDRequestVM, updatedBy string)
	}{
		{
			name: "should_convert_update_vm_to_dto_successfully",
			setupVM: func(t *testing.T) *protobuf_boilerplate.UpdateGuestByIDRequestVM {
				return &protobuf_boilerplate.UpdateGuestByIDRequestVM{
					Id:      "550e8400-e29b-41d4-a716-446655440000",
					Name:    "Updated Name",
					Address: "Updated Address",
				}
			},
			updatedBy: "admin",
			validate: func(t *testing.T, dto *dtos.UpdateGuestByIDRequestDTO, vm *protobuf_boilerplate.UpdateGuestByIDRequestVM, updatedBy string) {
				assert.NotNil(t, dto)
				assert.Equal(t, vm.GetId(), dto.ID)
				assert.Equal(t, vm.GetName(), dto.Name)
				assert.Equal(t, vm.GetAddress(), dto.Address)
				assert.Equal(t, updatedBy, dto.UpdatedBy)
			},
		},
		{
			name: "should_convert_update_vm_with_empty_address",
			setupVM: func(t *testing.T) *protobuf_boilerplate.UpdateGuestByIDRequestVM {
				return &protobuf_boilerplate.UpdateGuestByIDRequestVM{
					Id:      "123e4567-e89b-12d3-a456-426614174000",
					Name:    "Test User",
					Address: "",
				}
			},
			updatedBy: "user123",
			validate: func(t *testing.T, dto *dtos.UpdateGuestByIDRequestDTO, vm *protobuf_boilerplate.UpdateGuestByIDRequestVM, updatedBy string) {
				assert.NotNil(t, dto)
				assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", dto.ID)
				assert.Equal(t, "Test User", dto.Name)
				assert.Equal(t, "", dto.Address)
				assert.Equal(t, "user123", dto.UpdatedBy)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := tt.setupVM(t)
			dto := UpdateGuestByIDRequestVMToDTO(vm, tt.updatedBy)
			tt.validate(t, dto, vm, tt.updatedBy)
		})
	}
}
