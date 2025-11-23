package boilerplate_database

import (
	"go-boilerplate/configs"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	tests := []struct {
		name         string
		setupConfig  func() *configs.Config
		expectPanic  bool
		validateConn func(t *testing.T, conn *BoilerplateDatabase)
	}{
		{
			name: "connect with invalid master driver should panic",
			setupConfig: func() *configs.Config {
				return &configs.Config{
					Datasource: struct {
						BoilerplateDatabase struct {
							Master struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							} `mapstructure:"MASTER"`
							Slave struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							} `mapstructure:"SLAVE"`
						} `mapstructure:"BOILERPLATE_DATABASE"`
						InMemoryDatabase struct {
							DataSourceName string `mapstructure:"DATA_SOURCE_NAME"`
						} `mapstructure:"IN_MEMORY_DATABASE"`
						EventProducer struct {
							DataSourceName string `mapstructure:"DATA_SOURCE_NAME"`
						} `mapstructure:"EVENT_PRODUCER"`
						WebhookSiteHTTPClient struct {
							BaseURL  string `mapstructure:"BASE_URL"`
							Endpoint struct {
								Webhook string `mapstructure:"WEBHOOK"`
							} `mapstructure:"ENDPOINT"`
						} `mapstructure:"WEBHOOK_SITE_HTTP_CLIENT"`
					}{
						BoilerplateDatabase: struct {
							Master struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							} `mapstructure:"MASTER"`
							Slave struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							} `mapstructure:"SLAVE"`
						}{
							Master: struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							}{
								DriverName:                  "invalid_driver",
								DataSourceName:              "invalid://localhost:5432/testdb",
								MaximumOpenConnections:      10,
								MaximumIddleConnections:     5,
								ConnectionMaximumIdleTime:   5 * time.Minute,
								ConnectionMaximumLifeTime:   1 * time.Hour,
								MaximumQueryDurationWarning: 1 * time.Second,
							},
							Slave: struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							}{
								DriverName:                  "postgres",
								DataSourceName:              "postgres://user:pass@localhost:5432/testdb",
								MaximumOpenConnections:      10,
								MaximumIddleConnections:     5,
								ConnectionMaximumIdleTime:   5 * time.Minute,
								ConnectionMaximumLifeTime:   1 * time.Hour,
								MaximumQueryDurationWarning: 1 * time.Second,
							},
						},
					},
				}
			},
			expectPanic:  true,
			validateConn: func(t *testing.T, conn *BoilerplateDatabase) {},
		},
		{
			name: "connect with invalid slave driver should panic",
			setupConfig: func() *configs.Config {
				return &configs.Config{
					Datasource: struct {
						BoilerplateDatabase struct {
							Master struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							} `mapstructure:"MASTER"`
							Slave struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							} `mapstructure:"SLAVE"`
						} `mapstructure:"BOILERPLATE_DATABASE"`
						InMemoryDatabase struct {
							DataSourceName string `mapstructure:"DATA_SOURCE_NAME"`
						} `mapstructure:"IN_MEMORY_DATABASE"`
						EventProducer struct {
							DataSourceName string `mapstructure:"DATA_SOURCE_NAME"`
						} `mapstructure:"EVENT_PRODUCER"`
						WebhookSiteHTTPClient struct {
							BaseURL  string `mapstructure:"BASE_URL"`
							Endpoint struct {
								Webhook string `mapstructure:"WEBHOOK"`
							} `mapstructure:"ENDPOINT"`
						} `mapstructure:"WEBHOOK_SITE_HTTP_CLIENT"`
					}{
						BoilerplateDatabase: struct {
							Master struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							} `mapstructure:"MASTER"`
							Slave struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							} `mapstructure:"SLAVE"`
						}{
							Master: struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							}{
								DriverName:                  "postgres",
								DataSourceName:              "postgres://user:pass@localhost:5432/testdb",
								MaximumOpenConnections:      10,
								MaximumIddleConnections:     5,
								ConnectionMaximumIdleTime:   5 * time.Minute,
								ConnectionMaximumLifeTime:   1 * time.Hour,
								MaximumQueryDurationWarning: 1 * time.Second,
							},
							Slave: struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							}{
								DriverName:                  "invalid_driver",
								DataSourceName:              "invalid://localhost:5432/testdb",
								MaximumOpenConnections:      10,
								MaximumIddleConnections:     5,
								ConnectionMaximumIdleTime:   5 * time.Minute,
								ConnectionMaximumLifeTime:   1 * time.Hour,
								MaximumQueryDurationWarning: 1 * time.Second,
							},
						},
					},
				}
			},
			expectPanic:  true,
			validateConn: func(t *testing.T, conn *BoilerplateDatabase) {},
		},
		{
			name: "connect with invalid master DSN should panic",
			setupConfig: func() *configs.Config {
				return &configs.Config{
					Datasource: struct {
						BoilerplateDatabase struct {
							Master struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							} `mapstructure:"MASTER"`
							Slave struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							} `mapstructure:"SLAVE"`
						} `mapstructure:"BOILERPLATE_DATABASE"`
						InMemoryDatabase struct {
							DataSourceName string `mapstructure:"DATA_SOURCE_NAME"`
						} `mapstructure:"IN_MEMORY_DATABASE"`
						EventProducer struct {
							DataSourceName string `mapstructure:"DATA_SOURCE_NAME"`
						} `mapstructure:"EVENT_PRODUCER"`
						WebhookSiteHTTPClient struct {
							BaseURL  string `mapstructure:"BASE_URL"`
							Endpoint struct {
								Webhook string `mapstructure:"WEBHOOK"`
							} `mapstructure:"ENDPOINT"`
						} `mapstructure:"WEBHOOK_SITE_HTTP_CLIENT"`
					}{
						BoilerplateDatabase: struct {
							Master struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							} `mapstructure:"MASTER"`
							Slave struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							} `mapstructure:"SLAVE"`
						}{
							Master: struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							}{
								DriverName:                  "postgres",
								DataSourceName:              "invalid-dsn-format",
								MaximumOpenConnections:      10,
								MaximumIddleConnections:     5,
								ConnectionMaximumIdleTime:   5 * time.Minute,
								ConnectionMaximumLifeTime:   1 * time.Hour,
								MaximumQueryDurationWarning: 1 * time.Second,
							},
							Slave: struct {
								DriverName                  string        `mapstructure:"DRIVER_NAME"`
								DataSourceName              string        `mapstructure:"DATA_SOURCE_NAME"`
								MaximumOpenConnections      int           `mapstructure:"MAXIMUM_OPEN_CONNECTIONS"`
								MaximumIddleConnections     int           `mapstructure:"MAXIMUM_IDLE_CONNECTIONS"`
								ConnectionMaximumIdleTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_IDLE_TIME"`
								ConnectionMaximumLifeTime   time.Duration `mapstructure:"CONNECTION_MAXIMUM_LIFE_TIME"`
								MaximumQueryDurationWarning time.Duration `mapstructure:"MAXIMUM_QUERY_DURATION_WARNING"`
							}{
								DriverName:                  "postgres",
								DataSourceName:              "postgres://user:pass@localhost:5432/testdb",
								MaximumOpenConnections:      10,
								MaximumIddleConnections:     5,
								ConnectionMaximumIdleTime:   5 * time.Minute,
								ConnectionMaximumLifeTime:   1 * time.Hour,
								MaximumQueryDurationWarning: 1 * time.Second,
							},
						},
					},
				}
			},
			expectPanic:  true,
			validateConn: func(t *testing.T, conn *BoilerplateDatabase) {},
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
				conn := Connect(cfg)
				assert.NotNil(t, conn)

				tt.validateConn(t, conn)

				err := conn.Disconnect()
				if err != nil {
					t.Logf("Warning: Failed to disconnect: %v", err)
				}
			}
		})
	}
}

func TestDisconnect(t *testing.T) {
	tests := []struct {
		name         string
		setupConn    func(t *testing.T) *BoilerplateDatabase
		expectPanic  bool
		validateFunc func(t *testing.T, err error)
	}{
		{
			name: "disconnect with nil Master should panic",
			setupConn: func(t *testing.T) *BoilerplateDatabase {
				return &BoilerplateDatabase{
					Master:                        nil,
					MasterMaxQueryDurationWarning: 1 * time.Second,
					Slave:                         nil,
					SlaveMaxQueryDurationWarning:  1 * time.Second,
				}
			},
			expectPanic:  true,
			validateFunc: func(t *testing.T, err error) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := tt.setupConn(t)

			if tt.expectPanic {
				assert.Panics(t, func() {
					conn.Disconnect()
				})
			} else {
				err := conn.Disconnect()

				if tt.validateFunc != nil {
					tt.validateFunc(t, err)
				}
			}
		})
	}
}
