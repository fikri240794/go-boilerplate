package cmd

import (
	"context"
	"go-boilerplate/configs"
	"go-boilerplate/pkg/tracer"
	"go-boilerplate/transports/event_consumer"
	"go-boilerplate/transports/grpc"
	"go-boilerplate/transports/http"
	"log"

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
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[ERROR] Panic recovered in PreRun: %v", r)
				}
			}()

			cfg = configs.Read(cfgPath)

			tracer.NewTracer(&goteletracer.Config{
				ServiceName:         cfg.Server.Tracer.ServiceName,
				ExporterGRPCAddress: cfg.Server.Tracer.ExporterGRPCAddress,
			})

			var task gotask.Task = gotask.NewTask(3)

			task.Go(func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[ERROR] Panic recovered while building HTTP server: %v", r)
					}
				}()
				httpServer = http.BuildHTTPServer(cfg)
			})

			task.Go(func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[ERROR] Panic recovered while building GRPC server: %v", r)
					}
				}()
				grpcServer = grpc.BuildGRPCServer(cfg)
			})

			task.Go(func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[ERROR] Panic recovered while building event consumer: %v", r)
					}
				}()
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

			errTask.Go(func() error {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[ERROR] Panic recovered while serving HTTP: %v", r)
					}
				}()
				return httpServer.ServeHTTP()
			})

			errTask.Go(func() error {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[ERROR] Panic recovered while serving GRPC: %v", r)
					}
				}()
				return grpcServer.ServeGRPC()
			})

			errTask.Go(func() error {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[ERROR] Panic recovered while consuming events: %v", r)
					}
				}()
				return eventConsumer.ConsumeEvents()
			})

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
