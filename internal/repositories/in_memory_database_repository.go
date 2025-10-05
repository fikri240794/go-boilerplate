package repositories

import (
	"context"
	"fmt"
	"go-boilerplate/datasources/in_memory_database"
	"go-boilerplate/pkg/constants"
	custom_context "go-boilerplate/pkg/context"
	"go-boilerplate/pkg/tracer"
	"net/http"
	"time"

	"github.com/fikri240794/gocerr"
	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

type IInMemoryDatabaseRepository[TEntity interface{}] interface {
	Delete(ctx context.Context, keys ...string) error
	Get(ctx context.Context, key string) (*TEntity, error)
	GetList(ctx context.Context, key string) ([]TEntity, error)
	GetCount(ctx context.Context, key string) (uint64, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
	Lock(ctx context.Context, key string, expiration time.Duration) error
	Set(ctx context.Context, key string, value *TEntity, expiration time.Duration) error
	SetList(ctx context.Context, key string, values []TEntity, expiration time.Duration) error
	SetCount(ctx context.Context, key string, value uint64, expiration time.Duration) error
	Unlock(ctx context.Context, key string) error
}

type InMemoryDatabaseRepository[TEntity interface{}] struct {
	inMemoryDatabase *in_memory_database.InMemoryDatabase
}

func NewInMemoryDatabaseRepository[TEntity interface{}](inMemoryDatabase *in_memory_database.InMemoryDatabase) *InMemoryDatabaseRepository[TEntity] {
	return &InMemoryDatabaseRepository[TEntity]{
		inMemoryDatabase: inMemoryDatabase,
	}
}

func (r *InMemoryDatabaseRepository[TEntity]) Delete(ctx context.Context, keys ...string) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		err       error
	)

	ctx, span = tracer.Start(ctx, "[InMemoryDatabaseRepository][Delete]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"keys":      keys,
	}

	_, err = r.inMemoryDatabase.RedisClient.Del(ctx, keys...).
		Result()
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[InMemoryDatabaseRepository][Delete][Del][Result] failed to delete")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[InMemoryDatabaseRepository][Delete][Del][Result] failed to delete")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	return nil
}

func (r *InMemoryDatabaseRepository[TEntity]) Get(ctx context.Context, key string) (*TEntity, error) {
	var (
		span      trace.Span
		logFields map[string]interface{}
		rawValue  string
		value     *TEntity
		errorCode int
		err       error
	)

	ctx, span = tracer.Start(ctx, "[InMemoryDatabaseRepository][Get]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"key":       key,
	}

	err = r.inMemoryDatabase.RedisClient.
		Get(ctx, key).
		Scan(&rawValue)
	if err != nil {
		errorCode = http.StatusNotFound
		if err != redis.Nil {
			errorCode = http.StatusInternalServerError
			log.Err(err).
				Ctx(ctx).
				Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
				Msg("[InMemoryDatabaseRepository][Get][Get][Scan] failed to get")
		}

		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[InMemoryDatabaseRepository][Get][Get][Scan] failed to get")
		err = gocerr.New(errorCode, err.Error())
		return nil, err
	}

	logFields["rawValue"] = rawValue

	value = new(TEntity)
	err = json.Unmarshal([]byte(rawValue), value)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[InMemoryDatabaseRepository][Get][Unmarshal] failed to unmarshal raw value")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[InMemoryDatabaseRepository][Get][Unmarshal] failed to unmarshal raw value")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return nil, err
	}

	return value, nil
}

func (r *InMemoryDatabaseRepository[TEntity]) GetList(ctx context.Context, key string) ([]TEntity, error) {
	var (
		span      trace.Span
		logFields map[string]interface{}
		rawValues string
		values    []TEntity
		errorCode int
		err       error
	)

	ctx, span = tracer.Start(ctx, "[InMemoryDatabaseRepository][GetList]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"key":       key,
	}

	err = r.inMemoryDatabase.RedisClient.
		Get(ctx, key).
		Scan(&rawValues)
	if err != nil {
		errorCode = http.StatusNotFound
		if err != redis.Nil {
			errorCode = http.StatusInternalServerError
			log.Err(err).
				Ctx(ctx).
				Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
				Msg("[InMemoryDatabaseRepository][GetList][Get][Scan] failed to get")
		}

		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[InMemoryDatabaseRepository][GetList][Get][Scan] failed to get")
		err = gocerr.New(errorCode, err.Error())
		return nil, err
	}

	logFields["rawValues"] = rawValues

	err = json.Unmarshal([]byte(rawValues), &values)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[InMemoryDatabaseRepository][GetList][Unmarshal] failed to unmarshal raw values")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[InMemoryDatabaseRepository][GetList][Unmarshal] failed to unmarshal raw values")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return nil, err
	}

	return values, nil
}

func (r *InMemoryDatabaseRepository[TEntity]) GetCount(ctx context.Context, key string) (uint64, error) {
	var (
		span      trace.Span
		logFields map[string]interface{}
		value     uint64
		errorCode int
		err       error
	)

	ctx, span = tracer.Start(ctx, "[InMemoryDatabaseRepository][GetCount]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"key":       key,
	}

	value, err = r.inMemoryDatabase.RedisClient.
		Get(ctx, key).
		Uint64()
	if err != nil {
		errorCode = http.StatusNotFound
		if err != redis.Nil {
			errorCode = http.StatusInternalServerError
			log.Err(err).
				Ctx(ctx).
				Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
				Msg("[InMemoryDatabaseRepository][GetCount][Get][Uint64] failed to get")
		}

		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[InMemoryDatabaseRepository][GetCount][Get][Uint64] failed to get")
		err = gocerr.New(errorCode, err.Error())
		return 0, err
	}

	return value, nil
}

func (r *InMemoryDatabaseRepository[TEntity]) Keys(ctx context.Context, pattern string) ([]string, error) {
	var (
		span      trace.Span
		logFields map[string]interface{}
		keys      []string
		err       error
	)

	ctx, span = tracer.Start(ctx, "[InMemoryDatabaseRepository][Keys]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"pattern":   pattern,
	}

	keys, err = r.inMemoryDatabase.RedisClient.Keys(ctx, pattern).
		Result()
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[InMemoryDatabaseRepository][Keys][Keys][Result] failed to get keys")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[InMemoryDatabaseRepository][Keys][Keys][Result] failed to get keys")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return nil, err
	}

	return keys, nil
}

func (r *InMemoryDatabaseRepository[TEntity]) Lock(ctx context.Context, key string, expiration time.Duration) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		counter   int64
		err       error
	)

	ctx, span = tracer.Start(ctx, "[InMemoryDatabaseRepository][Lock]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid":  custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"key":        key,
		"expiration": expiration,
	}

	counter, err = r.inMemoryDatabase.RedisClient.Incr(ctx, key).
		Result()
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[InMemoryDatabaseRepository][Lock][Incr][Result] failed to increment")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[InMemoryDatabaseRepository][Lock][Incr][Result] failed to increment")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}
	logFields["counter"] = counter

	if expiration > 0 {
		_, err = r.inMemoryDatabase.RedisClient.Expire(
			ctx,
			key,
			expiration,
		).
			Result()
		if err != nil {
			log.Err(err).
				Ctx(ctx).
				Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
				Msg("[InMemoryDatabaseRepository][Lock][Expire][Result] failed to set expire")
			log.Debug().
				Ctx(ctx).
				Err(err).
				Fields(logFields).
				Msg("[InMemoryDatabaseRepository][Lock][Expire][Result] failed to set expire")
			err = gocerr.New(http.StatusInternalServerError, err.Error())
			return err
		}
	}

	if counter > 1 {
		err = gocerr.New(http.StatusConflict, fmt.Sprintf("%s is already locked", key))
		return err
	}

	return nil
}

func (r *InMemoryDatabaseRepository[TEntity]) Set(ctx context.Context, key string, value *TEntity, expiration time.Duration) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		rawValue  []byte
		err       error
	)

	ctx, span = tracer.Start(ctx, "[InMemoryDatabaseRepository][Set]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid":  custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"key":        key,
		"value":      value,
		"expiration": expiration,
	}

	rawValue, err = json.Marshal(value)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[InMemoryDatabaseRepository][Set][Marshal] failed to marshal value")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[InMemoryDatabaseRepository][Set][Marshal] failed to marshal value")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	_, err = r.inMemoryDatabase.RedisClient.Set(
		ctx,
		key,
		string(rawValue),
		expiration,
	).
		Result()
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[InMemoryDatabaseRepository][Set][Set][Result] failed to set")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[InMemoryDatabaseRepository][Set][Set][Result] failed to set")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	return nil
}

func (r *InMemoryDatabaseRepository[TEntity]) SetList(ctx context.Context, key string, values []TEntity, expiration time.Duration) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		rawValues []byte
		err       error
	)

	ctx, span = tracer.Start(ctx, "[InMemoryDatabaseRepository][SetList]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid":  custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"key":        key,
		"values":     values,
		"expiration": expiration,
	}

	rawValues, err = json.Marshal(values)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[InMemoryDatabaseRepository][SetList][Marshal] failed to marshal value")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[InMemoryDatabaseRepository][SetList][Marshal] failed to marshal value")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	_, err = r.inMemoryDatabase.RedisClient.Set(
		ctx,
		key,
		string(rawValues),
		expiration,
	).
		Result()
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[InMemoryDatabaseRepository][SetList][Set][Result] failed to set")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[InMemoryDatabaseRepository][SetList][Set][Result] failed to set")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	return nil
}

func (r *InMemoryDatabaseRepository[TEntity]) SetCount(ctx context.Context, key string, value uint64, expiration time.Duration) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		err       error
	)

	ctx, span = tracer.Start(ctx, "[InMemoryDatabaseRepository][SetCount]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid":  custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"key":        key,
		"value":      value,
		"expiration": expiration,
	}

	_, err = r.inMemoryDatabase.RedisClient.Set(
		ctx,
		key,
		value,
		expiration,
	).
		Result()
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[InMemoryDatabaseRepository][SetCount][Set][Result] failed to set")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[InMemoryDatabaseRepository][SetCount][Set][Result] failed to set")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	return nil
}

func (r *InMemoryDatabaseRepository[TEntity]) Unlock(ctx context.Context, key string) error {
	var (
		logFields map[string]interface{}
		err       error
	)

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"key":       key,
	}

	_, err = r.inMemoryDatabase.RedisClient.Del(ctx, key).
		Result()
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[InMemoryDatabaseRepository][Unlock][Delete][Result] failed to delete")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[InMemoryDatabaseRepository][Unlock][Delete][Result] failed to delete")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	return nil
}
