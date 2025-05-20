package grpc

//go:generate go run github.com/google/wire/cmd/wire

import (
	"fmt"
	"go-boilerplate/configs"
	"go-boilerplate/datasources"
	"go-boilerplate/pkg/protobuf_boilerplate"
	"go-boilerplate/transports/grpc/handlers"
	"go-boilerplate/transports/grpc/middlewares"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zs5460/art"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	cfg         *configs.Config
	datasources *datasources.Datasources
	server      *grpc.Server
	handlers    *handlers.ImplementedBoilerplateServer
	middlewares *middlewares.Middlewares
}

func NewGRPCServer(
	cfg *configs.Config,
	ds *datasources.Datasources,
	h *handlers.ImplementedBoilerplateServer,
	mw *middlewares.Middlewares,
) *GRPCServer {
	var grpcServer *grpc.Server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(mw.GetUnaryServerInterceptors()...),
	)
	protobuf_boilerplate.RegisterBoilerplateServer(grpcServer, h)

	return &GRPCServer{
		cfg:         cfg,
		datasources: ds,
		server:      grpcServer,
		handlers:    h,
		middlewares: mw,
	}
}

func (s *GRPCServer) gracefullyShutdown() {
	var (
		ticker               *time.Ticker
		tickCounter          float64
		tickMessage          string
		maxTickMessageLength int
		shutdownCompleteChan = make(chan bool)
	)

	tickCounter = 0
	ticker = time.NewTicker(1 * time.Millisecond)

	go func() {
		s.datasources.Disconnect()
		s.server.GracefulStop()
		shutdownCompleteChan <- true
	}()

	fmt.Print("\n\n")

	for {
		select {
		case <-ticker.C:
			tickMessage = fmt.Sprintf("shutting down GRPC server in %.3fs", tickCounter/1000)

			if len(tickMessage) > maxTickMessageLength {
				maxTickMessageLength = len(tickMessage)
			}

			fmt.Printf("\r%*s", maxTickMessageLength, "")
			fmt.Printf("\r%s", tickMessage)

			tickCounter++

		case <-shutdownCompleteChan:
			ticker.Stop()

			tickMessage = "GRPC server shutdown process finished successfully\n\n"

			fmt.Printf("\r%*s", maxTickMessageLength, "")
			fmt.Printf("\r%s", tickMessage)
			return
		}
	}
}

func (s *GRPCServer) setGlobalLog() {
	zerolog.SetGlobalLevel(zerolog.Level(s.cfg.Server.LogLevel))
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})
}

func (s *GRPCServer) ServeGRPC() error {
	var (
		netListener       net.Listener
		serverErrListener chan error
		signalListener    chan os.Signal
		err               error
	)

	s.setGlobalLog()

	netListener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.Server.GRPC.Port))
	if err != nil {
		return err
	}

	serverErrListener = make(chan error, 1)
	signalListener = make(chan os.Signal, 1)

	go func() {
		serverErrListener <- s.server.Serve(netListener)
	}()

	fmt.Println(art.String("GRPC"))
	fmt.Printf("GRPC Server is listening on :%d\n\n", s.cfg.Server.GRPC.Port)

	signal.Notify(signalListener, os.Interrupt)

	select {
	case err = <-serverErrListener:
		s.gracefullyShutdown()
		return err
	case <-signalListener:
		s.gracefullyShutdown()
	}

	return nil
}
