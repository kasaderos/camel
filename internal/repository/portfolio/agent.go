package portfolio

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kasaderos/camel/internal/model"
	"github.com/samber/lo"
)

type dbAgent struct {
	ID          string  `db:"id"`
	PortfolioID string  `db:"portfolio_id"`
	AssetID     string  `db:"asset_id"`
	AssetQty    float64 `db:"asset_qty"`
	Score       float64 `db:"score"`
}

func (r *Repository) CreatePortfolioAgent(
	ctx context.Context,
	agent model.PortfolioAgent,
) error {
	query := `
		INSERT INTO portfolio_agents (
			id,
			portfolio_id,
			asset_id,
			asset_qty,
			score
		)
		VALUES (
			:id,
			:portfolio_id,
			:asset_id,
			:asset_qty,
			:score
		)
	`

	_, err := r.db.NamedExecContext(
		ctx,
		query,
		dbAgent{
			ID:          agent.ID,
			PortfolioID: agent.PortfolioID,
			AssetID:     agent.AssetID,
			AssetQty:    agent.AssetQty,
			Score:       agent.Score,
		})
	if err != nil {
		return fmt.Errorf("could not create agent: %w", err)
	}

	return nil
}

func (r *Repository) FetchPortfolioAgents(
	ctx context.Context,
	portfolioID string,
) ([]model.PortfolioAgent, error) {
	var agents []dbAgent

	query := `SELECT
				id,
				portfolio_id,
				asset_id,
				asset_qty,
				score
			  FROM portfolio_agents WHERE portfolio_id = $1`

	err := r.db.SelectContext(ctx, &agents, query, portfolioID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("agent not found: %w", err)
		}

		return nil, err
	}

	return lo.Map(agents, func(a dbAgent, _ int) model.PortfolioAgent {
		return model.PortfolioAgent{
			ID:          a.ID,
			PortfolioID: a.PortfolioID,
			AssetID:     a.AssetID,
			AssetQty:    a.AssetQty,
			Score:       a.Score,
		}
	}), nil
}

func (r *Repository) UpdatePortfolioAgent(
	ctx context.Context,
	agent model.PortfolioAgent,
) error {
	query := `UPDATE portfolio_agents
			  SET score = $1,
				  asset_qty = $2
			  WHERE id = $3`

	res, err := r.db.ExecContext(
		ctx,
		query,
		agent.Score,
		agent.AssetQty,
		agent.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update agent state: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("agent not found")
	}

	return nil
}

func (r *Repository) withTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
