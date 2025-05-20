package entities

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/guregu/null/v5"
)

const (
	GuestEntityDatabaseFieldID        string = "id"
	GuestEntityDatabaseFieldName      string = "name"
	GuestEntityDatabaseFieldAddress   string = "address"
	GuestEntityDatabaseFieldCreatedAt string = "created_at"
	GuestEntityDatabaseFieldCreatedBy string = "created_by"
	GuestEntityDatabaseFieldUpdatedAt string = "updated_at"
	GuestEntityDatabaseFieldUpdatedBy string = "updated_by"
	GuestEntityDatabaseFieldDeletedAt string = "deleted_at"
	GuestEntityDatabaseFieldDeletedBy string = "deleted_by"
)

type GuestEntity struct {
	Table string `table:"guests" db:"-" json:"-"`

	ID        uuid.UUID   `db:"id" json:"id"`
	Name      string      `db:"name" json:"name"`
	Address   null.String `db:"address" json:"address"`
	CreatedAt int64       `db:"created_at" json:"created_at"`
	CreatedBy string      `db:"created_by" json:"created_by"`
	UpdatedAt null.Int64  `db:"updated_at" json:"updated_at"`
	UpdatedBy null.String `db:"updated_by" json:"updated_by"`
	DeletedAt null.Int64  `db:"deleted_at" json:"deleted_at"`
	DeletedBy null.String `db:"deleted_by" json:"deleted_by"`
}

func (entity *GuestEntity) MarkAsDeleted(deletedBy string) *GuestEntity {
	entity.DeletedAt = null.IntFrom(time.Now().UnixMilli())
	entity.DeletedBy = null.StringFrom(deletedBy)

	return entity
}
