package dtos

import (
	"go-boilerplate/internal/models/entities"
	custom_uuid "go-boilerplate/pkg/uuid"
	"go-boilerplate/pkg/validator"
	"net/http"
	"strings"
	"time"

	"github.com/fikri240794/gocerr"
	"github.com/fikri240794/goqube"
	"github.com/guregu/null/v5"
)

type CreateGuestRequestDTO struct {
	Name      string `json:"name" validate:"required"`
	Address   string `json:"address,omitempty"`
	CreatedBy string `json:"created_by" validate:"required"`
}

func (dto *CreateGuestRequestDTO) Validate() error {
	return validator.ValidateStruct(dto)
}

func (dto *CreateGuestRequestDTO) ToEntity() *entities.GuestEntity {
	var entity *entities.GuestEntity = &entities.GuestEntity{
		ID:        custom_uuid.NewV7(),
		Name:      dto.Name,
		Address:   null.NewString(dto.Address, dto.Address != ""),
		CreatedAt: time.Now().UnixMilli(),
		CreatedBy: dto.CreatedBy,
	}

	return entity
}

type DeleteGuestByIDRequestDTO struct {
	ID        string `json:"id" validate:"uuid_rfc4122"`
	DeletedBy string `json:"deleted_by" validate:"required"`
}

func (dto *DeleteGuestByIDRequestDTO) Validate() error {
	return validator.ValidateStruct(dto)
}

type FindAllGuestRequestDTO struct {
	Keyword string `json:"keyword,omitempty"`
	Sorts   string `json:"sorts,omitempty"`
	Take    uint64 `json:"take,omitempty"`
	Skip    uint64 `json:"skip,omitempty"`
}

func NewFindAllGuestRequestDTO() *FindAllGuestRequestDTO {
	return &FindAllGuestRequestDTO{
		Take: 10,
	}
}
func (dto *FindAllGuestRequestDTO) ToFilterAndSorts() (*goqube.Filter, []goqube.Sort, error) {
	var (
		filter         *goqube.Filter
		splittedString []string
		sorts          []goqube.Sort
		err            error
	)

	filter = &goqube.Filter{
		Logic: goqube.LogicAnd,
		Filters: []goqube.Filter{
			{
				Field:    goqube.Field{Column: entities.GuestEntityDatabaseFieldDeletedAt},
				Operator: goqube.OperatorIsNull,
				Value:    goqube.FilterValue{Value: nil},
			},
		},
	}

	if dto.Keyword != "" {
		filter.Filters = append(filter.Filters, goqube.Filter{
			Logic: goqube.LogicOr,
			Filters: []goqube.Filter{
				{
					Field:    goqube.Field{Column: entities.GuestEntityDatabaseFieldName},
					Operator: goqube.OperatorLike,
					Value:    goqube.FilterValue{Value: dto.Keyword},
				},
				{
					Field:    goqube.Field{Column: entities.GuestEntityDatabaseFieldAddress},
					Operator: goqube.OperatorLike,
					Value:    goqube.FilterValue{Value: dto.Keyword},
				},
			},
		})
	}

	if dto.Sorts != "" {
		splittedString = strings.Split(dto.Sorts, ",")
		if len(splittedString) <= 0 {
			sorts = append(sorts, goqube.Sort{
				Field:     goqube.Field{Column: entities.GuestEntityDatabaseFieldName},
				Direction: goqube.SortDirectionAscending,
			})
		}
	}

	for i := range splittedString {
		var sortFieldAndDirection []string = strings.Split(splittedString[i], ".")
		if len(sortFieldAndDirection) <= 0 {
			err = gocerr.New(
				http.StatusBadRequest,
				http.StatusText(http.StatusBadRequest),
				gocerr.NewErrorField("sorts", "invalid sorts value"),
			)
			return nil, nil, err
		}

		if len(sortFieldAndDirection) == 1 {
			if sortFieldAndDirection[0] != entities.GuestEntityDatabaseFieldName &&
				sortFieldAndDirection[0] != entities.GuestEntityDatabaseFieldAddress {
				err = gocerr.New(
					http.StatusBadRequest,
					http.StatusText(http.StatusBadRequest),
					gocerr.NewErrorField("sorts", "invalid sorts field"),
				)
				return nil, nil, err
			}

			sorts = append(sorts, goqube.Sort{
				Field:     goqube.Field{Column: sortFieldAndDirection[0]},
				Direction: goqube.SortDirectionAscending,
			})
		}

		if len(sortFieldAndDirection) >= 2 {
			if sortFieldAndDirection[0] != entities.GuestEntityDatabaseFieldName &&
				sortFieldAndDirection[0] != entities.GuestEntityDatabaseFieldAddress {
				err = gocerr.New(
					http.StatusBadRequest,
					http.StatusText(http.StatusBadRequest),
					gocerr.NewErrorField("sorts", "invalid sorts field"),
				)
				return nil, nil, err
			}

			if goqube.SortDirection(sortFieldAndDirection[1]) != goqube.SortDirectionAscending &&
				goqube.SortDirection(sortFieldAndDirection[1]) != goqube.SortDirectionDescending {
				err = gocerr.New(
					http.StatusBadRequest,
					http.StatusText(http.StatusBadRequest),
					gocerr.NewErrorField("sorts", "invalid sorts direction"),
				)
				return nil, nil, err
			}

			sorts = append(sorts, goqube.Sort{
				Field:     goqube.Field{Column: sortFieldAndDirection[0]},
				Direction: goqube.SortDirection(sortFieldAndDirection[1]),
			})
		}
	}

	return filter, sorts, nil
}

type FindAllGuestResponseDTO struct {
	List         []GuestResponseDTO
	Count        uint64
	PreviousPage string
	NextPage     string
}

func NewFindAllGuestResponseDTO(listEntity []entities.GuestEntity, count uint64) *FindAllGuestResponseDTO {
	var responseDTO *FindAllGuestResponseDTO = &FindAllGuestResponseDTO{
		Count: count,
	}

	if len(listEntity) <= 0 {
		return responseDTO
	}

	for i := range listEntity {
		var dto *GuestResponseDTO = NewGuestResponseDTO(&listEntity[i])
		responseDTO.List = append(responseDTO.List, *dto)
	}

	return responseDTO
}

type FindGuestByIDRequestDTO struct {
	ID string `json:"id" validate:"uuid_rfc4122"`
}

func (dto *FindGuestByIDRequestDTO) Validate() error {
	return validator.ValidateStruct(dto)
}

type GuestResponseDTO struct {
	ID        string
	Name      string
	Address   string
	CreatedAt int64
	CreatedBy string
	UpdatedAt int64
	UpdatedBy string
}

func NewGuestResponseDTO(entity *entities.GuestEntity) *GuestResponseDTO {
	return &GuestResponseDTO{
		ID:        entity.ID.String(),
		Name:      entity.Name,
		Address:   entity.Address.ValueOrZero(),
		CreatedAt: entity.CreatedAt,
		CreatedBy: entity.CreatedBy,
		UpdatedAt: entity.UpdatedAt.ValueOrZero(),
		UpdatedBy: entity.UpdatedBy.ValueOrZero(),
	}
}

type UpdateGuestByIDRequestDTO struct {
	ID        string `json:"id" validate:"uuid_rfc4122"`
	Name      string `json:"name" validate:"required"`
	Address   string `json:"address,omitempty"`
	UpdatedBy string `json:"updated_by" validate:"required"`
}

func (dto *UpdateGuestByIDRequestDTO) Validate() error {
	return validator.ValidateStruct(dto)
}

func (dto *UpdateGuestByIDRequestDTO) ToExistingEntity(existingEntity *entities.GuestEntity) *entities.GuestEntity {
	existingEntity.Name = dto.Name
	existingEntity.Address = null.NewString(dto.Address, dto.Address != "")
	existingEntity.UpdatedAt = null.IntFrom(time.Now().UnixMilli())
	existingEntity.UpdatedBy = null.StringFrom(dto.UpdatedBy)

	return existingEntity
}
