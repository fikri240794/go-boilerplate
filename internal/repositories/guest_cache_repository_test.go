package repositories

import (
	"go-boilerplate/datasources/in_memory_database"
	in_memory_database_mocks "go-boilerplate/datasources/in_memory_database/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewGuestCacheRepository(t *testing.T) {
	tests := []struct {
		name             string
		inMemoryDatabase *in_memory_database.InMemoryDatabase
		expectNil        bool
	}{
		{
			name: "create guest cache repository with in memory database",
			inMemoryDatabase: &in_memory_database.InMemoryDatabase{
				RedisClient: in_memory_database_mocks.NewRedisClientMock(t),
			},
			expectNil: false,
		},
		{
			name:             "create guest cache repository without in memory database",
			inMemoryDatabase: nil,
			expectNil:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewGuestCacheRepository(tt.inMemoryDatabase)

			if tt.expectNil {
				assert.Nil(t, repo, "NewGuestCacheRepository() expected nil, got %v", repo)
			} else {
				assert.NotNil(t, repo, "NewGuestCacheRepository() expected non-nil repository, got nil")
				assert.Equal(t, tt.inMemoryDatabase, repo.inMemoryDatabase, "NewGuestCacheRepository() inMemoryDatabase mismatch")
			}
		})
	}
}
