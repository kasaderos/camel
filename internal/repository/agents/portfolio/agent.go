package portfolio

import (
	"time"

	"github.com/kasaderos/camel/internal/model"
)

type PortfolioAgent struct {
	ID          string    `db:"id"`
	PortfolioID string    `db:"portfolio_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (a PortfolioAgent) toModel(assetAgents []model.AssetAgent) model.PortfolioAgent {
	return model.PortfolioAgent{
		ID:          a.ID,
		PortfolioID: a.PortfolioID,
		AssetAgents: assetAgents,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

func fromModel(a model.PortfolioAgent) PortfolioAgent {
	return PortfolioAgent{
		ID:          a.ID,
		PortfolioID: a.PortfolioID,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}
