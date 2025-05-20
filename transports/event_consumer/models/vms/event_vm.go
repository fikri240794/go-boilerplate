package vms

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type EventRequestVM[Tvm interface{}] struct {
	TracerPropagator map[string]string `json:"tracer_propagator"`
	Name             string            `json:"event_name"`
	Message          *Tvm              `json:"message"`
}

func (vm *EventRequestVM[Tvm]) ExtractTracerPropagator(ctx context.Context) context.Context {
	if len(vm.TracerPropagator) > 0 {
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(vm.TracerPropagator))
	}

	return ctx
}
