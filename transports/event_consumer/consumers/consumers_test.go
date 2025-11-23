package consumers

import (
	"go-boilerplate/configs"
	"go-boilerplate/internal/services/mocks"
	"go-boilerplate/transports/event_consumer/handlers"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsumers_ConsumeEvents(t *testing.T) {
	tests := []struct {
		name        string
		setupGuest  func(t *testing.T) *GuestConsumer
		validateErr func(t *testing.T, err error)
	}{
		{
			name: "should_consume_events_successfully_with_no_consumers_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Server.EventConsumer.DataSourceName = "localhost:4161"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = false

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validateErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "should_consume_events_with_created_consumer_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Server.EventConsumer.DataSourceName = "localhost:4161"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created-consumers-test"
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = false

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validateErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "should_consume_events_with_deleted_consumer_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Server.EventConsumer.DataSourceName = "localhost:4161"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted-consumers-test"
				cfg.Guest.Event.Updated.Enable = false

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validateErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "should_consume_events_with_updated_consumer_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Server.EventConsumer.DataSourceName = "localhost:4161"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated-consumers-test"

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validateErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "should_consume_events_with_all_consumers_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Server.EventConsumer.DataSourceName = "localhost:4161"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created-consumers-test"
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted-consumers-test"
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated-consumers-test"

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validateErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "should_consume_events_with_created_and_deleted_consumers_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Server.EventConsumer.DataSourceName = "localhost:4161"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created-consumers-test"
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted-consumers-test"
				cfg.Guest.Event.Updated.Enable = false

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validateErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "should_consume_events_with_created_and_updated_consumers_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Server.EventConsumer.DataSourceName = "localhost:4161"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created-consumers-test"
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated-consumers-test"

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validateErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "should_consume_events_with_deleted_and_updated_consumers_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Server.EventConsumer.DataSourceName = "localhost:4161"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted-consumers-test"
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated-consumers-test"

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validateErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guestConsumer := tt.setupGuest(t)
			consumers := &Consumers{
				Guest: guestConsumer,
			}

			err := consumers.ConsumeEvents()

			if tt.validateErr != nil {
				tt.validateErr(t, err)
			}

			if guestConsumer.createdGuestConsumer != nil || guestConsumer.deletedGuestConsumer != nil || guestConsumer.updatedGuestConsumer != nil {
				consumers.Stop()
			}
		})
	}
}

func TestConsumers_Stop(t *testing.T) {
	tests := []struct {
		name       string
		setupGuest func(t *testing.T) *GuestConsumer
		validate   func(t *testing.T, consumers *Consumers)
	}{
		{
			name: "should_stop_successfully_with_no_consumers_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = false

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validate: func(t *testing.T, consumers *Consumers) {
				assert.NotPanics(t, func() {
					consumers.Stop()
				})
			},
		},
		{
			name: "should_stop_successfully_with_created_consumer_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created-stop-test"
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = false

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validate: func(t *testing.T, consumers *Consumers) {
				assert.NotPanics(t, func() {
					consumers.Stop()
				})
			},
		},
		{
			name: "should_stop_successfully_with_deleted_consumer_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted-stop-test"
				cfg.Guest.Event.Updated.Enable = false

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validate: func(t *testing.T, consumers *Consumers) {
				assert.NotPanics(t, func() {
					consumers.Stop()
				})
			},
		},
		{
			name: "should_stop_successfully_with_updated_consumer_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated-stop-test"

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validate: func(t *testing.T, consumers *Consumers) {
				assert.NotPanics(t, func() {
					consumers.Stop()
				})
			},
		},
		{
			name: "should_stop_successfully_with_all_consumers_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created-stop-test"
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted-stop-test"
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated-stop-test"

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validate: func(t *testing.T, consumers *Consumers) {
				assert.NotPanics(t, func() {
					consumers.Stop()
				})
			},
		},
		{
			name: "should_stop_successfully_with_created_and_deleted_consumers_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created-stop-test"
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted-stop-test"
				cfg.Guest.Event.Updated.Enable = false

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validate: func(t *testing.T, consumers *Consumers) {
				assert.NotPanics(t, func() {
					consumers.Stop()
				})
			},
		},
		{
			name: "should_stop_successfully_with_created_and_updated_consumers_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest-created-stop-test"
				cfg.Guest.Event.Deleted.Enable = false
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated-stop-test"

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validate: func(t *testing.T, consumers *Consumers) {
				assert.NotPanics(t, func() {
					consumers.Stop()
				})
			},
		},
		{
			name: "should_stop_successfully_with_deleted_and_updated_consumers_enabled",
			setupGuest: func(t *testing.T) *GuestConsumer {
				cfg := &configs.Config{}
				cfg.Server.Name = "test-service"
				cfg.Guest.Event.Created.Enable = false
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest-deleted-stop-test"
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest-updated-stop-test"

				mockService := mocks.NewGuestServiceMock(t)
				handler := handlers.NewGuestHandler(mockService)
				return NewGuestConsumer(cfg, handler)
			},
			validate: func(t *testing.T, consumers *Consumers) {
				assert.NotPanics(t, func() {
					consumers.Stop()
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guestConsumer := tt.setupGuest(t)
			consumers := &Consumers{
				Guest: guestConsumer,
			}

			if tt.validate != nil {
				tt.validate(t, consumers)
			}
		})
	}
}
