package configs

import (
	"time"

	"github.com/spf13/viper"
)

const DefaultConfigPath string = "./configs/.env"

type Config struct {
	Environment string `mapstructure:"ENVIRONMENT"`
	Server      struct {
		Name     string `mapstructure:"NAME"`
		LogLevel int8   `mapstructure:"LOG_LEVEL"`
		HTTP     struct {
			Port                       int           `mapstructure:"PORT"`
			Prefork                    bool          `mapstructure:"PREFORK"`
			PrintRoutes                bool          `mapstructure:"PRINT_ROUTES"`
			RequestTimeout             time.Duration `mapstructure:"REQUEST_TIMEOUT"`
			GracefullyShutdownDuration time.Duration `mapstructure:"GRACEFULLY_SHUTDOWN_DURATION"`
			CORS                       struct {
				AllowOrigins string `mapstructure:"ALLOW_ORIGINS"`
				AllowMethods string `mapstructure:"ALLOW_METHODS"`
			} `mapstructure:"CORS"`
			Docs struct {
				Swagger struct {
					Enable   bool   `mapstructure:"ENABLE"`
					FilePath string `mapstructure:"FILE_PATH"`
					Path     string `mapstructure:"PATH"`
					Title    string `mapstructure:"TITLE"`
				} `mapstructure:"SWAGGER"`
			} `mapstructure:"DOCS"`
		} `mapstructure:"HTTP"`
		GRPC struct {
			Port           int           `mapstructure:"PORT"`
			RequestTimeout time.Duration `mapstructure:"REQUEST_TIMEOUT"`
		} `mapstructure:"GRPC"`
		EventConsumer struct {
			DataSourceName string `mapstructure:"DATA_SOURCE_NAME"`
		} `mapstructure:"EVENT_CONSUMER"`
		Tracer struct {
			ServiceName         string `mapstructure:"SERVICE_NAME"`
			ExporterGRPCAddress string `mapstructure:"EXPORTER_GRPC_ADDRESS"`
		} `mapstructure:"TRACER"`
	} `mapstructure:"SERVER"`
	Datasource struct {
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
	} `mapstructure:"DATASOURCE"`
	Guest struct {
		Cache struct {
			Enable   bool          `mapstructure:"ENABLE"`
			Keyf     string        `mapstructure:"KEYF"`
			Duration time.Duration `mapstructure:"DURATION"`
		} `mapstructure:"CACHE"`
		Event struct {
			Created struct {
				Enable bool   `mapstructure:"ENABLE"`
				Topic  string `mapstructure:"TOPIC"`
			} `mapstructure:"CREATED"`
			Deleted struct {
				Enable bool   `mapstructure:"ENABLE"`
				Topic  string `mapstructure:"TOPIC"`
			} `mapstructure:"DELETED"`
			Updated struct {
				Enable bool   `mapstructure:"ENABLE"`
				Topic  string `mapstructure:"TOPIC"`
			} `mapstructure:"UPDATED"`
		} `mapstructure:"EVENT"`
	} `mapstructure:"GUEST"`
}

func Read(cfgpath string) *Config {
	var (
		config *Config
		err    error
	)

	viper.SetConfigFile(cfgpath)
	err = viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	config = &Config{}
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}

	return config
}
