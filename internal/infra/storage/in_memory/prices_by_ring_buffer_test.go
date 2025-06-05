package in_memory

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

func TestPricesByRingBufferStore(t *testing.T) {
	var (
		maxSize = 3
		now     = time.Now()
		BtcUsd  = domain.NewPair(domain.BTC, domain.USD)

		update1 = domain.PriceUpdate{Pair: BtcUsd, Price: decimal.NewFromFloat(50001.0), ReceivedAt: now.Add(-5 * time.Second)}
		update2 = domain.PriceUpdate{Pair: BtcUsd, Price: decimal.NewFromFloat(50002.0), ReceivedAt: now.Add(-4 * time.Second)}
		update3 = domain.PriceUpdate{Pair: BtcUsd, Price: decimal.NewFromFloat(50003.0), ReceivedAt: now.Add(-3 * time.Second)}
		update4 = domain.PriceUpdate{Pair: BtcUsd, Price: decimal.NewFromFloat(50004.0), ReceivedAt: now.Add(-2 * time.Second)}
		update5 = domain.PriceUpdate{Pair: BtcUsd, Price: decimal.NewFromFloat(50005.0), ReceivedAt: now.Add(-1 * time.Second)}
	)

	repo := NewPricesByRingBuffer(maxSize)
	assert.NotNil(t, repo)

	t.Run("Should return no items", func(t *testing.T) {
		history := repo.GetAll(BtcUsd)
		assert.Len(t, history, 0, "Expected empty history")
	})

	t.Run("Should insert 5 items and keep size as 3", func(t *testing.T) {
		repo.Store(update1)
		repo.Store(update2)
		repo.Store(update3)
		repo.Store(update4)
		repo.Store(update5)

		all := repo.GetAll(BtcUsd)
		assert.Len(t, all, maxSize, "Expected only the latest 3 items to be stored")
		assert.True(t, all[0].Price.Equal(update3.Price), "Expected first price to be %s, got %s", update3.Price, all[0].Price)
	})

	t.Run("Should return latest 2 inserted items", func(t *testing.T) {
		latest2Items := repo.GetSince(BtcUsd, update4.ReceivedAt)
		assert.Len(t, latest2Items, 2, "Expected only the latest 2 items to be stored")
	})

	t.Run("Should return all items", func(t *testing.T) {
		items := repo.GetSince(BtcUsd, now.Add(-10*time.Hour))
		assert.Len(t, items, maxSize, "Expected all items to be returned")
	})

	t.Run("Should return latest item inserted", func(t *testing.T) {
		latest, exists := repo.GetLatest(BtcUsd)
		assert.True(t, exists, "Expected price update to exist")
		assert.True(t, latest.Price.Equal(update5.Price), "Expected price to be %s, got %s", update5.Price, latest.Price)
		assert.True(t, latest.ReceivedAt.Equal(update5.ReceivedAt), "Expected receivedAt to be %v, got %v", now, latest.ReceivedAt)
	})

	t.Run("Should clean all items", func(t *testing.T) {
		repo.Clear()

		_, exists := repo.GetLatest(BtcUsd)
		assert.False(t, exists, "Expected no price update to exist after clear")

		history := repo.GetAll(BtcUsd)
		assert.Len(t, history, 0, "Expected empty history after clear")
	})
}
