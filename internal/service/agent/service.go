package agent

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/kasaderos/camel/internal/agents/asset"
	"github.com/kasaderos/camel/internal/agents/portfolio"
	"github.com/kasaderos/camel/internal/model"
)

// Service tracks live asset agents and drives their monitoring lifecycle.
type Service struct {
	assetAgentRepository     AssetAgentRepository
	portfolioAgentRepository PortfolioAgentRepository
	market                   MarketService
}

func New(
	assetRepo AssetAgentRepository,
	portfolioRepo PortfolioAgentRepository,
	market MarketService,
) *Service {
	return &Service{
		assetAgentRepository:     assetRepo,
		portfolioAgentRepository: portfolioRepo,
		market:                   market,
	}
}

// InitAssetAgent loads an asset agent from persistent state by ID.
func (m *Service) InitAssetAgent(ctx context.Context, id string) (*asset.Agent, error) {
	agent, err := m.assetAgentRepository.FetchInfo(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetch asset agent: %w", err)
	}

	return asset.NewAgent(
		agent,
		m.assetAgentRepository,
		m.market,
	), nil
}

// InitPortfolioAgent loads a portfolio agent and all of its asset agents by ID.
func (m *Service) InitPortfolioAgent(ctx context.Context, id string) (*portfolio.Agent, error) {
	pa, err := m.portfolioAgentRepository.Fetch(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetch portfolio agent: %w", err)
	}

	assetAgents, err := lo.MapErr(pa.AssetAgentIDs, func(agentID string, _ int) (PortfolioAssetAgent, error) {
		return m.InitAssetAgent(ctx, agentID)
	})
	if err != nil {
		return nil, fmt.Errorf("init asset agents: %w", err)
	}

	return portfolio.NewAgent(
		pa,
		m.portfolioAgentRepository,
		assetAgents,
	), nil
}

// CreatePortfolioAgent persists a new portfolio with one asset agent per asset,
// then returns it fully initialized and ready for use.
func (m *Service) CreatePortfolioAgent(ctx context.Context, assets []model.Asset) (*portfolio.Agent, error) {
	agentModels := lo.Map(assets, func(a model.Asset, i int) model.AssetAgent {
		return model.AssetAgent{
			ID:      fmt.Sprintf("asset-agent-%d", i+1),
			AssetID: a.ID,
		}
	})

	pa, err := m.portfolioAgentRepository.Create(ctx, agentModels)
	if err != nil {
		return nil, fmt.Errorf("create portfolio: %w", err)
	}

	return m.InitPortfolioAgent(ctx, pa.ID)
}
