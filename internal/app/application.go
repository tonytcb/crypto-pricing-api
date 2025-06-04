package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/tonytcb/crypto-pricing-api/internal/api"
	"github.com/tonytcb/crypto-pricing-api/internal/api/http_handlers"
	"github.com/tonytcb/crypto-pricing-api/internal/app/config"
	"github.com/tonytcb/crypto-pricing-api/internal/infra/coindesk"
	"github.com/tonytcb/crypto-pricing-api/internal/infra/event_listener"
	"github.com/tonytcb/crypto-pricing-api/internal/infra/event_provider"
	"github.com/tonytcb/crypto-pricing-api/internal/infra/sse"
	"github.com/tonytcb/crypto-pricing-api/internal/infra/storage/in_memory"
)

type Application struct {
	cfg            *config.Config
	log            *slog.Logger
	httpServer     *api.HTTPServer
	eventProvider  *event_provider.HTTPPulling
	eventListener  *event_listener.PricesListener
	clientsManager *sse.Hub
}

func NewApplication(ctx context.Context, cfg *config.Config, log *slog.Logger) (*Application, error) {
	pairToMonitor, err := cfg.PairToMonitor()
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse pair to monitor configuration")
	}

	var (
		coinDeskHTTPClient  = &http.Client{Timeout: cfg.CoinDeskClientTimeout}
		priceAPI            = coindesk.NewPricingAPI(coinDeskHTTPClient, cfg)
		pricesEventProvider = event_provider.NewHTTPPulling(priceAPI, cfg.PricesPullingInterval, cfg.PricesChannelBufferSize)
		pricesRepo          = in_memory.NewPricesBySliceRepo(cfg.StoreMaxItems)
		clientsManager      = sse.NewHub(pricesRepo, cfg.SSEClientsCleanUpInterval)
	)

	eventsChan, err := pricesEventProvider.Start(ctx, pairToMonitor)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start event provider")
	}

	eventListener := event_listener.NewPricesListener(clientsManager, eventsChan)

	handlers := api.HTTPHandlers{
		HealthHandler:         http_handlers.NewHealthHandler(),
		PriceStreamingHandler: http_handlers.NewPriceStreamer(cfg, clientsManager),
	}

	httpServer := api.NewHTTPServer(log, cfg, handlers)

	return &Application{
		cfg:            cfg,
		log:            log,
		httpServer:     httpServer,
		eventProvider:  pricesEventProvider,
		eventListener:  eventListener,
		clientsManager: clientsManager,
	}, nil
}

func (a Application) Run(ctx context.Context) error {
	errGroup, _ := errgroup.WithContext(ctx)

	a.log.Info("Running application")

	errGroup.Go(func() error {
		a.log.Info("Starting http server", "port", a.cfg.RestAPIPort)

		return a.httpServer.Start()
	})

	errGroup.Go(func() error {
		a.log.Info("Starting prices event listener")

		return a.eventListener.Start(ctx)
	})

	errGroup.Go(func() error {
		a.log.Info("Starting clients manager")

		a.clientsManager.Start()

		return nil
	})

	return errGroup.Wait()
}

func (a Application) Stop() {
	a.log.Info("Stopping application")

	_ = a.httpServer.Stop()
	a.clientsManager.Stop()
}
