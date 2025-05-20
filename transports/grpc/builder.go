//go:build wireinject
// +build wireinject

package grpc

import (
	"go-boilerplate/configs"
	"go-boilerplate/datasources"
	"go-boilerplate/internal/repositories"
	"go-boilerplate/internal/services"
	"go-boilerplate/transports/grpc/handlers"
	"go-boilerplate/transports/grpc/middlewares"

	"github.com/google/wire"
)

func BuildGRPCServer(cfg *configs.Config) *GRPCServer {
	wire.Build(
		datasources.Provider,
		wire.Struct(new(datasources.Datasources), "*"),
		repositories.Provider,
		services.Provider,
		handlers.NewImplementedBoilerplateServer,
		middlewares.Provider,
		wire.Struct(new(middlewares.Middlewares), "*"),
		NewGRPCServer,
	)

	return &GRPCServer{}
}
