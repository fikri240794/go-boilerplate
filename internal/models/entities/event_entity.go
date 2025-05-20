package entities

import (
	"context"

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
	e.TracerPropagator = map[string]string{}
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(e.TracerPropagator))

	return ctx, e
}
