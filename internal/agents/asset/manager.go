package asset

import "context"

// AgentManager constructs and initializes asset agents from persistent state.
// It acts as a factory: given an agent ID it fetches the stored agent data and
// returns a fully initialized Asset Agent.
type AgentManager struct {
	repo   AgentRepository
	market MarketService
}

func NewAgentManager(repo AgentRepository, market MarketService) *AgentManager {
	return &AgentManager{repo: repo, market: market}
}

func (m *AgentManager) FetchAssetAgent(ctx context.Context, id string) (*Agent, error) {
	return NewAgent(m.repo, m.market, WithInitialize(ctx, id))
}
