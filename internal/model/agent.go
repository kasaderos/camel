package model

import "time"

// State values
const (
	StateDate            = "date"
	EMA20Lookback5Change = "ema20_lookback5_change"
)

type AssetAgent struct {
	ID      string
	AssetID string

	AssetQty float64
	Cash     float64

	State State

	PortfolioAgentID *string
}

func (a AssetAgent) HasCash() bool {
	return a.Cash > 0
}

type PortfolioAgent struct {
	ID          string
	PortfolioID string

	AssetAgentIDs []string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (a AssetAgent) NoPositions() bool {
	return a.AssetQty < 1e-3
}
