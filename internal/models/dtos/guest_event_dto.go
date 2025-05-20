package dtos

import "go-boilerplate/internal/models/entities"

type GuestEventRequestDTO struct {
	ID        string
	Name      string
	Address   string
	CreatedAt int64
	CreatedBy string
	UpdatedAt int64
	UpdatedBy string
	DeletedAt int64
	DeletedBy string
}

func (dto *GuestEventRequestDTO) ToEntity() *entities.GuestEventEntity {
	var entity entities.GuestEventEntity = entities.GuestEventEntity(*dto)
	return &entity
}

type GuestEventResponseDTO struct {
	ID        string
	Name      string
	Address   string
	CreatedAt int64
	CreatedBy string
	UpdatedAt int64
	UpdatedBy string
	DeletedAt int64
	DeletedBy string
}

func NewGuestEventResponseDTO(entity *entities.GuestEventEntity) *GuestEventResponseDTO {
	var dto GuestEventResponseDTO = GuestEventResponseDTO(*entity)
	return &dto
}
