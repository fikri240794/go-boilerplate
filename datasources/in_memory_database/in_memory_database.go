package in_memory_database

import (
	"go-boilerplate/configs"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

type InMemoryDatabase struct {
	RedisClient *redis.Client
}

func Connect(cfg *configs.Config) *InMemoryDatabase {
	var (
		redisOpts   *redis.Options
		redisClient *redis.Client
		err         error
	)

	redisOpts, err = redis.ParseURL(cfg.Datasource.InMemoryDatabase.DataSourceName)
	if err != nil {
		panic(err)
	}

	redisClient = redis.NewClient(redisOpts)

	err = redisotel.InstrumentTracing(redisClient)
	if err != nil {
		panic(err)
	}

	return &InMemoryDatabase{
		RedisClient: redisClient,
	}
}

func (r *InMemoryDatabase) Disconnect() error {
	return r.RedisClient.Close()
}
