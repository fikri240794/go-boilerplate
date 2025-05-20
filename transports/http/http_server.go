package http

//go:generate go run github.com/google/wire/cmd/wire
//go:generate sh -c "cd ./../../ && go run github.com/swaggo/swag/cmd/swag init --generalInfo ./transports/http/http_server.go --parseDependency --output ./transports/http/docs/swagger"

import (
	"fmt"
	"go-boilerplate/configs"
	"go-boilerplate/datasources"
	"go-boilerplate/transports/http/handlers"
	"go-boilerplate/transports/http/middlewares"
	"os"
	"os/signal"
	"time"

	"github.com/fikri240794/gocerr"
	"github.com/fikri240794/gores"
	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// @title Boilerplate API Docs
// @description Boilerplate API Docs

// remove the first // to enable and add "@Security Bearer" annotation to each API that requires it
// // @securityDefinitions.apikey Bearer
// // @in  header
// // @name Authorization
type HTTPServer struct {
	cfg         *configs.Config
	server      *fiber.App
	datasources *datasources.Datasources
	middlewares *middlewares.Middlewares
	handlers    *handlers.Handlers
}

func NewHTTPServer(
	cfg *configs.Config,
	ds *datasources.Datasources,
	mw *middlewares.Middlewares,
	h *handlers.Handlers,
) *HTTPServer {
	var httpServer *HTTPServer = &HTTPServer{
		cfg: cfg,
		server: fiber.New(fiber.Config{
			AppName:           cfg.Server.Name,
			Prefork:           cfg.Server.HTTP.Prefork,
			EnablePrintRoutes: cfg.Server.HTTP.PrintRoutes,
			ReduceMemoryUsage: true,
			JSONEncoder:       json.Marshal,
			JSONDecoder:       json.Unmarshal,
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				err = gocerr.New(fiber.StatusInternalServerError, err.Error())
				var responseVM *gores.ResponseVM[string] = gores.NewResponseVM[string]().
					SetErrorFromError(err)
				return c.Status(responseVM.Code).
					JSON(responseVM)
			},
		}),
		datasources: ds,
		middlewares: mw,
		handlers:    h,
	}

	return httpServer
}

func (s *HTTPServer) gracefullyShutdown() {
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
		s.server.ShutdownWithTimeout(s.cfg.Server.HTTP.GracefullyShutdownDuration)
		shutdownCompleteChan <- true
	}()

	fmt.Print("\n\n")

	for {
		select {
		case <-ticker.C:
			tickMessage = fmt.Sprintf("shutting down HTTP server in %.3fs", tickCounter/1000)

			if len(tickMessage) > maxTickMessageLength {
				maxTickMessageLength = len(tickMessage)
			}

			fmt.Printf("\r%*s", maxTickMessageLength, "")
			fmt.Printf("\r%s", tickMessage)

			tickCounter++

		case <-shutdownCompleteChan:
			ticker.Stop()

			tickMessage = "HTTP server shutdown process finished successfully\n\n"

			fmt.Printf("\r%*s", maxTickMessageLength, "")
			fmt.Printf("\r%s", tickMessage)
			return
		}
	}
}

func (s *HTTPServer) setGlobalLog() {
	zerolog.SetGlobalLevel(zerolog.Level(s.cfg.Server.LogLevel))
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})
}

func (s *HTTPServer) setupGlobalMiddlewares() {
	s.server.Use(
		s.middlewares.Recover.Recover,
		s.middlewares.Tracer.Start,
		s.middlewares.RequestID.Generate,
		s.middlewares.Log.Log,
		cors.New(cors.Config{
			AllowOrigins: s.cfg.Server.HTTP.CORS.AllowOrigins,
			AllowMethods: s.cfg.Server.HTTP.CORS.AllowMethods,
		}),
		etag.New(),
		favicon.New(),
	)

	if s.cfg.Server.HTTP.Docs.Swagger.Enable {
		s.server.Use(swagger.New(swagger.Config{
			FilePath: s.cfg.Server.HTTP.Docs.Swagger.FilePath,
			Path:     s.cfg.Server.HTTP.Docs.Swagger.Path,
			Title:    s.cfg.Server.HTTP.Docs.Swagger.Title,
			CacheAge: 1,
		}))
	}

	s.server.Use(s.middlewares.Timeout.Timeout)
}

func (s *HTTPServer) ServeHTTP() error {
	var (
		serverErrListener chan error
		signalListener    chan os.Signal
		err               error
	)

	s.setGlobalLog()
	s.setupGlobalMiddlewares()
	s.handlers.SetupRoutes(s.server)

	serverErrListener = make(chan error, 1)
	signalListener = make(chan os.Signal, 1)

	go func() {
		serverErrListener <- s.server.Listen(fmt.Sprintf(":%d", s.cfg.Server.HTTP.Port))
	}()

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
