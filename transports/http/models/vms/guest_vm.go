package vms

import (
	"go-boilerplate/internal/models/dtos"
)

type CreateGuestRequestVM struct {
	Name    string `json:"name" example:"John Snow"`
	Address string `json:"address" example:"123 Main Street, Apt. 4B, New York, NY 10001, USA"`
}

func (vm *CreateGuestRequestVM) ToDTO(createdBy string) *dtos.CreateGuestRequestDTO {
	var dto *dtos.CreateGuestRequestDTO = &dtos.CreateGuestRequestDTO{
		Name:      vm.Name,
		Address:   vm.Address,
		CreatedBy: createdBy,
	}

	return dto
}

type DeleteGuestByIDRequestVM struct {
	ID string `params:"id"`
}

func (vm *DeleteGuestByIDRequestVM) ToDTO(deletedBy string) *dtos.DeleteGuestByIDRequestDTO {
	var dto *dtos.DeleteGuestByIDRequestDTO = &dtos.DeleteGuestByIDRequestDTO{
		ID:        vm.ID,
		DeletedBy: deletedBy,
	}

	return dto
}

type FindAllGuestRequestVM struct {
	Keyword string `query:"keyword"`
	Sorts   string `query:"sorts"`
	Take    uint64 `query:"take"`
	Skip    uint64 `query:"skip"`
}

func (vm *FindAllGuestRequestVM) ToDTO() *dtos.FindAllGuestRequestDTO {
	var dto *dtos.FindAllGuestRequestDTO = dtos.NewFindAllGuestRequestDTO()

	dto.Keyword = vm.Keyword

	if vm.Sorts != "" {
		dto.Sorts = vm.Sorts
	}

	if vm.Take > 0 {
		dto.Take = vm.Take
	}

	if vm.Skip > 0 {
		dto.Skip = vm.Skip
	}

	return dto
}

type FindAllGuestResponseVM struct {
	List  []GuestResponseVM `json:"list"`
	Count uint64            `json:"count" example:"10"`
}

func NewFindAllGuestResponseVM(dto *dtos.FindAllGuestResponseDTO) *FindAllGuestResponseVM {
	var vm *FindAllGuestResponseVM = &FindAllGuestResponseVM{
		Count: dto.Count,
	}

	for i := range dto.List {
		var listItem *GuestResponseVM = NewGuestResponseVM(&dto.List[i])
		vm.List = append(vm.List, *listItem)
	}

	return vm
}

type FindGuestByIDRequestVM struct {
	ID string `params:"id"`
}

func (vm *FindGuestByIDRequestVM) ToDTO() *dtos.FindGuestByIDRequestDTO {
	var dto *dtos.FindGuestByIDRequestDTO = &dtos.FindGuestByIDRequestDTO{
		ID: vm.ID,
	}

	return dto
}

type GuestResponseVM struct {
	ID        string `json:"id" example:"01932293-d710-7f55-a9f6-66e6248ae72f"`
	Name      string `json:"name" example:"John Snow"`
	Address   string `json:"address,omitempty" example:"123 Main Street, Apt. 4B, New York, NY 10001, USA"`
	CreatedAt int64  `json:"created_at" example:"1731452061534"`
	CreatedBy string `json:"created_by" example:"Daenerys"`
	UpdatedAt int64  `json:"updated_at,omitempty" example:"1731452061534"`
	UpdatedBy string `json:"updated_by,omitempty" example:"Daenerys"`
}

func NewGuestResponseVM(dto *dtos.GuestResponseDTO) *GuestResponseVM {
	var vm GuestResponseVM = GuestResponseVM(*dto)
	return &vm
}

type UpdateGuestByIDRequestVM struct {
	ID      string `json:"-" params:"id"`
	Name    string `json:"name" example:"John Snow"`
	Address string `json:"address,omitempty" example:"123 Main Street, Apt. 4B, New York, NY 10001, USA"`
}

func (vm *UpdateGuestByIDRequestVM) ToDTO(updatedBy string) *dtos.UpdateGuestByIDRequestDTO {
	var dto *dtos.UpdateGuestByIDRequestDTO = &dtos.UpdateGuestByIDRequestDTO{
		ID:        vm.ID,
		Name:      vm.Name,
		Address:   vm.Address,
		UpdatedBy: updatedBy,
	}

	return dto
}
