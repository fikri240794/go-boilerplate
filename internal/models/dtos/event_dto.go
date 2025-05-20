package dtos

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type EventRequestDTO[Tdto interface{}] struct {
	TracerPropagator map[string]string `json:"tracer_propagator"`
	Name             string            `json:"event_name"`
	Message          Tdto              `json:"message"`
}

func (dto *EventRequestDTO[Tdto]) ExtractTracerPropagator(ctx context.Context) context.Context {
	if len(dto.TracerPropagator) > 0 {
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(dto.TracerPropagator))
	}

	return ctx
}
