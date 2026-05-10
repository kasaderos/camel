package asset

import (
	"context"
	"time"

	"github.com/kasaderos/camel/internal/model"
)

type AgentRepository interface {
	CreateAgent(ctx context.Context, agent *model.AssetAgent) error
	FetchInfo(ctx context.Context, id string) (model.AssetAgent, error)
	Withdraw(ctx context.Context, id string, q float64) error
	Deposit(ctx context.Context, id string, q float64) error
	UpdateState(ctx context.Context, id string, state model.State) error
}

type MarketService interface {
	FetchBars(
		ctx context.Context,
		symbol string,
		start time.Time,
		end time.Time,
	) ([]model.Bar, error)
}

type AssetAgent interface {
	FetchInfo(context.Context) model.AssetAgent
	FetchState(context.Context) model.State
	UpdateState(context.Context) error
}
