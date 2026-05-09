package model

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

	State map[string]string

	PortfolioAgentID *string
}
