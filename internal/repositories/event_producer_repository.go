package repositories

import (
	"context"
	"go-boilerplate/datasources/event_producer"
	"go-boilerplate/internal/models/entities"
	"go-boilerplate/pkg/tracer"
	"net/http"
	"time"

	"github.com/fikri240794/gocerr"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

//mockery:generate: true
//mockery:structname: EventProducerRepositoryMock
//mockery:filename: event_producer_repository_mock.go
//mockery:output: internal/repositories/mocks/
type IEventProducerRepository[TEntity interface{}] interface {
	Publish(
		ctx context.Context,
		topic string,
		message *entities.EventEntity[TEntity],
	) error

	PublishWithDelay(
		ctx context.Context,
		topic string,
		delay time.Duration,
		message *entities.EventEntity[TEntity],
	) error

	PublishBulk(
		ctx context.Context,
		topic string,
		message *entities.EventEntity[[]TEntity],
	) error

	PublishBulkWithDelay(
		ctx context.Context,
		topic string,
		delay time.Duration,
		message *entities.EventEntity[[]TEntity],
	) error
}

type EventProducerRepository[TEntity interface{}] struct {
	eventProducer *event_producer.EventProducer
}

func NewEventProducerRepository[TEntity interface{}](eventProducer *event_producer.EventProducer) *EventProducerRepository[TEntity] {
	return &EventProducerRepository[TEntity]{
		eventProducer: eventProducer,
	}
}

func (r *EventProducerRepository[TEntity]) Publish(
	ctx context.Context,
	topic string,
	message *entities.EventEntity[TEntity],
) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		bMessage  []byte
		err       error
	)

	ctx, span = tracer.Start(ctx, "[EventProducerRepository][Publish]")
	defer span.End()

	ctx, message = message.InjectTracerPropagator(ctx)

	logFields = map[string]interface{}{
		"topic":   topic,
		"message": message,
	}

	bMessage, err = json.Marshal(message)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[EventProducerRepository][Publish][Marshal] failed to marshal message")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	err = r.eventProducer.NSQProducer.Publish(topic, bMessage)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[EventProducerRepository][Publish][Publish] failed to publish message")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	return nil
}

func (r *EventProducerRepository[TEntity]) PublishWithDelay(
	ctx context.Context,
	topic string,
	delay time.Duration,
	message *entities.EventEntity[TEntity],
) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		bMessage  []byte
		err       error
	)

	ctx, span = tracer.Start(ctx, "[EventProducerRepository][PublishWithDelay]")
	defer span.End()

	ctx, message = message.InjectTracerPropagator(ctx)

	logFields = map[string]interface{}{
		"topic":   topic,
		"delay":   delay,
		"message": message,
	}

	bMessage, err = json.Marshal(message)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[EventProducerRepository][PublishWithDelay][Marshal] failed to marshal message")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	err = r.eventProducer.NSQProducer.DeferredPublish(topic, delay, bMessage)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[EventProducerRepository][PublishWithDelay][DeferredPublish] failed to publish message")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	return nil
}

func (r *EventProducerRepository[TEntity]) PublishBulk(
	ctx context.Context,
	topic string,
	message *entities.EventEntity[[]TEntity],
) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		bMessage  []byte
		err       error
	)

	ctx, span = tracer.Start(ctx, "[EventProducerRepository][PublishBulk]")
	defer span.End()

	ctx, message = message.InjectTracerPropagator(ctx)

	logFields = map[string]interface{}{
		"topic":   topic,
		"message": message,
	}

	bMessage, err = json.Marshal(message)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[EventProducerRepository][PublishBulk][Marshal] failed to marshal message")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	err = r.eventProducer.NSQProducer.Publish(topic, bMessage)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[EventProducerRepository][PublishBulk][Publish] failed to publish message")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	return nil
}

func (r *EventProducerRepository[TEntity]) PublishBulkWithDelay(
	ctx context.Context,
	topic string,
	delay time.Duration,
	message *entities.EventEntity[[]TEntity],
) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		bMessage  []byte
		err       error
	)

	ctx, span = tracer.Start(ctx, "[EventProducerRepository][PublishBulkWithDelay]")
	defer span.End()

	ctx, message = message.InjectTracerPropagator(ctx)

	logFields = map[string]interface{}{
		"topic":   topic,
		"delay":   delay,
		"message": message,
	}

	bMessage, err = json.Marshal(message)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[EventProducerRepository][PublishBulkWithDelay][Marshal] failed to marshal message")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	err = r.eventProducer.NSQProducer.DeferredPublish(topic, delay, bMessage)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[EventProducerRepository][PublishBulkWithDelay][DeferredPublish] failed to publish message")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	return nil
}
