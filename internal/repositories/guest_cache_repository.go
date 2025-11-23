package repositories

import (
	"go-boilerplate/datasources/in_memory_database"
	"go-boilerplate/internal/models/entities"
)

//go:generate go run github.com/vektra/mockery/v2 --name IGuestCacheRepository --structname GuestCacheRepositoryMock --filename guest_cache_repository_mock.go
type IGuestCacheRepository interface {
	IInMemoryDatabaseRepository[entities.GuestEntity]
}

type GuestCacheRepository struct {
	InMemoryDatabaseRepository[entities.GuestEntity]
}

func NewGuestCacheRepository(inMemoryDatabase *in_memory_database.InMemoryDatabase) *GuestCacheRepository {
	return &GuestCacheRepository{
		InMemoryDatabaseRepository: InMemoryDatabaseRepository[entities.GuestEntity]{
			inMemoryDatabase: inMemoryDatabase,
		},
	}
}
