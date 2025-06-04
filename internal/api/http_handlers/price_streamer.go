package http_handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/tonytcb/crypto-pricing-api/internal/app/config"
	"github.com/tonytcb/crypto-pricing-api/internal/domain"
	"github.com/tonytcb/crypto-pricing-api/internal/infra/sse"
)

type SseClientsManager interface {
	RegisterClient(client *sse.Client)
	UnregisterClient(client *sse.Client)
	GetHistory(pair domain.Pair, since time.Time) []domain.PriceUpdate
}

type PriceStreamer struct {
	log            *slog.Logger
	cfg            *config.Config
	clientsManager SseClientsManager
}

func NewPriceStreamer(cfg *config.Config, clientsManager SseClientsManager) *PriceStreamer {
	return &PriceStreamer{
		log:            slog.Default(),
		cfg:            cfg,
		clientsManager: clientsManager,
	}
}

func (h *PriceStreamer) Stream(c *gin.Context) {
	w := c.Writer

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// @TODO handle since query param retrieving history price updates

	clientID := uuid.New().String()
	client, err := sse.NewClient(clientID, c.Writer, h.cfg.SseClientsBufferSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	var pair = domain.Pair{
		From: domain.BTC,
		To:   domain.USD,
	}

	if pairParam := c.Param("pair"); pairParam != "" {
		v, err := domain.NewPairFromString(pairParam)
		if err != nil {
			h.log.Error("Invalid pair parameter", "pair", pairParam, "error", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pair parameter"})
			return
		}

		pair = v
	}

	h.clientsManager.RegisterClient(client)
	defer h.clientsManager.UnregisterClient(client)

	go client.Listen(pair)

	<-c.Request.Context().Done() // wait until a client is connected
}
