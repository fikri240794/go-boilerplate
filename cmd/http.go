package cmd

import (
	"go-boilerplate/configs"
	"go-boilerplate/pkg/tracer"
	"go-boilerplate/transports/http"

	"github.com/fikri240794/goteletracer"
	"github.com/spf13/cobra"
)

var (
	httpServer *http.HTTPServer
	httpCmd    *cobra.Command
)

func initHTTP() {
	httpCmd = &cobra.Command{
		Use:   "http",
		Short: "http server",
		Long:  "http server command",
		PreRun: func(cmd *cobra.Command, args []string) {
			cfg = configs.Read(cfgPath)
			tracer.NewTracer(&goteletracer.Config{
				ServiceName:         cfg.Server.Tracer.ServiceName,
				ExporterGRPCAddress: cfg.Server.Tracer.ExporterGRPCAddress,
			})
			httpServer = http.BuildHTTPServer(cfg)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return httpServer.ServeHTTP()
		},
	}
	httpCmd.Flags().
		StringVarP(
			&cfgPath,
			"cfgpath",
			"c",
			configs.DefaultConfigPath,
			".env config path",
		)
}
