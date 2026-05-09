package portfolio

import (
	"context"
	"fmt"

	"github.com/kasaderos/camel/internal/model"
)

type PortfolioAgentService struct {
	assetAgentService AssetAgentService
	agentRepo         AgentRepository
}

func New(
	assetAgentService AssetAgentService,
	agentRepo AgentRepository,
) *PortfolioAgentService {
	return &PortfolioAgentService{
		assetAgentService: assetAgentService,
		agentRepo:         agentRepo,
	}
}

func (s *PortfolioAgentService) CreatePortfolio(
	ctx context.Context,
	assets []model.Asset,
) (model.PortfolioAgent, error) {
	assetAgents := make([]model.AssetAgent, len(assets))
	for i, asset := range assets {
		assetAgents[i] = model.AssetAgent{
			ID:      fmt.Sprintf("asset-agent-%d", i+1),
			AssetID: asset.ID,
		}
	}

	portfolioAgent := model.PortfolioAgent{
		ID:          "portfolio-agent-1",
		PortfolioID: "portfolio-1",
		AssetAgents: assetAgents,
	}

	err := s.agentRepo.Create(ctx, portfolioAgent)
	if err != nil {
		return model.PortfolioAgent{}, fmt.Errorf("create agent: %w", err)
	}

	// Persist asset agents with a portfolio_id link.
	for _, assetAgent := range portfolioAgent.AssetAgents {
		assetAgent.PortfolioID = new(portfolioAgent.PortfolioID)

		if err := s.assetAgentService.CreateAgent(ctx, assetAgent); err != nil {
			return model.PortfolioAgent{}, fmt.Errorf("create asset agent: %w", err)
		}
	}

	return portfolioAgent, nil
}

func (s *PortfolioAgentService) Fetch(
	ctx context.Context,
	agentID string,
) (model.PortfolioAgent, error) {
	agent, err := s.agentRepo.Fetch(ctx, agentID)
	if err != nil {
		return model.PortfolioAgent{}, fmt.Errorf("fetch agent: %w", err)
	}

	return agent, nil
}

func (s *PortfolioAgentService) Rebalance(
	ctx context.Context,
	agentID string,
) error {
	agent, err := s.Fetch(ctx, agentID)
	if err != nil {
		return fmt.Errorf("fetch agent: %w", err)
	}

	// update state of all asset agents in portfolio
	for _, assetAgent := range agent.AssetAgents {
		err = s.assetAgentService.UpdateState(ctx, assetAgent.ID)
		if err != nil {
			return fmt.Errorf("update asset agent state: %w", err)
		}
	}

	// refetch
	agent, err = s.Fetch(ctx, agentID)
	if err != nil {
		return fmt.Errorf("fetch agent: %w", err)
	}

	return nil
}
