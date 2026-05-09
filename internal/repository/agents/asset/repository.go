package asset

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kasaderos/camel/internal/model"
)

type AgentRepository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

func (r *AgentRepository) CreateAgent(ctx context.Context, agent model.AssetAgent) error {
	dbAgent := fromModel(agent)

	query := `
		INSERT INTO asset_agents (id, asset_id, portfolio_id, asset_qty, cash, state)
		VALUES (:id, :asset_id, :portfolio_id, :asset_qty, :cash, :state)
	`

	// NamedExecContext automatically maps struct fields to :name placeholders
	_, err := r.db.NamedExecContext(ctx, query, dbAgent)
	if err != nil {
		return fmt.Errorf("could not create agent: %w", err)
	}

	return nil
}

func (r *AgentRepository) FetchInfo(ctx context.Context, agentID string) (model.AssetAgent, error) {
	var agent AssetAgent

	query := `SELECT id, asset_id, portfolio_id, asset_qty, cash, state FROM asset_agents WHERE id = $1`

	err := r.db.GetContext(ctx, &agent, query, agentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssetAgent{}, fmt.Errorf("agent not found: %w", err)
		}

		return model.AssetAgent{}, err
	}

	return agent.toModel(), nil
}

func (r *AgentRepository) Withdraw(ctx context.Context, agentID string, q float64) error {
	return r.withTransaction(ctx, func(tx *sqlx.Tx) error {
		query := `UPDATE asset_agents SET cash = cash - $1 WHERE id = $2 AND cash >= $1`

		res, err := tx.ExecContext(ctx, query, q, agentID)
		if err != nil {
			return err
		}

		rows, _ := res.RowsAffected()
		if rows == 0 {
			return errors.New("insufficient funds or agent not found")
		}

		return nil
	})
}

func (r *AgentRepository) Deposit(ctx context.Context, agentID string, q float64) error {
	return r.withTransaction(ctx, func(tx *sqlx.Tx) error {
		query := `UPDATE asset_agents SET cash = cash + $1 WHERE id = $2`

		res, err := tx.ExecContext(ctx, query, q, agentID)
		if err != nil {
			return err
		}

		rows, _ := res.RowsAffected()
		if rows == 0 {
			return errors.New("agent not found")
		}

		return nil
	})
}

func (r *AgentRepository) UpdateState(ctx context.Context, agentID string, state map[string]string) error {
	query := `UPDATE asset_agents SET state = $1 WHERE id = $2`

	res, err := r.db.ExecContext(ctx, query, state, agentID)
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

// withTransaction is a helper to ensure tx.Rollback() or tx.Commit() is always called
func (r *AgentRepository) withTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
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
