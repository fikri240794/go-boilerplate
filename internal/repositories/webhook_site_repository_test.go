package repositories

import (
	"context"
	"go-boilerplate/configs"
	"go-boilerplate/datasources/webhook_site_http_client"
	"go-boilerplate/internal/models/entities"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func Test_NewWebhookSiteRepository(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *configs.Config
		httpClient *webhook_site_http_client.WebhookSiteHTTPClient
		expectNil  bool
	}{
		{
			name: "create webhook site repository with config and http client",
			cfg: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.WebhookSiteHTTPClient.BaseURL = "https://webhook.site"
				cfg.Datasource.WebhookSiteHTTPClient.Endpoint.Webhook = "/test-webhook"
				return cfg
			}(),
			httpClient: &webhook_site_http_client.WebhookSiteHTTPClient{
				HttpClient: resty.New(),
			},
			expectNil: false,
		},
		{
			name:       "create webhook site repository without dependencies",
			cfg:        nil,
			httpClient: nil,
			expectNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewWebhookSiteRepository(tt.cfg, tt.httpClient)

			if tt.expectNil {
				assert.Nil(t, repo, "NewWebhookSiteRepository() expected nil")
			} else {
				assert.NotNil(t, repo, "NewWebhookSiteRepository() expected not nil")
				assert.Equal(t, tt.cfg, repo.cfg, "NewWebhookSiteRepository() cfg not set correctly")
				assert.Equal(t, tt.httpClient, repo.httpClient, "NewWebhookSiteRepository() httpClient not set correctly")
			}
		})
	}
}

func Test_WebhookSiteRepository_SendWebhook(t *testing.T) {
	tests := []struct {
		name          string
		setupRepo     func(serverURL string) *WebhookSiteRepository
		requestData   *entities.GuestEventEntity
		mockServer    func() *httptest.Server
		expectError   bool
		validateError func(t *testing.T, err error)
	}{
		{
			name: "send webhook successfully",
			setupRepo: func(serverURL string) *WebhookSiteRepository {
				cfg := &configs.Config{}
				cfg.Datasource.WebhookSiteHTTPClient.BaseURL = serverURL
				cfg.Datasource.WebhookSiteHTTPClient.Endpoint.Webhook = "/webhook"
				httpClient := resty.New()
				httpClient.SetBaseURL(serverURL)
				return NewWebhookSiteRepository(cfg, &webhook_site_http_client.WebhookSiteHTTPClient{
					HttpClient: httpClient,
				})
			},
			requestData: &entities.GuestEventEntity{
				ID:        "550e8400-e29b-41d4-a716-446655440000",
				Name:      "Test Guest",
				Address:   "Test Address",
				CreatedAt: 1234567890,
				CreatedBy: "test_user",
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"success": true}`))
				}))
			},
			expectError: false,
			validateError: func(t *testing.T, err error) {
				assert.NoError(t, err, "SendWebhook() unexpected error")
			},
		},
		{
			name: "send webhook with http request error",
			setupRepo: func(serverURL string) *WebhookSiteRepository {
				cfg := &configs.Config{}
				cfg.Datasource.WebhookSiteHTTPClient.BaseURL = "http://invalid-url-that-does-not-exist.local"
				cfg.Datasource.WebhookSiteHTTPClient.Endpoint.Webhook = "/webhook"
				return NewWebhookSiteRepository(cfg, &webhook_site_http_client.WebhookSiteHTTPClient{
					HttpClient: resty.New(),
				})
			},
			requestData: &entities.GuestEventEntity{
				ID:        "550e8400-e29b-41d4-a716-446655440001",
				Name:      "Test Guest",
				CreatedAt: 1234567890,
				CreatedBy: "test_user",
			},
			mockServer: func() *httptest.Server {
				return nil
			},
			expectError: true,
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err, "SendWebhook() expected error for http request failure")
			},
		},
		{
			name: "send webhook with 4xx client error response",
			setupRepo: func(serverURL string) *WebhookSiteRepository {
				cfg := &configs.Config{}
				cfg.Datasource.WebhookSiteHTTPClient.BaseURL = serverURL
				cfg.Datasource.WebhookSiteHTTPClient.Endpoint.Webhook = "/webhook"
				httpClient := resty.New()
				httpClient.SetBaseURL(serverURL)
				return NewWebhookSiteRepository(cfg, &webhook_site_http_client.WebhookSiteHTTPClient{
					HttpClient: httpClient,
				})
			},
			requestData: &entities.GuestEventEntity{
				ID:        "550e8400-e29b-41d4-a716-446655440002",
				Name:      "Test Guest",
				CreatedAt: 1234567890,
				CreatedBy: "test_user",
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error": "bad request"}`))
				}))
			},
			expectError: true,
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err, "SendWebhook() expected error for 4xx response")
			},
		},
		{
			name: "send webhook with 5xx server error response",
			setupRepo: func(serverURL string) *WebhookSiteRepository {
				cfg := &configs.Config{}
				cfg.Datasource.WebhookSiteHTTPClient.BaseURL = serverURL
				cfg.Datasource.WebhookSiteHTTPClient.Endpoint.Webhook = "/webhook"
				httpClient := resty.New()
				httpClient.SetBaseURL(serverURL)
				return NewWebhookSiteRepository(cfg, &webhook_site_http_client.WebhookSiteHTTPClient{
					HttpClient: httpClient,
				})
			},
			requestData: &entities.GuestEventEntity{
				ID:        "550e8400-e29b-41d4-a716-446655440003",
				Name:      "Test Guest",
				CreatedAt: 1234567890,
				CreatedBy: "test_user",
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "internal server error"}`))
				}))
			},
			expectError: true,
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err, "SendWebhook() expected error for 5xx response")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.mockServer != nil {
				server = tt.mockServer()
				if server != nil {
					defer server.Close()
				}
			}

			serverURL := ""
			if server != nil {
				serverURL = server.URL
			}

			repo := tt.setupRepo(serverURL)
			ctx := context.Background()

			err := repo.SendWebhook(ctx, tt.requestData)

			if tt.validateError != nil {
				tt.validateError(t, err)
			}
		})
	}
}
