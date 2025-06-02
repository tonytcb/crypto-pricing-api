package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type PriceUpdate struct {
	Price      decimal.Decimal
	ReceivedAt time.Time
}
