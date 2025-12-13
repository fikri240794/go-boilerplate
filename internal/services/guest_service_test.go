package services

import (
	"context"
	"errors"
	"go-boilerplate/configs"
	"go-boilerplate/internal/models/dtos"
	"go-boilerplate/internal/models/entities"
	"go-boilerplate/internal/repositories"
	repo_mocks "go-boilerplate/internal/repositories/mocks"
	"net/http"
	"testing"
	"time"

	"github.com/fikri240794/gocerr"
	"github.com/fikri240794/goqube"
	"github.com/gofrs/uuid/v5"
	"github.com/guregu/null/v5"
	"github.com/stretchr/testify/mock"
)

func newTestGuestEntity(id, name, address, createdBy string, createdAt int64) *entities.GuestEntity {
	return &entities.GuestEntity{
		ID:        uuid.FromStringOrNil(id),
		Name:      name,
		Address:   null.StringFrom(address),
		CreatedAt: createdAt,
		CreatedBy: createdBy,
	}
}

func Test_NewGuestService(t *testing.T) {
	tests := []struct {
		name                         string
		cfg                          *configs.Config
		guestRepository              repositories.IGuestRepository
		guestCacheRepository         repositories.IGuestCacheRepository
		guestEventProducerRepository repositories.IGuestEventProducerRepository
		webhookSiteRepository        repositories.IWebhookSiteRepository
		expectNil                    bool
	}{
		{
			name: "create guest service with all dependencies",
			cfg: func() *configs.Config {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Keyf = "guest:%s"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest.created"
				return cfg
			}(),
			guestRepository:              repo_mocks.NewGuestRepositoryMock(t),
			guestCacheRepository:         repo_mocks.NewGuestCacheRepositoryMock(t),
			guestEventProducerRepository: repo_mocks.NewGuestEventProducerRepositoryMock(t),
			webhookSiteRepository:        repo_mocks.NewWebhookSiteRepositoryMock(t),
			expectNil:                    false,
		},
		{
			name:                         "create guest service without dependencies",
			cfg:                          nil,
			guestRepository:              nil,
			guestCacheRepository:         nil,
			guestEventProducerRepository: nil,
			webhookSiteRepository:        nil,
			expectNil:                    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGuestService(
				tt.cfg,
				tt.guestRepository,
				tt.guestCacheRepository,
				tt.guestEventProducerRepository,
				tt.webhookSiteRepository,
			)

			if tt.expectNil && service != nil {
				t.Error("NewGuestService() expected nil, got not nil")
			}

			if !tt.expectNil && service == nil {
				t.Error("NewGuestService() expected not nil, got nil")
			}

			if !tt.expectNil && service != nil {
				if service.cfg != tt.cfg {
					t.Error("NewGuestService() cfg not set correctly")
				}

				if service.guestRepository != tt.guestRepository {
					t.Error("NewGuestService() guestRepository not set correctly")
				}

				if service.guestCacheRepository != tt.guestCacheRepository {
					t.Error("NewGuestService() guestCacheRepository not set correctly")
				}

				if service.guestEventProducerRepository != tt.guestEventProducerRepository {
					t.Error("NewGuestService() guestEventProducerRepository not set correctly")
				}

				if service.webhookSiteRepository != tt.webhookSiteRepository {
					t.Error("NewGuestService() webhookSiteRepository not set correctly")
				}
			}
		})
	}
}

func Test_GuestService_deleteEntityCaches(t *testing.T) {
	tests := []struct {
		name          string
		setupService  func(t *testing.T) *GuestService
		expectError   bool
		validateError func(t *testing.T, err error)
	}{
		{
			name: "delete caches successfully",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Keys", mock.Anything, "guest:*").Return([]string{"guest:key1", "guest:key2"}, nil)
				mockCache.On("Delete", mock.Anything, mock.MatchedBy(func(keys []string) bool {
					return len(keys) == 2 && keys[0] == "guest:key1" && keys[1] == "guest:key2"
				})).Return(nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			expectError: false,
			validateError: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("deleteEntityCaches() unexpected error: %v", err)
				}
			},
		},
		{
			name: "delete caches with no keys found",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Keys", mock.Anything, "guest:*").Return([]string{}, nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			expectError: false,
			validateError: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("deleteEntityCaches() unexpected error: %v", err)
				}
			},
		},
		{
			name: "delete caches with Keys error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Keys", mock.Anything, "guest:*").Return([]string(nil), errors.New("redis keys error"))

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			expectError: true,
			validateError: func(t *testing.T, err error) {
				if err == nil {
					t.Error("deleteEntityCaches() expected error for Keys failure, got nil")
				}
			},
		},
		{
			name: "delete caches with Delete error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Keys", mock.Anything, "guest:*").Return([]string{"guest:key1"}, nil)
				mockCache.On("Delete", mock.Anything, mock.MatchedBy(func(keys []string) bool {
					return len(keys) == 1 && keys[0] == "guest:key1"
				})).Return(errors.New("redis delete error"))

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			expectError: true,
			validateError: func(t *testing.T, err error) {
				if err == nil {
					t.Error("deleteEntityCaches() expected error for Delete failure, got nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			err := service.deleteEntityCaches(ctx)

			if tt.validateError != nil {
				tt.validateError(t, err)
			}
		})
	}
}

func Test_GuestService_Create(t *testing.T) {
	tests := []struct {
		name         string
		setupService func(t *testing.T) *GuestService
		requestDTO   *dtos.CreateGuestRequestDTO
		expectError  bool
		validate     func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error)
	}{
		{
			name: "create successfully without cache and event",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = false
				cfg.Guest.Cache.Keyf = "guest:%s"
				cfg.Guest.Event.Created.Enable = false

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(nil)

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.GuestEntity")).Return(nil)

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Keys", mock.Anything, "guest:*").Return([]string{}, nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.CreateGuestRequestDTO{
				Name:    "John Doe",
				Address: "123 Main St",

				CreatedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err != nil {
					t.Errorf("Create() unexpected error: %v", err)
				}
				if responseDTO == nil {
					t.Error("Create() expected responseDTO, got nil")
				}
			},
		},
		{
			name: "create with nil requestDTO",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO:  nil,
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("Create() expected error for nil requestDTO, got nil")
				}
				if responseDTO != nil {
					t.Error("Create() expected nil responseDTO on error")
				}
			},
		},
		{
			name: "create with validation error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.CreateGuestRequestDTO{
				Name:      "",
				Address:   "",
				CreatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("Create() expected validation error, got nil")
				}
				if responseDTO != nil {
					t.Error("Create() expected nil responseDTO on validation error")
				}
			},
		},
		{
			name: "create with begin transaction error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(nil, errors.New("begin transaction error"))

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.CreateGuestRequestDTO{
				Name:    "John Doe",
				Address: "123 Main St",

				CreatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("Create() expected begin transaction error, got nil")
				}
				if responseDTO != nil {
					t.Error("Create() expected nil responseDTO on error")
				}
			},
		},
		{
			name: "create with repository create error and rollback success",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Rollback").Return(nil)

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.GuestEntity")).Return(errors.New("create entity error"))

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.CreateGuestRequestDTO{
				Name:    "John Doe",
				Address: "123 Main St",

				CreatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("Create() expected create error, got nil")
				}
				if err.Error() != "create entity error" {
					t.Errorf("Create() expected 'create entity error', got %v", err.Error())
				}
				if responseDTO != nil {
					t.Error("Create() expected nil responseDTO on error")
				}
			},
		},
		{
			name: "create with repository create error and rollback error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Rollback").Return(errors.New("rollback error"))

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.GuestEntity")).Return(errors.New("create entity error"))

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.CreateGuestRequestDTO{
				Name:    "John Doe",
				Address: "123 Main St",

				CreatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("Create() expected create error, got nil")
				}
				if responseDTO != nil {
					t.Error("Create() expected nil responseDTO on error")
				}
			},
		},
		{
			name: "create with commit error and rollback success",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(errors.New("commit error"))
				mockTx.On("Rollback").Return(nil)

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.GuestEntity")).Return(nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.CreateGuestRequestDTO{
				Name:    "John Doe",
				Address: "123 Main St",

				CreatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("Create() expected commit error, got nil")
				}
				if err.Error() != "commit error" {
					t.Errorf("Create() expected 'commit error', got %v", err.Error())
				}
				if responseDTO != nil {
					t.Error("Create() expected nil responseDTO on error")
				}
			},
		},
		{
			name: "create with commit error and rollback error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(errors.New("commit error"))
				mockTx.On("Rollback").Return(errors.New("rollback error"))

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.GuestEntity")).Return(nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.CreateGuestRequestDTO{
				Name:    "John Doe",
				Address: "123 Main St",

				CreatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("Create() expected commit error, got nil")
				}
				if responseDTO != nil {
					t.Error("Create() expected nil responseDTO on error")
				}
			},
		},
		{
			name: "create successfully with cache delete error (non-fatal)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Keyf = "guest:%s"
				cfg.Guest.Event.Created.Enable = false

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Keys", mock.Anything, "guest:*").Return([]string{"guest:key1"}, nil)
				mockCache.On("Delete", mock.Anything, mock.MatchedBy(func(keys []string) bool {
					return len(keys) == 1 && keys[0] == "guest:key1"
				})).Return(errors.New("cache delete error"))

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(nil)

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.GuestEntity")).Return(nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.CreateGuestRequestDTO{
				Name:    "John Doe",
				Address: "123 Main St",

				CreatedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err != nil {
					t.Errorf("Create() unexpected error: %v", err)
				}
				if responseDTO == nil {
					t.Error("Create() expected responseDTO, got nil")
				}
			},
		},
		{
			name: "create successfully with event publish success",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Keyf = "guest:%s"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest.created"

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(nil)

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.GuestEntity")).Return(nil)

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Keys", mock.Anything, "guest:*").Return([]string{}, nil)

				mockEventProducer := repo_mocks.NewGuestEventProducerRepositoryMock(t)
				mockEventProducer.On("Publish", mock.Anything, "guest.created", mock.AnythingOfType("*entities.EventEntity[go-boilerplate/internal/models/entities.GuestEventEntity]")).Return(nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					mockCache,
					mockEventProducer,
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.CreateGuestRequestDTO{
				Name:    "John Doe",
				Address: "123 Main St",

				CreatedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err != nil {
					t.Errorf("Create() unexpected error: %v", err)
				}
				if responseDTO == nil {
					t.Error("Create() expected responseDTO, got nil")
				}
			},
		},
		{
			name: "create successfully with event publish error (non-fatal)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Keyf = "guest:%s"
				cfg.Guest.Event.Created.Enable = true
				cfg.Guest.Event.Created.Topic = "guest.created"

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(nil)

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.GuestEntity")).Return(nil)

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Keys", mock.Anything, "guest:*").Return([]string{}, nil)

				mockEventProducer := repo_mocks.NewGuestEventProducerRepositoryMock(t)
				mockEventProducer.On("Publish", mock.Anything, "guest.created", mock.AnythingOfType("*entities.EventEntity[go-boilerplate/internal/models/entities.GuestEventEntity]")).Return(errors.New("event publish error"))

				return NewGuestService(
					cfg,
					mockGuestRepo,
					mockCache,
					mockEventProducer,
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.CreateGuestRequestDTO{
				Name:    "John Doe",
				Address: "123 Main St",

				CreatedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err != nil {
					t.Errorf("Create() unexpected error: %v", err)
				}
				if responseDTO == nil {
					t.Error("Create() expected responseDTO, got nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			responseDTO, err := service.Create(ctx, tt.requestDTO)

			if tt.expectError && err == nil {
				t.Error("Create() expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Create() unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, responseDTO, err)
			}
		})
	}
}

func Test_GuestService_DeleteByID(t *testing.T) {
	tests := []struct {
		name         string
		setupService func(t *testing.T) *GuestService
		requestDTO   *dtos.DeleteGuestByIDRequestDTO
		expectError  bool
		validate     func(t *testing.T, err error)
	}{
		{
			name: "delete successfully without cache and event",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Keyf = "guest:%s"

				entity := newTestGuestEntity(
					"019a9a5f-aaf4-7506-a942-6ed217773e2a",
					"John Doe",
					"123 Main St",
					"admin",
					1763526552308,
				)

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(nil)

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(entity, nil)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(nil)

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Keys", mock.Anything, "guest:*").Return([]string{}, nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "019a9a5f-aaf4-7506-a942-6ed217773e2a",
				DeletedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("DeleteByID() unexpected error: %v", err)
				}
			},
		},
		{
			name: "delete with nil requestDTO",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO:  nil,
			expectError: true,
			validate: func(t *testing.T, err error) {
				if err == nil {
					t.Error("DeleteByID() expected error for nil requestDTO, got nil")
				}
			},
		},
		{
			name: "delete with validation error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "",
				DeletedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				if err == nil {
					t.Error("DeleteByID() expected validation error, got nil")
				}
			},
		},
		{
			name: "delete with FindOne error (404)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(nil, gocerr.New(http.StatusNotFound, "guest not found"))

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "019a9a5f-aaf4-7506-a942-6ed217773e2a",
				DeletedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				if err == nil {
					t.Error("DeleteByID() expected FindOne error, got nil")
				}
			},
		},
		{
			name: "delete with FindOne error (500)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(nil, gocerr.New(http.StatusInternalServerError, "database error"))

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "019a9a5f-aaf4-7506-a942-6ed217773e2a",
				DeletedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				if err == nil {
					t.Error("DeleteByID() expected FindOne error, got nil")
				}
			},
		},
		{
			name: "delete with begin transaction error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				entity := newTestGuestEntity(
					"019a9a5f-aaf4-7506-a942-6ed217773e2a",
					"John Doe",
					"123 Main St",
					"admin",
					1763526552308,
				)

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(entity, nil)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(nil, errors.New("begin transaction error"))

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "019a9a5f-aaf4-7506-a942-6ed217773e2a",
				DeletedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				if err == nil {
					t.Error("DeleteByID() expected begin transaction error, got nil")
				}
			},
		},
		{
			name: "delete with repository update error and rollback success",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				entity := newTestGuestEntity(
					"019a9a5f-aaf4-7506-a942-6ed217773e2a",
					"John Doe",
					"123 Main St",
					"admin",
					1763526552308,
				)

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Rollback").Return(nil)

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(entity, nil)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(errors.New("update entity error"))

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "019a9a5f-aaf4-7506-a942-6ed217773e2a",
				DeletedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				if err == nil {
					t.Error("DeleteByID() expected update error, got nil")
				}
			},
		},
		{
			name: "delete with repository update error and rollback error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				entity := newTestGuestEntity(
					"019a9a5f-aaf4-7506-a942-6ed217773e2a",
					"John Doe",
					"123 Main St",
					"admin",
					1763526552308,
				)

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Rollback").Return(errors.New("rollback error"))

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(entity, nil)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(errors.New("update entity error"))

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "019a9a5f-aaf4-7506-a942-6ed217773e2a",
				DeletedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				if err == nil {
					t.Error("DeleteByID() expected update error, got nil")
				}
			},
		},
		{
			name: "delete with commit error and rollback success",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				entity := newTestGuestEntity(
					"019a9a5f-aaf4-7506-a942-6ed217773e2a",
					"John Doe",
					"123 Main St",
					"admin",
					1763526552308,
				)

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(errors.New("commit error"))
				mockTx.On("Rollback").Return(nil)

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(entity, nil)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "019a9a5f-aaf4-7506-a942-6ed217773e2a",
				DeletedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				if err == nil {
					t.Error("DeleteByID() expected commit error, got nil")
				}
			},
		},
		{
			name: "delete with commit error and rollback error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				entity := newTestGuestEntity(
					"019a9a5f-aaf4-7506-a942-6ed217773e2a",
					"John Doe",
					"123 Main St",
					"admin",
					1763526552308,
				)

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(errors.New("commit error"))
				mockTx.On("Rollback").Return(errors.New("rollback error"))

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(entity, nil)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "019a9a5f-aaf4-7506-a942-6ed217773e2a",
				DeletedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				if err == nil {
					t.Error("DeleteByID() expected commit error, got nil")
				}
			},
		},
		{
			name: "delete successfully with cache delete error (non-fatal)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Keyf = "guest:%s"

				entity := newTestGuestEntity(
					"019a9a5f-aaf4-7506-a942-6ed217773e2a",
					"John Doe",
					"123 Main St",
					"admin",
					1763526552308,
				)

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(nil)

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(entity, nil)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(nil)

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("Keys", mock.Anything, "guest:*").Return([]string{"guest:key1"}, nil)
				mockCacheRepo.On("Delete", mock.Anything, mock.MatchedBy(func(keys []string) bool {
					return len(keys) == 1 && keys[0] == "guest:key1"
				})).Return(errors.New("cache delete error"))

				return NewGuestService(
					cfg,
					mockGuestRepo,
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "019a9a5f-aaf4-7506-a942-6ed217773e2a",
				DeletedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("DeleteByID() unexpected error: %v", err)
				}
			},
		},
		{
			name: "delete successfully with event publish success",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Keyf = "guest:%s"
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest.deleted"

				entity := newTestGuestEntity(
					"019a9a5f-aaf4-7506-a942-6ed217773e2a",
					"John Doe",
					"123 Main St",
					"admin",
					1763526552308,
				)

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(nil)

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(entity, nil)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(nil)

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Keys", mock.Anything, "guest:*").Return([]string{}, nil)

				mockEventProducer := repo_mocks.NewGuestEventProducerRepositoryMock(t)
				mockEventProducer.On("Publish", mock.Anything, "guest.deleted", mock.AnythingOfType("*entities.EventEntity[go-boilerplate/internal/models/entities.GuestEventEntity]")).Return(nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					mockCache,
					mockEventProducer,
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "019a9a5f-aaf4-7506-a942-6ed217773e2a",
				DeletedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("DeleteByID() unexpected error: %v", err)
				}
			},
		},
		{
			name: "delete successfully with event publish error (non-fatal)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Keyf = "guest:%s"
				cfg.Guest.Event.Deleted.Enable = true
				cfg.Guest.Event.Deleted.Topic = "guest.deleted"

				entity := newTestGuestEntity(
					"019a9a5f-aaf4-7506-a942-6ed217773e2a",
					"John Doe",
					"123 Main St",
					"admin",
					1763526552308,
				)

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(nil)

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(entity, nil)
				mockGuestRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockGuestRepo.On("WithTransaction", mockTx).Return(mockGuestRepo)
				mockGuestRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(nil)

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Keys", mock.Anything, "guest:*").Return([]string{}, nil)

				mockEventRepo := repo_mocks.NewGuestEventProducerRepositoryMock(t)
				mockEventRepo.On("Publish", mock.Anything, "guest.deleted", mock.AnythingOfType("*entities.EventEntity[go-boilerplate/internal/models/entities.GuestEventEntity]")).Return(errors.New("event publish error"))

				return NewGuestService(
					cfg,
					mockGuestRepo,
					mockCache,
					mockEventRepo,
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.DeleteGuestByIDRequestDTO{
				ID:        "019a9a5f-aaf4-7506-a942-6ed217773e2a",
				DeletedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("DeleteByID() unexpected error: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			err := service.DeleteByID(ctx, tt.requestDTO)

			if tt.expectError && err == nil {
				t.Error("DeleteByID() expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("DeleteByID() unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}
		})
	}
}

func Test_GuestService_getListEntityCache(t *testing.T) {
	tests := []struct {
		name               string
		setupService       func(t *testing.T) *GuestService
		listEntityCacheKey string
		expectError        bool
		validate           func(t *testing.T, result []entities.GuestEntity, err error)
	}{
		{
			name: "get list entity cache successfully",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				testEntities := []entities.GuestEntity{
					*newTestGuestEntity(
						"019a9a5f-aaf4-7506-a942-6ed217773e2a",
						"John Doe",
						"123 Main St",
						"admin",
						1763526552308,
					),
					*newTestGuestEntity(
						"019a9a5f-aaf4-7506-a942-6ed217773e2b",
						"Jane Doe",
						"456 Oak Ave",
						"admin",
						1763526552309,
					),
				}

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("GetList", mock.Anything, "guest:test-key").Return(testEntities, nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			expectError:        false,
			validate: func(t *testing.T, result []entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("getListEntityCache() unexpected error: %v", err)
				}
				if result == nil {
					t.Error("getListEntityCache() expected result, got nil")
				}
				if len(result) != 2 {
					t.Errorf("getListEntityCache() expected 2 entities, got %d", len(result))
				}
			},
		},
		{
			name: "get list entity cache with error (not found)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("GetList", mock.Anything, "guest:test-key").Return([]entities.GuestEntity(nil), gocerr.New(http.StatusNotFound, "cache not found"))

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			expectError:        true,
			validate: func(t *testing.T, result []entities.GuestEntity, err error) {
				if err == nil {
					t.Error("getListEntityCache() expected error, got nil")
				}
				if result != nil {
					t.Error("getListEntityCache() expected nil result on error")
				}
			},
		},
		{
			name: "get list entity cache with error (500 - internal server error)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("GetList", mock.Anything, "guest:test-key").Return([]entities.GuestEntity(nil), gocerr.New(http.StatusInternalServerError, "cache server error"))

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			expectError:        true,
			validate: func(t *testing.T, result []entities.GuestEntity, err error) {
				if err == nil {
					t.Error("getListEntityCache() expected error, got nil")
				}
				if result != nil {
					t.Error("getListEntityCache() expected nil result on error")
				}
			},
		},
		{
			name: "get list entity cache returns empty list",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("GetList", mock.Anything, "guest:test-key").Return([]entities.GuestEntity{}, nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			expectError:        false,
			validate: func(t *testing.T, result []entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("getListEntityCache() unexpected error: %v", err)
				}
				if result == nil {
					t.Error("getListEntityCache() expected empty result, got nil")
				}
				if len(result) != 0 {
					t.Errorf("getListEntityCache() expected empty list, got %d items", len(result))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			result, err := service.getListEntityCache(ctx, tt.listEntityCacheKey)

			if tt.expectError && err == nil {
				t.Error("getListEntityCache() expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("getListEntityCache() unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, result, err)
			}
		})
	}
}

func Test_GuestService_setListEntityCache(t *testing.T) {
	tests := []struct {
		name               string
		setupService       func(t *testing.T) *GuestService
		listEntityCacheKey string
		listEntity         []entities.GuestEntity
		expectError        bool
		validate           func(t *testing.T, err error)
	}{
		{
			name: "set list entity cache successfully",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Duration = 300

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("SetList", mock.Anything, "guest:test-key", mock.AnythingOfType("[]entities.GuestEntity"), time.Duration(300)).Return(nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			listEntity: []entities.GuestEntity{
				*newTestGuestEntity(
					"019a9a5f-aaf4-7506-a942-6ed217773e2a",
					"John Doe",
					"123 Main St",
					"admin",
					1763526552308,
				),
				*newTestGuestEntity(
					"019a9a5f-aaf4-7506-a942-6ed217773e2b",
					"Jane Doe",
					"456 Oak Ave",
					"admin",
					1763526552309,
				),
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("setListEntityCache() unexpected error: %v", err)
				}
			},
		},
		{
			name: "set list entity cache with error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Duration = 300

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("SetList", mock.Anything, "guest:test-key", mock.AnythingOfType("[]entities.GuestEntity"), time.Duration(300)).Return(errors.New("cache set error"))

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			listEntity: []entities.GuestEntity{
				*newTestGuestEntity(
					"019a9a5f-aaf4-7506-a942-6ed217773e2a",
					"John Doe",
					"123 Main St",
					"admin",
					1763526552308,
				),
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				if err == nil {
					t.Error("setListEntityCache() expected error, got nil")
				}
				if err != nil && err.Error() != "cache set error" {
					t.Errorf("setListEntityCache() expected 'cache set error', got: %v", err)
				}
			},
		},
		{
			name: "set list entity cache with empty list",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Duration = 300

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("SetList", mock.Anything, "guest:test-key", mock.AnythingOfType("[]entities.GuestEntity"), time.Duration(300)).Return(nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			listEntity:         []entities.GuestEntity{},
			expectError:        false,
			validate: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("setListEntityCache() unexpected error: %v", err)
				}
			},
		},
		{
			name: "set list entity cache with nil list",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Duration = 300

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("SetList", mock.Anything, "guest:test-key", mock.AnythingOfType("[]entities.GuestEntity"), time.Duration(300)).Return(nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			listEntity:         nil,
			expectError:        false,
			validate: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("setListEntityCache() unexpected error: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			err := service.setListEntityCache(ctx, tt.listEntityCacheKey, tt.listEntity)

			if tt.expectError && err == nil {
				t.Error("setListEntityCache() expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("setListEntityCache() unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}
		})
	}
}

func Test_GuestService_findListEntity(t *testing.T) {
	testEntities := []entities.GuestEntity{
		*newTestGuestEntity(
			"019a9a5f-aaf4-7506-a942-6ed217773e2a",
			"John Doe",
			"123 Main St",
			"admin",
			1763526552308,
		),
		*newTestGuestEntity(
			"019a9a5f-aaf4-7506-a942-6ed217773e2b",
			"Jane Doe",
			"456 Oak Ave",
			"admin",
			1763526552309,
		),
	}

	tests := []struct {
		name               string
		setupService       func(t *testing.T) *GuestService
		listEntityCacheKey string
		filter             *goqube.Filter
		sorts              []goqube.Sort
		take               uint64
		skip               uint64
		expectError        bool
		validate           func(t *testing.T, result []entities.GuestEntity, err error)
	}{
		{
			name: "find list entity from cache successfully (cache enabled and cache hit)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("GetList", mock.Anything, "guest:test-key").Return(testEntities, nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			filter:             &goqube.Filter{},
			sorts:              []goqube.Sort{},
			take:               10,
			skip:               0,
			expectError:        false,
			validate: func(t *testing.T, result []entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("findListEntity() unexpected error: %v", err)
				}
				if len(result) != 2 {
					t.Errorf("findListEntity() expected 2 entities from cache, got %d", len(result))
				}
			},
		},
		{
			name: "find list entity from repository (cache disabled)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = false

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, uint64(10), uint64(0), false).Return(testEntities, nil)

				return NewGuestService(
					cfg,
					mockRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			filter:             &goqube.Filter{},
			sorts:              []goqube.Sort{},
			take:               10,
			skip:               0,
			expectError:        false,
			validate: func(t *testing.T, result []entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("findListEntity() unexpected error: %v", err)
				}
				if len(result) != 2 {
					t.Errorf("findListEntity() expected 2 entities from repository, got %d", len(result))
				}
			},
		},
		{
			name: "find list entity from repository (cache miss - 404)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Duration = 300

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("GetList", mock.Anything, "guest:test-key").Return([]entities.GuestEntity(nil), gocerr.New(http.StatusNotFound, "cache not found"))
				mockCacheRepo.On("SetList", mock.Anything, "guest:test-key", mock.AnythingOfType("[]entities.GuestEntity"), time.Duration(300)).Return(nil)

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, uint64(10), uint64(0), false).Return(testEntities, nil)

				return NewGuestService(
					cfg,
					mockRepo,
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			filter:             &goqube.Filter{},
			sorts:              []goqube.Sort{},
			take:               10,
			skip:               0,
			expectError:        false,
			validate: func(t *testing.T, result []entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("findListEntity() unexpected error: %v", err)
				}
				if len(result) != 2 {
					t.Errorf("findListEntity() expected 2 entities from repository after cache miss, got %d", len(result))
				}
			},
		},
		{
			name: "find list entity from repository with cache error (500 - non-fatal)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Duration = 300

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("GetList", mock.Anything, "guest:test-key").Return([]entities.GuestEntity(nil), gocerr.New(http.StatusInternalServerError, "cache server error"))
				mockCacheRepo.On("SetList", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, uint64(10), uint64(0), false).Return(testEntities, nil)

				return NewGuestService(
					cfg,
					mockRepo,
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			filter:             &goqube.Filter{},
			sorts:              []goqube.Sort{},
			take:               10,
			skip:               0,
			expectError:        false,
			validate: func(t *testing.T, result []entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("findListEntity() unexpected error: %v", err)
				}
				if len(result) != 2 {
					t.Errorf("findListEntity() expected 2 entities from repository after cache error, got %d", len(result))
				}
			},
		},
		{
			name: "find list entity from repository and set cache successfully",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Duration = 300

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("GetList", mock.Anything, "guest:test-key").Return([]entities.GuestEntity{}, nil)
				mockCacheRepo.On("SetList", mock.Anything, "guest:test-key", mock.AnythingOfType("[]entities.GuestEntity"), time.Duration(300)).Return(nil)

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, uint64(10), uint64(0), false).Return(testEntities, nil)

				return NewGuestService(
					cfg,
					mockRepo,
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			filter:             &goqube.Filter{},
			sorts:              []goqube.Sort{},
			take:               10,
			skip:               0,
			expectError:        false,
			validate: func(t *testing.T, result []entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("findListEntity() unexpected error: %v", err)
				}
				if len(result) != 2 {
					t.Errorf("findListEntity() expected 2 entities, got %d", len(result))
				}
			},
		},
		{
			name: "find list entity with repository error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = false

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, uint64(10), uint64(0), false).Return([]entities.GuestEntity(nil), gocerr.New(http.StatusInternalServerError, "database error"))

				return NewGuestService(
					cfg,
					mockRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			filter:             &goqube.Filter{},
			sorts:              []goqube.Sort{},
			take:               10,
			skip:               0,
			expectError:        true,
			validate: func(t *testing.T, result []entities.GuestEntity, err error) {
				if err == nil {
					t.Error("findListEntity() expected error from repository, got nil")
				}
				if result != nil {
					t.Error("findListEntity() expected nil result on error")
				}
			},
		},
		{
			name: "find list entity from repository with set cache error (non-fatal)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Duration = 300

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("GetList", mock.Anything, "guest:test-key").Return([]entities.GuestEntity{}, nil)
				mockCacheRepo.On("SetList", mock.Anything, "guest:test-key", mock.AnythingOfType("[]entities.GuestEntity"), time.Duration(300)).Return(errors.New("cache set error"))

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, uint64(10), uint64(0), false).Return(testEntities, nil)

				return NewGuestService(
					cfg,
					mockRepo,
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			filter:             &goqube.Filter{},
			sorts:              []goqube.Sort{},
			take:               10,
			skip:               0,
			expectError:        false,
			validate: func(t *testing.T, result []entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("findListEntity() unexpected error: %v (set cache error should be non-fatal)", err)
				}
				if len(result) != 2 {
					t.Errorf("findListEntity() expected 2 entities despite cache set error, got %d", len(result))
				}
			},
		},
		{
			name: "find list entity returns empty list from repository",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = false

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, uint64(10), uint64(0), false).Return([]entities.GuestEntity{}, nil)

				return NewGuestService(
					cfg,
					mockRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			listEntityCacheKey: "guest:test-key",
			filter:             &goqube.Filter{},
			sorts:              []goqube.Sort{},
			take:               10,
			skip:               0,
			expectError:        false,
			validate: func(t *testing.T, result []entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("findListEntity() unexpected error: %v", err)
				}
				if len(result) != 0 {
					t.Errorf("findListEntity() expected empty list, got %d items", len(result))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			result, err := service.findListEntity(
				ctx,
				tt.listEntityCacheKey,
				tt.filter,
				tt.sorts,
				tt.take,
				tt.skip,
			)

			if tt.expectError && err == nil {
				t.Error("findListEntity() expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("findListEntity() unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, result, err)
			}
		})
	}
}

func Test_GuestService_getCountEntitiesCache(t *testing.T) {
	tests := []struct {
		name                  string
		setupService          func(t *testing.T) *GuestService
		entitiesCountCacheKey string
		expectError           bool
		validate              func(t *testing.T, result uint64, err error)
	}{
		{
			name: "get count entities cache successfully",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("GetCount", mock.Anything, "guest:count-key").Return(uint64(42), nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			entitiesCountCacheKey: "guest:count-key",
			expectError:           false,
			validate: func(t *testing.T, result uint64, err error) {
				if err != nil {
					t.Errorf("getCountEntitiesCache() unexpected error: %v", err)
				}
				if result != 42 {
					t.Errorf("getCountEntitiesCache() expected count 42, got %d", result)
				}
			},
		},
		{
			name: "get count entities cache with error (not found)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("GetCount", mock.Anything, "guest:count-key").Return(uint64(0), gocerr.New(http.StatusNotFound, "cache not found"))

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			entitiesCountCacheKey: "guest:count-key",
			expectError:           true,
			validate: func(t *testing.T, result uint64, err error) {
				if err == nil {
					t.Error("getCountEntitiesCache() expected error, got nil")
				}
				if result != 0 {
					t.Errorf("getCountEntitiesCache() expected 0 on error, got %d", result)
				}
			},
		},
		{
			name: "get count entities cache with error (500 - internal server error)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("GetCount", mock.Anything, "guest:count-key").Return(uint64(0), gocerr.New(http.StatusInternalServerError, "cache server error"))

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			entitiesCountCacheKey: "guest:count-key",
			expectError:           true,
			validate: func(t *testing.T, result uint64, err error) {
				if err == nil {
					t.Error("getCountEntitiesCache() expected error, got nil")
				}
				if result != 0 {
					t.Errorf("getCountEntitiesCache() expected 0 on error, got %d", result)
				}
			},
		},
		{
			name: "get count entities cache returns zero",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("GetCount", mock.Anything, "guest:count-key").Return(uint64(0), nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			entitiesCountCacheKey: "guest:count-key",
			expectError:           false,
			validate: func(t *testing.T, result uint64, err error) {
				if err != nil {
					t.Errorf("getCountEntitiesCache() unexpected error: %v", err)
				}
				if result != 0 {
					t.Errorf("getCountEntitiesCache() expected 0, got %d", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			result, err := service.getCountEntitiesCache(ctx, tt.entitiesCountCacheKey)

			if tt.expectError && err == nil {
				t.Error("getCountEntitiesCache() expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("getCountEntitiesCache() unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, result, err)
			}
		})
	}
}

func Test_GuestService_setEntitiesCountCache(t *testing.T) {
	tests := []struct {
		name                  string
		setupService          func(t *testing.T) *GuestService
		entitiesCountCacheKey string
		entitiesCount         uint64
		expectError           bool
		validate              func(t *testing.T, err error)
	}{
		{
			name: "set entities count cache successfully",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Duration = 300

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("SetCount", mock.Anything, "guest:count-key", uint64(42), time.Duration(300)).Return(nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			entitiesCountCacheKey: "guest:count-key",
			entitiesCount:         42,
			expectError:           false,
			validate: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("setEntitiesCountCache() unexpected error: %v", err)
				}
			},
		},
		{
			name: "set entities count cache with error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Duration = 300

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("SetCount", mock.Anything, "guest:count-key", uint64(42), time.Duration(300)).Return(errors.New("cache set error"))

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			entitiesCountCacheKey: "guest:count-key",
			entitiesCount:         42,
			expectError:           true,
			validate: func(t *testing.T, err error) {
				if err == nil {
					t.Error("setEntitiesCountCache() expected error, got nil")
				}
				if err != nil && err.Error() != "cache set error" {
					t.Errorf("setEntitiesCountCache() expected 'cache set error', got: %v", err)
				}
			},
		},
		{
			name: "set entities count cache with zero value",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Duration = 300

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("SetCount", mock.Anything, "guest:count-key", uint64(0), time.Duration(300)).Return(nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			entitiesCountCacheKey: "guest:count-key",
			entitiesCount:         0,
			expectError:           false,
			validate: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("setEntitiesCountCache() unexpected error: %v", err)
				}
			},
		},
		{
			name: "set entities count cache with large value",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Duration = 300

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("SetCount", mock.Anything, "guest:count-key", uint64(999999), time.Duration(300)).Return(nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			entitiesCountCacheKey: "guest:count-key",
			entitiesCount:         999999,
			expectError:           false,
			validate: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("setEntitiesCountCache() unexpected error: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			err := service.setEntitiesCountCache(ctx, tt.entitiesCountCacheKey, tt.entitiesCount)

			if tt.expectError && err == nil {
				t.Error("setEntitiesCountCache() expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("setEntitiesCountCache() unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}
		})
	}
}

func Test_GuestService_countEntities(t *testing.T) {
	tests := []struct {
		name                  string
		setupService          func(t *testing.T) *GuestService
		entitiesCountCacheKey string
		filter                *goqube.Filter
		expectedCount         uint64
		expectError           bool
		validate              func(t *testing.T, count uint64, err error)
	}{
		{
			name: "count entities with cache disabled",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = false
				cfg.Guest.Cache.Duration = 300

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(5), nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			entitiesCountCacheKey: "test:count",
			filter: &goqube.Filter{
				Logic: goqube.LogicAnd,
				Filters: []goqube.Filter{
					{
						Field:    goqube.Field{Column: "deleted_at"},
						Operator: goqube.OperatorIsNull,
						Value:    goqube.FilterValue{Value: nil},
					},
				},
			},
			expectedCount: 5,
			expectError:   false,
			validate: func(t *testing.T, count uint64, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if count != 5 {
					t.Errorf("expected count 5, got %d", count)
				}
			},
		},
		{
			name: "count entities with cache hit - count greater than 0",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Duration = 300

				mockGuestCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockGuestCacheRepo.On("GetCount", mock.Anything, "test:count").Return(uint64(10), nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockGuestCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			entitiesCountCacheKey: "test:count",
			filter: &goqube.Filter{
				Logic: goqube.LogicAnd,
				Filters: []goqube.Filter{
					{
						Field:    goqube.Field{Column: "deleted_at"},
						Operator: goqube.OperatorIsNull,
						Value:    goqube.FilterValue{Value: nil},
					},
				},
			},
			expectedCount: 10,
			expectError:   false,
			validate: func(t *testing.T, count uint64, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if count != 10 {
					t.Errorf("expected count 10, got %d", count)
				}
			},
		},
		{
			name: "count entities with cache miss (404) - fetch from repository",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Duration = 300

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(7), nil)

				mockGuestCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockGuestCacheRepo.On("GetCount", mock.Anything, "test:count").Return(uint64(0), gocerr.New(http.StatusNotFound, "cache not found"))
				mockGuestCacheRepo.On("SetCount", mock.Anything, "test:count", uint64(7), time.Duration(300)).Return(nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					mockGuestCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			entitiesCountCacheKey: "test:count",
			filter: &goqube.Filter{
				Logic: goqube.LogicAnd,
				Filters: []goqube.Filter{
					{
						Field:    goqube.Field{Column: "deleted_at"},
						Operator: goqube.OperatorIsNull,
						Value:    goqube.FilterValue{Value: nil},
					},
				},
			},
			expectedCount: 7,
			expectError:   false,
			validate: func(t *testing.T, count uint64, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if count != 7 {
					t.Errorf("expected count 7, got %d", count)
				}
			},
		},
		{
			name: "count entities with cache error (500) - fetch from repository",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Duration = 300

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(3), nil)

				mockGuestCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockGuestCacheRepo.On("GetCount", mock.Anything, "test:count").Return(uint64(0), gocerr.New(http.StatusInternalServerError, "cache server error"))
				mockGuestCacheRepo.On("SetCount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				return NewGuestService(cfg, mockGuestRepo, mockGuestCacheRepo, repo_mocks.NewGuestEventProducerRepositoryMock(t), repo_mocks.NewWebhookSiteRepositoryMock(t))
			},
			entitiesCountCacheKey: "test:count",
			filter: &goqube.Filter{
				Logic: goqube.LogicAnd,
				Filters: []goqube.Filter{
					{
						Field:    goqube.Field{Column: "deleted_at"},
						Operator: goqube.OperatorIsNull,
						Value:    goqube.FilterValue{Value: nil},
					},
				},
			},
			expectedCount: 3,
			expectError:   false,
			validate: func(t *testing.T, count uint64, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if count != 3 {
					t.Errorf("expected count 3, got %d", count)
				}
			},
		},
		{
			name: "count entities with cache hit but count is 0 - fetch from repository",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Duration = 300

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(15), nil)

				mockGuestCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockGuestCacheRepo.On("GetCount", mock.Anything, "test:count").Return(uint64(0), nil)
				mockGuestCacheRepo.On("SetCount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				return NewGuestService(cfg, mockGuestRepo, mockGuestCacheRepo, repo_mocks.NewGuestEventProducerRepositoryMock(t), repo_mocks.NewWebhookSiteRepositoryMock(t))
			},
			entitiesCountCacheKey: "test:count",
			filter: &goqube.Filter{
				Logic: goqube.LogicAnd,
				Filters: []goqube.Filter{
					{
						Field:    goqube.Field{Column: "deleted_at"},
						Operator: goqube.OperatorIsNull,
						Value:    goqube.FilterValue{Value: nil},
					},
				},
			},
			expectedCount: 15,
			expectError:   false,
			validate: func(t *testing.T, count uint64, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if count != 15 {
					t.Errorf("expected count 15, got %d", count)
				}
			},
		},
		{
			name: "count entities with repository error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = false
				cfg.Guest.Cache.Duration = 300

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(0), gocerr.New(http.StatusInternalServerError, "database error"))

				mockGuestCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)

				return NewGuestService(cfg, mockGuestRepo, mockGuestCacheRepo, repo_mocks.NewGuestEventProducerRepositoryMock(t), repo_mocks.NewWebhookSiteRepositoryMock(t))
			},
			entitiesCountCacheKey: "test:count",
			filter: &goqube.Filter{
				Logic: goqube.LogicAnd,
				Filters: []goqube.Filter{
					{
						Field:    goqube.Field{Column: "deleted_at"},
						Operator: goqube.OperatorIsNull,
						Value:    goqube.FilterValue{Value: nil},
					},
				},
			},
			expectedCount: 0,
			expectError:   true,
			validate: func(t *testing.T, count uint64, err error) {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if count != 0 {
					t.Errorf("expected count 0, got %d", count)
				}
				if gocerr.GetErrorCode(err) != http.StatusInternalServerError {
					t.Errorf("expected status 500, got %d", gocerr.GetErrorCode(err))
				}
			},
		},
		{
			name: "count entities - set cache successfully after fetch",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Duration = 300

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(20), nil)

				mockGuestCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockGuestCacheRepo.On("GetCount", mock.Anything, "test:count").Return(uint64(0), gocerr.New(http.StatusNotFound, "cache not found"))
				mockGuestCacheRepo.On("SetCount", mock.Anything, "test:count", uint64(20), time.Duration(300)).Return(nil)

				return NewGuestService(cfg, mockGuestRepo, mockGuestCacheRepo, repo_mocks.NewGuestEventProducerRepositoryMock(t), repo_mocks.NewWebhookSiteRepositoryMock(t))
			},
			entitiesCountCacheKey: "test:count",
			filter: &goqube.Filter{
				Logic: goqube.LogicAnd,
				Filters: []goqube.Filter{
					{
						Field:    goqube.Field{Column: "deleted_at"},
						Operator: goqube.OperatorIsNull,
						Value:    goqube.FilterValue{Value: nil},
					},
				},
			},
			expectedCount: 20,
			expectError:   false,
			validate: func(t *testing.T, count uint64, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if count != 20 {
					t.Errorf("expected count 20, got %d", count)
				}
			},
		},
		{
			name: "count entities - set cache error after fetch (non-fatal)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Duration = 300

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(12), nil)

				mockGuestCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockGuestCacheRepo.On("GetCount", mock.Anything, "test:count").Return(uint64(0), gocerr.New(http.StatusNotFound, "cache not found"))
				mockGuestCacheRepo.On("SetCount", mock.Anything, "test:count", uint64(12), time.Duration(300)).Return(gocerr.New(http.StatusInternalServerError, "cache set error"))

				return NewGuestService(cfg, mockGuestRepo, mockGuestCacheRepo, repo_mocks.NewGuestEventProducerRepositoryMock(t), repo_mocks.NewWebhookSiteRepositoryMock(t))
			},
			entitiesCountCacheKey: "test:count",
			filter: &goqube.Filter{
				Logic: goqube.LogicAnd,
				Filters: []goqube.Filter{
					{
						Field:    goqube.Field{Column: "deleted_at"},
						Operator: goqube.OperatorIsNull,
						Value:    goqube.FilterValue{Value: nil},
					},
				},
			},
			expectedCount: 12,
			expectError:   false,
			validate: func(t *testing.T, count uint64, err error) {
				if err != nil {
					t.Errorf("expected no error (cache error is non-fatal), got %v", err)
				}
				if count != 12 {
					t.Errorf("expected count 12, got %d", count)
				}
			},
		},
		{
			name: "count entities - repository returns 0 count - should not set cache",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Duration = 300

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(0), nil)

				mockGuestCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockGuestCacheRepo.On("GetCount", mock.Anything, "test:count").Return(uint64(0), gocerr.New(http.StatusNotFound, "cache not found"))

				return NewGuestService(cfg, mockGuestRepo, mockGuestCacheRepo, repo_mocks.NewGuestEventProducerRepositoryMock(t), repo_mocks.NewWebhookSiteRepositoryMock(t))
			},
			entitiesCountCacheKey: "test:count",
			filter: &goqube.Filter{
				Logic: goqube.LogicAnd,
				Filters: []goqube.Filter{
					{
						Field:    goqube.Field{Column: "deleted_at"},
						Operator: goqube.OperatorIsNull,
						Value:    goqube.FilterValue{Value: nil},
					},
				},
			},
			expectedCount: 0,
			expectError:   false,
			validate: func(t *testing.T, count uint64, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if count != 0 {
					t.Errorf("expected count 0, got %d", count)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			count, err := service.countEntities(ctx, tt.entitiesCountCacheKey, tt.filter)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, count, err)
			}
		})
	}
}

func Test_GuestService_FindAll(t *testing.T) {
	testEntity1 := newTestGuestEntity(
		"00000000-0000-0000-0000-000000000001",
		"John Doe",
		"123 Main St",
		"admin",
		time.Now().Unix(),
	)
	testEntity2 := newTestGuestEntity(
		"00000000-0000-0000-0000-000000000002",
		"Jane Smith",
		"456 Oak Ave",
		"admin",
		time.Now().Unix(),
	)

	tests := []struct {
		name         string
		setupService func(t *testing.T) *GuestService
		requestDTO   *dtos.FindAllGuestRequestDTO
		expectError  bool
		validate     func(t *testing.T, responseDTO *dtos.FindAllGuestResponseDTO, err error)
	}{
		{
			name: "find all with nil requestDTO - should use default",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = false
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, false).Return([]entities.GuestEntity{*testEntity1, *testEntity2}, nil)
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(2), nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO:  nil,
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.FindAllGuestResponseDTO, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if responseDTO == nil {
					t.Error("expected responseDTO, got nil")
					return
				}
				if len(responseDTO.List) != 2 {
					t.Errorf("expected 2 guests, got %d", len(responseDTO.List))
				}
				if responseDTO.Count != 2 {
					t.Errorf("expected count 2, got %d", responseDTO.Count)
				}
			},
		},
		{
			name: "find all successfully with valid requestDTO",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = false
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, false).Return([]entities.GuestEntity{*testEntity1}, nil)
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(1), nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.FindAllGuestRequestDTO{
				Keyword: "John",
				Take:    10,
				Skip:    0,
			},
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.FindAllGuestResponseDTO, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if responseDTO == nil {
					t.Error("expected responseDTO, got nil")
					return
				}
				if len(responseDTO.List) != 1 {
					t.Errorf("expected 1 guest, got %d", len(responseDTO.List))
				}
				if responseDTO.Count != 1 {
					t.Errorf("expected count 1, got %d", responseDTO.Count)
				}
			},
		},
		{
			name: "find all with empty result",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = false
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, false).Return([]entities.GuestEntity{}, nil)
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(0), nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.FindAllGuestRequestDTO{
				Take: 10,
				Skip: 0,
			},
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.FindAllGuestResponseDTO, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if responseDTO == nil {
					t.Error("expected responseDTO, got nil")
					return
				}
				if len(responseDTO.List) != 0 {
					t.Errorf("expected 0 guests, got %d", len(responseDTO.List))
				}
				if responseDTO.Count != 0 {
					t.Errorf("expected count 0, got %d", responseDTO.Count)
				}
			},
		},
		{
			name: "find all with invalid requestDTO - ToFilterAndSorts error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = false
				cfg.Guest.Cache.Keyf = "guest:%s"

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.FindAllGuestRequestDTO{
				Sorts: "invalid|sort|format",
				Take:  10,
				Skip:  0,
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.FindAllGuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
			},
		},
		{
			name: "find all - findListEntity error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = false
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, false).Return([]entities.GuestEntity(nil), gocerr.New(http.StatusInternalServerError, "database error"))
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(0), nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.FindAllGuestRequestDTO{
				Take: 10,
				Skip: 0,
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.FindAllGuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
				if gocerr.GetErrorCode(err) != http.StatusInternalServerError {
					t.Errorf("expected status 500, got %d", gocerr.GetErrorCode(err))
				}
			},
		},
		{
			name: "find all - countEntities error (404)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = false
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, false).Return([]entities.GuestEntity{}, nil).Maybe()
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(0), gocerr.New(http.StatusNotFound, "count not found"))

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.FindAllGuestRequestDTO{
				Take: 10,
				Skip: 0,
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.FindAllGuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
				if gocerr.GetErrorCode(err) != http.StatusNotFound {
					t.Errorf("expected status 404, got %d", gocerr.GetErrorCode(err))
				}
			},
		},
		{
			name: "find all - countEntities error (500)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = false
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, false).Return([]entities.GuestEntity{}, nil).Maybe()
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(0), gocerr.New(http.StatusInternalServerError, "count database error"))

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.FindAllGuestRequestDTO{
				Take: 10,
				Skip: 0,
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.FindAllGuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
				if gocerr.GetErrorCode(err) != http.StatusInternalServerError {
					t.Errorf("expected status 500, got %d", gocerr.GetErrorCode(err))
				}
			},
		},
		{
			name: "find all with cache enabled - both from cache",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Keyf = "guest:%s"
				cfg.Guest.Cache.Duration = 300

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("GetList", mock.Anything, mock.Anything).Return([]entities.GuestEntity{*testEntity1, *testEntity2}, nil)
				mockCacheRepo.On("GetCount", mock.Anything, mock.Anything).Return(uint64(2), nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.FindAllGuestRequestDTO{
				Take: 10,
				Skip: 0,
			},
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.FindAllGuestResponseDTO, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if responseDTO == nil {
					t.Error("expected responseDTO, got nil")
					return
				}
				if len(responseDTO.List) != 2 {
					t.Errorf("expected 2 guests, got %d", len(responseDTO.List))
				}
				if responseDTO.Count != 2 {
					t.Errorf("expected count 2, got %d", responseDTO.Count)
				}
			},
		},
		{
			name: "find all with cache enabled - cache miss, fetch from repository",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Keyf = "guest:%s"
				cfg.Guest.Cache.Duration = 300

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, false).Return([]entities.GuestEntity{*testEntity1}, nil)
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(1), nil)

				mockCacheRepo := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCacheRepo.On("GetList", mock.Anything, mock.Anything).Return([]entities.GuestEntity(nil), gocerr.New(http.StatusNotFound, "cache not found"))
				mockCacheRepo.On("GetCount", mock.Anything, mock.Anything).Return(uint64(0), gocerr.New(http.StatusNotFound, "cache not found"))
				mockCacheRepo.On("SetList", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockCacheRepo.On("SetCount", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					mockCacheRepo,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.FindAllGuestRequestDTO{
				Take: 10,
				Skip: 0,
			},
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.FindAllGuestResponseDTO, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if responseDTO == nil {
					t.Error("expected responseDTO, got nil")
					return
				}
				if len(responseDTO.List) != 1 {
					t.Errorf("expected 1 guest, got %d", len(responseDTO.List))
				}
				if responseDTO.Count != 1 {
					t.Errorf("expected count 1, got %d", responseDTO.Count)
				}
			},
		},
		{
			name: "find all with pagination",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = false
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockGuestRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockGuestRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, false).Return([]entities.GuestEntity{*testEntity2}, nil)
				mockGuestRepo.On("Count", mock.Anything, mock.Anything, false).Return(uint64(2), nil)

				return NewGuestService(
					cfg,
					mockGuestRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.FindAllGuestRequestDTO{
				Take: 1,
				Skip: 1,
			},
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.FindAllGuestResponseDTO, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if responseDTO == nil {
					t.Error("expected responseDTO, got nil")
					return
				}
				if len(responseDTO.List) != 1 {
					t.Errorf("expected 1 guest, got %d", len(responseDTO.List))
				}
				if responseDTO.Count != 2 {
					t.Errorf("expected count 2, got %d", responseDTO.Count)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			responseDTO, err := service.FindAll(ctx, tt.requestDTO)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, responseDTO, err)
			}
		})
	}
}

func Test_GuestService_getEntityByIDCache(t *testing.T) {
	testEntity := newTestGuestEntity(
		"00000000-0000-0000-0000-000000000001",
		"John Doe",
		"123 Main St",
		"admin",
		time.Now().Unix(),
	)

	tests := []struct {
		name         string
		setupService func(t *testing.T) *GuestService
		cacheKey     string
		expectError  bool
		validate     func(t *testing.T, entity *entities.GuestEntity, err error)
	}{
		{
			name: "get entity by id cache successfully",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Get", mock.Anything, "guest:00000000-0000-0000-0000-000000000001").Return(testEntity, nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey:    "guest:00000000-0000-0000-0000-000000000001",
			expectError: false,
			validate: func(t *testing.T, entity *entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if entity == nil {
					t.Error("expected entity, got nil")
					return
				}
				if entity.ID.String() != testEntity.ID.String() {
					t.Errorf("expected entity ID %s, got %s", testEntity.ID.String(), entity.ID.String())
				}
				if entity.Name != testEntity.Name {
					t.Errorf("expected entity name %s, got %s", testEntity.Name, entity.Name)
				}
			},
		},
		{
			name: "get entity by id cache with 404 error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Get", mock.Anything, "guest:00000000-0000-0000-0000-000000000001").Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusNotFound, "cache not found"))

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey:    "guest:00000000-0000-0000-0000-000000000001",
			expectError: true,
			validate: func(t *testing.T, entity *entities.GuestEntity, err error) {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if entity != nil {
					t.Errorf("expected nil entity, got %v", entity)
				}
				if gocerr.GetErrorCode(err) != http.StatusNotFound {
					t.Errorf("expected status 404, got %d", gocerr.GetErrorCode(err))
				}
			},
		},
		{
			name: "get entity by id cache with 500 error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Get", mock.Anything, "guest:00000000-0000-0000-0000-000000000001").Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusInternalServerError, "cache server error"))

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey:    "guest:00000000-0000-0000-0000-000000000001",
			expectError: true,
			validate: func(t *testing.T, entity *entities.GuestEntity, err error) {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if entity != nil {
					t.Errorf("expected nil entity, got %v", entity)
				}
				if gocerr.GetErrorCode(err) != http.StatusInternalServerError {
					t.Errorf("expected status 500, got %d", gocerr.GetErrorCode(err))
				}
			},
		},
		{
			name: "get entity by id cache with nil entity result",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Get", mock.Anything, "guest:00000000-0000-0000-0000-000000000001").Return((*entities.GuestEntity)(nil), nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey:    "guest:00000000-0000-0000-0000-000000000001",
			expectError: false,
			validate: func(t *testing.T, entity *entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if entity != nil {
					t.Errorf("expected nil entity, got %v", entity)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			entity, err := service.getEntityByIDCache(ctx, tt.cacheKey)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, entity, err)
			}
		})
	}
}

func Test_GuestService_setEntityByIDCache(t *testing.T) {
	testEntity := newTestGuestEntity(
		"00000000-0000-0000-0000-000000000001",
		"John Doe",
		"123 Main St",
		"admin",
		time.Now().Unix(),
	)

	tests := []struct {
		name         string
		setupService func(t *testing.T) *GuestService
		cacheKey     string
		entity       *entities.GuestEntity
		expectError  bool
		validate     func(t *testing.T, err error)
	}{
		{
			name: "set entity by id cache successfully",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Duration = 300

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey:    "guest:00000000-0000-0000-0000-000000000001",
			entity:      testEntity,
			expectError: false,
			validate: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			},
		},
		{
			name: "set entity by id cache with error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Duration = 300

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(gocerr.New(http.StatusInternalServerError, "cache set error"))

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey:    "guest:00000000-0000-0000-0000-000000000001",
			entity:      testEntity,
			expectError: true,
			validate: func(t *testing.T, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}

				code := gocerr.GetErrorCode(err)
				if code != http.StatusInternalServerError {
					t.Errorf("expected error code %d, got %d", http.StatusInternalServerError, code)
				}
			},
		},
		{
			name: "set entity by id cache with nil entity",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Duration = 300

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey:    "guest:00000000-0000-0000-0000-000000000001",
			entity:      nil,
			expectError: false,
			validate: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			},
		},
		{
			name: "set entity by id cache with large entity data",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Duration = 300

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey: "guest:00000000-0000-0000-0000-000000000002",
			entity: newTestGuestEntity(
				"00000000-0000-0000-0000-000000000002",
				"Very Long Name Data",
				"Very Long Address Data",
				"admin",
				time.Now().Unix(),
			),
			expectError: false,
			validate: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			err := service.setEntityByIDCache(ctx, tt.cacheKey, tt.entity)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}
		})
	}
}

func Test_GuestService_findEntityByID(t *testing.T) {
	testEntity := newTestGuestEntity(
		"00000000-0000-0000-0000-000000000001",
		"John Doe",
		"123 Main St",
		"admin",
		time.Now().Unix(),
	)

	testFilter := &goqube.Filter{
		Logic: goqube.LogicAnd,
		Filters: []goqube.Filter{
			{
				Field:    goqube.Field{Column: entities.GuestEntityDatabaseFieldID},
				Operator: goqube.OperatorEqual,
				Value:    goqube.FilterValue{Value: "00000000-0000-0000-0000-000000000001"},
			},
			{
				Field:    goqube.Field{Column: entities.GuestEntityDatabaseFieldDeletedAt},
				Operator: goqube.OperatorIsNull,
				Value:    goqube.FilterValue{Value: nil},
			},
		},
	}

	tests := []struct {
		name         string
		setupService func(t *testing.T) *GuestService
		cacheKey     string
		filter       *goqube.Filter
		expectError  bool
		validate     func(t *testing.T, entity *entities.GuestEntity, err error)
	}{
		{
			name: "find entity by id with cache disabled",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = false

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(testEntity, nil)

				return NewGuestService(
					cfg,
					mockRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey:    "guest:00000000-0000-0000-0000-000000000001",
			filter:      testFilter,
			expectError: false,
			validate: func(t *testing.T, entity *entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if entity == nil {
					t.Error("expected entity, got nil")
					return
				}
				if entity.ID.String() != "00000000-0000-0000-0000-000000000001" {
					t.Errorf("expected entity ID %s, got %s", "00000000-0000-0000-0000-000000000001", entity.ID.String())
				}
			},
		},
		{
			name: "find entity by id with cache hit",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Get", mock.Anything, "guest:00000000-0000-0000-0000-000000000001").Return(testEntity, nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey:    "guest:00000000-0000-0000-0000-000000000001",
			filter:      testFilter,
			expectError: false,
			validate: func(t *testing.T, entity *entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if entity == nil {
					t.Error("expected entity, got nil")
					return
				}
				if entity.ID.String() != "00000000-0000-0000-0000-000000000001" {
					t.Errorf("expected entity ID %s, got %s", "00000000-0000-0000-0000-000000000001", entity.ID.String())
				}
			},
		},
		{
			name: "find entity by id with cache miss 404",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(testEntity, nil)

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Get", mock.Anything, "guest:00000000-0000-0000-0000-000000000001").Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusNotFound, "cache not found"))
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				return NewGuestService(
					cfg,
					mockRepo,
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey:    "guest:00000000-0000-0000-0000-000000000001",
			filter:      testFilter,
			expectError: false,
			validate: func(t *testing.T, entity *entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if entity == nil {
					t.Error("expected entity from repository, got nil")
					return
				}
				if entity.ID.String() != "00000000-0000-0000-0000-000000000001" {
					t.Errorf("expected entity ID %s, got %s", "00000000-0000-0000-0000-000000000001", entity.ID.String())
				}
			},
		},
		{
			name: "find entity by id with cache error 500",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(testEntity, nil)

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Get", mock.Anything, "guest:00000000-0000-0000-0000-000000000001").Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusInternalServerError, "cache server error"))
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				return NewGuestService(
					cfg,
					mockRepo,
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey:    "guest:00000000-0000-0000-0000-000000000001",
			filter:      testFilter,
			expectError: false,
			validate: func(t *testing.T, entity *entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if entity == nil {
					t.Error("expected entity from repository, got nil")
					return
				}
				if entity.ID.String() != "00000000-0000-0000-0000-000000000001" {
					t.Errorf("expected entity ID %s, got %s", "00000000-0000-0000-0000-000000000001", entity.ID.String())
				}
			},
		},
		{
			name: "find entity by id with repository error 404",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusNotFound, "entity not found"))

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Get", mock.Anything, "guest:00000000-0000-0000-0000-000000000001").Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusNotFound, "cache not found"))

				return NewGuestService(
					cfg,
					mockRepo,
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey:    "guest:00000000-0000-0000-0000-000000000001",
			filter:      testFilter,
			expectError: true,
			validate: func(t *testing.T, entity *entities.GuestEntity, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if entity != nil {
					t.Errorf("expected nil entity, got %v", entity)
				}
				code := gocerr.GetErrorCode(err)
				if code != http.StatusNotFound {
					t.Errorf("expected error code %d, got %d", http.StatusNotFound, code)
				}
			},
		},
		{
			name: "find entity by id with repository error 500",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusInternalServerError, "repository server error"))

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Get", mock.Anything, "guest:00000000-0000-0000-0000-000000000001").Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusNotFound, "cache not found"))

				return NewGuestService(
					cfg,
					mockRepo,
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey:    "guest:00000000-0000-0000-0000-000000000001",
			filter:      testFilter,
			expectError: true,
			validate: func(t *testing.T, entity *entities.GuestEntity, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if entity != nil {
					t.Errorf("expected nil entity, got %v", entity)
				}
				code := gocerr.GetErrorCode(err)
				if code != http.StatusInternalServerError {
					t.Errorf("expected error code %d, got %d", http.StatusInternalServerError, code)
				}
			},
		},
		{
			name: "find entity by id with cache miss and set cache success",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Duration = 300

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(testEntity, nil)

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Get", mock.Anything, "guest:00000000-0000-0000-0000-000000000001").Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusNotFound, "cache not found"))
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				return NewGuestService(
					cfg,
					mockRepo,
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey:    "guest:00000000-0000-0000-0000-000000000001",
			filter:      testFilter,
			expectError: false,
			validate: func(t *testing.T, entity *entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if entity == nil {
					t.Error("expected entity, got nil")
					return
				}
				if entity.ID.String() != "00000000-0000-0000-0000-000000000001" {
					t.Errorf("expected entity ID %s, got %s", "00000000-0000-0000-0000-000000000001", entity.ID.String())
				}
			},
		},
		{
			name: "find entity by id with cache miss and set cache error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Duration = 300

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(testEntity, nil)

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Get", mock.Anything, "guest:00000000-0000-0000-0000-000000000001").Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusNotFound, "cache not found"))
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(gocerr.New(http.StatusInternalServerError, "cache set error"))

				return NewGuestService(
					cfg,
					mockRepo,
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			cacheKey:    "guest:00000000-0000-0000-0000-000000000001",
			filter:      testFilter,
			expectError: false,
			validate: func(t *testing.T, entity *entities.GuestEntity, err error) {
				if err != nil {
					t.Errorf("expected no error (cache set error is non-fatal), got %v", err)
					return
				}
				if entity == nil {
					t.Error("expected entity, got nil")
					return
				}
				if entity.ID.String() != "00000000-0000-0000-0000-000000000001" {
					t.Errorf("expected entity ID %s, got %s", "00000000-0000-0000-0000-000000000001", entity.ID.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			entity, err := service.findEntityByID(ctx, tt.cacheKey, tt.filter)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, entity, err)
			}
		})
	}
}

func Test_GuestService_FindByID(t *testing.T) {
	testEntity := newTestGuestEntity(
		"00000000-0000-0000-0000-000000000001",
		"John Doe",
		"123 Main St",
		"admin",
		time.Now().Unix(),
	)

	tests := []struct {
		name         string
		setupService func(t *testing.T) *GuestService
		requestDTO   *dtos.FindGuestByIDRequestDTO
		expectError  bool
		validate     func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error)
	}{
		{
			name: "find by id with nil requestDTO",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockEventProducer := repo_mocks.NewGuestEventProducerRepositoryMock(t)
				mockWebhook := repo_mocks.NewWebhookSiteRepositoryMock(t)

				return NewGuestService(cfg, mockRepo, mockCache, mockEventProducer, mockWebhook)
			},
			requestDTO:  nil,
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
				code := gocerr.GetErrorCode(err)
				if code != http.StatusBadRequest {
					t.Errorf("expected error code %d, got %d", http.StatusBadRequest, code)
				}
			},
		},
		{
			name: "find by id with invalid requestDTO",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockEventProducer := repo_mocks.NewGuestEventProducerRepositoryMock(t)
				mockWebhook := repo_mocks.NewWebhookSiteRepositoryMock(t)

				return NewGuestService(cfg, mockRepo, mockCache, mockEventProducer, mockWebhook)
			},
			requestDTO: &dtos.FindGuestByIDRequestDTO{
				ID: "",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected validation error, got nil")
					return
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
			},
		},
		{
			name: "find by id successfully",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(testEntity, nil)

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Get", mock.Anything, mock.Anything).Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusNotFound, "cache not found"))
				mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				return NewGuestService(
					cfg,
					mockRepo,
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.FindGuestByIDRequestDTO{
				ID: "00000000-0000-0000-0000-000000000001",
			},
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if responseDTO == nil {
					t.Error("expected responseDTO, got nil")
					return
				}
				if responseDTO.ID != "00000000-0000-0000-0000-000000000001" {
					t.Errorf("expected ID %s, got %s", "00000000-0000-0000-0000-000000000001", responseDTO.ID)
				}
				if responseDTO.Name != "John Doe" {
					t.Errorf("expected Name %s, got %s", "John Doe", responseDTO.Name)
				}
			},
		},
		{
			name: "find by id with findEntityByID error 404",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusNotFound, "entity not found"))

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Get", mock.Anything, mock.Anything).Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusNotFound, "cache not found"))

				return NewGuestService(
					cfg,
					mockRepo,
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.FindGuestByIDRequestDTO{
				ID: "00000000-0000-0000-0000-000000000001",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
				code := gocerr.GetErrorCode(err)
				if code != http.StatusNotFound {
					t.Errorf("expected error code %d, got %d", http.StatusNotFound, code)
				}
			},
		},
		{
			name: "find by id with findEntityByID error 500",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Enable = true
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusInternalServerError, "repository server error"))

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Get", mock.Anything, mock.Anything).Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusNotFound, "cache not found"))

				return NewGuestService(
					cfg,
					mockRepo,
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.FindGuestByIDRequestDTO{
				ID: "00000000-0000-0000-0000-000000000001",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
				code := gocerr.GetErrorCode(err)
				if code != http.StatusInternalServerError {
					t.Errorf("expected error code %d, got %d", http.StatusInternalServerError, code)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			responseDTO, err := service.FindByID(ctx, tt.requestDTO)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, responseDTO, err)
			}
		})
	}
}

func Test_GuestService_UpdateByID(t *testing.T) {
	testEntity := newTestGuestEntity(
		"00000000-0000-0000-0000-000000000001",
		"John Doe",
		"123 Main St",
		"admin",
		time.Now().Unix(),
	)

	tests := []struct {
		name         string
		setupService func(t *testing.T) *GuestService
		requestDTO   *dtos.UpdateGuestByIDRequestDTO
		expectError  bool
		validate     func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error)
	}{
		{
			name: "update by id with nil requestDTO",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockEventProducer := repo_mocks.NewGuestEventProducerRepositoryMock(t)
				mockWebhook := repo_mocks.NewWebhookSiteRepositoryMock(t)

				return NewGuestService(cfg, mockRepo, mockCache, mockEventProducer, mockWebhook)
			},
			requestDTO:  nil,
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
				code := gocerr.GetErrorCode(err)
				if code != http.StatusBadRequest {
					t.Errorf("expected error code %d, got %d", http.StatusBadRequest, code)
				}
			},
		},
		{
			name: "update by id with invalid requestDTO",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockEventProducer := repo_mocks.NewGuestEventProducerRepositoryMock(t)
				mockWebhook := repo_mocks.NewWebhookSiteRepositoryMock(t)

				return NewGuestService(cfg, mockRepo, mockCache, mockEventProducer, mockWebhook)
			},
			requestDTO: &dtos.UpdateGuestByIDRequestDTO{
				ID: "",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected validation error, got nil")
					return
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
			},
		},
		{
			name: "update by id with FindOne error 404",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusNotFound, "entity not found"))

				return NewGuestService(
					cfg,
					mockRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "00000000-0000-0000-0000-000000000001",
				Name:      "Jane Doe",
				Address:   "456 New St",
				UpdatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
				code := gocerr.GetErrorCode(err)
				if code != http.StatusNotFound {
					t.Errorf("expected error code %d, got %d", http.StatusNotFound, code)
				}
			},
		},
		{
			name: "update by id with FindOne error 500",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return((*entities.GuestEntity)(nil), gocerr.New(http.StatusInternalServerError, "database error"))

				return NewGuestService(
					cfg,
					mockRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "00000000-0000-0000-0000-000000000001",
				Name:      "Jane Doe",
				Address:   "456 New St",
				UpdatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
				code := gocerr.GetErrorCode(err)
				if code != http.StatusInternalServerError {
					t.Errorf("expected error code %d, got %d", http.StatusInternalServerError, code)
				}
			},
		},
		{
			name: "update by id with BeginTransaction error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(testEntity, nil)
				mockRepo.On("BeginTransaction", mock.Anything).Return(nil, errors.New("transaction error"))

				return NewGuestService(
					cfg,
					mockRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "00000000-0000-0000-0000-000000000001",
				Name:      "Jane Doe",
				Address:   "456 New St",
				UpdatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
			},
		},
		{
			name: "update by id with Update error and rollback",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Rollback").Return(nil)

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(testEntity, nil)
				mockRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockRepo.On("WithTransaction", mockTx).Return(mockRepo)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(errors.New("update error"))

				return NewGuestService(
					cfg,
					mockRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "00000000-0000-0000-0000-000000000001",
				Name:      "Jane Doe",
				Address:   "456 New St",
				UpdatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
			},
		},
		{
			name: "update by id with Update error and rollback error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Rollback").Return(errors.New("rollback error"))

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(testEntity, nil)
				mockRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockRepo.On("WithTransaction", mockTx).Return(mockRepo)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(errors.New("update error"))

				return NewGuestService(
					cfg,
					mockRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "00000000-0000-0000-0000-000000000001",
				Name:      "Jane Doe",
				Address:   "456 New St",
				UpdatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
			},
		},
		{
			name: "update by id with Commit error and rollback",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(errors.New("commit error"))
				mockTx.On("Rollback").Return(nil)

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(testEntity, nil)
				mockRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockRepo.On("WithTransaction", mockTx).Return(mockRepo)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(nil)

				return NewGuestService(
					cfg,
					mockRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "00000000-0000-0000-0000-000000000001",
				Name:      "Jane Doe",
				Address:   "456 New St",
				UpdatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
			},
		},
		{
			name: "update by id with Commit error and rollback error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(errors.New("commit error"))
				mockTx.On("Rollback").Return(errors.New("rollback error"))

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(testEntity, nil)
				mockRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockRepo.On("WithTransaction", mockTx).Return(mockRepo)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(nil)

				return NewGuestService(
					cfg,
					mockRepo,
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "00000000-0000-0000-0000-000000000001",
				Name:      "Jane Doe",
				Address:   "456 New St",
				UpdatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
			},
		},
		{
			name: "update by id successfully without cache delete error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(nil)

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(testEntity, nil)
				mockRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockRepo.On("WithTransaction", mockTx).Return(mockRepo)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(nil)

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Keys", mock.Anything, "guest:*").Return([]string{}, nil)

				return NewGuestService(
					cfg,
					mockRepo,
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "00000000-0000-0000-0000-000000000001",
				Name:      "Jane Doe",
				Address:   "456 New St",
				UpdatedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if responseDTO == nil {
					t.Error("expected responseDTO, got nil")
					return
				}
				if responseDTO.ID != "00000000-0000-0000-0000-000000000001" {
					t.Errorf("expected ID %s, got %s", "00000000-0000-0000-0000-000000000001", responseDTO.ID)
				}
			},
		},
		{
			name: "update by id successfully with cache delete error (non-fatal)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Keyf = "guest:%s"

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(nil)

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(testEntity, nil)
				mockRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockRepo.On("WithTransaction", mockTx).Return(mockRepo)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(nil)

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Keys", mock.Anything, "guest:*").Return([]string{}, errors.New("cache keys error"))

				return NewGuestService(
					cfg,
					mockRepo,
					mockCache,
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "00000000-0000-0000-0000-000000000001",
				Name:      "Jane Doe",
				Address:   "456 New St",
				UpdatedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err != nil {
					t.Errorf("expected no error (cache delete is non-fatal), got %v", err)
					return
				}
				if responseDTO == nil {
					t.Error("expected responseDTO, got nil")
					return
				}
			},
		},
		{
			name: "update by id successfully with event enabled and publish success",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Keyf = "guest:%s"
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest.updated"

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(nil)

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(testEntity, nil)
				mockRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockRepo.On("WithTransaction", mockTx).Return(mockRepo)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(nil)

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Keys", mock.Anything, "guest:*").Return([]string{}, nil)

				mockEventProducer := repo_mocks.NewGuestEventProducerRepositoryMock(t)
				mockEventProducer.On("Publish", mock.Anything, "guest.updated", mock.AnythingOfType("*entities.EventEntity[go-boilerplate/internal/models/entities.GuestEventEntity]")).Return(nil)

				return NewGuestService(
					cfg,
					mockRepo,
					mockCache,
					mockEventProducer,
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "00000000-0000-0000-0000-000000000001",
				Name:      "Jane Doe",
				Address:   "456 New St",
				UpdatedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if responseDTO == nil {
					t.Error("expected responseDTO, got nil")
					return
				}
			},
		},
		{
			name: "update by id successfully with event enabled and publish error (non-fatal)",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}
				cfg.Guest.Cache.Keyf = "guest:%s"
				cfg.Guest.Event.Updated.Enable = true
				cfg.Guest.Event.Updated.Topic = "guest.updated"

				mockTx := repo_mocks.NewBoilerplateDatabaseTransactionMock(t)
				mockTx.On("Commit").Return(nil)

				mockRepo := repo_mocks.NewGuestRepositoryMock(t)
				mockRepo.On("FindOne", mock.Anything, mock.Anything, mock.Anything, false).Return(testEntity, nil)
				mockRepo.On("BeginTransaction", mock.Anything).Return(mockTx, nil)
				mockRepo.On("WithTransaction", mockTx).Return(mockRepo)
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.GuestEntity"), mock.Anything).Return(nil)

				mockCache := repo_mocks.NewGuestCacheRepositoryMock(t)
				mockCache.On("Keys", mock.Anything, "guest:*").Return([]string{}, nil)

				mockEventProducer := repo_mocks.NewGuestEventProducerRepositoryMock(t)
				mockEventProducer.On("Publish", mock.Anything, "guest.updated", mock.AnythingOfType("*entities.EventEntity[go-boilerplate/internal/models/entities.GuestEventEntity]")).Return(errors.New("publish error"))

				return NewGuestService(
					cfg,
					mockRepo,
					mockCache,
					mockEventProducer,
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO: &dtos.UpdateGuestByIDRequestDTO{
				ID:        "00000000-0000-0000-0000-000000000001",
				Name:      "Jane Doe",
				Address:   "456 New St",
				UpdatedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.GuestResponseDTO, err error) {
				if err != nil {
					t.Errorf("expected no error (event publish is non-fatal), got %v", err)
					return
				}
				if responseDTO == nil {
					t.Error("expected responseDTO, got nil")
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			responseDTO, err := service.UpdateByID(ctx, tt.requestDTO)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, responseDTO, err)
			}
		})
	}
}

func Test_GuestService_ProcessEvent(t *testing.T) {
	tests := []struct {
		name         string
		setupService func(t *testing.T) *GuestService
		requestDTO   *dtos.GuestEventRequestDTO
		expectError  bool
		validate     func(t *testing.T, responseDTO *dtos.GuestEventResponseDTO, err error)
	}{
		{
			name: "process event with nil requestDTO",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					repo_mocks.NewWebhookSiteRepositoryMock(t),
				)
			},
			requestDTO:  nil,
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestEventResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
				code := gocerr.GetErrorCode(err)
				if code != http.StatusBadRequest {
					t.Errorf("expected error code %d, got %d", http.StatusBadRequest, code)
				}
			},
		},
		{
			name: "process event with SendWebhook error",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockWebhook := repo_mocks.NewWebhookSiteRepositoryMock(t)
				mockWebhook.On("SendWebhook", mock.Anything, mock.AnythingOfType("*entities.GuestEventEntity")).Return(errors.New("webhook error"))

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					mockWebhook,
				)
			},
			requestDTO: &dtos.GuestEventRequestDTO{
				ID:        "00000000-0000-0000-0000-000000000001",
				Name:      "John Doe",
				Address:   "123 Main St",
				CreatedAt: 1700000000,
				CreatedBy: "admin",
				UpdatedAt: 1700000000,
				UpdatedBy: "admin",
			},
			expectError: true,
			validate: func(t *testing.T, responseDTO *dtos.GuestEventResponseDTO, err error) {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if responseDTO != nil {
					t.Errorf("expected nil responseDTO, got %v", responseDTO)
				}
			},
		},
		{
			name: "process event successfully",
			setupService: func(t *testing.T) *GuestService {
				cfg := &configs.Config{}

				mockWebhook := repo_mocks.NewWebhookSiteRepositoryMock(t)
				mockWebhook.On("SendWebhook", mock.Anything, mock.AnythingOfType("*entities.GuestEventEntity")).Return(nil)

				return NewGuestService(
					cfg,
					repo_mocks.NewGuestRepositoryMock(t),
					repo_mocks.NewGuestCacheRepositoryMock(t),
					repo_mocks.NewGuestEventProducerRepositoryMock(t),
					mockWebhook,
				)
			},
			requestDTO: &dtos.GuestEventRequestDTO{
				ID:        "00000000-0000-0000-0000-000000000001",
				Name:      "John Doe",
				Address:   "123 Main St",
				CreatedAt: 1700000000,
				CreatedBy: "admin",
				UpdatedAt: 1700000000,
				UpdatedBy: "admin",
			},
			expectError: false,
			validate: func(t *testing.T, responseDTO *dtos.GuestEventResponseDTO, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
					return
				}
				if responseDTO == nil {
					t.Error("expected responseDTO, got nil")
					return
				}
				if responseDTO.ID != "00000000-0000-0000-0000-000000000001" {
					t.Errorf("expected ID %s, got %s", "00000000-0000-0000-0000-000000000001", responseDTO.ID)
				}
				if responseDTO.Name != "John Doe" {
					t.Errorf("expected Name %s, got %s", "John Doe", responseDTO.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupService(t)
			ctx := context.Background()

			responseDTO, err := service.ProcessEvent(ctx, tt.requestDTO)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, responseDTO, err)
			}
		})
	}
}
