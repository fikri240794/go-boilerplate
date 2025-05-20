package middlewares

import (
	"context"
	"go-boilerplate/configs"
	"go-boilerplate/pkg/tracer"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

type TimeoutMiddleware struct {
	cfg *configs.Config
}

func NewTimeoutMiddleware(cfg *configs.Config) *TimeoutMiddleware {
	return &TimeoutMiddleware{
		cfg: cfg,
	}
}

func (mw *TimeoutMiddleware) Timeout(c *fiber.Ctx) error {
	var (
		ctx    context.Context
		span   trace.Span
		cancel context.CancelFunc
	)

	ctx = c.UserContext()

	ctx, span = tracer.Start(ctx, "[TimeoutMiddleware][Timeout]")
	defer span.End()

	ctx, cancel = context.WithTimeout(ctx, mw.cfg.Server.HTTP.RequestTimeout)
	defer cancel()

	c.SetUserContext(ctx)

	return c.Next()
}
