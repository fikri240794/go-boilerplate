package configs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	tests := []struct {
		name           string
		setupFile      func(t *testing.T) string
		cleanupFile    func(t *testing.T, filePath string)
		expectPanic    bool
		validateConfig func(t *testing.T, config *Config)
	}{
		{
			name: "read valid config file with all fields",
			setupFile: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "test.env")
				content := `ENVIRONMENT=development

SERVER.NAME=test-server
SERVER.LOG_LEVEL=1

SERVER.HTTP.PORT=8080
SERVER.HTTP.PREFORK=false
SERVER.HTTP.PRINT_ROUTES=true
SERVER.HTTP.REQUEST_TIMEOUT=30s
SERVER.HTTP.GRACEFULLY_SHUTDOWN_DURATION=10s
SERVER.HTTP.CORS.ALLOW_ORIGINS=*
SERVER.HTTP.CORS.ALLOW_METHODS=GET,POST,PUT,DELETE
SERVER.HTTP.DOCS.SWAGGER.ENABLE=true
SERVER.HTTP.DOCS.SWAGGER.FILE_PATH=./docs/swagger.json
SERVER.HTTP.DOCS.SWAGGER.PATH=/swagger
SERVER.HTTP.DOCS.SWAGGER.TITLE=Test API

SERVER.GRPC.PORT=9090
SERVER.GRPC.REQUEST_TIMEOUT=30s

SERVER.EVENT_CONSUMER.DATA_SOURCE_NAME=amqp://guest:guest@localhost:5672/

SERVER.TRACER.SERVICE_NAME=test-service
SERVER.TRACER.EXPORTER_GRPC_ADDRESS=localhost:4317

DATASOURCE.BOILERPLATE_DATABASE.MASTER.DRIVER_NAME=postgres
DATASOURCE.BOILERPLATE_DATABASE.MASTER.DATA_SOURCE_NAME=postgres://user:pass@localhost:5432/db
DATASOURCE.BOILERPLATE_DATABASE.MASTER.MAXIMUM_OPEN_CONNECTIONS=10
DATASOURCE.BOILERPLATE_DATABASE.MASTER.MAXIMUM_IDLE_CONNECTIONS=5
DATASOURCE.BOILERPLATE_DATABASE.MASTER.CONNECTION_MAXIMUM_IDLE_TIME=5m
DATASOURCE.BOILERPLATE_DATABASE.MASTER.CONNECTION_MAXIMUM_LIFE_TIME=1h
DATASOURCE.BOILERPLATE_DATABASE.MASTER.MAXIMUM_QUERY_DURATION_WARNING=1s

DATASOURCE.BOILERPLATE_DATABASE.SLAVE.DRIVER_NAME=postgres
DATASOURCE.BOILERPLATE_DATABASE.SLAVE.DATA_SOURCE_NAME=postgres://user:pass@localhost:5433/db
DATASOURCE.BOILERPLATE_DATABASE.SLAVE.MAXIMUM_OPEN_CONNECTIONS=10
DATASOURCE.BOILERPLATE_DATABASE.SLAVE.MAXIMUM_IDLE_CONNECTIONS=5
DATASOURCE.BOILERPLATE_DATABASE.SLAVE.CONNECTION_MAXIMUM_IDLE_TIME=5m
DATASOURCE.BOILERPLATE_DATABASE.SLAVE.CONNECTION_MAXIMUM_LIFE_TIME=1h
DATASOURCE.BOILERPLATE_DATABASE.SLAVE.MAXIMUM_QUERY_DURATION_WARNING=1s

DATASOURCE.IN_MEMORY_DATABASE.DATA_SOURCE_NAME=redis://localhost:6379

DATASOURCE.EVENT_PRODUCER.DATA_SOURCE_NAME=amqp://guest:guest@localhost:5672/

DATASOURCE.WEBHOOK_SITE_HTTP_CLIENT.BASE_URL=https://webhook.site
DATASOURCE.WEBHOOK_SITE_HTTP_CLIENT.ENDPOINT.WEBHOOK=/webhook-endpoint

GUEST.CACHE.ENABLE=true
GUEST.CACHE.KEYF=guest:%s
GUEST.CACHE.DURATION=1h

GUEST.EVENT.CREATED.ENABLE=true
GUEST.EVENT.CREATED.TOPIC=guest.created

GUEST.EVENT.DELETED.ENABLE=true
GUEST.EVENT.DELETED.TOPIC=guest.deleted

GUEST.EVENT.UPDATED.ENABLE=true
GUEST.EVENT.UPDATED.TOPIC=guest.updated
`
				err := os.WriteFile(tmpFile, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test config file: %v", err)
				}
				return tmpFile
			},
			cleanupFile: func(t *testing.T, filePath string) {
				os.Remove(filePath)
			},
			expectPanic: false,
			validateConfig: func(t *testing.T, config *Config) {
				assert.Equal(t, "development", config.Environment)
				assert.Equal(t, "test-server", config.Server.Name)
				assert.Equal(t, int8(1), config.Server.LogLevel)
				assert.Equal(t, 8080, config.Server.HTTP.Port)
				assert.False(t, config.Server.HTTP.Prefork)
				assert.True(t, config.Server.HTTP.PrintRoutes)
				assert.Equal(t, 30*time.Second, config.Server.HTTP.RequestTimeout)
				assert.Equal(t, "*", config.Server.HTTP.CORS.AllowOrigins)
				assert.Equal(t, 9090, config.Server.GRPC.Port)
				assert.True(t, config.Guest.Cache.Enable)
				assert.Equal(t, "guest:%s", config.Guest.Cache.Keyf)
				assert.Equal(t, "guest.created", config.Guest.Event.Created.Topic)
			},
		},
		{
			name: "read config file with minimal fields",
			setupFile: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "minimal.env")
				content := `ENVIRONMENT=production

SERVER.NAME=minimal-server
SERVER.HTTP.PORT=3000
`
				err := os.WriteFile(tmpFile, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test config file: %v", err)
				}
				return tmpFile
			},
			cleanupFile: func(t *testing.T, filePath string) {
				os.Remove(filePath)
			},
			expectPanic: false,
			validateConfig: func(t *testing.T, config *Config) {
				assert.Equal(t, "production", config.Environment)
				assert.Equal(t, "minimal-server", config.Server.Name)
				assert.Equal(t, 3000, config.Server.HTTP.Port)
				assert.Equal(t, int8(0), config.Server.LogLevel)
				assert.Equal(t, 0, config.Server.GRPC.Port)
			},
		},
		{
			name: "panic when config file does not exist",
			setupFile: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent.env")
			},
			cleanupFile:    func(t *testing.T, filePath string) {},
			expectPanic:    true,
			validateConfig: func(t *testing.T, config *Config) {},
		},
		{
			name: "panic when config file has invalid format",
			setupFile: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "invalid.env")
				content := `INVALID CONTENT WITHOUT EQUAL SIGN
THIS IS NOT A VALID ENV FILE
{ "json": "object" }
`
				err := os.WriteFile(tmpFile, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test config file: %v", err)
				}
				return tmpFile
			},
			cleanupFile: func(t *testing.T, filePath string) {
				os.Remove(filePath)
			},
			expectPanic:    true,
			validateConfig: func(t *testing.T, config *Config) {},
		},
		{
			name: "read config with special characters in values",
			setupFile: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "special.env")
				content := `ENVIRONMENT=development

SERVER.NAME=test-server-with-special-chars

DATASOURCE.BOILERPLATE_DATABASE.MASTER.DATA_SOURCE_NAME=postgres://user:p@ssw0rd!@localhost:5432/db?sslmode=disable
DATASOURCE.WEBHOOK_SITE_HTTP_CLIENT.BASE_URL=https://webhook.site
DATASOURCE.WEBHOOK_SITE_HTTP_CLIENT.ENDPOINT.WEBHOOK=/a46fd97b-b775-428c-890d-9d71851a6c32
`
				err := os.WriteFile(tmpFile, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test config file: %v", err)
				}
				return tmpFile
			},
			cleanupFile: func(t *testing.T, filePath string) {
				os.Remove(filePath)
			},
			expectPanic: false,
			validateConfig: func(t *testing.T, config *Config) {
				assert.Equal(t, "test-server-with-special-chars", config.Server.Name)
				assert.Equal(t, "postgres://user:p@ssw0rd!@localhost:5432/db?sslmode=disable", config.Datasource.BoilerplateDatabase.Master.DataSourceName)
				assert.Equal(t, "https://webhook.site", config.Datasource.WebhookSiteHTTPClient.BaseURL)
				assert.Equal(t, "/a46fd97b-b775-428c-890d-9d71851a6c32", config.Datasource.WebhookSiteHTTPClient.Endpoint.Webhook)
			},
		},
		{
			name: "panic when unmarshal fails due to invalid type conversion",
			setupFile: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "invalid_type.env")
				content := `ENVIRONMENT=development

SERVER.LOG_LEVEL=invalid_number
SERVER.HTTP.PORT=not_a_number
SERVER.HTTP.PREFORK=not_a_boolean
`
				err := os.WriteFile(tmpFile, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test config file: %v", err)
				}
				return tmpFile
			},
			cleanupFile: func(t *testing.T, filePath string) {
				os.Remove(filePath)
			},
			expectPanic:    true,
			validateConfig: func(t *testing.T, config *Config) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile(t)
			defer tt.cleanupFile(t, filePath)

			if tt.expectPanic {
				assert.Panics(t, func() {
					Read(filePath)
				})
			} else {
				config := Read(filePath)
				assert.NotNil(t, config)
				tt.validateConfig(t, config)
			}
		})
	}
}
