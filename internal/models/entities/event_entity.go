package entities

import (
	"context"
	"go-boilerplate/pkg/constants"

	custom_context "go-boilerplate/pkg/context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type EventEntity[TEntity interface{}] struct {
	TracerPropagator map[string]string `json:"tracer_propagator"`
	Name             string            `json:"event_name"`
	Message          *TEntity          `json:"message"`
}

func NewEventEntity[TEntity interface{}](name string, message *TEntity) *EventEntity[TEntity] {
	return &EventEntity[TEntity]{
		Name:    name,
		Message: message,
	}
}

func (e *EventEntity[TEntity]) InjectTracerPropagator(ctx context.Context) (context.Context, *EventEntity[TEntity]) {
	var requestID string

	e.TracerPropagator = map[string]string{}

	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(e.TracerPropagator))

	requestID = custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)
	if requestID != "" {
		e.TracerPropagator[string(constants.ContextKeyRequestID)] = requestID
	}

	return ctx, e
}
