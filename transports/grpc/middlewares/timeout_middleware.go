package middlewares

import (
	"context"
	"go-boilerplate/configs"
	"go-boilerplate/pkg/tracer"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TimeoutMiddleware struct {
	cfg *configs.Config
}

func NewTimeoutMiddleware(cfg *configs.Config) *TimeoutMiddleware {
	return &TimeoutMiddleware{
		cfg: cfg,
	}
}

func (mw *TimeoutMiddleware) Timeout(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	var (
		span       trace.Span
		ctxTimeout context.Context
		cancel     context.CancelFunc
		resChan    chan interface{}
		errChan    chan error
		res        interface{}
		err        error
	)

	ctx, span = tracer.Start(ctx, "[TimeoutMiddleware][Timeout]")
	defer span.End()

	ctxTimeout, cancel = context.WithTimeout(ctx, mw.cfg.Server.GRPC.RequestTimeout)
	defer cancel()

	resChan = make(chan interface{})
	errChan = make(chan error)

	go func() {
		var (
			responseHandler interface{}
			errHandler      error
		)

		responseHandler, errHandler = handler(ctxTimeout, req)
		resChan <- responseHandler
		errChan <- errHandler
	}()

	select {
	case <-ctxTimeout.Done():
		return nil, status.Error(codes.DeadlineExceeded, codes.DeadlineExceeded.String())
	case res = <-resChan:
		err = <-errChan
		return res, err
	}
}
