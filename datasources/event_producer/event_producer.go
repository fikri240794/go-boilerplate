package event_producer

import (
	"go-boilerplate/configs"

	"github.com/nsqio/go-nsq"
)

type EventProducer struct {
	NSQProducer *nsq.Producer
}

func Connect(cfg *configs.Config) *EventProducer {
	var (
		nsqProducer *nsq.Producer
		err         error
	)

	nsqProducer, err = nsq.NewProducer(cfg.Datasource.EventProducer.DataSourceName, nsq.NewConfig())
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

func (p *EventProducer) Disconnect() error {
	p.NSQProducer.Stop()
	return nil
}
