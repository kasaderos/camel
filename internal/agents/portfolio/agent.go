package portfolio

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"sort"
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

func (a *Agent) Rebalance(ctx context.Context) error {
	const threshold = 0.02

	// 1. Update state of all asset agents in portfolio
	err := a.Coordinate(ctx, func(ctx context.Context, agent AssetAgent) error {
		return agent.UpdateState(ctx)
	})
	if err != nil {
		return fmt.Errorf("update asset agent states: %w", err)
	}

	// 2. Get target newWeights based on current states
	newWeights, err := a.Portfolio(ctx, threshold)
	if err != nil {
		return fmt.Errorf("calculate portfolio weights: %w", err)
	}

	// 3. Calculate total portfolio value (all assets + cash)
	var totalPortfolioValue float64

	err = a.Coordinate(ctx, func(ctx context.Context, agent AssetAgent) error {
		sum, err := agent.FetchTotalSum(ctx)
		if err != nil {
			return fmt.Errorf("agent fetch total sum: %w", err)
		}

		totalPortfolioValue += sum

		return nil
	})
	if err != nil {
		return err
	}

	// 4. First pass: Sell/withdraw assets to generate free cash
	freeCash := 0.0

	err = a.Coordinate(ctx, func(ctx context.Context, agent AssetAgent) error {
		agentInfo := agent.FetchInfo(ctx)
		newWeight, exist := newWeights[agentInfo.AssetID]

		// Close positions for assets not in target weights
		if !exist {
			agentInfo, err := agent.ClosePosition(ctx)
			if err != nil {
				return fmt.Errorf("agent close position: %w", err)
			}

			freeCash += agentInfo.Cash

			return nil
		}

		// Sell excess if agent has more than target
		agentTotalSum, err := agent.FetchTotalSum(ctx)
		if err != nil {
			return fmt.Errorf("agent fetch total sum: %w", err)
		}

		targetSum := roundMoney(totalPortfolioValue * newWeight)
		excessSum := roundMoney(agentTotalSum - targetSum)

		// Withdraw excess or clean up small positions
		if excessSum > 0 {
			// Normal case: agent has more than target
			agentInfo, err := agent.WithdrawWithSell(ctx, excessSum)
			if err != nil {
				return fmt.Errorf("agent withdraw with sell: %w", err)
			}
			freeCash += agentInfo.Cash
		} else if agentInfo.AssetQty < 1e-3 && agentTotalSum > 0 {
			// Clean up: agent has negligible quantity but some cash/value
			// Withdraw all remaining value
			agentInfo, err := agent.WithdrawWithSell(ctx, agentTotalSum)
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

	// 5. Second pass: Buy/deposit assets using available free cash
	err = a.Coordinate(ctx, func(ctx context.Context, agent AssetAgent) error {
		agentInfo := agent.FetchInfo(ctx)
		newWeight, exist := newWeights[agentInfo.AssetID]

		// Skip assets not in target weights
		if !exist {
			return nil
		}

		// Buy more if agent has less than target and we have cash
		agentTotalSum, err := agent.FetchTotalSum(ctx)
		if err != nil {
			return fmt.Errorf("agent fetch total sum: %w", err)
		}

		targetSum := roundMoney(totalPortfolioValue * newWeight)
		deficitSum := roundMoney(targetSum - agentTotalSum)

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

func (a *Agent) Portfolio(ctx context.Context, threshold float64) (map[string]float64, error) {
	type candidate struct {
		assetID string
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

		candidates = append(candidates, candidate{
			assetID: agentInfo.AssetID,
			score:   score,
		})

		totalScore += score

		return nil
	})
	if err != nil {
		return nil, err
	}

	weights := make(map[string]float64)

	if totalScore == 0 {
		return weights, nil
	}

	for _, c := range candidates {
		weights[c.assetID] = c.score / totalScore
	}

	return weights, nil
}

func (a *Agent) PrintInfo(ctx context.Context, w io.Writer) {
	fmt.Fprintf(w, "portfolio_agent_id=%s portfolio_id=%s created_at=%s updated_at=%s\n",
		a.ID,
		a.PortfolioID,
		a.CreatedAt.Format(time.RFC3339),
		a.UpdatedAt.Format(time.RFC3339),
	)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "asset agents:")

	for _, agent := range a.assetAgents {
		info := agent.FetchInfo(ctx)
		fmt.Fprintf(w, "- id=%s asset_id=%s asset_qty=%.4f cash=%.2f state=%v\n",
			info.ID,
			info.AssetID,
			info.AssetQty,
			info.Cash,
			info.State,
		)
	}

	type summary struct {
		AssetID string
		Count   int
		Qty     float64
		Cash    float64
	}

	byAsset := map[string]*summary{}

	for _, agent := range a.assetAgents {
		info := agent.FetchInfo(ctx)

		s := byAsset[info.AssetID]
		if s == nil {
			s = &summary{AssetID: info.AssetID}
			byAsset[info.AssetID] = s
		}

		s.Count++
		s.Qty += info.AssetQty
		s.Cash += info.Cash
	}

	keys := make([]string, 0, len(byAsset))
	for k := range byAsset {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	fmt.Fprintln(w, "")

	weights, err := a.Portfolio(ctx, 0.02)
	if err != nil {
		fmt.Fprintf(w, "error calculating portfolio weights: %v\n", err)
		return
	}

	if len(weights) > 0 {
		wKeys := make([]string, 0, len(weights))
		for k := range weights {
			wKeys = append(wKeys, k)
		}

		sort.Strings(wKeys)

		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "portfolio weights:")

		for _, k := range wKeys {
			fmt.Fprintf(w, "- asset_id=%s weight=%.4f\n", k, weights[k])
		}
	}
}

// roundMoney rounds a float64 to 2 decimal places (cents)
func roundMoney(amount float64) float64 {
	return math.Round(amount*100) / 100
}
