package entities

type GuestEventEntity struct {
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

func NewGuestEventEntity(entity *GuestEntity) *GuestEventEntity {
	return &GuestEventEntity{
		ID:        entity.ID.String(),
		Name:      entity.Name,
		Address:   entity.Address.ValueOrZero(),
		CreatedAt: entity.CreatedAt,
		CreatedBy: entity.CreatedBy,
		UpdatedAt: entity.UpdatedAt.ValueOrZero(),
		UpdatedBy: entity.UpdatedBy.ValueOrZero(),
		DeletedAt: entity.DeletedAt.ValueOrZero(),
		DeletedBy: entity.DeletedBy.ValueOrZero(),
	}
}
