package boilerplate_database

import (
	"context"
	"go-boilerplate/configs"
	"time"

	"github.com/fikri240794/gotask"
	"github.com/jmoiron/sqlx"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
)

type BoilerplateDatabase struct {
	Master                        *sqlx.DB
	MasterMaxQueryDurationWarning time.Duration
	Slave                         *sqlx.DB
	SlaveMaxQueryDurationWarning  time.Duration
}

func Connect(cfg *configs.Config) *BoilerplateDatabase {
	var connection *BoilerplateDatabase = &BoilerplateDatabase{
		Master: otelsqlx.MustConnect(
			cfg.Datasource.BoilerplateDatabase.Master.DriverName,
			cfg.Datasource.BoilerplateDatabase.Master.DataSourceName,
		),
		MasterMaxQueryDurationWarning: cfg.Datasource.BoilerplateDatabase.Master.MaximumQueryDurationWarning,
		Slave: otelsqlx.MustConnect(
			cfg.Datasource.BoilerplateDatabase.Slave.DriverName,
			cfg.Datasource.BoilerplateDatabase.Slave.DataSourceName,
		),
		SlaveMaxQueryDurationWarning: cfg.Datasource.BoilerplateDatabase.Slave.MaximumQueryDurationWarning,
	}

	connection.Master.SetMaxOpenConns(cfg.Datasource.BoilerplateDatabase.Master.MaximumOpenConnections)
	connection.Master.SetMaxIdleConns(cfg.Datasource.BoilerplateDatabase.Master.MaximumIddleConnections)
	connection.Master.SetConnMaxIdleTime(cfg.Datasource.BoilerplateDatabase.Master.ConnectionMaximumIdleTime)
	connection.Master.SetConnMaxLifetime(cfg.Datasource.BoilerplateDatabase.Master.ConnectionMaximumLifeTime)

	connection.Slave.SetMaxOpenConns(cfg.Datasource.BoilerplateDatabase.Slave.MaximumOpenConnections)
	connection.Slave.SetMaxIdleConns(cfg.Datasource.BoilerplateDatabase.Slave.MaximumIddleConnections)
	connection.Slave.SetConnMaxIdleTime(cfg.Datasource.BoilerplateDatabase.Slave.ConnectionMaximumIdleTime)
	connection.Slave.SetConnMaxLifetime(cfg.Datasource.BoilerplateDatabase.Slave.ConnectionMaximumLifeTime)

	return connection
}

func (db *BoilerplateDatabase) Disconnect() error {
	var (
		errTask gotask.ErrorTask
		err     error
	)

	errTask, _ = gotask.NewErrorTask(context.Background(), 2)

	errTask.Go(db.Slave.Close)

	errTask.Go(db.Master.Close)

	err = errTask.Wait()

	return err
}
