//go:build wireinject
// +build wireinject

package datasources

import (
	"go-boilerplate/datasources/boilerplate_database"
	"go-boilerplate/datasources/event_producer"
	"go-boilerplate/datasources/in_memory_database"
	"go-boilerplate/datasources/webhook_site_http_client"

	"github.com/google/wire"
)

var Provider wire.ProviderSet = wire.NewSet(
	boilerplate_database.Connect,
	in_memory_database.Connect,
	event_producer.Connect,
	webhook_site_http_client.NewWebhookSiteHTTPClient,
)
