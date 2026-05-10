package agents

import (
	"time"

	"github.com/kasaderos/camel/internal/model"
)

func fromAssetAgent(a model.AssetAgent) AssetAgent {
	return AssetAgent{
		ID:               a.ID,
		AssetID:          a.AssetID,
		PortfolioAgentID: a.PortfolioAgentID,
		AssetQty:         a.AssetQty,
		Cash:             a.Cash,
		State:            a.State.Data(),
	}
}

type PortfolioAgent struct {
	ID          string    `db:"id"`
	PortfolioID string    `db:"portfolio_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (a PortfolioAgent) toModel(assetAgentIDs []string) model.PortfolioAgent {
	return model.PortfolioAgent{
		ID:          a.ID,
		PortfolioID: a.PortfolioID,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,

		AssetAgentIDs: assetAgentIDs,
	}
}

func fromPortfolioAgent(a model.PortfolioAgent) PortfolioAgent {
	return PortfolioAgent{
		ID:          a.ID,
		PortfolioID: a.PortfolioID,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

type AssetAgent struct {
	ID               string         `db:"id"`
	AssetID          string         `db:"asset_id"`
	PortfolioAgentID *string        `db:"portfolio_agent_id"`
	AssetQty         float64        `db:"asset_qty"`
	Cash             float64        `db:"cash"`
	State            map[string]any `db:"state"`
}

func (a AssetAgent) toModel() *model.AssetAgent {
	state := model.State{}
	state.Load(a.State)

	return &model.AssetAgent{
		ID:               a.ID,
		AssetID:          a.AssetID,
		AssetQty:         a.AssetQty,
		Cash:             a.Cash,
		State:            state,
		PortfolioAgentID: a.PortfolioAgentID,
	}
}
