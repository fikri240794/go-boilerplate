package webhook_site_http_client

import (
	"go-boilerplate/configs"

	"github.com/go-resty/resty/v2"
)

type WebhookSiteHTTPClient struct {
	HttpClient *resty.Client
}

func NewWebhookSiteHTTPClient(cfg *configs.Config) *WebhookSiteHTTPClient {
	var httpClient *WebhookSiteHTTPClient = &WebhookSiteHTTPClient{
		HttpClient: resty.New().
			SetBaseURL(cfg.Datasource.WebhookSiteHTTPClient.BaseURL),
	}

	return httpClient
}
