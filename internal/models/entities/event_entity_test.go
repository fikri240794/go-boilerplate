package entities

import (
	"context"
	"testing"

	"go-boilerplate/pkg/constants"

	"github.com/stretchr/testify/assert"
)

type TestMessage struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

type AnotherTestMessage struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestNewEventEntity(t *testing.T) {

	tests := []struct {
		name      string
		eventName string
		message   interface{}
		validate  func(t *testing.T, result interface{})
	}{
		{
			name:      "create event entity with TestMessage",
			eventName: "user.created",
			message:   &TestMessage{ID: 1, Content: "test message"},
			validate: func(t *testing.T, result interface{}) {
				entity, ok := result.(*EventEntity[TestMessage])
				assert.True(t, ok, "Result should be *EventEntity[TestMessage]")
				assert.Equal(t, "user.created", entity.Name)
				assert.NotNil(t, entity.Message)
				assert.Equal(t, 1, entity.Message.ID)
				assert.Equal(t, "test message", entity.Message.Content)
				assert.Nil(t, entity.TracerPropagator)
			},
		},
		{
			name:      "create event entity with AnotherTestMessage",
			eventName: "order.processed",
			message:   &AnotherTestMessage{Name: "order123", Value: 100},
			validate: func(t *testing.T, result interface{}) {
				entity, ok := result.(*EventEntity[AnotherTestMessage])
				assert.True(t, ok, "Result should be *EventEntity[AnotherTestMessage]")
				assert.Equal(t, "order.processed", entity.Name)
				assert.NotNil(t, entity.Message)
				assert.Equal(t, "order123", entity.Message.Name)
				assert.Equal(t, 100, entity.Message.Value)
			},
		},
		{
			name:      "create event entity with nil message",
			eventName: "system.shutdown",
			message:   (*TestMessage)(nil),
			validate: func(t *testing.T, result interface{}) {
				entity, ok := result.(*EventEntity[TestMessage])
				assert.True(t, ok, "Result should be *EventEntity[TestMessage]")
				assert.Equal(t, "system.shutdown", entity.Name)
				assert.Nil(t, entity.Message)
			},
		},
		{
			name:      "create event entity with empty event name",
			eventName: "",
			message:   &TestMessage{ID: 2, Content: "empty name test"},
			validate: func(t *testing.T, result interface{}) {
				entity, ok := result.(*EventEntity[TestMessage])
				assert.True(t, ok, "Result should be *EventEntity[TestMessage]")
				assert.Equal(t, "", entity.Name)
				assert.NotNil(t, entity.Message)
				assert.Equal(t, 2, entity.Message.ID)
			},
		},
		{
			name:      "create event entity with string type",
			eventName: "notification.sent",
			message:   func() *string { s := "Hello World"; return &s }(),
			validate: func(t *testing.T, result interface{}) {
				entity, ok := result.(*EventEntity[string])
				assert.True(t, ok, "Result should be *EventEntity[string]")
				assert.Equal(t, "notification.sent", entity.Name)
				assert.NotNil(t, entity.Message)
				assert.Equal(t, "Hello World", *entity.Message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result interface{}

			switch msg := tt.message.(type) {
			case *TestMessage:
				result = NewEventEntity(tt.eventName, msg)
			case *AnotherTestMessage:
				result = NewEventEntity(tt.eventName, msg)
			case *string:
				result = NewEventEntity(tt.eventName, msg)
			default:
				assert.Fail(t, "Unsupported message type: %T", tt.message)
				return
			}

			assert.NotNil(t, result)

			tt.validate(t, result)
		})
	}
}

func TestEventEntity_InjectTracerPropagator(t *testing.T) {
	type contextKey string
	const testKey contextKey = "test-key"

	tests := []struct {
		name     string
		entity   *EventEntity[TestMessage]
		ctx      context.Context
		validate func(t *testing.T, resultCtx context.Context, resultEntity *EventEntity[TestMessage], originalEntity *EventEntity[TestMessage])
	}{
		{
			name: "inject tracer propagator with background context",
			entity: &EventEntity[TestMessage]{
				Name:             "test.event",
				Message:          &TestMessage{ID: 1, Content: "test"},
				TracerPropagator: nil,
			},
			ctx: context.Background(),
			validate: func(t *testing.T, resultCtx context.Context, resultEntity *EventEntity[TestMessage], originalEntity *EventEntity[TestMessage]) {
				assert.NotNil(t, resultCtx)
				assert.NotNil(t, resultEntity)
				assert.Equal(t, originalEntity, resultEntity)
				assert.NotNil(t, resultEntity.TracerPropagator)
				assert.Equal(t, "test.event", resultEntity.Name)
				assert.NotNil(t, resultEntity.Message)
				assert.Equal(t, 1, resultEntity.Message.ID)
			},
		},
		{
			name: "inject tracer propagator with existing TracerPropagator",
			entity: &EventEntity[TestMessage]{
				Name:             "existing.event",
				Message:          &TestMessage{ID: 2, Content: "existing"},
				TracerPropagator: map[string]string{"existing": "value"},
			},
			ctx: context.Background(),
			validate: func(t *testing.T, resultCtx context.Context, resultEntity *EventEntity[TestMessage], originalEntity *EventEntity[TestMessage]) {
				assert.NotNil(t, resultEntity.TracerPropagator)
			},
		},
		{
			name: "inject tracer propagator with context containing values",
			entity: &EventEntity[TestMessage]{
				Name:             "context.event",
				Message:          &TestMessage{ID: 3, Content: "context test"},
				TracerPropagator: nil,
			},
			ctx: context.WithValue(context.Background(), testKey, "test-value"),
			validate: func(t *testing.T, resultCtx context.Context, resultEntity *EventEntity[TestMessage], originalEntity *EventEntity[TestMessage]) {
				assert.NotNil(t, resultEntity.TracerPropagator)
				assert.Equal(t, "test-value", resultCtx.Value(testKey))
			},
		},
		{
			name: "inject tracer propagator multiple times",
			entity: &EventEntity[TestMessage]{
				Name:             "multiple.event",
				Message:          &TestMessage{ID: 4, Content: "multiple test"},
				TracerPropagator: nil,
			},
			ctx: context.Background(),
			validate: func(t *testing.T, resultCtx context.Context, resultEntity *EventEntity[TestMessage], originalEntity *EventEntity[TestMessage]) {
				assert.NotNil(t, resultEntity.TracerPropagator)

				ctx2, entity2 := resultEntity.InjectTracerPropagator(context.Background())

				assert.NotNil(t, ctx2)
				assert.Equal(t, resultEntity, entity2)
				assert.NotNil(t, entity2.TracerPropagator)
			},
		},
		{
			name: "inject tracer propagator with requestid in context",
			entity: &EventEntity[TestMessage]{
				Name:             "requestid.event",
				Message:          &TestMessage{ID: 5, Content: "requestid test"},
				TracerPropagator: nil,
			},
			ctx: context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-123"),
			validate: func(t *testing.T, resultCtx context.Context, resultEntity *EventEntity[TestMessage], originalEntity *EventEntity[TestMessage]) {
				assert.NotNil(t, resultEntity.TracerPropagator)
				assert.Contains(t, resultEntity.TracerPropagator, string(constants.ContextKeyRequestID))
				assert.Equal(t, "test-request-123", resultEntity.TracerPropagator[string(constants.ContextKeyRequestID)])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultCtx, resultEntity := tt.entity.InjectTracerPropagator(tt.ctx)

			tt.validate(t, resultCtx, resultEntity, tt.entity)
		})
	}
}
