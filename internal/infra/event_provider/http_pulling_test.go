package event_provider

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

type MockPriceAPI struct {
	mock.Mock
}

func (m *MockPriceAPI) GetPrice(ctx context.Context, pair domain.Pair) (decimal.Decimal, error) {
	args := m.Called(ctx, pair)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func TestHTTPPulling_Start(t *testing.T) {
	t.Run("Should pull prices at regular intervals", func(t *testing.T) {
		var (
			mockAPI      = new(MockPriceAPI)
			pullInterval = 50 * time.Millisecond
			bufferSize   = 10
			puller       = NewHTTPPulling(mockAPI, pullInterval, bufferSize)
		)

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		pair := domain.NewPair(domain.BTC, domain.USD)
		price := decimal.NewFromFloat(50000.0)

		mockAPI.On("GetPrice", mock.Anything, pair).Return(price, nil)

		ch, err := puller.Start(ctx, pair)

		assert.NoError(t, err)
		assert.NotNil(t, ch)

		var updates []domain.PriceUpdate
		for update := range ch {
			updates = append(updates, update)
		}
		assert.GreaterOrEqual(t, len(updates), 3, "Expected at least 3 price updates")

		for _, update := range updates {
			assert.Equal(t, pair, update.Pair)
			assert.True(t, price.Equal(update.Price))
		}

		mockAPI.AssertExpectations(t)
	})
}
