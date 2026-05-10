package portfolio

import (
	"context"
	"fmt"
	"io"
	"log/slog"
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

func (a *Agent) Rebalance(ctx context.Context) error {
	const threshold = 0.02

	// update state of all asset agents in portfolio
	for _, assetAgent := range a.assetAgents {
		err := assetAgent.UpdateState(ctx)
		if err != nil {
			return fmt.Errorf("update asset agent state: %w", err)
		}
	}

	return nil
}

func (a *Agent) Portfolio(ctx context.Context, threshold float64) (map[string]float64, error) {
	type candidate struct {
		assetID string
		score   float64
	}

	candidates := make([]candidate, 0, len(a.assetAgents))

	var totalScore float64

	for _, agent := range a.assetAgents {
		agentInfo := agent.FetchInfo(ctx)
		agentState := agent.FetchState(ctx)

		score, ok := agentState.EmaChange()
		if !ok {
			slog.Error("agent state ema_change invalid", "id", agentInfo.ID)
			continue
		}

		// long-only threshold filter
		if score < threshold {
			continue
		}

		candidates = append(candidates, candidate{
			assetID: agentInfo.AssetID,
			score:   score,
		})

		totalScore += score
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

	weights, _ := a.Portfolio(ctx, 0.02)
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
