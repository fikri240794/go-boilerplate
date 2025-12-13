package logger

import (
	"context"
	"go-boilerplate/pkg/constants"

	custom_context "go-boilerplate/pkg/context"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

// ContextHook is a zerolog hook that automatically extracts and injects
// requestid and traceid from context into log entries
type ContextHook struct{}

// NewContextHook creates a new ContextHook
func NewContextHook() ContextHook {
	return ContextHook{}
}

// Run implements zerolog.Hook interface
// This is called for every log event and automatically adds context values
func (h ContextHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	var (
		ctx       context.Context
		requestID string
		span      trace.Span
	)

	// Get context from the event
	ctx = e.GetCtx()

	// Extract and add request ID from context
	requestID = custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)
	if requestID != "" {
		e.Str(string(constants.ContextKeyRequestID), requestID)
	}

	// Extract and add trace ID from OpenTelemetry span
	span = trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		e.Str(string(constants.ContextKeyTraceID), span.SpanContext().TraceID().String())
	}
}
