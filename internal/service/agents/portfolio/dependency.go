package portfolio

import (
	"context"

	"github.com/kasaderos/camel/internal/model"
)

type AgentRepository interface {
	Fetch(ctx context.Context, id string) (model.PortfolioAgent, error)
	Create(ctx context.Context, agent model.PortfolioAgent) error
}

type AssetAgentService interface {
	CreateAgent(
		ctx context.Context,
		agent model.AssetAgent,
	) error

	FetchInfo(
		ctx context.Context,
		agentID string,
	) (model.AssetAgent, error)

	Withdraw(
		ctx context.Context,
		agentID string,
		amount float64,
	) error

	Deposit(
		ctx context.Context,
		agentID string,
		amount float64,
	) error

	UpdateState(
		ctx context.Context,
		agentID string,
	) error
}
