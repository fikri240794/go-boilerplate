package middlewares

import (
	"context"
	"fmt"
	"go-boilerplate/pkg/constants"
	custom_context "go-boilerplate/pkg/context"
	"go-boilerplate/pkg/tracer"
	"net/http"
	"time"

	"github.com/fikri240794/gostacode"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type LogMiddleware struct{}

func NewLogMiddleware() *LogMiddleware {
	return &LogMiddleware{}
}

func (mw *LogMiddleware) Log(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	var (
		span         trace.Span
		logLevel     zerolog.Level
		processStart time.Time
		processEnd   time.Time
		latency      time.Duration
		res          interface{}
		grpcStatus   *status.Status
		ok           bool
		logFields    map[string]interface{}
		err          error
	)

	ctx, span = tracer.Start(ctx, "[LogMiddleware][Log]")
	defer span.End()

	logLevel = zerolog.InfoLevel
	processStart = time.Now()
	res, err = handler(ctx, req)
	processEnd = time.Now()
	latency = processEnd.Sub(processStart)

	if err != nil {
		grpcStatus, ok = status.FromError(err)
		if ok && gostacode.HTTPStatusCodeFromGRPCCode(grpcStatus.Code()) >= http.StatusInternalServerError {
			logLevel = zerolog.ErrorLevel
		}
	}

	logFields = map[string]interface{}{
		"requestid":         custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID),
		"req":               req,
		"unary server info": info,
		"res":               res,
		"err":               err,
		"latency":           fmt.Sprintf("%.3f ms", (float64(latency) / float64(time.Millisecond))),
	}

	log.WithLevel(logLevel).
		Ctx(ctx).
		Fields(logFields).
		Msg("request response")

	return res, err
}
