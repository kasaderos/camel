package portfolio

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kasaderos/camel/internal/model"
	assetrepo "github.com/kasaderos/camel/internal/repository/agents/asset"
)

func (r *AgentRepository) Fetch(ctx context.Context, id string) (model.PortfolioAgent, error) {
	var rows []fetchRow

	query := `
		SELECT
			p.id AS id,
			p.portfolio_id AS portfolio_id,
			p.created_at AS created_at,
			p.updated_at AS updated_at,
			a.id AS asset_agent_id,
			a.asset_id AS asset_id,
			a.asset_qty AS asset_qty,
			a.cash AS cash,
			a.state AS state
		FROM portfolio_agents p
		LEFT JOIN asset_agents a
			ON a.portfolio_agent_id = p.id
		WHERE p.id = $1
		ORDER BY a.id
	`

	err := r.db.SelectContext(ctx, &rows, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PortfolioAgent{}, fmt.Errorf("agent not found: %w", err)
		}

		return model.PortfolioAgent{}, err
	}

	if len(rows) == 0 {
		return model.PortfolioAgent{}, fmt.Errorf("agent not found: %w", sql.ErrNoRows)
	}

	row := rows[0]

	agent := PortfolioAgent{
		ID:          row.ID,
		PortfolioID: row.PortfolioID,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}

	assetAgents := make([]model.AssetAgent, 0, len(rows))
	for _, row := range rows {
		if !row.AssetAgentID.Valid {
			continue
		}

		assetAgents = append(assetAgents, model.AssetAgent{
			ID:       row.AssetAgentID.String,
			AssetID:  row.AssetID.String,
			AssetQty: row.AssetQty.Float64,
			Cash:     row.Cash.Float64,
			State:    map[string]string(row.State),
		})
	}

	return agent.toModel(assetAgents), nil
}

type fetchRow struct {
	ID          string    `db:"id"`
	PortfolioID string    `db:"portfolio_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`

	AssetAgentID sql.NullString       `db:"asset_agent_id"`
	AssetID      sql.NullString       `db:"asset_id"`
	AssetQty     sql.NullFloat64      `db:"asset_qty"`
	Cash         sql.NullFloat64      `db:"cash"`
	State        assetrepo.AgentState `db:"state"`
}
