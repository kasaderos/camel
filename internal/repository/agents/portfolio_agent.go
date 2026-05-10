package agents

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kasaderos/camel/internal/model"
)

func (r *AgentRepository) Create(ctx context.Context, assets []model.AssetAgent) (model.PortfolioAgent, error) {
	now := time.Now()
	pa := PortfolioAgent{
		ID:          fmt.Sprintf("portfolio-agent-%d", now.UnixNano()),
		PortfolioID: fmt.Sprintf("portfolio-%d", now.UnixNano()),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query := `
		INSERT INTO portfolio_agents (id, portfolio_id, created_at, updated_at)
		VALUES (:id, :portfolio_id, :created_at, :updated_at)
		ON CONFLICT DO NOTHING
	`

	if _, err := r.db.NamedExecContext(ctx, query, pa); err != nil {
		return model.PortfolioAgent{}, fmt.Errorf("create portfolio agent: %w", err)
	}

	assetAgentIDs := make([]string, 0, len(assets))
	for i := range assets {
		assets[i].PortfolioAgentID = &pa.ID
		if err := r.CreateAgent(ctx, &assets[i]); err != nil {
			return model.PortfolioAgent{}, fmt.Errorf("create asset agent: %w", err)
		}
		assetAgentIDs = append(assetAgentIDs, assets[i].ID)
	}

	return pa.toModel(assetAgentIDs), nil
}

func (r *AgentRepository) Fetch(ctx context.Context, id string) (model.PortfolioAgent, error) {
	var rows []fetchRow

	query := `
		SELECT
			p.id AS id,
			p.portfolio_id AS portfolio_id,
			p.created_at AS created_at,
			p.updated_at AS updated_at,
			a.id AS asset_agent_id
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

	assetAgentIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		if !row.AssetAgentID.Valid {
			continue
		}

		assetAgentIDs = append(assetAgentIDs, row.AssetAgentID.String)
	}

	return agent.toModel(assetAgentIDs), nil
}

type fetchRow struct {
	ID           string         `db:"id"`
	PortfolioID  string         `db:"portfolio_id"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
	AssetAgentID sql.NullString `db:"asset_agent_id"`
}
