package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/spf13/cobra"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var (
	databaseMigrationDataSourceName     string
	databaseMigrator                    *migrate.Migrate
	databaseMigrationDirectory          string
	upDatabaseMigrationStep             int
	downDatabaseMigrationStep           int
	toDatabaseMigrationVersion          uint
	forceDatabaseMigrationVersion       uint
	newDatabaseMigrationFileName        string
	createEmptyDatabaseMigrationFileCmd *cobra.Command
	showDatabaseVersionCmd              *cobra.Command
	upDatabaseMigrationCmd              *cobra.Command
	downDatabaseMigrationCmd            *cobra.Command
	toDatabaseMigrationCmd              *cobra.Command
	forceDatabaseMigrationCmd           *cobra.Command
	databaseMigrationCmd                *cobra.Command
)

func createEmptyMigrationFiles(migrationDirectory string, fileName string) error {
	var (
		currentTime   int64
		fileExtension string
		createFile    func(filename string) error
		err           error
	)

	if fileName == "" || regexp.MustCompile(`\s`).MatchString(fileName) {
		return errors.New("filename is empty or have whitespace character(s)")
	}

	migrationDirectory, err = filepath.Abs(migrationDirectory)
	if err != nil {
		return err
	}

	currentTime = time.Now().Unix()
	fileExtension = "sql"
	createFile = func(filename string) error {
		var (
			f   *os.File
			err error
		)

		f, err = os.Create(filename)
		if err != nil {
			return err
		}

		defer f.Close()

		return nil
	}

	err = createFile(fmt.Sprintf(
		"%s/%d_%s.up.%s",
		migrationDirectory,
		currentTime,
		fileName,
		fileExtension,
	))
	if err != nil {
		return err
	}

	err = createFile(fmt.Sprintf(
		"%s/%d_%s.down.%s",
		migrationDirectory,
		currentTime,
		fileName,
		fileExtension,
	))
	if err != nil {
		return err
	}

	return nil
}

func getMigrator(dataSourceName string, migrationDirectory string) (*migrate.Migrate, error) {
	var (
		sourceURL string
		migrator  *migrate.Migrate
		err       error
	)

	sourceURL = fmt.Sprintf("file://%s", migrationDirectory)

	migrator, err = migrate.New(sourceURL, dataSourceName)
	if err != nil {
		return nil, err
	}

	return migrator, nil
}

func showVersion(databaseMigrator *migrate.Migrate) error {
	var (
		version uint
		dirty   bool
		err     error
	)

	version, dirty, err = databaseMigrator.Version()
	if err != nil {
		return err
	}

	fmt.Printf("current database migration version: %d\n", version)
	fmt.Printf("dirty: %t\n", dirty)

	return nil
}

func up(databaseMigrator *migrate.Migrate, migrationStep int) error {
	var err error

	if migrationStep == 0 {
		err = databaseMigrator.Up()
	}

	if migrationStep >= 1 {
		err = databaseMigrator.Steps(migrationStep)
	}

	if migrationStep < 0 {
		err = errors.New("database migration step is less than zero")
	}

	if err != nil {
		return err
	}

	return nil
}

func down(databaseMigrator *migrate.Migrate, migrationStep int) error {
	var err error

	if migrationStep == 0 {
		err = databaseMigrator.Steps(-1)
	}

	if migrationStep <= -1 {
		err = databaseMigrator.Steps(migrationStep)
	}

	if migrationStep > 0 {
		err = errors.New("database migration step is greater than zero")
	}

	if err != nil {
		return err
	}

	return nil
}

func to(databaseMigrator *migrate.Migrate, version uint) error {
	var err error

	if version == 0 {
		err = errors.New("database migration version parameter equal to zero")
		return err
	}

	err = databaseMigrator.Migrate(version)
	if err != nil {
		return err
	}

	return nil
}

func force(databaseMigrator *migrate.Migrate, version uint) error {
	var err error

	if version == 0 {
		err = errors.New("database migration version parameter equal to zero")
		return err
	}

	err = databaseMigrator.Force(int(version))
	if err != nil {
		return err
	}

	return nil
}

func initCreateEmptyDatabaseMigrationFileCmd() {
	createEmptyDatabaseMigrationFileCmd = &cobra.Command{
		Use:   "new",
		Short: "new database migration file",
		Long:  "create empty database migration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return createEmptyMigrationFiles(
				databaseMigrationDirectory,
				newDatabaseMigrationFileName,
			)
		},
	}
	createEmptyDatabaseMigrationFileCmd.Flags().
		StringVarP(
			&databaseMigrationDirectory,
			"migration_directory",
			"d",
			"",
			"relative path to migration directory",
		)
	createEmptyDatabaseMigrationFileCmd.MarkFlagRequired("migration_directory")
	createEmptyDatabaseMigrationFileCmd.Flags().
		StringVarP(
			&newDatabaseMigrationFileName,
			"filename",
			"f",
			"",
			"migration file name (no whitespace)",
		)
	createEmptyDatabaseMigrationFileCmd.MarkFlagRequired("filename")
}

func setRequiredFlags(cmd *cobra.Command) {
	cmd.Flags().
		StringVarP(
			&databaseMigrationDataSourceName,
			"data_source_name",
			"t",
			"",
			"database connection string in URL format, example for postgres: \"postgres://user:password@host:port/database_name?sslmode=disable\" (with double quotation mark)",
		)
	cmd.MarkFlagRequired("data_source_name")

	cmd.Flags().
		StringVarP(
			&databaseMigrationDirectory,
			"migration_directory",
			"d",
			"",
			"relative path to migration directory, example: ./datasources/boilerplate_database/migrations",
		)
	cmd.MarkFlagRequired("migration_directory")
}

func initShowDatabaseVersionCmd() {
	showDatabaseVersionCmd = &cobra.Command{
		Use:   "version",
		Short: "database migration version",
		Long:  "show database migration version",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			databaseMigrator, err = getMigrator(
				databaseMigrationDataSourceName,
				databaseMigrationDirectory,
			)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			defer databaseMigrator.Close()
			return showVersion(databaseMigrator)
		},
	}
	setRequiredFlags(showDatabaseVersionCmd)
}

func initUpDatabaseMigrationCmd() {
	upDatabaseMigrationCmd = &cobra.Command{
		Use:   "up",
		Short: "up database migration",
		Long:  "apply all or N up database migration(s)",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			databaseMigrator, err = getMigrator(
				databaseMigrationDataSourceName,
				databaseMigrationDirectory,
			)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			defer databaseMigrator.Close()
			return up(databaseMigrator, upDatabaseMigrationStep)
		},
	}
	setRequiredFlags(upDatabaseMigrationCmd)
	upDatabaseMigrationCmd.Flags().
		IntVarP(
			&upDatabaseMigrationStep,
			"step",
			"s",
			0,
			"N migration step",
		)
}

func initDownDatabaseMigrationCmd() {
	downDatabaseMigrationCmd = &cobra.Command{
		Use:   "down",
		Short: "down database migration",
		Long:  "down -1 or N up database migration(s)",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			databaseMigrator, err = getMigrator(
				databaseMigrationDataSourceName,
				databaseMigrationDirectory,
			)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			defer databaseMigrator.Close()
			return down(databaseMigrator, downDatabaseMigrationStep)
		},
	}
	setRequiredFlags(downDatabaseMigrationCmd)
	downDatabaseMigrationCmd.Flags().
		IntVarP(
			&downDatabaseMigrationStep,
			"step",
			"s",
			-1,
			"N migration step",
		)
}

func initToDatabaseMigrationCmd() {
	toDatabaseMigrationCmd = &cobra.Command{
		Use:   "to",
		Short: "database migration to version V",
		Long:  "run database migration (up/down) to specific V version",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			databaseMigrator, err = getMigrator(
				databaseMigrationDataSourceName,
				databaseMigrationDirectory,
			)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			defer databaseMigrator.Close()
			return to(databaseMigrator, toDatabaseMigrationVersion)
		},
	}
	setRequiredFlags(toDatabaseMigrationCmd)
	toDatabaseMigrationCmd.Flags().
		UintVarP(
			&toDatabaseMigrationVersion,
			"version",
			"v",
			0,
			"V migration version",
		)
	toDatabaseMigrationCmd.MarkFlagRequired("version")
}

func initForceDatabaseMigrationCmd() {
	forceDatabaseMigrationCmd = &cobra.Command{
		Use:   "force",
		Short: "force update database migration version to V",
		Long:  "force update database migration version to V version without run interface{} database migration script",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			databaseMigrator, err = getMigrator(
				databaseMigrationDataSourceName,
				databaseMigrationDirectory,
			)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			defer databaseMigrator.Close()
			return force(databaseMigrator, forceDatabaseMigrationVersion)
		},
	}
	setRequiredFlags(forceDatabaseMigrationCmd)
	forceDatabaseMigrationCmd.Flags().
		UintVarP(
			&forceDatabaseMigrationVersion,
			"version",
			"v",
			0,
			"V migration version",
		)
	forceDatabaseMigrationCmd.MarkFlagRequired("version")
}

func initDatabaseMigration() {
	initCreateEmptyDatabaseMigrationFileCmd()
	initShowDatabaseVersionCmd()
	initUpDatabaseMigrationCmd()
	initDownDatabaseMigrationCmd()
	initToDatabaseMigrationCmd()
	initForceDatabaseMigrationCmd()

	databaseMigrationCmd = &cobra.Command{
		Use:   "migrate",
		Short: "database migration",
		Long:  "database migration command",
	}

	databaseMigrationCmd.AddCommand(
		createEmptyDatabaseMigrationFileCmd,
		showDatabaseVersionCmd,
		upDatabaseMigrationCmd,
		downDatabaseMigrationCmd,
		toDatabaseMigrationCmd,
	)
}
