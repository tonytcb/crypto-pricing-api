package event_listener

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

type PricingAPI interface {
	GetPrice(ctx context.Context, pair domain.Pair) (decimal.Decimal, error)
}

type Channel struct {
	pricingAPI   PricingAPI
	pullInterval time.Duration
	bufferSize   int
}

func NewChannel(pricingAPI PricingAPI, pullInterval time.Duration, bufferSize int) *Channel {
	return &Channel{
		pricingAPI:   pricingAPI,
		pullInterval: pullInterval,
		bufferSize:   bufferSize,
	}
}

func (c *Channel) Listen(ctx context.Context, pair domain.Pair) <-chan domain.PriceUpdate {
	ticker := time.NewTicker(c.pullInterval)

	ch := make(chan domain.PriceUpdate, c.bufferSize)

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
				price, err := c.pricingAPI.GetPrice(ctx, pair)
				if err != nil {
					// @TODO handle error, maybe log it
					continue
				}

				ch <- domain.PriceUpdate{
					Price:      price,
					ReceivedAt: time.Now().UTC(),
				}
			}
		}
	}()

	return ch
}
