package in_memory

import (
	"sync"
	"time"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

type PricesBySliceRepo struct {
	prices         map[domain.Pair][]domain.PriceUpdate
	mutex          sync.RWMutex
	maxHistorySize int
}

func NewPricesBySliceRepo(maxHistorySize int) *PricesBySliceRepo {
	return &PricesBySliceRepo{
		prices:         make(map[domain.Pair][]domain.PriceUpdate),
		maxHistorySize: maxHistorySize,
	}
}

func (r *PricesBySliceRepo) Store(priceUpdate domain.PriceUpdate) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	pair := priceUpdate.Pair

	history, exists := r.prices[pair]
	if !exists {
		history = []domain.PriceUpdate{}
	}

	history = append(history, priceUpdate)

	if len(history) > r.maxHistorySize {
		history = history[len(history)-r.maxHistorySize:]
	}

	r.prices[pair] = history
}

func (r *PricesBySliceRepo) GetLatest(pair domain.Pair) (domain.PriceUpdate, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	history, exists := r.prices[pair]
	if !exists || len(history) == 0 {
		return domain.PriceUpdate{}, false
	}

	return history[len(history)-1], true
}

func (r *PricesBySliceRepo) GetAll(pair domain.Pair) []domain.PriceUpdate {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	history, exists := r.prices[pair]
	if !exists {
		return []domain.PriceUpdate{}
	}

	result := make([]domain.PriceUpdate, len(history))
	copy(result, history)

	return result
}

func (r *PricesBySliceRepo) GetSince(pair domain.Pair, since time.Time) []domain.PriceUpdate {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	history, exists := r.prices[pair]
	if !exists || len(history) == 0 {
		return []domain.PriceUpdate{}
	}

	startIndex := 0
	for i, update := range history {
		if !update.ReceivedAt.Before(since) {
			startIndex = i
			break
		}
		if i == len(history)-1 {
			startIndex = len(history)
		}
	}

	result := make([]domain.PriceUpdate, len(history)-startIndex)
	copy(result, history[startIndex:])

	return result
}

func (r *PricesBySliceRepo) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.prices = make(map[domain.Pair][]domain.PriceUpdate)
}
