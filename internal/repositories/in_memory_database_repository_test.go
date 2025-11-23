package repositories

import (
	"context"
	"go-boilerplate/datasources/in_memory_database"
	in_memory_database_mocks "go-boilerplate/datasources/in_memory_database/mocks"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testInMemoryEntity struct {
	ID   int
	Name string
}

type testInMemoryEntityWithChannel struct {
	ID      int
	Name    string
	Channel chan int
}

func Test_NewInMemoryDatabaseRepository(t *testing.T) {
	tests := []struct {
		name             string
		inMemoryDatabase *in_memory_database.InMemoryDatabase
		expectNil        bool
	}{
		{
			name: "create in memory database repository with in memory database",
			inMemoryDatabase: &in_memory_database.InMemoryDatabase{
				RedisClient: in_memory_database_mocks.NewRedisClientMock(t),
			},
			expectNil: false,
		},
		{
			name:             "create in memory database repository without in memory database",
			inMemoryDatabase: nil,
			expectNil:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInMemoryDatabaseRepository[testInMemoryEntity](tt.inMemoryDatabase)

			if tt.expectNil {
				assert.Nil(t, repo, "NewInMemoryDatabaseRepository() expected nil, got %v", repo)
			} else {
				assert.NotNil(t, repo, "NewInMemoryDatabaseRepository() expected non-nil repository, got nil")
				assert.Equal(t, tt.inMemoryDatabase, repo.inMemoryDatabase, "NewInMemoryDatabaseRepository() inMemoryDatabase mismatch")
			}
		})
	}
}

func Test_InMemoryDatabaseRepository_Delete(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity]
		keys        []string
		expectError bool
	}{
		{
			name: "delete successfully",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewIntCmd(context.Background())
				cmd.SetVal(2)
				mockRedis.On("Del", mock.Anything, "key1", "key2").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			keys:        []string{"key1", "key2"},
			expectError: false,
		},
		{
			name: "delete with redis error",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewIntCmd(context.Background())
				cmd.SetErr(redis.TxFailedErr)
				mockRedis.On("Del", mock.Anything, "key1").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			keys:        []string{"key1"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo(t)
			ctx := context.Background()

			err := repo.Delete(ctx, tt.keys...)

			if tt.expectError {
				assert.Error(t, err, "Delete() expected error, got nil")
			} else {
				assert.NoError(t, err, "Delete() unexpected error")
			}
		})
	}
}

func Test_InMemoryDatabaseRepository_Get(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity]
		key         string
		expectError bool
		validate    func(t *testing.T, result *testInMemoryEntity, err error)
	}{
		{
			name: "get successfully",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStringCmd(context.Background())
				jsonData := `{"ID":1,"Name":"test"}`
				cmd.SetVal(jsonData)
				mockRedis.On("Get", mock.Anything, "test_key").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "test_key",
			expectError: false,
			validate: func(t *testing.T, result *testInMemoryEntity, err error) {
				assert.NoError(t, err, "Get() unexpected error")
				assert.NotNil(t, result, "Get() expected non-nil result")
				assert.Equal(t, 1, result.ID, "Get() result.ID mismatch")
				assert.Equal(t, "test", result.Name, "Get() result.Name mismatch")
			},
		},
		{
			name: "get with redis.Nil error (not found)",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetErr(redis.Nil)
				mockRedis.On("Get", mock.Anything, "missing_key").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "missing_key",
			expectError: true,
			validate: func(t *testing.T, result *testInMemoryEntity, err error) {
				assert.Error(t, err, "Get() expected error, got nil")
				assert.Nil(t, result, "Get() expected nil result")
			},
		},
		{
			name: "get with redis error (internal error)",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetErr(redis.TxFailedErr)
				mockRedis.On("Get", mock.Anything, "error_key").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "error_key",
			expectError: true,
			validate: func(t *testing.T, result *testInMemoryEntity, err error) {
				assert.Error(t, err, "Get() expected error, got nil")
				assert.Nil(t, result, "Get() expected nil result")
			},
		},
		{
			name: "get with json unmarshal error",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetVal("invalid json data")
				mockRedis.On("Get", mock.Anything, "invalid_json_key").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "invalid_json_key",
			expectError: true,
			validate: func(t *testing.T, result *testInMemoryEntity, err error) {
				assert.Error(t, err, "Get() expected error, got nil")
				assert.Nil(t, result, "Get() expected nil result")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo(t)
			ctx := context.Background()

			result, err := repo.Get(ctx, tt.key)

			if tt.validate != nil {
				tt.validate(t, result, err)
			}
		})
	}
}

func Test_InMemoryDatabaseRepository_GetList(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity]
		key         string
		expectError bool
		validate    func(t *testing.T, result []testInMemoryEntity, err error)
	}{
		{
			name: "get list successfully",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStringCmd(context.Background())
				jsonData := `[{"ID":1,"Name":"test1"},{"ID":2,"Name":"test2"}]`
				cmd.SetVal(jsonData)
				mockRedis.On("Get", mock.Anything, "test_list_key").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "test_list_key",
			expectError: false,
			validate: func(t *testing.T, result []testInMemoryEntity, err error) {
				assert.NoError(t, err, "GetList() unexpected error")
				assert.NotNil(t, result, "GetList() expected non-nil result")
				assert.Len(t, result, 2, "GetList() result length mismatch")
				assert.Equal(t, 1, result[0].ID, "GetList() result[0].ID mismatch")
				assert.Equal(t, "test1", result[0].Name, "GetList() result[0].Name mismatch")
				assert.Equal(t, 2, result[1].ID, "GetList() result[1].ID mismatch")
				assert.Equal(t, "test2", result[1].Name, "GetList() result[1].Name mismatch")
			},
		},
		{
			name: "get list with redis.Nil error (not found)",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetErr(redis.Nil)
				mockRedis.On("Get", mock.Anything, "missing_list_key").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "missing_list_key",
			expectError: true,
			validate: func(t *testing.T, result []testInMemoryEntity, err error) {
				assert.Error(t, err, "GetList() expected error, got nil")
				assert.Nil(t, result, "GetList() expected nil result")
			},
		},
		{
			name: "get list with redis error (internal error)",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetErr(redis.TxFailedErr)
				mockRedis.On("Get", mock.Anything, "error_list_key").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "error_list_key",
			expectError: true,
			validate: func(t *testing.T, result []testInMemoryEntity, err error) {
				assert.Error(t, err, "GetList() expected error, got nil")
				assert.Nil(t, result, "GetList() expected nil result")
			},
		},
		{
			name: "get list with json unmarshal error",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetVal("invalid json array")
				mockRedis.On("Get", mock.Anything, "invalid_json_list_key").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "invalid_json_list_key",
			expectError: true,
			validate: func(t *testing.T, result []testInMemoryEntity, err error) {
				assert.Error(t, err, "GetList() expected error, got nil")
				assert.Nil(t, result, "GetList() expected nil result")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo(t)
			ctx := context.Background()

			result, err := repo.GetList(ctx, tt.key)

			if tt.validate != nil {
				tt.validate(t, result, err)
			}
		})
	}
}

func Test_InMemoryDatabaseRepository_GetCount(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity]
		key         string
		expectError bool
		validate    func(t *testing.T, result uint64, err error)
	}{
		{
			name: "get count successfully",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetVal("42")
				mockRedis.On("Get", mock.Anything, "count_key").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "count_key",
			expectError: false,
			validate: func(t *testing.T, result uint64, err error) {
				assert.NoError(t, err, "GetCount() unexpected error")
				assert.Equal(t, uint64(42), result, "GetCount() result mismatch")
			},
		},
		{
			name: "get count with redis.Nil error (not found)",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetErr(redis.Nil)
				mockRedis.On("Get", mock.Anything, "missing_count_key").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "missing_count_key",
			expectError: true,
			validate: func(t *testing.T, result uint64, err error) {
				assert.Error(t, err, "GetCount() expected error, got nil")
				assert.Equal(t, uint64(0), result, "GetCount() expected 0 result")
			},
		},
		{
			name: "get count with redis error (internal error)",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetErr(redis.TxFailedErr)
				mockRedis.On("Get", mock.Anything, "error_count_key").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "error_count_key",
			expectError: true,
			validate: func(t *testing.T, result uint64, err error) {
				assert.Error(t, err, "GetCount() expected error, got nil")
				assert.Equal(t, uint64(0), result, "GetCount() expected 0 result")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo(t)
			ctx := context.Background()

			result, err := repo.GetCount(ctx, tt.key)

			if tt.validate != nil {
				tt.validate(t, result, err)
			}
		})
	}
}

func Test_InMemoryDatabaseRepository_Keys(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity]
		pattern     string
		expectError bool
		validate    func(t *testing.T, result []string, err error)
	}{
		{
			name: "get keys successfully",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStringSliceCmd(context.Background())
				cmd.SetVal([]string{"key1", "key2", "key3"})
				mockRedis.On("Keys", mock.Anything, "key*").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			pattern:     "key*",
			expectError: false,
			validate: func(t *testing.T, result []string, err error) {
				assert.NoError(t, err, "Keys() unexpected error")
				assert.NotNil(t, result, "Keys() expected non-nil result")
				assert.Len(t, result, 3, "Keys() result length mismatch")
				assert.Equal(t, []string{"key1", "key2", "key3"}, result, "Keys() result mismatch")
			},
		},
		{
			name: "get keys with redis error",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStringSliceCmd(context.Background())
				cmd.SetErr(redis.TxFailedErr)
				mockRedis.On("Keys", mock.Anything, "error*").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			pattern:     "error*",
			expectError: true,
			validate: func(t *testing.T, result []string, err error) {
				assert.Error(t, err, "Keys() expected error, got nil")
				assert.Nil(t, result, "Keys() expected nil result")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo(t)
			ctx := context.Background()

			result, err := repo.Keys(ctx, tt.pattern)

			if tt.validate != nil {
				tt.validate(t, result, err)
			}
		})
	}
}

func Test_InMemoryDatabaseRepository_Lock(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity]
		key         string
		expiration  time.Duration
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "lock successfully with expiration",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetVal(1)
				expireCmd := redis.NewBoolCmd(context.Background())
				expireCmd.SetVal(true)
				mockRedis.On("Incr", mock.Anything, "lock_key").Return(incrCmd)
				mockRedis.On("Expire", mock.Anything, "lock_key", 5*time.Second).Return(expireCmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "lock_key",
			expiration:  5 * time.Second,
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, "Lock() unexpected error")
			},
		},
		{
			name: "lock successfully without expiration",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetVal(1)
				mockRedis.On("Incr", mock.Anything, "lock_key_no_expire").Return(incrCmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "lock_key_no_expire",
			expiration:  0,
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, "Lock() unexpected error")
			},
		},
		{
			name: "lock with incr error",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetErr(redis.TxFailedErr)
				mockRedis.On("Incr", mock.Anything, "lock_key_incr_error").Return(incrCmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "lock_key_incr_error",
			expiration:  5 * time.Second,
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "Lock() expected error, got nil")
			},
		},
		{
			name: "lock with expire error",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetVal(1)
				expireCmd := redis.NewBoolCmd(context.Background())
				expireCmd.SetErr(redis.TxFailedErr)
				mockRedis.On("Incr", mock.Anything, "lock_key_expire_error").Return(incrCmd)
				mockRedis.On("Expire", mock.Anything, "lock_key_expire_error", 5*time.Second).Return(expireCmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "lock_key_expire_error",
			expiration:  5 * time.Second,
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "Lock() expected error, got nil")
			},
		},
		{
			name: "lock when already locked (counter > 1)",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				incrCmd := redis.NewIntCmd(context.Background())
				incrCmd.SetVal(2)
				expireCmd := redis.NewBoolCmd(context.Background())
				expireCmd.SetVal(true)
				mockRedis.On("Incr", mock.Anything, "already_locked_key").Return(incrCmd)
				mockRedis.On("Expire", mock.Anything, "already_locked_key", 5*time.Second).Return(expireCmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "already_locked_key",
			expiration:  5 * time.Second,
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "Lock() expected error, got nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo(t)
			ctx := context.Background()

			err := repo.Lock(ctx, tt.key, tt.expiration)

			if tt.validate != nil {
				tt.validate(t, err)
			}
		})
	}
}

func Test_InMemoryDatabaseRepository_Set(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(t *testing.T) interface{}
		key         string
		value       interface{}
		expiration  time.Duration
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "set successfully",
			setupRepo: func(t *testing.T) interface{} {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetVal("OK")
				mockRedis.On("Set", mock.Anything, "test_key", mock.Anything, 5*time.Second).Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key: "test_key",
			value: &testInMemoryEntity{
				ID:   1,
				Name: "test",
			},
			expiration:  5 * time.Second,
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, "Set() unexpected error")
			},
		},
		{
			name: "set with redis error",
			setupRepo: func(t *testing.T) interface{} {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetErr(redis.TxFailedErr)
				mockRedis.On("Set", mock.Anything, "test_key_error", mock.Anything, 5*time.Second).Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key: "test_key_error",
			value: &testInMemoryEntity{
				ID:   1,
				Name: "test",
			},
			expiration:  5 * time.Second,
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "Set() expected error, got nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo(t).(*InMemoryDatabaseRepository[testInMemoryEntity])
			ctx := context.Background()

			err := repo.Set(ctx, tt.key, tt.value.(*testInMemoryEntity), tt.expiration)

			if tt.validate != nil {
				tt.validate(t, err)
			}
		})
	}

	t.Run("set with json marshal error", func(t *testing.T) {
		mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
		repo := NewInMemoryDatabaseRepository[testInMemoryEntityWithChannel](&in_memory_database.InMemoryDatabase{
			RedisClient: mockRedis,
		})
		ctx := context.Background()

		value := &testInMemoryEntityWithChannel{
			ID:      1,
			Name:    "test",
			Channel: make(chan int),
		}

		err := repo.Set(ctx, "test_key", value, 5*time.Second)

		assert.Error(t, err, "Set() expected error for unmarshalable type, got nil")
	})
}

func Test_InMemoryDatabaseRepository_SetList(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity]
		key         string
		values      []testInMemoryEntity
		expiration  time.Duration
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "set list successfully",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetVal("OK")
				mockRedis.On("Set", mock.Anything, "list_key", mock.Anything, 5*time.Second).Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key: "list_key",
			values: []testInMemoryEntity{
				{ID: 1, Name: "test1"},
				{ID: 2, Name: "test2"},
			},
			expiration:  5 * time.Second,
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, "SetList() unexpected error")
			},
		},
		{
			name: "set list with redis error",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetErr(redis.TxFailedErr)
				mockRedis.On("Set", mock.Anything, "error_list_key", mock.Anything, 5*time.Second).Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key: "error_list_key",
			values: []testInMemoryEntity{
				{ID: 1, Name: "test1"},
			},
			expiration:  5 * time.Second,
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "SetList() expected error, got nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo(t)
			ctx := context.Background()

			err := repo.SetList(ctx, tt.key, tt.values, tt.expiration)

			if tt.validate != nil {
				tt.validate(t, err)
			}
		})
	}

	t.Run("set list with json marshal error", func(t *testing.T) {
		mockRedis := in_memory_database_mocks.NewRedisClientMock(t)

		repo := NewInMemoryDatabaseRepository[testInMemoryEntityWithChannel](&in_memory_database.InMemoryDatabase{
			RedisClient: mockRedis,
		})
		ctx := context.Background()

		values := []testInMemoryEntityWithChannel{
			{
				ID:      1,
				Name:    "test",
				Channel: make(chan int),
			},
		}

		err := repo.SetList(ctx, "test_list_key", values, 5*time.Second)

		assert.Error(t, err, "SetList() expected error for unmarshalable type, got nil")
	})
}

func Test_InMemoryDatabaseRepository_SetCount(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity]
		key         string
		value       uint64
		expiration  time.Duration
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "set count successfully",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetVal("OK")
				mockRedis.On("Set", mock.Anything, "count_key", uint64(100), 5*time.Second).Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "count_key",
			value:       100,
			expiration:  5 * time.Second,
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, "SetCount() unexpected error")
			},
		},
		{
			name: "set count with redis error",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetErr(redis.TxFailedErr)
				mockRedis.On("Set", mock.Anything, "error_count_key", uint64(200), 5*time.Second).Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "error_count_key",
			value:       200,
			expiration:  5 * time.Second,
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "SetCount() expected error, got nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo(t)
			ctx := context.Background()

			err := repo.SetCount(ctx, tt.key, tt.value, tt.expiration)

			if tt.validate != nil {
				tt.validate(t, err)
			}
		})
	}
}

func Test_InMemoryDatabaseRepository_Unlock(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity]
		key         string
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "unlock successfully",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewIntCmd(context.Background())
				cmd.SetVal(1)
				mockRedis.On("Del", mock.Anything, "lock_key").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "lock_key",
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, "Unlock() unexpected error")
			},
		},
		{
			name: "unlock with redis error",
			setupRepo: func(t *testing.T) *InMemoryDatabaseRepository[testInMemoryEntity] {
				mockRedis := in_memory_database_mocks.NewRedisClientMock(t)
				cmd := redis.NewIntCmd(context.Background())
				cmd.SetErr(redis.TxFailedErr)
				mockRedis.On("Del", mock.Anything, "error_lock_key").Return(cmd)
				return NewInMemoryDatabaseRepository[testInMemoryEntity](&in_memory_database.InMemoryDatabase{
					RedisClient: mockRedis,
				})
			},
			key:         "error_lock_key",
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err, "Unlock() expected error, got nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo(t)
			ctx := context.Background()

			err := repo.Unlock(ctx, tt.key)

			if tt.validate != nil {
				tt.validate(t, err)
			}
		})
	}
}
