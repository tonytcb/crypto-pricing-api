package sse

import (
	"log/slog"
	"sync"
	"time"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

type PricesRepository interface {
	Store(priceUpdate domain.PriceUpdate)
	GetSince(pair domain.Pair, since time.Time) []domain.PriceUpdate
}

type Hub struct {
	mu              sync.RWMutex
	pricesRepo      PricesRepository
	cleanUpInterval time.Duration
	log             *slog.Logger
	clients         map[*Client]struct{}
	register        chan *Client
	unregister      chan *Client
	broadcast       chan domain.PriceUpdate
	done            chan struct{}
}

func NewHub(pricesRepo PricesRepository, cleanUpInterval time.Duration) *Hub {
	return &Hub{
		pricesRepo:      pricesRepo,
		cleanUpInterval: cleanUpInterval,
		log:             slog.Default(),
		clients:         make(map[*Client]struct{}),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		broadcast:       make(chan domain.PriceUpdate),
		done:            make(chan struct{}),
	}
}

func (h *Hub) Start() {
	cleanupTicker := time.NewTicker(h.cleanUpInterval)
	defer cleanupTicker.Stop()

	for {
		select {
		case client := <-h.register:
			h.addClient(client)

		case client := <-h.unregister:
			h.removeClient(client)

		case update := <-h.broadcast:
			h.broadcastUpdate(update)
			h.pricesRepo.Store(update)

		case <-cleanupTicker.C:
			h.cleanupDisconnectedClients()

		case <-h.done:
			h.closeAllClients()
			return
		}
	}
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

func (h *Hub) Broadcast(update domain.PriceUpdate) {
	select {
	case h.broadcast <- update:
	default:
		h.log.Error("Broadcast channel full, dropping update")
	}
}

func (h *Hub) GetHistory(pair domain.Pair, since time.Time) []domain.PriceUpdate {
	return h.pricesRepo.GetSince(pair, since)
}

func (h *Hub) Stop() {
	close(h.done)
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) addClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = struct{}{}

	h.log.Info("Client connected", "client_id", client.ID())
}

func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[client]; exists {
		delete(h.clients, client)
		client.Close()
	}

	h.log.Info("Client disconnected", "client_id", client.ID())
}

func (h *Hub) broadcastUpdate(update domain.PriceUpdate) {
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if err := client.Send(update); err != nil {
			h.log.Error("Failed to send update to client", "client_id", client.ID(), "error", err.Error())
		}
	}
}

func (h *Hub) cleanupDisconnectedClients() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		if client.IsClosed() {
			delete(h.clients, client)
			h.log.Info("Cleaning up disconnected client", "client_id", client.ID())
		}
	}
}

func (h *Hub) closeAllClients() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		client.Close()
	}

	h.clients = make(map[*Client]struct{})
}
