package datasources

import (
	"context"
	"database/sql"
	"errors"
	"go-boilerplate/datasources/boilerplate_database"
	"go-boilerplate/datasources/event_producer"
	"go-boilerplate/datasources/in_memory_database"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type mockRedisClient struct {
	closeError  error
	closed      bool
	delError    error
	getError    error
	setError    error
	keysError   error
	incrError   error
	expireError error
}

func (m *mockRedisClient) Close() error {
	m.closed = true
	return m.closeError
}

func (m *mockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	if m.delError != nil {
		cmd.SetErr(m.delError)
	}
	return cmd
}

func (m *mockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx)
	if m.getError != nil {
		cmd.SetErr(m.getError)
	}
	return cmd
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	if m.setError != nil {
		cmd.SetErr(m.setError)
	}
	return cmd
}

func (m *mockRedisClient) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	cmd := redis.NewStringSliceCmd(ctx)
	if m.keysError != nil {
		cmd.SetErr(m.keysError)
	}
	return cmd
}

func (m *mockRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	if m.incrError != nil {
		cmd.SetErr(m.incrError)
	}
	return cmd
}

func (m *mockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	cmd := redis.NewBoolCmd(ctx)
	if m.expireError != nil {
		cmd.SetErr(m.expireError)
	}
	return cmd
}

type mockNSQProducer struct {
	pingError            error
	stopped              bool
	publishError         error
	deferredPublishError error
}

func (m *mockNSQProducer) Ping() error {
	return m.pingError
}

func (m *mockNSQProducer) Stop() {
	m.stopped = true
}

func (m *mockNSQProducer) Publish(topic string, body []byte) error {
	return m.publishError
}

func (m *mockNSQProducer) DeferredPublish(topic string, delay time.Duration, body []byte) error {
	return m.deferredPublishError
}

func TestDisconnect(t *testing.T) {
	tests := []struct {
		name                  string
		setupDatasources      func() *Datasources
		expectError           bool
		expectedErrorContains []string
		validate              func(t *testing.T, err error, ds *Datasources)
	}{
		{
			name: "disconnect with valid mock clients should succeed",
			setupDatasources: func() *Datasources {
				masterDB, _ := sql.Open("postgres", "host=localhost")
				slaveDB, _ := sql.Open("postgres", "host=localhost")

				return &Datasources{
					BoilerplateDatabase: &boilerplate_database.BoilerplateDatabase{
						Master: sqlx.NewDb(masterDB, "postgres"),
						Slave:  sqlx.NewDb(slaveDB, "postgres"),
					},
					InMemoryDatabase: &in_memory_database.InMemoryDatabase{
						RedisClient: &mockRedisClient{},
					},
					EventProducer: &event_producer.EventProducer{
						NSQProducer: &mockNSQProducer{},
					},
				}
			},
			expectError: false,
			validate: func(t *testing.T, err error, ds *Datasources) {
				assert.NoError(t, err)

				if redis, ok := ds.InMemoryDatabase.RedisClient.(*mockRedisClient); ok {
					assert.True(t, redis.closed, "Expected Redis client to be closed")
				}

				if nsq, ok := ds.EventProducer.NSQProducer.(*mockNSQProducer); ok {
					assert.True(t, nsq.stopped, "Expected NSQ producer to be stopped")
				}
			},
		},
		{
			name: "disconnect with redis close error should return error",
			setupDatasources: func() *Datasources {
				masterDB, _ := sql.Open("postgres", "host=localhost")
				slaveDB, _ := sql.Open("postgres", "host=localhost")

				return &Datasources{
					BoilerplateDatabase: &boilerplate_database.BoilerplateDatabase{
						Master: sqlx.NewDb(masterDB, "postgres"),
						Slave:  sqlx.NewDb(slaveDB, "postgres"),
					},
					InMemoryDatabase: &in_memory_database.InMemoryDatabase{
						RedisClient: &mockRedisClient{
							closeError: errors.New("redis close error"),
						},
					},
					EventProducer: &event_producer.EventProducer{
						NSQProducer: &mockNSQProducer{},
					},
				}
			},
			expectError:           true,
			expectedErrorContains: []string{"redis close error"},
			validate: func(t *testing.T, err error, ds *Datasources) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "redis close error")
			},
		},
		{
			name: "disconnect with database close errors should aggregate them",
			setupDatasources: func() *Datasources {
				masterDB, _ := sql.Open("postgres", "host=localhost")
				slaveDB, _ := sql.Open("postgres", "host=localhost")

				return &Datasources{
					BoilerplateDatabase: &boilerplate_database.BoilerplateDatabase{
						Master: sqlx.NewDb(masterDB, "postgres"),
						Slave:  sqlx.NewDb(slaveDB, "postgres"),
					},
					InMemoryDatabase: &in_memory_database.InMemoryDatabase{
						RedisClient: &mockRedisClient{
							closeError: errors.New("redis disconnect error"),
						},
					},
					EventProducer: &event_producer.EventProducer{
						NSQProducer: &mockNSQProducer{},
					},
				}
			},
			expectError: true,
			validate: func(t *testing.T, err error, ds *Datasources) {
				assert.Error(t, err)
				errMsg := err.Error()
				assert.Contains(t, errMsg, "redis disconnect error")
			},
		},
		{
			name: "disconnect calls all three datasource disconnect methods",
			setupDatasources: func() *Datasources {
				masterDB, _ := sql.Open("postgres", "host=localhost")
				slaveDB, _ := sql.Open("postgres", "host=localhost")

				return &Datasources{
					BoilerplateDatabase: &boilerplate_database.BoilerplateDatabase{
						Master: sqlx.NewDb(masterDB, "postgres"),
						Slave:  sqlx.NewDb(slaveDB, "postgres"),
					},
					InMemoryDatabase: &in_memory_database.InMemoryDatabase{
						RedisClient: &mockRedisClient{},
					},
					EventProducer: &event_producer.EventProducer{
						NSQProducer: &mockNSQProducer{},
					},
				}
			},
			expectError: false,
			validate: func(t *testing.T, err error, ds *Datasources) {
				redis, redisOk := ds.InMemoryDatabase.RedisClient.(*mockRedisClient)
				nsq, nsqOk := ds.EventProducer.NSQProducer.(*mockNSQProducer)

				assert.True(t, redisOk, "Expected RedisClient to be mockRedisClient")
				assert.True(t, nsqOk, "Expected NSQProducer to be mockNSQProducer")

				assert.True(t, redis.closed, "Expected Redis disconnect to be called")
				assert.True(t, nsq.stopped, "Expected NSQ disconnect to be called")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := tt.setupDatasources()

			err := ds.Disconnect()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedErrorContains != nil && err != nil {
				errMsg := err.Error()
				for _, expectedStr := range tt.expectedErrorContains {
					assert.Contains(t, errMsg, expectedStr)
				}
			}

			if tt.validate != nil {
				tt.validate(t, err, ds)
			}
		})
	}
}
