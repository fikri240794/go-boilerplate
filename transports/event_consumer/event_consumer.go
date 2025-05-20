package event_consumer

//go:generate go run github.com/google/wire/cmd/wire

import (
	"fmt"
	"go-boilerplate/configs"
	"go-boilerplate/datasources"
	"go-boilerplate/transports/event_consumer/consumers"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type EventConsumer struct {
	cfg            *configs.Config
	datasources    *datasources.Datasources
	eventConsumers *consumers.Consumers
}

func NewEventConsumer(
	cfg *configs.Config,
	ds *datasources.Datasources,
	eventConsumers *consumers.Consumers,
) *EventConsumer {
	return &EventConsumer{
		cfg:            cfg,
		datasources:    ds,
		eventConsumers: eventConsumers,
	}
}

func (c *EventConsumer) gracefullyShutdown() {
	var (
		ticker               *time.Ticker
		tickCounter          float64
		tickMessage          string
		maxTickMessageLength int
		stopCompleteChan     = make(chan bool)
	)

	tickCounter = 0
	ticker = time.NewTicker(1 * time.Millisecond)

	go func() {
		c.datasources.Disconnect()
		c.eventConsumers.Stop()
		stopCompleteChan <- true
	}()

	fmt.Print("\n\n")

	for {
		select {
		case <-ticker.C:
			tickMessage = fmt.Sprintf("shutting down Event Consumer in %.3fs", tickCounter/1000)

			if len(tickMessage) > maxTickMessageLength {
				maxTickMessageLength = len(tickMessage)
			}

			fmt.Printf("\r%*s", maxTickMessageLength, "")
			fmt.Printf("\r%s", tickMessage)

			tickCounter++

		case <-stopCompleteChan:
			ticker.Stop()

			tickMessage = "Event Consumer shutdown process finished successfully\n\n"

			fmt.Printf("\r%*s", maxTickMessageLength, "")
			fmt.Printf("\r%s", tickMessage)
			return
		}
	}
}

func (c *EventConsumer) setGlobalLog() {
	zerolog.SetGlobalLevel(zerolog.Level(c.cfg.Server.LogLevel))
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})
}

func (c *EventConsumer) ConsumeEvents() error {
	var (
		signalListener chan os.Signal
		err            error
	)

	c.setGlobalLog()

	err = c.eventConsumers.ConsumeEvents()
	if err != nil {
		return err
	}

	signalListener = make(chan os.Signal, 1)
	signal.Notify(signalListener, os.Interrupt)
	<-signalListener

	c.gracefullyShutdown()

	return nil
}
