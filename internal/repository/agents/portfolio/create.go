package portfolio

import (
	"context"
	"fmt"
	"time"

	"github.com/kasaderos/camel/internal/model"
)

func (r *AgentRepository) Create(ctx context.Context, agent model.PortfolioAgent) error {
	a := fromModel(agent)

	a.CreatedAt = time.Now()
	a.UpdatedAt = a.CreatedAt

	query := `
		INSERT INTO portfolio_agents (id, portfolio_id, created_at, updated_at)
		VALUES (:id, :portfolio_id, :created_at, :updated_at)
		ON CONFLICT DO NOTHING
	`

	_, err := r.db.NamedExecContext(ctx, query, a)
	if err != nil {
		return fmt.Errorf("could not create agent: %w", err)
	}

	return nil
}
