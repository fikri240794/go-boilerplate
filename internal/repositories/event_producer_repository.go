package repositories

import (
	"context"
	"go-boilerplate/datasources/event_producer"
	"go-boilerplate/internal/models/entities"
	"go-boilerplate/pkg/constants"
	custom_context "go-boilerplate/pkg/context"
	"go-boilerplate/pkg/tracer"
	"net/http"
	"time"

	"github.com/fikri240794/gocerr"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

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
		"requestid": custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID),
		"topic":     topic,
		"message":   message,
	}

	bMessage, err = json.Marshal(message)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[EventProducerRepository][Publish][Marshal] failed to marshal message")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[EventProducerRepository][Publish][Marshal] failed to marshal message")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[EventProducerRepository][Publish] publishing message")

	err = r.eventProducer.NSQProducer.Publish(topic, bMessage)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[EventProducerRepository][Publish][Publish] failed to publish message")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[EventProducerRepository][Publish][Publish] failed to publish message")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[EventProducerRepository][Publish] message published")

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
		"requestid": custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID),
		"topic":     topic,
		"delay":     delay,
		"message":   message,
	}

	bMessage, err = json.Marshal(message)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[EventProducerRepository][PublishWithDelay][Marshal] failed to marshal message")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[EventProducerRepository][PublishWithDelay][Marshal] failed to marshal message")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[EventProducerRepository][PublishWithDelay] publishing message")

	err = r.eventProducer.NSQProducer.DeferredPublish(topic, delay, bMessage)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID)).
			Msg("[EventProducerRepository][PublishWithDelay][DeferredPublish] failed to publish message")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[EventProducerRepository][PublishWithDelay][DeferredPublish] failed to publish message")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[EventProducerRepository][PublishWithDelay] message published")

	return nil
}
