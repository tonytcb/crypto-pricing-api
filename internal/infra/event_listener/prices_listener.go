package event_listener

import (
	"context"
	"log/slog"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

type Notifier interface {
	Broadcast(update domain.PriceUpdate)
}

type PricesListener struct {
	log        *slog.Logger
	notifier   Notifier
	eventsChan <-chan domain.PriceUpdate
}

func NewPricesListener(notifier Notifier, eventsChan <-chan domain.PriceUpdate) *PricesListener {
	return &PricesListener{notifier: notifier, eventsChan: eventsChan}
}

func (l PricesListener) Start(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				l.log.Info("Stopping PricesListener")
				return

			case update, ok := <-l.eventsChan:
				if !ok {
					l.log.Info("Events channel closed, stopping PricesListener")
					return
				}

				l.log.Debug("Received price update", "update", update)
				l.notifier.Broadcast(update)
			}
		}
	}()

	return nil
}
