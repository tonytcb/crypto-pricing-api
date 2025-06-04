package coindesk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/tonytcb/crypto-pricing-api/internal/app/config"
	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type PriceAPI struct {
	client HTTPClient
	config *config.Config
}

func NewPricingAPI(client HTTPClient, config *config.Config) *PriceAPI {
	return &PriceAPI{
		client: client,
		config: config,
	}
}

func (a PriceAPI) GetPrice(ctx context.Context, pair domain.Pair) (decimal.Decimal, error) {
	var price decimal.Decimal
	var err error

	// Retry mechanism with exponential backoff
	for attempt := 0; attempt < a.config.CoinDeskRetryMaxAttempts; attempt++ {
		price, err = a.fetchPrice(ctx, pair)
		if err == nil {
			return price, nil
		}

		if attempt == a.config.CoinDeskRetryMaxAttempts-1 {
			return decimal.Zero, errors.Wrap(err, "failed to fetch price after max retry attempts")
		}

		backoffDuration := a.calculateBackoff(attempt)

		timer := time.NewTimer(backoffDuration)
		select {
		case <-ctx.Done():
			timer.Stop()
			return decimal.Zero, ctx.Err()
		case <-timer.C:
			// Timer expired, continue to next attempt
		}
	}

	return decimal.Zero, errors.New("failed to fetch price")
}

func (a PriceAPI) fetchPrice(ctx context.Context, pair domain.Pair) (decimal.Decimal, error) {
	url := fmt.Sprintf("%s?fsym=%s&tsyms=%s",
		a.config.CoinDeskAPIURL,
		pair.From,
		pair.To)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return decimal.Zero, errors.Wrap(err, "failed to create request")
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return decimal.Zero, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return decimal.Zero, errors.Wrap(err, "failed to decode response")
	}

	priceValue, ok := response[string(pair.To)]
	if !ok {
		return decimal.Zero, errors.Errorf("price for %s not found in response", pair.To)
	}

	price, err := decimal.NewFromString(fmt.Sprintf("%v", priceValue))
	if err != nil {
		return decimal.Zero, errors.Wrap(err, "failed to convert price to decimal")
	}

	return price, nil
}

func (a PriceAPI) calculateBackoff(attempt int) time.Duration {
	// Calculate exponential backoff: initialWait * 2^attempt
	backoff := a.config.CoinDeskRetryInitialWait * time.Duration(1<<uint(attempt))

	if backoff > a.config.CoinDeskRetryMaxWait {
		backoff = a.config.CoinDeskRetryMaxWait
	}

	return backoff
}
