package consumers

import (
	"context"

	"github.com/fikri240794/gotask"
)

type Consumers struct {
	Guest *GuestConsumer
}

func (c *Consumers) ConsumeEvents() error {
	var errTask gotask.ErrorTask
	errTask, _ = gotask.NewErrorTask(context.Background(), 1)

	errTask.Go(c.Guest.ConsumeEvents)

	return errTask.Wait()
}

func (c *Consumers) Stop() {
	var task gotask.Task = gotask.NewTask(1)

	task.Go(c.Guest.Stop)

	task.Wait()
}
