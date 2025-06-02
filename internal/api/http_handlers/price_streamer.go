package http_handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
	"github.com/tonytcb/crypto-pricing-api/internal/infra/event_listener"
)

type EventListener interface {
	Listen(context.Context, domain.Pair) <-chan domain.PriceUpdate
}

type PriceAPI interface {
	GetPrice(ctx context.Context, pair domain.Pair) (decimal.Decimal, error)
}

type PriceStreamer struct {
	mu        sync.Mutex
	priceAPI  PriceAPI
	listeners map[EventListener]struct{}
}

func NewPriceStreamer(priceAPI PriceAPI) *PriceStreamer {
	return &PriceStreamer{
		listeners: make(map[EventListener]struct{}),
		priceAPI:  priceAPI,
	}
}

func (h *PriceStreamer) Stream(c *gin.Context) {
	w := c.Writer

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// @TODO Replace it by api input, if desired
	var btcUsd = domain.Pair{
		From: domain.BTC,
		To:   domain.USD,
	}

	// @TODO replace it by injected event listener
	var eventListener = event_listener.NewChannel(h.priceAPI, time.Second*3, 10)

	// @TODO handle since query param retrieving history price updates

	h.addEventListener(eventListener)
	defer h.removeEventListener(eventListener)

	var (
		ctx        = c.Request.Context()
		eventsChan = eventListener.Listen(ctx, btcUsd)
	)

	for {
		select {
		case <-ctx.Done():
			// @TODO debug log client disconnection
			return

		case update := <-eventsChan:
			res := PriceStreamResponse{
				Pair:       btcUsd.String(),
				Price:      update.Price.String(),
				ReceivedAt: update.ReceivedAt.Format(time.RFC3339),
			}
			data, err := json.Marshal(res)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Error marshalling response"))
				return
			}

			_, _ = w.Write([]byte("data: "))
			_, _ = w.Write(data)
			_, _ = w.Write([]byte("\n\n"))
			w.Flush()
		}
	}
}

func (h *PriceStreamer) addEventListener(c EventListener) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.listeners[c] = struct{}{}
}

func (h *PriceStreamer) removeEventListener(c EventListener) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.listeners, c)
}

type PriceStreamResponse struct {
	Pair       string `json:"pair"`
	Price      string `json:"price"`
	ReceivedAt string `json:"received_at"`
}
