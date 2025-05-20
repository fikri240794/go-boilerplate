//go:build wireinject
// +build wireinject

package handlers

import "github.com/google/wire"

var Provider wire.ProviderSet = wire.NewSet(
	// guests
	NewGuestHandler,
)
