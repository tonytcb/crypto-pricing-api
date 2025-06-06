package event_listener

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) Broadcast(update domain.PriceUpdate) {
	m.Called(update)
}

func TestPricesListener_Start(t *testing.T) {
	t.Run("Should broadcast multiple updates", func(t *testing.T) {
		var (
			mockNotifier = new(MockNotifier)
			eventsChan   = make(chan domain.PriceUpdate, 10)
			listener     = NewPricesListener(mockNotifier, eventsChan)
		)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		pair := domain.NewPair(domain.BTC, domain.USD)
		update1 := domain.PriceUpdate{
			Pair:       pair,
			Price:      decimal.NewFromFloat(50000.0),
			ReceivedAt: time.Now(),
		}
		update2 := domain.PriceUpdate{
			Pair:       pair,
			Price:      decimal.NewFromFloat(51000.0),
			ReceivedAt: time.Now().Add(time.Second),
		}

		mockNotifier.On("Broadcast", update1).Return()
		mockNotifier.On("Broadcast", update2).Return()

		err := listener.Start(ctx)
		assert.NoError(t, err)

		// Send updates
		eventsChan <- update1
		eventsChan <- update2

		assert.Eventually(t, func() bool {
			return mockNotifier.AssertNumberOfCalls(t, "Broadcast", 2)
		}, 100*time.Millisecond, 10*time.Millisecond, "Notifier should be called twice")

		mockNotifier.AssertCalled(t, "Broadcast", update1)
		mockNotifier.AssertCalled(t, "Broadcast", update2)
	})
}
