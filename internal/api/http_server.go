package api

import (
	"context"
	"github.com/pkg/errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tonytcb/crypto-pricing-api/internal/app/config"
)

type HealthHandler interface {
	IsHealthy(c *gin.Context)
}

type HTTPHandlers struct {
	HealthHandler HealthHandler
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

	m.srv = &http.Server{
		Addr:         m.cfg.RestAPIPort,
		Handler:      router.Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
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
