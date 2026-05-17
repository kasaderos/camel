package portfolio

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"time"

	"github.com/kasaderos/camel/internal/model"
)

type Agent struct {
	model.PortfolioAgent

	assetAgents []AssetAgent
	repository  AgentRepository
}

func NewAgent(agent model.PortfolioAgent, repo AgentRepository, assetAgents []AssetAgent) *Agent {
	return &Agent{
		PortfolioAgent: agent,
		repository:     repo,
		assetAgents:    assetAgents,
	}
}

// Coordinate executes a function on each asset agent in the portfolio
func (a *Agent) Coordinate(ctx context.Context, fn func(context.Context, AssetAgent) error) error {
	for _, assetAgent := range a.assetAgents {
		if err := fn(ctx, assetAgent); err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) UpdatePortfolio(ctx context.Context) (*model.Portfolio, error) {
	err := a.Coordinate(ctx, func(ctx context.Context, agent AssetAgent) error {
		_, err := agent.UpdateState(ctx)

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("update agent states: %w", err)
	}

	newPortfolio, err := a.FetchPortfolio(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch portfolio weights: %w", err)
	}

	return newPortfolio, nil
}

func (a *Agent) Rebalance(ctx context.Context) error {
	portfolio, err := a.FetchPortfolio(ctx)
	if err != nil {
		return fmt.Errorf("fetch portfolio weights: %w", err)
	}

	// 1. Update portfolio
	newPortfolio, err := a.UpdatePortfolio(ctx)
	if err != nil {
		return fmt.Errorf("update portfolio: %w", err)
	}

	// 2. First pass: Sell/withdraw assets to generate free cash
	freeCash := 0.0

	err = a.Coordinate(ctx, func(ctx context.Context, agent AssetAgent) error {
		agentInfo := agent.FetchInfo(ctx)

		asset, _ := portfolio.Asset(agentInfo.AssetID)
		newAsset, exist := newPortfolio.Asset(agentInfo.AssetID)

		// Close positions for assets not in target weights
		if !exist {
			agentInfo, err := agent.ClosePosition(ctx)
			if err != nil {
				return fmt.Errorf("agent close position: %w", err)
			}

			freeCash += agentInfo.Cash

			return nil
		}

		excessSum := roundMoney(asset.Sum() - newAsset.Sum())

		// Withdraw excess or clean up small positions
		if excessSum > 0 {
			// Normal case: agent has more than target
			agentInfo, err := agent.WithdrawWithSell(ctx, excessSum)
			if err != nil {
				return fmt.Errorf("agent withdraw with sell: %w", err)
			}

			freeCash += agentInfo.Cash
		} else if agentInfo.NoPositions() && agentInfo.HasCash() {
			// Withdraw all remaining
			agentInfo, err := agent.Withdraw(ctx, agentInfo.Cash)
			if err != nil {
				return fmt.Errorf("agent withdraw with sell: %w", err)
			}

			freeCash += agentInfo.Cash
		}

		return nil
	})
	if err != nil {
		return err
	}

	// 3. Second pass: Buy/deposit assets using available free cash
	err = a.Coordinate(ctx, func(ctx context.Context, agent AssetAgent) error {
		agentInfo := agent.FetchInfo(ctx)

		asset, _ := portfolio.Asset(agentInfo.AssetID)
		targetAsset, exist := newPortfolio.Asset(agentInfo.AssetID)

		// Skip assets not in target weights
		if !exist {
			return nil
		}

		deficitSum := roundMoney(targetAsset.Sum() - asset.Sum())

		if deficitSum > 0 && freeCash >= deficitSum {
			_, err := agent.DepositWithBuy(ctx, deficitSum)
			if err != nil {
				return fmt.Errorf("agent deposit with buy: %w", err)
			}

			freeCash -= deficitSum
		}

		return nil
	})

	return nil
}

func (a *Agent) FetchPortfolio(ctx context.Context) (*model.Portfolio, error) {
	const threshold = 0.01

	type candidate struct {
		assetID string
		price   float64
		qty     float64
		score   float64
	}

	candidates := make([]candidate, 0, len(a.assetAgents))

	var totalScore float64

	// Collect candidates above threshold
	err := a.Coordinate(ctx, func(ctx context.Context, agent AssetAgent) error {
		agentInfo := agent.FetchInfo(ctx)

		score, ok := agentInfo.State.EmaChange()
		if !ok {
			slog.Error("agent state ema_change invalid", "id", agentInfo.ID)
			return nil
		}

		// long-only threshold filter
		if score < threshold {
			return nil
		}

		price, err := agent.FetchPrice(ctx)
		if err != nil {
			return fmt.Errorf("fetch price: %w", err)
		}

		candidates = append(candidates, candidate{
			assetID: agentInfo.AssetID,
			score:   score,
			price:   price,
			qty:     agentInfo.AssetQty,
		})

		totalScore += score

		return nil
	})
	if err != nil {
		return nil, err
	}

	assets := make(map[string]model.PortfolioAsset)

	if totalScore == 0 {
		return &model.Portfolio{Assets: assets}, nil
	}

	for _, c := range candidates {
		assets[c.assetID] = model.PortfolioAsset{
			AssetID: c.assetID,
			Price:   c.price,
			Qty:     c.qty,
			Weight:  c.score / totalScore,
		}
	}

	return &model.Portfolio{Assets: assets}, nil
}

func (a *Agent) PrintInfo(ctx context.Context, w io.Writer) {
	fmt.Fprintf(w, "portfolio_agent_id=%s portfolio_id=%s created_at=%s updated_at=%s\n",
		a.ID,
		a.PortfolioID,
		a.CreatedAt.Format(time.RFC3339),
		a.UpdatedAt.Format(time.RFC3339),
	)
	fmt.Fprintln(w, "")

	portfolio, err := a.FetchPortfolio(ctx)
	if err != nil {
		fmt.Fprintf(w, "error calculating portfolio weights: %v\n", err)
		return
	}

	portfolio.Print(w)
}

// roundMoney rounds a float64 to 2 decimal places (cents)
func roundMoney(amount float64) float64 {
	return math.Round(amount*100) / 100
}
