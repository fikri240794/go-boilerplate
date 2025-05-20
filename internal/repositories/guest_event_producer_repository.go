package repositories

import (
	"go-boilerplate/datasources/event_producer"
	"go-boilerplate/internal/models/entities"
)

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
