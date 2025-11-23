package event_producer

import (
	"errors"
	"go-boilerplate/configs"
	"go-boilerplate/datasources/event_producer/mocks"
	"testing"

	"github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/assert"
)

func Test_connectToNSQProducer(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *configs.Config
		fn          nsqProducer
		expectPanic bool
		validate    func(t *testing.T, producer *EventProducer)
	}{
		{
			name: "connect with fn returns error should panic",
			setupConfig: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.EventProducer.DataSourceName = "test-host:4150"
				return cfg
			},
			fn: func(addr string, config *nsq.Config) (INSQProducer, error) {
				return nil, errors.New("producer error")
			},
			expectPanic: true,
		},
		{
			name: "connect with ping error should panic",
			setupConfig: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.EventProducer.DataSourceName = "test-host:4150"
				return cfg
			},
			fn: func(addr string, config *nsq.Config) (INSQProducer, error) {
				mockProducer := mocks.NewNSQProducerMock(t)
				mockProducer.On("Ping").Return(errors.New("ping failed"))
				return mockProducer, nil
			},
			expectPanic: true,
		},
		{
			name: "connect successfully with mock producer",
			setupConfig: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.EventProducer.DataSourceName = "test-host:4150"
				return cfg
			},
			fn: func(addr string, config *nsq.Config) (INSQProducer, error) {
				mockProducer := mocks.NewNSQProducerMock(t)
				mockProducer.On("Ping").Return(nil)
				mockProducer.On("Stop").Return()
				return mockProducer, nil
			},
			expectPanic: false,
			validate: func(t *testing.T, producer *EventProducer) {
				assert.NotNil(t, producer)
				assert.NotNil(t, producer.NSQProducer)
				assert.IsType(t, &mocks.NSQProducerMock{}, producer.NSQProducer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()

			if tt.expectPanic {
				assert.Panics(t, func() {
					connectToNSQProducer(cfg, tt.fn)
				})
			} else {
				producer := connectToNSQProducer(cfg, tt.fn)

				if tt.validate != nil {
					tt.validate(t, producer)
				}

				err := producer.Disconnect()
				assert.NoError(t, err)
			}
		})
	}
}

func TestConnect(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *configs.Config
		expectPanic bool
	}{
		{
			name: "connect with empty data source name should panic",
			setupConfig: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.EventProducer.DataSourceName = ""
				return cfg
			},
			expectPanic: true,
		},
		{
			name: "connect with invalid data source name should panic",
			setupConfig: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.EventProducer.DataSourceName = "invalid-host:0"
				return cfg
			},
			expectPanic: true,
		},
		{
			name: "connect with unreachable host should panic on ping",
			setupConfig: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Datasource.EventProducer.DataSourceName = "localhost:99999"
				return cfg
			},
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()

			if tt.expectPanic {
				assert.Panics(t, func() {
					Connect(cfg)
				})
			} else {
				producer := Connect(cfg)
				assert.NotNil(t, producer)
				assert.NotNil(t, producer.NSQProducer)

				err := producer.Disconnect()
				assert.NoError(t, err)
			}
		})
	}
}

func TestDisconnect(t *testing.T) {
	tests := []struct {
		name          string
		setupProducer func(t *testing.T) *EventProducer
		expectError   bool
		expectPanic   bool
	}{
		{
			name: "disconnect with nil NSQProducer should panic",
			setupProducer: func(t *testing.T) *EventProducer {
				return &EventProducer{
					NSQProducer: nil,
				}
			},
			expectError: false,
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer := tt.setupProducer(t)

			if tt.expectPanic {
				assert.Panics(t, func() {
					producer.Disconnect()
				})
			} else {
				err := producer.Disconnect()

				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}
