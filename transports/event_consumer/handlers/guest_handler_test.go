package handlers

import (
	"context"
	"errors"
	"go-boilerplate/internal/models/dtos"
	"go-boilerplate/internal/services"
	"go-boilerplate/internal/services/mocks"
	"go-boilerplate/pkg/constants"
	"go-boilerplate/transports/event_consumer/models/vms"
	"testing"

	"github.com/goccy/go-json"
	"github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewGuestHandler(t *testing.T) {
	tests := []struct {
		name         string
		setupService func() services.IGuestService
		validate     func(t *testing.T, handler *GuestHandler)
	}{
		{
			name: "should_create_guest_handler_successfully",
			setupService: func() services.IGuestService {
				return mocks.NewGuestServiceMock(t)
			},
			validate: func(t *testing.T, handler *GuestHandler) {
				assert.NotNil(t, handler)
				assert.NotNil(t, handler.guestService)
			},
		},
		{
			name: "should_create_guest_handler_with_nil_service",
			setupService: func() services.IGuestService {
				return nil
			},
			validate: func(t *testing.T, handler *GuestHandler) {
				assert.NotNil(t, handler)
				assert.Nil(t, handler.guestService)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService()

			handler := NewGuestHandler(service)

			tt.validate(t, handler)
		})
	}
}

func TestGuestHandler_HandleCreated(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func() context.Context
		setupMessage func() *nsq.Message
		setupMock    func(mock *mocks.GuestServiceMock)
		wantErr      bool
		validateErr  func(t *testing.T, err error)
	}{
		{
			name: "should_handle_created_event_successfully",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-123")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[vms.GuestEventRequestVM]{
					Name: "guest.created",
					Message: &vms.GuestEventRequestVM{
						ID:        "guest-123",
						Name:      "John Doe",
						Address:   "123 Main St",
						CreatedAt: 1700000000,
						CreatedBy: "user-1",
					},
					TracerPropagator: map[string]string{
						"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-123",
					Name:      "John Doe",
					Address:   "123 Main St",
					CreatedAt: 1700000000,
					CreatedBy: "user-1",
				}).Return(&dtos.GuestEventResponseDTO{
					ID: "guest-123",
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "should_return_error_when_unmarshal_fails",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-456")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				return &nsq.Message{Body: []byte("invalid json")}
			},
			setupMock: func(mock *mocks.GuestServiceMock) {

			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "invalid character")
			},
		},
		{
			name: "should_return_error_when_message_is_nil",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-789")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[vms.GuestEventRequestVM]{
					Name:    "guest.created",
					Message: nil,
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mock *mocks.GuestServiceMock) {

			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "message is nil")
			},
		},
		{
			name: "should_return_error_when_process_event_fails",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-999")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[vms.GuestEventRequestVM]{
					Name: "guest.created",
					Message: &vms.GuestEventRequestVM{
						ID:        "guest-999",
						Name:      "Jane Smith",
						CreatedAt: 1700001000,
						CreatedBy: "user-2",
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-999",
					Name:      "Jane Smith",
					CreatedAt: 1700001000,
					CreatedBy: "user-2",
				}).Return((*dtos.GuestEventResponseDTO)(nil), errors.New("service error"))
			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "service error")
			},
		},
		{
			name: "should_handle_empty_tracer_propagator",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-empty-tracer")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[vms.GuestEventRequestVM]{
					Name: "guest.created",
					Message: &vms.GuestEventRequestVM{
						ID:        "guest-empty",
						Name:      "Empty Tracer",
						CreatedAt: 1700002000,
						CreatedBy: "user-3",
					},
					TracerPropagator: map[string]string{},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-empty",
					Name:      "Empty Tracer",
					CreatedAt: 1700002000,
					CreatedBy: "user-3",
				}).Return(&dtos.GuestEventResponseDTO{
					ID: "guest-empty",
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "should_handle_complete_guest_data",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constants.ContextKeyRequestID, "req-complete")
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[vms.GuestEventRequestVM]{
					Name: "guest.created",
					Message: &vms.GuestEventRequestVM{
						ID:        "guest-complete",
						Name:      "Complete Guest",
						Address:   "456 Oak Ave",
						CreatedAt: 1700003000,
						CreatedBy: "admin",
						UpdatedAt: 1700003100,
						UpdatedBy: "admin",
						DeletedAt: 1700003200,
						DeletedBy: "admin",
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-complete",
					Name:      "Complete Guest",
					Address:   "456 Oak Ave",
					CreatedAt: 1700003000,
					CreatedBy: "admin",
					UpdatedAt: 1700003100,
					UpdatedBy: "admin",
					DeletedAt: 1700003200,
					DeletedBy: "admin",
				}).Return(&dtos.GuestEventResponseDTO{
					ID: "guest-complete",
				}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewGuestServiceMock(t)
			tt.setupMock(mockService)

			handler := NewGuestHandler(mockService)
			ctx := tt.setupContext()
			msg := tt.setupMessage()

			err := handler.HandleCreated(ctx, msg)

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

func TestGuestHandler_HandleDeleted(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func() context.Context
		setupMessage func() *nsq.Message
		setupMock    func(mock *mocks.GuestServiceMock)
		wantErr      bool
		validateErr  func(t *testing.T, err error)
	}{
		{
			name: "should_handle_deleted_event_successfully",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-del-123")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[vms.GuestEventRequestVM]{
					Name: "guest.deleted",
					Message: &vms.GuestEventRequestVM{
						ID:        "guest-del-123",
						Name:      "Delete Test",
						DeletedAt: 1700004000,
						DeletedBy: "admin",
					},
					TracerPropagator: map[string]string{
						"traceparent": "00-5bf92f3577b34da6a3ce929d0e0e4737-00f067aa0ba902b8-01",
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-del-123",
					Name:      "Delete Test",
					DeletedAt: 1700004000,
					DeletedBy: "admin",
				}).Return(&dtos.GuestEventResponseDTO{
					ID: "guest-del-123",
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "should_return_error_when_unmarshal_fails",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-del-invalid")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				return &nsq.Message{Body: []byte("{invalid json}")}
			},
			setupMock: func(mock *mocks.GuestServiceMock) {

			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "invalid character")
			},
		},
		{
			name: "should_return_error_when_message_is_nil",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-del-nil")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[vms.GuestEventRequestVM]{
					Name:    "guest.deleted",
					Message: nil,
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mock *mocks.GuestServiceMock) {

			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "message is nil")
			},
		},
		{
			name: "should_return_error_when_process_event_fails",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-del-fail")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[vms.GuestEventRequestVM]{
					Name: "guest.deleted",
					Message: &vms.GuestEventRequestVM{
						ID:        "guest-del-fail",
						Name:      "Fail Test",
						DeletedAt: 1700005000,
						DeletedBy: "admin",
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-del-fail",
					Name:      "Fail Test",
					DeletedAt: 1700005000,
					DeletedBy: "admin",
				}).Return((*dtos.GuestEventResponseDTO)(nil), errors.New("delete service error"))
			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "delete service error")
			},
		},
		{
			name: "should_handle_soft_delete_with_full_data",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constants.ContextKeyRequestID, "req-soft-del")
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[vms.GuestEventRequestVM]{
					Name: "guest.deleted",
					Message: &vms.GuestEventRequestVM{
						ID:        "guest-soft-del",
						Name:      "Soft Delete",
						Address:   "789 Pine Rd",
						CreatedAt: 1700006000,
						CreatedBy: "system",
						UpdatedAt: 1700006100,
						UpdatedBy: "system",
						DeletedAt: 1700006200,
						DeletedBy: "system",
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-soft-del",
					Name:      "Soft Delete",
					Address:   "789 Pine Rd",
					CreatedAt: 1700006000,
					CreatedBy: "system",
					UpdatedAt: 1700006100,
					UpdatedBy: "system",
					DeletedAt: 1700006200,
					DeletedBy: "system",
				}).Return(&dtos.GuestEventResponseDTO{
					ID: "guest-soft-del",
				}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewGuestServiceMock(t)
			tt.setupMock(mockService)

			handler := NewGuestHandler(mockService)
			ctx := tt.setupContext()
			msg := tt.setupMessage()

			err := handler.HandleDeleted(ctx, msg)

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

func TestGuestHandler_HandleUpdated(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func() context.Context
		setupMessage func() *nsq.Message
		setupMock    func(mock *mocks.GuestServiceMock)
		wantErr      bool
		validateErr  func(t *testing.T, err error)
	}{
		{
			name: "should_handle_updated_event_successfully",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-upd-123")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[vms.GuestEventRequestVM]{
					Name: "guest.updated",
					Message: &vms.GuestEventRequestVM{
						ID:        "guest-upd-123",
						Name:      "Updated Guest",
						Address:   "Updated Address",
						CreatedAt: 1700007000,
						CreatedBy: "user-1",
						UpdatedAt: 1700007100,
						UpdatedBy: "user-2",
					},
					TracerPropagator: map[string]string{
						"traceparent": "00-6bf92f3577b34da6a3ce929d0e0e4738-00f067aa0ba902b9-01",
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-upd-123",
					Name:      "Updated Guest",
					Address:   "Updated Address",
					CreatedAt: 1700007000,
					CreatedBy: "user-1",
					UpdatedAt: 1700007100,
					UpdatedBy: "user-2",
				}).Return(&dtos.GuestEventResponseDTO{
					ID: "guest-upd-123",
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "should_return_error_when_unmarshal_fails",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-upd-invalid")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				return &nsq.Message{Body: []byte("not a json")}
			},
			setupMock: func(mock *mocks.GuestServiceMock) {

			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "invalid character")
			},
		},
		{
			name: "should_return_error_when_message_is_nil",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-upd-nil")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[vms.GuestEventRequestVM]{
					Name:    "guest.updated",
					Message: nil,
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mock *mocks.GuestServiceMock) {

			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "message is nil")
			},
		},
		{
			name: "should_return_error_when_process_event_fails",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-upd-fail")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[vms.GuestEventRequestVM]{
					Name: "guest.updated",
					Message: &vms.GuestEventRequestVM{
						ID:        "guest-upd-fail",
						Name:      "Update Fail",
						UpdatedAt: 1700008000,
						UpdatedBy: "user-3",
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-upd-fail",
					Name:      "Update Fail",
					UpdatedAt: 1700008000,
					UpdatedBy: "user-3",
				}).Return((*dtos.GuestEventResponseDTO)(nil), errors.New("update service error"))
			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "update service error")
			},
		},
		{
			name: "should_handle_partial_update",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constants.ContextKeyRequestID, "req-partial-upd")
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[vms.GuestEventRequestVM]{
					Name: "guest.updated",
					Message: &vms.GuestEventRequestVM{
						ID:        "guest-partial",
						Name:      "Partial Update",
						UpdatedAt: 1700009000,
						UpdatedBy: "moderator",
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-partial",
					Name:      "Partial Update",
					UpdatedAt: 1700009000,
					UpdatedBy: "moderator",
				}).Return(&dtos.GuestEventResponseDTO{
					ID: "guest-partial",
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "should_handle_name_and_address_update",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), constants.ContextKeyRequestID, "req-name-addr-upd")
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[vms.GuestEventRequestVM]{
					Name: "guest.updated",
					Message: &vms.GuestEventRequestVM{
						ID:        "guest-name-addr",
						Name:      "New Name",
						Address:   "New Address 123",
						CreatedAt: 1700010000,
						CreatedBy: "admin",
						UpdatedAt: 1700010100,
						UpdatedBy: "admin",
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-name-addr",
					Name:      "New Name",
					Address:   "New Address 123",
					CreatedAt: 1700010000,
					CreatedBy: "admin",
					UpdatedAt: 1700010100,
					UpdatedBy: "admin",
				}).Return(&dtos.GuestEventResponseDTO{
					ID: "guest-name-addr",
				}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewGuestServiceMock(t)
			tt.setupMock(mockService)

			handler := NewGuestHandler(mockService)
			ctx := tt.setupContext()
			msg := tt.setupMessage()

			err := handler.HandleUpdated(ctx, msg)

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

func TestGuestHandler_HandleBulkCreated(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func() context.Context
		setupMessage func() *nsq.Message
		setupMock    func(mock *mocks.GuestServiceMock)
		wantErr      bool
		validateErr  func(t *testing.T, err error)
	}{
		{
			name: "should_handle_bulk_created_event_successfully",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-bulk-1")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[[]vms.GuestEventRequestVM]{
					Name: "guest.bulk.created",
					Message: &[]vms.GuestEventRequestVM{
						{
							ID:        "guest-1",
							Name:      "John Doe",
							Address:   "123 Main St",
							CreatedAt: 1700000000,
							CreatedBy: "user-1",
						},
					},
					TracerPropagator: map[string]string{
						"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-1",
					Name:      "John Doe",
					Address:   "123 Main St",
					CreatedAt: 1700000000,
					CreatedBy: "user-1",
				}).Return(&dtos.GuestEventResponseDTO{
					ID: "guest-1",
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "should_return_error_when_bulk_created_unmarshal_fails",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-bulk-2")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				return &nsq.Message{Body: []byte("invalid json")}
			},
			setupMock: func(mock *mocks.GuestServiceMock) {},
			wantErr:   true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "invalid character")
			},
		},
		{
			name: "should_return_error_when_bulk_created_message_is_nil",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-bulk-3")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[[]vms.GuestEventRequestVM]{
					Name:    "guest.bulk.created",
					Message: nil,
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mock *mocks.GuestServiceMock) {},
			wantErr:   true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "message is nil")
			},
		},
		{
			name: "should_return_error_when_bulk_created_process_event_fails",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-bulk-4")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[[]vms.GuestEventRequestVM]{
					Name: "guest.bulk.created",
					Message: &[]vms.GuestEventRequestVM{
						{
							ID:        "guest-2",
							Name:      "Jane Smith",
							CreatedAt: 1700001000,
							CreatedBy: "user-2",
						},
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-2",
					Name:      "Jane Smith",
					CreatedAt: 1700001000,
					CreatedBy: "user-2",
				}).Return((*dtos.GuestEventResponseDTO)(nil), errors.New("service error"))
			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "service error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewGuestServiceMock(t)
			tt.setupMock(mockService)

			handler := NewGuestHandler(mockService)
			ctx := tt.setupContext()
			msg := tt.setupMessage()

			err := handler.HandleBulkCreated(ctx, msg)

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

func TestGuestHandler_HandleBulkUpdated(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func() context.Context
		setupMessage func() *nsq.Message
		setupMock    func(mock *mocks.GuestServiceMock)
		wantErr      bool
		validateErr  func(t *testing.T, err error)
	}{
		{
			name: "should_handle_bulk_updated_event_successfully",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-bulk-upd-1")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[[]vms.GuestEventRequestVM]{
					Name: "guest.bulk.updated",
					Message: &[]vms.GuestEventRequestVM{
						{
							ID:        "guest-upd-1",
							Name:      "Updated Name",
							UpdatedAt: 1700002000,
							UpdatedBy: "user-upd",
						},
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-upd-1",
					Name:      "Updated Name",
					UpdatedAt: 1700002000,
					UpdatedBy: "user-upd",
				}).Return(&dtos.GuestEventResponseDTO{
					ID: "guest-upd-1",
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "should_return_error_when_bulk_updated_unmarshal_fails",
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMessage: func() *nsq.Message {
				return &nsq.Message{Body: []byte("invalid json")}
			},
			setupMock: func(mock *mocks.GuestServiceMock) {},
			wantErr:   true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "invalid character")
			},
		},
		{
			name: "should_return_error_when_bulk_updated_message_is_nil",
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[[]vms.GuestEventRequestVM]{
					Name:    "guest.bulk.updated",
					Message: nil,
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mock *mocks.GuestServiceMock) {},
			wantErr:   true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "message is nil")
			},
		},
		{
			name: "should_return_error_when_bulk_updated_process_event_fails",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-bulk-upd-err")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[[]vms.GuestEventRequestVM]{
					Name: "guest.bulk.updated",
					Message: &[]vms.GuestEventRequestVM{
						{
							ID:        "guest-upd-err",
							Name:      "Error Test",
							UpdatedAt: 1700005000,
							UpdatedBy: "user-err",
						},
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-upd-err",
					Name:      "Error Test",
					UpdatedAt: 1700005000,
					UpdatedBy: "user-err",
				}).Return((*dtos.GuestEventResponseDTO)(nil), errors.New("service error"))
			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "service error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewGuestServiceMock(t)
			tt.setupMock(mockService)

			handler := NewGuestHandler(mockService)
			ctx := tt.setupContext()
			msg := tt.setupMessage()

			err := handler.HandleBulkUpdated(ctx, msg)

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

func TestGuestHandler_HandleBulkDeleted(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func() context.Context
		setupMessage func() *nsq.Message
		setupMock    func(mock *mocks.GuestServiceMock)
		wantErr      bool
		validateErr  func(t *testing.T, err error)
	}{
		{
			name: "should_handle_bulk_deleted_event_successfully",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-bulk-del-1")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[[]vms.GuestEventRequestVM]{
					Name: "guest.bulk.deleted",
					Message: &[]vms.GuestEventRequestVM{
						{
							ID:        "guest-del-1",
							DeletedAt: 1700003000,
							DeletedBy: "user-del",
						},
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-del-1",
					DeletedAt: 1700003000,
					DeletedBy: "user-del",
				}).Return(&dtos.GuestEventResponseDTO{
					ID: "guest-del-1",
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "should_return_error_when_bulk_deleted_unmarshal_fails",
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMessage: func() *nsq.Message {
				return &nsq.Message{Body: []byte("invalid json")}
			},
			setupMock: func(mock *mocks.GuestServiceMock) {},
			wantErr:   true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "invalid character")
			},
		},
		{
			name: "should_return_error_when_bulk_deleted_message_is_nil",
			setupContext: func() context.Context {
				return context.Background()
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[[]vms.GuestEventRequestVM]{
					Name:    "guest.bulk.deleted",
					Message: nil,
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mock *mocks.GuestServiceMock) {},
			wantErr:   true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "message is nil")
			},
		},
		{
			name: "should_return_error_when_bulk_deleted_process_event_fails",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, constants.ContextKeyRequestID, "req-bulk-del-err")
				return ctx
			},
			setupMessage: func() *nsq.Message {
				eventVM := vms.EventRequestVM[[]vms.GuestEventRequestVM]{
					Name: "guest.bulk.deleted",
					Message: &[]vms.GuestEventRequestVM{
						{
							ID:        "guest-del-err",
							DeletedAt: 1700006000,
							DeletedBy: "user-del-err",
						},
					},
				}
				body, _ := json.Marshal(eventVM)
				return &nsq.Message{Body: body}
			},
			setupMock: func(mockService *mocks.GuestServiceMock) {
				mockService.On("ProcessEvent", mock.Anything, &dtos.GuestEventRequestDTO{
					ID:        "guest-del-err",
					DeletedAt: 1700006000,
					DeletedBy: "user-del-err",
				}).Return((*dtos.GuestEventResponseDTO)(nil), errors.New("service error"))
			},
			wantErr: true,
			validateErr: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "service error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewGuestServiceMock(t)
			tt.setupMock(mockService)

			handler := NewGuestHandler(mockService)
			ctx := tt.setupContext()
			msg := tt.setupMessage()

			err := handler.HandleBulkDeleted(ctx, msg)

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
