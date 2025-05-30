package app

import (
	"context"
	"github.com/tonytcb/crypto-pricing-api/internal/api/http_handlers"
	"log/slog"

	"golang.org/x/sync/errgroup"

	"github.com/tonytcb/crypto-pricing-api/internal/api"
	"github.com/tonytcb/crypto-pricing-api/internal/app/config"
)

type Application struct {
	cfg        *config.Config
	log        *slog.Logger
	httpServer *api.HTTPServer
}

func NewApplication(ctx context.Context, cfg *config.Config, log *slog.Logger) (*Application, error) {
	handlers := api.HTTPHandlers{
		HealthHandler: http_handlers.NewHealthHandler(),
	}

	httpServer := api.NewHTTPServer(log, cfg, handlers)

	return &Application{
		cfg:        cfg,
		log:        log,
		httpServer: httpServer,
	}, nil
}

func (a Application) Run(ctx context.Context) error {
	errGroup, _ := errgroup.WithContext(ctx)

	a.log.Info("Running application")

	errGroup.Go(func() error {
		a.log.Info("Starting http server", "port", a.cfg.RestAPIPort)

		return a.httpServer.Start()
	})

	return nil
}

func (a Application) Stop() {
	a.log.Info("Stopping application")

	_ = a.httpServer.Stop()
}
