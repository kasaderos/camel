package model

type PortfolioAgent struct {
	ID          string
	PortfolioID string

	AssetID  string
	AssetQty float64

	Score float64
}
