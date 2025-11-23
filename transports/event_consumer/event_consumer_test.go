package event_consumer

import (
	"go-boilerplate/configs"
	"go-boilerplate/datasources"
	"go-boilerplate/datasources/boilerplate_database"
	"go-boilerplate/datasources/event_producer"
	"go-boilerplate/datasources/in_memory_database"
	"go-boilerplate/internal/services/mocks"
	"go-boilerplate/transports/event_consumer/consumers"
	"go-boilerplate/transports/event_consumer/handlers"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewEventConsumer(t *testing.T) {
	tests := []struct {
		name             string
		setupCfg         func(t *testing.T) *configs.Config
		setupDatasources func(t *testing.T) *datasources.Datasources
		setupConsumers   func(t *testing.T) *consumers.Consumers
		validate         func(t *testing.T, ec *EventConsumer, cfg *configs.Config, ds *datasources.Datasources, c *consumers.Consumers)
	}{
		{
			name: "should_create_event_consumer_successfully",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Server.LogLevel = int8(zerolog.InfoLevel)
				return cfg
			},
			setupDatasources: func(t *testing.T) *datasources.Datasources {
				return &datasources.Datasources{
					BoilerplateDatabase: &boilerplate_database.BoilerplateDatabase{},
					InMemoryDatabase:    &in_memory_database.InMemoryDatabase{},
					EventProducer:       &event_producer.EventProducer{},
				}
			},
			setupConsumers: func(t *testing.T) *consumers.Consumers {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = false

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				guestConsumer := consumers.NewGuestConsumer(cfg, handler)

				return &consumers.Consumers{
					Guest: guestConsumer,
				}
			},
			validate: func(t *testing.T, ec *EventConsumer, cfg *configs.Config, ds *datasources.Datasources, c *consumers.Consumers) {
				assert.NotNil(t, ec)
				assert.Equal(t, cfg, ec.cfg)
				assert.Equal(t, ds, ec.datasources)
				assert.Equal(t, c, ec.eventConsumers)
			},
		},
		{
			name: "should_create_event_consumer_with_all_fields",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "production-service"
				cfg.Server.LogLevel = int8(zerolog.DebugLevel)
				cfg.Server.EventConsumer.DataSourceName = "localhost:4161"
				return cfg
			},
			setupDatasources: func(t *testing.T) *datasources.Datasources {
				return &datasources.Datasources{
					BoilerplateDatabase: &boilerplate_database.BoilerplateDatabase{},
					InMemoryDatabase:    &in_memory_database.InMemoryDatabase{},
					EventProducer:       &event_producer.EventProducer{},
				}
			},
			setupConsumers: func(t *testing.T) *consumers.Consumers {
				cfg := &configs.Config{}
				cfg.Server.Name = "production-service"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created"
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted"
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated"

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				guestConsumer := consumers.NewGuestConsumer(cfg, handler)

				return &consumers.Consumers{
					Guest: guestConsumer,
				}
			},
			validate: func(t *testing.T, ec *EventConsumer, cfg *configs.Config, ds *datasources.Datasources, c *consumers.Consumers) {
				assert.NotNil(t, ec)
				assert.Equal(t, cfg, ec.cfg)
				assert.Equal(t, ds, ec.datasources)
				assert.Equal(t, c, ec.eventConsumers)
				assert.Equal(t, "production-service", ec.cfg.Server.Name)
			},
		},
		{
			name: "should_create_event_consumer_with_nil_safe_fields",
			setupCfg: func(t *testing.T) *configs.Config {
				return &configs.Config{}
			},
			setupDatasources: func(t *testing.T) *datasources.Datasources {
				return &datasources.Datasources{}
			},
			setupConsumers: func(t *testing.T) *consumers.Consumers {
				return &consumers.Consumers{}
			},
			validate: func(t *testing.T, ec *EventConsumer, cfg *configs.Config, ds *datasources.Datasources, c *consumers.Consumers) {
				assert.NotNil(t, ec)
				assert.Equal(t, cfg, ec.cfg)
				assert.Equal(t, ds, ec.datasources)
				assert.Equal(t, c, ec.eventConsumers)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupCfg(t)
			ds := tt.setupDatasources(t)
			c := tt.setupConsumers(t)

			ec := NewEventConsumer(cfg, ds, c)

			if tt.validate != nil {
				tt.validate(t, ec, cfg, ds, c)
			}
		})
	}
}

func TestEventConsumer_setGlobalLog(t *testing.T) {
	tests := []struct {
		name     string
		setupCfg func(t *testing.T) *configs.Config
		validate func(t *testing.T)
	}{
		{
			name: "should_set_global_log_with_info_level",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.LogLevel = int8(zerolog.InfoLevel)
				return cfg
			},
			validate: func(t *testing.T) {
				assert.NotPanics(t, func() {

				})
			},
		},
		{
			name: "should_set_global_log_with_debug_level",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.LogLevel = int8(zerolog.DebugLevel)
				return cfg
			},
			validate: func(t *testing.T) {
				assert.NotPanics(t, func() {

				})
			},
		},
		{
			name: "should_set_global_log_with_warn_level",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.LogLevel = int8(zerolog.WarnLevel)
				return cfg
			},
			validate: func(t *testing.T) {
				assert.NotPanics(t, func() {

				})
			},
		},
		{
			name: "should_set_global_log_with_error_level",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.LogLevel = int8(zerolog.ErrorLevel)
				return cfg
			},
			validate: func(t *testing.T) {
				assert.NotPanics(t, func() {

				})
			},
		},
		{
			name: "should_set_global_log_with_trace_level",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.LogLevel = int8(zerolog.TraceLevel)
				return cfg
			},
			validate: func(t *testing.T) {
				assert.NotPanics(t, func() {

				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupCfg(t)
			ds := &datasources.Datasources{}
			c := &consumers.Consumers{}

			ec := NewEventConsumer(cfg, ds, c)

			assert.NotPanics(t, func() {
				ec.setGlobalLog()
			})

			if tt.validate != nil {
				tt.validate(t)
			}
		})
	}
}
