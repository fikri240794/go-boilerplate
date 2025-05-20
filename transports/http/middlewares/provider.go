//go:build wireinject
// +build wireinject

package middlewares

import "github.com/google/wire"

var Provider wire.ProviderSet = wire.NewSet(
	NewRecoverMiddleware,
	NewTracerMiddleware,
	NewRequestIDMiddleware,
	NewLogMiddleware,
	NewTimeoutMiddleware,
)
