package webhook_site_http_client

import (
	"go-boilerplate/configs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWebhookSiteHTTPClient(t *testing.T) {
	tests := []struct {
		name     string
		config   func() *configs.Config
		validate func(t *testing.T, client *WebhookSiteHTTPClient)
	}{
		{
			name: "create client with valid base URL",
			config: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.WebhookSiteHTTPClient.BaseURL = "https://webhook.site"
				return cfg
			},
			validate: func(t *testing.T, client *WebhookSiteHTTPClient) {
				assert.NotNil(t, client)
				assert.NotNil(t, client.HttpClient)
				assert.Equal(t, "https://webhook.site", client.HttpClient.BaseURL)
			},
		},
		{
			name: "create client with empty base URL",
			config: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.WebhookSiteHTTPClient.BaseURL = ""
				return cfg
			},
			validate: func(t *testing.T, client *WebhookSiteHTTPClient) {
				assert.NotNil(t, client)
				assert.NotNil(t, client.HttpClient)
				assert.Equal(t, "", client.HttpClient.BaseURL)
			},
		},
		{
			name: "create client with localhost URL",
			config: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.WebhookSiteHTTPClient.BaseURL = "http://localhost:8080"
				return cfg
			},
			validate: func(t *testing.T, client *WebhookSiteHTTPClient) {
				assert.NotNil(t, client)
				assert.NotNil(t, client.HttpClient)
				assert.Equal(t, "http://localhost:8080", client.HttpClient.BaseURL)
			},
		},
		{
			name: "create client with URL containing path",
			config: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.WebhookSiteHTTPClient.BaseURL = "https://api.example.com/v1"
				return cfg
			},
			validate: func(t *testing.T, client *WebhookSiteHTTPClient) {
				assert.NotNil(t, client)
				assert.NotNil(t, client.HttpClient)
				assert.Equal(t, "https://api.example.com/v1", client.HttpClient.BaseURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewWebhookSiteHTTPClient(tt.config())

			if tt.validate != nil {
				tt.validate(t, client)
			}
		})
	}
}
