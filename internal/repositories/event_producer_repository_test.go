package repositories

import (
	"context"
	"errors"
	"go-boilerplate/datasources/event_producer"
	event_producer_mocks "go-boilerplate/datasources/event_producer/mocks"
	"go-boilerplate/internal/models/entities"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testEntity struct {
	ID   int
	Name string
}

type testEntityWithChannel struct {
	ID      int
	Name    string
	Channel chan int
}

func Test_NewEventProducerRepository(t *testing.T) {
	tests := []struct {
		name          string
		eventProducer *event_producer.EventProducer
		validate      func(t *testing.T, repo *EventProducerRepository[testEntity])
	}{
		{
			name: "create new event producer repository successfully",
			eventProducer: &event_producer.EventProducer{
				NSQProducer: event_producer_mocks.NewNSQProducerMock(t),
			},
			validate: func(t *testing.T, repo *EventProducerRepository[testEntity]) {
				assert.NotNil(t, repo, "expected repository, got nil")
				assert.NotNil(t, repo.eventProducer, "expected eventProducer to be set, got nil")
			},
		},
		{
			name:          "create new event producer repository with nil event producer",
			eventProducer: nil,
			validate: func(t *testing.T, repo *EventProducerRepository[testEntity]) {
				assert.NotNil(t, repo, "expected repository, got nil")
				assert.Nil(t, repo.eventProducer, "expected eventProducer to be nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			repo := NewEventProducerRepository[testEntity](tt.eventProducer)

			if tt.validate != nil {
				tt.validate(t, repo)
			}
		})
	}
}

func Test_EventProducerRepository_Publish(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func() *EventProducerRepository[testEntity]
		topic       string
		message     *entities.EventEntity[testEntity]
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "publish successfully",
			setupRepo: func() *EventProducerRepository[testEntity] {
				mockProducer := event_producer_mocks.NewNSQProducerMock(t)
				mockProducer.On("Publish", "test-topic", mock.AnythingOfType("[]uint8")).Return(nil)
				return &EventProducerRepository[testEntity]{
					eventProducer: &event_producer.EventProducer{
						NSQProducer: mockProducer,
					},
				}
			},
			topic: "test-topic",
			message: entities.NewEventEntity("test-event", &testEntity{
				ID:   1,
				Name: "Test Entity",
			}),
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, "unexpected error: %v", err)
			},
		},
		{
			name: "publish with NSQProducer.Publish error",
			setupRepo: func() *EventProducerRepository[testEntity] {
				mockProducer := event_producer_mocks.NewNSQProducerMock(t)
				mockProducer.On("Publish", "test-topic", mock.AnythingOfType("[]uint8")).Return(errors.New("nsq publish error"))
				return &EventProducerRepository[testEntity]{
					eventProducer: &event_producer.EventProducer{
						NSQProducer: mockProducer,
					},
				}
			},
			topic: "test-topic",
			message: entities.NewEventEntity("test-event", &testEntity{
				ID:   2,
				Name: "Test Entity 2",
			}),
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "expected error from NSQProducer.Publish, got nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			repo := tt.setupRepo()

			ctx := context.Background()
			err := repo.Publish(ctx, tt.topic, tt.message)

			if tt.expectError {
				assert.NotNil(t, err, "expected error, got nil")
			} else {
				assert.NoError(t, err, "unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}
		})
	}

	t.Run("publish with json.Marshal error", func(t *testing.T) {
		mockProducer := event_producer_mocks.NewNSQProducerMock(t)
		repo := &EventProducerRepository[testEntityWithChannel]{
			eventProducer: &event_producer.EventProducer{
				NSQProducer: mockProducer,
			},
		}

		message := entities.NewEventEntity("test-event", &testEntityWithChannel{
			ID:      1,
			Name:    "Test",
			Channel: make(chan int),
		})

		ctx := context.Background()
		err := repo.Publish(ctx, "test-topic", message)

		assert.NotNil(t, err, "expected error from json.Marshal, got nil")
	})
}

func Test_EventProducerRepository_PublishWithDelay(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func() *EventProducerRepository[testEntity]
		topic       string
		delay       time.Duration
		message     *entities.EventEntity[testEntity]
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "publish with delay successfully",
			setupRepo: func() *EventProducerRepository[testEntity] {
				mockProducer := event_producer_mocks.NewNSQProducerMock(t)
				mockProducer.On("DeferredPublish", "test-topic", 5*time.Second, mock.AnythingOfType("[]uint8")).Return(nil)
				return &EventProducerRepository[testEntity]{
					eventProducer: &event_producer.EventProducer{
						NSQProducer: mockProducer,
					},
				}
			},
			topic: "test-topic",
			delay: 5 * time.Second,
			message: entities.NewEventEntity("test-event", &testEntity{
				ID:   1,
				Name: "Test Entity",
			}),
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, "unexpected error: %v", err)
			},
		},
		{
			name: "publish with delay with NSQProducer.DeferredPublish error",
			setupRepo: func() *EventProducerRepository[testEntity] {
				mockProducer := event_producer_mocks.NewNSQProducerMock(t)
				mockProducer.On("DeferredPublish", "test-topic", 10*time.Second, mock.AnythingOfType("[]uint8")).Return(errors.New("nsq deferred publish error"))
				return &EventProducerRepository[testEntity]{
					eventProducer: &event_producer.EventProducer{
						NSQProducer: mockProducer,
					},
				}
			},
			topic: "test-topic",
			delay: 10 * time.Second,
			message: entities.NewEventEntity("test-event", &testEntity{
				ID:   2,
				Name: "Test Entity 2",
			}),
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "expected error from NSQProducer.DeferredPublish, got nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			repo := tt.setupRepo()

			ctx := context.Background()
			err := repo.PublishWithDelay(ctx, tt.topic, tt.delay, tt.message)

			if tt.expectError {
				assert.NotNil(t, err, "expected error, got nil")
			} else {
				assert.NoError(t, err, "unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}
		})
	}

	t.Run("publish with delay with json.Marshal error", func(t *testing.T) {
		mockProducer := event_producer_mocks.NewNSQProducerMock(t)
		repo := &EventProducerRepository[testEntityWithChannel]{
			eventProducer: &event_producer.EventProducer{
				NSQProducer: mockProducer,
			},
		}

		message := entities.NewEventEntity("test-event", &testEntityWithChannel{
			ID:      1,
			Name:    "Test",
			Channel: make(chan int),
		})

		ctx := context.Background()
		err := repo.PublishWithDelay(ctx, "test-topic", 5*time.Second, message)

		assert.NotNil(t, err, "expected error from json.Marshal, got nil")
	})
}

func Test_EventProducerRepository_PublishBulk(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func() *EventProducerRepository[testEntity]
		topic       string
		message     *entities.EventEntity[[]testEntity]
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "publish bulk successfully",
			setupRepo: func() *EventProducerRepository[testEntity] {
				mockProducer := event_producer_mocks.NewNSQProducerMock(t)
				mockProducer.On("Publish", "test-topic", mock.AnythingOfType("[]uint8")).Return(nil)
				return &EventProducerRepository[testEntity]{
					eventProducer: &event_producer.EventProducer{
						NSQProducer: mockProducer,
					},
				}
			},
			topic: "test-topic",
			message: entities.NewEventEntity("test-event", &[]testEntity{
				{ID: 1, Name: "Entity 1"},
				{ID: 2, Name: "Entity 2"},
			}),
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, "unexpected error: %v", err)
			},
		},
		{
			name: "publish bulk with NSQProducer.Publish error",
			setupRepo: func() *EventProducerRepository[testEntity] {
				mockProducer := event_producer_mocks.NewNSQProducerMock(t)
				mockProducer.On("Publish", "test-topic", mock.AnythingOfType("[]uint8")).Return(errors.New("nsq publish error"))
				return &EventProducerRepository[testEntity]{
					eventProducer: &event_producer.EventProducer{
						NSQProducer: mockProducer,
					},
				}
			},
			topic: "test-topic",
			message: entities.NewEventEntity("test-event", &[]testEntity{
				{ID: 3, Name: "Entity 3"},
			}),
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "expected error from NSQProducer.Publish, got nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			ctx := context.Background()
			err := repo.PublishBulk(ctx, tt.topic, tt.message)

			if tt.expectError {
				assert.NotNil(t, err, "expected error, got nil")
			} else {
				assert.NoError(t, err, "unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}
		})
	}

	t.Run("publish bulk with json.Marshal error", func(t *testing.T) {
		mockProducer := event_producer_mocks.NewNSQProducerMock(t)
		repo := &EventProducerRepository[testEntityWithChannel]{
			eventProducer: &event_producer.EventProducer{
				NSQProducer: mockProducer,
			},
		}

		message := entities.NewEventEntity("test-event", &[]testEntityWithChannel{
			{ID: 1, Name: "Test", Channel: make(chan int)},
		})

		ctx := context.Background()
		err := repo.PublishBulk(ctx, "test-topic", message)

		assert.NotNil(t, err, "expected error from json.Marshal, got nil")
	})
}

func Test_EventProducerRepository_PublishBulkWithDelay(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func() *EventProducerRepository[testEntity]
		topic       string
		delay       time.Duration
		message     *entities.EventEntity[[]testEntity]
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "publish bulk with delay successfully",
			setupRepo: func() *EventProducerRepository[testEntity] {
				mockProducer := event_producer_mocks.NewNSQProducerMock(t)
				mockProducer.On("DeferredPublish", "test-topic", 5*time.Second, mock.AnythingOfType("[]uint8")).Return(nil)
				return &EventProducerRepository[testEntity]{
					eventProducer: &event_producer.EventProducer{
						NSQProducer: mockProducer,
					},
				}
			},
			topic: "test-topic",
			delay: 5 * time.Second,
			message: entities.NewEventEntity("test-event", &[]testEntity{
				{ID: 1, Name: "Entity 1"},
			}),
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, "unexpected error: %v", err)
			},
		},
		{
			name: "publish bulk with delay with NSQProducer.DeferredPublish error",
			setupRepo: func() *EventProducerRepository[testEntity] {
				mockProducer := event_producer_mocks.NewNSQProducerMock(t)
				mockProducer.On("DeferredPublish", "test-topic", 10*time.Second, mock.AnythingOfType("[]uint8")).Return(errors.New("nsq deferred publish error"))
				return &EventProducerRepository[testEntity]{
					eventProducer: &event_producer.EventProducer{
						NSQProducer: mockProducer,
					},
				}
			},
			topic: "test-topic",
			delay: 10 * time.Second,
			message: entities.NewEventEntity("test-event", &[]testEntity{
				{ID: 2, Name: "Entity 2"},
			}),
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "expected error from NSQProducer.DeferredPublish, got nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			ctx := context.Background()
			err := repo.PublishBulkWithDelay(ctx, tt.topic, tt.delay, tt.message)

			if tt.expectError {
				assert.NotNil(t, err, "expected error, got nil")
			} else {
				assert.NoError(t, err, "unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}
		})
	}

	t.Run("publish bulk with delay with json.Marshal error", func(t *testing.T) {
		mockProducer := event_producer_mocks.NewNSQProducerMock(t)
		repo := &EventProducerRepository[testEntityWithChannel]{
			eventProducer: &event_producer.EventProducer{
				NSQProducer: mockProducer,
			},
		}

		message := entities.NewEventEntity("test-event", &[]testEntityWithChannel{
			{ID: 1, Name: "Test", Channel: make(chan int)},
		})

		ctx := context.Background()
		err := repo.PublishBulkWithDelay(ctx, "test-topic", 5*time.Second, message)

		assert.NotNil(t, err, "expected error from json.Marshal, got nil")
	})
}
