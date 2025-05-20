package repositories

import (
	"go-boilerplate/datasources/boilerplate_database"
	"go-boilerplate/internal/models/entities"
)

type IGuestRepository interface {
	IBoilerplateDatabaseRepository[entities.GuestEntity]

	WithTransaction(tx IBoilerplateDatabaseTransaction) IGuestRepository
}

type GuestRepository struct {
	BoilerplateDatabaseRepository[entities.GuestEntity]
}

func NewGuestRepository(databaseConnection *boilerplate_database.BoilerplateDatabase) *GuestRepository {
	return &GuestRepository{
		BoilerplateDatabaseRepository[entities.GuestEntity]{
			db: databaseConnection,
		},
	}
}

func (r *GuestRepository) WithTransaction(tx IBoilerplateDatabaseTransaction) IGuestRepository {
	return &GuestRepository{
		BoilerplateDatabaseRepository[entities.GuestEntity]{
			db: r.db,
			tx: tx,
		},
	}
}
