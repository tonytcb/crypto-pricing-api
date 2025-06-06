package http_handlers

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/tonytcb/crypto-pricing-api/internal/app/config"
	"github.com/tonytcb/crypto-pricing-api/internal/domain"
	"github.com/tonytcb/crypto-pricing-api/internal/infra/sse"
	"github.com/tonytcb/crypto-pricing-api/test/mocks"
)

type MockSseClientsManager struct {
	mock.Mock
}

func (m *MockSseClientsManager) RegisterClient(client *sse.Client) {
	m.Called(client)
}

func (m *MockSseClientsManager) UnregisterClient(client *sse.Client) {
	m.Called(client)
}

func (m *MockSseClientsManager) GetHistory(pair domain.Pair, since time.Time) []domain.PriceUpdate {
	args := m.Called(pair, since)
	return args.Get(0).([]domain.PriceUpdate)
}

func TestPriceStreamer_Stream(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Default BTCUSD pair streaming", func(t *testing.T) {
		clientsManager := new(MockSseClientsManager)
		clientsManager.On("RegisterClient", mock.Anything).Return()

		cfg := &config.Config{SseClientsBufferSize: 10}
		handler := NewPriceStreamer(cfg, clientsManager)

		w := mocks.NewThreadSafeRecorder()
		c, _ := gin.CreateTestContext(w.Recorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/stream", nil)

		go handler.Stream(c)

		assert.Eventually(t, func() bool {
			return w.Code() == http.StatusOK
		}, 100*time.Millisecond, 10*time.Millisecond, "Status code should be OK")

		assert.Empty(t, w.BodyString())

		clientsManager.AssertExpectations(t)
	})

	t.Run("Custom pair streaming", func(t *testing.T) {
		clientsManager := new(MockSseClientsManager)
		clientsManager.On("RegisterClient", mock.Anything).Return()

		cfg := &config.Config{SseClientsBufferSize: 10}
		handler := NewPriceStreamer(cfg, clientsManager)

		w := mocks.NewThreadSafeRecorder()
		c, _ := gin.CreateTestContext(w.Recorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/stream/ETHUSD", nil)
		c.Params = []gin.Param{{Key: "pair", Value: "ETHUSD"}}

		go handler.Stream(c)

		assert.Eventually(t, func() bool {
			return w.Code() == http.StatusOK
		}, 100*time.Millisecond, 10*time.Millisecond, "Status code should be OK")

		assert.Empty(t, w.BodyString())

		clientsManager.AssertExpectations(t)
	})

	t.Run("Historical data streaming", func(t *testing.T) {
		now := time.Now()
		sinceTime := now.Add(-1 * time.Hour)
		sinceTimestamp := strconv.FormatInt(sinceTime.Unix(), 10)

		btcUsd := domain.Pair{From: domain.BTC, To: domain.USD}

		history := []domain.PriceUpdate{
			{Pair: btcUsd, Price: decimal.NewFromFloat(50000), ReceivedAt: now.Add(-30 * time.Second)},
			{Pair: btcUsd, Price: decimal.NewFromFloat(51000), ReceivedAt: now.Add(-15 * time.Second)},
		}

		clientsManager := new(MockSseClientsManager)
		clientsManager.On("RegisterClient", mock.Anything).Return()
		clientsManager.On("GetHistory", btcUsd, mock.Anything).Return(history)

		cfg := &config.Config{SseClientsBufferSize: 10}
		handler := NewPriceStreamer(cfg, clientsManager)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/?since="+sinceTimestamp, nil)
		c.Request.URL.RawQuery = "since=" + sinceTimestamp

		// Create a channel to signal when we're done with the test
		done := make(chan struct{})

		go func() {
			go func() {
				time.Sleep(50 * time.Millisecond)
				close(done)
			}()
			handler.Stream(c)
		}()

		// Wait for the test to complete
		<-done

		clientsManager.AssertExpectations(t)
	})
}
