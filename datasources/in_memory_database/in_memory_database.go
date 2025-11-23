package in_memory_database

import (
	"context"
	"go-boilerplate/configs"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

//go:generate go run github.com/vektra/mockery/v2 --name IRedisTracer --structname RedisTracerMock --filename redis_tracer_mock.go
type IRedisTracer interface {
	InstrumentTracing(client *redis.Client) error
}

type redisTracer func(client *redis.Client) error

var defaultRedisTracer redisTracer = func(client *redis.Client) error {
	return redisotel.InstrumentTracing(client)
}

//go:generate go run github.com/vektra/mockery/v2 --name IRedisClient --structname RedisClientMock --filename redis_client_mock.go
type IRedisClient interface {
	Close() error
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Keys(ctx context.Context, pattern string) *redis.StringSliceCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
}

type redisClient func(dsn string) (IRedisClient, error)

func connectToRedisTracer(dsn string, fn redisTracer) (IRedisClient, error) {
	var (
		redisOpts   *redis.Options
		redisClient *redis.Client
		err         error
	)

	redisOpts, err = redis.ParseURL(dsn)
	if err != nil {
		return nil, err
	}

	redisClient = redis.NewClient(redisOpts)

	err = fn(redisClient)
	if err != nil {
		return nil, err
	}

	return redisClient, nil
}

var defaultRedisClient redisClient = func(dsn string) (IRedisClient, error) {
	return connectToRedisTracer(dsn, defaultRedisTracer)
}

type InMemoryDatabase struct {
	RedisClient IRedisClient
}

func connectToRedis(cfg *configs.Config, fn redisClient) *InMemoryDatabase {
	var (
		redisClient IRedisClient
		err         error
	)

	redisClient, err = fn(cfg.Datasource.InMemoryDatabase.DataSourceName)
	if err != nil {
		panic(err)
	}

	return &InMemoryDatabase{
		RedisClient: redisClient,
	}
}

func Connect(cfg *configs.Config) *InMemoryDatabase {
	return connectToRedis(cfg, defaultRedisClient)
}

func (r *InMemoryDatabase) Disconnect() error {
	return r.RedisClient.Close()
}
