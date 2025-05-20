package cmd

import (
	"go-boilerplate/configs"

	"github.com/spf13/cobra"
)

var (
	cfgPath string
	cfg     *configs.Config
	rootCmd *cobra.Command
)

func init() {
	initDatabaseMigration()
	initHTTP()
	initGRPC()
	initEventConsumer()
	initApp()
	rootCmd = &cobra.Command{
		Long: "boilerplate",
	}
	rootCmd.AddCommand(
		databaseMigrationCmd,
		httpCmd,
		grpcCmd,
		eventConsumerCmd,
		appCmd,
	)
}

func Execute() {
	var err error = rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
