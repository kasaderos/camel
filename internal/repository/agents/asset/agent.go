package asset

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/kasaderos/camel/internal/model"
)

type AgentState map[string]string

// Value handles saving to DB
func (a AgentState) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan handles reading from DB
func (a *AgentState) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

type AssetAgent struct {
	ID          string     `db:"id"`
	AssetID     string     `db:"asset_id"`
	PortfolioID *string    `db:"portfolio_id"`
	AssetQty    float64    `db:"asset_qty"`
	Cash        float64    `db:"cash"`
	State       AgentState `db:"state"`
}

func (a AssetAgent) toModel() model.AssetAgent {
	return model.AssetAgent{
		ID:          a.ID,
		AssetID:     a.AssetID,
		PortfolioID: a.PortfolioID,
		AssetQty:    a.AssetQty,
		Cash:        a.Cash,
		State:       map[string]string(a.State),
	}
}

func fromModel(a model.AssetAgent) AssetAgent {
	return AssetAgent{
		ID:          a.ID,
		AssetID:     a.AssetID,
		PortfolioID: a.PortfolioID,
		AssetQty:    a.AssetQty,
		Cash:        a.Cash,
		State:       AgentState(a.State),
	}
}
