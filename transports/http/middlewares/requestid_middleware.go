package middlewares

import (
	"context"
	"go-boilerplate/pkg/constants"
	"go-boilerplate/pkg/tracer"
	"go-boilerplate/pkg/uuid"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

const ()

type RequestIDMiddleware struct{}

func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

func (mw *RequestIDMiddleware) Generate(c *fiber.Ctx) error {
	var (
		ctx       context.Context
		span      trace.Span
		requestID string
	)

	ctx = c.UserContext()

	ctx, span = tracer.Start(ctx, "[RequestIDMiddleware][Generate]")
	defer span.End()

	requestID = c.Get(constants.HeaderKeyRequestID)
	if requestID == "" {
		requestID = uuid.NewV7().String()
	}

	ctx = context.WithValue(ctx, constants.ContextKeyRequestID, requestID)
	c.SetUserContext(ctx)
	c.Set(constants.HeaderKeyRequestID, requestID)

	return c.Next()
}
