package middlewares

import (
	"fmt"
	"runtime/debug"

	"go-boilerplate/pkg/constants"
	custom_context "go-boilerplate/pkg/context"

	"github.com/fikri240794/gocerr"
	"github.com/fikri240794/gores"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type RecoverMiddleware struct{}

func NewRecoverMiddleware() *RecoverMiddleware {
	return &RecoverMiddleware{}
}

func (mw *RecoverMiddleware) Recover(c *fiber.Ctx) error {
	defer func() {
		var (
			r          interface{}
			logFields  map[string]interface{}
			responseVM *gores.ResponseVM[string]
			err        error
		)

		r = recover()
		if r != nil {
			err = gocerr.New(fiber.StatusInternalServerError, "panic")
			responseVM = gores.NewResponseVM[string]().
				SetErrorFromError(err)
			c.Status(responseVM.Code).
				JSON(responseVM)

			logFields = map[string]interface{}{
				"requestid":            custom_context.SafeCtxValue[string](c.UserContext(), constants.ContextKeyRequestID),
				"path":                 c.Path(),
				"method":               c.Method(),
				"request headers":      c.GetReqHeaders(),
				"request queries":      c.Queries(),
				"request body":         string(c.Body()),
				"response status code": c.Response().StatusCode(),
				"response headers":     c.GetRespHeaders(),
				"response body":        string(c.Response().Body()),
			}

			log.Debug().
				Err(err).
				Ctx(c.UserContext()).
				Fields(logFields).
				Msg("panic")

			fmt.Printf("panic: %v\n%s\n", r, debug.Stack())
		}
	}()

	return c.Next()
}
