package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name         string
		validateFunc func(t *testing.T)
	}{
		{
			name: "should initialize databaseMigrationCmd",
			validateFunc: func(t *testing.T) {
				assert.NotNil(t, databaseMigrationCmd)
			},
		},
		{
			name: "should initialize httpCmd",
			validateFunc: func(t *testing.T) {
				assert.NotNil(t, httpCmd)
			},
		},
		{
			name: "should initialize grpcCmd",
			validateFunc: func(t *testing.T) {
				assert.NotNil(t, grpcCmd)
			},
		},
		{
			name: "should initialize eventConsumerCmd",
			validateFunc: func(t *testing.T) {
				assert.NotNil(t, eventConsumerCmd)
			},
		},
		{
			name: "should initialize appCmd",
			validateFunc: func(t *testing.T) {
				assert.NotNil(t, appCmd)
			},
		},
		{
			name: "should initialize rootCmd",
			validateFunc: func(t *testing.T) {
				assert.NotNil(t, rootCmd)
			},
		},
		{
			name: "should set correct Long field for rootCmd",
			validateFunc: func(t *testing.T) {
				assert.Equal(t, "boilerplate", rootCmd.Long)
			},
		},
		{
			name: "should add databaseMigrationCmd to rootCmd",
			validateFunc: func(t *testing.T) {
				commands := rootCmd.Commands()
				var found bool
				for _, cmd := range commands {
					if cmd.Use == "migrate" || cmd.Name() == "migrate" {
						found = true
						break
					}
				}
				assert.True(t, found, "databaseMigrationCmd (migrate) should be added to rootCmd")
			},
		},
		{
			name: "should add httpCmd to rootCmd",
			validateFunc: func(t *testing.T) {
				commands := rootCmd.Commands()
				var found bool
				for _, cmd := range commands {
					if cmd.Use == "http" || cmd.Name() == "http" {
						found = true
						break
					}
				}
				assert.True(t, found, "httpCmd should be added to rootCmd")
			},
		},
		{
			name: "should add grpcCmd to rootCmd",
			validateFunc: func(t *testing.T) {
				commands := rootCmd.Commands()
				var found bool
				for _, cmd := range commands {
					if cmd.Use == "grpc" || cmd.Name() == "grpc" {
						found = true
						break
					}
				}
				assert.True(t, found, "grpcCmd should be added to rootCmd")
			},
		},
		{
			name: "should add eventConsumerCmd to rootCmd",
			validateFunc: func(t *testing.T) {
				commands := rootCmd.Commands()
				var found bool
				for _, cmd := range commands {
					if cmd.Use == "event-consumer" || cmd.Name() == "event-consumer" {
						found = true
						break
					}
				}
				assert.True(t, found, "eventConsumerCmd should be added to rootCmd")
			},
		},
		{
			name: "should add appCmd to rootCmd",
			validateFunc: func(t *testing.T) {
				commands := rootCmd.Commands()
				var found bool
				for _, cmd := range commands {
					if cmd.Use == "app" || cmd.Name() == "app" {
						found = true
						break
					}
				}
				assert.True(t, found, "appCmd should be added to rootCmd")
			},
		},
		{
			name: "should have exactly 5 subcommands",
			validateFunc: func(t *testing.T) {
				commands := rootCmd.Commands()
				assert.Equal(t, 5, len(commands))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.validateFunc != nil {
				tt.validateFunc(t)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(t *testing.T)
		validateFunc func(t *testing.T)
	}{
		{
			name: "should panic when Execute returns error",
			setupFunc: func(t *testing.T) {
				// Setup rootCmd with invalid command to trigger error
				rootCmd.SetArgs([]string{"invalid-command"})
			},
			validateFunc: func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						// Expected panic from Execute()
						t.Logf("Expected panic recovered: %v", r)
						assert.NotNil(t, r)
					} else {
						t.Error("Expected Execute() to panic but it didn't")
					}
				}()

				Execute()
			},
		},
		{
			name: "should not panic when Execute succeeds with help flag",
			setupFunc: func(t *testing.T) {
				// Setup rootCmd with valid help flag
				rootCmd.SetArgs([]string{"--help"})
			},
			validateFunc: func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Did not expect panic but got: %v", r)
					}
				}()

				// Execute with --help should not panic
				// Note: This will exit with code 0, but in test it just returns
				Execute()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc(t)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t)
			}
		})
	}
}
