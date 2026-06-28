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

	ID        uuid.UUID   `db:"id" json:"id" primary_key:"true" db_type:"uuid"`
	Name      string      `db:"name" json:"name" db_type:"text"`
	Address   null.String `db:"address" json:"address" db_type:"text"`
	CreatedAt int64       `db:"created_at" json:"created_at" db_type:"bigint"`
	CreatedBy string      `db:"created_by" json:"created_by" db_type:"text"`
	UpdatedAt null.Int64  `db:"updated_at" json:"updated_at" db_type:"bigint"`
	UpdatedBy null.String `db:"updated_by" json:"updated_by" db_type:"text"`
	DeletedAt null.Int64  `db:"deleted_at" json:"deleted_at" db_type:"bigint"`
	DeletedBy null.String `db:"deleted_by" json:"deleted_by" db_type:"text"`
}

func (entity *GuestEntity) MarkAsDeleted(deletedBy string) *GuestEntity {
	entity.DeletedAt = null.IntFrom(time.Now().UnixMilli())
	entity.DeletedBy = null.StringFrom(deletedBy)

	return entity
}
