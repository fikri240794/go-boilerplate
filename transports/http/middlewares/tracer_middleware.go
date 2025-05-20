package middlewares

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type TracerMiddleware struct{}

func NewTracerMiddleware() *TracerMiddleware {
	return &TracerMiddleware{}
}

func (mw *TracerMiddleware) Start(c *fiber.Ctx) error {
	var ctx context.Context = otel.GetTextMapPropagator().Extract(c.UserContext(), propagation.HeaderCarrier(c.GetReqHeaders()))
	c.SetUserContext(ctx)
	return c.Next()
}
