package in_memory

import "github.com/tonytcb/crypto-pricing-api/internal/domain"

// ringBuffer is not thread-safe and depends on locks in the high-level repository
type ringBuffer struct {
	items    []domain.PriceUpdate
	size     int
	capacity int
	head     int
	tail     int
}

func newRingBuffer(capacity int) *ringBuffer {
	return &ringBuffer{
		items:    make([]domain.PriceUpdate, capacity),
		capacity: capacity,
		head:     0,
		tail:     0,
		size:     0,
	}
}

func (rb *ringBuffer) push(item domain.PriceUpdate) {
	rb.items[rb.tail] = item
	rb.tail = (rb.tail + 1) % rb.capacity

	if rb.size < rb.capacity {
		rb.size++
	} else {
		// If the buffer is full, move the head forward
		rb.head = (rb.head + 1) % rb.capacity
	}
}

func (rb *ringBuffer) toSlice() []domain.PriceUpdate {
	if rb.size == 0 {
		return []domain.PriceUpdate{}
	}

	result := make([]domain.PriceUpdate, rb.size)

	for i := 0; i < rb.size; i++ {
		idx := (rb.head + i) % rb.capacity
		result[i] = rb.items[idx]
	}

	return result
}
