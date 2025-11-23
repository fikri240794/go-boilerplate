package repositories

import (
	"go-boilerplate/datasources/event_producer"
	event_producer_mocks "go-boilerplate/datasources/event_producer/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewGuestEventProducerRepository(t *testing.T) {
	tests := []struct {
		name          string
		eventProducer *event_producer.EventProducer
		expectNil     bool
	}{
		{
			name: "create guest event producer repository with event producer",
			eventProducer: &event_producer.EventProducer{
				NSQProducer: event_producer_mocks.NewNSQProducerMock(t),
			},
			expectNil: false,
		},
		{
			name:          "create guest event producer repository without event producer",
			eventProducer: nil,
			expectNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewGuestEventProducerRepository(tt.eventProducer)

			if tt.expectNil {
				assert.Nil(t, repo, "NewGuestEventProducerRepository() expected nil, got %v", repo)
			} else {
				assert.NotNil(t, repo, "NewGuestEventProducerRepository() expected non-nil repository, got nil")
				assert.Equal(t, tt.eventProducer, repo.eventProducer, "NewGuestEventProducerRepository() eventProducer mismatch")
			}
		})
	}
}
