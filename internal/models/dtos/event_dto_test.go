package dtos

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventRequestDTO_ExtractTracerPropagator(t *testing.T) {
	tests := []struct {
		name             string
		tracerPropagator map[string]string
		validate         func(t *testing.T, originalCtx, resultCtx context.Context, tracerPropagator map[string]string)
	}{
		{
			name:             "empty tracer propagator map",
			tracerPropagator: map[string]string{},
			validate: func(t *testing.T, originalCtx, resultCtx context.Context, tracerPropagator map[string]string) {
				assert.Equal(t, originalCtx, resultCtx, "Expected context to remain unchanged when TracerPropagator is empty")
			},
		},
		{
			name:             "nil tracer propagator map",
			tracerPropagator: nil,
			validate: func(t *testing.T, originalCtx, resultCtx context.Context, tracerPropagator map[string]string) {
				assert.Equal(t, originalCtx, resultCtx, "Expected context to remain unchanged when TracerPropagator is nil")
			},
		},
		{
			name: "tracer propagator with valid trace data",
			tracerPropagator: map[string]string{
				"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
				"tracestate":  "congo=t61rcWkgMzE",
			},
			validate: func(t *testing.T, originalCtx, resultCtx context.Context, tracerPropagator map[string]string) {
				assert.NotNil(t, resultCtx, "Expected valid context to be returned")
			},
		},
		{
			name: "tracer propagator with custom headers",
			tracerPropagator: map[string]string{
				"x-custom-trace": "custom-value",
				"x-request-id":   "req-123",
			},
			validate: func(t *testing.T, originalCtx, resultCtx context.Context, tracerPropagator map[string]string) {
				assert.NotNil(t, resultCtx, "Expected valid context to be returned")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type dummyMsg struct{ Value string }

			dto := &EventRequestDTO[dummyMsg]{
				TracerPropagator: tt.tracerPropagator,
				Name:             "test_event",
				Message:          dummyMsg{Value: "test"},
			}

			originalCtx := context.Background()
			resultCtx := dto.ExtractTracerPropagator(originalCtx)

			assert.NotNil(t, resultCtx, "ExtractTracerPropagator should never return nil context")

			tt.validate(t, originalCtx, resultCtx, tt.tracerPropagator)
		})
	}
}
