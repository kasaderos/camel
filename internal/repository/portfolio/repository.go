package portfolio

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/kasaderos/camel/pkg/slices"

	"github.com/kasaderos/camel/internal/model"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FetchPortfolio(ctx context.Context, id int64) (*model.Portfolio, error) {
	var p portfolio

	err := r.db.SelectContext(
		ctx,
		&p,
		"SELECT id, name, weights, cache, policy_id FROM portfolios WHERE id = ?",
		id,
	)
	if err != nil {
		return nil, err
	}

	return toModel(p), nil
}

func (r *Repository) SearchPortfolios(ctx context.Context, offset, limit int) ([]*model.Portfolio, error) {
	var portfolios []portfolio

	err := r.db.SelectContext(
		ctx,
		&portfolios,
		"SELECT id, name, weights, cache, policy_id FROM portfolios ORDER BY id LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, err
	}

	return slices.Map(portfolios, func(p portfolio) (*model.Portfolio, error) {
		return toModel(p), nil
	})
}

func (r *Repository) CreatePortfolio(ctx context.Context, p model.Portfolio) (model.Portfolio, error) {
	result, err := r.db.ExecContext(
		ctx,
		"INSERT INTO portfolios (name, weights, cache, policy_id) VALUES (?, ?, ?, ?)",
		p.Name, p.Weights, p.Cache,
	)
	if err != nil {
		return p, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return p, err
	}

	p.ID = int64(id)

	return p, nil
}

func (r *Repository) UpdatePortfolio(ctx context.Context, p model.Portfolio) error {
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE portfolios SET name = ?, weights = ? WHERE id = ?",
		p.Name, p.Weights, p.ID,
	)

	return err
}

func (r *Repository) DeletePortfolio(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(
		ctx,
		"DELETE FROM portfolios WHERE id = ?",
		id,
	)

	return err
}
