package portfolio

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kasaderos/camel/internal/agents/portfolio"
	"github.com/kasaderos/camel/internal/model"
	"github.com/samber/lo"
)

type Service struct {
	exchange Exchanger
	repo     PortfolioRepository
	market   portfolio.MarketService
}

func New(
	exchange Exchanger,
	repo PortfolioRepository,
	market portfolio.MarketService,
) *Service {
	return &Service{
		exchange: exchange,
		repo:     repo,
		market:   market,
	}
}

func (s *Service) CreatePortfolio(
	ctx context.Context,
	id string,
	assets []model.Asset,
	cash float64,
) error {
	err := s.repo.CreatePortfolio(ctx, model.Portfolio{
		ID:      id,
		Cash:    cash,
		Weights: map[string]float64{},
	})
	if err != nil {
		return fmt.Errorf("create portfolio: %w", err)
	}

	for _, asset := range assets {
		err := s.repo.CreatePortfolioAgent(
			ctx,
			model.PortfolioAgent{
				ID:          fmt.Sprintf("%s-%s", id, asset.ID),
				PortfolioID: id,
				AssetID:     asset.ID,
			},
		)
		if err != nil {
			return fmt.Errorf("create portfolio agent: %w", err)
		}
	}

	return nil
}

func (s *Service) Rebalance(ctx context.Context, p model.Portfolio) error {
	slog.Info("rebalance portfolio", "id", p.ID, "cash", p.Cash)

	agents, err := s.FetchPortfolioAgents(ctx, p.ID)
	if err != nil {
		return fmt.Errorf("fetch portfolio agents: %w", err)
	}

	slog.Info("portfolio agents", "id", p.ID, "total", len(agents))

	totalAgentsCash := p.Cash
	for _, agent := range agents {
		price, err := s.exchange.FetchPrice(ctx, agent.AssetID)
		if err != nil {
			return fmt.Errorf("fetch price: %w", err)
		}

		totalAgentsCash += agent.AssetQty * price
	}

	slog.Info("total portfolio value", "total agents cash", totalAgentsCash)

	weights := map[string]float64{}
	targetSums := map[string]float64{}

	scores, totalScore, err := fetchAgentScores(ctx, agents)
	if err != nil {
		return fmt.Errorf("fetch agent scores: %w", err)
	}

	for _, agent := range agents {
		score := scores[agent.AssetID]

		newWeight := 0.0
		if totalScore > 0 {
			newWeight = score / totalScore
		}

		weights[agent.AssetID] = newWeight
		targetSums[agent.AssetID] = newWeight * totalAgentsCash

		slog.Info(
			"current weights",
			"assetID", agent.AssetID,
			"weight", newWeight,
		)
	}

	for _, agent := range agents {
		sum := targetSums[agent.AssetID]

		_, err := agent.AdjustTargetSum(ctx, sum)
		if err != nil {
			return fmt.Errorf("adjust target sum: %w", err)
		}

		totalAgentsCash -= sum
	}

	if totalAgentsCash < 0 {
		slog.Warn("total agents cash negative", "remaining cash", totalAgentsCash)
		totalAgentsCash = 0
	}

	slog.Info(
		"update portfolio",
		"id", p.ID,
		"weights", weights,
		"cash", totalAgentsCash,
	)

	err = s.repo.UpdatePortfolio(ctx, model.Portfolio{
		ID:      p.ID,
		Weights: weights,
		Cash:    totalAgentsCash,
	})
	if err != nil {
		return fmt.Errorf("update portfolio: %w", err)
	}

	return nil
}

func fetchAgentScores(
	ctx context.Context,
	agents []*portfolio.Agent,
) (map[string]float64, float64, error) {
	const threshold = 0.01
	scores := map[string]float64{}
	totalScore := 0.0

	for _, a := range agents {
		score, err := a.FetchScore(ctx)
		if err != nil {
			return nil, 0.0, fmt.Errorf("fetch score: %w", err)
		}

		slog.Info(
			"current scores",
			"assetID", a.AssetID,
			"score", score,
		)

		if score > threshold {
			scores[a.AssetID] = score
			totalScore += score
		}
	}

	return scores, totalScore, nil
}

func (s *Service) FetchPortfolio(
	ctx context.Context,
	id string,
) (model.Portfolio, error) {
	return s.repo.FetchPortfolio(ctx, id)
}

func (s *Service) UpdatePortfolio(
	ctx context.Context,
	p model.Portfolio,
) error {
	return s.repo.UpdatePortfolio(ctx, p)
}

func (s *Service) FetchPortfolioAgents(
	ctx context.Context,
	id string,
) ([]*portfolio.Agent, error) {
	agents, err := s.repo.FetchPortfolioAgents(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetch portfolio agents: %w", err)
	}

	return lo.Map(agents, func(a model.PortfolioAgent, _ int) *portfolio.Agent {
		return portfolio.NewAgent(
			a,
			s.repo,
			s.market,
			s.exchange,
		)
	}), nil
}
