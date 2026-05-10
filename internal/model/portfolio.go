package model

import (
	"strconv"
	"time"
)

type PortfolioAgent struct {
	ID          string
	PortfolioID string

	AssetAgents []*AssetAgent

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *PortfolioAgent) Portfolio(threshold float64) map[string]float64 {
	agents := p.AssetAgents

	type candidate struct {
		assetID string
		score   float64
	}

	candidates := make([]candidate, 0, len(agents))

	var totalScore float64

	for _, agent := range agents {
		raw, ok := agent.State[EMA20Lookback5Change]
		if !ok {
			continue
		}

		score, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			continue
		}

		// long-only threshold filter
		if score < threshold {
			continue
		}

		candidates = append(candidates, candidate{
			assetID: agent.AssetID,
			score:   score,
		})

		totalScore += score
	}

	weights := make(map[string]float64)

	if totalScore == 0 {
		return weights
	}

	for _, c := range candidates {
		weights[c.assetID] = c.score / totalScore
	}

	return weights
}
