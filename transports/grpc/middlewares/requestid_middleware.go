package middlewares

import (
	"context"
	"go-boilerplate/pkg/constants"
	"go-boilerplate/pkg/grpc_metadata"
	"go-boilerplate/pkg/tracer"
	"go-boilerplate/pkg/uuid"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type RequestIDMiddleware struct{}

func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

func (mw *RequestIDMiddleware) Generate(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	var (
		span         trace.Span
		md           metadata.MD
		ok           bool
		requestID    string
		res          interface{}
		err          error
		setHeaderErr error
	)

	ctx, span = tracer.Start(ctx, "[RequestIDMiddleware][Generate]")
	defer span.End()

	md, ok = metadata.FromIncomingContext(ctx)
	if ok {
		requestID = grpc_metadata.MDGetString(md, constants.HeaderKeyRequestID)
	}

	if requestID == "" {
		requestID = uuid.NewV7().String()
		ctx = context.WithValue(ctx, constants.ContextKeyRequestID, requestID)
	}

	res, err = handler(ctx, req)

	md = metadata.New(map[string]string{
		constants.HeaderKeyRequestID: requestID,
	})
	setHeaderErr = grpc.SetHeader(ctx, md)
	if setHeaderErr != nil {
		log.Err(setHeaderErr).
			Ctx(ctx).
			Interface("req", req).
			Interface("unary server info", info).
			Interface("res", res).
			AnErr("err", err).
			Msg("[RequestIDMiddleware][Generate][SetHeader] failed to set response header")
	}

	return res, err
}
