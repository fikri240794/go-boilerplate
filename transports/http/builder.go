//go:build wireinject
// +build wireinject

package http

import (
	"go-boilerplate/configs"
	"go-boilerplate/datasources"
	"go-boilerplate/internal/repositories"
	"go-boilerplate/internal/services"
	"go-boilerplate/transports/http/handlers"
	"go-boilerplate/transports/http/middlewares"

	"github.com/google/wire"
)

func BuildHTTPServer(cfg *configs.Config) *HTTPServer {
	wire.Build(
		datasources.Provider,
		wire.Struct(new(datasources.Datasources), "*"),
		repositories.Provider,
		services.Provider,
		handlers.Provider,
		wire.Struct(new(handlers.Handlers), "*"),
		middlewares.Provider,
		wire.Struct(new(middlewares.Middlewares), "*"),
		NewHTTPServer,
	)

	return &HTTPServer{}
}
