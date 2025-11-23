package in_memory_database

import (
	"errors"
	"go-boilerplate/configs"
	"go-boilerplate/datasources/in_memory_database/mocks"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func Test_defaultRedisTracer(t *testing.T) {
	tests := []struct {
		name        string
		setupClient func() *redis.Client
		expectError bool
	}{
		{
			name: "instrument tracing successfully",
			setupClient: func() *redis.Client {
				opts, _ := redis.ParseURL("redis://localhost:6379/0")
				return redis.NewClient(opts)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupClient()
			err := defaultRedisTracer(client)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			client.Close()
		})
	}
}

func Test_connectToRedisTracer(t *testing.T) {
	tests := []struct {
		name        string
		dsn         string
		fn          redisTracer
		expectError bool
		validate    func(t *testing.T, client IRedisClient)
	}{
		{
			name: "parse URL error should return error",
			dsn:  "invalid-dsn",
			fn: func(client *redis.Client) error {
				return nil
			},
			expectError: true,
		},
		{
			name: "tracer returns error should return error",
			dsn:  "redis://localhost:6379/0",
			fn: func(client *redis.Client) error {
				return errors.New("tracer error")
			},
			expectError: true,
		},
		{
			name: "connect successfully with tracer",
			dsn:  "redis://localhost:6379/0",
			fn: func(client *redis.Client) error {
				return nil
			},
			expectError: false,
			validate: func(t *testing.T, client IRedisClient) {
				assert.NotNil(t, client)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := connectToRedisTracer(tt.dsn, tt.fn)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)

				if tt.validate != nil {
					tt.validate(t, client)
				}

				client.Close()
			}
		})
	}
}

func Test_defaultRedisClient(t *testing.T) {
	tests := []struct {
		name        string
		dsn         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "parse URL error should return error",
			dsn:         "invalid-dsn",
			expectError: true,
			errorMsg:    "redis: invalid URL scheme:",
		},
		{
			name:        "valid DSN should create client successfully",
			dsn:         "redis://localhost:6379/0",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := defaultRedisClient(tt.dsn)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				client.Close()
			}
		})
	}
}

func Test_connectToRedis(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *configs.Config
		fn          func(t *testing.T, dsn string) (IRedisClient, error)
		expectPanic bool
		validate    func(t *testing.T, db *InMemoryDatabase)
	}{
		{
			name: "connect with fn returns error should panic",
			setupConfig: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.InMemoryDatabase.DataSourceName = "redis://localhost:6379"
				return cfg
			},
			fn: func(t *testing.T, dsn string) (IRedisClient, error) {
				return nil, errors.New("factory error")
			},
			expectPanic: true,
		},
		{
			name: "connect with fn returns parse error should panic",
			setupConfig: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.InMemoryDatabase.DataSourceName = "redis://localhost:6379"
				return cfg
			},
			fn: func(t *testing.T, dsn string) (IRedisClient, error) {
				return nil, errors.New("parse URL error")
			},
			expectPanic: true,
		},
		{
			name: "connect successfully with mock client",
			setupConfig: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.InMemoryDatabase.DataSourceName = "redis://localhost:6379"
				return cfg
			},
			fn: func(t *testing.T, dsn string) (IRedisClient, error) {
				mockClient := mocks.NewRedisClientMock(t)
				mockClient.On("Close").Return(nil)
				return mockClient, nil
			},
			expectPanic: false,
			validate: func(t *testing.T, db *InMemoryDatabase) {
				assert.NotNil(t, db)
				assert.NotNil(t, db.RedisClient)
				assert.IsType(t, &mocks.RedisClientMock{}, db.RedisClient)
			},
		},
		{
			name: "connect with instrumentation error should panic",
			setupConfig: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.InMemoryDatabase.DataSourceName = "redis://localhost:6379"
				return cfg
			},
			fn: func(t *testing.T, dsn string) (IRedisClient, error) {
				return nil, errors.New("instrumentation error")
			},
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()

			if tt.expectPanic {
				assert.Panics(t, func() {
					connectToRedis(cfg, func(dsn string) (IRedisClient, error) {
						return tt.fn(t, dsn)
					})
				})
			} else {
				db := connectToRedis(cfg, func(dsn string) (IRedisClient, error) {
					return tt.fn(t, dsn)
				})

				if tt.validate != nil {
					tt.validate(t, db)
				}

				err := db.Disconnect()
				assert.NoError(t, err)
			}
		})
	}
}

func TestConnect(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *configs.Config
		expectPanic bool
	}{
		{
			name: "connect with invalid DSN should panic",
			setupConfig: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.InMemoryDatabase.DataSourceName = "invalid-dsn-format"
				return cfg
			},
			expectPanic: true,
		},
		{
			name: "connect with empty DSN should panic",
			setupConfig: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.InMemoryDatabase.DataSourceName = ""
				return cfg
			},
			expectPanic: true,
		},
		{
			name: "connect with malformed URL should panic",
			setupConfig: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.InMemoryDatabase.DataSourceName = "://invalid"
				return cfg
			},
			expectPanic: true,
		},
		{
			name: "connect with valid DSN should succeed and cover default factory",
			setupConfig: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.InMemoryDatabase.DataSourceName = "redis://localhost:6379/0"
				return cfg
			},
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()

			if tt.expectPanic {
				assert.Panics(t, func() {
					Connect(cfg)
				})
			} else {
				db := Connect(cfg)
				assert.NotNil(t, db)
				assert.NotNil(t, db.RedisClient)

				err := db.Disconnect()
				assert.NoError(t, err)
			}
		})
	}
}

func TestDisconnect(t *testing.T) {
	tests := []struct {
		name        string
		setupDB     func(t *testing.T) *InMemoryDatabase
		expectError bool
		expectPanic bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "disconnect with nil RedisClient should panic",
			setupDB: func(t *testing.T) *InMemoryDatabase {
				return &InMemoryDatabase{
					RedisClient: nil,
				}
			},
			expectError: false,
			expectPanic: true,
		},
		{
			name: "disconnect successfully",
			setupDB: func(t *testing.T) *InMemoryDatabase {
				mockClient := mocks.NewRedisClientMock(t)
				mockClient.On("Close").Return(nil)
				return &InMemoryDatabase{
					RedisClient: mockClient,
				}
			},
			expectError: false,
			expectPanic: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "disconnect with close error",
			setupDB: func(t *testing.T) *InMemoryDatabase {
				mockClient := mocks.NewRedisClientMock(t)
				mockClient.On("Close").Return(errors.New("close error"))
				return &InMemoryDatabase{
					RedisClient: mockClient,
				}
			},
			expectError: true,
			expectPanic: false,
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Equal(t, "close error", err.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setupDB(t)

			if tt.expectPanic {
				assert.Panics(t, func() {
					db.Disconnect()
				})
			} else {
				err := db.Disconnect()

				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}

				if tt.validate != nil {
					tt.validate(t, err)
				}
			}
		})
	}
}
