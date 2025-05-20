package middlewares

import (
	"context"
	"fmt"
	"go-boilerplate/pkg/constants"
	custom_context "go-boilerplate/pkg/context"
	"go-boilerplate/pkg/tracer"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

type LogMiddleware struct{}

func NewLogMiddleware() *LogMiddleware {
	return &LogMiddleware{}
}

func (mw *LogMiddleware) Log(c *fiber.Ctx) error {
	var (
		ctx          context.Context
		span         trace.Span
		logLevel     zerolog.Level
		processStart time.Time
		processEnd   time.Time
		latency      time.Duration
		logFields    map[string]interface{}
		err          error
	)

	ctx = c.UserContext()

	ctx, span = tracer.Start(ctx, "[LogMiddleware][Log]")
	defer span.End()

	c.SetUserContext(ctx)
	logLevel = zerolog.InfoLevel
	processStart = time.Now()
	err = c.Next()
	processEnd = time.Now()
	latency = processEnd.Sub(processStart)

	if err != nil || c.Response().StatusCode() >= fiber.StatusInternalServerError {
		logLevel = zerolog.ErrorLevel
	}

	if c.Response().StatusCode() >= fiber.StatusBadRequest && c.Response().StatusCode() < fiber.StatusInternalServerError {
		logLevel = zerolog.WarnLevel
	}

	logFields = map[string]interface{}{
		"requestid":            custom_context.SafeCtxValue[string](ctx, constants.ContextKeyRequestID),
		"path":                 c.Path(),
		"method":               c.Method(),
		"response status code": c.Response().StatusCode(),
		"latency":              fmt.Sprintf("%.3f ms", (float64(latency) / float64(time.Millisecond))),
	}

	log.WithLevel(logLevel).
		Ctx(ctx).
		Fields(logFields).
		Msg("request response")

	logFields["request headers"] = c.GetReqHeaders()
	logFields["request queries"] = c.Queries()
	logFields["request body"] = string(c.Body())
	logFields["response headers"] = c.GetRespHeaders()
	logFields["response body"] = string(c.Response().Body())

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("request response")

	return err
}
