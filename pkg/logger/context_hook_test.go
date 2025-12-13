package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"go-boilerplate/pkg/constants"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestNewContextHook(t *testing.T) {
	hook := NewContextHook()

	assert.NotNil(t, hook, "Expected non-nil ContextHook")
	assert.IsType(t, ContextHook{}, hook, "Expected ContextHook type")
}

func TestContextHook_Run(t *testing.T) {
	tests := []struct {
		name                string
		setupContext        func() context.Context
		useContext          bool
		level               zerolog.Level
		msg                 string
		expectedRequestID   string
		expectedTraceID     bool
		shouldHaveRequestID bool
		directCall          bool
		shouldNotPanic      bool
	}{
		{
			name: "context with requestid should add requestid field",
			setupContext: func() context.Context {
				return context.WithValue(
					context.Background(),
					constants.ContextKeyRequestID,
					"test-request-id-123",
				)
			},
			useContext:          true,
			level:               zerolog.InfoLevel,
			msg:                 "test message",
			expectedRequestID:   "test-request-id-123",
			shouldHaveRequestID: true,
			expectedTraceID:     false,
			directCall:          false,
			shouldNotPanic:      false,
		},
		{
			name: "event without context should not add any fields",
			setupContext: func() context.Context {
				return context.Background()
			},
			useContext:          false,
			level:               zerolog.InfoLevel,
			msg:                 "test message",
			shouldHaveRequestID: false,
			expectedTraceID:     false,
			directCall:          false,
			shouldNotPanic:      false,
		},
		{
			name: "empty context should not add any fields",
			setupContext: func() context.Context {
				return context.Background()
			},
			useContext:          true,
			level:               zerolog.InfoLevel,
			msg:                 "test message",
			shouldHaveRequestID: false,
			expectedTraceID:     false,
			directCall:          false,
			shouldNotPanic:      false,
		},
		{
			name: "context with empty requestid should not add requestid field",
			setupContext: func() context.Context {
				return context.WithValue(
					context.Background(),
					constants.ContextKeyRequestID,
					"",
				)
			},
			useContext:          true,
			level:               zerolog.InfoLevel,
			msg:                 "test message",
			shouldHaveRequestID: false,
			expectedTraceID:     false,
			directCall:          false,
			shouldNotPanic:      false,
		},
		{
			name: "context with valid trace span should add traceid and spanid fields",
			setupContext: func() context.Context {
				// Create a real tracer with in-memory exporter
				exporter := tracetest.NewInMemoryExporter()
				tp := trace.NewTracerProvider(
					trace.WithSyncer(exporter),
				)
				tracer := tp.Tracer("test-tracer")

				// Start a real span
				ctx, span := tracer.Start(context.Background(), "test-operation")
				defer span.End()

				return ctx
			},
			useContext:          true,
			level:               zerolog.ErrorLevel,
			msg:                 "error message",
			shouldHaveRequestID: false,
			expectedTraceID:     true,
			directCall:          false,
			shouldNotPanic:      false,
		},
		{
			name: "context with both requestid and valid trace should add all fields",
			setupContext: func() context.Context {
				exporter := tracetest.NewInMemoryExporter()
				tp := trace.NewTracerProvider(
					trace.WithSyncer(exporter),
				)
				tracer := tp.Tracer("test-tracer")

				ctx, span := tracer.Start(context.Background(), "combined-operation")
				defer span.End()

				return context.WithValue(
					ctx,
					constants.ContextKeyRequestID,
					"combined-request-id",
				)
			},
			useContext:          true,
			level:               zerolog.WarnLevel,
			msg:                 "warning message",
			expectedRequestID:   "combined-request-id",
			shouldHaveRequestID: true,
			expectedTraceID:     true,
			directCall:          false,
			shouldNotPanic:      false,
		},
		{
			name: "different log levels should work the same",
			setupContext: func() context.Context {
				return context.WithValue(
					context.Background(),
					constants.ContextKeyRequestID,
					"debug-request-id",
				)
			},
			useContext:          true,
			level:               zerolog.DebugLevel,
			msg:                 "debug message",
			expectedRequestID:   "debug-request-id",
			shouldHaveRequestID: true,
			expectedTraceID:     false,
			directCall:          false,
			shouldNotPanic:      false,
		},
		{
			name: "context with invalid span should not add trace fields",
			setupContext: func() context.Context {
				return context.WithValue(
					context.Background(),
					constants.ContextKeyRequestID,
					"request-with-invalid-span",
				)
			},
			useContext:          true,
			level:               zerolog.InfoLevel,
			msg:                 "test invalid span",
			expectedRequestID:   "request-with-invalid-span",
			shouldHaveRequestID: true,
			expectedTraceID:     false,
			directCall:          false,
			shouldNotPanic:      false,
		},
		{
			name: "event with nil context from GetCtx should not panic",
			setupContext: func() context.Context {
				return context.Background()
			},
			useContext:          false,
			level:               zerolog.InfoLevel,
			msg:                 "test",
			shouldHaveRequestID: false,
			expectedTraceID:     false,
			directCall:          true,
			shouldNotPanic:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := NewContextHook()

			if tt.directCall {
				// Create a buffer to capture log output
				var buf bytes.Buffer
				logger := zerolog.New(&buf)

				// Create log event based on level without hook
				var event *zerolog.Event
				switch tt.level {
				case zerolog.DebugLevel:
					event = logger.Debug()
				case zerolog.InfoLevel:
					event = logger.Info()
				case zerolog.WarnLevel:
					event = logger.Warn()
				case zerolog.ErrorLevel:
					event = logger.Error()
				default:
					event = logger.Info()
				}

				ctx := tt.setupContext()
				if tt.useContext {
					event = event.Ctx(ctx)
				}

				// Call hook.Run directly
				if tt.shouldNotPanic {
					assert.NotPanics(t, func() {
						hook.Run(event, tt.level, tt.msg)
					}, "Run should not panic")
				}
			} else {
				// Create a buffer to capture log output
				var buf bytes.Buffer
				logger := zerolog.New(&buf)

				logger = logger.Hook(hook)

				ctx := tt.setupContext()

				// Create log event based on level
				var event *zerolog.Event
				switch tt.level {
				case zerolog.DebugLevel:
					event = logger.Debug()
				case zerolog.InfoLevel:
					event = logger.Info()
				case zerolog.WarnLevel:
					event = logger.Warn()
				case zerolog.ErrorLevel:
					event = logger.Error()
				default:
					event = logger.Info()
				}

				if tt.useContext {
					event = event.Ctx(ctx)
				}

				event.Msg(tt.msg)

				// Parse the JSON log output
				var logOutput map[string]interface{}
				if buf.Len() > 0 {
					err := json.Unmarshal(buf.Bytes(), &logOutput)
					assert.NoError(t, err, "Should parse log output as JSON")

					// Verify requestid field
					if tt.shouldHaveRequestID {
						assert.Contains(t, logOutput, "requestid", "Should contain requestid field")
						assert.Equal(t, tt.expectedRequestID, logOutput["requestid"], "RequestID should match")
					} else {
						assert.NotContains(t, logOutput, "requestid", "Should not contain requestid field")
					}

					// Verify traceid and spanid fields
					if tt.expectedTraceID {
						assert.Contains(t, logOutput, "traceid", "Should contain traceid field")
						assert.NotEmpty(t, logOutput["traceid"], "TraceID should not be empty")
					}

					// Verify message
					assert.Equal(t, tt.msg, logOutput["message"], "Message should match")
				}
			}
		})
	}
}
