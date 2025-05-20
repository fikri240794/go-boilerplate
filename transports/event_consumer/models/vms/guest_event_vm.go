package vms

import "go-boilerplate/internal/models/dtos"

type GuestEventRequestVM struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address,omitempty"`
	CreatedAt int64  `json:"created_at"`
	CreatedBy string `json:"created_by"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
	UpdatedBy string `json:"updated_by,omitempty"`
	DeletedAt int64  `json:"deleted_at,omitempty"`
	DeletedBy string `json:"deleted_by,omitempty"`
}

func (vm *GuestEventRequestVM) ToDTO() *dtos.GuestEventRequestDTO {
	var dto dtos.GuestEventRequestDTO = dtos.GuestEventRequestDTO(*vm)
	return &dto
}
