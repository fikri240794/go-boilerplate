package grpc

import (
	"go-boilerplate/configs"
	"go-boilerplate/datasources"
	"go-boilerplate/datasources/boilerplate_database"
	"go-boilerplate/datasources/event_producer"
	"go-boilerplate/datasources/in_memory_database"
	"go-boilerplate/transports/grpc/handlers"
	"go-boilerplate/transports/grpc/middlewares"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewGRPCServer(t *testing.T) {
	tests := []struct {
		name             string
		setupConfig      func() *configs.Config
		setupDatasources func() *datasources.Datasources
		setupHandlers    func() *handlers.ImplementedBoilerplateServer
		setupMiddlewares func() *middlewares.Middlewares
		validate         func(t *testing.T, server *GRPCServer)
	}{
		{
			name: "should_create_new_grpc_server_successfully",
			setupConfig: func() *configs.Config {
				return &configs.Config{}
			},
			setupDatasources: func() *datasources.Datasources {
				return &datasources.Datasources{
					BoilerplateDatabase: &boilerplate_database.BoilerplateDatabase{},
					InMemoryDatabase:    &in_memory_database.InMemoryDatabase{},
					EventProducer:       &event_producer.EventProducer{},
				}
			},
			setupHandlers: func() *handlers.ImplementedBoilerplateServer {
				return &handlers.ImplementedBoilerplateServer{}
			},
			setupMiddlewares: func() *middlewares.Middlewares {
				return &middlewares.Middlewares{}
			},
			validate: func(t *testing.T, server *GRPCServer) {
				assert.NotNil(t, server)
				assert.NotNil(t, server.cfg)
				assert.NotNil(t, server.datasources)
				assert.NotNil(t, server.server)
				assert.NotNil(t, server.handlers)
				assert.NotNil(t, server.middlewares)
			},
		},
		{
			name: "should_register_boilerplate_server_with_grpc_server",
			setupConfig: func() *configs.Config {
				return &configs.Config{}
			},
			setupDatasources: func() *datasources.Datasources {
				return &datasources.Datasources{}
			},
			setupHandlers: func() *handlers.ImplementedBoilerplateServer {
				return &handlers.ImplementedBoilerplateServer{}
			},
			setupMiddlewares: func() *middlewares.Middlewares {
				return &middlewares.Middlewares{}
			},
			validate: func(t *testing.T, server *GRPCServer) {
				assert.NotNil(t, server)
				assert.NotNil(t, server.server)

				info := server.server.GetServiceInfo()
				_, exists := info["protobuf_boilerplate.Boilerplate"]
				assert.True(t, exists, "Boilerplate service should be registered")
			},
		},
		{
			name: "should_chain_unary_interceptors_from_middlewares",
			setupConfig: func() *configs.Config {
				return &configs.Config{}
			},
			setupDatasources: func() *datasources.Datasources {
				return &datasources.Datasources{}
			},
			setupHandlers: func() *handlers.ImplementedBoilerplateServer {
				return &handlers.ImplementedBoilerplateServer{}
			},
			setupMiddlewares: func() *middlewares.Middlewares {
				return &middlewares.Middlewares{}
			},
			validate: func(t *testing.T, server *GRPCServer) {
				assert.NotNil(t, server)
				assert.NotNil(t, server.server)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()
			ds := tt.setupDatasources()
			h := tt.setupHandlers()
			mw := tt.setupMiddlewares()

			server := NewGRPCServer(cfg, ds, h, mw)

			tt.validate(t, server)
		})
	}
}

func TestGRPCServer_setGlobalLog(t *testing.T) {
	tests := []struct {
		name        string
		setupServer func() *GRPCServer
		validate    func(t *testing.T)
	}{
		{
			name: "should_set_global_log_level_to_debug",
			setupServer: func() *GRPCServer {
				cfg := &configs.Config{}
				cfg.Server.LogLevel = int8(zerolog.DebugLevel)
				ds := &datasources.Datasources{}
				h := &handlers.ImplementedBoilerplateServer{}
				mw := &middlewares.Middlewares{}

				return NewGRPCServer(cfg, ds, h, mw)
			},
			validate: func(t *testing.T) {
				assert.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel())
			},
		},
		{
			name: "should_set_global_log_level_to_info",
			setupServer: func() *GRPCServer {
				cfg := &configs.Config{}
				cfg.Server.LogLevel = int8(zerolog.InfoLevel)
				ds := &datasources.Datasources{}
				h := &handlers.ImplementedBoilerplateServer{}
				mw := &middlewares.Middlewares{}

				return NewGRPCServer(cfg, ds, h, mw)
			},
			validate: func(t *testing.T) {
				assert.Equal(t, zerolog.InfoLevel, zerolog.GlobalLevel())
			},
		},
		{
			name: "should_set_global_log_level_to_error",
			setupServer: func() *GRPCServer {
				cfg := &configs.Config{}
				cfg.Server.LogLevel = int8(zerolog.ErrorLevel)
				ds := &datasources.Datasources{}
				h := &handlers.ImplementedBoilerplateServer{}
				mw := &middlewares.Middlewares{}

				return NewGRPCServer(cfg, ds, h, mw)
			},
			validate: func(t *testing.T) {
				assert.Equal(t, zerolog.ErrorLevel, zerolog.GlobalLevel())
			},
		},
		{
			name: "should_set_global_log_level_to_warn",
			setupServer: func() *GRPCServer {
				cfg := &configs.Config{}
				cfg.Server.LogLevel = int8(zerolog.WarnLevel)
				ds := &datasources.Datasources{}
				h := &handlers.ImplementedBoilerplateServer{}
				mw := &middlewares.Middlewares{}

				return NewGRPCServer(cfg, ds, h, mw)
			},
			validate: func(t *testing.T) {
				assert.Equal(t, zerolog.WarnLevel, zerolog.GlobalLevel())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer()

			server.setGlobalLog()

			tt.validate(t)
		})
	}
}
