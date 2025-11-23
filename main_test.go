package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func()
		shouldPanic bool
	}{
		{
			name: "should call Execute with help flag",
			setupFunc: func() {
				os.Args = []string{"cmd", "--help"}
			},
			shouldPanic: false,
		},
		{
			name: "should call Execute with no args",
			setupFunc: func() {
				os.Args = []string{"cmd"}
			},
			shouldPanic: false,
		},
		{
			name: "should call Execute with http help",
			setupFunc: func() {
				os.Args = []string{"cmd", "http", "--help"}
			},
			shouldPanic: false,
		},
		{
			name: "should call Execute with grpc help",
			setupFunc: func() {
				os.Args = []string{"cmd", "grpc", "--help"}
			},
			shouldPanic: false,
		},
		{
			name: "should call Execute with event-consumer help",
			setupFunc: func() {
				os.Args = []string{"cmd", "event-consumer", "--help"}
			},
			shouldPanic: false,
		},
		{
			name: "should call Execute with app help",
			setupFunc: func() {
				os.Args = []string{"cmd", "app", "--help"}
			},
			shouldPanic: false,
		},
		{
			name: "should call Execute with migrate help",
			setupFunc: func() {
				os.Args = []string{"cmd", "migrate", "--help"}
			},
			shouldPanic: false,
		},
		{
			name: "should panic with invalid command",
			setupFunc: func() {
				os.Args = []string{"cmd", "invalid-command"}
			},
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			// Test
			if tt.shouldPanic {
				assert.Panics(t, func() {
					main()
				})
			} else {
				assert.NotPanics(t, func() {
					main()
				})
			}
		})
	}
}
