package repositories

import (
	"go-boilerplate/datasources/event_producer"
	"go-boilerplate/internal/models/entities"
)

//go:generate go run github.com/vektra/mockery/v2 --name IGuestEventProducerRepository --structname GuestEventProducerRepositoryMock --filename guest_event_producer_repository_mock.go
type IGuestEventProducerRepository interface {
	IEventProducerRepository[entities.GuestEventEntity]
}

type GuestEventProducerRepository struct {
	EventProducerRepository[entities.GuestEventEntity]
}

func NewGuestEventProducerRepository(eventProducer *event_producer.EventProducer) *GuestEventProducerRepository {
	return &GuestEventProducerRepository{
		EventProducerRepository[entities.GuestEventEntity]{
			eventProducer: eventProducer,
		},
	}
}
