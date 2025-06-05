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
)

// MockSseClientsManager is a mock implementation of SseClientsManager
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
		// Setup
		clientsManager := new(MockSseClientsManager)
		clientsManager.On("RegisterClient", mock.Anything).Return()

		cfg := &config.Config{SseClientsBufferSize: 10}
		handler := NewPriceStreamer(cfg, clientsManager)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/stream", nil)

		// Execute
		go handler.Stream(c)
		time.Sleep(10 * time.Millisecond) // Give time for goroutine to start

		// Asserts
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
		assert.Equal(t, "keep-alive", w.Header().Get("Connection"))
		assert.Equal(t, "chunked", w.Header().Get("Transfer-Encoding"))

		// No data should be in the response body yet as we're just setting up the stream
		assert.Empty(t, w.Body.String())

		clientsManager.AssertExpectations(t)
	})

	t.Run("Custom pair streaming", func(t *testing.T) {
		// Setup
		clientsManager := new(MockSseClientsManager)
		clientsManager.On("RegisterClient", mock.Anything).Return()

		cfg := &config.Config{SseClientsBufferSize: 10}
		handler := NewPriceStreamer(cfg, clientsManager)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/stream/ethusd", nil)
		c.Params = []gin.Param{{Key: "pair", Value: "ethusd"}}

		// Execute
		go handler.Stream(c)
		time.Sleep(10 * time.Millisecond) // Give time for goroutine to start

		// Asserts
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Body.String())

		clientsManager.AssertExpectations(t)
	})

	t.Run("Historical data streaming", func(t *testing.T) {
		// Setup
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
		c.Request = httptest.NewRequest(http.MethodGet, "/stream?since="+sinceTimestamp, nil)
		c.Request.URL.RawQuery = "since=" + sinceTimestamp

		// Execute
		go handler.Stream(c)
		time.Sleep(10 * time.Millisecond) // Give time for goroutine to start

		responseBody := w.Body.String()

		// Asserts
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, responseBody, `"price":"50000"`)
		assert.Contains(t, responseBody, `"price":"51000"`)
		assert.Contains(t, responseBody, `"pair":"BTCUSD"`)
		assert.Contains(t, responseBody, `"received_at"`)

		clientsManager.AssertExpectations(t)
	})
}
