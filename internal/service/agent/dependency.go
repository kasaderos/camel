package agent

import (
	"github.com/kasaderos/camel/internal/agents/asset"
	"github.com/kasaderos/camel/internal/agents/portfolio"
)

type (
	AssetAgentRepository     = asset.AgentRepository
	MarketService            = asset.MarketService
	PortfolioAssetAgent      = portfolio.AssetAgent
	PortfolioAgentRepository = portfolio.AgentRepository
)
