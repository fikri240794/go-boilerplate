//go:build wireinject
// +build wireinject

package event_consumer

import (
	"go-boilerplate/configs"
	"go-boilerplate/datasources"
	"go-boilerplate/internal/repositories"
	"go-boilerplate/internal/services"
	"go-boilerplate/transports/event_consumer/consumers"
	"go-boilerplate/transports/event_consumer/handlers"

	"github.com/google/wire"
)

func BuildEventConsumer(cfg *configs.Config) *EventConsumer {
	wire.Build(
		datasources.Provider,
		wire.Struct(new(datasources.Datasources), "*"),
		repositories.Provider,
		services.Provider,
		handlers.Provider,
		consumers.Provider,
		wire.Struct(new(consumers.Consumers), "*"),
		NewEventConsumer,
	)

	return &EventConsumer{}
}
