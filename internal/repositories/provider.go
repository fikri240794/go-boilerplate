//go:build wireinject
// +build wireinject

package repositories

import "github.com/google/wire"

var Provider wire.ProviderSet = wire.NewSet(
	// guests
	NewGuestRepository,
	wire.Bind(new(IGuestRepository), new(*GuestRepository)),
	NewGuestCacheRepository,
	wire.Bind(new(IGuestCacheRepository), new(*GuestCacheRepository)),
	NewGuestEventProducerRepository,
	wire.Bind(new(IGuestEventProducerRepository), new(*GuestEventProducerRepository)),

	// webhook.site
	NewWebhookSiteRepository,
	wire.Bind(new(IWebhookSiteRepository), new(*WebhookSiteRepository)),
)
