package handlers

import (
	"context"
	"go-boilerplate/internal/models/dtos"
	service_mocks "go-boilerplate/internal/services/mocks"
	"go-boilerplate/pkg/constants"
	"go-boilerplate/pkg/protobuf_boilerplate"
	"net/http"
	"testing"

	"github.com/fikri240794/gocerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestImplementedBoilerplateServer_CreateGuest(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func(t *testing.T) (*protobuf_boilerplate.CreateGuestRequestVM, context.Context)
		setupMock     func(t *testing.T, mockService *service_mocks.GuestServiceMock)
		validateError func(t *testing.T, err error)
		validate      func(t *testing.T, responseVM *protobuf_boilerplate.GuestResponseVM, err error)
	}{
		{
			name: "should_create_guest_successfully",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.CreateGuestRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				requestVM := &protobuf_boilerplate.CreateGuestRequestVM{
					Name:    "John Doe",
					Address: "123 Main St",
				}
				return requestVM, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("Create", mock.Anything, mock.AnythingOfType("*dtos.CreateGuestRequestDTO")).
					Return(&dtos.GuestResponseDTO{
						ID:        "550e8400-e29b-41d4-a716-446655440000",
						Name:      "John Doe",
						Address:   "123 Main St",
						CreatedAt: 1700000000000,
						CreatedBy: "00000000-0000-0000-0000-000000000000",
					}, nil)
			},
			validateError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.GuestResponseVM, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, responseVM)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", responseVM.Id)
				assert.Equal(t, "John Doe", responseVM.Name)
				assert.Equal(t, "123 Main St", responseVM.Address)
			},
		},
		{
			name: "should_return_error_when_request_vm_is_nil",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.CreateGuestRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				return nil, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {

			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "requestVM is nil")
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.GuestResponseVM, err error) {
				assert.Error(t, err)
				assert.Nil(t, responseVM)
			},
		},
		{
			name: "should_return_error_when_service_create_fails_with_4xx",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.CreateGuestRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				requestVM := &protobuf_boilerplate.CreateGuestRequestVM{
					Name:    "Jane Doe",
					Address: "456 Oak Ave",
				}
				return requestVM, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("Create", mock.Anything, mock.AnythingOfType("*dtos.CreateGuestRequestDTO")).
					Return(nil, gocerr.New(http.StatusBadRequest, "validation error"))
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.GuestResponseVM, err error) {
				assert.Error(t, err)
				assert.Nil(t, responseVM)
			},
		},
		{
			name: "should_return_error_when_service_create_fails_with_5xx",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.CreateGuestRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				requestVM := &protobuf_boilerplate.CreateGuestRequestVM{
					Name:    "Test User",
					Address: "Test Address",
				}
				return requestVM, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("Create", mock.Anything, mock.AnythingOfType("*dtos.CreateGuestRequestDTO")).
					Return(nil, gocerr.New(http.StatusInternalServerError, "internal server error"))
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.GuestResponseVM, err error) {
				assert.Error(t, err)
				assert.Nil(t, responseVM)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := service_mocks.NewGuestServiceMock(t)
			tt.setupMock(t, mockService)

			handler := NewImplementedBoilerplateServer(mockService)

			requestVM, ctx := tt.setupRequest(t)
			responseVM, err := handler.CreateGuest(ctx, requestVM)

			if tt.validateError != nil {
				tt.validateError(t, err)
			}
			if tt.validate != nil {
				tt.validate(t, responseVM, err)
			}
		})
	}
}

func TestImplementedBoilerplateServer_DeleteGuestByID(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func(t *testing.T) (*protobuf_boilerplate.DeleteGuestByIDRequestVM, context.Context)
		setupMock     func(t *testing.T, mockService *service_mocks.GuestServiceMock)
		validateError func(t *testing.T, err error)
		validate      func(t *testing.T, err error)
	}{
		{
			name: "should_delete_guest_successfully",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.DeleteGuestByIDRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				requestVM := &protobuf_boilerplate.DeleteGuestByIDRequestVM{
					Id: "550e8400-e29b-41d4-a716-446655440000",
				}
				return requestVM, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("DeleteByID", mock.Anything, mock.AnythingOfType("*dtos.DeleteGuestByIDRequestDTO")).
					Return(nil)
			},
			validateError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "should_return_error_when_request_vm_is_nil",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.DeleteGuestByIDRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				return nil, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {

			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "requestVM is nil")
			},
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "should_return_error_when_service_delete_fails_with_4xx",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.DeleteGuestByIDRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				requestVM := &protobuf_boilerplate.DeleteGuestByIDRequestVM{
					Id: "invalid-id",
				}
				return requestVM, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("DeleteByID", mock.Anything, mock.AnythingOfType("*dtos.DeleteGuestByIDRequestDTO")).
					Return(gocerr.New(http.StatusNotFound, "guest not found"))
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "should_return_error_when_service_delete_fails_with_5xx",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.DeleteGuestByIDRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				requestVM := &protobuf_boilerplate.DeleteGuestByIDRequestVM{
					Id: "550e8400-e29b-41d4-a716-446655440000",
				}
				return requestVM, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("DeleteByID", mock.Anything, mock.AnythingOfType("*dtos.DeleteGuestByIDRequestDTO")).
					Return(gocerr.New(http.StatusInternalServerError, "database error"))
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
			validate: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := service_mocks.NewGuestServiceMock(t)
			tt.setupMock(t, mockService)

			handler := NewImplementedBoilerplateServer(mockService)

			requestVM, ctx := tt.setupRequest(t)
			_, err := handler.DeleteGuestByID(ctx, requestVM)

			if tt.validateError != nil {
				tt.validateError(t, err)
			}
			if tt.validate != nil {
				tt.validate(t, err)
			}
		})
	}
}

func TestImplementedBoilerplateServer_FindAllGuest(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func(t *testing.T) (*protobuf_boilerplate.FindAllGuestRequestVM, context.Context)
		setupMock     func(t *testing.T, mockService *service_mocks.GuestServiceMock)
		validateError func(t *testing.T, err error)
		validate      func(t *testing.T, responseVM *protobuf_boilerplate.FindAllGuestResponseVM, err error)
	}{
		{
			name: "should_find_all_guests_successfully_with_request",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.FindAllGuestRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				requestVM := &protobuf_boilerplate.FindAllGuestRequestVM{
					Keyword: "john",
					Sorts:   "name.asc",
					Take:    10,
					Skip:    0,
				}
				return requestVM, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("FindAll", mock.Anything, mock.AnythingOfType("*dtos.FindAllGuestRequestDTO")).
					Return(&dtos.FindAllGuestResponseDTO{
						List: []dtos.GuestResponseDTO{
							{
								ID:        "550e8400-e29b-41d4-a716-446655440000",
								Name:      "John Doe",
								Address:   "123 Main St",
								CreatedAt: 1700000000000,
								CreatedBy: "user1",
							},
						},
						Count: 1,
					}, nil)
			},
			validateError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.FindAllGuestResponseVM, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, responseVM)
				assert.Equal(t, uint64(1), responseVM.Count)
				assert.Equal(t, 1, len(responseVM.List))
				assert.Equal(t, "John Doe", responseVM.List[0].Name)
			},
		},
		{
			name: "should_find_all_guests_with_nil_request_vm",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.FindAllGuestRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				return nil, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("FindAll", mock.Anything, mock.AnythingOfType("*dtos.FindAllGuestRequestDTO")).
					Return(&dtos.FindAllGuestResponseDTO{
						List:  []dtos.GuestResponseDTO{},
						Count: 0,
					}, nil)
			},
			validateError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.FindAllGuestResponseVM, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, responseVM)
				assert.Equal(t, uint64(0), responseVM.Count)
			},
		},
		{
			name: "should_return_error_when_service_findall_fails_with_4xx",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.FindAllGuestRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				requestVM := &protobuf_boilerplate.FindAllGuestRequestVM{
					Keyword: "test",
				}
				return requestVM, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("FindAll", mock.Anything, mock.AnythingOfType("*dtos.FindAllGuestRequestDTO")).
					Return(nil, gocerr.New(http.StatusBadRequest, "invalid query"))
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.FindAllGuestResponseVM, err error) {
				assert.Error(t, err)
				assert.Nil(t, responseVM)
			},
		},
		{
			name: "should_return_error_when_service_findall_fails_with_5xx",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.FindAllGuestRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				requestVM := &protobuf_boilerplate.FindAllGuestRequestVM{}
				return requestVM, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("FindAll", mock.Anything, mock.AnythingOfType("*dtos.FindAllGuestRequestDTO")).
					Return(nil, gocerr.New(http.StatusInternalServerError, "database connection error"))
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.FindAllGuestResponseVM, err error) {
				assert.Error(t, err)
				assert.Nil(t, responseVM)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := service_mocks.NewGuestServiceMock(t)
			tt.setupMock(t, mockService)

			handler := NewImplementedBoilerplateServer(mockService)

			requestVM, ctx := tt.setupRequest(t)
			responseVM, err := handler.FindAllGuest(ctx, requestVM)

			if tt.validateError != nil {
				tt.validateError(t, err)
			}
			if tt.validate != nil {
				tt.validate(t, responseVM, err)
			}
		})
	}
}

func TestImplementedBoilerplateServer_FindGuestByID(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func(t *testing.T) (*protobuf_boilerplate.FindGuestByIDRequestVM, context.Context)
		setupMock     func(t *testing.T, mockService *service_mocks.GuestServiceMock)
		validateError func(t *testing.T, err error)
		validate      func(t *testing.T, responseVM *protobuf_boilerplate.GuestResponseVM, err error)
	}{
		{
			name: "should_find_guest_by_id_successfully",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.FindGuestByIDRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				requestVM := &protobuf_boilerplate.FindGuestByIDRequestVM{
					Id: "550e8400-e29b-41d4-a716-446655440000",
				}
				return requestVM, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("FindByID", mock.Anything, mock.AnythingOfType("*dtos.FindGuestByIDRequestDTO")).
					Return(&dtos.GuestResponseDTO{
						ID:        "550e8400-e29b-41d4-a716-446655440000",
						Name:      "John Doe",
						Address:   "123 Main St",
						CreatedAt: 1700000000000,
						CreatedBy: "user1",
					}, nil)
			},
			validateError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.GuestResponseVM, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, responseVM)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", responseVM.Id)
				assert.Equal(t, "John Doe", responseVM.Name)
			},
		},
		{
			name: "should_return_error_when_request_vm_is_nil",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.FindGuestByIDRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				return nil, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {

			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "requestVM is nil")
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.GuestResponseVM, err error) {
				assert.Error(t, err)
				assert.Nil(t, responseVM)
			},
		},
		{
			name: "should_return_error_when_service_findbyid_fails_with_4xx",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.FindGuestByIDRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				requestVM := &protobuf_boilerplate.FindGuestByIDRequestVM{
					Id: "invalid-id",
				}
				return requestVM, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("FindByID", mock.Anything, mock.AnythingOfType("*dtos.FindGuestByIDRequestDTO")).
					Return(nil, gocerr.New(http.StatusNotFound, "guest not found"))
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.GuestResponseVM, err error) {
				assert.Error(t, err)
				assert.Nil(t, responseVM)
			},
		},
		{
			name: "should_return_error_when_service_findbyid_fails_with_5xx",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.FindGuestByIDRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				requestVM := &protobuf_boilerplate.FindGuestByIDRequestVM{
					Id: "550e8400-e29b-41d4-a716-446655440000",
				}
				return requestVM, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("FindByID", mock.Anything, mock.AnythingOfType("*dtos.FindGuestByIDRequestDTO")).
					Return(nil, gocerr.New(http.StatusInternalServerError, "database error"))
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.GuestResponseVM, err error) {
				assert.Error(t, err)
				assert.Nil(t, responseVM)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := service_mocks.NewGuestServiceMock(t)
			tt.setupMock(t, mockService)

			handler := NewImplementedBoilerplateServer(mockService)

			requestVM, ctx := tt.setupRequest(t)
			responseVM, err := handler.FindGuestByID(ctx, requestVM)

			if tt.validateError != nil {
				tt.validateError(t, err)
			}
			if tt.validate != nil {
				tt.validate(t, responseVM, err)
			}
		})
	}
}

func TestImplementedBoilerplateServer_UpdateGuestByID(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func(t *testing.T) (*protobuf_boilerplate.UpdateGuestByIDRequestVM, context.Context)
		setupMock     func(t *testing.T, mockService *service_mocks.GuestServiceMock)
		validateError func(t *testing.T, err error)
		validate      func(t *testing.T, responseVM *protobuf_boilerplate.GuestResponseVM, err error)
	}{
		{
			name: "should_update_guest_successfully",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.UpdateGuestByIDRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				requestVM := &protobuf_boilerplate.UpdateGuestByIDRequestVM{
					Id:      "550e8400-e29b-41d4-a716-446655440000",
					Name:    "Updated Name",
					Address: "Updated Address",
				}
				return requestVM, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("UpdateByID", mock.Anything, mock.AnythingOfType("*dtos.UpdateGuestByIDRequestDTO")).
					Return(&dtos.GuestResponseDTO{
						ID:        "550e8400-e29b-41d4-a716-446655440000",
						Name:      "Updated Name",
						Address:   "Updated Address",
						CreatedAt: 1700000000000,
						CreatedBy: "user1",
						UpdatedAt: 1700000001000,
						UpdatedBy: "00000000-0000-0000-0000-000000000000",
					}, nil)
			},
			validateError: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.GuestResponseVM, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, responseVM)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", responseVM.Id)
				assert.Equal(t, "Updated Name", responseVM.Name)
				assert.Equal(t, "Updated Address", responseVM.Address)
			},
		},
		{
			name: "should_return_error_when_request_vm_is_nil",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.UpdateGuestByIDRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				return nil, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {

			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "requestVM is nil")
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.GuestResponseVM, err error) {
				assert.Error(t, err)
				assert.Nil(t, responseVM)
			},
		},
		{
			name: "should_return_error_when_service_update_fails_with_4xx",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.UpdateGuestByIDRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				requestVM := &protobuf_boilerplate.UpdateGuestByIDRequestVM{
					Id:      "invalid-id",
					Name:    "Test",
					Address: "Test",
				}
				return requestVM, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("UpdateByID", mock.Anything, mock.AnythingOfType("*dtos.UpdateGuestByIDRequestDTO")).
					Return(nil, gocerr.New(http.StatusNotFound, "guest not found"))
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.GuestResponseVM, err error) {
				assert.Error(t, err)
				assert.Nil(t, responseVM)
			},
		},
		{
			name: "should_return_error_when_service_update_fails_with_5xx",
			setupRequest: func(t *testing.T) (*protobuf_boilerplate.UpdateGuestByIDRequestVM, context.Context) {
				ctx := context.WithValue(context.Background(), constants.ContextKeyRequestID, "test-request-id")
				requestVM := &protobuf_boilerplate.UpdateGuestByIDRequestVM{
					Id:      "550e8400-e29b-41d4-a716-446655440000",
					Name:    "Test",
					Address: "Test",
				}
				return requestVM, ctx
			},
			setupMock: func(t *testing.T, mockService *service_mocks.GuestServiceMock) {
				mockService.On("UpdateByID", mock.Anything, mock.AnythingOfType("*dtos.UpdateGuestByIDRequestDTO")).
					Return(nil, gocerr.New(http.StatusInternalServerError, "database error"))
			},
			validateError: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
			validate: func(t *testing.T, responseVM *protobuf_boilerplate.GuestResponseVM, err error) {
				assert.Error(t, err)
				assert.Nil(t, responseVM)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := service_mocks.NewGuestServiceMock(t)
			tt.setupMock(t, mockService)

			handler := NewImplementedBoilerplateServer(mockService)

			requestVM, ctx := tt.setupRequest(t)
			responseVM, err := handler.UpdateGuestByID(ctx, requestVM)

			if tt.validateError != nil {
				tt.validateError(t, err)
			}
			if tt.validate != nil {
				tt.validate(t, responseVM, err)
			}
		})
	}
}
