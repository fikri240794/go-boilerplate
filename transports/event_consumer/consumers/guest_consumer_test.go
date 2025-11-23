package consumers

import (
	"go-boilerplate/configs"
	"go-boilerplate/internal/services/mocks"
	"go-boilerplate/transports/event_consumer/handlers"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGuestConsumer(t *testing.T) {
	tests := []struct {
		name        string
		setupCfg    func(t *testing.T) *configs.Config
		setupMock   func(t *testing.T) *mocks.GuestServiceMock
		validate    func(t *testing.T, consumer *GuestConsumer, cfg *configs.Config)
		shouldPanic bool
	}{
		{
			name: "should_create_guest_consumer_with_all_events_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created"
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted"
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated"
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer, cfg *configs.Config) {
				assert.NotNil(t, consumer)
				assert.Equal(t, cfg, consumer.cfg)
				assert.NotNil(t, consumer.createdGuestConsumer)
				assert.NotNil(t, consumer.deletedGuestConsumer)
				assert.NotNil(t, consumer.updatedGuestConsumer)
			},
			shouldPanic: false,
		},
		{
			name: "should_create_guest_consumer_with_only_created_event_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created"
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = false
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer, cfg *configs.Config) {
				assert.NotNil(t, consumer)
				assert.Equal(t, cfg, consumer.cfg)
				assert.NotNil(t, consumer.createdGuestConsumer)
				assert.Nil(t, consumer.deletedGuestConsumer)
				assert.Nil(t, consumer.updatedGuestConsumer)
			},
			shouldPanic: false,
		},
		{
			name: "should_create_guest_consumer_with_only_deleted_event_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted"
				cfg.Guest.Event.Updated.Enable = false
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer, cfg *configs.Config) {
				assert.NotNil(t, consumer)
				assert.Equal(t, cfg, consumer.cfg)
				assert.Nil(t, consumer.createdGuestConsumer)
				assert.NotNil(t, consumer.deletedGuestConsumer)
				assert.Nil(t, consumer.updatedGuestConsumer)
			},
			shouldPanic: false,
		},
		{
			name: "should_create_guest_consumer_with_only_updated_event_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated"
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer, cfg *configs.Config) {
				assert.NotNil(t, consumer)
				assert.Equal(t, cfg, consumer.cfg)
				assert.Nil(t, consumer.createdGuestConsumer)
				assert.Nil(t, consumer.deletedGuestConsumer)
				assert.NotNil(t, consumer.updatedGuestConsumer)
			},
			shouldPanic: false,
		},
		{
			name: "should_create_guest_consumer_with_no_events_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = false
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer, cfg *configs.Config) {
				assert.NotNil(t, consumer)
				assert.Equal(t, cfg, consumer.cfg)
				assert.Nil(t, consumer.createdGuestConsumer)
				assert.Nil(t, consumer.deletedGuestConsumer)
				assert.Nil(t, consumer.updatedGuestConsumer)
			},
			shouldPanic: false,
		},
		{
			name: "should_create_guest_consumer_with_created_and_deleted_events_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created"
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted"
				cfg.Guest.Event.Updated.Enable = false
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer, cfg *configs.Config) {
				assert.NotNil(t, consumer)
				assert.Equal(t, cfg, consumer.cfg)
				assert.NotNil(t, consumer.createdGuestConsumer)
				assert.NotNil(t, consumer.deletedGuestConsumer)
				assert.Nil(t, consumer.updatedGuestConsumer)
			},
			shouldPanic: false,
		},
		{
			name: "should_create_guest_consumer_with_created_and_updated_events_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created"
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated"
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer, cfg *configs.Config) {
				assert.NotNil(t, consumer)
				assert.Equal(t, cfg, consumer.cfg)
				assert.NotNil(t, consumer.createdGuestConsumer)
				assert.Nil(t, consumer.deletedGuestConsumer)
				assert.NotNil(t, consumer.updatedGuestConsumer)
			},
			shouldPanic: false,
		},
		{
			name: "should_create_guest_consumer_with_deleted_and_updated_events_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted"
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated"
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer, cfg *configs.Config) {
				assert.NotNil(t, consumer)
				assert.Equal(t, cfg, consumer.cfg)
				assert.Nil(t, consumer.createdGuestConsumer)
				assert.NotNil(t, consumer.deletedGuestConsumer)
				assert.NotNil(t, consumer.updatedGuestConsumer)
			},
			shouldPanic: false,
		},
		{
			name: "should_panic_when_created_consumer_initialization_fails",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = ""
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = ""
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = false
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate:    func(t *testing.T, consumer *GuestConsumer, cfg *configs.Config) {},
			shouldPanic: true,
		},
		{
			name: "should_panic_when_deleted_consumer_initialization_fails",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = ""
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = ""
				cfg.Guest.Event.Updated.Enable = false
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate:    func(t *testing.T, consumer *GuestConsumer, cfg *configs.Config) {},
			shouldPanic: true,
		},
		{
			name: "should_panic_when_updated_consumer_initialization_fails",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = ""
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = ""
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate:    func(t *testing.T, consumer *GuestConsumer, cfg *configs.Config) {},
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupCfg(t)
			mockService := tt.setupMock(t)
			handler := handlers.NewGuestHandler(mockService)

			if tt.shouldPanic {
				assert.Panics(t, func() {
					NewGuestConsumer(cfg, handler)
				})
			} else {
				consumer := NewGuestConsumer(cfg, handler)
				if tt.validate != nil {
					tt.validate(t, consumer, cfg)
				}
			}
		})
	}
}

func TestGuestConsumer_ConsumeEvents(t *testing.T) {
	tests := []struct {
		name        string
		setupCfg    func(t *testing.T) *configs.Config
		setupMock   func(t *testing.T) *mocks.GuestServiceMock
		validateErr func(t *testing.T, err error)
	}{
		{
			name: "should_not_error_when_no_consumers_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Server.EventConsumer.DataSourceName = "localhost:4161"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = false
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validateErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "should_start_consume_events_with_created_consumer_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Server.EventConsumer.DataSourceName = "localhost:4161"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created-test"
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = false
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validateErr: func(t *testing.T, err error) {

				assert.NoError(t, err)
			},
		},
		{
			name: "should_start_consume_events_with_deleted_consumer_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Server.EventConsumer.DataSourceName = "localhost:4161"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted-test"
				cfg.Guest.Event.Updated.Enable = false
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validateErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "should_start_consume_events_with_updated_consumer_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Server.EventConsumer.DataSourceName = "localhost:4161"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated-test"
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validateErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "should_start_consume_events_with_all_consumers_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Server.EventConsumer.DataSourceName = "localhost:4161"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created-test"
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted-test"
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated-test"
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validateErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupCfg(t)
			mockService := tt.setupMock(t)
			handler := handlers.NewGuestHandler(mockService)
			consumer := NewGuestConsumer(cfg, handler)

			err := consumer.ConsumeEvents()

			if tt.validateErr != nil {
				tt.validateErr(t, err)
			}

			if consumer.createdGuestConsumer != nil || consumer.deletedGuestConsumer != nil || consumer.updatedGuestConsumer != nil {
				consumer.Stop()
			}
		})
	}
}

func TestGuestConsumer_Stop(t *testing.T) {
	tests := []struct {
		name      string
		setupCfg  func(t *testing.T) *configs.Config
		setupMock func(t *testing.T) *mocks.GuestServiceMock
		validate  func(t *testing.T, consumer *GuestConsumer)
	}{
		{
			name: "should_stop_all_consumers_when_all_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created"
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted"
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated"
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer) {
				assert.NotPanics(t, func() {
					consumer.Stop()
				})
			},
		},
		{
			name: "should_stop_only_created_consumer_when_only_created_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created"
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = false
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer) {
				assert.NotPanics(t, func() {
					consumer.Stop()
				})
			},
		},
		{
			name: "should_stop_only_deleted_consumer_when_only_deleted_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted"
				cfg.Guest.Event.Updated.Enable = false
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer) {
				assert.NotPanics(t, func() {
					consumer.Stop()
				})
			},
		},
		{
			name: "should_stop_only_updated_consumer_when_only_updated_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated"
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer) {
				assert.NotPanics(t, func() {
					consumer.Stop()
				})
			},
		},
		{
			name: "should_handle_stop_when_no_consumers_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = false
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer) {
				assert.NotPanics(t, func() {
					consumer.Stop()
				})
			},
		},
		{
			name: "should_stop_created_and_deleted_consumers_when_both_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created"
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted"
				cfg.Guest.Event.Updated.Enable = false
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer) {
				assert.NotPanics(t, func() {
					consumer.Stop()
				})
			},
		},
		{
			name: "should_stop_created_and_updated_consumers_when_both_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created"
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated"
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer) {
				assert.NotPanics(t, func() {
					consumer.Stop()
				})
			},
		},
		{
			name: "should_stop_deleted_and_updated_consumers_when_both_enabled",
			setupCfg: func(t *testing.T) *configs.Config {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted"
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated"
				return cfg
			},
			setupMock: func(t *testing.T) *mocks.GuestServiceMock {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, consumer *GuestConsumer) {
				assert.NotPanics(t, func() {
					consumer.Stop()
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupCfg(t)
			mockService := tt.setupMock(t)
			handler := handlers.NewGuestHandler(mockService)
			consumer := NewGuestConsumer(cfg, handler)

			if tt.validate != nil {
				tt.validate(t, consumer)
			}
		})
	}
}
