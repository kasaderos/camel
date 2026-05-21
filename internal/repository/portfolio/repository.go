package portfolio

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kasaderos/camel/internal/model"
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

type JSONMap map[string]float64

func (m *JSONMap) Scan(src any) error {
	return json.Unmarshal(src.([]byte), m)
}

type Portfolio struct {
	ID      string  `db:"id"`
	Cash    float64 `db:"cash"`
	Weights JSONMap `db:"weights"`
}

func (r *Repository) CreatePortfolio(
	ctx context.Context,
	portfolio model.Portfolio,
) error {
	const query = `
		INSERT INTO portfolios (id, cash, weights)
		VALUES (:id, :cash, :weights)`

	_, err := r.db.NamedExecContext(ctx, query, Portfolio{
		ID:      portfolio.ID,
		Cash:    portfolio.Cash,
		Weights: portfolio.Weights,
	})
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}

func (r *Repository) FetchPortfolio(
	ctx context.Context,
	id string,
) (model.Portfolio, error) {
	const query = `
		SELECT
			id,
			cash,
			weights
		FROM portfolios
		WHERE id = $1
	`

	var portfolio Portfolio

	err := r.db.GetContext(ctx, &portfolio, query, id)
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("fetch portfolio: %w", err)
	}

	return model.Portfolio{
		ID:      portfolio.ID,
		Cash:    portfolio.Cash,
		Weights: portfolio.Weights,
	}, nil
}

func (r *Repository) UpdatePortfolio(
	ctx context.Context,
	portfolio model.Portfolio,
) error {
	const query = `
		UPDATE portfolios
		SET cash = :cash,
	        weights = :weights
		WHERE id = :id
	`

	_, err := r.db.NamedExecContext(ctx, query, Portfolio{
		ID:      portfolio.ID,
		Cash:    portfolio.Cash,
		Weights: portfolio.Weights,
	})
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return nil
}
