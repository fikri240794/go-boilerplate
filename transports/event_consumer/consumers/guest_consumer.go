package consumers

import (
	"context"
	"go-boilerplate/configs"
	"go-boilerplate/transports/event_consumer/handlers"

	"github.com/fikri240794/gotask"
	"github.com/nsqio/go-nsq"
)

type GuestConsumer struct {
	cfg                      *configs.Config
	createdGuestConsumer     *nsq.Consumer
	deletedGuestConsumer     *nsq.Consumer
	updatedGuestConsumer     *nsq.Consumer
	bulkCreatedGuestConsumer *nsq.Consumer
	bulkUpdatedGuestConsumer *nsq.Consumer
	bulkDeletedGuestConsumer *nsq.Consumer
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

	if cfg.Guest.Event.BulkCreated.Enable {
		consumer.bulkCreatedGuestConsumer, err = nsq.NewConsumer(
			cfg.Guest.Event.BulkCreated.Topic,
			cfg.Server.Name,
			nsqConfig,
		)
		if err != nil {
			panic(err)
		}

		consumer.bulkCreatedGuestConsumer.AddHandler(handlers.NewMessageHandler(handler.HandleBulkCreated))
	}

	if cfg.Guest.Event.BulkUpdated.Enable {
		consumer.bulkUpdatedGuestConsumer, err = nsq.NewConsumer(
			cfg.Guest.Event.BulkUpdated.Topic,
			cfg.Server.Name,
			nsqConfig,
		)
		if err != nil {
			panic(err)
		}

		consumer.bulkUpdatedGuestConsumer.AddHandler(handlers.NewMessageHandler(handler.HandleBulkUpdated))
	}

	if cfg.Guest.Event.BulkDeleted.Enable {
		consumer.bulkDeletedGuestConsumer, err = nsq.NewConsumer(
			cfg.Guest.Event.BulkDeleted.Topic,
			cfg.Server.Name,
			nsqConfig,
		)
		if err != nil {
			panic(err)
		}

		consumer.bulkDeletedGuestConsumer.AddHandler(handlers.NewMessageHandler(handler.HandleBulkDeleted))
	}

	return consumer
}

func (c *GuestConsumer) ConsumeEvents() error {
	var (
		taskCount int
		errTask   gotask.ErrorTask
	)

	if c.createdGuestConsumer != nil {
		taskCount++
	}
	if c.deletedGuestConsumer != nil {
		taskCount++
	}
	if c.updatedGuestConsumer != nil {
		taskCount++
	}
	if c.bulkCreatedGuestConsumer != nil {
		taskCount++
	}
	if c.bulkUpdatedGuestConsumer != nil {
		taskCount++
	}
	if c.bulkDeletedGuestConsumer != nil {
		taskCount++
	}

	errTask, _ = gotask.NewErrorTask(context.Background(), taskCount)

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

	if c.bulkCreatedGuestConsumer != nil {
		errTask.Go(func() error {
			return c.bulkCreatedGuestConsumer.ConnectToNSQLookupd(c.cfg.Server.EventConsumer.DataSourceName)
		})
	}

	if c.bulkUpdatedGuestConsumer != nil {
		errTask.Go(func() error {
			return c.bulkUpdatedGuestConsumer.ConnectToNSQLookupd(c.cfg.Server.EventConsumer.DataSourceName)
		})
	}

	if c.bulkDeletedGuestConsumer != nil {
		errTask.Go(func() error {
			return c.bulkDeletedGuestConsumer.ConnectToNSQLookupd(c.cfg.Server.EventConsumer.DataSourceName)
		})
	}

	return errTask.Wait()
}

func (c *GuestConsumer) Stop() {
	var (
		taskCount int
		task      gotask.Task
	)

	if c.createdGuestConsumer != nil {
		taskCount++
	}
	if c.deletedGuestConsumer != nil {
		taskCount++
	}
	if c.updatedGuestConsumer != nil {
		taskCount++
	}
	if c.bulkCreatedGuestConsumer != nil {
		taskCount++
	}
	if c.bulkUpdatedGuestConsumer != nil {
		taskCount++
	}
	if c.bulkDeletedGuestConsumer != nil {
		taskCount++
	}

	task = gotask.NewTask(taskCount)

	if c.createdGuestConsumer != nil {
		task.Go(c.createdGuestConsumer.Stop)
	}
	if c.deletedGuestConsumer != nil {
		task.Go(c.deletedGuestConsumer.Stop)
	}
	if c.updatedGuestConsumer != nil {
		task.Go(c.updatedGuestConsumer.Stop)
	}
	if c.bulkCreatedGuestConsumer != nil {
		task.Go(c.bulkCreatedGuestConsumer.Stop)
	}
	if c.bulkUpdatedGuestConsumer != nil {
		task.Go(c.bulkUpdatedGuestConsumer.Stop)
	}
	if c.bulkDeletedGuestConsumer != nil {
		task.Go(c.bulkDeletedGuestConsumer.Stop)
	}

	task.Wait()
}
