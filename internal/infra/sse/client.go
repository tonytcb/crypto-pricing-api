package sse

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

type Client struct {
	mu      sync.RWMutex
	log     *slog.Logger
	id      string
	ch      chan domain.PriceUpdate
	writer  http.ResponseWriter
	flusher http.Flusher
	done    chan struct{}
}

type PriceStreamResponse struct {
	Pair       string `json:"pair"`
	Price      string `json:"price"`
	ReceivedAt string `json:"received_at"`
}

func NewClient(id string, w http.ResponseWriter, bufferSize int) (*Client, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("response writer does not support streaming")
	}

	client := &Client{
		log:     slog.Default(),
		id:      id,
		ch:      make(chan domain.PriceUpdate, bufferSize),
		writer:  w,
		flusher: flusher,
		done:    make(chan struct{}),
	}

	return client, nil
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) Send(update domain.PriceUpdate) error {
	select {
	case c.ch <- update:
		return nil
	default:
		return errors.New("client buffer is full")
	}
}

func (c *Client) Listen(pair domain.Pair) {
	for {
		select {
		case update := <-c.ch:
			if update.Pair != pair {
				c.log.Debug("Ignoring update for unregistered pair", "pair", update.Pair.String(), "client_id", c.id)
				continue
			}

			if err := c.writeUpdate(update); err != nil {
				c.log.Error("Failed to write update to client", "error", err.Error())
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *Client) writeUpdate(update domain.PriceUpdate) error {
	response := PriceStreamResponse{
		Pair:       update.Pair.String(),
		Price:      update.Price.String(),
		ReceivedAt: update.ReceivedAt.Format(time.RFC3339),
	}

	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	_, err = fmt.Fprintf(c.writer, "data: %s\n\n", data)
	if err != nil {
		return errors.Wrap(err, "failed to stream update to client")
	}

	c.flusher.Flush()

	return nil
}

func (c *Client) Close() {
	select {
	case <-c.done:
		return
	default:
		close(c.done)
	}
}

func (c *Client) IsClosed() bool {
	select {
	case <-c.done:
		return true
	default:
		return false
	}
}
