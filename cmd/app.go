package cmd

import (
	"context"
	"go-boilerplate/configs"
	"go-boilerplate/pkg/tracer"
	"go-boilerplate/transports/event_consumer"
	"go-boilerplate/transports/grpc"
	"go-boilerplate/transports/http"

	"github.com/fikri240794/gotask"
	"github.com/fikri240794/goteletracer"
	"github.com/spf13/cobra"
)

var (
	appCmd *cobra.Command
)

func initApp() {
	appCmd = &cobra.Command{
		Use:   "app",
		Short: "app",
		Long:  "app command",
		PreRun: func(cmd *cobra.Command, args []string) {
			cfg = configs.Read(cfgPath)

			tracer.NewTracer(&goteletracer.Config{
				ServiceName:         cfg.Server.Tracer.ServiceName,
				ExporterGRPCAddress: cfg.Server.Tracer.ExporterGRPCAddress,
			})

			var task gotask.Task = gotask.NewTask(3)

			task.Go(func() {
				httpServer = http.BuildHTTPServer(cfg)
			})

			task.Go(func() {
				grpcServer = grpc.BuildGRPCServer(cfg)
			})

			task.Go(func() {
				eventConsumer = event_consumer.BuildEventConsumer(cfg)
			})

			task.Wait()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				errTask gotask.ErrorTask
				err     error
			)

			errTask, _ = gotask.NewErrorTask(context.Background(), 3)

			errTask.Go(httpServer.ServeHTTP)

			errTask.Go(grpcServer.ServeGRPC)

			errTask.Go(eventConsumer.ConsumeEvents)

			err = errTask.Wait()

			return err
		},
	}
	appCmd.Flags().
		StringVarP(
			&cfgPath,
			"cfgpath",
			"c",
			configs.DefaultConfigPath,
			".env config path",
		)
}
