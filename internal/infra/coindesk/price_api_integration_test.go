//go:build integration
// +build integration

package coindesk

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tonytcb/crypto-pricing-api/internal/app/config"
	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

func TestPriceAPI_Integration(t *testing.T) {
	cfg := &config.Config{
		CoinDeskAPIURL:           "https://min-api.cryptocompare.com/data/price",
		CoinDeskRetryMaxAttempts: 3,
		CoinDeskClientTimeout:    3 * time.Second,
		CoinDeskRetryTimeout:     100 * time.Millisecond,
		CoinDeskRetryInitialWait: 100 * time.Millisecond,
		CoinDeskRetryMaxWait:     1 * time.Second,
	}

	httpClient := &http.Client{
		Timeout: cfg.CoinDeskClientTimeout,
	}

	priceAPI := NewPricingAPI(httpClient, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pair := domain.NewPair(domain.BTC, domain.USD)

	price, err := priceAPI.GetPrice(ctx, pair)
	require.NoError(t, err)
	assert.True(t, price.GreaterThan(decimal.Zero), "Expected price to be greater than zero, got %s", price)

	t.Logf("Current price for %s/%s: %s", pair.From, pair.To, price)
}
