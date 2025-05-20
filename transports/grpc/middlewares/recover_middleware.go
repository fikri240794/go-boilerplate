package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"go-boilerplate/pkg/grpc_error"

	"github.com/fikri240794/gocerr"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type RecoverMiddleware struct{}

func NewRecoverMiddleware() *RecoverMiddleware {
	return &RecoverMiddleware{}
}

func (mw *RecoverMiddleware) Recover(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (res interface{}, err error) {
	defer func() {
		var (
			r         interface{}
			logFields map[string]interface{}
		)

		r = recover()
		if r != nil {
			err = gocerr.New(http.StatusInternalServerError, fmt.Sprintf("panic: %v", r))

			logFields = map[string]interface{}{
				"req":               req,
				"unary server info": info,
			}

			log.Err(err).
				Ctx(ctx).
				Fields(logFields).
				Msg("panic")

			fmt.Printf("panic: %v\n%s\n", r, debug.Stack())

			err = grpc_error.FromError(err)
		}
	}()

	res, err = handler(ctx, req)

	return
}
