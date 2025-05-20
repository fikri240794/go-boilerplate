package vms

import (
	"go-boilerplate/internal/models/dtos"
	"go-boilerplate/pkg/protobuf_boilerplate"
)

func CreateGuestRequestVMToDTO(vm *protobuf_boilerplate.CreateGuestRequestVM, createdBy string) *dtos.CreateGuestRequestDTO {
	var dto *dtos.CreateGuestRequestDTO = &dtos.CreateGuestRequestDTO{
		Name:      vm.GetName(),
		Address:   vm.GetAddress(),
		CreatedBy: createdBy,
	}

	return dto
}

func DeleteGuestByIDRequestVMToDTO(vm *protobuf_boilerplate.DeleteGuestByIDRequestVM, deletedBy string) *dtos.DeleteGuestByIDRequestDTO {
	var dto *dtos.DeleteGuestByIDRequestDTO = &dtos.DeleteGuestByIDRequestDTO{
		ID:        vm.GetId(),
		DeletedBy: deletedBy,
	}

	return dto
}

func FindAllGuestRequestVMToDTO(vm *protobuf_boilerplate.FindAllGuestRequestVM) *dtos.FindAllGuestRequestDTO {
	var dto *dtos.FindAllGuestRequestDTO = dtos.NewFindAllGuestRequestDTO()

	dto.Keyword = vm.GetKeyword()

	if vm.GetSorts() != "" {
		dto.Sorts = vm.GetSorts()
	}

	if vm.GetTake() > 0 {
		dto.Take = vm.GetTake()
	}

	if vm.GetSkip() > 0 {
		dto.Skip = vm.GetSkip()
	}

	return dto
}

func NewFindAllGuestResponseVM(dto *dtos.FindAllGuestResponseDTO) *protobuf_boilerplate.FindAllGuestResponseVM {
	var vm *protobuf_boilerplate.FindAllGuestResponseVM = &protobuf_boilerplate.FindAllGuestResponseVM{
		Count: dto.Count,
	}

	for i := range dto.List {
		var listItem *protobuf_boilerplate.GuestResponseVM = NewGuestResponseVM(&dto.List[i])
		vm.List = append(vm.List, listItem)
	}

	return vm
}

func FindGuestByIDRequestVMToDTO(vm *protobuf_boilerplate.FindGuestByIDRequestVM) *dtos.FindGuestByIDRequestDTO {
	var dto *dtos.FindGuestByIDRequestDTO = &dtos.FindGuestByIDRequestDTO{
		ID: vm.GetId(),
	}

	return dto
}

func NewGuestResponseVM(dto *dtos.GuestResponseDTO) *protobuf_boilerplate.GuestResponseVM {
	return &protobuf_boilerplate.GuestResponseVM{
		Id:        dto.ID,
		Name:      dto.Name,
		Address:   dto.Address,
		CreatedAt: dto.CreatedAt,
		CreatedBy: dto.CreatedBy,
		UpdatedAt: dto.UpdatedAt,
		UpdatedBy: dto.UpdatedBy,
	}
}

func UpdateGuestByIDRequestVMToDTO(vm *protobuf_boilerplate.UpdateGuestByIDRequestVM, updatedBy string) *dtos.UpdateGuestByIDRequestDTO {
	var dto *dtos.UpdateGuestByIDRequestDTO = &dtos.UpdateGuestByIDRequestDTO{
		ID:        vm.GetId(),
		Name:      vm.GetName(),
		Address:   vm.GetAddress(),
		UpdatedBy: updatedBy,
	}

	return dto
}
