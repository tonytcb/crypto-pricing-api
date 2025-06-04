package event_provider

import (
	"context"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

type PriceAPI interface {
	GetPrice(ctx context.Context, pair domain.Pair) (decimal.Decimal, error)
}

type HTTPPulling struct {
	log          *slog.Logger
	priceAPI     PriceAPI
	pullInterval time.Duration
	bufferSize   int
}

func NewHTTPPulling(priceAPI PriceAPI, pullInterval time.Duration, bufferSize int) *HTTPPulling {
	return &HTTPPulling{
		log:          slog.Default(),
		priceAPI:     priceAPI,
		pullInterval: pullInterval,
		bufferSize:   bufferSize,
	}
}

func (p *HTTPPulling) Start(ctx context.Context, pair domain.Pair) (<-chan domain.PriceUpdate, error) {
	ticker := time.NewTicker(p.pullInterval)

	ch := make(chan domain.PriceUpdate, p.bufferSize)

	go func() {
		defer func() {
			ticker.Stop()
			close(ch)
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				price, err := p.priceAPI.GetPrice(ctx, pair)
				if err != nil {
					p.log.Error("Error getting price", "error", err.Error())
					continue
				}

				ch <- domain.PriceUpdate{
					Pair:       pair,
					Price:      price,
					ReceivedAt: time.Now().UTC(),
				}
			}
		}
	}()

	return ch, nil
}

func (p *HTTPPulling) Stop() {
	// todo
}
