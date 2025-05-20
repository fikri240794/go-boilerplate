package middlewares

import (
	"context"
	"go-boilerplate/pkg/grpc_metadata"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type TracerMiddleware struct{}

func NewTracerMiddleware() *TracerMiddleware {
	return &TracerMiddleware{}
}

func (mw *TracerMiddleware) Start(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	var (
		md             metadata.MD
		ok             bool
		traceparentMap map[string]string
		res            interface{}
		err            error
	)

	md, ok = metadata.FromIncomingContext(ctx)
	if ok {
		traceparentMap = grpc_metadata.MDToMapString(md)
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(traceparentMap))
	}

	res, err = handler(ctx, req)

	return res, err
}
