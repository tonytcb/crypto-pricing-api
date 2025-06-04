package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type PriceUpdate struct {
	Pair       Pair
	Price      decimal.Decimal
	ReceivedAt time.Time
}
