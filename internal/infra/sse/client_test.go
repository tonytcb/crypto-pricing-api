package sse

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

func TestNewClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		client, err := NewClient("test-client-1", w, 10)

		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "test-client-1", client.ID())
		assert.False(t, client.IsClosed())
	})

	t.Run("Error - Writer does not support flushing", func(t *testing.T) {
		mockWriter := &mockResponseWriter{}
		client, err := NewClient("test-client-1", mockWriter, 10)

		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "does not support streaming")
	})
}

func TestClient_ID(t *testing.T) {
	w := httptest.NewRecorder()
	client, err := NewClient("test-client-id", w, 10)
	require.NoError(t, err)

	assert.Equal(t, "test-client-id", client.ID())
}

func TestClient_Send(t *testing.T) {
	slog.SetDefault(newNoopLogger())

	t.Run("Buffer full", func(t *testing.T) {
		w := httptest.NewRecorder()
		client, err := NewClient("test-client-1", w, 1) // Buffer size of 1
		require.NoError(t, err)

		btcUsd := domain.NewPair(domain.BTC, domain.USD)
		update1 := domain.PriceUpdate{
			Pair:       btcUsd,
			Price:      decimal.NewFromFloat(50000),
			ReceivedAt: time.Now(),
		}
		update2 := domain.PriceUpdate{
			Pair:       btcUsd,
			Price:      decimal.NewFromFloat(51000),
			ReceivedAt: time.Now().Add(time.Second),
		}

		assert.NoError(t, client.Send(update1))

		err = client.Send(update2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client buffer is full")
	})
}

func TestClient_Listen(t *testing.T) {
	t.Run("Receive updates for registered pair", func(t *testing.T) {
		w := httptest.NewRecorder()
		client, err := NewClient("test-client-1", w, 10)
		require.NoError(t, err)

		btcUsd := domain.NewPair(domain.BTC, domain.USD)
		update := domain.PriceUpdate{
			Pair:       btcUsd,
			Price:      decimal.NewFromFloat(50000),
			ReceivedAt: time.Now(),
		}

		go client.Listen(btcUsd)

		require.NoError(t, client.Send(update))

		// Check the response
		assert.Eventually(t, func() bool {
			response := w.Body.String()
			if !strings.Contains(response, "data: ") {
				return false
			}

			// Parse the JSON from the response
			jsonStr := strings.TrimPrefix(strings.TrimSuffix(response, "\n\n"), "data: ")
			var respObj PriceStreamResponse
			err := json.Unmarshal([]byte(jsonStr), &respObj)
			if err != nil {
				return false
			}

			return respObj.Pair == "BTCUSD" &&
				respObj.Price == "50000" &&
				respObj.ReceivedAt != ""
		}, 100*time.Millisecond, 10*time.Millisecond, "Client should receive the update for registered pair")

		// Close the client to stop the listener
		client.Close()
	})

	t.Run("Ignore updates for unregistered pair", func(t *testing.T) {
		w := httptest.NewRecorder()
		client, err := NewClient("test-client-1", w, 10)
		require.NoError(t, err)

		btcUsd := domain.NewPair(domain.BTC, domain.USD)
		ethUsd := domain.NewPair(domain.ETH, domain.USD)
		update := domain.PriceUpdate{
			Pair:       ethUsd,
			Price:      decimal.NewFromFloat(0.00002),
			ReceivedAt: time.Now(),
		}

		go client.Listen(btcUsd)

		require.NoError(t, client.Send(update))

		// Check the response - should remain empty as the update should be ignored
		assert.Eventually(t, func() bool {
			return w.Body.String() == ""
		}, 100*time.Millisecond, 10*time.Millisecond, "Response should remain empty for unregistered pair")

		// Close the client to stop the listener
		client.Close()
	})

	t.Run("Stop listening when client is closed", func(t *testing.T) {
		w := httptest.NewRecorder()
		client, err := NewClient("test-client-1", w, 10)
		require.NoError(t, err)

		btcUsd := domain.NewPair(domain.BTC, domain.USD)
		update := domain.PriceUpdate{
			Pair:       btcUsd,
			Price:      decimal.NewFromFloat(50000),
			ReceivedAt: time.Now(),
		}

		go client.Listen(btcUsd)

		require.NoError(t, client.Send(update))

		var mu sync.Mutex

		assert.Eventually(t, func() bool {
			mu.Lock()
			response := w.Body.String()
			mu.Unlock()
			return strings.Contains(response, "data: ")
		}, 100*time.Millisecond, 10*time.Millisecond, "Client should receive the update")

		w.Body.Reset()

		client.Close()

		// Send another update that should not be processed
		update.Price = decimal.NewFromFloat(55000)
		_ = client.Send(update) // Ignoring error as the client might be closed

		// Verify the client is closed, and no more updates are processed
		assert.Eventually(t, func() bool {
			mu.Lock()
			response := w.Body.String()
			mu.Unlock()
			return client.IsClosed() && response == ""
		}, 100*time.Millisecond, 10*time.Millisecond, "Client should be closed and not process updates")
	})
}

// mockResponseWriter is a mock http.ResponseWriter that doesn't implement http.Flusher
type mockResponseWriter struct {
	headers map[string][]string
	body    []byte
	status  int
}

func (m *mockResponseWriter) Header() http.Header {
	if m.headers == nil {
		m.headers = make(map[string][]string)
	}
	return m.headers
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	if m.body == nil {
		m.body = make([]byte, 0)
	}
	m.body = append(m.body, b...)
	return len(b), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}

func newNoopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
