package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/tonytcb/crypto-pricing-api/internal/app/config"
)

type HealthHandler interface {
	IsHealthy(c *gin.Context)
}

type CorsHandler interface {
	Allowed(c *gin.Context)
}

type PriceStreamingHandler interface {
	Stream(c *gin.Context)
}

type HTTPHandlers struct {
	HealthHandler         HealthHandler
	CorsHandler           CorsHandler
	PriceStreamingHandler PriceStreamingHandler
}

type HTTPServer struct {
	log      *slog.Logger
	srv      *http.Server
	cfg      *config.Config
	handlers HTTPHandlers
}

func NewHTTPServer(
	log *slog.Logger,
	cfg *config.Config,
	handlers HTTPHandlers,
) *HTTPServer {
	return &HTTPServer{
		log:      log,
		cfg:      cfg,
		handlers: handlers,
	}
}

func (m *HTTPServer) Start() error {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/health", m.handlers.HealthHandler.IsHealthy)
	router.GET("/prices/:pair/stream", m.handlers.CorsHandler.Allowed, m.handlers.PriceStreamingHandler.Stream)

	m.srv = &http.Server{
		Addr:    m.cfg.RestAPIPort,
		Handler: router.Handler(),
	}

	if err := m.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (m *HTTPServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.srv.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "error to shutdown http server")
	}

	return nil
}
