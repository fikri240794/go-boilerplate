package repositories

import (
	"go-boilerplate/datasources/in_memory_database"
	"go-boilerplate/internal/models/entities"
)

//mockery:generate: true
//mockery:structname: GuestCacheRepositoryMock
//mockery:filename: guest_cache_repository_mock.go
//mockery:output: internal/repositories/mocks/
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
