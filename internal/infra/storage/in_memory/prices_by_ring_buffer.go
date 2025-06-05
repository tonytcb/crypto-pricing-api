package in_memory

import (
	"sync"
	"time"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

type PricesByRingBuffer struct {
	mutex          sync.RWMutex
	buffers        map[domain.Pair]*ringBuffer
	maxHistorySize int
}

func NewPricesByRingBuffer(maxHistorySize int) *PricesByRingBuffer {
	return &PricesByRingBuffer{
		buffers:        make(map[domain.Pair]*ringBuffer),
		maxHistorySize: maxHistorySize,
	}
}

func (r *PricesByRingBuffer) Store(priceUpdate domain.PriceUpdate) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	pair := priceUpdate.Pair

	buffer, exists := r.buffers[pair]
	if !exists {
		buffer = newRingBuffer(r.maxHistorySize)
		r.buffers[pair] = buffer
	}

	// Add the item to the buffer
	buffer.push(priceUpdate)
}

func (r *PricesByRingBuffer) GetLatest(pair domain.Pair) (domain.PriceUpdate, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	buffer, exists := r.buffers[pair]
	if !exists || buffer.size == 0 {
		return domain.PriceUpdate{}, false
	}

	// The latest item is at (tail-1) in the ring buffer
	latestIndex := (buffer.tail - 1 + buffer.capacity) % buffer.capacity
	return buffer.items[latestIndex], true
}

func (r *PricesByRingBuffer) GetAll(pair domain.Pair) []domain.PriceUpdate {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	buffer, exists := r.buffers[pair]
	if !exists || buffer.size == 0 {
		return []domain.PriceUpdate{}
	}

	return buffer.toSlice()
}

func (r *PricesByRingBuffer) GetSince(pair domain.Pair, since time.Time) []domain.PriceUpdate {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	buffer, exists := r.buffers[pair]
	if !exists || buffer.size == 0 {
		return []domain.PriceUpdate{}
	}

	allUpdates := buffer.toSlice()

	// Find the first update that is not before 'since'
	startIndex := len(allUpdates)
	for i, update := range allUpdates {
		if !update.ReceivedAt.Before(since) {
			startIndex = i
			break
		}
	}

	if startIndex >= len(allUpdates) {
		return []domain.PriceUpdate{}
	}

	result := make([]domain.PriceUpdate, len(allUpdates)-startIndex)
	copy(result, allUpdates[startIndex:])

	return result
}

// Clear removes all price updates from the repository.
func (r *PricesByRingBuffer) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.buffers = make(map[domain.Pair]*ringBuffer)
}
