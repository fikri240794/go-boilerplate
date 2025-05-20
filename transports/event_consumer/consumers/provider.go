//go:build wireinject
// +build wireinject

package consumers

import "github.com/google/wire"

var Provider wire.ProviderSet = wire.NewSet(
	// guests
	NewGuestConsumer,
)
