package model

import "time"

type PortfolioAgent struct {
	ID          string
	PortfolioID string

	AssetAgents []AssetAgent

	CreatedAt time.Time
	UpdatedAt time.Time
}
