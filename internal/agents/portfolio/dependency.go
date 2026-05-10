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
}
