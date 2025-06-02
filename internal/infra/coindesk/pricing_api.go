package coindesk

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/tonytcb/crypto-pricing-api/internal/domain"
)

type PricingAPI struct {
}

func NewPricingAPI() *PricingAPI {
	return &PricingAPI{}
}

func (a PricingAPI) GetPrice(ctx context.Context, pair domain.Pair) (decimal.Decimal, error) {
	return decimal.Zero, nil
}
