package sse

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

func TestHub_Start(t *testing.T) {
	pricesRepo := new(MockPricesRepository)
	hub := NewHub(pricesRepo, 100*time.Millisecond)

	w1 := httptest.NewRecorder()
	client1, err := NewClient("test-client-1", w1, 10)
	assert.NoError(t, err)

	w2 := httptest.NewRecorder()
	client2, err := NewClient("test-client-2", w2, 10)
	assert.NoError(t, err)

	btcUsd := domain.NewPair(domain.BTC, domain.USD)
	update := domain.PriceUpdate{
		Pair:       btcUsd,
		Price:      decimal.NewFromFloat(50000),
		ReceivedAt: time.Now(),
	}

	pricesRepo.On("Store", mock.Anything).Return()

	go hub.Start()
	defer hub.Stop()

	assert.Equal(t, 0, hub.ClientCount())

	hub.RegisterClient(client1)
	hub.RegisterClient(client2)

	assert.Eventually(t, func() bool {
		return hub.ClientCount() == 2
	}, 100*time.Millisecond, 10*time.Millisecond, "Both clients should be registered")

	go client1.Listen(btcUsd)
	go client2.Listen(btcUsd)
	time.Sleep(10 * time.Millisecond) // Give time for goroutines to start

	hub.Broadcast(update)

	assert.Eventually(t, func() bool {
		response1 := w1.Body.String()
		response2 := w2.Body.String()
		return strings.Contains(response1, "50000") && strings.Contains(response2, "50000")
	}, 100*time.Millisecond, 10*time.Millisecond, "Clients should receive the broadcast")

	pricesRepo.AssertCalled(t, "Store", mock.Anything)

	hub.UnregisterClient(client1)

	assert.Eventually(t, func() bool {
		return hub.ClientCount() == 1
	}, 100*time.Millisecond, 10*time.Millisecond, "Client1 should be unregistered")

	client2.Close()

	assert.Eventually(t, func() bool {
		return hub.ClientCount() == 0
	}, 200*time.Millisecond, 10*time.Millisecond, "Disconnected client should be cleaned up")
}

func TestHub_CleanupDisconnectedClients(t *testing.T) {
	pricesRepo := new(MockPricesRepository)
	hub := NewHub(pricesRepo, time.Minute)

	w1 := httptest.NewRecorder()
	client1, err := NewClient("test-client-1", w1, 10)
	assert.NoError(t, err)

	w2 := httptest.NewRecorder()
	client2, err := NewClient("test-client-2", w2, 10)
	assert.NoError(t, err)
	client2.Close() // Mark as closed

	hub.addClient(client1)
	hub.addClient(client2)
	assert.Equal(t, 2, hub.ClientCount())

	hub.cleanupDisconnectedClients()
	assert.Equal(t, 1, hub.ClientCount())
}

func TestHub_CloseAllClients(t *testing.T) {
	pricesRepo := new(MockPricesRepository)
	hub := NewHub(pricesRepo, time.Minute)

	w1 := httptest.NewRecorder()
	client1, err := NewClient("test-client-1", w1, 10)
	assert.NoError(t, err)

	w2 := httptest.NewRecorder()
	client2, err := NewClient("test-client-2", w2, 10)
	assert.NoError(t, err)

	hub.addClient(client1)
	hub.addClient(client2)
	assert.Equal(t, 2, hub.ClientCount())

	hub.closeAllClients()
	assert.Equal(t, 0, hub.ClientCount())

	assert.True(t, client1.IsClosed())
	assert.True(t, client2.IsClosed())
}

func TestHub_ClientCount(t *testing.T) {
	pricesRepo := new(MockPricesRepository)
	hub := NewHub(pricesRepo, time.Minute)

	assert.Equal(t, 0, hub.ClientCount())

	w1 := httptest.NewRecorder()
	client1, err := NewClient("test-client-1", w1, 10)
	assert.NoError(t, err)

	w2 := httptest.NewRecorder()
	client2, err := NewClient("test-client-2", w2, 10)
	assert.NoError(t, err)

	hub.addClient(client1)
	assert.Equal(t, 1, hub.ClientCount())

	hub.addClient(client2)
	assert.Equal(t, 2, hub.ClientCount())

	hub.removeClient(client1)
	assert.Equal(t, 1, hub.ClientCount())
}

type MockPricesRepository struct {
	mock.Mock
}

func (m *MockPricesRepository) Store(priceUpdate domain.PriceUpdate) {
	m.Called(priceUpdate)
}

func (m *MockPricesRepository) GetSince(pair domain.Pair, since time.Time) []domain.PriceUpdate {
	args := m.Called(pair, since)
	return args.Get(0).([]domain.PriceUpdate)
}
