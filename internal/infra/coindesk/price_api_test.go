package coindesk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/tonytcb/crypto-pricing-api/internal/app/config"
	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

func TestPricingAPI_GetPrice(t *testing.T) {
	tests := []struct {
		name           string
		pair           domain.Pair
		serverResponse string
		serverStatus   interface{} // Can be int or []int for retry tests
		serverDelay    time.Duration
		maxAttempts    int
		expectedPrice  string
		expectError    bool
	}{
		{
			name:           "successful response",
			pair:           domain.NewPair(domain.BTC, domain.USD),
			serverResponse: `{"USD": 50000.25}`,
			serverStatus:   http.StatusOK,
			maxAttempts:    3,
			expectedPrice:  "50000.25",
			expectError:    false,
		},
		{
			name:           "server error with retry success",
			pair:           domain.NewPair(domain.BTC, domain.USD),
			serverResponse: `{"USD": 50000.25}`,
			serverStatus:   []int{http.StatusInternalServerError, http.StatusOK},
			maxAttempts:    3,
			expectedPrice:  "50000.25",
			expectError:    false,
		},
		{
			name:          "max retries exceeded",
			pair:          domain.NewPair(domain.BTC, domain.USD),
			serverStatus:  http.StatusInternalServerError,
			maxAttempts:   2,
			expectedPrice: "0",
			expectError:   true,
		},
		{
			name:           "invalid response format",
			pair:           domain.NewPair(domain.BTC, domain.USD),
			serverResponse: `invalid json`,
			serverStatus:   http.StatusOK,
			maxAttempts:    1,
			expectedPrice:  "0",
			expectError:    true,
		},
		{
			name:           "missing currency in response",
			pair:           domain.NewPair(domain.BTC, domain.USD),
			serverResponse: `{"EUR": 45000.75}`,
			serverStatus:   http.StatusOK,
			maxAttempts:    1,
			expectedPrice:  "0",
			expectError:    true,
		},
		{
			name:           "context cancellation",
			pair:           domain.NewPair(domain.BTC, domain.USD),
			serverResponse: `{"USD": 50000.25}`,
			serverStatus:   http.StatusOK,
			serverDelay:    100 * time.Millisecond,
			maxAttempts:    3,
			expectedPrice:  "0",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testAttempt := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/data/price", r.URL.Path)
				assert.Equal(t, string(tt.pair.From), r.URL.Query().Get("fsym"))
				assert.Equal(t, string(tt.pair.To), r.URL.Query().Get("tsyms"))

				if tt.serverDelay > 0 {
					time.Sleep(tt.serverDelay)
				}

				var status int
				if statusArray, ok := tt.serverStatus.([]int); ok {
					if testAttempt < len(statusArray) {
						status = statusArray[testAttempt]
					} else {
						status = statusArray[len(statusArray)-1]
					}
				} else {
					status = tt.serverStatus.(int)
				}
				testAttempt++

				w.WriteHeader(status)
				if tt.serverResponse != "" {
					_, _ = w.Write([]byte(tt.serverResponse))
				}
			}))
			defer server.Close()

			cfg := &config.Config{
				CoinDeskAPIURL:           server.URL + "/data/price",
				CoinDeskRetryMaxAttempts: tt.maxAttempts,
				CoinDeskRetryInitialWait: 10 * time.Millisecond,
				CoinDeskRetryMaxWait:     50 * time.Millisecond,
			}

			api := NewPricingAPI(server.Client(), cfg)

			ctx := context.Background()
			if tt.name == "context cancellation" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 50*time.Millisecond)
				defer cancel()
			}

			price, err := api.GetPrice(ctx, tt.pair)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			expectedPrice, err := decimal.NewFromString(tt.expectedPrice)
			assert.NoError(t, err)
			assert.True(t, expectedPrice.Equal(price), "Expected price %s, got %s", expectedPrice, price)
		})
	}
}
