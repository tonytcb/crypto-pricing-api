package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tonytcb/crypto-pricing-api/internal/app"
	"github.com/tonytcb/crypto-pricing-api/internal/app/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load("./default.env", "./config/default.env")
	if err != nil {
		panic("unable to load env configurations:" + err.Error())
	}

	var log = slog.Default()

	log.Info("Initializing application")

	application, err := app.NewApplication(ctx, cfg, log)
	if err != nil {
		panic("Failed to create application:" + err.Error())
	}

	if err = application.Run(ctx); err != nil {
		panic("Failed to start application: " + err.Error())
	}

	<-ctx.Done() // wait until we receive a signal to stop the app

	log.Info("Shutting down application")

	application.Stop()

	log.Info("Exit application")

	os.Exit(0)
}
