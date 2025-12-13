package repositories

import (
	"go-boilerplate/datasources/event_producer"
	"go-boilerplate/internal/models/entities"
)

//mockery:generate: true
//mockery:structname: GuestEventProducerRepositoryMock
//mockery:filename: guest_event_producer_repository_mock.go
//mockery:output: internal/repositories/mocks/
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
