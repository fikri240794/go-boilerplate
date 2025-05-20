package cmd

import (
	"go-boilerplate/configs"
	"go-boilerplate/pkg/tracer"
	"go-boilerplate/transports/event_consumer"

	"github.com/fikri240794/goteletracer"
	"github.com/spf13/cobra"
)

var (
	eventConsumer    *event_consumer.EventConsumer
	eventConsumerCmd *cobra.Command
)

func initEventConsumer() {
	eventConsumerCmd = &cobra.Command{
		Use:   "event-consumer",
		Short: "event consumer",
		Long:  "event consumer command",
		PreRun: func(cmd *cobra.Command, args []string) {
			cfg = configs.Read(cfgPath)
			tracer.NewTracer(&goteletracer.Config{
				ServiceName:         cfg.Server.Tracer.ServiceName,
				ExporterGRPCAddress: cfg.Server.Tracer.ExporterGRPCAddress,
			})
			eventConsumer = event_consumer.BuildEventConsumer(cfg)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return eventConsumer.ConsumeEvents()
		},
	}
	eventConsumerCmd.Flags().
		StringVarP(
			&cfgPath,
			"cfgpath",
			"c",
			configs.DefaultConfigPath,
			".env config path",
		)
}
