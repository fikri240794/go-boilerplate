package consumers

import (
	"context"
	"go-boilerplate/configs"
	"go-boilerplate/transports/event_consumer/handlers"

	"github.com/fikri240794/gotask"
	"github.com/nsqio/go-nsq"
)

type GuestConsumer struct {
	cfg                  *configs.Config
	createdGuestConsumer *nsq.Consumer
	deletedGuestConsumer *nsq.Consumer
	updatedGuestConsumer *nsq.Consumer
}

func NewGuestConsumer(cfg *configs.Config, handler *handlers.GuestHandler) *GuestConsumer {
	var (
		consumer  *GuestConsumer
		nsqConfig *nsq.Config
		err       error
	)

	consumer = &GuestConsumer{
		cfg: cfg,
	}
	nsqConfig = nsq.NewConfig()

	if cfg.Guest.Event.Created.Enable {
		consumer.createdGuestConsumer, err = nsq.NewConsumer(
			cfg.Guest.Event.Created.Topic,
			cfg.Server.Name,
			nsqConfig,
		)
		if err != nil {
			panic(err)
		}

		consumer.createdGuestConsumer.AddHandler(handlers.NewMessageHandler(handler.HandleCreated))
	}

	if cfg.Guest.Event.Deleted.Enable {
		consumer.deletedGuestConsumer, err = nsq.NewConsumer(
			cfg.Guest.Event.Deleted.Topic,
			cfg.Server.Name,
			nsqConfig,
		)
		if err != nil {
			panic(err)
		}

		consumer.deletedGuestConsumer.AddHandler(handlers.NewMessageHandler(handler.HandleDeleted))
	}

	if cfg.Guest.Event.Updated.Enable {
		consumer.updatedGuestConsumer, err = nsq.NewConsumer(
			cfg.Guest.Event.Updated.Topic,
			cfg.Server.Name,
			nsqConfig,
		)
		if err != nil {
			panic(err)
		}

		consumer.updatedGuestConsumer.AddHandler(handlers.NewMessageHandler(handler.HandleUpdated))
	}

	return consumer
}

func (c *GuestConsumer) ConsumeEvents() error {
	var errTask gotask.ErrorTask
	errTask, _ = gotask.NewErrorTask(context.Background(), 3)

	if c.createdGuestConsumer != nil {
		errTask.Go(func() error {
			return c.createdGuestConsumer.ConnectToNSQLookupd(c.cfg.Server.EventConsumer.DataSourceName)
		})
	}

	if c.deletedGuestConsumer != nil {
		errTask.Go(func() error {
			return c.deletedGuestConsumer.ConnectToNSQLookupd(c.cfg.Server.EventConsumer.DataSourceName)
		})
	}

	if c.updatedGuestConsumer != nil {
		errTask.Go(func() error {
			return c.updatedGuestConsumer.ConnectToNSQLookupd(c.cfg.Server.EventConsumer.DataSourceName)
		})
	}

	return errTask.Wait()
}

func (c *GuestConsumer) Stop() {
	var task gotask.Task = gotask.NewTask(3)

	if c.createdGuestConsumer != nil {
		task.Go(c.createdGuestConsumer.Stop)
	}

	if c.deletedGuestConsumer != nil {
		task.Go(c.deletedGuestConsumer.Stop)
	}

	if c.updatedGuestConsumer != nil {
		task.Go(c.updatedGuestConsumer.Stop)
	}

	task.Wait()
}
