package cmd

import (
	"go-boilerplate/configs"
	"go-boilerplate/pkg/tracer"
	"go-boilerplate/transports/grpc"

	"github.com/fikri240794/goteletracer"
	"github.com/spf13/cobra"
)

var (
	grpcServer *grpc.GRPCServer
	grpcCmd    *cobra.Command
)

func initGRPC() {
	grpcCmd = &cobra.Command{
		Use:   "grpc",
		Short: "grpc server",
		Long:  "grpc server command",
		PreRun: func(cmd *cobra.Command, args []string) {
			cfg = configs.Read(cfgPath)
			tracer.NewTracer(&goteletracer.Config{
				ServiceName:         cfg.Server.Tracer.ServiceName,
				ExporterGRPCAddress: cfg.Server.Tracer.ExporterGRPCAddress,
			})
			grpcServer = grpc.BuildGRPCServer(cfg)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return grpcServer.ServeGRPC()
		},
	}
	grpcCmd.Flags().
		StringVarP(
			&cfgPath,
			"cfgpath",
			"c",
			configs.DefaultConfigPath,
			".env config path",
		)
}
