package in_memory

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

func TestPricesBySliceRepoStore(t *testing.T) {
	repo := NewPricesBySliceRepo(3)
	assert.NotNil(t, repo)
	pair := domain.NewPair(domain.BTC, domain.USD)

	now := time.Now()
	price := decimal.NewFromFloat(50000.0)
	update := domain.PriceUpdate{Pair: pair, Price: price, ReceivedAt: now}

	repo.Store(update)

	latest, exists := repo.GetLatest(pair)
	assert.True(t, exists, "Expected price update to exist")
	assert.True(t, latest.Price.Equal(price), "Expected price to be %s, got %s", price, latest.Price)
	assert.True(t, latest.ReceivedAt.Equal(now), "Expected receivedAt to be %v, got %v", now, latest.ReceivedAt)
}

func TestPricesBySliceRepoGetLatest(t *testing.T) {
	repo := NewPricesBySliceRepo(3)
	pair := domain.NewPair(domain.BTC, domain.USD)

	_, exists := repo.GetLatest(pair)
	assert.False(t, exists, "Expected no price update to exist")

	now := time.Now()
	price1 := decimal.NewFromFloat(50000.0)
	update1 := domain.PriceUpdate{Pair: pair, Price: price1, ReceivedAt: now.Add(-2 * time.Hour)}

	price2 := decimal.NewFromFloat(51000.0)
	update2 := domain.PriceUpdate{Pair: pair, Price: price2, ReceivedAt: now.Add(-1 * time.Hour)}

	price3 := decimal.NewFromFloat(52000.0)
	update3 := domain.PriceUpdate{Pair: pair, Price: price3, ReceivedAt: now}

	repo.Store(update1)
	repo.Store(update2)
	repo.Store(update3)

	latest, exists := repo.GetLatest(pair)
	assert.True(t, exists, "Expected price update to exist")
	assert.True(t, latest.Price.Equal(price3), "Expected latest price to be %s, got %s", price3, latest.Price)
}

func TestPricesBySliceRepoGetAll(t *testing.T) {
	repo := NewPricesBySliceRepo(3)
	pair := domain.NewPair(domain.BTC, domain.USD)

	// Test with no data
	history := repo.GetAll(pair)
	assert.Len(t, history, 0, "Expected empty history")

	// Store multiple updates
	now := time.Now()
	price1 := decimal.NewFromFloat(50000.0)
	update1 := domain.PriceUpdate{Pair: pair, Price: price1, ReceivedAt: now.Add(-2 * time.Hour)}

	price2 := decimal.NewFromFloat(51000.0)
	update2 := domain.PriceUpdate{Pair: pair, Price: price2, ReceivedAt: now.Add(-1 * time.Hour)}

	repo.Store(update1)
	repo.Store(update2)

	// Verify history
	history = repo.GetAll(pair)
	assert.Len(t, history, 2, "Expected 2 items in history")

	// Verify the history is returned in chronological order
	assert.True(t, history[0].Price.Equal(price1), "Expected first price to be %s, got %s", price1, history[0].Price)
	assert.True(t, history[1].Price.Equal(price2), "Expected second price to be %s, got %s", price2, history[1].Price)
}

func TestPricesBySliceRepoHistoryCleanup(t *testing.T) {
	repo := NewPricesBySliceRepo(2) // Only keep 2 items
	pair := domain.NewPair(domain.BTC, domain.USD)

	// Store 3 updates
	now := time.Now()
	price1 := decimal.NewFromFloat(50000.0)
	update1 := domain.PriceUpdate{Pair: pair, Price: price1, ReceivedAt: now.Add(-2 * time.Hour)}

	price2 := decimal.NewFromFloat(51000.0)
	update2 := domain.PriceUpdate{Pair: pair, Price: price2, ReceivedAt: now.Add(-1 * time.Hour)}

	price3 := decimal.NewFromFloat(52000.0)
	update3 := domain.PriceUpdate{Pair: pair, Price: price3, ReceivedAt: now}

	repo.Store(update1)
	repo.Store(update2)
	repo.Store(update3)

	// Verify only the latest 2 are kept
	history := repo.GetAll(pair)
	assert.Len(t, history, 2, "Expected 2 items in history")

	// Verify the oldest was removed
	assert.True(t, history[0].Price.Equal(price2), "Expected first price to be %s, got %s", price2, history[0].Price)
	assert.True(t, history[1].Price.Equal(price3), "Expected second price to be %s, got %s", price3, history[1].Price)
}

func TestPricesBySliceRepoGetSince(t *testing.T) {
	repo := NewPricesBySliceRepo(5)
	pair := domain.NewPair(domain.BTC, domain.USD)

	// Test with no data
	history := repo.GetSince(pair, time.Now())
	assert.Len(t, history, 0, "Expected empty history with no data")

	// Store multiple updates
	now := time.Now()
	times := []time.Time{
		now.Add(-4 * time.Hour),
		now.Add(-3 * time.Hour),
		now.Add(-2 * time.Hour),
		now.Add(-1 * time.Hour),
		now,
	}

	prices := []decimal.Decimal{
		decimal.NewFromFloat(50000.0),
		decimal.NewFromFloat(51000.0),
		decimal.NewFromFloat(52000.0),
		decimal.NewFromFloat(53000.0),
		decimal.NewFromFloat(54000.0),
	}

	for i := 0; i < 5; i++ {
		update := domain.PriceUpdate{Pair: pair, Price: prices[i], ReceivedAt: times[i]}
		repo.Store(update)
	}

	// Test 1: Get history from a timestamp before all updates
	// Should return all updates
	beforeAll := now.Add(-5 * time.Hour)
	history = repo.GetSince(pair, beforeAll)
	assert.Len(t, history, 5, "Expected 5 items when starting before all updates")
	for i := 0; i < 5; i++ {
		assert.True(t, history[i].Price.Equal(prices[i]), "Expected price at index %d to be %s, got %s", i, prices[i], history[i].Price)
	}

	// Test 2: Get history from a timestamp after all updates
	// Should return empty slice
	afterAll := now.Add(1 * time.Hour)
	history = repo.GetSince(pair, afterAll)
	assert.Len(t, history, 0, "Expected 0 items when starting after all updates")

	// Test 3: Get history from a timestamp in the middle
	// Should return updates from that point onwards
	middleTime := times[2] // Exactly at the third update
	history = repo.GetSince(pair, middleTime)
	assert.Len(t, history, 3, "Expected 3 items when starting from the middle")
	for i := 0; i < 3; i++ {
		expectedPrice := prices[i+2]
		assert.True(t, history[i].Price.Equal(expectedPrice), "Expected price at index %d to be %s, got %s", i, expectedPrice, history[i].Price)
	}

	// Test 4: Get history from a timestamp between updates
	// Should return updates from the next point onwards
	betweenTime := times[1].Add(30 * time.Minute) // Between second and third update
	history = repo.GetSince(pair, betweenTime)
	assert.Len(t, history, 3, "Expected 3 items when starting between updates")
	for i := 0; i < 3; i++ {
		expectedPrice := prices[i+2]
		assert.True(t, history[i].Price.Equal(expectedPrice), "Expected price at index %d to be %s, got %s", i, expectedPrice, history[i].Price)
	}
}

func TestPricesBySliceRepoClear(t *testing.T) {
	repo := NewPricesBySliceRepo(3)
	pair := domain.NewPair(domain.BTC, domain.USD)

	now := time.Now()
	price := decimal.NewFromFloat(50000.0)
	update := domain.PriceUpdate{Pair: pair, Price: price, ReceivedAt: now}

	repo.Store(update)
	repo.Clear()

	_, exists := repo.GetLatest(pair)
	assert.False(t, exists, "Expected no price update to exist after clear")

	history := repo.GetAll(pair)
	assert.Len(t, history, 0, "Expected empty history after clear")
}
