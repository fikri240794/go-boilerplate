package tracer

import (
	"context"

	"github.com/fikri240794/goteletracer"
	"go.opentelemetry.io/otel/trace"
)

var globalTracer trace.Tracer

func init() {
	globalTracer = goteletracer.NewTracer(nil)
}

func NewTracer(cfg *goteletracer.Config) {
	globalTracer = goteletracer.NewTracer(cfg)
}

func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return globalTracer.Start(ctx, spanName, opts...)
}
