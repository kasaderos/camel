package portfolio

import (
	"context"
	"time"

	"github.com/kasaderos/camel/internal/model"
)

type Exchanger interface {
	FetchPrice(
		ctx context.Context,
		assetID string,
	) (float64, time.Time, error)
	CreateMarketOrder(
		ctx context.Context,
		symbol string,
		qty float64,
		side string, // "buy" or "sell"
	) (*model.Order, error)
}

type PortfolioRepository interface {
	CreatePortfolio(ctx context.Context, p model.Portfolio) error
	FetchPortfolio(ctx context.Context, id string) (model.Portfolio, error)
	UpdatePortfolio(ctx context.Context, p model.Portfolio) error

	CreatePortfolioAgent(
		ctx context.Context,
		agent model.PortfolioAgent,
	) error
	UpdatePortfolioAgent(
		ctx context.Context,
		agent model.PortfolioAgent,
	) error
	FetchPortfolioAgents(
		ctx context.Context,
		portfolioID string,
	) ([]model.PortfolioAgent, error)
}
