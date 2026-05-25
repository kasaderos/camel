package portfolio

import (
	"context"
	"time"

	"github.com/kasaderos/camel/internal/model"
	"github.com/kasaderos/camel/pkg/alpaca"
)

type AgentRepository interface {
	UpdatePortfolioAgent(ctx context.Context, a model.PortfolioAgent) error
}

type MarketService interface {
	FetchBars(
		ctx context.Context,
		symbol string,
		start time.Time,
		end time.Time,
	) ([]alpaca.Bar, error)
}

type Exchanger interface {
	CreateMarketOrder(
		ctx context.Context,
		symbol string,
		qty float64,
		side string, // "buy" or "sell"
	) (*model.Order, error)
	FetchPrice(
		ctx context.Context,
		assetID string,
	) (float64, time.Time, error)
}
