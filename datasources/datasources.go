package datasources

import (
	"context"
	"go-boilerplate/datasources/boilerplate_database"
	"go-boilerplate/datasources/event_producer"
	"go-boilerplate/datasources/in_memory_database"

	"github.com/fikri240794/gotask"
)

type Datasources struct {
	BoilerplateDatabase *boilerplate_database.BoilerplateDatabase
	InMemoryDatabase    *in_memory_database.InMemoryDatabase
	EventProducer       *event_producer.EventProducer
}

func (ds *Datasources) Disconnect() error {
	var (
		errTask gotask.ErrorTask
		err     error
	)

	errTask, _ = gotask.NewErrorTask(context.Background(), 3)

	errTask.Go(ds.BoilerplateDatabase.Disconnect)

	errTask.Go(ds.InMemoryDatabase.Disconnect)

	errTask.Go(ds.EventProducer.Disconnect)

	err = errTask.Wait()

	return err
}
