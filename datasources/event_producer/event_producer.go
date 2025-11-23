package event_producer

import (
	"go-boilerplate/configs"
	"time"

	"github.com/nsqio/go-nsq"
)

//go:generate go run github.com/vektra/mockery/v2 --name INSQProducer --structname NSQProducerMock --filename nsq_producer_mock.go
type INSQProducer interface {
	Ping() error
	Stop()
	Publish(topic string, body []byte) error
	DeferredPublish(topic string, delay time.Duration, body []byte) error
}

type EventProducer struct {
	NSQProducer INSQProducer
}

type nsqProducer func(addr string, config *nsq.Config) (INSQProducer, error)

var defaultNSQProducer nsqProducer = func(addr string, config *nsq.Config) (INSQProducer, error) {
	return nsq.NewProducer(addr, config)
}

func connectToNSQProducer(cfg *configs.Config, fn nsqProducer) *EventProducer {
	var (
		nsqProducer INSQProducer
		err         error
	)

	nsqProducer, err = fn(cfg.Datasource.EventProducer.DataSourceName, nsq.NewConfig())
	if err != nil {
		panic(err)
	}

	err = nsqProducer.Ping()
	if err != nil {
		panic(err)
	}

	return &EventProducer{
		NSQProducer: nsqProducer,
	}
}

func Connect(cfg *configs.Config) *EventProducer {
	return connectToNSQProducer(cfg, defaultNSQProducer)
}

func (p *EventProducer) Disconnect() error {
	p.NSQProducer.Stop()
	return nil
}
