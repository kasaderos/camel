package portfolio

import (
	"context"

	"github.com/kasaderos/camel/internal/model"
)

type AgentRepository interface {
	Fetch(ctx context.Context, id string) (model.PortfolioAgent, error)
	Create(ctx context.Context, assets []model.AssetAgent) (model.PortfolioAgent, error)
}

type AssetAgent interface {
	FetchInfo(context.Context) model.AssetAgent
	FetchState(context.Context) model.State
	UpdateState(context.Context) error
	BuyAsset(ctx context.Context, sum float64) error
	SellAsset(ctx context.Context, qty float64) error
	Withdraw(ctx context.Context, sum float64) (model.AssetAgent, error)
	WithdrawWithBuy(ctx context.Context, sum float64) (model.AssetAgent, error)
	WithdrawWithSell(ctx context.Context, sum float64) (model.AssetAgent, error)
	Deposit(ctx context.Context, amount float64) (model.AssetAgent, error)
	DepositWithBuy(ctx context.Context, sum float64) (model.AssetAgent, error)
	FetchPrice(ctx context.Context) (float64, error)
	FetchTotalSum(ctx context.Context) (float64, error)
	ClosePosition(ctx context.Context) (model.AssetAgent, error)
}
