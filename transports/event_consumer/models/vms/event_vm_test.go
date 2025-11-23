package vms

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TestEventRequestVM_ExtractTracerPropagator(t *testing.T) {
	tests := []struct {
		name            string
		setupVM         func() *EventRequestVM[string]
		setupContext    func() context.Context
		setupPropagator func()
		validate        func(t *testing.T, resultCtx context.Context)
	}{
		{
			name: "should_extract_tracer_propagator_when_map_is_not_empty",
			setupVM: func() *EventRequestVM[string] {
				return &EventRequestVM[string]{
					TracerPropagator: map[string]string{
						"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
						"tracestate":  "congo=t61rcWkgMzE",
					},
					Name:    "test_event",
					Message: stringPtr("test message"),
				}
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			setupPropagator: func() {
				otel.SetTextMapPropagator(propagation.TraceContext{})
			},
			validate: func(t *testing.T, resultCtx context.Context) {
				assert.NotNil(t, resultCtx)

				spanCtx := trace.SpanContextFromContext(resultCtx)
				assert.True(t, spanCtx.IsValid(), "Span context should be valid after extraction")
			},
		},
		{
			name: "should_return_original_context_when_tracer_propagator_is_empty",
			setupVM: func() *EventRequestVM[string] {
				return &EventRequestVM[string]{
					TracerPropagator: map[string]string{},
					Name:             "test_event",
					Message:          stringPtr("test message"),
				}
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			setupPropagator: func() {
				otel.SetTextMapPropagator(propagation.TraceContext{})
			},
			validate: func(t *testing.T, resultCtx context.Context) {
				assert.NotNil(t, resultCtx)
				spanCtx := trace.SpanContextFromContext(resultCtx)
				assert.False(t, spanCtx.IsValid(), "Span context should not be valid when propagator is empty")
			},
		},
		{
			name: "should_return_original_context_when_tracer_propagator_is_nil",
			setupVM: func() *EventRequestVM[string] {
				return &EventRequestVM[string]{
					TracerPropagator: nil,
					Name:             "test_event",
					Message:          stringPtr("test message"),
				}
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			setupPropagator: func() {
				otel.SetTextMapPropagator(propagation.TraceContext{})
			},
			validate: func(t *testing.T, resultCtx context.Context) {
				assert.NotNil(t, resultCtx)
				spanCtx := trace.SpanContextFromContext(resultCtx)
				assert.False(t, spanCtx.IsValid(), "Span context should not be valid when propagator is nil")
			},
		},
		{
			name: "should_handle_context_with_existing_values",
			setupVM: func() *EventRequestVM[string] {
				return &EventRequestVM[string]{
					TracerPropagator: map[string]string{
						"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
					},
					Name:    "test_event",
					Message: stringPtr("test message"),
				}
			},
			setupContext: func() context.Context {
				ctx := context.Background()
				type contextKey string
				return context.WithValue(ctx, contextKey("test_key"), "test_value")
			},
			setupPropagator: func() {
				otel.SetTextMapPropagator(propagation.TraceContext{})
			},
			validate: func(t *testing.T, resultCtx context.Context) {
				assert.NotNil(t, resultCtx)

				spanCtx := trace.SpanContextFromContext(resultCtx)
				assert.True(t, spanCtx.IsValid())
			},
		},
		{
			name: "should_extract_with_composite_propagator",
			setupVM: func() *EventRequestVM[string] {
				return &EventRequestVM[string]{
					TracerPropagator: map[string]string{
						"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
						"baggage":     "key1=value1,key2=value2",
					},
					Name:    "test_event",
					Message: stringPtr("test message"),
				}
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			setupPropagator: func() {
				otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
					propagation.TraceContext{},
					propagation.Baggage{},
				))
			},
			validate: func(t *testing.T, resultCtx context.Context) {
				assert.NotNil(t, resultCtx)
				spanCtx := trace.SpanContextFromContext(resultCtx)
				assert.True(t, spanCtx.IsValid())
			},
		},
		{
			name: "should_handle_invalid_traceparent_format",
			setupVM: func() *EventRequestVM[string] {
				return &EventRequestVM[string]{
					TracerPropagator: map[string]string{
						"traceparent": "invalid-format",
					},
					Name:    "test_event",
					Message: stringPtr("test message"),
				}
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			setupPropagator: func() {
				otel.SetTextMapPropagator(propagation.TraceContext{})
			},
			validate: func(t *testing.T, resultCtx context.Context) {
				assert.NotNil(t, resultCtx)

				spanCtx := trace.SpanContextFromContext(resultCtx)
				assert.False(t, spanCtx.IsValid())
			},
		},
		{
			name: "should_work_with_generic_type_int",
			setupVM: func() *EventRequestVM[string] {
				vm := &EventRequestVM[int]{
					TracerPropagator: map[string]string{
						"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
					},
					Name:    "test_event",
					Message: intPtr(123),
				}

				return &EventRequestVM[string]{
					TracerPropagator: vm.TracerPropagator,
					Name:             vm.Name,
					Message:          stringPtr("converted"),
				}
			},
			setupContext: func() context.Context {
				return context.Background()
			},
			setupPropagator: func() {
				otel.SetTextMapPropagator(propagation.TraceContext{})
			},
			validate: func(t *testing.T, resultCtx context.Context) {
				assert.NotNil(t, resultCtx)
				spanCtx := trace.SpanContextFromContext(resultCtx)
				assert.True(t, spanCtx.IsValid())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.setupPropagator()

			vm := tt.setupVM()
			ctx := tt.setupContext()

			resultCtx := vm.ExtractTracerPropagator(ctx)

			tt.validate(t, resultCtx)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
