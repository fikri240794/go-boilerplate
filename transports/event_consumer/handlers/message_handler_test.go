package handlers

import (
	"context"
	"errors"
	"go-boilerplate/pkg/constants"
	"testing"

	"github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/assert"
)

func TestNewMessageHandler(t *testing.T) {
	tests := []struct {
		name            string
		setupHandleFunc func() func(ctx context.Context, m *nsq.Message) error
		validate        func(t *testing.T, handler nsq.Handler)
	}{
		{
			name: "should_create_message_handler_successfully",
			setupHandleFunc: func() func(ctx context.Context, m *nsq.Message) error {
				return func(ctx context.Context, m *nsq.Message) error {
					return nil
				}
			},
			validate: func(t *testing.T, handler nsq.Handler) {
				assert.NotNil(t, handler)

				assert.Implements(t, (*nsq.Handler)(nil), handler)
			},
		},
		{
			name: "should_create_message_handler_with_custom_function",
			setupHandleFunc: func() func(ctx context.Context, m *nsq.Message) error {
				return func(ctx context.Context, m *nsq.Message) error {
					return errors.New("custom error")
				}
			},
			validate: func(t *testing.T, handler nsq.Handler) {
				assert.NotNil(t, handler)

				mh, ok := handler.(*messageHandler)
				assert.True(t, ok)
				assert.NotNil(t, mh.handleMessageFunc)
			},
		},
		{
			name: "should_create_message_handler_with_nil_function",
			setupHandleFunc: func() func(ctx context.Context, m *nsq.Message) error {
				return nil
			},
			validate: func(t *testing.T, handler nsq.Handler) {
				assert.NotNil(t, handler)

				mh, ok := handler.(*messageHandler)
				assert.True(t, ok)
				assert.Nil(t, mh.handleMessageFunc)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleFunc := tt.setupHandleFunc()

			handler := NewMessageHandler(handleFunc)

			tt.validate(t, handler)
		})
	}
}

func TestMessageHandler_HandleMessage(t *testing.T) {
	tests := []struct {
		name            string
		setupMessage    func() *nsq.Message
		setupHandleFunc func(t *testing.T) func(ctx context.Context, m *nsq.Message) error
		wantErr         bool
		validateErr     func(t *testing.T, err error)
	}{
		{
			name: "should_handle_message_successfully",
			setupMessage: func() *nsq.Message {
				jsonBody := []byte(`{"tracer_propagator":{"traceparent":"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"},"event_name":"test.event","message":{"data":"test"}}`)
				msg := nsq.NewMessage(nsq.MessageID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, jsonBody)
				return msg
			},
			setupHandleFunc: func(t *testing.T) func(ctx context.Context, m *nsq.Message) error {
				return func(ctx context.Context, m *nsq.Message) error {

					requestID := ctx.Value(constants.ContextKeyRequestID)
					assert.NotNil(t, requestID)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "should_return_error_when_unmarshal_fails",
			setupMessage: func() *nsq.Message {
				msg := nsq.NewMessage(nsq.MessageID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, []byte("invalid json"))
				return msg
			},
			setupHandleFunc: func(t *testing.T) func(ctx context.Context, m *nsq.Message) error {
				return func(ctx context.Context, m *nsq.Message) error {
					t.Error("handleMessageFunc should not be called when unmarshal fails")
					return nil
				}
			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "invalid character")
			},
		},
		{
			name: "should_return_error_when_handle_func_fails",
			setupMessage: func() *nsq.Message {
				jsonBody := []byte(`{"event_name":"test.error","message":{"data":"error test"}}`)
				msg := nsq.NewMessage(nsq.MessageID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, jsonBody)
				return msg
			},
			setupHandleFunc: func(t *testing.T) func(ctx context.Context, m *nsq.Message) error {
				return func(ctx context.Context, m *nsq.Message) error {
					return errors.New("handle function error")
				}
			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, "handle function error", err.Error())
			},
		},
		{
			name: "should_set_request_id_in_context",
			setupMessage: func() *nsq.Message {
				jsonBody := []byte(`{"event_name":"test.context","message":{"data":"context test"}}`)
				msgID := nsq.MessageID{10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120, 130, 140, 150, 160}
				msg := nsq.NewMessage(msgID, jsonBody)
				return msg
			},
			setupHandleFunc: func(t *testing.T) func(ctx context.Context, m *nsq.Message) error {
				return func(ctx context.Context, m *nsq.Message) error {
					requestID := ctx.Value(constants.ContextKeyRequestID)
					assert.NotNil(t, requestID)

					requestIDStr, ok := requestID.(string)
					assert.True(t, ok)
					assert.NotEmpty(t, requestIDStr)

					expectedID := string(m.ID[:])
					assert.Equal(t, expectedID, requestIDStr)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "should_extract_tracer_propagator_to_context",
			setupMessage: func() *nsq.Message {
				jsonBody := []byte(`{"tracer_propagator":{"traceparent":"00-5bf92f3577b34da6a3ce929d0e0e4737-00f067aa0ba902b8-01","tracestate":"key=value"},"event_name":"test.tracer","message":{"data":"tracer test"}}`)
				msg := nsq.NewMessage(nsq.MessageID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, jsonBody)
				return msg
			},
			setupHandleFunc: func(t *testing.T) func(ctx context.Context, m *nsq.Message) error {
				return func(ctx context.Context, m *nsq.Message) error {

					assert.NotNil(t, ctx)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "should_handle_empty_tracer_propagator",
			setupMessage: func() *nsq.Message {
				jsonBody := []byte(`{"tracer_propagator":{},"event_name":"test.empty.tracer","message":{"data":"empty tracer"}}`)
				msg := nsq.NewMessage(nsq.MessageID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, jsonBody)
				return msg
			},
			setupHandleFunc: func(t *testing.T) func(ctx context.Context, m *nsq.Message) error {
				return func(ctx context.Context, m *nsq.Message) error {
					assert.NotNil(t, ctx)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "should_handle_nil_tracer_propagator",
			setupMessage: func() *nsq.Message {
				jsonBody := []byte(`{"tracer_propagator":null,"event_name":"test.nil.tracer","message":{"data":"nil tracer"}}`)
				msg := nsq.NewMessage(nsq.MessageID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, jsonBody)
				return msg
			},
			setupHandleFunc: func(t *testing.T) func(ctx context.Context, m *nsq.Message) error {
				return func(ctx context.Context, m *nsq.Message) error {
					assert.NotNil(t, ctx)
					requestID := ctx.Value(constants.ContextKeyRequestID)
					assert.NotNil(t, requestID)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "should_handle_complex_message_data",
			setupMessage: func() *nsq.Message {
				jsonBody := []byte(`{"event_name":"test.complex","message":{"id":"complex-123","name":"Complex Test","nested":{"key":"value"},"array":[1,2,3],"boolean":true,"number":42.5}}`)
				msg := nsq.NewMessage(nsq.MessageID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, jsonBody)
				return msg
			},
			setupHandleFunc: func(t *testing.T) func(ctx context.Context, m *nsq.Message) error {
				return func(ctx context.Context, m *nsq.Message) error {

					assert.NotEmpty(t, m.Body)
					assert.Contains(t, string(m.Body), "test.complex")
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "should_handle_empty_message_body",
			setupMessage: func() *nsq.Message {
				jsonBody := []byte(`{"event_name":"test.empty","message":null}`)
				msg := nsq.NewMessage(nsq.MessageID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, jsonBody)
				return msg
			},
			setupHandleFunc: func(t *testing.T) func(ctx context.Context, m *nsq.Message) error {
				return func(ctx context.Context, m *nsq.Message) error {
					assert.NotEmpty(t, m.Body)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "should_handle_malformed_json",
			setupMessage: func() *nsq.Message {
				msg := nsq.NewMessage(nsq.MessageID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, []byte("{incomplete json"))
				return msg
			},
			setupHandleFunc: func(t *testing.T) func(ctx context.Context, m *nsq.Message) error {
				return func(ctx context.Context, m *nsq.Message) error {
					t.Error("handleMessageFunc should not be called with malformed JSON")
					return nil
				}
			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "invalid character")
			},
		},
		{
			name: "should_preserve_message_object_in_handler",
			setupMessage: func() *nsq.Message {
				jsonBody := []byte(`{"event_name":"test.preserve","message":{"preserve":"data"}}`)
				msg := nsq.NewMessage(nsq.MessageID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, jsonBody)
				return msg
			},
			setupHandleFunc: func(t *testing.T) func(ctx context.Context, m *nsq.Message) error {
				return func(ctx context.Context, m *nsq.Message) error {

					assert.NotNil(t, m)
					assert.NotEmpty(t, m.Body)
					assert.NotEqual(t, nsq.MessageID{}, m.ID)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "should_handle_message_with_special_characters",
			setupMessage: func() *nsq.Message {
				jsonBody := []byte(`{"event_name":"test.special","message":{"text":"Hello \"World\" with 'quotes' and symbols: @#$%^&*()"}}`)
				msg := nsq.NewMessage(nsq.MessageID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, jsonBody)
				return msg
			},
			setupHandleFunc: func(t *testing.T) func(ctx context.Context, m *nsq.Message) error {
				return func(ctx context.Context, m *nsq.Message) error {
					assert.NotNil(t, ctx)
					requestID := ctx.Value(constants.ContextKeyRequestID)
					assert.NotNil(t, requestID)
					return nil
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.setupMessage()
			handleFunc := tt.setupHandleFunc(t)
			handler := NewMessageHandler(handleFunc)

			err := handler.HandleMessage(msg)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.validateErr != nil {
					tt.validateErr(t, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
