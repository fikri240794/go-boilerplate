//go:build wireinject
// +build wireinject

package services

import "github.com/google/wire"

var Provider wire.ProviderSet = wire.NewSet(
	// guests
	NewGuestService,
	wire.Bind(new(IGuestService), new(*GuestService)),
)
