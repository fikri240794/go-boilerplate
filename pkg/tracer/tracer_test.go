package tracer

import (
	"context"
	"testing"

	"github.com/fikri240794/goteletracer"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func TestNewTracer(t *testing.T) {
	originalTracer := globalTracer
	defer func() {
		globalTracer = originalTracer
	}()

	tests := []struct {
		name   string
		config *goteletracer.Config
	}{
		{
			name:   "nil config should initialize tracer",
			config: nil,
		},
		{
			name: "with config should initialize tracer",
			config: &goteletracer.Config{
				ServiceName:         "test-service",
				ExporterGRPCAddress: "localhost:4317",
			},
		},
		{
			name: "with partial config should initialize tracer",
			config: &goteletracer.Config{
				ServiceName: "test-service",
			},
		},
		{
			name: "with empty config should initialize tracer",
			config: &goteletracer.Config{
				ServiceName:         "",
				ExporterGRPCAddress: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalTracer = nil

			NewTracer(tt.config)

			assert.NotNil(t, globalTracer, "Expected globalTracer to be initialized")
		})
	}
}

func TestStart(t *testing.T) {
	originalTracer := globalTracer
	defer func() {
		globalTracer = originalTracer
	}()

	NewTracer(nil)

	tests := []struct {
		name     string
		ctx      context.Context
		spanName string
		opts     []trace.SpanStartOption
	}{
		{
			name:     "start span with background context",
			ctx:      context.Background(),
			spanName: "test-span",
			opts:     nil,
		},
		{
			name:     "start span with empty span name",
			ctx:      context.Background(),
			spanName: "",
			opts:     nil,
		},
		{
			name:     "start span with TODO context",
			ctx:      context.TODO(),
			spanName: "test-span-todo",
			opts:     nil,
		},
		{
			name:     "start span with span options",
			ctx:      context.Background(),
			spanName: "test-span-with-options",
			opts:     []trace.SpanStartOption{},
		},
		{
			name:     "start span with special characters in name",
			ctx:      context.Background(),
			spanName: "test-span-@#$%^&*()",
			opts:     nil,
		},
		{
			name:     "start span with long span name",
			ctx:      context.Background(),
			spanName: "this-is-a-very-long-span-name-that-should-still-work-without-any-issues",
			opts:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultCtx, span := Start(tt.ctx, tt.spanName, tt.opts...)

			assert.NotNil(t, resultCtx, "Expected non-nil context")
			assert.NotNil(t, span, "Expected non-nil span")

			assert.NotPanics(t, func() {
				span.End()
			}, "Unexpected panic when ending span")
		})
	}
}
